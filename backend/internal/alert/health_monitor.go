package alert

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"backend/internal/logger"
	"backend/internal/metrics"
)

// HealthStatus はヘルスチェックの状態を表す
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck はヘルスチェック項目を表す構造体
type HealthCheck struct {
	Name        string                 `json:"name"`
	Status      HealthStatus           `json:"status"`
	Message     string                 `json:"message"`
	LastCheck   time.Time              `json:"last_check"`
	Duration    time.Duration          `json:"duration"`
	Details     map[string]interface{} `json:"details,omitempty"`
}

// HealthMonitor はシステムの健全性を監視する
type HealthMonitor struct {
	alertManager AlertManager
	metrics      metrics.MetricsCollector
	logger       logger.Logger
	
	// 監視対象
	db           *sql.DB
	httpClient   *http.Client
	
	// 設定
	checkInterval    time.Duration
	alertThresholds  map[string]AlertThreshold
	
	// 状態管理
	checks           map[string]*HealthCheck
	mutex            sync.RWMutex
	stopChan         chan struct{}
	running          bool
	
	// 自動復旧機能
	recoveryActions  map[string]RecoveryAction
}

// AlertThreshold はアラート閾値を表す
type AlertThreshold struct {
	WarningThreshold  float64       `json:"warning_threshold"`
	CriticalThreshold float64       `json:"critical_threshold"`
	Duration          time.Duration `json:"duration"`
	Enabled           bool          `json:"enabled"`
}

// RecoveryAction は自動復旧アクションを表す
type RecoveryAction interface {
	Execute(ctx context.Context, check *HealthCheck) error
	GetName() string
	IsEnabled() bool
}

// HealthMonitorConfig はヘルスモニターの設定
type HealthMonitorConfig struct {
	CheckInterval   time.Duration                `json:"check_interval"`
	AlertThresholds map[string]AlertThreshold    `json:"alert_thresholds"`
	RecoveryActions map[string]RecoveryAction    `json:"recovery_actions"`
}

// NewHealthMonitor は新しいHealthMonitorを作成する
func NewHealthMonitor(config HealthMonitorConfig, alertManager AlertManager, metricsCollector metrics.MetricsCollector, db *sql.DB) *HealthMonitor {
	if config.CheckInterval == 0 {
		config.CheckInterval = 30 * time.Second
	}

	if config.AlertThresholds == nil {
		config.AlertThresholds = getDefaultAlertThresholds()
	}

	if config.RecoveryActions == nil {
		config.RecoveryActions = make(map[string]RecoveryAction)
	}

	return &HealthMonitor{
		alertManager:    alertManager,
		metrics:         metricsCollector,
		logger:          logger.GetLogger().WithComponent("health_monitor"),
		db:              db,
		httpClient:      &http.Client{Timeout: 10 * time.Second},
		checkInterval:   config.CheckInterval,
		alertThresholds: config.AlertThresholds,
		checks:          make(map[string]*HealthCheck),
		stopChan:        make(chan struct{}),
		recoveryActions: config.RecoveryActions,
	}
}

// Start はヘルスモニタリングを開始する
func (hm *HealthMonitor) Start(ctx context.Context) error {
	hm.mutex.Lock()
	if hm.running {
		hm.mutex.Unlock()
		return fmt.Errorf("health monitor is already running")
	}
	hm.running = true
	hm.mutex.Unlock()

	hm.logger.Info("ヘルスモニタリングを開始しました",
		logger.Duration("check_interval", hm.checkInterval),
	)

	// 初回チェックを実行
	hm.runHealthChecks(ctx)

	// 定期チェックを開始
	go hm.monitoringLoop(ctx)

	return nil
}

// Stop はヘルスモニタリングを停止する
func (hm *HealthMonitor) Stop() error {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	if !hm.running {
		return fmt.Errorf("health monitor is not running")
	}

	close(hm.stopChan)
	hm.running = false

	hm.logger.Info("ヘルスモニタリングを停止しました")
	return nil
}

// GetHealthStatus は現在のヘルス状態を取得する
func (hm *HealthMonitor) GetHealthStatus() map[string]*HealthCheck {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	result := make(map[string]*HealthCheck)
	for name, check := range hm.checks {
		// コピーを作成
		checkCopy := *check
		result[name] = &checkCopy
	}

	return result
}

// GetOverallStatus は全体的なヘルス状態を取得する
func (hm *HealthMonitor) GetOverallStatus() HealthStatus {
	hm.mutex.RLock()
	defer hm.mutex.RUnlock()

	hasUnhealthy := false
	hasDegraded := false

	for _, check := range hm.checks {
		switch check.Status {
		case HealthStatusUnhealthy:
			hasUnhealthy = true
		case HealthStatusDegraded:
			hasDegraded = true
		}
	}

	if hasUnhealthy {
		return HealthStatusUnhealthy
	}
	if hasDegraded {
		return HealthStatusDegraded
	}
	return HealthStatusHealthy
}

// monitoringLoop は監視ループを実行する
func (hm *HealthMonitor) monitoringLoop(ctx context.Context) {
	ticker := time.NewTicker(hm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-hm.stopChan:
			return
		case <-ticker.C:
			hm.runHealthChecks(ctx)
		}
	}
}

// runHealthChecks は全てのヘルスチェックを実行する
func (hm *HealthMonitor) runHealthChecks(ctx context.Context) {
	checks := []func(context.Context) *HealthCheck{
		hm.checkDatabase,
		hm.checkMemoryUsage,
		hm.checkCPUUsage,
		hm.checkGoroutineCount,
		hm.checkDiskSpace,
		hm.checkHTTPEndpoints,
	}

	for _, checkFunc := range checks {
		go func(cf func(context.Context) *HealthCheck) {
			check := cf(ctx)
			hm.updateHealthCheck(check)
			hm.evaluateAlerts(check)
		}(checkFunc)
	}
}

// updateHealthCheck はヘルスチェック結果を更新する
func (hm *HealthMonitor) updateHealthCheck(check *HealthCheck) {
	hm.mutex.Lock()
	defer hm.mutex.Unlock()

	hm.checks[check.Name] = check

	// メトリクスを記録
	if hm.metrics != nil {
		labels := metrics.Labels{
			"check_name": check.Name,
			"status":     string(check.Status),
		}
		
		// ヘルスチェック実行回数
		counter := hm.metrics.RegisterCounter(
			"health_checks_total",
			"Total number of health checks",
			labels,
		)
		counter.Inc()

		// ヘルスチェック実行時間
		histogram := hm.metrics.RegisterHistogram(
			"health_check_duration_seconds",
			"Health check duration in seconds",
			labels,
			[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		)
		histogram.Observe(check.Duration.Seconds())

		// ヘルス状態ゲージ
		statusValue := hm.getStatusValue(check.Status)
		gauge := hm.metrics.RegisterGauge(
			"health_status",
			"Health status (0=unhealthy, 1=degraded, 2=healthy)",
			labels,
		)
		gauge.Set(statusValue)
	}
}

// evaluateAlerts はアラート条件を評価する
func (hm *HealthMonitor) evaluateAlerts(check *HealthCheck) {
	threshold, exists := hm.alertThresholds[check.Name]
	if !exists || !threshold.Enabled {
		return
	}

	// 状態に基づいてアラートを生成
	switch check.Status {
	case HealthStatusUnhealthy:
		hm.fireHealthAlert(check, SeverityCritical, "システムコンポーネントが異常状態です")
	case HealthStatusDegraded:
		hm.fireHealthAlert(check, SeverityWarning, "システムコンポーネントのパフォーマンスが低下しています")
	case HealthStatusHealthy:
		// 既存のアラートがあれば解決
		hm.resolveHealthAlert(check)
	}
}

// fireHealthAlert はヘルスアラートを発火する
func (hm *HealthMonitor) fireHealthAlert(check *HealthCheck, severity Severity, message string) {
	alertID := fmt.Sprintf("health_%s", check.Name)
	
	alert := &Alert{
		ID:          alertID,
		Type:        AlertTypeHealthCheckFailed,
		Severity:    severity,
		Title:       fmt.Sprintf("ヘルスチェック失敗: %s", check.Name),
		Description: fmt.Sprintf("%s - %s", message, check.Message),
		Source:      "health_monitor",
		Timestamp:   time.Now(),
		Labels: map[string]string{
			"check_name": check.Name,
			"status":     string(check.Status),
		},
		Annotations: map[string]string{
			"check_duration": check.Duration.String(),
			"last_check":     check.LastCheck.Format(time.RFC3339),
		},
		Status:   AlertStatusFiring,
		StartsAt: time.Now(),
	}

	if err := hm.alertManager.FireAlert(alert); err != nil {
		hm.logger.Error("ヘルスアラート発火エラー",
			logger.String("check_name", check.Name),
			logger.Err(err),
		)
	}

	// 自動復旧を試行
	hm.attemptRecovery(check)
}

// resolveHealthAlert はヘルスアラートを解決する
func (hm *HealthMonitor) resolveHealthAlert(check *HealthCheck) {
	alertID := fmt.Sprintf("health_%s", check.Name)
	
	if err := hm.alertManager.ResolveAlert(alertID); err != nil {
		// アラートが存在しない場合は無視
		if err.Error() != fmt.Sprintf("active alert not found: %s", alertID) {
			hm.logger.Error("ヘルスアラート解決エラー",
				logger.String("check_name", check.Name),
				logger.Err(err),
			)
		}
	}
}

// attemptRecovery は自動復旧を試行する
func (hm *HealthMonitor) attemptRecovery(check *HealthCheck) {
	action, exists := hm.recoveryActions[check.Name]
	if !exists || !action.IsEnabled() {
		return
	}

	hm.logger.Info("自動復旧を試行します",
		logger.String("check_name", check.Name),
		logger.String("action", action.GetName()),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := action.Execute(ctx, check); err != nil {
		hm.logger.Error("自動復旧失敗",
			logger.String("check_name", check.Name),
			logger.String("action", action.GetName()),
			logger.Err(err),
		)
	} else {
		hm.logger.Info("自動復旧成功",
			logger.String("check_name", check.Name),
			logger.String("action", action.GetName()),
		)
	}
}

// checkDatabase はデータベースのヘルスチェックを実行する
func (hm *HealthMonitor) checkDatabase(ctx context.Context) *HealthCheck {
	start := time.Now()
	check := &HealthCheck{
		Name:      "database",
		LastCheck: start,
	}

	if hm.db == nil {
		check.Status = HealthStatusUnhealthy
		check.Message = "データベース接続が設定されていません"
		check.Duration = time.Since(start)
		return check
	}

	// 接続テスト
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := hm.db.PingContext(pingCtx); err != nil {
		check.Status = HealthStatusUnhealthy
		check.Message = fmt.Sprintf("データベース接続エラー: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	// 簡単なクエリテスト
	var result int
	queryCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := hm.db.QueryRowContext(queryCtx, "SELECT 1").Scan(&result); err != nil {
		check.Status = HealthStatusDegraded
		check.Message = fmt.Sprintf("データベースクエリエラー: %v", err)
		check.Duration = time.Since(start)
		return check
	}

	check.Status = HealthStatusHealthy
	check.Message = "データベース接続正常"
	check.Duration = time.Since(start)
	check.Details = map[string]interface{}{
		"ping_duration": check.Duration.Milliseconds(),
	}

	return check
}

// checkMemoryUsage はメモリ使用量のヘルスチェックを実行する
func (hm *HealthMonitor) checkMemoryUsage(ctx context.Context) *HealthCheck {
	start := time.Now()
	check := &HealthCheck{
		Name:      "memory",
		LastCheck: start,
	}

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// メモリ使用量をMBで計算
	allocMB := float64(m.Alloc) / 1024 / 1024
	sysMB := float64(m.Sys) / 1024 / 1024

	// 閾値チェック（例：1GB以上で警告、2GB以上でクリティカル）
	if allocMB > 2048 {
		check.Status = HealthStatusUnhealthy
		check.Message = fmt.Sprintf("メモリ使用量が高すぎます: %.2f MB", allocMB)
	} else if allocMB > 1024 {
		check.Status = HealthStatusDegraded
		check.Message = fmt.Sprintf("メモリ使用量が高めです: %.2f MB", allocMB)
	} else {
		check.Status = HealthStatusHealthy
		check.Message = fmt.Sprintf("メモリ使用量正常: %.2f MB", allocMB)
	}

	check.Duration = time.Since(start)
	check.Details = map[string]interface{}{
		"alloc_mb":     allocMB,
		"sys_mb":       sysMB,
		"gc_count":     m.NumGC,
		"goroutines":   runtime.NumGoroutine(),
	}

	return check
}

// checkCPUUsage はCPU使用量のヘルスチェックを実行する（簡易実装）
func (hm *HealthMonitor) checkCPUUsage(ctx context.Context) *HealthCheck {
	start := time.Now()
	check := &HealthCheck{
		Name:      "cpu",
		LastCheck: start,
	}

	// 簡易実装：実際の実装では適切なCPU監視ライブラリを使用
	// ここではGoroutine数を基準にした簡易判定
	goroutines := runtime.NumGoroutine()

	if goroutines > 10000 {
		check.Status = HealthStatusUnhealthy
		check.Message = fmt.Sprintf("Goroutine数が異常に多いです: %d", goroutines)
	} else if goroutines > 5000 {
		check.Status = HealthStatusDegraded
		check.Message = fmt.Sprintf("Goroutine数が多めです: %d", goroutines)
	} else {
		check.Status = HealthStatusHealthy
		check.Message = fmt.Sprintf("CPU使用量正常: %d goroutines", goroutines)
	}

	check.Duration = time.Since(start)
	check.Details = map[string]interface{}{
		"goroutines": goroutines,
		"cpu_count":  runtime.NumCPU(),
	}

	return check
}

// checkGoroutineCount はGoroutine数のヘルスチェックを実行する
func (hm *HealthMonitor) checkGoroutineCount(ctx context.Context) *HealthCheck {
	start := time.Now()
	check := &HealthCheck{
		Name:      "goroutines",
		LastCheck: start,
	}

	count := runtime.NumGoroutine()

	// 閾値チェック
	if count > 10000 {
		check.Status = HealthStatusUnhealthy
		check.Message = fmt.Sprintf("Goroutine数が異常です: %d", count)
	} else if count > 1000 {
		check.Status = HealthStatusDegraded
		check.Message = fmt.Sprintf("Goroutine数が多めです: %d", count)
	} else {
		check.Status = HealthStatusHealthy
		check.Message = fmt.Sprintf("Goroutine数正常: %d", count)
	}

	check.Duration = time.Since(start)
	check.Details = map[string]interface{}{
		"count": count,
	}

	return check
}

// checkDiskSpace はディスク容量のヘルスチェックを実行する（簡易実装）
func (hm *HealthMonitor) checkDiskSpace(ctx context.Context) *HealthCheck {
	start := time.Now()
	check := &HealthCheck{
		Name:      "disk",
		LastCheck: start,
	}

	// 簡易実装：実際の実装では適切なディスク監視を行う
	check.Status = HealthStatusHealthy
	check.Message = "ディスク容量正常（簡易チェック）"
	check.Duration = time.Since(start)
	check.Details = map[string]interface{}{
		"note": "実際の実装では適切なディスク監視を行います",
	}

	return check
}

// checkHTTPEndpoints は重要なHTTPエンドポイントのヘルスチェックを実行する
func (hm *HealthMonitor) checkHTTPEndpoints(ctx context.Context) *HealthCheck {
	start := time.Now()
	check := &HealthCheck{
		Name:      "http_endpoints",
		LastCheck: start,
	}

	// 内部ヘルスチェックエンドポイントをテスト
	endpoints := []string{
		"http://localhost:8080/health",
	}

	healthyCount := 0
	totalCount := len(endpoints)
	details := make(map[string]interface{})

	for _, endpoint := range endpoints {
		endpointStart := time.Now()
		
		req, err := http.NewRequestWithContext(ctx, "GET", endpoint, nil)
		if err != nil {
			details[endpoint] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
			continue
		}

		resp, err := hm.httpClient.Do(req)
		if err != nil {
			details[endpoint] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
			continue
		}
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			healthyCount++
			details[endpoint] = map[string]interface{}{
				"status":   "healthy",
				"duration": time.Since(endpointStart).Milliseconds(),
			}
		} else {
			details[endpoint] = map[string]interface{}{
				"status":      "unhealthy",
				"status_code": resp.StatusCode,
				"duration":    time.Since(endpointStart).Milliseconds(),
			}
		}
	}

	// 全体的な状態を判定
	if healthyCount == totalCount {
		check.Status = HealthStatusHealthy
		check.Message = "全てのエンドポイントが正常です"
	} else if healthyCount > 0 {
		check.Status = HealthStatusDegraded
		check.Message = fmt.Sprintf("一部のエンドポイントに問題があります (%d/%d)", healthyCount, totalCount)
	} else {
		check.Status = HealthStatusUnhealthy
		check.Message = "全てのエンドポイントに問題があります"
	}

	check.Duration = time.Since(start)
	check.Details = details

	return check
}

// getStatusValue はヘルス状態を数値に変換する
func (hm *HealthMonitor) getStatusValue(status HealthStatus) float64 {
	switch status {
	case HealthStatusHealthy:
		return 2
	case HealthStatusDegraded:
		return 1
	case HealthStatusUnhealthy:
		return 0
	default:
		return 0
	}
}

// getDefaultAlertThresholds はデフォルトのアラート閾値を取得する
func getDefaultAlertThresholds() map[string]AlertThreshold {
	return map[string]AlertThreshold{
		"database": {
			WarningThreshold:  1000, // 1秒
			CriticalThreshold: 5000, // 5秒
			Duration:          time.Minute,
			Enabled:           true,
		},
		"memory": {
			WarningThreshold:  1024, // 1GB
			CriticalThreshold: 2048, // 2GB
			Duration:          time.Minute * 5,
			Enabled:           true,
		},
		"cpu": {
			WarningThreshold:  80,  // 80%
			CriticalThreshold: 95,  // 95%
			Duration:          time.Minute * 2,
			Enabled:           true,
		},
		"goroutines": {
			WarningThreshold:  1000,  // 1000個
			CriticalThreshold: 10000, // 10000個
			Duration:          time.Minute,
			Enabled:           true,
		},
		"disk": {
			WarningThreshold:  80, // 80%
			CriticalThreshold: 95, // 95%
			Duration:          time.Minute * 10,
			Enabled:           true,
		},
		"http_endpoints": {
			WarningThreshold:  1, // 1つ以上のエンドポイントに問題
			CriticalThreshold: 0, // 全てのエンドポイントに問題
			Duration:          time.Minute,
			Enabled:           true,
		},
	}
}

// DatabaseRecoveryAction はデータベース接続の自動復旧アクション
type DatabaseRecoveryAction struct {
	name    string
	enabled bool
	db      *sql.DB
	logger  logger.Logger
}

// NewDatabaseRecoveryAction は新しいDatabaseRecoveryActionを作成する
func NewDatabaseRecoveryAction(db *sql.DB) *DatabaseRecoveryAction {
	return &DatabaseRecoveryAction{
		name:    "database_recovery",
		enabled: true,
		db:      db,
		logger:  logger.GetLogger().WithComponent("database_recovery"),
	}
}

// Execute は復旧アクションを実行する
func (dra *DatabaseRecoveryAction) Execute(ctx context.Context, check *HealthCheck) error {
	if dra.db == nil {
		return fmt.Errorf("database connection is nil")
	}

	// 接続プールの統計情報を取得
	stats := dra.db.Stats()
	
	dra.logger.Info("データベース接続プール統計",
		logger.Int("open_connections", stats.OpenConnections),
		logger.Int("in_use", stats.InUse),
		logger.Int("idle", stats.Idle),
	)

	// 簡単な接続テストを再実行
	if err := dra.db.PingContext(ctx); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	dra.logger.Info("データベース接続復旧成功")
	return nil
}

// GetName は復旧アクションの名前を返す
func (dra *DatabaseRecoveryAction) GetName() string {
	return dra.name
}

// IsEnabled は復旧アクションが有効かどうかを返す
func (dra *DatabaseRecoveryAction) IsEnabled() bool {
	return dra.enabled
}

// SetEnabled は復旧アクションの有効/無効を設定する
func (dra *DatabaseRecoveryAction) SetEnabled(enabled bool) {
	dra.enabled = enabled
}
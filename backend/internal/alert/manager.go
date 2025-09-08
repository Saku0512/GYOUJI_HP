package alert

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend/internal/logger"
	"backend/internal/metrics"
)

// DefaultAlertManager はデフォルトのアラートマネージャー実装
type DefaultAlertManager struct {
	store         AlertStore
	notifiers     map[string]AlertNotifier
	rules         map[string]*AlertRule
	activeAlerts  map[string]*Alert
	silencedUntil map[string]time.Time
	mutex         sync.RWMutex
	logger        logger.Logger
	metrics       metrics.MetricsCollector
	
	// 評価関連
	evaluationInterval time.Duration
	stopChan          chan struct{}
	running           bool
}

// NewDefaultAlertManager は新しいDefaultAlertManagerを作成する
func NewDefaultAlertManager(store AlertStore, metricsCollector metrics.MetricsCollector) *DefaultAlertManager {
	return &DefaultAlertManager{
		store:              store,
		notifiers:          make(map[string]AlertNotifier),
		rules:              make(map[string]*AlertRule),
		activeAlerts:       make(map[string]*Alert),
		silencedUntil:      make(map[string]time.Time),
		logger:             logger.GetLogger().WithComponent("alert_manager"),
		metrics:            metricsCollector,
		evaluationInterval: 30 * time.Second, // デフォルト30秒間隔
		stopChan:          make(chan struct{}),
	}
}

// AddRule はアラートルールを追加する
func (am *DefaultAlertManager) AddRule(rule *AlertRule) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if err := am.store.SaveRule(rule); err != nil {
		return fmt.Errorf("failed to save rule: %w", err)
	}

	am.rules[rule.ID] = rule
	am.logger.Info("アラートルールを追加しました",
		logger.String("rule_id", rule.ID),
		logger.String("rule_name", rule.Name),
		logger.String("type", string(rule.Type)),
		logger.String("severity", string(rule.Severity)),
	)

	return nil
}

// RemoveRule はアラートルールを削除する
func (am *DefaultAlertManager) RemoveRule(ruleID string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if _, exists := am.rules[ruleID]; !exists {
		return fmt.Errorf("rule not found: %s", ruleID)
	}

	if err := am.store.DeleteRule(ruleID); err != nil {
		return fmt.Errorf("failed to delete rule: %w", err)
	}

	delete(am.rules, ruleID)
	am.logger.Info("アラートルールを削除しました",
		logger.String("rule_id", ruleID),
	)

	return nil
}

// GetRule はアラートルールを取得する
func (am *DefaultAlertManager) GetRule(ruleID string) (*AlertRule, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	rule, exists := am.rules[ruleID]
	if !exists {
		return am.store.GetRule(ruleID)
	}
	return rule, nil
}

// GetRules は全てのアラートルールを取得する
func (am *DefaultAlertManager) GetRules() []*AlertRule {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var rules []*AlertRule
	for _, rule := range am.rules {
		rules = append(rules, rule)
	}
	return rules
}

// UpdateRule はアラートルールを更新する
func (am *DefaultAlertManager) UpdateRule(rule *AlertRule) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if _, exists := am.rules[rule.ID]; !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}

	rule.UpdatedAt = time.Now()
	if err := am.store.UpdateRule(rule); err != nil {
		return fmt.Errorf("failed to update rule: %w", err)
	}

	am.rules[rule.ID] = rule
	am.logger.Info("アラートルールを更新しました",
		logger.String("rule_id", rule.ID),
		logger.String("rule_name", rule.Name),
	)

	return nil
}

// FireAlert はアラートを発火する
func (am *DefaultAlertManager) FireAlert(alert *Alert) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	// サイレンス中かチェック
	if silencedUntil, exists := am.silencedUntil[alert.ID]; exists {
		if time.Now().Before(silencedUntil) {
			am.logger.Debug("アラートはサイレンス中です",
				logger.String("alert_id", alert.ID),
				logger.Time("silenced_until", silencedUntil),
			)
			return nil
		}
		// サイレンス期間が終了した場合は削除
		delete(am.silencedUntil, alert.ID)
	}

	// アラートを保存
	if err := am.store.SaveAlert(alert); err != nil {
		return fmt.Errorf("failed to save alert: %w", err)
	}

	am.activeAlerts[alert.ID] = alert

	am.logger.Error("アラートが発火しました",
		logger.String("alert_id", alert.ID),
		logger.String("type", string(alert.Type)),
		logger.String("severity", string(alert.Severity)),
		logger.String("title", alert.Title),
		logger.String("description", alert.Description),
		logger.Float64("value", alert.Value),
		logger.Float64("threshold", alert.Threshold),
	)

	// メトリクスを記録
	am.recordAlertMetrics(alert, "fired")

	// 通知を送信
	go am.sendNotifications(context.Background(), alert)

	return nil
}

// ResolveAlert はアラートを解決する
func (am *DefaultAlertManager) ResolveAlert(alertID string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	alert, exists := am.activeAlerts[alertID]
	if !exists {
		return fmt.Errorf("active alert not found: %s", alertID)
	}

	// アラートの状態を更新
	now := time.Now()
	alert.Status = AlertStatusResolved
	alert.EndsAt = &now

	if err := am.store.UpdateAlert(alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	delete(am.activeAlerts, alertID)

	am.logger.Info("アラートが解決されました",
		logger.String("alert_id", alertID),
		logger.String("type", string(alert.Type)),
		logger.Duration("duration", now.Sub(alert.StartsAt)),
	)

	// メトリクスを記録
	am.recordAlertMetrics(alert, "resolved")

	return nil
}

// SilenceAlert はアラートをサイレンスする
func (am *DefaultAlertManager) SilenceAlert(alertID string, duration time.Duration) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	alert, exists := am.activeAlerts[alertID]
	if !exists {
		return fmt.Errorf("active alert not found: %s", alertID)
	}

	silenceUntil := time.Now().Add(duration)
	am.silencedUntil[alertID] = silenceUntil

	alert.Status = AlertStatusSilenced
	if err := am.store.UpdateAlert(alert); err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	am.logger.Info("アラートをサイレンスしました",
		logger.String("alert_id", alertID),
		logger.Duration("duration", duration),
		logger.Time("until", silenceUntil),
	)

	return nil
}

// GetAlert はアラートを取得する
func (am *DefaultAlertManager) GetAlert(alertID string) (*Alert, error) {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	if alert, exists := am.activeAlerts[alertID]; exists {
		return alert, nil
	}
	return am.store.GetAlert(alertID)
}

// GetAlerts は全てのアラートを取得する
func (am *DefaultAlertManager) GetAlerts() []*Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	alerts, _ := am.store.GetAlerts(nil)
	return alerts
}

// GetActiveAlerts はアクティブなアラートを取得する
func (am *DefaultAlertManager) GetActiveAlerts() []*Alert {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var alerts []*Alert
	for _, alert := range am.activeAlerts {
		alerts = append(alerts, alert)
	}
	return alerts
}

// AddNotifier は通知機能を追加する
func (am *DefaultAlertManager) AddNotifier(notifier AlertNotifier) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	am.notifiers[notifier.GetName()] = notifier
	am.logger.Info("通知機能を追加しました",
		logger.String("notifier", notifier.GetName()),
	)

	return nil
}

// RemoveNotifier は通知機能を削除する
func (am *DefaultAlertManager) RemoveNotifier(name string) error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if _, exists := am.notifiers[name]; !exists {
		return fmt.Errorf("notifier not found: %s", name)
	}

	delete(am.notifiers, name)
	am.logger.Info("通知機能を削除しました",
		logger.String("notifier", name),
	)

	return nil
}

// GetNotifiers は全ての通知機能を取得する
func (am *DefaultAlertManager) GetNotifiers() []AlertNotifier {
	am.mutex.RLock()
	defer am.mutex.RUnlock()

	var notifiers []AlertNotifier
	for _, notifier := range am.notifiers {
		notifiers = append(notifiers, notifier)
	}
	return notifiers
}

// EvaluateRules はアラートルールを評価する
func (am *DefaultAlertManager) EvaluateRules(ctx context.Context) error {
	am.mutex.RLock()
	rules := make([]*AlertRule, 0, len(am.rules))
	for _, rule := range am.rules {
		if rule.Enabled {
			rules = append(rules, rule)
		}
	}
	am.mutex.RUnlock()

	for _, rule := range rules {
		if err := am.evaluateRule(ctx, rule); err != nil {
			am.logger.Error("ルール評価エラー",
				logger.String("rule_id", rule.ID),
				logger.String("rule_name", rule.Name),
				logger.Err(err),
			)
		}
	}

	return nil
}

// Start はアラートマネージャーを開始する
func (am *DefaultAlertManager) Start(ctx context.Context) error {
	am.mutex.Lock()
	if am.running {
		am.mutex.Unlock()
		return fmt.Errorf("alert manager is already running")
	}
	am.running = true
	am.mutex.Unlock()

	// ストアからルールを読み込み
	if err := am.loadRulesFromStore(); err != nil {
		return fmt.Errorf("failed to load rules: %w", err)
	}

	am.logger.Info("アラートマネージャーを開始しました",
		logger.Duration("evaluation_interval", am.evaluationInterval),
	)

	// 評価ループを開始
	go am.evaluationLoop(ctx)

	return nil
}

// Stop はアラートマネージャーを停止する
func (am *DefaultAlertManager) Stop() error {
	am.mutex.Lock()
	defer am.mutex.Unlock()

	if !am.running {
		return fmt.Errorf("alert manager is not running")
	}

	close(am.stopChan)
	am.running = false

	am.logger.Info("アラートマネージャーを停止しました")
	return nil
}

// loadRulesFromStore はストアからルールを読み込む
func (am *DefaultAlertManager) loadRulesFromStore() error {
	rules, err := am.store.GetRules()
	if err != nil {
		return err
	}

	am.mutex.Lock()
	defer am.mutex.Unlock()

	for _, rule := range rules {
		am.rules[rule.ID] = rule
	}

	am.logger.Info("ルールを読み込みました",
		logger.Int("count", len(rules)),
	)

	return nil
}

// evaluationLoop は評価ループを実行する
func (am *DefaultAlertManager) evaluationLoop(ctx context.Context) {
	ticker := time.NewTicker(am.evaluationInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-am.stopChan:
			return
		case <-ticker.C:
			if err := am.EvaluateRules(ctx); err != nil {
				am.logger.Error("ルール評価エラー", logger.Err(err))
			}
		}
	}
}

// evaluateRule は個別のルールを評価する
func (am *DefaultAlertManager) evaluateRule(ctx context.Context, rule *AlertRule) error {
	// メトリクスから値を取得（簡易実装）
	// 実際の実装では、メトリクスクエリエンジンを使用
	value := am.getMetricValue(rule.Query)
	
	condition := &AlertCondition{
		MetricName: rule.Query,
		Operator:   rule.Operator,
		Threshold:  rule.Threshold,
		Duration:   rule.Duration,
	}

	if condition.Evaluate(value) {
		// アラート条件に一致した場合
		alertID := fmt.Sprintf("%s_%s", rule.ID, rule.Type)
		
		// 既にアクティブなアラートがあるかチェック
		if _, exists := am.activeAlerts[alertID]; !exists {
			alert := &Alert{
				ID:          alertID,
				Type:        rule.Type,
				Severity:    rule.Severity,
				Title:       fmt.Sprintf("Alert: %s", rule.Name),
				Description: rule.Description,
				Source:      "alert_manager",
				Timestamp:   time.Now(),
				Labels:      rule.Labels,
				Annotations: rule.Annotations,
				Value:       value,
				Threshold:   rule.Threshold,
				Status:      AlertStatusFiring,
				StartsAt:    time.Now(),
			}

			return am.FireAlert(alert)
		}
	} else {
		// アラート条件に一致しない場合、解決する
		alertID := fmt.Sprintf("%s_%s", rule.ID, rule.Type)
		if _, exists := am.activeAlerts[alertID]; exists {
			return am.ResolveAlert(alertID)
		}
	}

	return nil
}

// getMetricValue はメトリクス値を取得する（簡易実装）
func (am *DefaultAlertManager) getMetricValue(query string) float64 {
	// 実際の実装では、メトリクスコレクターから値を取得
	// ここでは簡易的にランダム値を返す
	return 0.0
}

// sendNotifications は通知を送信する
func (am *DefaultAlertManager) sendNotifications(ctx context.Context, alert *Alert) {
	am.mutex.RLock()
	notifiers := make([]AlertNotifier, 0, len(am.notifiers))
	for _, notifier := range am.notifiers {
		if notifier.IsEnabled() {
			notifiers = append(notifiers, notifier)
		}
	}
	am.mutex.RUnlock()

	for _, notifier := range notifiers {
		if err := notifier.Notify(ctx, alert); err != nil {
			am.logger.Error("通知送信エラー",
				logger.String("notifier", notifier.GetName()),
				logger.String("alert_id", alert.ID),
				logger.Err(err),
			)
		} else {
			am.logger.Info("通知を送信しました",
				logger.String("notifier", notifier.GetName()),
				logger.String("alert_id", alert.ID),
			)
		}
	}
}

// recordAlertMetrics はアラートメトリクスを記録する
func (am *DefaultAlertManager) recordAlertMetrics(alert *Alert, action string) {
	if am.metrics != nil {
		labels := metrics.Labels{
			"type":     string(alert.Type),
			"severity": string(alert.Severity),
			"action":   action,
		}
		
		counter := am.metrics.RegisterCounter(
			"alerts_total",
			"Total number of alerts",
			labels,
		)
		counter.Inc()
	}
}

// グローバルアラートマネージャー
var globalAlertManager AlertManager

// Init はグローバルアラートマネージャーを初期化する
func Init(store AlertStore, metricsCollector metrics.MetricsCollector) {
	globalAlertManager = NewDefaultAlertManager(store, metricsCollector)
}

// GetManager はグローバルアラートマネージャーを取得する
func GetManager() AlertManager {
	if globalAlertManager == nil {
		Init(NewMemoryAlertStore(), metrics.GetCollector())
	}
	return globalAlertManager
}
package alert

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"strconv"
	"time"

	"backend/internal/logger"
	"backend/internal/metrics"
)

// AlertSystemConfig はアラートシステムの設定
type AlertSystemConfig struct {
	// アラートマネージャー設定
	EvaluationInterval time.Duration `json:"evaluation_interval"`
	
	// ヘルスモニター設定
	HealthCheckInterval time.Duration `json:"health_check_interval"`
	EnableHealthMonitor bool          `json:"enable_health_monitor"`
	
	// 通知設定
	EnableNotifications bool     `json:"enable_notifications"`
	NotifierTypes      []string `json:"notifier_types"`
	
	// ストレージ設定
	UseMemoryStore bool   `json:"use_memory_store"`
	DatabaseURL    string `json:"database_url"`
	
	// 自動復旧設定
	EnableAutoRecovery bool `json:"enable_auto_recovery"`
}

// AlertSystem はアラートシステム全体を管理する
type AlertSystem struct {
	config        AlertSystemConfig
	manager       AlertManager
	healthMonitor *HealthMonitor
	store         AlertStore
	notifiers     []AlertNotifier
	logger        logger.Logger
	
	// 依存関係
	db      *sql.DB
	metrics metrics.MetricsCollector
}

// NewAlertSystem は新しいアラートシステムを作成する
func NewAlertSystem(config AlertSystemConfig, db *sql.DB, metricsCollector metrics.MetricsCollector) *AlertSystem {
	return &AlertSystem{
		config:  config,
		db:      db,
		metrics: metricsCollector,
		logger:  logger.GetLogger().WithComponent("alert_system"),
	}
}

// Initialize はアラートシステムを初期化する
func (as *AlertSystem) Initialize() error {
	as.logger.Info("アラートシステムを初期化しています...")

	// ストレージを初期化
	if err := as.initializeStore(); err != nil {
		return fmt.Errorf("failed to initialize alert store: %w", err)
	}

	// アラートマネージャーを初期化
	if err := as.initializeManager(); err != nil {
		return fmt.Errorf("failed to initialize alert manager: %w", err)
	}

	// 通知機能を初期化
	if as.config.EnableNotifications {
		if err := as.initializeNotifiers(); err != nil {
			return fmt.Errorf("failed to initialize notifiers: %w", err)
		}
	}

	// ヘルスモニターを初期化
	if as.config.EnableHealthMonitor {
		if err := as.initializeHealthMonitor(); err != nil {
			return fmt.Errorf("failed to initialize health monitor: %w", err)
		}
	}

	// デフォルトのアラートルールを追加
	if err := as.addDefaultAlertRules(); err != nil {
		return fmt.Errorf("failed to add default alert rules: %w", err)
	}

	as.logger.Info("アラートシステムの初期化が完了しました")
	return nil
}

// Start はアラートシステムを開始する
func (as *AlertSystem) Start(ctx context.Context) error {
	as.logger.Info("アラートシステムを開始しています...")

	// アラートマネージャーを開始
	if err := as.manager.Start(ctx); err != nil {
		return fmt.Errorf("failed to start alert manager: %w", err)
	}

	// ヘルスモニターを開始
	if as.healthMonitor != nil {
		if err := as.healthMonitor.Start(ctx); err != nil {
			return fmt.Errorf("failed to start health monitor: %w", err)
		}
	}

	as.logger.Info("アラートシステムが開始されました")
	return nil
}

// Stop はアラートシステムを停止する
func (as *AlertSystem) Stop() error {
	as.logger.Info("アラートシステムを停止しています...")

	// ヘルスモニターを停止
	if as.healthMonitor != nil {
		if err := as.healthMonitor.Stop(); err != nil {
			as.logger.Error("ヘルスモニター停止エラー", logger.Err(err))
		}
	}

	// アラートマネージャーを停止
	if err := as.manager.Stop(); err != nil {
		as.logger.Error("アラートマネージャー停止エラー", logger.Err(err))
	}

	as.logger.Info("アラートシステムが停止されました")
	return nil
}

// GetManager はアラートマネージャーを取得する
func (as *AlertSystem) GetManager() AlertManager {
	return as.manager
}

// GetHealthMonitor はヘルスモニターを取得する
func (as *AlertSystem) GetHealthMonitor() *HealthMonitor {
	return as.healthMonitor
}

// initializeStore はアラートストレージを初期化する
func (as *AlertSystem) initializeStore() error {
	if as.config.UseMemoryStore {
		as.store = NewMemoryAlertStore()
		as.logger.Info("メモリアラートストアを初期化しました")
	} else {
		// 実際の実装では、データベースストアを使用
		as.store = NewMemoryAlertStore()
		as.logger.Info("データベースアラートストア（未実装のためメモリストア）を初期化しました")
	}
	return nil
}

// initializeManager はアラートマネージャーを初期化する
func (as *AlertSystem) initializeManager() error {
	as.manager = NewDefaultAlertManager(as.store, as.metrics)
	as.logger.Info("アラートマネージャーを初期化しました")
	return nil
}

// initializeNotifiers は通知機能を初期化する
func (as *AlertSystem) initializeNotifiers() error {
	// 環境変数から通知機能を読み込み
	notifiers := LoadNotifiersFromEnv()
	
	// 設定された通知機能を追加
	for _, notifier := range notifiers {
		if err := as.manager.AddNotifier(notifier); err != nil {
			as.logger.Error("通知機能追加エラー",
				logger.String("notifier", notifier.GetName()),
				logger.Err(err),
			)
		} else {
			as.logger.Info("通知機能を追加しました",
				logger.String("notifier", notifier.GetName()),
			)
		}
	}

	as.notifiers = notifiers
	return nil
}

// initializeHealthMonitor はヘルスモニターを初期化する
func (as *AlertSystem) initializeHealthMonitor() error {
	config := HealthMonitorConfig{
		CheckInterval:   as.config.HealthCheckInterval,
		AlertThresholds: getDefaultAlertThresholds(),
		RecoveryActions: make(map[string]RecoveryAction),
	}

	// 自動復旧が有効な場合は復旧アクションを追加
	if as.config.EnableAutoRecovery && as.db != nil {
		config.RecoveryActions["database"] = NewDatabaseRecoveryAction(as.db)
	}

	as.healthMonitor = NewHealthMonitor(config, as.manager, as.metrics, as.db)
	as.logger.Info("ヘルスモニターを初期化しました")
	return nil
}

// addDefaultAlertRules はデフォルトのアラートルールを追加する
func (as *AlertSystem) addDefaultAlertRules() error {
	defaultRules := []*AlertRule{
		{
			ID:          "default_high_error_rate",
			Name:        "高エラー率",
			Type:        AlertTypeHighErrorRate,
			Severity:    SeverityWarning,
			Description: "HTTPエラー率が高い状態が続いています",
			Query:       "http_error_rate",
			Threshold:   5.0, // 5%
			Operator:    OperatorGreaterThan,
			Duration:    5 * time.Minute,
			Enabled:     true,
			Labels: map[string]string{
				"category": "http",
				"default":  "true",
			},
			Annotations: map[string]string{
				"summary":     "HTTPエラー率が閾値を超えています",
				"description": "過去5分間のHTTPエラー率が5%を超えています",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "default_slow_response",
			Name:        "レスポンス時間遅延",
			Type:        AlertTypeSlowResponse,
			Severity:    SeverityWarning,
			Description: "APIレスポンス時間が遅い状態が続いています",
			Query:       "http_response_time_p95",
			Threshold:   2.0, // 2秒
			Operator:    OperatorGreaterThan,
			Duration:    3 * time.Minute,
			Enabled:     true,
			Labels: map[string]string{
				"category": "performance",
				"default":  "true",
			},
			Annotations: map[string]string{
				"summary":     "APIレスポンス時間が遅延しています",
				"description": "95パーセンタイルのレスポンス時間が2秒を超えています",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "default_db_connection_error",
			Name:        "データベース接続エラー",
			Type:        AlertTypeDBConnectionError,
			Severity:    SeverityCritical,
			Description: "データベース接続でエラーが発生しています",
			Query:       "db_connection_errors",
			Threshold:   1.0,
			Operator:    OperatorGreaterThan,
			Duration:    1 * time.Minute,
			Enabled:     true,
			Labels: map[string]string{
				"category": "database",
				"default":  "true",
			},
			Annotations: map[string]string{
				"summary":     "データベース接続エラーが発生しています",
				"description": "データベースへの接続でエラーが発生しています",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "default_high_memory_usage",
			Name:        "高メモリ使用量",
			Type:        AlertTypeHighMemory,
			Severity:    SeverityWarning,
			Description: "メモリ使用量が高い状態が続いています",
			Query:       "memory_usage_percent",
			Threshold:   80.0, // 80%
			Operator:    OperatorGreaterThan,
			Duration:    5 * time.Minute,
			Enabled:     true,
			Labels: map[string]string{
				"category": "system",
				"default":  "true",
			},
			Annotations: map[string]string{
				"summary":     "メモリ使用量が高くなっています",
				"description": "システムのメモリ使用量が80%を超えています",
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, rule := range defaultRules {
		if err := as.manager.AddRule(rule); err != nil {
			as.logger.Error("デフォルトアラートルール追加エラー",
				logger.String("rule_id", rule.ID),
				logger.String("rule_name", rule.Name),
				logger.Err(err),
			)
		} else {
			as.logger.Info("デフォルトアラートルールを追加しました",
				logger.String("rule_id", rule.ID),
				logger.String("rule_name", rule.Name),
			)
		}
	}

	return nil
}

// LoadConfigFromEnv は環境変数から設定を読み込む
func LoadConfigFromEnv() AlertSystemConfig {
	config := AlertSystemConfig{
		// デフォルト値
		EvaluationInterval:  30 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		EnableHealthMonitor: true,
		EnableNotifications: true,
		UseMemoryStore:      true,
		EnableAutoRecovery:  true,
	}

	// 環境変数から設定を読み込み
	if interval := os.Getenv("ALERT_EVALUATION_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			config.EvaluationInterval = d
		}
	}

	if interval := os.Getenv("ALERT_HEALTH_CHECK_INTERVAL"); interval != "" {
		if d, err := time.ParseDuration(interval); err == nil {
			config.HealthCheckInterval = d
		}
	}

	if enable := os.Getenv("ALERT_ENABLE_HEALTH_MONITOR"); enable != "" {
		if b, err := strconv.ParseBool(enable); err == nil {
			config.EnableHealthMonitor = b
		}
	}

	if enable := os.Getenv("ALERT_ENABLE_NOTIFICATIONS"); enable != "" {
		if b, err := strconv.ParseBool(enable); err == nil {
			config.EnableNotifications = b
		}
	}

	if enable := os.Getenv("ALERT_USE_MEMORY_STORE"); enable != "" {
		if b, err := strconv.ParseBool(enable); err == nil {
			config.UseMemoryStore = b
		}
	}

	if enable := os.Getenv("ALERT_ENABLE_AUTO_RECOVERY"); enable != "" {
		if b, err := strconv.ParseBool(enable); err == nil {
			config.EnableAutoRecovery = b
		}
	}

	if dbURL := os.Getenv("ALERT_DATABASE_URL"); dbURL != "" {
		config.DatabaseURL = dbURL
	}

	return config
}

// InitializeGlobalAlertSystem はグローバルアラートシステムを初期化する
func InitializeGlobalAlertSystem(db *sql.DB, metricsCollector metrics.MetricsCollector) (*AlertSystem, error) {
	config := LoadConfigFromEnv()
	
	alertSystem := NewAlertSystem(config, db, metricsCollector)
	
	if err := alertSystem.Initialize(); err != nil {
		return nil, fmt.Errorf("failed to initialize alert system: %w", err)
	}

	// グローバル変数を設定
	globalAlertManager = alertSystem.GetManager()
	
	return alertSystem, nil
}

// StartGlobalAlertSystem はグローバルアラートシステムを開始する
func StartGlobalAlertSystem(ctx context.Context, alertSystem *AlertSystem) error {
	return alertSystem.Start(ctx)
}

// StopGlobalAlertSystem はグローバルアラートシステムを停止する
func StopGlobalAlertSystem(alertSystem *AlertSystem) error {
	return alertSystem.Stop()
}
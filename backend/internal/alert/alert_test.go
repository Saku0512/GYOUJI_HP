package alert

import (
	"context"
	"testing"
	"time"

	"backend/internal/metrics"
)

func TestAlertCreation(t *testing.T) {
	alert := NewAlert(
		AlertTypeHighCPU,
		SeverityWarning,
		"高CPU使用率",
		"CPU使用率が80%を超えています",
		"system_monitor",
	)

	if alert.Type != AlertTypeHighCPU {
		t.Errorf("Expected alert type %s, got %s", AlertTypeHighCPU, alert.Type)
	}

	if alert.Severity != SeverityWarning {
		t.Errorf("Expected severity %s, got %s", SeverityWarning, alert.Severity)
	}

	if alert.Status != AlertStatusFiring {
		t.Errorf("Expected status %s, got %s", AlertStatusFiring, alert.Status)
	}

	if alert.ID == "" {
		t.Error("Alert ID should not be empty")
	}
}

func TestAlertRuleCreation(t *testing.T) {
	rule := NewAlertRule(
		"高CPU使用率ルール",
		AlertTypeHighCPU,
		SeverityWarning,
	)

	if rule.Name != "高CPU使用率ルール" {
		t.Errorf("Expected rule name '高CPU使用率ルール', got '%s'", rule.Name)
	}

	if rule.Type != AlertTypeHighCPU {
		t.Errorf("Expected rule type %s, got %s", AlertTypeHighCPU, rule.Type)
	}

	if !rule.Enabled {
		t.Error("Rule should be enabled by default")
	}

	if rule.ID == "" {
		t.Error("Rule ID should not be empty")
	}
}

func TestAlertConditionEvaluation(t *testing.T) {
	tests := []struct {
		name      string
		condition AlertCondition
		value     float64
		expected  bool
	}{
		{
			name: "Greater than - true",
			condition: AlertCondition{
				Operator:  OperatorGreaterThan,
				Threshold: 80.0,
			},
			value:    85.0,
			expected: true,
		},
		{
			name: "Greater than - false",
			condition: AlertCondition{
				Operator:  OperatorGreaterThan,
				Threshold: 80.0,
			},
			value:    75.0,
			expected: false,
		},
		{
			name: "Less than - true",
			condition: AlertCondition{
				Operator:  OperatorLessThan,
				Threshold: 20.0,
			},
			value:    15.0,
			expected: true,
		},
		{
			name: "Equal - true",
			condition: AlertCondition{
				Operator:  OperatorEqual,
				Threshold: 50.0,
			},
			value:    50.0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.condition.Evaluate(tt.value)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestMemoryAlertStore(t *testing.T) {
	store := NewMemoryAlertStore()

	// アラートを保存
	alert := NewAlert(
		AlertTypeHighCPU,
		SeverityWarning,
		"テストアラート",
		"テスト用のアラートです",
		"test",
	)

	err := store.SaveAlert(alert)
	if err != nil {
		t.Fatalf("Failed to save alert: %v", err)
	}

	// アラートを取得
	retrieved, err := store.GetAlert(alert.ID)
	if err != nil {
		t.Fatalf("Failed to get alert: %v", err)
	}

	if retrieved.ID != alert.ID {
		t.Errorf("Expected alert ID %s, got %s", alert.ID, retrieved.ID)
	}

	// アラートを更新
	alert.Status = AlertStatusResolved
	err = store.UpdateAlert(alert)
	if err != nil {
		t.Fatalf("Failed to update alert: %v", err)
	}

	// 更新されたアラートを取得
	updated, err := store.GetAlert(alert.ID)
	if err != nil {
		t.Fatalf("Failed to get updated alert: %v", err)
	}

	if updated.Status != AlertStatusResolved {
		t.Errorf("Expected status %s, got %s", AlertStatusResolved, updated.Status)
	}

	// フィルターでアラートを取得
	filters := map[string]interface{}{
		"status": string(AlertStatusResolved),
	}
	alerts, err := store.GetAlerts(filters)
	if err != nil {
		t.Fatalf("Failed to get alerts with filters: %v", err)
	}

	if len(alerts) != 1 {
		t.Errorf("Expected 1 alert, got %d", len(alerts))
	}

	// アラートを削除
	err = store.DeleteAlert(alert.ID)
	if err != nil {
		t.Fatalf("Failed to delete alert: %v", err)
	}

	// 削除されたアラートの取得を試行
	_, err = store.GetAlert(alert.ID)
	if err == nil {
		t.Error("Expected error when getting deleted alert")
	}
}

func TestDefaultAlertManager(t *testing.T) {
	store := NewMemoryAlertStore()
	collector := metrics.NewDefaultCollector()
	manager := NewDefaultAlertManager(store, collector)

	// ルールを追加
	rule := NewAlertRule(
		"テストルール",
		AlertTypeHighCPU,
		SeverityWarning,
	)
	rule.Threshold = 80.0
	rule.Operator = OperatorGreaterThan

	err := manager.AddRule(rule)
	if err != nil {
		t.Fatalf("Failed to add rule: %v", err)
	}

	// ルールを取得
	retrieved, err := manager.GetRule(rule.ID)
	if err != nil {
		t.Fatalf("Failed to get rule: %v", err)
	}

	if retrieved.ID != rule.ID {
		t.Errorf("Expected rule ID %s, got %s", rule.ID, retrieved.ID)
	}

	// アラートを発火
	alert := NewAlert(
		AlertTypeHighCPU,
		SeverityWarning,
		"テストアラート",
		"テスト用のアラートです",
		"test",
	)

	err = manager.FireAlert(alert)
	if err != nil {
		t.Fatalf("Failed to fire alert: %v", err)
	}

	// アクティブアラートを取得
	activeAlerts := manager.GetActiveAlerts()
	if len(activeAlerts) != 1 {
		t.Errorf("Expected 1 active alert, got %d", len(activeAlerts))
	}

	// アラートを解決
	err = manager.ResolveAlert(alert.ID)
	if err != nil {
		t.Fatalf("Failed to resolve alert: %v", err)
	}

	// アクティブアラートが0になることを確認
	activeAlerts = manager.GetActiveAlerts()
	if len(activeAlerts) != 0 {
		t.Errorf("Expected 0 active alerts, got %d", len(activeAlerts))
	}

	// ルールを削除
	err = manager.RemoveRule(rule.ID)
	if err != nil {
		t.Fatalf("Failed to remove rule: %v", err)
	}

	// 削除されたルールの取得を試行
	_, err = manager.GetRule(rule.ID)
	if err == nil {
		t.Error("Expected error when getting deleted rule")
	}
}

func TestLogNotifier(t *testing.T) {
	notifier := NewLogNotifier()

	if notifier.GetName() != "log" {
		t.Errorf("Expected notifier name 'log', got '%s'", notifier.GetName())
	}

	if !notifier.IsEnabled() {
		t.Error("Log notifier should be enabled by default")
	}

	// アラート通知をテスト
	alert := NewAlert(
		AlertTypeHighCPU,
		SeverityWarning,
		"テストアラート",
		"テスト用のアラートです",
		"test",
	)

	ctx := context.Background()
	err := notifier.Notify(ctx, alert)
	if err != nil {
		t.Fatalf("Failed to notify alert: %v", err)
	}
}

func TestWebhookNotifier(t *testing.T) {
	config := WebhookConfig{
		URL:     "http://example.com/webhook",
		Timeout: 5 * time.Second,
		Headers: map[string]string{
			"Authorization": "Bearer token",
		},
	}

	notifier := NewWebhookNotifier(config)

	if notifier.GetName() != "webhook" {
		t.Errorf("Expected notifier name 'webhook', got '%s'", notifier.GetName())
	}

	if !notifier.IsEnabled() {
		t.Error("Webhook notifier should be enabled by default")
	}

	// 無効なURLでのテスト（実際のHTTPリクエストは送信されない）
	notifier.SetEnabled(false)
	
	alert := NewAlert(
		AlertTypeHighCPU,
		SeverityWarning,
		"テストアラート",
		"テスト用のアラートです",
		"test",
	)

	ctx := context.Background()
	err := notifier.Notify(ctx, alert)
	if err != nil {
		t.Fatalf("Disabled notifier should not return error: %v", err)
	}
}

func TestSlackNotifier(t *testing.T) {
	config := SlackConfig{
		WebhookURL: "http://example.com/slack",
		Channel:    "#alerts",
		Username:   "Alert Bot",
	}

	notifier := NewSlackNotifier(config)

	if notifier.GetName() != "slack" {
		t.Errorf("Expected notifier name 'slack', got '%s'", notifier.GetName())
	}

	// 色の判定をテスト
	color := notifier.getColorBySeverity(SeverityCritical)
	if color != "danger" {
		t.Errorf("Expected color 'danger' for critical severity, got '%s'", color)
	}

	color = notifier.getColorBySeverity(SeverityWarning)
	if color != "warning" {
		t.Errorf("Expected color 'warning' for warning severity, got '%s'", color)
	}

	color = notifier.getColorBySeverity(SeverityInfo)
	if color != "good" {
		t.Errorf("Expected color 'good' for info severity, got '%s'", color)
	}
}

func TestNotifierFactory(t *testing.T) {
	factory := &NotifierFactory{}

	// ログ通知機能の作成
	logNotifier, err := factory.CreateNotifier("log", nil)
	if err != nil {
		t.Fatalf("Failed to create log notifier: %v", err)
	}

	if logNotifier.GetName() != "log" {
		t.Errorf("Expected log notifier, got %s", logNotifier.GetName())
	}

	// Webhook通知機能の作成
	webhookConfig := map[string]interface{}{
		"url":     "http://example.com/webhook",
		"timeout": "10s",
	}

	webhookNotifier, err := factory.CreateNotifier("webhook", webhookConfig)
	if err != nil {
		t.Fatalf("Failed to create webhook notifier: %v", err)
	}

	if webhookNotifier.GetName() != "webhook" {
		t.Errorf("Expected webhook notifier, got %s", webhookNotifier.GetName())
	}

	// 不明な通知機能タイプ
	_, err = factory.CreateNotifier("unknown", nil)
	if err == nil {
		t.Error("Expected error for unknown notifier type")
	}
}

func TestAlertSystemConfig(t *testing.T) {
	config := LoadConfigFromEnv()

	// デフォルト値のテスト
	if config.EvaluationInterval != 30*time.Second {
		t.Errorf("Expected default evaluation interval 30s, got %v", config.EvaluationInterval)
	}

	if config.HealthCheckInterval != 30*time.Second {
		t.Errorf("Expected default health check interval 30s, got %v", config.HealthCheckInterval)
	}

	if !config.EnableHealthMonitor {
		t.Error("Health monitor should be enabled by default")
	}

	if !config.EnableNotifications {
		t.Error("Notifications should be enabled by default")
	}

	if !config.UseMemoryStore {
		t.Error("Memory store should be used by default")
	}

	if !config.EnableAutoRecovery {
		t.Error("Auto recovery should be enabled by default")
	}
}

// ベンチマークテスト

func BenchmarkAlertCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewAlert(
			AlertTypeHighCPU,
			SeverityWarning,
			"ベンチマークアラート",
			"ベンチマーク用のアラートです",
			"benchmark",
		)
	}
}

func BenchmarkMemoryStoreOperations(b *testing.B) {
	store := NewMemoryAlertStore()
	alert := NewAlert(
		AlertTypeHighCPU,
		SeverityWarning,
		"ベンチマークアラート",
		"ベンチマーク用のアラートです",
		"benchmark",
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		store.SaveAlert(alert)
		store.GetAlert(alert.ID)
		store.UpdateAlert(alert)
	}
}

func BenchmarkAlertManagerOperations(b *testing.B) {
	store := NewMemoryAlertStore()
	collector := metrics.NewDefaultCollector()
	manager := NewDefaultAlertManager(store, collector)

	rule := NewAlertRule(
		"ベンチマークルール",
		AlertTypeHighCPU,
		SeverityWarning,
	)

	alert := NewAlert(
		AlertTypeHighCPU,
		SeverityWarning,
		"ベンチマークアラート",
		"ベンチマーク用のアラートです",
		"benchmark",
	)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		manager.AddRule(rule)
		manager.FireAlert(alert)
		manager.ResolveAlert(alert.ID)
		manager.RemoveRule(rule.ID)
	}
}
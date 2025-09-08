package alert

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Severity はアラートの重要度を表す
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// AlertType はアラートの種類を表す
type AlertType string

const (
	// システム関連アラート
	AlertTypeHighCPU           AlertType = "high_cpu"
	AlertTypeHighMemory        AlertType = "high_memory"
	AlertTypeHighDiskUsage     AlertType = "high_disk_usage"
	AlertTypeHighErrorRate     AlertType = "high_error_rate"
	AlertTypeSlowResponse      AlertType = "slow_response"
	AlertTypeHighLatency       AlertType = "high_latency"

	// データベース関連アラート
	AlertTypeDBConnectionError AlertType = "db_connection_error"
	AlertTypeDBSlowQuery       AlertType = "db_slow_query"
	AlertTypeDBHighConnections AlertType = "db_high_connections"

	// アプリケーション関連アラート
	AlertTypeAuthFailure       AlertType = "auth_failure"
	AlertTypeAPIRateLimit      AlertType = "api_rate_limit"
	AlertTypeWebSocketError    AlertType = "websocket_error"
	AlertTypeBusinessLogicError AlertType = "business_logic_error"

	// インフラ関連アラート
	AlertTypeServiceDown       AlertType = "service_down"
	AlertTypeHealthCheckFailed AlertType = "health_check_failed"
	AlertTypeExternalAPIError  AlertType = "external_api_error"
)

// Alert はアラート情報を表す構造体
type Alert struct {
	ID          string                 `json:"id"`
	Type        AlertType              `json:"type"`
	Severity    Severity               `json:"severity"`
	Title       string                 `json:"title"`
	Description string                 `json:"description"`
	Source      string                 `json:"source"`
	Timestamp   time.Time              `json:"timestamp"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Value       float64                `json:"value"`
	Threshold   float64                `json:"threshold"`
	Status      AlertStatus            `json:"status"`
	StartsAt    time.Time              `json:"starts_at"`
	EndsAt      *time.Time             `json:"ends_at,omitempty"`
	Tags        map[string]interface{} `json:"tags,omitempty"`
}

// AlertStatus はアラートの状態を表す
type AlertStatus string

const (
	AlertStatusFiring   AlertStatus = "firing"
	AlertStatusResolved AlertStatus = "resolved"
	AlertStatusSilenced AlertStatus = "silenced"
)

// AlertRule はアラートルールを表す構造体
type AlertRule struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Type        AlertType              `json:"type"`
	Severity    Severity               `json:"severity"`
	Description string                 `json:"description"`
	Query       string                 `json:"query"`
	Threshold   float64                `json:"threshold"`
	Operator    ComparisonOperator     `json:"operator"`
	Duration    time.Duration          `json:"duration"`
	Labels      map[string]string      `json:"labels"`
	Annotations map[string]string      `json:"annotations"`
	Enabled     bool                   `json:"enabled"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Tags        map[string]interface{} `json:"tags,omitempty"`
}

// ComparisonOperator は比較演算子を表す
type ComparisonOperator string

const (
	OperatorGreaterThan    ComparisonOperator = "gt"
	OperatorGreaterOrEqual ComparisonOperator = "gte"
	OperatorLessThan       ComparisonOperator = "lt"
	OperatorLessOrEqual    ComparisonOperator = "lte"
	OperatorEqual          ComparisonOperator = "eq"
	OperatorNotEqual       ComparisonOperator = "ne"
)

// AlertCondition はアラート条件を表す構造体
type AlertCondition struct {
	MetricName string             `json:"metric_name"`
	Labels     map[string]string  `json:"labels"`
	Operator   ComparisonOperator `json:"operator"`
	Threshold  float64            `json:"threshold"`
	Duration   time.Duration      `json:"duration"`
}

// Evaluate は条件を評価する
func (ac *AlertCondition) Evaluate(value float64) bool {
	switch ac.Operator {
	case OperatorGreaterThan:
		return value > ac.Threshold
	case OperatorGreaterOrEqual:
		return value >= ac.Threshold
	case OperatorLessThan:
		return value < ac.Threshold
	case OperatorLessOrEqual:
		return value <= ac.Threshold
	case OperatorEqual:
		return value == ac.Threshold
	case OperatorNotEqual:
		return value != ac.Threshold
	default:
		return false
	}
}

// AlertNotifier はアラート通知のインターフェース
type AlertNotifier interface {
	Notify(ctx context.Context, alert *Alert) error
	GetName() string
	IsEnabled() bool
}

// AlertManager はアラート管理のインターフェース
type AlertManager interface {
	// ルール管理
	AddRule(rule *AlertRule) error
	RemoveRule(ruleID string) error
	GetRule(ruleID string) (*AlertRule, error)
	GetRules() []*AlertRule
	UpdateRule(rule *AlertRule) error

	// アラート管理
	FireAlert(alert *Alert) error
	ResolveAlert(alertID string) error
	SilenceAlert(alertID string, duration time.Duration) error
	GetAlert(alertID string) (*Alert, error)
	GetAlerts() []*Alert
	GetActiveAlerts() []*Alert

	// 通知管理
	AddNotifier(notifier AlertNotifier) error
	RemoveNotifier(name string) error
	GetNotifiers() []AlertNotifier

	// 評価とモニタリング
	EvaluateRules(ctx context.Context) error
	Start(ctx context.Context) error
	Stop() error
}

// AlertStore はアラートの永続化インターフェース
type AlertStore interface {
	SaveAlert(alert *Alert) error
	GetAlert(id string) (*Alert, error)
	GetAlerts(filters map[string]interface{}) ([]*Alert, error)
	UpdateAlert(alert *Alert) error
	DeleteAlert(id string) error

	SaveRule(rule *AlertRule) error
	GetRule(id string) (*AlertRule, error)
	GetRules() ([]*AlertRule, error)
	UpdateRule(rule *AlertRule) error
	DeleteRule(id string) error
}

// MemoryAlertStore はメモリ内アラートストア
type MemoryAlertStore struct {
	alerts map[string]*Alert
	rules  map[string]*AlertRule
	mutex  sync.RWMutex
}

// NewMemoryAlertStore は新しいMemoryAlertStoreを作成する
func NewMemoryAlertStore() *MemoryAlertStore {
	return &MemoryAlertStore{
		alerts: make(map[string]*Alert),
		rules:  make(map[string]*AlertRule),
	}
}

// SaveAlert はアラートを保存する
func (s *MemoryAlertStore) SaveAlert(alert *Alert) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.alerts[alert.ID] = alert
	return nil
}

// GetAlert はアラートを取得する
func (s *MemoryAlertStore) GetAlert(id string) (*Alert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	alert, exists := s.alerts[id]
	if !exists {
		return nil, fmt.Errorf("alert not found: %s", id)
	}
	return alert, nil
}

// GetAlerts はフィルターに基づいてアラートを取得する
func (s *MemoryAlertStore) GetAlerts(filters map[string]interface{}) ([]*Alert, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var alerts []*Alert
	for _, alert := range s.alerts {
		if s.matchesFilters(alert, filters) {
			alerts = append(alerts, alert)
		}
	}
	return alerts, nil
}

// UpdateAlert はアラートを更新する
func (s *MemoryAlertStore) UpdateAlert(alert *Alert) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.alerts[alert.ID]; !exists {
		return fmt.Errorf("alert not found: %s", alert.ID)
	}
	s.alerts[alert.ID] = alert
	return nil
}

// DeleteAlert はアラートを削除する
func (s *MemoryAlertStore) DeleteAlert(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.alerts[id]; !exists {
		return fmt.Errorf("alert not found: %s", id)
	}
	delete(s.alerts, id)
	return nil
}

// SaveRule はルールを保存する
func (s *MemoryAlertStore) SaveRule(rule *AlertRule) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.rules[rule.ID] = rule
	return nil
}

// GetRule はルールを取得する
func (s *MemoryAlertStore) GetRule(id string) (*AlertRule, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	
	rule, exists := s.rules[id]
	if !exists {
		return nil, fmt.Errorf("rule not found: %s", id)
	}
	return rule, nil
}

// GetRules は全てのルールを取得する
func (s *MemoryAlertStore) GetRules() ([]*AlertRule, error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var rules []*AlertRule
	for _, rule := range s.rules {
		rules = append(rules, rule)
	}
	return rules, nil
}

// UpdateRule はルールを更新する
func (s *MemoryAlertStore) UpdateRule(rule *AlertRule) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.rules[rule.ID]; !exists {
		return fmt.Errorf("rule not found: %s", rule.ID)
	}
	rule.UpdatedAt = time.Now()
	s.rules[rule.ID] = rule
	return nil
}

// DeleteRule はルールを削除する
func (s *MemoryAlertStore) DeleteRule(id string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	
	if _, exists := s.rules[id]; !exists {
		return fmt.Errorf("rule not found: %s", id)
	}
	delete(s.rules, id)
	return nil
}

// matchesFilters はアラートがフィルターに一致するかチェックする
func (s *MemoryAlertStore) matchesFilters(alert *Alert, filters map[string]interface{}) bool {
	for key, value := range filters {
		switch key {
		case "status":
			if alert.Status != AlertStatus(value.(string)) {
				return false
			}
		case "severity":
			if alert.Severity != Severity(value.(string)) {
				return false
			}
		case "type":
			if alert.Type != AlertType(value.(string)) {
				return false
			}
		case "source":
			if alert.Source != value.(string) {
				return false
			}
		}
	}
	return true
}

// NewAlert は新しいアラートを作成する
func NewAlert(alertType AlertType, severity Severity, title, description, source string) *Alert {
	return &Alert{
		ID:          generateAlertID(),
		Type:        alertType,
		Severity:    severity,
		Title:       title,
		Description: description,
		Source:      source,
		Timestamp:   time.Now(),
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		Status:      AlertStatusFiring,
		StartsAt:    time.Now(),
		Tags:        make(map[string]interface{}),
	}
}

// NewAlertRule は新しいアラートルールを作成する
func NewAlertRule(name string, alertType AlertType, severity Severity) *AlertRule {
	return &AlertRule{
		ID:          generateRuleID(),
		Name:        name,
		Type:        alertType,
		Severity:    severity,
		Labels:      make(map[string]string),
		Annotations: make(map[string]string),
		Enabled:     true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
		Tags:        make(map[string]interface{}),
	}
}

// generateAlertID はアラートIDを生成する
func generateAlertID() string {
	return fmt.Sprintf("alert_%d", time.Now().UnixNano())
}

// generateRuleID はルールIDを生成する
func generateRuleID() string {
	return fmt.Sprintf("rule_%d", time.Now().UnixNano())
}
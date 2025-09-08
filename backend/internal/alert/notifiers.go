package alert

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"backend/internal/logger"
)

// LogNotifier はログ出力による通知機能
type LogNotifier struct {
	name    string
	enabled bool
	logger  logger.Logger
}

// NewLogNotifier は新しいLogNotifierを作成する
func NewLogNotifier() *LogNotifier {
	return &LogNotifier{
		name:    "log",
		enabled: true,
		logger:  logger.GetLogger().WithComponent("alert_notifier_log"),
	}
}

// Notify はアラートをログに出力する
func (n *LogNotifier) Notify(ctx context.Context, alert *Alert) error {
	logLevel := n.getLogLevel(alert.Severity)
	
	fields := []logger.Field{
		logger.String("alert_id", alert.ID),
		logger.String("type", string(alert.Type)),
		logger.String("severity", string(alert.Severity)),
		logger.String("title", alert.Title),
		logger.String("description", alert.Description),
		logger.String("source", alert.Source),
		logger.Float64("value", alert.Value),
		logger.Float64("threshold", alert.Threshold),
		logger.String("status", string(alert.Status)),
		logger.Time("starts_at", alert.StartsAt),
	}

	// ラベルを追加
	for k, v := range alert.Labels {
		fields = append(fields, logger.String(fmt.Sprintf("label_%s", k), v))
	}

	switch logLevel {
	case "error":
		n.logger.Error("アラート通知", fields...)
	case "warn":
		n.logger.Warn("アラート通知", fields...)
	case "info":
		n.logger.Info("アラート通知", fields...)
	default:
		n.logger.Debug("アラート通知", fields...)
	}

	return nil
}

// GetName は通知機能の名前を返す
func (n *LogNotifier) GetName() string {
	return n.name
}

// IsEnabled は通知機能が有効かどうかを返す
func (n *LogNotifier) IsEnabled() bool {
	return n.enabled
}

// SetEnabled は通知機能の有効/無効を設定する
func (n *LogNotifier) SetEnabled(enabled bool) {
	n.enabled = enabled
}

// getLogLevel はアラートの重要度に基づいてログレベルを決定する
func (n *LogNotifier) getLogLevel(severity Severity) string {
	switch severity {
	case SeverityCritical:
		return "error"
	case SeverityError:
		return "error"
	case SeverityWarning:
		return "warn"
	case SeverityInfo:
		return "info"
	default:
		return "debug"
	}
}

// WebhookNotifier はWebhookによる通知機能
type WebhookNotifier struct {
	name     string
	enabled  bool
	url      string
	timeout  time.Duration
	headers  map[string]string
	client   *http.Client
	logger   logger.Logger
}

// WebhookConfig はWebhook通知の設定
type WebhookConfig struct {
	URL     string            `json:"url"`
	Timeout time.Duration     `json:"timeout"`
	Headers map[string]string `json:"headers"`
}

// NewWebhookNotifier は新しいWebhookNotifierを作成する
func NewWebhookNotifier(config WebhookConfig) *WebhookNotifier {
	if config.Timeout == 0 {
		config.Timeout = 10 * time.Second
	}

	if config.Headers == nil {
		config.Headers = make(map[string]string)
	}

	// デフォルトヘッダーを設定
	if _, exists := config.Headers["Content-Type"]; !exists {
		config.Headers["Content-Type"] = "application/json"
	}

	return &WebhookNotifier{
		name:    "webhook",
		enabled: true,
		url:     config.URL,
		timeout: config.Timeout,
		headers: config.Headers,
		client: &http.Client{
			Timeout: config.Timeout,
		},
		logger: logger.GetLogger().WithComponent("alert_notifier_webhook"),
	}
}

// WebhookPayload はWebhookペイロードの構造体
type WebhookPayload struct {
	Alert     *Alert    `json:"alert"`
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
}

// Notify はアラートをWebhookで送信する
func (n *WebhookNotifier) Notify(ctx context.Context, alert *Alert) error {
	if !n.enabled || n.url == "" {
		return nil
	}

	payload := WebhookPayload{
		Alert:     alert,
		Timestamp: time.Now(),
		Source:    "tournament-api-alert-system",
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create webhook request: %w", err)
	}

	// ヘッダーを設定
	for key, value := range n.headers {
		req.Header.Set(key, value)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		n.logger.Error("Webhook送信エラー",
			logger.String("url", n.url),
			logger.String("alert_id", alert.ID),
			logger.Err(err),
		)
		return fmt.Errorf("failed to send webhook: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		n.logger.Error("Webhook送信失敗",
			logger.String("url", n.url),
			logger.String("alert_id", alert.ID),
			logger.Int("status_code", resp.StatusCode),
		)
		return fmt.Errorf("webhook returned non-2xx status: %d", resp.StatusCode)
	}

	n.logger.Info("Webhook送信成功",
		logger.String("url", n.url),
		logger.String("alert_id", alert.ID),
		logger.Int("status_code", resp.StatusCode),
	)

	return nil
}

// GetName は通知機能の名前を返す
func (n *WebhookNotifier) GetName() string {
	return n.name
}

// IsEnabled は通知機能が有効かどうかを返す
func (n *WebhookNotifier) IsEnabled() bool {
	return n.enabled
}

// SetEnabled は通知機能の有効/無効を設定する
func (n *WebhookNotifier) SetEnabled(enabled bool) {
	n.enabled = enabled
}

// EmailNotifier はメール通知機能（簡易実装）
type EmailNotifier struct {
	name     string
	enabled  bool
	smtpHost string
	smtpPort int
	username string
	password string
	from     string
	to       []string
	logger   logger.Logger
}

// EmailConfig はメール通知の設定
type EmailConfig struct {
	SMTPHost string   `json:"smtp_host"`
	SMTPPort int      `json:"smtp_port"`
	Username string   `json:"username"`
	Password string   `json:"password"`
	From     string   `json:"from"`
	To       []string `json:"to"`
}

// NewEmailNotifier は新しいEmailNotifierを作成する
func NewEmailNotifier(config EmailConfig) *EmailNotifier {
	return &EmailNotifier{
		name:     "email",
		enabled:  true,
		smtpHost: config.SMTPHost,
		smtpPort: config.SMTPPort,
		username: config.Username,
		password: config.Password,
		from:     config.From,
		to:       config.To,
		logger:   logger.GetLogger().WithComponent("alert_notifier_email"),
	}
}

// Notify はアラートをメールで送信する
func (n *EmailNotifier) Notify(ctx context.Context, alert *Alert) error {
	if !n.enabled || len(n.to) == 0 {
		return nil
	}

	// 簡易実装：実際の実装ではSMTPライブラリを使用
	n.logger.Info("メール通知送信（簡易実装）",
		logger.String("alert_id", alert.ID),
		logger.String("severity", string(alert.Severity)),
		logger.String("title", alert.Title),
		logger.Strings("recipients", n.to),
	)

	return nil
}

// GetName は通知機能の名前を返す
func (n *EmailNotifier) GetName() string {
	return n.name
}

// IsEnabled は通知機能が有効かどうかを返す
func (n *EmailNotifier) IsEnabled() bool {
	return n.enabled
}

// SetEnabled は通知機能の有効/無効を設定する
func (n *EmailNotifier) SetEnabled(enabled bool) {
	n.enabled = enabled
}

// SlackNotifier はSlack通知機能
type SlackNotifier struct {
	name     string
	enabled  bool
	webhookURL string
	channel  string
	username string
	client   *http.Client
	logger   logger.Logger
}

// SlackConfig はSlack通知の設定
type SlackConfig struct {
	WebhookURL string `json:"webhook_url"`
	Channel    string `json:"channel"`
	Username   string `json:"username"`
}

// NewSlackNotifier は新しいSlackNotifierを作成する
func NewSlackNotifier(config SlackConfig) *SlackNotifier {
	if config.Username == "" {
		config.Username = "Alert Bot"
	}

	return &SlackNotifier{
		name:       "slack",
		enabled:    true,
		webhookURL: config.WebhookURL,
		channel:    config.Channel,
		username:   config.Username,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
		logger: logger.GetLogger().WithComponent("alert_notifier_slack"),
	}
}

// SlackMessage はSlackメッセージの構造体
type SlackMessage struct {
	Channel     string            `json:"channel,omitempty"`
	Username    string            `json:"username,omitempty"`
	Text        string            `json:"text"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
}

// SlackAttachment はSlack添付ファイルの構造体
type SlackAttachment struct {
	Color     string       `json:"color,omitempty"`
	Title     string       `json:"title,omitempty"`
	Text      string       `json:"text,omitempty"`
	Fields    []SlackField `json:"fields,omitempty"`
	Timestamp int64        `json:"ts,omitempty"`
}

// SlackField はSlackフィールドの構造体
type SlackField struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

// Notify はアラートをSlackで送信する
func (n *SlackNotifier) Notify(ctx context.Context, alert *Alert) error {
	if !n.enabled || n.webhookURL == "" {
		return nil
	}

	color := n.getColorBySeverity(alert.Severity)
	
	message := SlackMessage{
		Channel:  n.channel,
		Username: n.username,
		Text:     fmt.Sprintf("🚨 アラート: %s", alert.Title),
		Attachments: []SlackAttachment{
			{
				Color:     color,
				Title:     alert.Title,
				Text:      alert.Description,
				Timestamp: alert.Timestamp.Unix(),
				Fields: []SlackField{
					{
						Title: "重要度",
						Value: string(alert.Severity),
						Short: true,
					},
					{
						Title: "タイプ",
						Value: string(alert.Type),
						Short: true,
					},
					{
						Title: "ソース",
						Value: alert.Source,
						Short: true,
					},
					{
						Title: "ステータス",
						Value: string(alert.Status),
						Short: true,
					},
				},
			},
		},
	}

	// 値と閾値がある場合は追加
	if alert.Value != 0 || alert.Threshold != 0 {
		message.Attachments[0].Fields = append(message.Attachments[0].Fields,
			SlackField{
				Title: "値 / 閾値",
				Value: fmt.Sprintf("%.2f / %.2f", alert.Value, alert.Threshold),
				Short: true,
			},
		)
	}

	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal slack message: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", n.webhookURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create slack request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := n.client.Do(req)
	if err != nil {
		n.logger.Error("Slack通知送信エラー",
			logger.String("alert_id", alert.ID),
			logger.Err(err),
		)
		return fmt.Errorf("failed to send slack notification: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		n.logger.Error("Slack通知送信失敗",
			logger.String("alert_id", alert.ID),
			logger.Int("status_code", resp.StatusCode),
		)
		return fmt.Errorf("slack returned non-200 status: %d", resp.StatusCode)
	}

	n.logger.Info("Slack通知送信成功",
		logger.String("alert_id", alert.ID),
	)

	return nil
}

// GetName は通知機能の名前を返す
func (n *SlackNotifier) GetName() string {
	return n.name
}

// IsEnabled は通知機能が有効かどうかを返す
func (n *SlackNotifier) IsEnabled() bool {
	return n.enabled
}

// SetEnabled は通知機能の有効/無効を設定する
func (n *SlackNotifier) SetEnabled(enabled bool) {
	n.enabled = enabled
}

// getColorBySeverity は重要度に基づいて色を決定する
func (n *SlackNotifier) getColorBySeverity(severity Severity) string {
	switch severity {
	case SeverityCritical:
		return "danger"  // 赤
	case SeverityError:
		return "danger"  // 赤
	case SeverityWarning:
		return "warning" // 黄
	case SeverityInfo:
		return "good"    // 緑
	default:
		return "#808080" // グレー
	}
}

// NotifierFactory は通知機能のファクトリー
type NotifierFactory struct{}

// CreateNotifier は設定に基づいて通知機能を作成する
func (f *NotifierFactory) CreateNotifier(notifierType string, config map[string]interface{}) (AlertNotifier, error) {
	switch strings.ToLower(notifierType) {
	case "log":
		return NewLogNotifier(), nil
		
	case "webhook":
		webhookConfig := WebhookConfig{}
		if url, ok := config["url"].(string); ok {
			webhookConfig.URL = url
		}
		if timeout, ok := config["timeout"].(string); ok {
			if d, err := time.ParseDuration(timeout); err == nil {
				webhookConfig.Timeout = d
			}
		}
		if headers, ok := config["headers"].(map[string]interface{}); ok {
			webhookConfig.Headers = make(map[string]string)
			for k, v := range headers {
				if str, ok := v.(string); ok {
					webhookConfig.Headers[k] = str
				}
			}
		}
		return NewWebhookNotifier(webhookConfig), nil
		
	case "email":
		emailConfig := EmailConfig{}
		if host, ok := config["smtp_host"].(string); ok {
			emailConfig.SMTPHost = host
		}
		if port, ok := config["smtp_port"].(float64); ok {
			emailConfig.SMTPPort = int(port)
		}
		if username, ok := config["username"].(string); ok {
			emailConfig.Username = username
		}
		if password, ok := config["password"].(string); ok {
			emailConfig.Password = password
		}
		if from, ok := config["from"].(string); ok {
			emailConfig.From = from
		}
		if to, ok := config["to"].([]interface{}); ok {
			for _, t := range to {
				if str, ok := t.(string); ok {
					emailConfig.To = append(emailConfig.To, str)
				}
			}
		}
		return NewEmailNotifier(emailConfig), nil
		
	case "slack":
		slackConfig := SlackConfig{}
		if url, ok := config["webhook_url"].(string); ok {
			slackConfig.WebhookURL = url
		}
		if channel, ok := config["channel"].(string); ok {
			slackConfig.Channel = channel
		}
		if username, ok := config["username"].(string); ok {
			slackConfig.Username = username
		}
		return NewSlackNotifier(slackConfig), nil
		
	default:
		return nil, fmt.Errorf("unknown notifier type: %s", notifierType)
	}
}

// LoadNotifiersFromEnv は環境変数から通知機能を読み込む
func LoadNotifiersFromEnv() []AlertNotifier {
	var notifiers []AlertNotifier
	factory := &NotifierFactory{}

	// ログ通知は常に有効
	notifiers = append(notifiers, NewLogNotifier())

	// Webhook通知
	if webhookURL := os.Getenv("ALERT_WEBHOOK_URL"); webhookURL != "" {
		config := map[string]interface{}{
			"url": webhookURL,
		}
		if timeout := os.Getenv("ALERT_WEBHOOK_TIMEOUT"); timeout != "" {
			config["timeout"] = timeout
		}
		if notifier, err := factory.CreateNotifier("webhook", config); err == nil {
			notifiers = append(notifiers, notifier)
		}
	}

	// Slack通知
	if slackURL := os.Getenv("ALERT_SLACK_WEBHOOK_URL"); slackURL != "" {
		config := map[string]interface{}{
			"webhook_url": slackURL,
		}
		if channel := os.Getenv("ALERT_SLACK_CHANNEL"); channel != "" {
			config["channel"] = channel
		}
		if username := os.Getenv("ALERT_SLACK_USERNAME"); username != "" {
			config["username"] = username
		}
		if notifier, err := factory.CreateNotifier("slack", config); err == nil {
			notifiers = append(notifiers, notifier)
		}
	}

	// メール通知
	if smtpHost := os.Getenv("ALERT_SMTP_HOST"); smtpHost != "" {
		config := map[string]interface{}{
			"smtp_host": smtpHost,
		}
		if smtpPort := os.Getenv("ALERT_SMTP_PORT"); smtpPort != "" {
			config["smtp_port"] = smtpPort
		}
		if username := os.Getenv("ALERT_SMTP_USERNAME"); username != "" {
			config["username"] = username
		}
		if password := os.Getenv("ALERT_SMTP_PASSWORD"); password != "" {
			config["password"] = password
		}
		if from := os.Getenv("ALERT_EMAIL_FROM"); from != "" {
			config["from"] = from
		}
		if to := os.Getenv("ALERT_EMAIL_TO"); to != "" {
			config["to"] = strings.Split(to, ",")
		}
		if notifier, err := factory.CreateNotifier("email", config); err == nil {
			notifiers = append(notifiers, notifier)
		}
	}

	return notifiers
}
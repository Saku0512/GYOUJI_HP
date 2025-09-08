package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"backend/internal/alert"
	"backend/internal/logger"
	"backend/internal/models"
)

// AlertHandler はアラート管理のHTTPハンドラー
type AlertHandler struct {
	*BaseHandler
	alertManager  alert.AlertManager
	healthMonitor *alert.HealthMonitor
	logger        logger.Logger
}

// NewAlertHandler は新しいAlertHandlerを作成する
func NewAlertHandler(alertManager alert.AlertManager, healthMonitor *alert.HealthMonitor) *AlertHandler {
	return &AlertHandler{
		BaseHandler:   NewBaseHandler(),
		alertManager:  alertManager,
		healthMonitor: healthMonitor,
		logger:        logger.GetLogger().WithComponent("alert_handler"),
	}
}

// GetAlerts はアラート一覧を取得する
// @Summary アラート一覧取得
// @Description 全てのアラートまたはフィルター条件に一致するアラートを取得する
// @Tags alerts
// @Accept json
// @Produce json
// @Param status query string false "アラートステータス" Enums(firing,resolved,silenced)
// @Param severity query string false "重要度" Enums(info,warning,error,critical)
// @Param type query string false "アラートタイプ"
// @Success 200 {object} models.APIResponse{data=[]alert.Alert}
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts [get]
func (h *AlertHandler) GetAlerts(c *gin.Context) {
	alerts := h.alertManager.GetAlerts()

	// フィルタリング
	status := c.Query("status")
	severity := c.Query("severity")
	alertType := c.Query("type")

	var filteredAlerts []*alert.Alert
	for _, a := range alerts {
		if status != "" && string(a.Status) != status {
			continue
		}
		if severity != "" && string(a.Severity) != severity {
			continue
		}
		if alertType != "" && string(a.Type) != alertType {
			continue
		}
		filteredAlerts = append(filteredAlerts, a)
	}

	h.SendSuccess(c, filteredAlerts, "アラート一覧を取得しました")
}

// GetActiveAlerts はアクティブなアラート一覧を取得する
// @Summary アクティブアラート一覧取得
// @Description 現在アクティブなアラートを取得する
// @Tags alerts
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]alert.Alert}
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/active [get]
func (h *AlertHandler) GetActiveAlerts(c *gin.Context) {
	alerts := h.alertManager.GetActiveAlerts()
	h.SendSuccess(c, alerts, "アクティブアラート一覧を取得しました")
}

// GetAlert は特定のアラートを取得する
// @Summary アラート詳細取得
// @Description 指定されたIDのアラート詳細を取得する
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path string true "アラートID"
// @Success 200 {object} models.APIResponse{data=alert.Alert}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/{id} [get]
func (h *AlertHandler) GetAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "アラートIDが必要です")
		return
	}

	alert, err := h.alertManager.GetAlert(alertID)
	if err != nil {
		h.SendError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "アラートが見つかりません")
		return
	}

	h.SendSuccess(c, alert, "アラート詳細を取得しました")
}

// SilenceAlert はアラートをサイレンスする
// @Summary アラートサイレンス
// @Description 指定されたアラートを一定期間サイレンスする
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path string true "アラートID"
// @Param request body SilenceAlertRequest true "サイレンス設定"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/{id}/silence [post]
func (h *AlertHandler) SilenceAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "アラートIDが必要です")
		return
	}

	var req SilenceAlertRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.SendValidationError(c, err)
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "無効な期間形式です")
		return
	}

	if err := h.alertManager.SilenceAlert(alertID, duration); err != nil {
		if err.Error() == "active alert not found: "+alertID {
			h.SendError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "アクティブなアラートが見つかりません")
			return
		}
		h.SendError(c, http.StatusInternalServerError, "SYSTEM_ERROR", "アラートのサイレンスに失敗しました")
		return
	}

	h.logger.Info("アラートをサイレンスしました",
		logger.String("alert_id", alertID),
		logger.Duration("duration", duration),
		logger.String("user_id", h.GetUserID(c)),
	)

	h.SendSuccess(c, nil, "アラートをサイレンスしました")
}

// ResolveAlert はアラートを手動で解決する
// @Summary アラート解決
// @Description 指定されたアラートを手動で解決する
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path string true "アラートID"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/{id}/resolve [post]
func (h *AlertHandler) ResolveAlert(c *gin.Context) {
	alertID := c.Param("id")
	if alertID == "" {
		h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "アラートIDが必要です")
		return
	}

	if err := h.alertManager.ResolveAlert(alertID); err != nil {
		if err.Error() == "active alert not found: "+alertID {
			h.SendError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "アクティブなアラートが見つかりません")
			return
		}
		h.SendError(c, http.StatusInternalServerError, "SYSTEM_ERROR", "アラートの解決に失敗しました")
		return
	}

	h.logger.Info("アラートを手動解決しました",
		logger.String("alert_id", alertID),
		logger.String("user_id", h.GetUserID(c)),
	)

	h.SendSuccess(c, nil, "アラートを解決しました")
}

// GetAlertRules はアラートルール一覧を取得する
// @Summary アラートルール一覧取得
// @Description 全てのアラートルールを取得する
// @Tags alerts
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=[]alert.AlertRule}
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/rules [get]
func (h *AlertHandler) GetAlertRules(c *gin.Context) {
	rules := h.alertManager.GetRules()
	h.SendSuccess(c, rules, "アラートルール一覧を取得しました")
}

// GetAlertRule は特定のアラートルールを取得する
// @Summary アラートルール詳細取得
// @Description 指定されたIDのアラートルール詳細を取得する
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path string true "ルールID"
// @Success 200 {object} models.APIResponse{data=alert.AlertRule}
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/rules/{id} [get]
func (h *AlertHandler) GetAlertRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "ルールIDが必要です")
		return
	}

	rule, err := h.alertManager.GetRule(ruleID)
	if err != nil {
		h.SendError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "アラートルールが見つかりません")
		return
	}

	h.SendSuccess(c, rule, "アラートルール詳細を取得しました")
}

// CreateAlertRule は新しいアラートルールを作成する
// @Summary アラートルール作成
// @Description 新しいアラートルールを作成する
// @Tags alerts
// @Accept json
// @Produce json
// @Param request body CreateAlertRuleRequest true "アラートルール作成リクエスト"
// @Success 201 {object} models.APIResponse{data=alert.AlertRule}
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/rules [post]
func (h *AlertHandler) CreateAlertRule(c *gin.Context) {
	var req CreateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.SendValidationError(c, err)
		return
	}

	duration, err := time.ParseDuration(req.Duration)
	if err != nil {
		h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "無効な期間形式です")
		return
	}

	rule := alert.NewAlertRule(req.Name, alert.AlertType(req.Type), alert.Severity(req.Severity))
	rule.Description = req.Description
	rule.Query = req.Query
	rule.Threshold = req.Threshold
	rule.Operator = alert.ComparisonOperator(req.Operator)
	rule.Duration = duration
	rule.Labels = req.Labels
	rule.Annotations = req.Annotations
	rule.Enabled = req.Enabled

	if err := h.alertManager.AddRule(rule); err != nil {
		h.SendError(c, http.StatusInternalServerError, "SYSTEM_ERROR", "アラートルールの作成に失敗しました")
		return
	}

	h.logger.Info("アラートルールを作成しました",
		logger.String("rule_id", rule.ID),
		logger.String("rule_name", rule.Name),
		logger.String("user_id", h.GetUserID(c)),
	)

	h.SendSuccess(c, rule, "アラートルールを作成しました")
}

// UpdateAlertRule はアラートルールを更新する
// @Summary アラートルール更新
// @Description 指定されたアラートルールを更新する
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path string true "ルールID"
// @Param request body UpdateAlertRuleRequest true "アラートルール更新リクエスト"
// @Success 200 {object} models.APIResponse{data=alert.AlertRule}
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/rules/{id} [put]
func (h *AlertHandler) UpdateAlertRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "ルールIDが必要です")
		return
	}

	var req UpdateAlertRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.SendValidationError(c, err)
		return
	}

	rule, err := h.alertManager.GetRule(ruleID)
	if err != nil {
		h.SendError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "アラートルールが見つかりません")
		return
	}

	// 更新可能なフィールドを更新
	if req.Name != nil {
		rule.Name = *req.Name
	}
	if req.Description != nil {
		rule.Description = *req.Description
	}
	if req.Query != nil {
		rule.Query = *req.Query
	}
	if req.Threshold != nil {
		rule.Threshold = *req.Threshold
	}
	if req.Operator != nil {
		rule.Operator = alert.ComparisonOperator(*req.Operator)
	}
	if req.Duration != nil {
		duration, err := time.ParseDuration(*req.Duration)
		if err != nil {
			h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "無効な期間形式です")
			return
		}
		rule.Duration = duration
	}
	if req.Enabled != nil {
		rule.Enabled = *req.Enabled
	}
	if req.Labels != nil {
		rule.Labels = req.Labels
	}
	if req.Annotations != nil {
		rule.Annotations = req.Annotations
	}

	if err := h.alertManager.UpdateRule(rule); err != nil {
		h.SendError(c, http.StatusInternalServerError, "SYSTEM_ERROR", "アラートルールの更新に失敗しました")
		return
	}

	h.logger.Info("アラートルールを更新しました",
		logger.String("rule_id", rule.ID),
		logger.String("rule_name", rule.Name),
		logger.String("user_id", h.GetUserID(c)),
	)

	h.SendSuccess(c, rule, "アラートルールを更新しました")
}

// DeleteAlertRule はアラートルールを削除する
// @Summary アラートルール削除
// @Description 指定されたアラートルールを削除する
// @Tags alerts
// @Accept json
// @Produce json
// @Param id path string true "ルールID"
// @Success 200 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/rules/{id} [delete]
func (h *AlertHandler) DeleteAlertRule(c *gin.Context) {
	ruleID := c.Param("id")
	if ruleID == "" {
		h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "ルールIDが必要です")
		return
	}

	if err := h.alertManager.RemoveRule(ruleID); err != nil {
		if err.Error() == "rule not found: "+ruleID {
			h.SendError(c, http.StatusNotFound, "RESOURCE_NOT_FOUND", "アラートルールが見つかりません")
			return
		}
		h.SendError(c, http.StatusInternalServerError, "SYSTEM_ERROR", "アラートルールの削除に失敗しました")
		return
	}

	h.logger.Info("アラートルールを削除しました",
		logger.String("rule_id", ruleID),
		logger.String("user_id", h.GetUserID(c)),
	)

	h.SendSuccess(c, nil, "アラートルールを削除しました")
}

// GetHealthStatus はシステムのヘルス状態を取得する
// @Summary ヘルス状態取得
// @Description システム全体のヘルス状態を取得する
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/health/status [get]
func (h *AlertHandler) GetHealthStatus(c *gin.Context) {
	if h.healthMonitor == nil {
		h.SendError(c, http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", "ヘルスモニターが利用できません")
		return
	}

	checks := h.healthMonitor.GetHealthStatus()
	overallStatus := h.healthMonitor.GetOverallStatus()

	response := map[string]interface{}{
		"overall_status": overallStatus,
		"checks":         checks,
		"timestamp":      time.Now(),
	}

	// 全体的な状態に基づいてHTTPステータスコードを設定
	var statusCode int
	switch overallStatus {
	case alert.HealthStatusHealthy:
		statusCode = http.StatusOK
	case alert.HealthStatusDegraded:
		statusCode = http.StatusOK // 警告だが200を返す
	case alert.HealthStatusUnhealthy:
		statusCode = http.StatusServiceUnavailable
	default:
		statusCode = http.StatusOK
	}

	c.JSON(statusCode, models.APIResponse{
		Success:   true,
		Data:      response,
		Message:   "ヘルス状態を取得しました",
		Code:      statusCode,
		Timestamp: models.Now(),
		RequestID: h.GetRequestID(c),
	})
}

// GetAlertStats はアラート統計情報を取得する
// @Summary アラート統計取得
// @Description アラートの統計情報を取得する
// @Tags alerts
// @Accept json
// @Produce json
// @Param period query string false "集計期間" Enums(1h,24h,7d,30d) default(24h)
// @Success 200 {object} models.APIResponse{data=map[string]interface{}}
// @Failure 500 {object} models.APIResponse
// @Router /api/v1/alerts/stats [get]
func (h *AlertHandler) GetAlertStats(c *gin.Context) {
	period := c.DefaultQuery("period", "24h")
	
	// 期間の解析
	duration, err := time.ParseDuration(period)
	if err != nil {
		h.SendError(c, http.StatusBadRequest, "VALIDATION_ERROR", "無効な期間形式です")
		return
	}

	alerts := h.alertManager.GetAlerts()
	cutoff := time.Now().Add(-duration)

	stats := map[string]interface{}{
		"total":    0,
		"active":   0,
		"resolved": 0,
		"silenced": 0,
		"by_severity": map[string]int{
			"info":     0,
			"warning":  0,
			"error":    0,
			"critical": 0,
		},
		"by_type": make(map[string]int),
		"period":  period,
	}

	severityStats := stats["by_severity"].(map[string]int)
	typeStats := stats["by_type"].(map[string]int)

	for _, alert := range alerts {
		// 期間内のアラートのみカウント
		if alert.Timestamp.Before(cutoff) {
			continue
		}

		stats["total"] = stats["total"].(int) + 1

		// ステータス別
		switch alert.Status {
		case alert.AlertStatusFiring:
			stats["active"] = stats["active"].(int) + 1
		case alert.AlertStatusResolved:
			stats["resolved"] = stats["resolved"].(int) + 1
		case alert.AlertStatusSilenced:
			stats["silenced"] = stats["silenced"].(int) + 1
		}

		// 重要度別
		severityStats[string(alert.Severity)]++

		// タイプ別
		typeStats[string(alert.Type)]++
	}

	h.SendSuccess(c, stats, "アラート統計情報を取得しました")
}

// リクエスト構造体

// SilenceAlertRequest はアラートサイレンスリクエスト
type SilenceAlertRequest struct {
	Duration string `json:"duration" binding:"required" example:"1h"`
	Reason   string `json:"reason" example:"メンテナンス中"`
}

// CreateAlertRuleRequest はアラートルール作成リクエスト
type CreateAlertRuleRequest struct {
	Name        string            `json:"name" binding:"required" example:"高CPU使用率"`
	Type        string            `json:"type" binding:"required" example:"high_cpu"`
	Severity    string            `json:"severity" binding:"required" example:"warning"`
	Description string            `json:"description" example:"CPU使用率が高い状態が続いています"`
	Query       string            `json:"query" binding:"required" example:"cpu_usage"`
	Threshold   float64           `json:"threshold" binding:"required" example:"80"`
	Operator    string            `json:"operator" binding:"required" example:"gt"`
	Duration    string            `json:"duration" binding:"required" example:"5m"`
	Labels      map[string]string `json:"labels" example:"{}"`
	Annotations map[string]string `json:"annotations" example:"{}"`
	Enabled     bool              `json:"enabled" example:"true"`
}

// UpdateAlertRuleRequest はアラートルール更新リクエスト
type UpdateAlertRuleRequest struct {
	Name        *string            `json:"name,omitempty" example:"高CPU使用率"`
	Description *string            `json:"description,omitempty" example:"CPU使用率が高い状態が続いています"`
	Query       *string            `json:"query,omitempty" example:"cpu_usage"`
	Threshold   *float64           `json:"threshold,omitempty" example:"80"`
	Operator    *string            `json:"operator,omitempty" example:"gt"`
	Duration    *string            `json:"duration,omitempty" example:"5m"`
	Enabled     *bool              `json:"enabled,omitempty" example:"true"`
	Labels      map[string]string  `json:"labels,omitempty" example:"{}"`
	Annotations map[string]string  `json:"annotations,omitempty" example:"{}"`
}
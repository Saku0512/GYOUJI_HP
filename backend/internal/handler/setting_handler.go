package handler

import (
	"log"
	"net/http"

	"github.com/saku0512/GYOUJI_HP/backend/internal/service"

	"github.com/gin-gonic/gin"
)

// SettingHandler は設定関連のHTTPリクエストを処理します。
type SettingHandler struct {
	settingService service.SettingService
}

// NewSettingHandler は SettingHandler の新しいインスタンスを生成します。
func NewSettingHandler(s service.SettingService) *SettingHandler {
	return &SettingHandler{settingService: s}
}

// GetVisibility godoc
// @Summary スコア表示設定を取得
// @Description 全体のスコア表示設定を取得します。
// @Tags Settings
// @Produce  json
// @Success 200 {object} map[string]bool
// @Failure 500 {object} map[string]string
// @Router /settings/visibility [get]
func (h *SettingHandler) GetVisibility(c *gin.Context) {
	settings, err := h.settingService.GetVisibility()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get settings"})
		return
	}
	c.JSON(http.StatusOK, settings)
}

// UpdateVisibility godoc
// @Summary スコア表示設定を更新
// @Description 全体のスコア表示設定を更新します。
// @Tags Settings
// @Accept  json
// @Produce  json
// @Param   body body map[string]bool true "表示設定"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/visibility [put]
func (h *SettingHandler) UpdateVisibility(c *gin.Context) {
	var req struct {
		ShowTotalScores         *bool `json:"showTotalScores"`
		ShowQuestionnaireButton *bool `json:"showQuestionnaireButton"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.ShowTotalScores != nil {
		if err := h.settingService.UpdateVisibility("showTotalScores", *req.ShowTotalScores); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update showTotalScores setting"})
			return
		}
	}

	if req.ShowQuestionnaireButton != nil {
		if err := h.settingService.UpdateVisibility("showQuestionnaireButton", *req.ShowQuestionnaireButton); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update showQuestionnaireButton setting"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// GetWeather godoc
// @Summary 天候設定を取得
// @Description 卓球の天候設定を取得します。
// @Tags Settings
// @Produce  json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/weather [get]
func (h *SettingHandler) GetWeather(c *gin.Context) {
	value, err := h.settingService.GetWeather()
	if err != nil {
		log.Printf("!!!!!! DATABASE ERROR on GetWeather: %v !!!!!!", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get weather setting"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tableTennisWeather": value})
}

// UpdateWeather godoc
// @Summary 天候設定を更新
// @Description 卓球の天候設定を更新します。
// @Tags Settings
// @Accept  json
// @Produce  json
// @Param   body body map[string]string true "天候設定"
// @Success 200 {object} map[string]string
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /settings/weather [put]
func (h *SettingHandler) UpdateWeather(c *gin.Context) {
	var req struct {
		Value string `json:"tableTennisWeather"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("Error binding JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("Received weather update request with value: %s", req.Value)

	if err := h.settingService.UpdateWeather(req.Value); err != nil {
		log.Printf("Error updating weather setting: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update weather setting"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Weather setting updated successfully"})
}

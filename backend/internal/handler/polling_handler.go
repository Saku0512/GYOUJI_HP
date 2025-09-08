package handler

import (
	"net/http"
	"strconv"

	"backend/internal/models"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

// PollingHandler はポーリング関連のハンドラー
type PollingHandler struct {
	*BaseHandler
	pollingService *service.PollingService
}

// NewPollingHandler は新しいPollingHandlerを作成する
func NewPollingHandler(pollingService *service.PollingService) *PollingHandler {
	return &PollingHandler{
		BaseHandler:    NewBaseHandler(),
		pollingService: pollingService,
	}
}

// CheckUpdates はデータの更新をチェックする
// @Summary データ更新チェック
// @Description 指定されたスポーツとデータタイプの更新をチェックする
// @Tags Polling
// @Accept json
// @Produce json
// @Param sport path string true "スポーツタイプ" Enums(volleyball,table_tennis,soccer)
// @Param data_type path string true "データタイプ" Enums(tournament,matches,bracket)
// @Param last_etag query string false "最後のETag"
// @Param last_check query string false "最後のチェック時刻"
// @Success 200 {object} models.DataResponse[service.PollingResponse] "更新チェック結果"
// @Failure 400 {object} models.ErrorResponse "リクエストエラー"
// @Failure 404 {object} models.ErrorResponse "データが見つからない"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/polling/{sport}/{data_type}/check [get]
func (h *PollingHandler) CheckUpdates(c *gin.Context) {
	// パラメータを取得
	sportParam := c.Param("sport")
	dataType := c.Param("data_type")
	lastETag := c.Query("last_etag")
	lastCheck := c.Query("last_check")

	// スポーツタイプを検証
	sport := models.SportType(sportParam)
	if !sport.IsValid() {
		h.SendErrorWithCode(c, models.ErrorValidationRequiredField, "無効なスポーツタイプです", http.StatusBadRequest)
		return
	}

	// データタイプを検証
	validDataTypes := []string{"tournament", "matches", "bracket"}
	isValidDataType := false
	for _, validType := range validDataTypes {
		if dataType == validType {
			isValidDataType = true
			break
		}
	}
	if !isValidDataType {
		h.SendErrorWithCode(c, models.ErrorValidationRequiredField, "無効なデータタイプです", http.StatusBadRequest)
		return
	}

	// リクエストを作成
	request := &service.UpdateCheckRequest{
		Sport:     sport,
		DataType:  dataType,
		LastETag:  lastETag,
		LastCheck: lastCheck,
	}

	// 更新をチェック
	response, err := h.pollingService.CheckForUpdates(c.Request.Context(), request)
	if err != nil {
		if serviceErr, ok := err.(*service.ServiceError); ok {
			switch serviceErr.Type {
			case "validation":
				h.SendErrorWithCode(c, models.ErrorValidationRequiredField, serviceErr.Message, http.StatusBadRequest)
			case "not_found":
				h.SendErrorWithCode(c, models.ErrorResourceNotFound, serviceErr.Message, http.StatusNotFound)
			default:
				h.SendErrorWithCode(c, models.ErrorSystemDatabaseError, serviceErr.Message, http.StatusInternalServerError)
			}
		} else {
			h.SendErrorWithCode(c, models.ErrorSystemUnknownError, "データの更新チェックに失敗しました", http.StatusInternalServerError)
		}
		return
	}

	h.SendSuccess(c, response, "データの更新チェックが完了しました", http.StatusOK)
}

// GetLatestData は最新のデータを取得する
// @Summary 最新データ取得
// @Description 指定されたスポーツとデータタイプの最新データを強制取得する
// @Tags Polling
// @Accept json
// @Produce json
// @Param sport path string true "スポーツタイプ" Enums(volleyball,table_tennis,soccer)
// @Param data_type path string true "データタイプ" Enums(tournament,matches,bracket)
// @Success 200 {object} models.DataResponse[service.PollingResponse] "最新データ"
// @Failure 400 {object} models.ErrorResponse "リクエストエラー"
// @Failure 404 {object} models.ErrorResponse "データが見つからない"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/polling/{sport}/{data_type}/latest [get]
func (h *PollingHandler) GetLatestData(c *gin.Context) {
	// パラメータを取得
	sportParam := c.Param("sport")
	dataType := c.Param("data_type")

	// スポーツタイプを検証
	sport := models.SportType(sportParam)
	if !sport.IsValid() {
		h.SendErrorWithCode(c, models.ErrorValidationRequiredField, "無効なスポーツタイプです", http.StatusBadRequest)
		return
	}

	// データタイプを検証
	validDataTypes := []string{"tournament", "matches", "bracket"}
	isValidDataType := false
	for _, validType := range validDataTypes {
		if dataType == validType {
			isValidDataType = true
			break
		}
	}
	if !isValidDataType {
		h.SendErrorWithCode(c, models.ErrorValidationRequiredField, "無効なデータタイプです", http.StatusBadRequest)
		return
	}

	// 最新データを取得
	response, err := h.pollingService.GetLatestData(c.Request.Context(), sport, dataType)
	if err != nil {
		if serviceErr, ok := err.(*service.ServiceError); ok {
			switch serviceErr.Type {
			case "validation":
				h.SendErrorWithCode(c, models.ErrorValidationRequiredField, serviceErr.Message, http.StatusBadRequest)
			case "not_found":
				h.SendErrorWithCode(c, models.ErrorResourceNotFound, serviceErr.Message, http.StatusNotFound)
			default:
				h.SendErrorWithCode(c, models.ErrorSystemDatabaseError, serviceErr.Message, http.StatusInternalServerError)
			}
		} else {
			h.SendErrorWithCode(c, models.ErrorSystemUnknownError, "最新データの取得に失敗しました", http.StatusInternalServerError)
		}
		return
	}

	h.SendSuccess(c, response, "最新データを取得しました", http.StatusOK)
}

// InvalidateCache はキャッシュを無効化する
// @Summary キャッシュ無効化
// @Description 指定されたスポーツとデータタイプのキャッシュを無効化する
// @Tags Polling
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sport path string true "スポーツタイプ" Enums(volleyball,table_tennis,soccer)
// @Param data_type path string true "データタイプ" Enums(tournament,matches,bracket)
// @Success 200 {object} models.DataResponse[interface{}] "キャッシュ無効化成功"
// @Failure 400 {object} models.ErrorResponse "リクエストエラー"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Failure 403 {object} models.ErrorResponse "権限エラー"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/polling/{sport}/{data_type}/invalidate [post]
func (h *PollingHandler) InvalidateCache(c *gin.Context) {
	// 管理者権限チェック
	role, exists := h.GetUserRole(c)
	if !exists || role != "admin" {
		h.SendForbidden(c, "管理者権限が必要です")
		return
	}

	// パラメータを取得
	sportParam := c.Param("sport")
	dataType := c.Param("data_type")

	// スポーツタイプを検証
	sport := models.SportType(sportParam)
	if !sport.IsValid() {
		h.SendErrorWithCode(c, models.ErrorValidationRequiredField, "無効なスポーツタイプです", http.StatusBadRequest)
		return
	}

	// データタイプを検証
	validDataTypes := []string{"tournament", "matches", "bracket"}
	isValidDataType := false
	for _, validType := range validDataTypes {
		if dataType == validType {
			isValidDataType = true
			break
		}
	}
	if !isValidDataType {
		h.SendErrorWithCode(c, models.ErrorValidationRequiredField, "無効なデータタイプです", http.StatusBadRequest)
		return
	}

	// キャッシュを無効化
	h.pollingService.InvalidateCache(sport, dataType)

	h.SendSuccess(c, map[string]interface{}{
		"sport":     sport,
		"data_type": dataType,
		"message":   "キャッシュを無効化しました",
	}, "キャッシュの無効化が完了しました", http.StatusOK)
}

// GetCacheStats はキャッシュ統計を取得する
// @Summary キャッシュ統計取得
// @Description ポーリングサービスのキャッシュ統計を取得する
// @Tags Polling
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.DataResponse[interface{}] "キャッシュ統計"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Failure 403 {object} models.ErrorResponse "権限エラー"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/polling/cache/stats [get]
func (h *PollingHandler) GetCacheStats(c *gin.Context) {
	// 管理者権限チェック
	role, exists := h.GetUserRole(c)
	if !exists || role != "admin" {
		h.SendForbidden(c, "管理者権限が必要です")
		return
	}

	stats := h.pollingService.GetCacheStats()
	h.SendSuccess(c, stats, "キャッシュ統計を取得しました", http.StatusOK)
}

// GetPollingConfig はポーリング設定を取得する
// @Summary ポーリング設定取得
// @Description ポーリングの推奨設定を取得する
// @Tags Polling
// @Accept json
// @Produce json
// @Success 200 {object} models.DataResponse[PollingConfig] "ポーリング設定"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/polling/config [get]
func (h *PollingHandler) GetPollingConfig(c *gin.Context) {
	config := &PollingConfig{
		DefaultInterval:    30,  // 30秒
		MinInterval:        5,   // 最小5秒
		MaxInterval:        300, // 最大5分
		UpdatedInterval:    10,  // 更新時は10秒
		SupportedDataTypes: []string{"tournament", "matches", "bracket"},
		SupportedSports:    []string{"volleyball", "table_tennis", "soccer"},
		CacheExpiry:        30,  // キャッシュ30秒
		UseETag:            true,
	}

	h.SendSuccess(c, config, "ポーリング設定を取得しました", http.StatusOK)
}

// PollingConfig はポーリング設定を表す構造体
type PollingConfig struct {
	DefaultInterval    int      `json:"default_interval"`     // デフォルトポーリング間隔（秒）
	MinInterval        int      `json:"min_interval"`         // 最小ポーリング間隔（秒）
	MaxInterval        int      `json:"max_interval"`         // 最大ポーリング間隔（秒）
	UpdatedInterval    int      `json:"updated_interval"`     // 更新時のポーリング間隔（秒）
	SupportedDataTypes []string `json:"supported_data_types"` // サポートされるデータタイプ
	SupportedSports    []string `json:"supported_sports"`     // サポートされるスポーツ
	CacheExpiry        int      `json:"cache_expiry"`         // キャッシュ有効期限（秒）
	UseETag            bool     `json:"use_etag"`             // ETag使用フラグ
}

// BatchCheckUpdates は複数のデータタイプの更新を一括チェックする
// @Summary 一括更新チェック
// @Description 複数のスポーツ・データタイプの更新を一括でチェックする
// @Tags Polling
// @Accept json
// @Produce json
// @Param request body BatchUpdateCheckRequest true "一括更新チェックリクエスト"
// @Success 200 {object} models.DataResponse[BatchUpdateCheckResponse] "一括更新チェック結果"
// @Failure 400 {object} models.ErrorResponse "リクエストエラー"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/polling/batch/check [post]
func (h *PollingHandler) BatchCheckUpdates(c *gin.Context) {
	var request BatchUpdateCheckRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.SendBindingError(c, err)
		return
	}

	// バリデーション
	if len(request.Checks) == 0 {
		h.SendErrorWithCode(c, models.ErrorValidationRequiredField, "チェック項目が指定されていません", http.StatusBadRequest)
		return
	}

	if len(request.Checks) > 10 {
		h.SendErrorWithCode(c, models.ErrorValidationOutOfRange, "チェック項目は最大10個までです", http.StatusBadRequest)
		return
	}

	// 一括チェックを実行
	response := &BatchUpdateCheckResponse{
		Results: make(map[string]*service.PollingResponse),
		Errors:  make(map[string]string),
	}

	for _, check := range request.Checks {
		key := check.Sport.String() + ":" + check.DataType
		
		result, err := h.pollingService.CheckForUpdates(c.Request.Context(), &check)
		if err != nil {
			response.Errors[key] = err.Error()
		} else {
			response.Results[key] = result
		}
	}

	h.SendSuccess(c, response, "一括更新チェックが完了しました", http.StatusOK)
}

// BatchUpdateCheckRequest は一括更新チェックリクエストを表す
type BatchUpdateCheckRequest struct {
	Checks []service.UpdateCheckRequest `json:"checks" validate:"required,min=1,max=10,dive"`
}

// BatchUpdateCheckResponse は一括更新チェックレスポンスを表す
type BatchUpdateCheckResponse struct {
	Results map[string]*service.PollingResponse `json:"results"`
	Errors  map[string]string                   `json:"errors"`
}
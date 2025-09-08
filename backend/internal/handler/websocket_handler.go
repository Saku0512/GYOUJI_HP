package handler

import (
	"fmt"
	"net/http"

	"backend/internal/models"
	websocketManager "backend/internal/websocket"

	"github.com/gin-gonic/gin"
)

// WebSocketHandler はWebSocket関連のハンドラー
type WebSocketHandler struct {
	*BaseHandler
	manager *websocketManager.Manager
}

// NewWebSocketHandler は新しいWebSocketHandlerを作成する
func NewWebSocketHandler(manager *websocketManager.Manager) *WebSocketHandler {
	return &WebSocketHandler{
		BaseHandler: NewBaseHandler(),
		manager:     manager,
	}
}

// HandleWebSocket はWebSocket接続を処理する
// @Summary WebSocket接続
// @Description WebSocketでリアルタイム更新を受信するための接続エンドポイント
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 101 {string} string "WebSocket接続成功"
// @Failure 400 {object} models.ErrorResponse "リクエストエラー"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /ws [get]
func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	// WebSocket接続をマネージャーに委譲
	h.manager.HandleWebSocket(c)
}

// GetStats はWebSocket統計情報を取得する
// @Summary WebSocket統計情報取得
// @Description WebSocket接続の統計情報を取得する
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.DataResponse[models.WebSocketStats] "統計情報"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Failure 403 {object} models.ErrorResponse "権限エラー"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/websocket/stats [get]
func (h *WebSocketHandler) GetStats(c *gin.Context) {
	// 管理者権限チェック
	role, exists := h.GetUserRole(c)
	if !exists || role != "admin" {
		h.SendForbidden(c, "管理者権限が必要です")
		return
	}

	stats := h.manager.GetStats()
	h.SendSuccess(c, stats, "WebSocket統計情報を取得しました", http.StatusOK)
}

// GetConnections は現在のWebSocket接続一覧を取得する
// @Summary WebSocket接続一覧取得
// @Description 現在のWebSocket接続一覧を取得する
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ListResponse[models.ConnectionInfo] "接続一覧"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Failure 403 {object} models.ErrorResponse "権限エラー"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/websocket/connections [get]
func (h *WebSocketHandler) GetConnections(c *gin.Context) {
	// 管理者権限チェック
	role, exists := h.GetUserRole(c)
	if !exists || role != "admin" {
		h.SendForbidden(c, "管理者権限が必要です")
		return
	}

	connections := h.manager.GetConnections()
	h.SendSuccess(c, connections, "WebSocket接続一覧を取得しました", http.StatusOK)
}

// BroadcastMessage は管理者がメッセージをブロードキャストする
// @Summary メッセージブロードキャスト
// @Description 管理者が全ユーザーまたは特定のスポーツ購読者にメッセージをブロードキャストする
// @Tags WebSocket
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body BroadcastRequest true "ブロードキャストリクエスト"
// @Success 200 {object} models.DataResponse[interface{}] "ブロードキャスト成功"
// @Failure 400 {object} models.ErrorResponse "リクエストエラー"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Failure 403 {object} models.ErrorResponse "権限エラー"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/websocket/broadcast [post]
func (h *WebSocketHandler) BroadcastMessage(c *gin.Context) {
	// 管理者権限チェック
	role, exists := h.GetUserRole(c)
	if !exists || role != "admin" {
		h.SendForbidden(c, "管理者権限が必要です")
		return
	}

	var request BroadcastRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.SendBindingError(c, err)
		return
	}

	// バリデーション
	if !h.ValidateRequest(c, func() models.ValidationErrors {
		return request.Validate()
	}) {
		return
	}

	// WebSocketメッセージを作成
	wsMessage, err := models.NewWebSocketMessage(request.Type, request.Data)
	if err != nil {
		h.SendErrorWithCode(c, models.ErrorSystemUnknownError, "メッセージの作成に失敗しました", http.StatusInternalServerError)
		return
	}

	// ブロードキャスト
	if len(request.Sports) > 0 {
		h.manager.BroadcastToSports(wsMessage, request.Sports)
	} else if len(request.UserIDs) > 0 {
		h.manager.BroadcastToUsers(wsMessage, request.UserIDs)
	} else {
		h.manager.BroadcastToAll(wsMessage)
	}

	h.SendSuccess(c, map[string]interface{}{
		"message_type": request.Type,
		"target_sports": request.Sports,
		"target_users": request.UserIDs,
	}, "メッセージをブロードキャストしました", http.StatusOK)
}

// BroadcastRequest はブロードキャストリクエストの構造体
type BroadcastRequest struct {
	Type    string              `json:"type" validate:"required"`
	Data    interface{}         `json:"data" validate:"required"`
	Sports  []models.SportType  `json:"sports,omitempty" validate:"omitempty,dive,oneof=volleyball table_tennis soccer"`
	UserIDs []int               `json:"user_ids,omitempty" validate:"omitempty,dive,min=1"`
}

// Validate はBroadcastRequestのバリデーションを実行する
func (r *BroadcastRequest) Validate() models.ValidationErrors {
	errors := models.NewValidationErrors()

	// タイプの検証
	if r.Type == "" {
		errors.AddFieldError("type", "メッセージタイプは必須です", "")
	}

	// データの検証
	if r.Data == nil {
		errors.AddFieldError("data", "メッセージデータは必須です", "")
	}

	// スポーツの検証
	for i, sport := range r.Sports {
		if !sport.IsValid() {
			errors.AddFieldError(
				fmt.Sprintf("sports[%d]", i),
				"無効なスポーツタイプです",
				sport.String(),
			)
		}
	}

	// ユーザーIDの検証
	for i, userID := range r.UserIDs {
		if userID <= 0 {
			errors.AddFieldError(
				fmt.Sprintf("user_ids[%d]", i),
				"ユーザーIDは1以上である必要があります",
				fmt.Sprintf("%d", userID),
			)
		}
	}

	return errors
}
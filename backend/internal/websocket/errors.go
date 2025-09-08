package websocket

import (
	"fmt"
	"log"

	"backend/internal/models"
)

// WebSocketError はWebSocket関連のエラーを表す
type WebSocketError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Error はerrorインターフェースを実装する
func (e *WebSocketError) Error() string {
	return fmt.Sprintf("WebSocket Error [%s]: %s", e.Code, e.Message)
}

// NewWebSocketError は新しいWebSocketエラーを作成する
func NewWebSocketError(code, message string) *WebSocketError {
	return &WebSocketError{
		Code:    code,
		Message: message,
	}
}

// NewWebSocketErrorWithDetails は詳細付きのWebSocketエラーを作成する
func NewWebSocketErrorWithDetails(code, message, details string) *WebSocketError {
	return &WebSocketError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// WebSocketエラーコード定数
const (
	// 接続関連エラー
	ErrorConnectionFailed    = "CONNECTION_FAILED"
	ErrorConnectionClosed    = "CONNECTION_CLOSED"
	ErrorConnectionTimeout   = "CONNECTION_TIMEOUT"
	ErrorConnectionLimit     = "CONNECTION_LIMIT_EXCEEDED"
	
	// 認証関連エラー
	ErrorAuthRequired        = "AUTH_REQUIRED"
	ErrorAuthFailed          = "AUTH_FAILED"
	ErrorAuthTokenInvalid    = "AUTH_TOKEN_INVALID"
	ErrorAuthTokenExpired    = "AUTH_TOKEN_EXPIRED"
	
	// メッセージ関連エラー
	ErrorInvalidMessage      = "INVALID_MESSAGE_FORMAT"
	ErrorMessageTooLarge     = "MESSAGE_TOO_LARGE"
	ErrorUnknownMessageType  = "UNKNOWN_MESSAGE_TYPE"
	ErrorMessageSendFailed   = "MESSAGE_SEND_FAILED"
	
	// 購読関連エラー
	ErrorInvalidSubscription = "INVALID_SUBSCRIPTION"
	ErrorSubscriptionFailed  = "SUBSCRIPTION_FAILED"
	ErrorNotSubscribed       = "NOT_SUBSCRIBED"
	
	// システム関連エラー
	ErrorSystemOverload      = "SYSTEM_OVERLOAD"
	ErrorSystemMaintenance   = "SYSTEM_MAINTENANCE"
	ErrorInternalError       = "INTERNAL_ERROR"
)

// ErrorHandler はWebSocketエラーを処理するハンドラー
type ErrorHandler struct {
	manager *Manager
}

// NewErrorHandler は新しいエラーハンドラーを作成する
func NewErrorHandler(manager *Manager) *ErrorHandler {
	return &ErrorHandler{
		manager: manager,
	}
}

// HandleConnectionError は接続エラーを処理する
func (h *ErrorHandler) HandleConnectionError(client *Client, err error) {
	log.Printf("Connection error for client %s: %v", client.ID, err)
	
	// エラーの種類に応じて適切な処理を行う
	wsError := h.classifyError(err)
	
	// クライアントにエラーを送信（可能な場合）
	if client.Connection != nil {
		client.sendError(wsError.Code, wsError.Message)
	}
	
	// 統計を更新
	h.manager.stats.ErrorCount++
	
	// 重大なエラーの場合は接続を閉じる
	if h.isCriticalError(wsError.Code) {
		h.manager.unregister <- client
	}
}

// HandleMessageError はメッセージエラーを処理する
func (h *ErrorHandler) HandleMessageError(client *Client, err error, messageType string) {
	log.Printf("Message error for client %s (type: %s): %v", client.ID, messageType, err)
	
	wsError := h.classifyError(err)
	client.sendError(wsError.Code, wsError.Message)
	
	// 統計を更新
	h.manager.stats.ErrorCount++
}

// HandleAuthError は認証エラーを処理する
func (h *ErrorHandler) HandleAuthError(client *Client, err error) {
	log.Printf("Auth error for client %s: %v", client.ID, err)
	
	wsError := NewWebSocketError(ErrorAuthFailed, "認証に失敗しました")
	client.sendError(wsError.Code, wsError.Message)
	
	// 認証エラーの場合は接続を維持するが、認証状態をリセット
	client.Info.UserID = 0
	client.Info.Username = "anonymous"
	client.Info.Role = "guest"
	client.Info.Sports = make([]models.SportType, 0)
}

// HandleSubscriptionError は購読エラーを処理する
func (h *ErrorHandler) HandleSubscriptionError(client *Client, err error, sports []models.SportType) {
	log.Printf("Subscription error for client %s (sports: %v): %v", client.ID, sports, err)
	
	wsError := NewWebSocketError(ErrorSubscriptionFailed, "購読に失敗しました")
	client.sendError(wsError.Code, wsError.Message)
}

// HandleBroadcastError はブロードキャストエラーを処理する
func (h *ErrorHandler) HandleBroadcastError(err error, messageType string, targetCount int) {
	log.Printf("Broadcast error (type: %s, targets: %d): %v", messageType, targetCount, err)
	
	// ブロードキャストエラーは統計のみ更新
	h.manager.stats.ErrorCount++
}

// classifyError はエラーを分類してWebSocketErrorに変換する
func (h *ErrorHandler) classifyError(err error) *WebSocketError {
	errMsg := err.Error()
	
	// エラーメッセージに基づいて分類
	switch {
	case contains(errMsg, "connection", "closed"):
		return NewWebSocketError(ErrorConnectionClosed, "接続が閉じられました")
	case contains(errMsg, "timeout"):
		return NewWebSocketError(ErrorConnectionTimeout, "接続がタイムアウトしました")
	case contains(errMsg, "auth", "token"):
		return NewWebSocketError(ErrorAuthTokenInvalid, "認証トークンが無効です")
	case contains(errMsg, "message", "format"):
		return NewWebSocketError(ErrorInvalidMessage, "メッセージ形式が無効です")
	case contains(errMsg, "too", "large"):
		return NewWebSocketError(ErrorMessageTooLarge, "メッセージが大きすぎます")
	default:
		return NewWebSocketError(ErrorInternalError, "内部エラーが発生しました")
	}
}

// isCriticalError は重大なエラーかどうかを判定する
func (h *ErrorHandler) isCriticalError(errorCode string) bool {
	criticalErrors := []string{
		ErrorConnectionFailed,
		ErrorConnectionClosed,
		ErrorConnectionTimeout,
		ErrorSystemOverload,
		ErrorInternalError,
	}
	
	for _, critical := range criticalErrors {
		if errorCode == critical {
			return true
		}
	}
	
	return false
}

// contains は文字列に指定されたキーワードが含まれているかチェックする
func contains(str string, keywords ...string) bool {
	for _, keyword := range keywords {
		if len(str) >= len(keyword) {
			for i := 0; i <= len(str)-len(keyword); i++ {
				if str[i:i+len(keyword)] == keyword {
					return true
				}
			}
		}
	}
	return false
}

// RecoveryHandler はパニックからの回復を処理する
func (h *ErrorHandler) RecoveryHandler() {
	if r := recover(); r != nil {
		log.Printf("WebSocket panic recovered: %v", r)
		
		// パニック統計を更新
		h.manager.stats.ErrorCount++
		
		// システムメンテナンスメッセージを送信
		maintenanceMsg, _ := models.NewWebSocketMessage(
			"system_maintenance",
			map[string]interface{}{
				"message": "システムメンテナンス中です。しばらくお待ちください。",
				"code":    ErrorSystemMaintenance,
			},
		)
		
		h.manager.BroadcastToAll(maintenanceMsg)
	}
}

// GetErrorStats はエラー統計を取得する
func (h *ErrorHandler) GetErrorStats() map[string]interface{} {
	return map[string]interface{}{
		"total_errors":     h.manager.stats.ErrorCount,
		"error_rate":       h.calculateErrorRate(),
		"last_error_time":  h.getLastErrorTime(),
	}
}

// calculateErrorRate はエラー率を計算する
func (h *ErrorHandler) calculateErrorRate() float64 {
	totalMessages := h.manager.stats.MessagesSent + h.manager.stats.MessagesReceived
	if totalMessages == 0 {
		return 0.0
	}
	
	return float64(h.manager.stats.ErrorCount) / float64(totalMessages) * 100
}

// getLastErrorTime は最後のエラー時刻を取得する（簡易実装）
func (h *ErrorHandler) getLastErrorTime() string {
	// 実際の実装では、最後のエラー時刻を記録する必要がある
	return h.manager.stats.LastUpdated
}
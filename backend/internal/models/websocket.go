package models

import (
	"encoding/json"
	"time"
)

// WebSocketMessage はWebSocketで送受信されるメッセージの基本構造
type WebSocketMessage struct {
	Type      string          `json:"type"`      // メッセージタイプ
	Data      json.RawMessage `json:"data"`      // メッセージデータ（JSON）
	Timestamp string          `json:"timestamp"` // タイムスタンプ（ISO 8601形式）
	RequestID string          `json:"request_id,omitempty"` // リクエストID（追跡用）
}

// NewWebSocketMessage は新しいWebSocketメッセージを作成する
func NewWebSocketMessage(messageType string, data interface{}) (*WebSocketMessage, error) {
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	return &WebSocketMessage{
		Type:      messageType,
		Data:      dataBytes,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

// SetRequestID はリクエストIDを設定する
func (m *WebSocketMessage) SetRequestID(requestID string) *WebSocketMessage {
	m.RequestID = requestID
	return m
}

// WebSocketMessageType はWebSocketメッセージタイプの定数
type WebSocketMessageType string

const (
	// 接続関連
	MessageTypeConnect    WebSocketMessageType = "connect"
	MessageTypeDisconnect WebSocketMessageType = "disconnect"
	MessageTypePing       WebSocketMessageType = "ping"
	MessageTypePong       WebSocketMessageType = "pong"
	
	// 購読関連
	MessageTypeSubscribe   WebSocketMessageType = "subscribe"
	MessageTypeUnsubscribe WebSocketMessageType = "unsubscribe"
	
	// 更新通知
	MessageTypeTournamentUpdate WebSocketMessageType = "tournament_update"
	MessageTypeMatchUpdate      WebSocketMessageType = "match_update"
	MessageTypeMatchResult      WebSocketMessageType = "match_result"
	MessageTypeBracketUpdate    WebSocketMessageType = "bracket_update"
	
	// エラー
	MessageTypeError WebSocketMessageType = "error"
	
	// 認証
	MessageTypeAuth WebSocketMessageType = "auth"
)

// String はWebSocketMessageTypeの文字列表現を返す
func (t WebSocketMessageType) String() string {
	return string(t)
}

// SubscribeRequest は購読リクエストの構造体
type SubscribeRequest struct {
	Sports []SportType `json:"sports" validate:"required,min=1,dive,oneof=volleyball table_tennis soccer"`
}

// UnsubscribeRequest は購読解除リクエストの構造体
type UnsubscribeRequest struct {
	Sports []SportType `json:"sports" validate:"required,min=1,dive,oneof=volleyball table_tennis soccer"`
}

// AuthRequest はWebSocket認証リクエストの構造体
type AuthRequest struct {
	Token string `json:"token" validate:"required"`
}

// UpdateNotification は更新通知の構造体
type UpdateNotification struct {
	Type      WebSocketMessageType `json:"type"`      // 更新タイプ
	Sport     SportType            `json:"sport"`     // 対象スポーツ
	Data      interface{}          `json:"data"`      // 更新データ
	Timestamp string               `json:"timestamp"` // タイムスタンプ
}

// NewUpdateNotification は新しい更新通知を作成する
func NewUpdateNotification(messageType WebSocketMessageType, sport SportType, data interface{}) *UpdateNotification {
	return &UpdateNotification{
		Type:      messageType,
		Sport:     sport,
		Data:      data,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// TournamentUpdateData はトーナメント更新データの構造体
type TournamentUpdateData struct {
	Tournament *Tournament `json:"tournament"`
	Action     string      `json:"action"` // "created", "updated", "deleted", "status_changed"
}

// MatchUpdateData は試合更新データの構造体
type MatchUpdateData struct {
	Match  *Match `json:"match"`
	Action string `json:"action"` // "created", "updated", "deleted", "result_updated"
}

// BracketUpdateData はブラケット更新データの構造体
type BracketUpdateData struct {
	Sport   SportType `json:"sport"`
	Bracket *Bracket  `json:"bracket"`
	Action  string    `json:"action"` // "updated", "regenerated"
}

// ErrorNotification はエラー通知の構造体
type ErrorNotification struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// NewErrorNotification は新しいエラー通知を作成する
func NewErrorNotification(code, message string) *ErrorNotification {
	return &ErrorNotification{
		Code:    code,
		Message: message,
	}
}

// ConnectionInfo はWebSocket接続情報の構造体
type ConnectionInfo struct {
	ID             string      `json:"id"`              // 接続ID
	UserID         int         `json:"user_id"`         // ユーザーID
	Username       string      `json:"username"`        // ユーザー名
	Role           string      `json:"role"`            // ユーザーロール
	Sports         []SportType `json:"sports"`          // 購読中のスポーツ
	ConnectedAt    string      `json:"connected_at"`    // 接続時刻
	LastActiveAt   string      `json:"last_active_at"`  // 最終アクティブ時刻
	RemoteAddr     string      `json:"remote_addr"`     // リモートアドレス
	UserAgent      string      `json:"user_agent"`      // ユーザーエージェント
}

// NewConnectionInfo は新しい接続情報を作成する
func NewConnectionInfo(id string, userID int, username, role, remoteAddr, userAgent string) *ConnectionInfo {
	now := time.Now().UTC().Format(time.RFC3339)
	return &ConnectionInfo{
		ID:           id,
		UserID:       userID,
		Username:     username,
		Role:         role,
		Sports:       make([]SportType, 0),
		ConnectedAt:  now,
		LastActiveAt: now,
		RemoteAddr:   remoteAddr,
		UserAgent:    userAgent,
	}
}

// UpdateLastActive は最終アクティブ時刻を更新する
func (c *ConnectionInfo) UpdateLastActive() {
	c.LastActiveAt = time.Now().UTC().Format(time.RFC3339)
}

// AddSport はスポーツを購読リストに追加する
func (c *ConnectionInfo) AddSport(sport SportType) {
	// 重複チェック
	for _, s := range c.Sports {
		if s == sport {
			return
		}
	}
	c.Sports = append(c.Sports, sport)
}

// RemoveSport はスポーツを購読リストから削除する
func (c *ConnectionInfo) RemoveSport(sport SportType) {
	for i, s := range c.Sports {
		if s == sport {
			c.Sports = append(c.Sports[:i], c.Sports[i+1:]...)
			return
		}
	}
}

// IsSubscribedTo は指定されたスポーツを購読しているかチェックする
func (c *ConnectionInfo) IsSubscribedTo(sport SportType) bool {
	for _, s := range c.Sports {
		if s == sport {
			return true
		}
	}
	return false
}

// WebSocketStats はWebSocket統計情報の構造体
type WebSocketStats struct {
	TotalConnections    int                        `json:"total_connections"`
	ActiveConnections   int                        `json:"active_connections"`
	ConnectionsBySport  map[SportType]int          `json:"connections_by_sport"`
	ConnectionsByUser   map[int]int                `json:"connections_by_user"`
	MessagesSent        int64                      `json:"messages_sent"`
	MessagesReceived    int64                      `json:"messages_received"`
	ErrorCount          int64                      `json:"error_count"`
	LastUpdated         string                     `json:"last_updated"`
}

// NewWebSocketStats は新しいWebSocket統計情報を作成する
func NewWebSocketStats() *WebSocketStats {
	return &WebSocketStats{
		ConnectionsBySport: make(map[SportType]int),
		ConnectionsByUser:  make(map[int]int),
		LastUpdated:        time.Now().UTC().Format(time.RFC3339),
	}
}

// UpdateStats は統計情報を更新する
func (s *WebSocketStats) UpdateStats() {
	s.LastUpdated = time.Now().UTC().Format(time.RFC3339)
}
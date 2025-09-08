package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"backend/internal/models"

	"github.com/gorilla/websocket"
)

const (
	// 書き込みタイムアウト
	writeWait = 10 * time.Second
	
	// Pongメッセージの待機時間
	pongWait = 60 * time.Second
	
	// Pingメッセージの送信間隔（pongWaitより短くする必要がある）
	pingPeriod = (pongWait * 9) / 10
	
	// 最大メッセージサイズ
	maxMessageSize = 512
)

// readPump はWebSocketからメッセージを読み取る
func (c *Client) readPump() {
	defer func() {
		c.Manager.unregister <- c
		c.Connection.Close()
	}()

	// 設定
	c.Connection.SetReadLimit(maxMessageSize)
	c.Connection.SetReadDeadline(time.Now().Add(pongWait))
	c.Connection.SetPongHandler(func(string) error {
		c.Connection.SetReadDeadline(time.Now().Add(pongWait))
		c.Info.UpdateLastActive()
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
		}

		// メッセージを読み取り
		_, messageBytes, err := c.Connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Manager.errorHandler.HandleConnectionError(c, err)
			}
			break
		}

		// 統計を更新
		c.Manager.stats.MessagesReceived++
		c.Info.UpdateLastActive()

		// メッセージを処理
		c.handleMessage(messageBytes)
	}
}

// writePump はWebSocketにメッセージを書き込む
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Connection.Close()
	}()

	for {
		select {
		case <-c.ctx.Done():
			return
			
		case message, ok := <-c.Send:
			c.Connection.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// チャンネルが閉じられた
				c.Connection.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Connection.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// キューに残っているメッセージも送信
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Connection.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Connection.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage は受信したメッセージを処理する
func (c *Client) handleMessage(messageBytes []byte) {
	var wsMessage models.WebSocketMessage
	if err := json.Unmarshal(messageBytes, &wsMessage); err != nil {
		c.Manager.errorHandler.HandleMessageError(c, err, "unknown")
		return
	}

	// メッセージタイプに応じて処理
	switch models.WebSocketMessageType(wsMessage.Type) {
	case models.MessageTypeAuth:
		c.handleAuth(wsMessage.Data)
		
	case models.MessageTypeSubscribe:
		c.handleSubscribe(wsMessage.Data)
		
	case models.MessageTypeUnsubscribe:
		c.handleUnsubscribe(wsMessage.Data)
		
	case models.MessageTypePong:
		// Pongメッセージは特に処理不要（readPumpで処理済み）
		
	default:
		log.Printf("Unknown message type from client %s: %s", c.ID, wsMessage.Type)
		c.sendError("UNKNOWN_MESSAGE_TYPE", "不明なメッセージタイプです")
	}
}

// handleAuth は認証メッセージを処理する
func (c *Client) handleAuth(data json.RawMessage) {
	var authRequest models.AuthRequest
	if err := json.Unmarshal(data, &authRequest); err != nil {
		c.sendError("INVALID_AUTH_REQUEST", "認証リクエストが無効です")
		return
	}

	// JWTトークンを検証（実際の実装では適切なJWT検証を行う）
	userID, username, role, err := c.validateToken(authRequest.Token)
	if err != nil {
		c.sendError("AUTH_FAILED", "認証に失敗しました")
		return
	}

	// 接続情報を更新
	c.Manager.mutex.Lock()
	c.Info.UserID = userID
	c.Info.Username = username
	c.Info.Role = role
	
	// ユーザー別統計を更新
	c.Manager.stats.ConnectionsByUser[userID]++
	c.Manager.mutex.Unlock()

	// 認証成功メッセージを送信
	authSuccessMsg, _ := models.NewWebSocketMessage(
		models.MessageTypeAuth.String(),
		map[string]interface{}{
			"success":  true,
			"user_id":  userID,
			"username": username,
			"role":     role,
			"message":  "認証が完了しました",
		},
	)
	c.sendMessage(authSuccessMsg)

	log.Printf("Client %s authenticated as user %d (%s)", c.ID, userID, username)
}

// handleSubscribe は購読メッセージを処理する
func (c *Client) handleSubscribe(data json.RawMessage) {
	var subscribeRequest models.SubscribeRequest
	if err := json.Unmarshal(data, &subscribeRequest); err != nil {
		c.sendError("INVALID_SUBSCRIBE_REQUEST", "購読リクエストが無効です")
		return
	}

	// 認証チェック
	if c.Info.UserID == 0 {
		c.sendError("AUTH_REQUIRED", "購読には認証が必要です")
		return
	}

	// スポーツを購読リストに追加
	c.Manager.mutex.Lock()
	for _, sport := range subscribeRequest.Sports {
		if !sport.IsValid() {
			continue
		}
		
		if !c.Info.IsSubscribedTo(sport) {
			c.Info.AddSport(sport)
			c.Manager.stats.ConnectionsBySport[sport]++
		}
	}
	c.Manager.mutex.Unlock()

	// 購読成功メッセージを送信
	subscribeSuccessMsg, _ := models.NewWebSocketMessage(
		models.MessageTypeSubscribe.String(),
		map[string]interface{}{
			"success": true,
			"sports":  subscribeRequest.Sports,
			"message": "購読が完了しました",
		},
	)
	c.sendMessage(subscribeSuccessMsg)

	log.Printf("Client %s subscribed to sports: %v", c.ID, subscribeRequest.Sports)
}

// handleUnsubscribe は購読解除メッセージを処理する
func (c *Client) handleUnsubscribe(data json.RawMessage) {
	var unsubscribeRequest models.UnsubscribeRequest
	if err := json.Unmarshal(data, &unsubscribeRequest); err != nil {
		c.sendError("INVALID_UNSUBSCRIBE_REQUEST", "購読解除リクエストが無効です")
		return
	}

	// 認証チェック
	if c.Info.UserID == 0 {
		c.sendError("AUTH_REQUIRED", "購読解除には認証が必要です")
		return
	}

	// スポーツを購読リストから削除
	c.Manager.mutex.Lock()
	for _, sport := range unsubscribeRequest.Sports {
		if !sport.IsValid() {
			continue
		}
		
		if c.Info.IsSubscribedTo(sport) {
			c.Info.RemoveSport(sport)
			if count, exists := c.Manager.stats.ConnectionsBySport[sport]; exists {
				if count <= 1 {
					delete(c.Manager.stats.ConnectionsBySport, sport)
				} else {
					c.Manager.stats.ConnectionsBySport[sport] = count - 1
				}
			}
		}
	}
	c.Manager.mutex.Unlock()

	// 購読解除成功メッセージを送信
	unsubscribeSuccessMsg, _ := models.NewWebSocketMessage(
		models.MessageTypeUnsubscribe.String(),
		map[string]interface{}{
			"success": true,
			"sports":  unsubscribeRequest.Sports,
			"message": "購読解除が完了しました",
		},
	)
	c.sendMessage(unsubscribeSuccessMsg)

	log.Printf("Client %s unsubscribed from sports: %v", c.ID, unsubscribeRequest.Sports)
}

// validateToken はJWTトークンを検証する
func (c *Client) validateToken(token string) (userID int, username, role string, err error) {
	if token == "" {
		return 0, "", "", fmt.Errorf("empty token")
	}
	
	// TODO: 実際のJWT検証サービスを使用する
	// 現在は簡易的な実装として、トークンの形式のみチェック
	if len(token) < 10 {
		return 0, "", "", fmt.Errorf("invalid token format")
	}
	
	// 仮の実装（実際にはJWTライブラリとAuthServiceを使用）
	// トークンが "admin_" で始まる場合は管理者として扱う
	if len(token) > 6 && token[:6] == "admin_" {
		return 1, "admin", "admin", nil
	}
	
	// その他は一般ユーザーとして扱う
	return 2, "user", "user", nil
}
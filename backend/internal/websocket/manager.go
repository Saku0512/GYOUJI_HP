package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/google/uuid"
)

// Manager はWebSocket接続を管理するマネージャー
type Manager struct {
	// 接続管理
	connections map[string]*Client
	mutex       sync.RWMutex
	
	// チャンネル
	register   chan *Client
	unregister chan *Client
	broadcast  chan *BroadcastMessage
	
	// 統計情報
	stats *models.WebSocketStats
	
	// エラーハンドラー
	errorHandler *ErrorHandler
	
	// 設定
	upgrader websocket.Upgrader
	
	// コンテキスト
	ctx    context.Context
	cancel context.CancelFunc
}

// Client はWebSocket接続クライアントを表す
type Client struct {
	ID         string
	Connection *websocket.Conn
	Manager    *Manager
	Info       *models.ConnectionInfo
	Send       chan []byte
	ctx        context.Context
	cancel     context.CancelFunc
}

// BroadcastMessage はブロードキャストメッセージを表す
type BroadcastMessage struct {
	Message *models.WebSocketMessage
	Sports  []models.SportType // 対象スポーツ（空の場合は全体）
	UserIDs []int              // 対象ユーザーID（空の場合は全ユーザー）
}

// NewManager は新しいWebSocketマネージャーを作成する
func NewManager() *Manager {
	ctx, cancel := context.WithCancel(context.Background())
	
	manager := &Manager{
		connections: make(map[string]*Client),
		register:    make(chan *Client, 256),
		unregister:  make(chan *Client, 256),
		broadcast:   make(chan *BroadcastMessage, 1024),
		stats:       models.NewWebSocketStats(),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// 本番環境では適切なオリジンチェックを実装
				return true
			},
		},
		ctx:    ctx,
		cancel: cancel,
	}
	
	// エラーハンドラーを初期化
	manager.errorHandler = NewErrorHandler(manager)
	
	return manager
}

// Start はマネージャーを開始する
func (m *Manager) Start() {
	go m.run()
	log.Println("WebSocket Manager started")
}

// Stop はマネージャーを停止する
func (m *Manager) Stop() {
	m.cancel()
	
	// 全ての接続を閉じる
	m.mutex.Lock()
	for _, client := range m.connections {
		client.cancel()
		client.Connection.Close()
	}
	m.mutex.Unlock()
	
	log.Println("WebSocket Manager stopped")
}

// run はマネージャーのメインループ
func (m *Manager) run() {
	ticker := time.NewTicker(30 * time.Second) // ヘルスチェック用
	defer ticker.Stop()

	for {
		select {
		case <-m.ctx.Done():
			return
			
		case client := <-m.register:
			m.registerClient(client)
			
		case client := <-m.unregister:
			m.unregisterClient(client)
			
		case message := <-m.broadcast:
			m.broadcastMessage(message)
			
		case <-ticker.C:
			m.healthCheck()
		}
	}
}

// HandleWebSocket はWebSocket接続をハンドルする
func (m *Manager) HandleWebSocket(c *gin.Context) {
	// WebSocket接続にアップグレード
	conn, err := m.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}

	// クライアント作成
	clientID := uuid.New().String()
	ctx, cancel := context.WithCancel(m.ctx)
	
	client := &Client{
		ID:         clientID,
		Connection: conn,
		Manager:    m,
		Send:       make(chan []byte, 256),
		ctx:        ctx,
		cancel:     cancel,
	}

	// 接続情報を初期化（認証前は匿名）
	client.Info = models.NewConnectionInfo(
		clientID,
		0, // 認証前は0
		"anonymous",
		"guest",
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	// クライアントを登録
	m.register <- client

	// ゴルーチンを開始
	go client.writePump()
	go client.readPump()
}

// registerClient はクライアントを登録する
func (m *Manager) registerClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	m.connections[client.ID] = client
	m.stats.TotalConnections++
	m.stats.ActiveConnections++
	m.stats.UpdateStats()
	
	log.Printf("Client registered: %s (Total: %d)", client.ID, m.stats.ActiveConnections)
	
	// 接続成功メッセージを送信
	connectMsg, _ := models.NewWebSocketMessage(
		models.MessageTypeConnect.String(),
		map[string]interface{}{
			"client_id": client.ID,
			"message":   "WebSocket接続が確立されました",
		},
	)
	client.sendMessage(connectMsg)
}

// unregisterClient はクライアントの登録を解除する
func (m *Manager) unregisterClient(client *Client) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	
	if _, exists := m.connections[client.ID]; exists {
		delete(m.connections, client.ID)
		close(client.Send)
		m.stats.ActiveConnections--
		
		// ユーザー別統計を更新
		if client.Info.UserID > 0 {
			if count, exists := m.stats.ConnectionsByUser[client.Info.UserID]; exists {
				if count <= 1 {
					delete(m.stats.ConnectionsByUser, client.Info.UserID)
				} else {
					m.stats.ConnectionsByUser[client.Info.UserID] = count - 1
				}
			}
		}
		
		// スポーツ別統計を更新
		for _, sport := range client.Info.Sports {
			if count, exists := m.stats.ConnectionsBySport[sport]; exists {
				if count <= 1 {
					delete(m.stats.ConnectionsBySport, sport)
				} else {
					m.stats.ConnectionsBySport[sport] = count - 1
				}
			}
		}
		
		m.stats.UpdateStats()
		
		log.Printf("Client unregistered: %s (Total: %d)", client.ID, m.stats.ActiveConnections)
	}
}

// broadcastMessage はメッセージをブロードキャストする
func (m *Manager) broadcastMessage(broadcastMsg *BroadcastMessage) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	messageBytes, err := json.Marshal(broadcastMsg.Message)
	if err != nil {
		log.Printf("Failed to marshal broadcast message: %v", err)
		return
	}
	
	sentCount := 0
	for _, client := range m.connections {
		// 対象ユーザーIDが指定されている場合はチェック
		if len(broadcastMsg.UserIDs) > 0 {
			found := false
			for _, userID := range broadcastMsg.UserIDs {
				if client.Info.UserID == userID {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// 対象スポーツが指定されている場合はチェック
		if len(broadcastMsg.Sports) > 0 {
			found := false
			for _, sport := range broadcastMsg.Sports {
				if client.Info.IsSubscribedTo(sport) {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		// メッセージを送信
		select {
		case client.Send <- messageBytes:
			sentCount++
		default:
			// 送信バッファが満杯の場合は接続を閉じる
			close(client.Send)
			delete(m.connections, client.ID)
		}
	}
	
	m.stats.MessagesSent += int64(sentCount)
	log.Printf("Broadcast message sent to %d clients", sentCount)
}

// BroadcastToSports は指定されたスポーツの購読者にメッセージをブロードキャストする
func (m *Manager) BroadcastToSports(message *models.WebSocketMessage, sports []models.SportType) {
	broadcastMsg := &BroadcastMessage{
		Message: message,
		Sports:  sports,
	}
	
	select {
	case m.broadcast <- broadcastMsg:
	default:
		log.Println("Broadcast channel is full, message dropped")
	}
}

// BroadcastToUsers は指定されたユーザーにメッセージをブロードキャストする
func (m *Manager) BroadcastToUsers(message *models.WebSocketMessage, userIDs []int) {
	broadcastMsg := &BroadcastMessage{
		Message: message,
		UserIDs: userIDs,
	}
	
	select {
	case m.broadcast <- broadcastMsg:
	default:
		log.Println("Broadcast channel is full, message dropped")
	}
}

// BroadcastToAll は全ての接続にメッセージをブロードキャストする
func (m *Manager) BroadcastToAll(message *models.WebSocketMessage) {
	broadcastMsg := &BroadcastMessage{
		Message: message,
	}
	
	select {
	case m.broadcast <- broadcastMsg:
	default:
		log.Println("Broadcast channel is full, message dropped")
	}
}

// GetStats は統計情報を取得する
func (m *Manager) GetStats() *models.WebSocketStats {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// 現在の接続数を更新
	m.stats.ActiveConnections = len(m.connections)
	m.stats.UpdateStats()
	
	return m.stats
}

// GetConnections は現在の接続一覧を取得する
func (m *Manager) GetConnections() []*models.ConnectionInfo {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	connections := make([]*models.ConnectionInfo, 0, len(m.connections))
	for _, client := range m.connections {
		connections = append(connections, client.Info)
	}
	
	return connections
}

// healthCheck は定期的なヘルスチェックを実行する
func (m *Manager) healthCheck() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	
	// Pingメッセージを全クライアントに送信
	pingMsg, _ := models.NewWebSocketMessage(
		models.MessageTypePing.String(),
		map[string]interface{}{
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		},
	)
	
	messageBytes, _ := json.Marshal(pingMsg)
	
	for _, client := range m.connections {
		select {
		case client.Send <- messageBytes:
		default:
			// 送信できない場合は接続を閉じる
			log.Printf("Health check failed for client %s, closing connection", client.ID)
			client.cancel()
		}
	}
}

// sendMessage はクライアントにメッセージを送信する
func (c *Client) sendMessage(message *models.WebSocketMessage) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		log.Printf("Failed to marshal message: %v", err)
		return
	}
	
	select {
	case c.Send <- messageBytes:
	default:
		log.Printf("Send channel full for client %s", c.ID)
	}
}

// sendError はクライアントにエラーメッセージを送信する
func (c *Client) sendError(code, message string) {
	errorNotification := models.NewErrorNotification(code, message)
	errorMsg, _ := models.NewWebSocketMessage(
		models.MessageTypeError.String(),
		errorNotification,
	)
	c.sendMessage(errorMsg)
}
package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"backend/internal/models"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

// NotificationMiddleware はリアルタイム通知を自動送信するミドルウェア
type NotificationMiddleware struct {
	notificationService *service.NotificationService
}

// NewNotificationMiddleware は新しい通知ミドルウェアを作成する
func NewNotificationMiddleware(notificationService *service.NotificationService) *NotificationMiddleware {
	return &NotificationMiddleware{
		notificationService: notificationService,
	}
}

// AutoNotify は自動通知ミドルウェアを返す
func (m *NotificationMiddleware) AutoNotify() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 通知サービスが無効な場合はスキップ
		if m.notificationService == nil || !m.notificationService.IsEnabled() {
			c.Next()
			return
		}

		// 管理者操作のみ通知対象とする
		if !m.isAdminOperation(c) {
			c.Next()
			return
		}

		// リクエストボディを読み取り（必要に応じて）
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// レスポンスライターをラップして結果を監視
		writer := &responseWriter{
			ResponseWriter: c.Writer,
			body:          &bytes.Buffer{},
		}
		c.Writer = writer

		// 次のハンドラーを実行
		c.Next()

		// レスポンスが成功の場合のみ通知を送信
		if writer.Status() >= 200 && writer.Status() < 300 {
			m.sendNotificationBasedOnEndpoint(c, requestBody, writer.body.Bytes())
		}
	}
}

// isAdminOperation は管理者操作かどうかを判定する
func (m *NotificationMiddleware) isAdminOperation(c *gin.Context) bool {
	path := c.Request.URL.Path
	method := c.Request.Method

	// 管理者専用エンドポイントのパターン
	adminPatterns := []string{
		"/api/v1/admin/",
		"/api/tournaments", // 旧API（POST, PUT, DELETE）
		"/api/matches",     // 旧API（POST, PUT, DELETE）
	}

	for _, pattern := range adminPatterns {
		if strings.Contains(path, pattern) {
			// 管理者エンドポイントまたは変更操作
			if strings.Contains(path, "/admin/") || 
			   method == "POST" || method == "PUT" || method == "DELETE" {
				return true
			}
		}
	}

	return false
}

// sendNotificationBasedOnEndpoint はエンドポイントに基づいて通知を送信する
func (m *NotificationMiddleware) sendNotificationBasedOnEndpoint(c *gin.Context, requestBody, responseBody []byte) {
	path := c.Request.URL.Path
	method := c.Request.Method

	// トーナメント関連の通知
	if strings.Contains(path, "tournament") {
		m.handleTournamentNotification(c, method, path, requestBody, responseBody)
	}

	// 試合関連の通知
	if strings.Contains(path, "match") {
		m.handleMatchNotification(c, method, path, requestBody, responseBody)
	}
}

// handleTournamentNotification はトーナメント関連の通知を処理する
func (m *NotificationMiddleware) handleTournamentNotification(c *gin.Context, method, path string, requestBody, responseBody []byte) {
	var action string
	var sport models.SportType

	// アクションを決定
	switch method {
	case "POST":
		action = "created"
	case "PUT":
		if strings.Contains(path, "/format") {
			action = "format_changed"
		} else if strings.Contains(path, "/complete") {
			action = "completed"
		} else {
			action = "updated"
		}
	case "DELETE":
		action = "deleted"
	default:
		return
	}

	// スポーツを抽出（パスまたはレスポンスから）
	sport = m.extractSportFromPath(path)
	if sport == "" {
		sport = m.extractSportFromResponse(responseBody)
	}

	// システムメッセージを送信
	if sport != "" {
		message := m.generateTournamentMessage(action, sport)
		m.notificationService.NotifySystemMessage(message, []models.SportType{sport}, nil)
	}
}

// handleMatchNotification は試合関連の通知を処理する
func (m *NotificationMiddleware) handleMatchNotification(c *gin.Context, method, path string, requestBody, responseBody []byte) {
	var action string
	var sport models.SportType

	// アクションを決定
	switch method {
	case "POST":
		action = "created"
	case "PUT":
		if strings.Contains(path, "/result") {
			action = "result_updated"
		} else {
			action = "updated"
		}
	case "DELETE":
		action = "deleted"
	default:
		return
	}

	// スポーツを抽出
	sport = m.extractSportFromPath(path)
	if sport == "" {
		sport = m.extractSportFromResponse(responseBody)
	}

	// システムメッセージを送信
	if sport != "" {
		message := m.generateMatchMessage(action, sport)
		m.notificationService.NotifySystemMessage(message, []models.SportType{sport}, nil)
	}
}

// extractSportFromPath はパスからスポーツを抽出する
func (m *NotificationMiddleware) extractSportFromPath(path string) models.SportType {
	if strings.Contains(path, "/volleyball") {
		return models.SportTypeVolleyball
	}
	if strings.Contains(path, "/table_tennis") {
		return models.SportTypeTableTennis
	}
	if strings.Contains(path, "/soccer") {
		return models.SportTypeSoccer
	}
	return ""
}

// extractSportFromResponse はレスポンスからスポーツを抽出する
func (m *NotificationMiddleware) extractSportFromResponse(responseBody []byte) models.SportType {
	var response map[string]interface{}
	if err := json.Unmarshal(responseBody, &response); err != nil {
		return ""
	}

	// データフィールドからスポーツを抽出
	if data, ok := response["data"].(map[string]interface{}); ok {
		if sport, ok := data["sport"].(string); ok {
			return models.SportType(sport)
		}
	}

	return ""
}

// generateTournamentMessage はトーナメント用のメッセージを生成する
func (m *NotificationMiddleware) generateTournamentMessage(action string, sport models.SportType) string {
	sportName := m.getSportDisplayName(sport)
	
	switch action {
	case "created":
		return sportName + "のトーナメントが作成されました"
	case "updated":
		return sportName + "のトーナメント情報が更新されました"
	case "format_changed":
		return sportName + "のトーナメント形式が変更されました"
	case "completed":
		return sportName + "のトーナメントが完了しました"
	case "deleted":
		return sportName + "のトーナメントが削除されました"
	default:
		return sportName + "のトーナメントが変更されました"
	}
}

// generateMatchMessage は試合用のメッセージを生成する
func (m *NotificationMiddleware) generateMatchMessage(action string, sport models.SportType) string {
	sportName := m.getSportDisplayName(sport)
	
	switch action {
	case "created":
		return sportName + "の新しい試合が作成されました"
	case "updated":
		return sportName + "の試合情報が更新されました"
	case "result_updated":
		return sportName + "の試合結果が更新されました"
	case "deleted":
		return sportName + "の試合が削除されました"
	default:
		return sportName + "の試合が変更されました"
	}
}

// getSportDisplayName はスポーツの表示名を取得する
func (m *NotificationMiddleware) getSportDisplayName(sport models.SportType) string {
	switch sport {
	case models.SportTypeVolleyball:
		return "バレーボール"
	case models.SportTypeTableTennis:
		return "卓球"
	case models.SportTypeSoccer:
		return "サッカー"
	default:
		return "スポーツ"
	}
}

// responseWriter はレスポンスを監視するためのラッパー
type responseWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Status() int {
	if w.status == 0 {
		return 200
	}
	return w.status
}
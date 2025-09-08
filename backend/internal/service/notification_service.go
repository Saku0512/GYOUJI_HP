package service

import (
	"log"

	"backend/internal/models"
	websocketManager "backend/internal/websocket"
)

// NotificationService はリアルタイム通知を管理するサービス
type NotificationService struct {
	wsManager *websocketManager.Manager
}

// NewNotificationService は新しいNotificationServiceを作成する
func NewNotificationService(wsManager *websocketManager.Manager) *NotificationService {
	return &NotificationService{
		wsManager: wsManager,
	}
}

// NotifyTournamentUpdate はトーナメント更新を通知する
func (s *NotificationService) NotifyTournamentUpdate(tournament *models.Tournament, action string) {
	if s.wsManager == nil {
		return
	}

	// 更新データを作成
	updateData := &models.TournamentUpdateData{
		Tournament: tournament,
		Action:     action,
	}

	// 更新通知を作成
	notification := models.NewUpdateNotification(
		models.MessageTypeTournamentUpdate,
		tournament.Sport,
		updateData,
	)

	// WebSocketメッセージを作成
	wsMessage, err := models.NewWebSocketMessage(
		models.MessageTypeTournamentUpdate.String(),
		notification,
	)
	if err != nil {
		log.Printf("Failed to create tournament update message: %v", err)
		return
	}

	// 該当スポーツの購読者にブロードキャスト
	s.wsManager.BroadcastToSports(wsMessage, []models.SportType{tournament.Sport})

	log.Printf("Tournament update notification sent: sport=%s, action=%s, id=%d", 
		tournament.Sport, action, tournament.ID)
}

// NotifyMatchUpdate は試合更新を通知する
func (s *NotificationService) NotifyMatchUpdate(match *models.Match, action string) {
	if s.wsManager == nil {
		return
	}

	// 試合が属するトーナメントのスポーツを取得する必要がある
	// 簡易実装として、試合にスポーツ情報を含める
	// 実際の実装では、TournamentServiceから取得する
	sport := s.getMatchSport(match)
	if sport == "" {
		log.Printf("Failed to determine sport for match %d", match.ID)
		return
	}

	// 更新データを作成
	updateData := &models.MatchUpdateData{
		Match:  match,
		Action: action,
	}

	// 更新通知を作成
	notification := models.NewUpdateNotification(
		models.MessageTypeMatchUpdate,
		sport,
		updateData,
	)

	// WebSocketメッセージを作成
	wsMessage, err := models.NewWebSocketMessage(
		models.MessageTypeMatchUpdate.String(),
		notification,
	)
	if err != nil {
		log.Printf("Failed to create match update message: %v", err)
		return
	}

	// 該当スポーツの購読者にブロードキャスト
	s.wsManager.BroadcastToSports(wsMessage, []models.SportType{sport})

	log.Printf("Match update notification sent: sport=%s, action=%s, id=%d", 
		sport, action, match.ID)
}

// NotifyMatchResult は試合結果更新を通知する
func (s *NotificationService) NotifyMatchResult(match *models.Match) {
	if s.wsManager == nil {
		return
	}

	// 試合が属するトーナメントのスポーツを取得
	sport := s.getMatchSport(match)
	if sport == "" {
		log.Printf("Failed to determine sport for match %d", match.ID)
		return
	}

	// 更新データを作成
	updateData := &models.MatchUpdateData{
		Match:  match,
		Action: "result_updated",
	}

	// 更新通知を作成
	notification := models.NewUpdateNotification(
		models.MessageTypeMatchResult,
		sport,
		updateData,
	)

	// WebSocketメッセージを作成
	wsMessage, err := models.NewWebSocketMessage(
		models.MessageTypeMatchResult.String(),
		notification,
	)
	if err != nil {
		log.Printf("Failed to create match result message: %v", err)
		return
	}

	// 該当スポーツの購読者にブロードキャスト
	s.wsManager.BroadcastToSports(wsMessage, []models.SportType{sport})

	log.Printf("Match result notification sent: sport=%s, id=%d", sport, match.ID)
}

// NotifyBracketUpdate はブラケット更新を通知する
func (s *NotificationService) NotifyBracketUpdate(sport models.SportType, bracket *models.Bracket, action string) {
	if s.wsManager == nil {
		return
	}

	// 更新データを作成
	updateData := &models.BracketUpdateData{
		Sport:   sport,
		Bracket: bracket,
		Action:  action,
	}

	// 更新通知を作成
	notification := models.NewUpdateNotification(
		models.MessageTypeBracketUpdate,
		sport,
		updateData,
	)

	// WebSocketメッセージを作成
	wsMessage, err := models.NewWebSocketMessage(
		models.MessageTypeBracketUpdate.String(),
		notification,
	)
	if err != nil {
		log.Printf("Failed to create bracket update message: %v", err)
		return
	}

	// 該当スポーツの購読者にブロードキャスト
	s.wsManager.BroadcastToSports(wsMessage, []models.SportType{sport})

	log.Printf("Bracket update notification sent: sport=%s, action=%s", sport, action)
}

// NotifySystemMessage はシステムメッセージを通知する
func (s *NotificationService) NotifySystemMessage(message string, sports []models.SportType, userIDs []int) {
	if s.wsManager == nil {
		return
	}

	// システムメッセージデータを作成
	systemData := map[string]interface{}{
		"message":   message,
		"timestamp": models.Now().String(),
		"type":      "system",
	}

	// WebSocketメッセージを作成
	wsMessage, err := models.NewWebSocketMessage("system_message", systemData)
	if err != nil {
		log.Printf("Failed to create system message: %v", err)
		return
	}

	// ブロードキャスト
	if len(userIDs) > 0 {
		s.wsManager.BroadcastToUsers(wsMessage, userIDs)
	} else if len(sports) > 0 {
		s.wsManager.BroadcastToSports(wsMessage, sports)
	} else {
		s.wsManager.BroadcastToAll(wsMessage)
	}

	log.Printf("System message notification sent: %s", message)
}

// getMatchSport は試合のスポーツを取得する（簡易実装）
// 実際の実装では、TournamentServiceやRepositoryから取得する
func (s *NotificationService) getMatchSport(match *models.Match) models.SportType {
	// TODO: 実際の実装では、match.TournamentIDからトーナメント情報を取得してスポーツを判定する
	// 現在は簡易的な実装として、試合IDやトーナメントIDから推測する
	
	// 仮の実装：トーナメントIDの範囲でスポーツを判定
	switch {
	case match.TournamentID >= 1 && match.TournamentID <= 10:
		return models.SportTypeVolleyball
	case match.TournamentID >= 11 && match.TournamentID <= 20:
		return models.SportTypeTableTennis
	case match.TournamentID >= 21 && match.TournamentID <= 30:
		return models.SportTypeSoccer
	default:
		// デフォルトはバレーボール
		return models.SportTypeVolleyball
	}
}

// SetWebSocketManager はWebSocketマネージャーを設定する
func (s *NotificationService) SetWebSocketManager(wsManager *websocketManager.Manager) {
	s.wsManager = wsManager
}

// IsEnabled はNotificationServiceが有効かどうかを返す
func (s *NotificationService) IsEnabled() bool {
	return s.wsManager != nil
}

// GetStats は通知サービスの統計情報を取得する
func (s *NotificationService) GetStats() map[string]interface{} {
	if s.wsManager == nil {
		return map[string]interface{}{
			"enabled": false,
		}
	}

	wsStats := s.wsManager.GetStats()
	return map[string]interface{}{
		"enabled":             true,
		"active_connections":  wsStats.ActiveConnections,
		"total_connections":   wsStats.TotalConnections,
		"messages_sent":       wsStats.MessagesSent,
		"messages_received":   wsStats.MessagesReceived,
		"connections_by_sport": wsStats.ConnectionsBySport,
		"last_updated":        wsStats.LastUpdated,
	}
}
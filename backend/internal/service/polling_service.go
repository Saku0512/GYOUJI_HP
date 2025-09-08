package service

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
)

// PollingService はポーリング機能を提供するサービス
type PollingService struct {
	tournamentRepo repository.TournamentRepository
	matchRepo      repository.MatchRepository
	
	// キャッシュ管理
	cache      map[string]*CacheEntry
	cacheMutex sync.RWMutex
	
	// 設定
	cacheExpiry time.Duration
}

// CacheEntry はキャッシュエントリを表す
type CacheEntry struct {
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	ETag      string      `json:"etag"`
}

// PollingResponse はポーリングレスポンスを表す
type PollingResponse struct {
	HasUpdates bool        `json:"has_updates"`
	Data       interface{} `json:"data,omitempty"`
	ETag       string      `json:"etag"`
	Timestamp  string      `json:"timestamp"`
	NextPoll   int         `json:"next_poll_seconds"`
}

// UpdateCheckRequest はデータ更新チェックリクエストを表す
type UpdateCheckRequest struct {
	Sport     models.SportType `json:"sport" validate:"required,oneof=volleyball table_tennis soccer"`
	DataType  string           `json:"data_type" validate:"required,oneof=tournament matches bracket"`
	LastETag  string           `json:"last_etag,omitempty"`
	LastCheck string           `json:"last_check,omitempty"`
}

// NewPollingService は新しいポーリングサービスを作成する
func NewPollingService(
	tournamentRepo repository.TournamentRepository,
	matchRepo repository.MatchRepository,
) *PollingService {
	return &PollingService{
		tournamentRepo: tournamentRepo,
		matchRepo:      matchRepo,
		cache:          make(map[string]*CacheEntry),
		cacheExpiry:    30 * time.Second, // 30秒でキャッシュ期限切れ
	}
}

// CheckForUpdates はデータの更新をチェックする
func (s *PollingService) CheckForUpdates(ctx context.Context, request *UpdateCheckRequest) (*PollingResponse, error) {
	// バリデーション
	if !request.Sport.IsValid() {
		return nil, NewValidationError("無効なスポーツタイプです")
	}

	// キャッシュキーを生成
	cacheKey := s.generateCacheKey(request.Sport, request.DataType)

	// キャッシュから現在のデータを取得
	currentEntry := s.getCacheEntry(cacheKey)
	
	// クライアントのETagと比較
	if request.LastETag != "" && currentEntry != nil && currentEntry.ETag == request.LastETag {
		// データに変更なし
		return &PollingResponse{
			HasUpdates: false,
			ETag:       currentEntry.ETag,
			Timestamp:  currentEntry.Timestamp.UTC().Format(time.RFC3339),
			NextPoll:   s.calculateNextPollInterval(false),
		}, nil
	}

	// データを取得（キャッシュまたはDB）
	data, err := s.fetchData(ctx, request.Sport, request.DataType)
	if err != nil {
		return nil, err
	}

	// 新しいETagを生成
	etag := s.generateETag(data)

	// キャッシュを更新
	s.updateCache(cacheKey, data, etag)

	// レスポンスを作成
	response := &PollingResponse{
		HasUpdates: true,
		Data:       data,
		ETag:       etag,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		NextPoll:   s.calculateNextPollInterval(true),
	}

	return response, nil
}

// GetLatestData は最新のデータを取得する（強制更新）
func (s *PollingService) GetLatestData(ctx context.Context, sport models.SportType, dataType string) (*PollingResponse, error) {
	// バリデーション
	if !sport.IsValid() {
		return nil, NewValidationError("無効なスポーツタイプです")
	}

	// データを強制取得
	data, err := s.fetchData(ctx, sport, dataType)
	if err != nil {
		return nil, err
	}

	// ETagを生成
	etag := s.generateETag(data)

	// キャッシュを更新
	cacheKey := s.generateCacheKey(sport, dataType)
	s.updateCache(cacheKey, data, etag)

	return &PollingResponse{
		HasUpdates: true,
		Data:       data,
		ETag:       etag,
		Timestamp:  time.Now().UTC().Format(time.RFC3339),
		NextPoll:   s.calculateNextPollInterval(true),
	}, nil
}

// InvalidateCache は指定されたキャッシュを無効化する
func (s *PollingService) InvalidateCache(sport models.SportType, dataType string) {
	cacheKey := s.generateCacheKey(sport, dataType)
	
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	delete(s.cache, cacheKey)
}

// InvalidateAllCache は全てのキャッシュを無効化する
func (s *PollingService) InvalidateAllCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	s.cache = make(map[string]*CacheEntry)
}

// fetchData はデータベースからデータを取得する
func (s *PollingService) fetchData(ctx context.Context, sport models.SportType, dataType string) (interface{}, error) {
	switch dataType {
	case "tournament":
		return s.fetchTournamentData(ctx, sport)
	case "matches":
		return s.fetchMatchesData(ctx, sport)
	case "bracket":
		return s.fetchBracketData(ctx, sport)
	default:
		return nil, NewValidationError("無効なデータタイプです")
	}
}

// fetchTournamentData はトーナメントデータを取得する
func (s *PollingService) fetchTournamentData(ctx context.Context, sport models.SportType) (interface{}, error) {
	tournaments, err := s.tournamentRepo.GetBySport(ctx, sport.String(), 1, 0)
	if err != nil {
		return nil, NewDatabaseError("トーナメントデータの取得に失敗しました")
	}
	
	if len(tournaments) == 0 {
		return nil, NewNotFoundError("トーナメントが見つかりません")
	}
	
	return tournaments[0], nil
}

// fetchMatchesData は試合データを取得する
func (s *PollingService) fetchMatchesData(ctx context.Context, sport models.SportType) (interface{}, error) {
	// スポーツに対応するトーナメントを取得
	tournaments, err := s.tournamentRepo.GetBySport(ctx, sport.String(), 1, 0)
	if err != nil {
		return nil, NewDatabaseError("トーナメントデータの取得に失敗しました")
	}
	
	if len(tournaments) == 0 {
		return []interface{}{}, nil // 空の配列を返す
	}
	
	// 試合データを取得
	matches, err := s.matchRepo.GetByTournamentID(ctx, uint(tournaments[0].ID))
	if err != nil {
		return nil, NewDatabaseError("試合データの取得に失敗しました")
	}
	
	return matches, nil
}

// fetchBracketData はブラケットデータを取得する
func (s *PollingService) fetchBracketData(ctx context.Context, sport models.SportType) (interface{}, error) {
	// 試合データを取得してブラケット形式に整形
	matchesData, err := s.fetchMatchesData(ctx, sport)
	if err != nil {
		return nil, err
	}
	
	matches, ok := matchesData.([]*models.Match)
	if !ok {
		return nil, NewDatabaseError("試合データの形式が無効です")
	}
	
	// ブラケット形式に変換（簡易実装）
	bracket := s.convertToBracket(matches)
	return bracket, nil
}

// convertToBracket は試合データをブラケット形式に変換する
func (s *PollingService) convertToBracket(matches []*models.Match) *models.Bracket {
	// 簡易的なブラケット作成
	bracket := &models.Bracket{
		Sport:   models.SportTypeVolleyball, // 実際にはパラメータから取得
		Rounds:  make(map[string][]*models.Match),
		Updated: models.Now(),
	}
	
	// ラウンド別に試合を分類
	for _, match := range matches {
		if bracket.Rounds[match.Round] == nil {
			bracket.Rounds[match.Round] = make([]*models.Match, 0)
		}
		bracket.Rounds[match.Round] = append(bracket.Rounds[match.Round], match)
	}
	
	return bracket
}

// generateCacheKey はキャッシュキーを生成する
func (s *PollingService) generateCacheKey(sport models.SportType, dataType string) string {
	return fmt.Sprintf("%s:%s", sport.String(), dataType)
}

// getCacheEntry はキャッシュエントリを取得する
func (s *PollingService) getCacheEntry(key string) *CacheEntry {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	entry, exists := s.cache[key]
	if !exists {
		return nil
	}
	
	// キャッシュの有効期限をチェック
	if time.Since(entry.Timestamp) > s.cacheExpiry {
		// 期限切れの場合は削除
		delete(s.cache, key)
		return nil
	}
	
	return entry
}

// updateCache はキャッシュを更新する
func (s *PollingService) updateCache(key string, data interface{}, etag string) {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	s.cache[key] = &CacheEntry{
		Data:      data,
		Timestamp: time.Now(),
		ETag:      etag,
	}
}

// generateETag はデータからETagを生成する
func (s *PollingService) generateETag(data interface{}) string {
	// データをJSONにシリアライズしてハッシュ化
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Sprintf("etag_%d", time.Now().Unix())
	}
	
	// 簡易的なハッシュ（実際の実装ではより堅牢なハッシュを使用）
	hash := 0
	for _, b := range jsonData {
		hash = hash*31 + int(b)
	}
	
	return fmt.Sprintf("etag_%x_%d", hash, time.Now().Unix())
}

// calculateNextPollInterval は次のポーリング間隔を計算する
func (s *PollingService) calculateNextPollInterval(hasUpdates bool) int {
	if hasUpdates {
		// 更新があった場合は短い間隔
		return 10 // 10秒
	}
	
	// 更新がない場合は長い間隔
	return 30 // 30秒
}

// GetCacheStats はキャッシュ統計を取得する
func (s *PollingService) GetCacheStats() map[string]interface{} {
	s.cacheMutex.RLock()
	defer s.cacheMutex.RUnlock()
	
	stats := make(map[string]interface{})
	stats["cache_entries"] = len(s.cache)
	stats["cache_expiry_seconds"] = int(s.cacheExpiry.Seconds())
	
	// キャッシュエントリの詳細
	entries := make([]map[string]interface{}, 0)
	for key, entry := range s.cache {
		entryInfo := map[string]interface{}{
			"key":       key,
			"timestamp": entry.Timestamp.UTC().Format(time.RFC3339),
			"etag":      entry.ETag,
			"age_seconds": int(time.Since(entry.Timestamp).Seconds()),
		}
		entries = append(entries, entryInfo)
	}
	stats["entries"] = entries
	
	return stats
}

// CleanupExpiredCache は期限切れのキャッシュをクリーンアップする
func (s *PollingService) CleanupExpiredCache() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()
	
	now := time.Now()
	for key, entry := range s.cache {
		if now.Sub(entry.Timestamp) > s.cacheExpiry {
			delete(s.cache, key)
		}
	}
}

// StartCacheCleanup は定期的なキャッシュクリーンアップを開始する
func (s *PollingService) StartCacheCleanup(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute) // 5分ごとにクリーンアップ
	defer ticker.Stop()
	
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.CleanupExpiredCache()
		}
	}
}
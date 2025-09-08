// Package cache はキャッシュ無効化戦略を提供する
package cache

import (
	"context"
	"log"
	"time"
)

// InvalidationStrategy はキャッシュ無効化戦略のインターフェース
type InvalidationStrategy interface {
	// データ更新時の無効化
	OnTournamentUpdate(ctx context.Context, sport string) error
	OnMatchUpdate(ctx context.Context, matchID int, sport string) error
	OnMatchResultUpdate(ctx context.Context, matchID int, sport string) error
	
	// 定期的な無効化
	SchedulePeriodicInvalidation(ctx context.Context, interval time.Duration)
	
	// 手動無効化
	InvalidateAll(ctx context.Context) error
	InvalidateBySport(ctx context.Context, sport string) error
}

// invalidationStrategyImpl はInvalidationStrategyの実装
type invalidationStrategyImpl struct {
	cache CacheManager
}

// NewInvalidationStrategy は新しい無効化戦略インスタンスを作成する
func NewInvalidationStrategy(cache CacheManager) InvalidationStrategy {
	return &invalidationStrategyImpl{
		cache: cache,
	}
}

// OnTournamentUpdate はトーナメント更新時のキャッシュ無効化を実行する
func (s *invalidationStrategyImpl) OnTournamentUpdate(ctx context.Context, sport string) error {
	log.Printf("トーナメント更新によるキャッシュ無効化: %s", sport)
	
	// トーナメント関連の全キャッシュを無効化
	if err := s.cache.InvalidateTournamentCache(ctx, sport); err != nil {
		log.Printf("トーナメントキャッシュ無効化エラー: %v", err)
		return err
	}
	
	// 統計キャッシュも無効化
	statsKeys := []string{
		sport + "_progress",
		sport + "_statistics",
		"tournament_list",
	}
	
	for _, key := range statsKeys {
		if err := s.cache.DeleteStatistics(ctx, key); err != nil {
			log.Printf("統計キャッシュ無効化エラー (%s): %v", key, err)
		}
	}
	
	return nil
}

// OnMatchUpdate は試合更新時のキャッシュ無効化を実行する
func (s *invalidationStrategyImpl) OnMatchUpdate(ctx context.Context, matchID int, sport string) error {
	log.Printf("試合更新によるキャッシュ無効化: match_id=%d, sport=%s", matchID, sport)
	
	// 試合関連キャッシュを無効化
	if err := s.cache.DeleteMatches(ctx, sport); err != nil {
		log.Printf("試合キャッシュ無効化エラー: %v", err)
		return err
	}
	
	// ブラケットキャッシュも無効化（試合更新でブラケット構造が変わる可能性）
	if err := s.cache.DeleteBracket(ctx, sport); err != nil {
		log.Printf("ブラケットキャッシュ無効化エラー: %v", err)
	}
	
	// 統計キャッシュも無効化
	statsKeys := []string{
		sport + "_statistics",
		sport + "_progress",
		"match_statistics",
	}
	
	for _, key := range statsKeys {
		if err := s.cache.DeleteStatistics(ctx, key); err != nil {
			log.Printf("統計キャッシュ無効化エラー (%s): %v", key, err)
		}
	}
	
	return nil
}

// OnMatchResultUpdate は試合結果更新時のキャッシュ無効化を実行する
func (s *invalidationStrategyImpl) OnMatchResultUpdate(ctx context.Context, matchID int, sport string) error {
	log.Printf("試合結果更新によるキャッシュ無効化: match_id=%d, sport=%s", matchID, sport)
	
	// 試合結果更新は特に重要なので、関連する全キャッシュを無効化
	if err := s.cache.InvalidateTournamentCache(ctx, sport); err != nil {
		log.Printf("トーナメントキャッシュ無効化エラー: %v", err)
		return err
	}
	
	// 統計キャッシュを無効化
	statsKeys := []string{
		sport + "_statistics",
		sport + "_progress",
		"match_statistics",
		"team_statistics",
		"tournament_progress",
	}
	
	for _, key := range statsKeys {
		if err := s.cache.DeleteStatistics(ctx, key); err != nil {
			log.Printf("統計キャッシュ無効化エラー (%s): %v", key, err)
		}
	}
	
	return nil
}

// SchedulePeriodicInvalidation は定期的なキャッシュ無効化をスケジュールする
func (s *invalidationStrategyImpl) SchedulePeriodicInvalidation(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	
	go func() {
		defer ticker.Stop()
		
		for {
			select {
			case <-ctx.Done():
				log.Println("定期キャッシュ無効化を停止します")
				return
			case <-ticker.C:
				log.Println("定期キャッシュ無効化を実行します")
				
				// 統計キャッシュのみ定期的に無効化（データの整合性を保つため）
				statsKeys := []string{
					"volleyball_statistics",
					"table_tennis_statistics", 
					"soccer_statistics",
					"tournament_progress",
					"match_statistics",
				}
				
				for _, key := range statsKeys {
					if err := s.cache.DeleteStatistics(ctx, key); err != nil {
						log.Printf("定期統計キャッシュ無効化エラー (%s): %v", key, err)
					}
				}
			}
		}
	}()
	
	log.Printf("定期キャッシュ無効化をスケジュールしました (間隔: %v)", interval)
}

// InvalidateAll は全キャッシュを無効化する
func (s *invalidationStrategyImpl) InvalidateAll(ctx context.Context) error {
	log.Println("全キャッシュ無効化を実行します")
	
	if err := s.cache.InvalidateAllCache(ctx); err != nil {
		log.Printf("全キャッシュ無効化エラー: %v", err)
		return err
	}
	
	log.Println("全キャッシュ無効化が完了しました")
	return nil
}

// InvalidateBySport はスポーツ別キャッシュを無効化する
func (s *invalidationStrategyImpl) InvalidateBySport(ctx context.Context, sport string) error {
	log.Printf("スポーツ別キャッシュ無効化を実行します: %s", sport)
	
	if err := s.cache.InvalidateTournamentCache(ctx, sport); err != nil {
		log.Printf("スポーツ別キャッシュ無効化エラー: %v", err)
		return err
	}
	
	// 統計キャッシュも無効化
	statsKeys := []string{
		sport + "_statistics",
		sport + "_progress",
	}
	
	for _, key := range statsKeys {
		if err := s.cache.DeleteStatistics(ctx, key); err != nil {
			log.Printf("統計キャッシュ無効化エラー (%s): %v", key, err)
		}
	}
	
	log.Printf("スポーツ別キャッシュ無効化が完了しました: %s", sport)
	return nil
}

// CacheWarmer はキャッシュのウォームアップ機能を提供する
type CacheWarmer interface {
	WarmupTournamentCache(ctx context.Context, sport string) error
	WarmupAllCache(ctx context.Context) error
}

// cacheWarmerImpl はCacheWarmerの実装
type cacheWarmerImpl struct {
	cache           CacheManager
	tournamentRepo  interface{ GetBySport(ctx context.Context, sport string) (*models.Tournament, error) }
	matchRepo       interface{ GetBySport(ctx context.Context, sport string) ([]models.Match, error) }
}

// NewCacheWarmer は新しいキャッシュウォーマーインスタンスを作成する
func NewCacheWarmer(
	cache CacheManager,
	tournamentRepo interface{ GetBySport(ctx context.Context, sport string) (*models.Tournament, error) },
	matchRepo interface{ GetBySport(ctx context.Context, sport string) ([]models.Match, error) },
) CacheWarmer {
	return &cacheWarmerImpl{
		cache:          cache,
		tournamentRepo: tournamentRepo,
		matchRepo:      matchRepo,
	}
}

// WarmupTournamentCache はスポーツ別キャッシュをウォームアップする
func (w *cacheWarmerImpl) WarmupTournamentCache(ctx context.Context, sport string) error {
	log.Printf("キャッシュウォームアップを開始します: %s", sport)
	
	// トーナメントデータをプリロード
	if tournament, err := w.tournamentRepo.GetBySport(ctx, sport); err == nil {
		if err := w.cache.SetTournament(ctx, sport, tournament); err != nil {
			log.Printf("トーナメントキャッシュウォームアップエラー: %v", err)
		}
	}
	
	// 試合データをプリロード
	if matches, err := w.matchRepo.GetBySport(ctx, sport); err == nil {
		if err := w.cache.SetMatches(ctx, sport, matches); err != nil {
			log.Printf("試合キャッシュウォームアップエラー: %v", err)
		}
	}
	
	log.Printf("キャッシュウォームアップが完了しました: %s", sport)
	return nil
}

// WarmupAllCache は全スポーツのキャッシュをウォームアップする
func (w *cacheWarmerImpl) WarmupAllCache(ctx context.Context) error {
	log.Println("全キャッシュウォームアップを開始します")
	
	sports := []string{"volleyball", "table_tennis", "soccer"}
	
	for _, sport := range sports {
		if err := w.WarmupTournamentCache(ctx, sport); err != nil {
			log.Printf("キャッシュウォームアップエラー (%s): %v", sport, err)
		}
	}
	
	log.Println("全キャッシュウォームアップが完了しました")
	return nil
}
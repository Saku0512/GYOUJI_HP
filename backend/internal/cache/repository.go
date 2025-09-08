// Package cache はキャッシュ対応リポジトリラッパーを提供する
package cache

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"

	"backend/internal/models"
	"backend/internal/repository"
)

// CachedTournamentRepository はキャッシュ機能付きトーナメントリポジトリ
type CachedTournamentRepository struct {
	repo  repository.TournamentRepository
	cache CacheManager
}

// NewCachedTournamentRepository は新しいキャッシュ付きトーナメントリポジトリを作成する
func NewCachedTournamentRepository(repo repository.TournamentRepository, cache CacheManager) repository.TournamentRepository {
	return &CachedTournamentRepository{
		repo:  repo,
		cache: cache,
	}
}

// GetBySport はスポーツ別トーナメントを取得する（キャッシュ対応）
func (r *CachedTournamentRepository) GetBySport(ctx context.Context, sport string) (*models.Tournament, error) {
	// キャッシュから取得を試行
	if tournament, err := r.cache.GetTournament(ctx, sport); err == nil {
		log.Printf("トーナメントキャッシュヒット: %s", sport)
		return tournament, nil
	} else if err != redis.Nil {
		log.Printf("キャッシュ取得エラー: %v", err)
	}

	// キャッシュミスの場合はDBから取得
	tournament, err := r.repo.GetBySport(ctx, sport)
	if err != nil {
		return nil, err
	}

	// 取得したデータをキャッシュに保存
	if err := r.cache.SetTournament(ctx, sport, tournament); err != nil {
		log.Printf("トーナメントキャッシュ保存エラー: %v", err)
	}

	return tournament, nil
}

// Update はトーナメントを更新し、キャッシュを無効化する
func (r *CachedTournamentRepository) Update(ctx context.Context, tournament *models.Tournament) error {
	// DBを更新
	if err := r.repo.Update(ctx, tournament); err != nil {
		return err
	}

	// 関連キャッシュを無効化
	if err := r.cache.InvalidateTournamentCache(ctx, tournament.Sport); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}

	return nil
}

// Create はトーナメントを作成し、キャッシュを無効化する
func (r *CachedTournamentRepository) Create(ctx context.Context, tournament *models.Tournament) error {
	// DBに作成
	if err := r.repo.Create(ctx, tournament); err != nil {
		return err
	}

	// 関連キャッシュを無効化
	if err := r.cache.InvalidateTournamentCache(ctx, tournament.Sport); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}

	return nil
}

// Delete はトーナメントを削除し、キャッシュを無効化する
func (r *CachedTournamentRepository) Delete(ctx context.Context, id int) error {
	// 削除前にスポーツ情報を取得
	tournament, err := r.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// DBから削除
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}

	// 関連キャッシュを無効化
	if err := r.cache.InvalidateTournamentCache(ctx, tournament.Sport); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}

	return nil
}

// 他のメソッドは元のリポジトリに委譲
func (r *CachedTournamentRepository) GetAll(ctx context.Context) ([]models.Tournament, error) {
	return r.repo.GetAll(ctx)
}

func (r *CachedTournamentRepository) GetByID(ctx context.Context, id int) (*models.Tournament, error) {
	return r.repo.GetByID(ctx, id)
}

func (r *CachedTournamentRepository) UpdateFormat(ctx context.Context, sport, format string) error {
	if err := r.repo.UpdateFormat(ctx, sport, format); err != nil {
		return err
	}
	
	// キャッシュを無効化
	if err := r.cache.InvalidateTournamentCache(ctx, sport); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}
	
	return nil
}

func (r *CachedTournamentRepository) UpdateStatus(ctx context.Context, sport, status string) error {
	if err := r.repo.UpdateStatus(ctx, sport, status); err != nil {
		return err
	}
	
	// キャッシュを無効化
	if err := r.cache.InvalidateTournamentCache(ctx, sport); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}
	
	return nil
}

// CachedMatchRepository はキャッシュ機能付き試合リポジトリ
type CachedMatchRepository struct {
	repo  repository.MatchRepository
	cache CacheManager
}

// NewCachedMatchRepository は新しいキャッシュ付き試合リポジトリを作成する
func NewCachedMatchRepository(repo repository.MatchRepository, cache CacheManager) repository.MatchRepository {
	return &CachedMatchRepository{
		repo:  repo,
		cache: cache,
	}
}

// GetBySport はスポーツ別試合一覧を取得する（キャッシュ対応）
func (r *CachedMatchRepository) GetBySport(ctx context.Context, sport string) ([]models.Match, error) {
	// キャッシュから取得を試行
	if matches, err := r.cache.GetMatches(ctx, sport); err == nil {
		log.Printf("試合キャッシュヒット: %s", sport)
		return matches, nil
	} else if err != redis.Nil {
		log.Printf("キャッシュ取得エラー: %v", err)
	}

	// キャッシュミスの場合はDBから取得
	matches, err := r.repo.GetBySport(ctx, sport)
	if err != nil {
		return nil, err
	}

	// 取得したデータをキャッシュに保存
	if err := r.cache.SetMatches(ctx, sport, matches); err != nil {
		log.Printf("試合キャッシュ保存エラー: %v", err)
	}

	return matches, nil
}

// Update は試合を更新し、キャッシュを無効化する
func (r *CachedMatchRepository) Update(ctx context.Context, match *models.Match) error {
	// 更新前にトーナメント情報を取得してスポーツを特定
	oldMatch, err := r.repo.GetByID(ctx, match.ID)
	if err != nil {
		return err
	}

	// DBを更新
	if err := r.repo.Update(ctx, match); err != nil {
		return err
	}

	// トーナメントIDからスポーツを特定してキャッシュを無効化
	// 注: 実際の実装では、トーナメントリポジトリを使ってスポーツを取得する必要がある
	// ここでは簡略化のため、試合キャッシュのみ無効化
	if err := r.cache.DeleteMatches(ctx, "all"); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}

	return nil
}

// Create は試合を作成し、キャッシュを無効化する
func (r *CachedMatchRepository) Create(ctx context.Context, match *models.Match) error {
	// DBに作成
	if err := r.repo.Create(ctx, match); err != nil {
		return err
	}

	// 全試合キャッシュを無効化（スポーツ特定が困難なため）
	if err := r.cache.DeleteMatches(ctx, "all"); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}

	return nil
}

// UpdateResult は試合結果を更新し、キャッシュを無効化する
func (r *CachedMatchRepository) UpdateResult(ctx context.Context, matchID int, score1, score2 int, winner string) error {
	// DBを更新
	if err := r.repo.UpdateResult(ctx, matchID, score1, score2, winner); err != nil {
		return err
	}

	// 全試合キャッシュを無効化
	if err := r.cache.DeleteMatches(ctx, "all"); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}

	return nil
}

// 他のメソッドは元のリポジトリに委譲
func (r *CachedMatchRepository) GetAll(ctx context.Context) ([]models.Match, error) {
	return r.repo.GetAll(ctx)
}

func (r *CachedMatchRepository) GetByID(ctx context.Context, id int) (*models.Match, error) {
	return r.repo.GetByID(ctx, id)
}

func (r *CachedMatchRepository) GetByTournamentID(ctx context.Context, tournamentID int) ([]models.Match, error) {
	return r.repo.GetByTournamentID(ctx, tournamentID)
}

func (r *CachedMatchRepository) Delete(ctx context.Context, id int) error {
	if err := r.repo.Delete(ctx, id); err != nil {
		return err
	}
	
	// 全試合キャッシュを無効化
	if err := r.cache.DeleteMatches(ctx, "all"); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}
	
	return nil
}

func (r *CachedMatchRepository) GetByRound(ctx context.Context, tournamentID int, round string) ([]models.Match, error) {
	return r.repo.GetByRound(ctx, tournamentID, round)
}

func (r *CachedMatchRepository) GetByStatus(ctx context.Context, status string) ([]models.Match, error) {
	return r.repo.GetByStatus(ctx, status)
}

func (r *CachedMatchRepository) UpdateStatus(ctx context.Context, matchID int, status string) error {
	if err := r.repo.UpdateStatus(ctx, matchID, status); err != nil {
		return err
	}
	
	// 全試合キャッシュを無効化
	if err := r.cache.DeleteMatches(ctx, "all"); err != nil {
		log.Printf("キャッシュ無効化エラー: %v", err)
	}
	
	return nil
}
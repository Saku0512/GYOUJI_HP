// Package pagination はページネーション対応リポジトリ拡張を提供する
package pagination

import (
	"context"
	"database/sql"

	"backend/internal/models"
	"backend/internal/repository"
)

// PaginatedTournamentRepository はページネーション機能付きトーナメントリポジトリ
type PaginatedTournamentRepository interface {
	repository.TournamentRepository
	
	// ページネーション付きメソッド
	GetAllPaginated(ctx context.Context, req *models.PaginationRequest) (*PaginatedQuery[*models.Tournament], error)
	GetByFilterPaginated(ctx context.Context, filter *models.FilterRequest, req *models.PaginationRequest) (*PaginatedQuery[*models.Tournament], error)
}

// paginatedTournamentRepositoryImpl はPaginatedTournamentRepositoryの実装
type paginatedTournamentRepositoryImpl struct {
	repository.TournamentRepository
	db interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
	}
}

// NewPaginatedTournamentRepository は新しいページネーション付きトーナメントリポジトリを作成する
func NewPaginatedTournamentRepository(
	repo repository.TournamentRepository,
	db interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
	},
) PaginatedTournamentRepository {
	return &paginatedTournamentRepositoryImpl{
		TournamentRepository: repo,
		db:                   db,
	}
}

// GetAllPaginated は全トーナメントをページネーション付きで取得する
func (r *paginatedTournamentRepositoryImpl) GetAllPaginated(
	ctx context.Context,
	req *models.PaginationRequest,
) (*PaginatedQuery[*models.Tournament], error) {
	
	baseQuery := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		ORDER BY created_at DESC
	`
	
	return ExecutePaginatedTournamentQuery(ctx, r.db, baseQuery, req)
}

// GetByFilterPaginated はフィルター条件付きでトーナメントをページネーション取得する
func (r *paginatedTournamentRepositoryImpl) GetByFilterPaginated(
	ctx context.Context,
	filter *models.FilterRequest,
	req *models.PaginationRequest,
) (*PaginatedQuery[*models.Tournament], error) {
	
	baseQuery := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		WHERE 1=1
	`
	
	var args []interface{}
	argIndex := 1
	
	// フィルター条件を追加
	if filter != nil {
		if filter.Sport != nil {
			baseQuery += " AND sport = ?"
			args = append(args, string(*filter.Sport))
			argIndex++
		}
		
		if filter.Status != nil {
			baseQuery += " AND status = ?"
			args = append(args, string(*filter.Status))
			argIndex++
		}
	}
	
	baseQuery += " ORDER BY created_at DESC"
	
	return ExecutePaginatedTournamentQuery(ctx, r.db, baseQuery, req, args...)
}

// PaginatedMatchRepository はページネーション機能付き試合リポジトリ
type PaginatedMatchRepository interface {
	repository.MatchRepository
	
	// ページネーション付きメソッド
	GetAllPaginated(ctx context.Context, req *models.PaginationRequest) (*PaginatedQuery[*models.Match], error)
	GetBySportPaginated(ctx context.Context, sport string, req *models.PaginationRequest) (*PaginatedQuery[*models.Match], error)
	GetByFilterPaginated(ctx context.Context, filter *models.MatchFilterRequest, req *models.PaginationRequest) (*PaginatedQuery[*models.Match], error)
}

// paginatedMatchRepositoryImpl はPaginatedMatchRepositoryの実装
type paginatedMatchRepositoryImpl struct {
	repository.MatchRepository
	db interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
	}
}

// NewPaginatedMatchRepository は新しいページネーション付き試合リポジトリを作成する
func NewPaginatedMatchRepository(
	repo repository.MatchRepository,
	db interface {
		Query(query string, args ...interface{}) (*sql.Rows, error)
		QueryRow(query string, args ...interface{}) *sql.Row
	},
) PaginatedMatchRepository {
	return &paginatedMatchRepositoryImpl{
		MatchRepository: repo,
		db:              db,
	}
}

// GetAllPaginated は全試合をページネーション付きで取得する
func (r *paginatedMatchRepositoryImpl) GetAllPaginated(
	ctx context.Context,
	req *models.PaginationRequest,
) (*PaginatedQuery[*models.Match], error) {
	
	baseQuery := `
		SELECT m.id, m.tournament_id, m.round, m.team1, m.team2, 
		       m.score1, m.score2, m.winner, m.status, 
		       m.scheduled_at, m.completed_at, m.created_at, m.updated_at
		FROM matches m
		ORDER BY m.scheduled_at DESC, m.created_at DESC
	`
	
	return ExecutePaginatedMatchQuery(ctx, r.db, baseQuery, req)
}

// GetBySportPaginated はスポーツ別試合をページネーション付きで取得する
func (r *paginatedMatchRepositoryImpl) GetBySportPaginated(
	ctx context.Context,
	sport string,
	req *models.PaginationRequest,
) (*PaginatedQuery[*models.Match], error) {
	
	baseQuery := `
		SELECT m.id, m.tournament_id, m.round, m.team1, m.team2, 
		       m.score1, m.score2, m.winner, m.status, 
		       m.scheduled_at, m.completed_at, m.created_at, m.updated_at
		FROM matches m
		JOIN tournaments t ON m.tournament_id = t.id
		WHERE t.sport = ?
		ORDER BY m.scheduled_at DESC, m.created_at DESC
	`
	
	return ExecutePaginatedMatchQuery(ctx, r.db, baseQuery, req, sport)
}

// GetByFilterPaginated はフィルター条件付きで試合をページネーション取得する
func (r *paginatedMatchRepositoryImpl) GetByFilterPaginated(
	ctx context.Context,
	filter *models.MatchFilterRequest,
	req *models.PaginationRequest,
) (*PaginatedQuery[*models.Match], error) {
	
	baseQuery := `
		SELECT m.id, m.tournament_id, m.round, m.team1, m.team2, 
		       m.score1, m.score2, m.winner, m.status, 
		       m.scheduled_at, m.completed_at, m.created_at, m.updated_at
		FROM matches m
		JOIN tournaments t ON m.tournament_id = t.id
		WHERE 1=1
	`
	
	var args []interface{}
	
	// フィルター条件を追加
	if filter != nil {
		if filter.Sport != nil {
			baseQuery += " AND t.sport = ?"
			args = append(args, string(*filter.Sport))
		}
		
		if filter.Status != nil {
			baseQuery += " AND m.status = ?"
			args = append(args, string(*filter.Status))
		}
		
		if filter.Round != nil {
			baseQuery += " AND m.round = ?"
			args = append(args, string(*filter.Round))
		}
		
		if filter.TournamentID != nil {
			baseQuery += " AND m.tournament_id = ?"
			args = append(args, *filter.TournamentID)
		}
	}
	
	baseQuery += " ORDER BY m.scheduled_at DESC, m.created_at DESC"
	
	return ExecutePaginatedMatchQuery(ctx, r.db, baseQuery, req, args...)
}
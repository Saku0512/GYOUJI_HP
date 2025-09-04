// Package repository はデータアクセス層の実装を提供する
package repository

import (
	"backend/internal/database"
)

// Repository は全てのリポジトリインターフェースを統合する
type Repository struct {
	Base       BaseRepository
	User       UserRepository
	Tournament TournamentRepository
	Match      MatchRepository
}

// NewRepository は新しいRepositoryインスタンスを作成する
func NewRepository(db *database.DB) *Repository {
	baseRepo := NewBaseRepository(db)
	userRepo := NewUserRepository(db)
	tournamentRepo := NewTournamentRepository(db)
	matchRepo := NewMatchRepository(db)
	
	return &Repository{
		Base:       baseRepo,
		User:       userRepo,
		Tournament: tournamentRepo,
		Match:      matchRepo,
	}
}

// Close は全てのリポジトリのリソースを解放する
func (r *Repository) Close() error {
	if r.Base != nil {
		return r.Base.Close()
	}
	return nil
}

// HealthCheck は全てのリポジトリの健全性をチェックする
func (r *Repository) HealthCheck() error {
	if r.Base != nil {
		return r.Base.Ping()
	}
	return NewRepositoryError(ErrTypeConnection, "ベースリポジトリが初期化されていません", nil)
}
// Package repository はデータアクセス層の実装を提供する
package repository

import (
	"backend/internal/database"
)

// Repository は全てのリポジトリインターフェースを統合する
type Repository struct {
	Base BaseRepository
	// 将来的に追加される他のリポジトリ
	// User UserRepository
	// Tournament TournamentRepository
	// Match MatchRepository
}

// NewRepository は新しいRepositoryインスタンスを作成する
func NewRepository(db *database.DB) *Repository {
	baseRepo := NewBaseRepository(db)
	
	return &Repository{
		Base: baseRepo,
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
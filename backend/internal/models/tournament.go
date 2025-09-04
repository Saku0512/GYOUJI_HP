package models

import (
	"errors"
	"strings"
	"time"
)

// Tournament はトーナメントを表すモデル
type Tournament struct {
	ID        int       `json:"id" db:"id"`
	Sport     string    `json:"sport" db:"sport"`
	Format    string    `json:"format" db:"format"` // 卓球の場合"standard", "rainy"
	Status    string    `json:"status" db:"status"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Validate はトーナメントデータの検証を行う
func (t *Tournament) Validate() error {
	if strings.TrimSpace(t.Sport) == "" {
		return errors.New("スポーツは必須です")
	}
	
	if !IsValidSport(t.Sport) {
		return errors.New("無効なスポーツです")
	}
	
	if strings.TrimSpace(t.Format) == "" {
		return errors.New("フォーマットは必須です")
	}
	
	if !IsValidTournamentFormat(t.Format) {
		return errors.New("無効なトーナメントフォーマットです")
	}
	
	if strings.TrimSpace(t.Status) == "" {
		return errors.New("ステータスは必須です")
	}
	
	if !IsValidTournamentStatus(t.Status) {
		return errors.New("無効なトーナメントステータスです")
	}
	
	return nil
}

// IsActive はトーナメントがアクティブかどうかを返す
func (t *Tournament) IsActive() bool {
	return t.Status == TournamentStatusActive
}

// IsCompleted はトーナメントが完了しているかどうかを返す
func (t *Tournament) IsCompleted() bool {
	return t.Status == TournamentStatusCompleted
}
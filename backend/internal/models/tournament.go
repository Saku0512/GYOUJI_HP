package models

import (
	"errors"
	"strings"
	"time"
)

// Tournament はトーナメントを表すモデル
type Tournament struct {
	ID        int       `json:"id" db:"id"`
	Sport     string    `json:"sport" db:"sport"`         // データベース互換性のため文字列型を維持
	Format    string    `json:"format" db:"format"`       // データベース互換性のため文字列型を維持
	Status    string    `json:"status" db:"status"`       // データベース互換性のため文字列型を維持
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// GetSportType はSportType列挙型を返す
func (t *Tournament) GetSportType() SportType {
	return SportType(t.Sport)
}

// SetSportType はSportType列挙型から文字列を設定する
func (t *Tournament) SetSportType(sport SportType) {
	t.Sport = string(sport)
}

// GetFormat はTournamentFormat列挙型を返す
func (t *Tournament) GetFormat() TournamentFormat {
	return TournamentFormat(t.Format)
}

// SetFormat はTournamentFormat列挙型から文字列を設定する
func (t *Tournament) SetFormat(format TournamentFormat) {
	t.Format = string(format)
}

// GetStatus はTournamentStatus列挙型を返す
func (t *Tournament) GetStatus() TournamentStatus {
	return TournamentStatus(t.Status)
}

// SetStatus はTournamentStatus列挙型から文字列を設定する
func (t *Tournament) SetStatus(status TournamentStatus) {
	t.Status = string(status)
}

// GetCreatedAt はDateTime型で作成日時を返す
func (t *Tournament) GetCreatedAt() DateTime {
	return NewDateTime(t.CreatedAt)
}

// GetUpdatedAt はDateTime型で更新日時を返す
func (t *Tournament) GetUpdatedAt() DateTime {
	return NewDateTime(t.UpdatedAt)
}

// Validate はトーナメントデータの検証を行う
func (t *Tournament) Validate() error {
	if strings.TrimSpace(t.Sport) == "" {
		return errors.New("スポーツは必須です")
	}
	
	if !t.GetSportType().IsValid() {
		return errors.New("無効なスポーツです")
	}
	
	if strings.TrimSpace(t.Format) == "" {
		return errors.New("フォーマットは必須です")
	}
	
	if !t.GetFormat().IsValid() {
		return errors.New("無効なトーナメントフォーマットです")
	}
	
	if strings.TrimSpace(t.Status) == "" {
		return errors.New("ステータスは必須です")
	}
	
	if !t.GetStatus().IsValid() {
		return errors.New("無効なトーナメントステータスです")
	}
	
	return nil
}

// IsActive はトーナメントがアクティブかどうかを返す
func (t *Tournament) IsActive() bool {
	return t.GetStatus() == TournamentStatusActiveEnum
}

// IsCompleted はトーナメントが完了しているかどうかを返す
func (t *Tournament) IsCompleted() bool {
	return t.GetStatus() == TournamentStatusCompletedEnum
}

// IsRegistration はトーナメントが登録中かどうかを返す
func (t *Tournament) IsRegistration() bool {
	return t.GetStatus() == TournamentStatusRegistrationEnum
}

// IsCancelled はトーナメントがキャンセルされているかどうかを返す
func (t *Tournament) IsCancelled() bool {
	return t.GetStatus() == TournamentStatusCancelledEnum
}
package models

import (
	"errors"
	"strings"
	"time"
)

// Match は試合を表すモデル
// @Description 試合を表すモデル
type Match struct {
	ID           int        `json:"id" db:"id"`
	TournamentID int        `json:"tournament_id" db:"tournament_id"`
	Round        string     `json:"round" db:"round"`        // データベース互換性のため文字列型を維持
	Team1        string     `json:"team1" db:"team1"`
	Team2        string     `json:"team2" db:"team2"`
	Score1       *int       `json:"score1,omitempty" db:"score1"` // 試合が行われるまでnull
	Score2       *int       `json:"score2,omitempty" db:"score2"`
	Winner       *string    `json:"winner,omitempty" db:"winner"`
	Status       string     `json:"status" db:"status"`           // データベース互換性のため文字列型を維持
	ScheduledAt  time.Time  `json:"scheduled_at" db:"scheduled_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty" db:"completed_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

// GetRound はRoundType列挙型を返す
func (m *Match) GetRound() RoundType {
	return RoundType(m.Round)
}

// SetRound はRoundType列挙型から文字列を設定する
func (m *Match) SetRound(round RoundType) {
	m.Round = string(round)
}

// GetStatus はMatchStatus列挙型を返す
func (m *Match) GetStatus() MatchStatus {
	return MatchStatus(m.Status)
}

// SetStatus はMatchStatus列挙型から文字列を設定する
func (m *Match) SetStatus(status MatchStatus) {
	m.Status = string(status)
}

// GetScheduledAt はDateTime型で予定日時を返す
func (m *Match) GetScheduledAt() DateTime {
	return NewDateTime(m.ScheduledAt)
}

// GetCompletedAt はNullableDateTime型で完了日時を返す
func (m *Match) GetCompletedAt() NullableDateTime {
	if m.CompletedAt == nil {
		return NullableDateTime{Valid: false}
	}
	return NewNullableDateTime(*m.CompletedAt)
}

// SetCompletedAt は完了日時を設定する
func (m *Match) SetCompletedAt(t *time.Time) {
	m.CompletedAt = t
}

// GetCreatedAt はDateTime型で作成日時を返す
func (m *Match) GetCreatedAt() DateTime {
	return NewDateTime(m.CreatedAt)
}

// GetUpdatedAt はDateTime型で更新日時を返す
func (m *Match) GetUpdatedAt() DateTime {
	return NewDateTime(m.UpdatedAt)
}

// MatchResult は試合結果を表す構造体
type MatchResult struct {
	Score1 int    `json:"score1"`
	Score2 int    `json:"score2"`
	Winner string `json:"winner"`
}

// Validate は試合データの検証を行う
func (m *Match) Validate() error {
	if m.TournamentID <= 0 {
		return errors.New("トーナメントIDは必須です")
	}
	
	if strings.TrimSpace(m.Round) == "" {
		return errors.New("ラウンドは必須です")
	}
	
	if !m.GetRound().IsValid() {
		return errors.New("無効なラウンドです")
	}
	
	if strings.TrimSpace(m.Team1) == "" {
		return errors.New("チーム1は必須です")
	}
	
	if len(m.Team1) > 100 {
		return errors.New("チーム1名は100文字以下である必要があります")
	}
	
	if strings.TrimSpace(m.Team2) == "" {
		return errors.New("チーム2は必須です")
	}
	
	if len(m.Team2) > 100 {
		return errors.New("チーム2名は100文字以下である必要があります")
	}
	
	if m.Team1 == m.Team2 {
		return errors.New("同じチーム同士の試合はできません")
	}
	
	if strings.TrimSpace(m.Status) == "" {
		return errors.New("ステータスは必須です")
	}
	
	if !m.GetStatus().IsValid() {
		return errors.New("無効な試合ステータスです")
	}
	
	// 予定日時の検証
	if m.ScheduledAt.IsZero() {
		return errors.New("予定日時は必須です")
	}
	
	return nil
}

// ValidateResult は試合結果の検証を行う
func (mr *MatchResult) Validate() error {
	if mr.Score1 < 0 || mr.Score2 < 0 {
		return errors.New("スコアは0以上である必要があります")
	}
	
	if strings.TrimSpace(mr.Winner) == "" {
		return errors.New("勝者は必須です")
	}
	
	if mr.Score1 == mr.Score2 {
		return errors.New("引き分けは許可されていません")
	}
	
	return nil
}

// ValidateResultWithTeams は試合結果とチーム名の整合性を検証する
func (mr *MatchResult) ValidateResultWithTeams(team1, team2 string) error {
	if err := mr.Validate(); err != nil {
		return err
	}
	
	// 勝者がいずれかのチームと一致するかチェック
	if mr.Winner != team1 && mr.Winner != team2 {
		return errors.New("勝者は参加チームのいずれかである必要があります")
	}
	
	// スコアと勝者の整合性チェック
	if mr.Score1 > mr.Score2 && mr.Winner != team1 {
		return errors.New("スコアと勝者が一致しません")
	}
	
	if mr.Score2 > mr.Score1 && mr.Winner != team2 {
		return errors.New("スコアと勝者が一致しません")
	}
	
	return nil
}

// IsPending は試合が未実施かどうかを返す
func (m *Match) IsPending() bool {
	return m.GetStatus() == MatchStatusPendingEnum
}

// IsInProgress は試合が進行中かどうかを返す
func (m *Match) IsInProgress() bool {
	return m.GetStatus() == MatchStatusInProgressEnum
}

// IsCompleted は試合が完了しているかどうかを返す
func (m *Match) IsCompleted() bool {
	return m.GetStatus() == MatchStatusCompletedEnum
}

// IsCancelled は試合がキャンセルされているかどうかを返す
func (m *Match) IsCancelled() bool {
	return m.GetStatus() == MatchStatusCancelledEnum
}

// HasResult は試合結果が入力されているかどうかを返す
func (m *Match) HasResult() bool {
	return m.Score1 != nil && m.Score2 != nil && m.Winner != nil
}

// CanUpdateResult は試合結果を更新可能かどうかを返す
func (m *Match) CanUpdateResult() bool {
	return m.IsPending() || m.IsInProgress()
}

// CanDelete は試合を削除可能かどうかを返す
func (m *Match) CanDelete() bool {
	return !m.IsCompleted()
}
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
	Round        string     `json:"round" db:"round"` // "1st_round", "quarterfinal"など
	Team1        string     `json:"team1" db:"team1"`
	Team2        string     `json:"team2" db:"team2"`
	Score1       *int       `json:"score1" db:"score1"` // 試合が行われるまでnull
	Score2       *int       `json:"score2" db:"score2"`
	Winner       *string    `json:"winner" db:"winner"`
	Status       string     `json:"status" db:"status"`
	ScheduledAt  time.Time  `json:"scheduled_at" db:"scheduled_at"`
	CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
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
	
	if !IsValidRound(m.Round) {
		return errors.New("無効なラウンドです")
	}
	
	if strings.TrimSpace(m.Team1) == "" {
		return errors.New("チーム1は必須です")
	}
	
	if strings.TrimSpace(m.Team2) == "" {
		return errors.New("チーム2は必須です")
	}
	
	if m.Team1 == m.Team2 {
		return errors.New("同じチーム同士の試合はできません")
	}
	
	if strings.TrimSpace(m.Status) == "" {
		return errors.New("ステータスは必須です")
	}
	
	if !IsValidMatchStatus(m.Status) {
		return errors.New("無効な試合ステータスです")
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
	
	// 勝者の判定ロジック
	if mr.Score1 > mr.Score2 && mr.Winner == "" {
		return errors.New("スコアに基づいて勝者を決定してください")
	}
	
	if mr.Score1 == mr.Score2 {
		return errors.New("引き分けは許可されていません")
	}
	
	return nil
}

// IsPending は試合が未実施かどうかを返す
func (m *Match) IsPending() bool {
	return m.Status == MatchStatusPending
}

// IsCompleted は試合が完了しているかどうかを返す
func (m *Match) IsCompleted() bool {
	return m.Status == MatchStatusCompleted
}

// HasResult は試合結果が入力されているかどうかを返す
func (m *Match) HasResult() bool {
	return m.Score1 != nil && m.Score2 != nil && m.Winner != nil
}
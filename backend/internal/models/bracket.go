package models

import (
	"errors"
	"strings"
)

// Bracket はトーナメントブラケットを表すモデル
type Bracket struct {
	TournamentID int     `json:"tournament_id"`
	Sport        string  `json:"sport"`
	Format       string  `json:"format"`
	Rounds       []Round `json:"rounds"`
}

// Round はトーナメントのラウンドを表すモデル
type Round struct {
	Name    string  `json:"name"`
	Matches []Match `json:"matches"`
}

// Validate はブラケットデータの検証を行う
func (b *Bracket) Validate() error {
	if b.TournamentID <= 0 {
		return errors.New("トーナメントIDは必須です")
	}
	
	if strings.TrimSpace(b.Sport) == "" {
		return errors.New("スポーツは必須です")
	}
	
	if !IsValidSport(b.Sport) {
		return errors.New("無効なスポーツです")
	}
	
	if strings.TrimSpace(b.Format) == "" {
		return errors.New("フォーマットは必須です")
	}
	
	if !IsValidTournamentFormat(b.Format) {
		return errors.New("無効なトーナメントフォーマットです")
	}
	
	if len(b.Rounds) == 0 {
		return errors.New("ラウンドは最低1つ必要です")
	}
	
	// 各ラウンドの検証
	for i, round := range b.Rounds {
		if err := round.Validate(); err != nil {
			return errors.New("ラウンド " + string(rune(i+1)) + " の検証エラー: " + err.Error())
		}
	}
	
	return nil
}

// Validate はラウンドデータの検証を行う
func (r *Round) Validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return errors.New("ラウンド名は必須です")
	}
	
	if !IsValidRound(r.Name) {
		return errors.New("無効なラウンド名です")
	}
	
	if len(r.Matches) == 0 {
		return errors.New("ラウンドには最低1つの試合が必要です")
	}
	
	// 各試合の検証
	for i, match := range r.Matches {
		if err := match.Validate(); err != nil {
			return errors.New("試合 " + string(rune(i+1)) + " の検証エラー: " + err.Error())
		}
	}
	
	return nil
}

// GetCompletedMatches は完了した試合の数を返す
func (b *Bracket) GetCompletedMatches() int {
	count := 0
	for _, round := range b.Rounds {
		for _, match := range round.Matches {
			if match.IsCompleted() {
				count++
			}
		}
	}
	return count
}

// GetTotalMatches は総試合数を返す
func (b *Bracket) GetTotalMatches() int {
	count := 0
	for _, round := range b.Rounds {
		count += len(round.Matches)
	}
	return count
}

// IsCompleted はブラケットが完了しているかどうかを返す
func (b *Bracket) IsCompleted() bool {
	return b.GetCompletedMatches() == b.GetTotalMatches()
}
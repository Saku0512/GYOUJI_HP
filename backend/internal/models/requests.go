package models

import (
	"errors"
	"strings"
	"time"
)

// BaseRequest は全てのリクエストの基底構造体
type BaseRequest struct {
	RequestID string `json:"request_id,omitempty"` // リクエスト追跡用ID
}

// AuthRequests - 認証関連のリクエスト構造体

// LoginRequest はログインリクエストの統一構造体
type LoginRequest struct {
	BaseRequest
	Username string `json:"username" binding:"required,min=1,max=50" validate:"alphanum" example:"admin"`
	Password string `json:"password" binding:"required,min=8,max=100" example:"password"`
}

// Validate はLoginRequestの検証を行う
func (r *LoginRequest) Validate() error {
	if strings.TrimSpace(r.Username) == "" {
		return errors.New("ユーザー名は必須です")
	}
	
	if len(r.Username) < 1 || len(r.Username) > 50 {
		return errors.New("ユーザー名は1文字以上50文字以下である必要があります")
	}
	
	if strings.TrimSpace(r.Password) == "" {
		return errors.New("パスワードは必須です")
	}
	
	if len(r.Password) < 8 || len(r.Password) > 100 {
		return errors.New("パスワードは8文字以上100文字以下である必要があります")
	}
	
	return nil
}

// RefreshTokenRequest はトークンリフレッシュリクエストの統一構造体
type RefreshTokenRequest struct {
	BaseRequest
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
}

// Validate はRefreshTokenRequestの検証を行う
func (r *RefreshTokenRequest) Validate() error {
	if strings.TrimSpace(r.Token) == "" {
		return errors.New("トークンは必須です")
	}
	
	return nil
}

// TournamentRequests - トーナメント関連のリクエスト構造体

// CreateTournamentRequest はトーナメント作成リクエストの統一構造体
type CreateTournamentRequest struct {
	BaseRequest
	Sport  SportType        `json:"sport" binding:"required" example:"volleyball"`
	Format TournamentFormat `json:"format" binding:"required" example:"standard"`
}

// Validate はCreateTournamentRequestの検証を行う
func (r *CreateTournamentRequest) Validate() error {
	if !r.Sport.IsValid() {
		return errors.New("無効なスポーツです")
	}
	
	if !r.Format.IsValid() {
		return errors.New("無効なトーナメントフォーマットです")
	}
	
	return nil
}

// UpdateTournamentRequest はトーナメント更新リクエストの統一構造体
type UpdateTournamentRequest struct {
	BaseRequest
	Format *TournamentFormat `json:"format,omitempty" example:"standard"`
	Status *TournamentStatus `json:"status,omitempty" example:"active"`
}

// Validate はUpdateTournamentRequestの検証を行う
func (r *UpdateTournamentRequest) Validate() error {
	if r.Format != nil && !r.Format.IsValid() {
		return errors.New("無効なトーナメントフォーマットです")
	}
	
	if r.Status != nil && !r.Status.IsValid() {
		return errors.New("無効なトーナメントステータスです")
	}
	
	return nil
}

// SwitchFormatRequest はトーナメント形式切り替えリクエストの統一構造体
type SwitchFormatRequest struct {
	BaseRequest
	Format TournamentFormat `json:"format" binding:"required" example:"rainy"`
}

// Validate はSwitchFormatRequestの検証を行う
func (r *SwitchFormatRequest) Validate() error {
	if !r.Format.IsValid() {
		return errors.New("無効なトーナメントフォーマットです")
	}
	
	return nil
}

// MatchRequests - 試合関連のリクエスト構造体

// CreateMatchRequest は試合作成リクエストの統一構造体
type CreateMatchRequest struct {
	BaseRequest
	TournamentID int       `json:"tournament_id" binding:"required,min=1"`
	Round        RoundType `json:"round" binding:"required" example:"1st_round"`
	Team1        string    `json:"team1" binding:"required,min=1,max=100" example:"チームA"`
	Team2        string    `json:"team2" binding:"required,min=1,max=100" example:"チームB"`
	ScheduledAt  DateTime  `json:"scheduled_at" binding:"required" example:"2024-01-01T10:00:00Z"`
}

// Validate はCreateMatchRequestの検証を行う
func (r *CreateMatchRequest) Validate() error {
	if r.TournamentID <= 0 {
		return errors.New("トーナメントIDは必須です")
	}
	
	if !r.Round.IsValid() {
		return errors.New("無効なラウンドです")
	}
	
	if strings.TrimSpace(r.Team1) == "" {
		return errors.New("チーム1は必須です")
	}
	
	if len(r.Team1) > 100 {
		return errors.New("チーム1名は100文字以下である必要があります")
	}
	
	if strings.TrimSpace(r.Team2) == "" {
		return errors.New("チーム2は必須です")
	}
	
	if len(r.Team2) > 100 {
		return errors.New("チーム2名は100文字以下である必要があります")
	}
	
	if r.Team1 == r.Team2 {
		return errors.New("同じチーム同士の試合はできません")
	}
	
	if r.ScheduledAt.IsZero() {
		return errors.New("予定日時は必須です")
	}
	
	// 過去の日時チェック
	if r.ScheduledAt.Time.Before(time.Now()) {
		return errors.New("予定日時は現在時刻より後である必要があります")
	}
	
	return nil
}

// UpdateMatchRequest は試合更新リクエストの統一構造体
type UpdateMatchRequest struct {
	BaseRequest
	Round       *RoundType `json:"round,omitempty" example:"quarterfinal"`
	Team1       *string    `json:"team1,omitempty" example:"チームA"`
	Team2       *string    `json:"team2,omitempty" example:"チームB"`
	Status      *MatchStatus `json:"status,omitempty" example:"in_progress"`
	ScheduledAt *DateTime  `json:"scheduled_at,omitempty" example:"2024-01-01T10:00:00Z"`
}

// Validate はUpdateMatchRequestの検証を行う
func (r *UpdateMatchRequest) Validate() error {
	if r.Round != nil && !r.Round.IsValid() {
		return errors.New("無効なラウンドです")
	}
	
	if r.Team1 != nil {
		if strings.TrimSpace(*r.Team1) == "" {
			return errors.New("チーム1名は空にできません")
		}
		if len(*r.Team1) > 100 {
			return errors.New("チーム1名は100文字以下である必要があります")
		}
	}
	
	if r.Team2 != nil {
		if strings.TrimSpace(*r.Team2) == "" {
			return errors.New("チーム2名は空にできません")
		}
		if len(*r.Team2) > 100 {
			return errors.New("チーム2名は100文字以下である必要があります")
		}
	}
	
	if r.Team1 != nil && r.Team2 != nil && *r.Team1 == *r.Team2 {
		return errors.New("同じチーム同士の試合はできません")
	}
	
	if r.Status != nil && !r.Status.IsValid() {
		return errors.New("無効な試合ステータスです")
	}
	
	if r.ScheduledAt != nil && r.ScheduledAt.Time.Before(time.Now()) {
		return errors.New("予定日時は現在時刻より後である必要があります")
	}
	
	return nil
}

// SubmitMatchResultRequest は試合結果提出リクエストの統一構造体
type SubmitMatchResultRequest struct {
	BaseRequest
	Score1 int    `json:"score1" binding:"required,min=0" example:"3"`
	Score2 int    `json:"score2" binding:"required,min=0" example:"1"`
	Winner string `json:"winner" binding:"required,min=1,max=100" example:"チームA"`
}

// Validate はSubmitMatchResultRequestの検証を行う
func (r *SubmitMatchResultRequest) Validate() error {
	if r.Score1 < 0 {
		return errors.New("チーム1のスコアは0以上である必要があります")
	}
	
	if r.Score2 < 0 {
		return errors.New("チーム2のスコアは0以上である必要があります")
	}
	
	if r.Score1 == r.Score2 {
		return errors.New("引き分けは許可されていません")
	}
	
	if strings.TrimSpace(r.Winner) == "" {
		return errors.New("勝者は必須です")
	}
	
	if len(r.Winner) > 100 {
		return errors.New("勝者名は100文字以下である必要があります")
	}
	
	return nil
}

// ValidateMatchResult は試合結果とチーム名の整合性を検証する
func (r *SubmitMatchResultRequest) ValidateMatchResult(team1, team2 string) error {
	if err := r.Validate(); err != nil {
		return err
	}
	
	// 勝者がいずれかのチームと一致するかチェック
	if r.Winner != team1 && r.Winner != team2 {
		return errors.New("勝者は参加チームのいずれかである必要があります")
	}
	
	// スコアと勝者の整合性チェック
	if r.Score1 > r.Score2 && r.Winner != team1 {
		return errors.New("スコアと勝者が一致しません")
	}
	
	if r.Score2 > r.Score1 && r.Winner != team2 {
		return errors.New("スコアと勝者が一致しません")
	}
	
	return nil
}

// PaginationRequest はページネーションリクエストの統一構造体
type PaginationRequest struct {
	Page     int `form:"page" binding:"min=1" example:"1"`
	PageSize int `form:"page_size" binding:"min=1,max=100" example:"20"`
}

// Validate はPaginationRequestの検証を行う
func (r *PaginationRequest) Validate() error {
	if r.Page < 1 {
		return errors.New("ページ番号は1以上である必要があります")
	}
	
	if r.PageSize < 1 || r.PageSize > 100 {
		return errors.New("ページサイズは1以上100以下である必要があります")
	}
	
	return nil
}

// GetOffset はページネーション用のオフセットを計算する
func (r *PaginationRequest) GetOffset() int {
	return (r.Page - 1) * r.PageSize
}

// GetLimit はページネーション用のリミットを返す
func (r *PaginationRequest) GetLimit() int {
	return r.PageSize
}

// FilterRequest はフィルタリングリクエストの統一構造体
type FilterRequest struct {
	Sport  *SportType        `form:"sport" example:"volleyball"`
	Status *TournamentStatus `form:"status" example:"active"`
	Round  *RoundType        `form:"round" example:"quarterfinal"`
}

// MatchFilterRequest は試合フィルタリングリクエストの統一構造体
type MatchFilterRequest struct {
	Sport        *SportType    `form:"sport" example:"volleyball"`
	Status       *MatchStatus  `form:"status" example:"completed"`
	Round        *RoundType    `form:"round" example:"quarterfinal"`
	TournamentID *int          `form:"tournament_id" example:"1"`
}

// Validate はFilterRequestの検証を行う
func (r *FilterRequest) Validate() error {
	if r.Sport != nil && !r.Sport.IsValid() {
		return errors.New("無効なスポーツです")
	}
	
	if r.Status != nil && !r.Status.IsValid() {
		return errors.New("無効なステータスです")
	}
	
	if r.Round != nil && !r.Round.IsValid() {
		return errors.New("無効なラウンドです")
	}
	
	return nil
}
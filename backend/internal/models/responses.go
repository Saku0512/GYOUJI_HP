package models

// BaseResponse は全てのレスポンスの基底構造体
type BaseResponse struct {
	Success   bool   `json:"success"`             // 成功フラグ
	Message   string `json:"message"`             // メッセージ
	Code      int    `json:"code"`                // HTTPステータスコード
	Timestamp string `json:"timestamp"`           // タイムスタンプ（ISO 8601形式）
	RequestID string `json:"request_id,omitempty"` // リクエストID（追跡用）
}

// DataResponse は単一データを含むレスポンスの統一構造体
type DataResponse[T any] struct {
	BaseResponse
	Data T `json:"data,omitempty"` // レスポンスデータ
}

// SetRequestID はリクエストIDを設定する
func (r *DataResponse[T]) SetRequestID(requestID string) *DataResponse[T] {
	r.RequestID = requestID
	return r
}

// ListResponse はリストデータを含むレスポンスの統一構造体
type ListResponse[T any] struct {
	BaseResponse
	Data  []T `json:"data"`           // レスポンスデータ配列
	Count int `json:"count"`          // 件数
}

// PaginatedResponse はページネーション付きリストレスポンスの統一構造体
type PaginatedResponse[T any] struct {
	BaseResponse
	Data       []T                 `json:"data"`                 // レスポンスデータ配列
	Pagination *PaginationResponse `json:"pagination,omitempty"` // ページネーション情報
}

// ErrorResponse はエラーレスポンスの統一構造体
type ErrorResponse struct {
	BaseResponse
	Error string `json:"error,omitempty"` // エラーコード
}

// SetRequestID はリクエストIDを設定する
func (r *ErrorResponse) SetRequestID(requestID string) *ErrorResponse {
	r.RequestID = requestID
	return r
}

// ValidationErrorResponse はバリデーションエラー専用のレスポンス構造体
type ValidationErrorResponse struct {
	BaseResponse
	Error   string                    `json:"error"`               // エラーコード
	Details []ValidationErrorDetail   `json:"details"`             // 詳細なバリデーションエラー情報
}

// SetRequestID はリクエストIDを設定する
func (r *ValidationErrorResponse) SetRequestID(requestID string) *ValidationErrorResponse {
	r.RequestID = requestID
	return r
}

// PaginationResponse はページネーション情報の統一構造体
type PaginationResponse struct {
	Page       int  `json:"page"`        // 現在のページ番号
	PageSize   int  `json:"page_size"`   // ページサイズ
	TotalItems int  `json:"total_items"` // 総アイテム数
	TotalPages int  `json:"total_pages"` // 総ページ数
	HasNext    bool `json:"has_next"`    // 次のページが存在するか
	HasPrev    bool `json:"has_prev"`    // 前のページが存在するか
}

// NewPaginationResponse はページネーション情報を作成する
func NewPaginationResponse(page, pageSize, totalItems int) *PaginationResponse {
	totalPages := (totalItems + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}
	
	return &PaginationResponse{
		Page:       page,
		PageSize:   pageSize,
		TotalItems: totalItems,
		TotalPages: totalPages,
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
	}
}

// AuthResponses - 認証関連のレスポンス構造体

// LoginResponse はログインレスポンスの統一構造体
type LoginResponse struct {
	Token    string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Username string `json:"username" example:"admin"`
	Role     string `json:"role" example:"admin"`
	ExpiresAt DateTime `json:"expires_at" example:"2024-01-02T09:00:00Z"`
}

// RefreshTokenResponse はトークンリフレッシュレスポンスの統一構造体
type RefreshTokenResponse struct {
	Token     string   `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt DateTime `json:"expires_at" example:"2024-01-02T09:00:00Z"`
}

// UserProfileResponse はユーザープロフィールレスポンスの統一構造体
type UserProfileResponse struct {
	UserID   int    `json:"user_id" example:"1"`
	Username string `json:"username" example:"admin"`
	Role     string `json:"role" example:"admin"`
}

// TokenValidationResponse はトークン検証レスポンスの統一構造体
type TokenValidationResponse struct {
	Valid     bool     `json:"valid" example:"true"`
	UserID    int      `json:"user_id" example:"1"`
	Username  string   `json:"username" example:"admin"`
	Role      string   `json:"role" example:"admin"`
	ExpiresAt DateTime `json:"expires_at" example:"2024-01-02T09:00:00Z"`
}

// TournamentResponses - トーナメント関連のレスポンス構造体

// TournamentResponse はトーナメントレスポンスの統一構造体
type TournamentResponse struct {
	ID        int               `json:"id" example:"1"`
	Sport     SportType         `json:"sport" example:"volleyball"`
	Format    TournamentFormat  `json:"format" example:"standard"`
	Status    TournamentStatus  `json:"status" example:"active"`
	CreatedAt DateTime          `json:"created_at" example:"2024-01-01T09:00:00Z"`
	UpdatedAt DateTime          `json:"updated_at" example:"2024-01-01T11:00:00Z"`
}

// NewTournamentResponse はTournamentモデルからレスポンス構造体を作成する
func NewTournamentResponse(tournament *Tournament) *TournamentResponse {
	return &TournamentResponse{
		ID:        tournament.ID,
		Sport:     SportType(tournament.Sport),
		Format:    TournamentFormat(tournament.Format),
		Status:    TournamentStatus(tournament.Status),
		CreatedAt: NewDateTime(tournament.CreatedAt),
		UpdatedAt: NewDateTime(tournament.UpdatedAt),
	}
}

// BracketResponse はブラケットレスポンスの統一構造体
type BracketResponse struct {
	TournamentID int           `json:"tournament_id" example:"1"`
	Sport        SportType     `json:"sport" example:"volleyball"`
	Format       TournamentFormat `json:"format" example:"standard"`
	Rounds       []RoundResponse `json:"rounds"`
}

// RoundResponse はラウンドレスポンスの統一構造体
type RoundResponse struct {
	Name    RoundType       `json:"name" example:"1st_round"`
	Matches []MatchResponse `json:"matches"`
}

// TournamentProgressResponse はトーナメント進行状況レスポンスの統一構造体
type TournamentProgressResponse struct {
	TournamentID     int      `json:"tournament_id" example:"1"`
	Sport            SportType `json:"sport" example:"volleyball"`
	Format           TournamentFormat `json:"format" example:"standard"`
	Status           TournamentStatus `json:"status" example:"active"`
	TotalMatches     int      `json:"total_matches" example:"16"`
	CompletedMatches int      `json:"completed_matches" example:"8"`
	PendingMatches   int      `json:"pending_matches" example:"8"`
	ProgressPercent  float64  `json:"progress_percent" example:"50.0"`
	CurrentRound     RoundType `json:"current_round" example:"quarterfinal"`
}

// MatchResponses - 試合関連のレスポンス構造体

// MatchResponse は試合レスポンスの統一構造体
type MatchResponse struct {
	ID           int                  `json:"id" example:"1"`
	TournamentID int                  `json:"tournament_id" example:"1"`
	Round        RoundType            `json:"round" example:"1st_round"`
	Team1        string               `json:"team1" example:"チームA"`
	Team2        string               `json:"team2" example:"チームB"`
	Score1       *int                 `json:"score1" example:"3"`
	Score2       *int                 `json:"score2" example:"1"`
	Winner       *string              `json:"winner" example:"チームA"`
	Status       MatchStatus          `json:"status" example:"completed"`
	ScheduledAt  DateTime             `json:"scheduled_at" example:"2024-01-01T10:00:00Z"`
	CompletedAt  *NullableDateTime    `json:"completed_at" example:"2024-01-01T11:00:00Z"`
	CreatedAt    DateTime             `json:"created_at" example:"2024-01-01T09:00:00Z"`
	UpdatedAt    DateTime             `json:"updated_at" example:"2024-01-01T11:00:00Z"`
}

// NewMatchResponse はMatchモデルからレスポンス構造体を作成する
func NewMatchResponse(match *Match) *MatchResponse {
	var completedAt *NullableDateTime
	if match.CompletedAt != nil {
		completedAt = &NullableDateTime{
			DateTime: NewDateTime(*match.CompletedAt),
			Valid:    true,
		}
	}
	
	return &MatchResponse{
		ID:           match.ID,
		TournamentID: match.TournamentID,
		Round:        RoundType(match.Round),
		Team1:        match.Team1,
		Team2:        match.Team2,
		Score1:       match.Score1,
		Score2:       match.Score2,
		Winner:       match.Winner,
		Status:       MatchStatus(match.Status),
		ScheduledAt:  NewDateTime(match.ScheduledAt),
		CompletedAt:  completedAt,
		CreatedAt:    NewDateTime(match.CreatedAt),
		UpdatedAt:    NewDateTime(match.UpdatedAt),
	}
}

// MatchStatisticsResponse は試合統計レスポンスの統一構造体
type MatchStatisticsResponse struct {
	TournamentID     int                        `json:"tournament_id" example:"1"`
	TotalMatches     int                        `json:"total_matches" example:"16"`
	CompletedMatches int                        `json:"completed_matches" example:"8"`
	PendingMatches   int                        `json:"pending_matches" example:"8"`
	MatchesByRound   map[RoundType]int          `json:"matches_by_round"`
	CompletionRate   float64                    `json:"completion_rate" example:"0.5"`
	AverageScore     map[string]float64         `json:"average_score"`
	TeamStats        map[string]*TeamStatsResponse `json:"team_stats"`
}

// TeamStatsResponse はチーム統計レスポンスの統一構造体
type TeamStatsResponse struct {
	TeamName      string  `json:"team_name" example:"チームA"`
	MatchesPlayed int     `json:"matches_played" example:"4"`
	Wins          int     `json:"wins" example:"3"`
	Losses        int     `json:"losses" example:"1"`
	TotalScore    int     `json:"total_score" example:"12"`
	AverageScore  float64 `json:"average_score" example:"3.0"`
}

// HealthCheckResponse はヘルスチェックレスポンスの統一構造体
type HealthCheckResponse struct {
	Status    string            `json:"status" example:"healthy"`
	Version   string            `json:"version" example:"1.0.0"`
	Timestamp DateTime          `json:"timestamp" example:"2024-01-01T12:00:00Z"`
	Services  map[string]string `json:"services"`
}

// NewHealthCheckResponse はヘルスチェックレスポンスを作成する
func NewHealthCheckResponse(version string, services map[string]string) *HealthCheckResponse {
	return &HealthCheckResponse{
		Status:    "healthy",
		Version:   version,
		Timestamp: Now(),
		Services:  services,
	}
}
package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"backend/internal/models"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

// TournamentHandler はトーナメント関連のHTTPハンドラー
type TournamentHandler struct {
	tournamentService service.TournamentService
}

// NewTournamentHandler は新しいTournamentHandlerを作成する
func NewTournamentHandler(tournamentService service.TournamentService) *TournamentHandler {
	return &TournamentHandler{
		tournamentService: tournamentService,
	}
}

// CreateTournamentRequest はトーナメント作成リクエストの構造体
type CreateTournamentRequest struct {
	Sport  string `json:"sport" binding:"required"`
	Format string `json:"format" binding:"required"`
}

// UpdateTournamentRequest はトーナメント更新リクエストの構造体
type UpdateTournamentRequest struct {
	Format string `json:"format"`
	Status string `json:"status"`
}

// SwitchFormatRequest はトーナメント形式切り替えリクエストの構造体
type SwitchFormatRequest struct {
	Format string `json:"format" binding:"required"`
}

// TournamentResponse はトーナメントレスポンスの構造体
type TournamentResponse struct {
	Success bool       `json:"success" example:"true"`              // 成功フラグ
	Message string     `json:"message" example:"トーナメント情報を取得しました"` // メッセージ
	Data    Tournament `json:"data"`                                // トーナメントデータ
}

// TournamentListResponse はトーナメント一覧レスポンスの構造体
type TournamentListResponse struct {
	Success bool         `json:"success" example:"true"`              // 成功フラグ
	Message string       `json:"message" example:"トーナメント一覧を取得しました"` // メッセージ
	Data    []Tournament `json:"data"`                                // トーナメントデータ配列
	Count   int          `json:"count" example:"3"`                   // 件数
}

// Tournament はSwagger用のトーナメント構造体
type Tournament struct {
	ID        int    `json:"id" example:"1"`                                      // トーナメントID
	Sport     string `json:"sport" example:"volleyball"`                         // スポーツ種目
	Format    string `json:"format" example:"standard"`                          // トーナメント形式
	Status    string `json:"status" example:"active"`                            // ステータス
	CreatedAt string `json:"created_at" example:"2024-01-01T09:00:00Z"`         // 作成日時
	UpdatedAt string `json:"updated_at" example:"2024-01-01T11:00:00Z"`         // 更新日時
}

// convertToSwaggerTournament はmodels.TournamentをSwagger用のTournamentに変換する
func convertToSwaggerTournament(tournament *models.Tournament) Tournament {
	return Tournament{
		ID:        tournament.ID,
		Sport:     tournament.Sport,
		Format:    tournament.Format,
		Status:    tournament.Status,
		CreatedAt: tournament.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: tournament.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// convertToSwaggerTournaments はmodels.Tournamentの配列をSwagger用のTournamentの配列に変換する
func convertToSwaggerTournaments(tournaments []*models.Tournament) []Tournament {
	result := make([]Tournament, len(tournaments))
	for i, tournament := range tournaments {
		result[i] = convertToSwaggerTournament(tournament)
	}
	return result
}

// convertToSwaggerBracket はmodels.BracketをSwagger用のBracketに変換する
func convertToSwaggerBracket(bracket *models.Bracket) Bracket {
	rounds := make([]Round, len(bracket.Rounds))
	for i, round := range bracket.Rounds {
		matches := make([]Match, len(round.Matches))
		for j, match := range round.Matches {
			var completedAt *string
			if match.CompletedAt != nil {
				completedAtStr := match.CompletedAt.Format("2006-01-02T15:04:05Z")
				completedAt = &completedAtStr
			}

			matches[j] = Match{
				ID:           match.ID,
				TournamentID: match.TournamentID,
				Round:        match.Round,
				Team1:        match.Team1,
				Team2:        match.Team2,
				Score1:       match.Score1,
				Score2:       match.Score2,
				Winner:       match.Winner,
				Status:       match.Status,
				ScheduledAt:  match.ScheduledAt.Format("2006-01-02T15:04:05Z"),
				CompletedAt:  completedAt,
				CreatedAt:    match.CreatedAt.Format("2006-01-02T15:04:05Z"),
				UpdatedAt:    match.UpdatedAt.Format("2006-01-02T15:04:05Z"),
			}
		}
		rounds[i] = Round{
			Name:    round.Name,
			Matches: matches,
		}
	}

	return Bracket{
		TournamentID: bracket.TournamentID,
		Sport:        bracket.Sport,
		Format:       bracket.Format,
		Rounds:       rounds,
	}
}

// convertToSwaggerTournamentProgress はservice.TournamentProgressをSwagger用のTournamentProgressに変換する
func convertToSwaggerTournamentProgress(progress *service.TournamentProgress) TournamentProgress {
	return TournamentProgress{
		TournamentID:     progress.TournamentID,
		Sport:            progress.Sport,
		Format:           progress.Format,
		Status:           progress.Status,
		TotalMatches:     progress.TotalMatches,
		CompletedMatches: progress.CompletedMatches,
		PendingMatches:   progress.PendingMatches,
		ProgressPercent:  progress.ProgressPercent,
		CurrentRound:     progress.CurrentRound,
	}
}

// convertMatchesToRounds はマッチデータをラウンド形式に変換する
func convertMatchesToRounds(matches []*models.Match) []models.Round {
	roundMap := make(map[string][]models.Match)
	
	// ラウンド別にマッチをグループ化（ポインタから値に変換）
	for _, match := range matches {
		roundMap[match.Round] = append(roundMap[match.Round], *match)
	}
	
	// ラウンドの順序を定義
	roundOrder := []string{
		"1st_round", "2nd_round", "3rd_round", "4th_round",
		"quarterfinal", "semifinal", "final",
	}
	
	var rounds []models.Round
	for _, roundName := range roundOrder {
		if roundMatches, exists := roundMap[roundName]; exists {
			rounds = append(rounds, models.Round{
				Name:    roundName,
				Matches: roundMatches,
			})
		}
	}
	
	// 定義されていないラウンドも追加
	for roundName, roundMatches := range roundMap {
		found := false
		for _, definedRound := range roundOrder {
			if roundName == definedRound {
				found = true
				break
			}
		}
		if !found {
			rounds = append(rounds, models.Round{
				Name:    roundName,
				Matches: roundMatches,
			})
		}
	}
	
	return rounds
}

// BracketResponse はブラケットレスポンスの構造体
type BracketResponse struct {
	Success bool    `json:"success" example:"true"`          // 成功フラグ
	Message string  `json:"message" example:"ブラケット情報を取得しました"` // メッセージ
	Data    Bracket `json:"data"`                            // ブラケットデータ
}

// ProgressResponse はトーナメント進行状況レスポンスの構造体
type ProgressResponse struct {
	Success bool               `json:"success" example:"true"`      // 成功フラグ
	Message string             `json:"message" example:"進行状況を取得しました"` // メッセージ
	Data    TournamentProgress `json:"data"`                        // 進行状況データ
}

// Bracket はSwagger用のブラケット構造体
type Bracket struct {
	TournamentID int     `json:"tournament_id" example:"1"`        // トーナメントID
	Sport        string  `json:"sport" example:"volleyball"`       // スポーツ種目
	Format       string  `json:"format" example:"standard"`        // トーナメント形式
	Rounds       []Round `json:"rounds"`                           // ラウンド配列
}

// Round はSwagger用のラウンド構造体
type Round struct {
	Name    string  `json:"name" example:"1st_round"`    // ラウンド名
	Matches []Match `json:"matches"`                     // 試合配列
}

// TournamentProgress はSwagger用のトーナメント進行状況構造体
type TournamentProgress struct {
	TournamentID     int     `json:"tournament_id" example:"1"`        // トーナメントID
	Sport            string  `json:"sport" example:"volleyball"`       // スポーツ種目
	Format           string  `json:"format" example:"standard"`        // トーナメント形式
	Status           string  `json:"status" example:"active"`          // ステータス
	TotalMatches     int     `json:"total_matches" example:"16"`       // 総試合数
	CompletedMatches int     `json:"completed_matches" example:"8"`    // 完了試合数
	PendingMatches   int     `json:"pending_matches" example:"8"`      // 未完了試合数
	ProgressPercent  float64 `json:"progress_percent" example:"50.0"`  // 進行率
	CurrentRound     string  `json:"current_round" example:"quarterfinal"` // 現在のラウンド
}

// CreateTournament はトーナメント作成エンドポイントハンドラー
// @Summary トーナメント作成
// @Description 新しいトーナメントを作成する（管理者のみ）
// @Tags tournaments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTournamentRequest true "トーナメント作成情報"
// @Success 201 {object} TournamentResponse "作成成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 409 {object} ErrorResponse "競合エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments [post]
func (h *TournamentHandler) CreateTournament(c *gin.Context) {
	var req CreateTournamentRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なリクエスト形式です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 入力値の検証
	if strings.TrimSpace(req.Sport) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "スポーツは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if strings.TrimSpace(req.Format) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "フォーマットは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// トーナメント作成
	tournament := &models.Tournament{
		Sport:  req.Sport,
		Format: req.Format,
		Status: models.TournamentStatusRegistration, // デフォルトステータス
	}
	
	err := h.tournamentService.CreateTournament(context.Background(), tournament)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") || strings.Contains(err.Error(), "既に") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Conflict",
				Message: "既にアクティブなトーナメントが存在します",
				Code:    http.StatusConflict,
			})
			return
		}
		
		if strings.Contains(err.Error(), "無効な") || strings.Contains(err.Error(), "validation") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの作成に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusCreated, TournamentResponse{
		Success: true,
		Data:    convertToSwaggerTournament(tournament),
		Message: "トーナメントを作成しました",
	})
}

// GetTournaments は全トーナメント取得エンドポイントハンドラー
// @Summary 全トーナメント取得
// @Description 全てのトーナメントを取得する
// @Tags tournaments
// @Produce json
// @Success 200 {object} TournamentListResponse "取得成功"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments [get]
func (h *TournamentHandler) GetTournaments(c *gin.Context) {
	tournaments, err := h.tournamentService.GetTournaments(context.Background(), 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメント一覧の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, TournamentListResponse{
		Success: true,
		Data:    convertToSwaggerTournaments(tournaments),
		Count:   len(tournaments),
		Message: "トーナメント一覧を取得しました",
	})
}

// GetTournamentBySport はスポーツ別トーナメント取得エンドポイントハンドラー
// @Summary スポーツ別トーナメント取得
// @Description 指定されたスポーツのトーナメントを取得する
// @Tags tournaments
// @Produce json
// @Param sport path string true "スポーツ名" Enums(volleyball,table_tennis,soccer)
// @Success 200 {object} TournamentResponse "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{sport} [get]
func (h *TournamentHandler) GetTournamentBySport(c *gin.Context) {
	sport := c.Param("sport")

	// 入力値の検証
	if strings.TrimSpace(sport) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "スポーツパラメータは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// スポーツ別トーナメント取得
	tournaments, err := h.tournamentService.GetTournamentBySport(context.Background(), sport, 1, 0)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: "指定されたスポーツのトーナメントが見つかりません",
				Code:    http.StatusNotFound,
			})
			return
		}

		if strings.Contains(err.Error(), "無効な") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}
	
	if len(tournaments) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Not Found",
			Message: "指定されたスポーツのトーナメントが見つかりません",
			Code:    http.StatusNotFound,
		})
		return
	}
	
	tournament := tournaments[0]


	// 成功レスポンス
	c.JSON(http.StatusOK, TournamentResponse{
		Success: true,
		Data:    convertToSwaggerTournament(tournament),
		Message: "トーナメントを取得しました",
	})
}

// GetTournamentByID はID別トーナメント取得エンドポイントハンドラー
// @Summary ID別トーナメント取得
// @Description 指定されたIDのトーナメントを取得する
// @Tags tournaments
// @Produce json
// @Param id path int true "トーナメントID"
// @Success 200 {object} TournamentResponse "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/id/{id} [get]
func (h *TournamentHandler) GetTournamentByID(c *gin.Context) {
	idStr := c.Param("id")

	// IDの変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なトーナメントIDです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// トーナメント取得
	tournament, err := h.tournamentService.GetTournament(context.Background(), uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, TournamentResponse{
		Success: true,
		Data:    convertToSwaggerTournament(tournament),
		Message: "トーナメントを取得しました",
	})
}

// UpdateTournament はトーナメント更新エンドポイントハンドラー
// @Summary トーナメント更新
// @Description 指定されたIDのトーナメントを更新する（管理者のみ）
// @Tags tournaments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "トーナメントID"
// @Param request body UpdateTournamentRequest true "トーナメント更新情報"
// @Success 200 {object} TournamentResponse "更新成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{id} [put]
func (h *TournamentHandler) UpdateTournament(c *gin.Context) {
	idStr := c.Param("id")
	var req UpdateTournamentRequest

	// IDの変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なトーナメントIDです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なリクエスト形式です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 既存のトーナメントを取得
	tournament, err := h.tournamentService.GetTournament(context.Background(), uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 更新フィールドを適用
	if strings.TrimSpace(req.Format) != "" {
		tournament.Format = req.Format
	}
	if strings.TrimSpace(req.Status) != "" {
		tournament.Status = req.Status
	}

	// トーナメント更新
	err = h.tournamentService.UpdateTournament(context.Background(), uint(id), tournament)
	if err != nil {
		if strings.Contains(err.Error(), "無効な") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの更新に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, TournamentResponse{
		Success: true,
		Data:    convertToSwaggerTournament(tournament),
		Message: "トーナメントを更新しました",
	})
}

// DeleteTournament はトーナメント削除エンドポイントハンドラー
// @Summary トーナメント削除
// @Description 指定されたIDのトーナメントを削除する（管理者のみ）
// @Tags tournaments
// @Produce json
// @Security BearerAuth
// @Param id path int true "トーナメントID"
// @Success 200 {object} map[string]string "削除成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 409 {object} ErrorResponse "競合エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{id} [delete]
func (h *TournamentHandler) DeleteTournament(c *gin.Context) {
	idStr := c.Param("id")

	// IDの変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なトーナメントIDです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// トーナメント削除
	err = h.tournamentService.DeleteTournament(context.Background(), uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "試合が存在する") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Conflict",
				Message: err.Error(),
				Code:    http.StatusConflict,
			})
			return
		}

		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: "削除対象のトーナメントが見つかりません",
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの削除に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, gin.H{
		"message": "トーナメントを削除しました",
	})
}

// GetTournamentBracket はトーナメントブラケット取得エンドポイントハンドラー
// @Summary トーナメントブラケット取得
// @Description 指定されたスポーツのトーナメントブラケットを取得する
// @Tags tournaments
// @Produce json
// @Param sport path string true "スポーツ名" Enums(volleyball,table_tennis,soccer)
// @Success 200 {object} BracketResponse "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{sport}/bracket [get]
func (h *TournamentHandler) GetTournamentBracket(c *gin.Context) {
	sport := c.Param("sport")

	// 入力値の検証
	if strings.TrimSpace(sport) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "スポーツパラメータは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// スポーツ別トーナメント取得
	tournaments, err := h.tournamentService.GetTournamentBySport(context.Background(), sport, 1, 0)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: "指定されたスポーツのトーナメントが見つかりません",
				Code:    http.StatusNotFound,
			})
			return
		}

		if strings.Contains(err.Error(), "無効な") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}
	
	if len(tournaments) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Not Found",
			Message: "指定されたスポーツのトーナメントが見つかりません",
			Code:    http.StatusNotFound,
		})
		return
	}
	
	tournament := tournaments[0]
	
	// ブラケット取得
	matches, err := h.tournamentService.GetBracket(context.Background(), uint(tournament.ID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "ブラケットの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}
	
	// マッチデータをブラケット形式に変換
	bracket := &models.Bracket{
		TournamentID: tournament.ID,
		Sport:        tournament.Sport,
		Format:       tournament.Format,
		Rounds:       convertMatchesToRounds(matches),
	}
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		if strings.Contains(err.Error(), "無効な") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "ブラケットの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, BracketResponse{
		Success: true,
		Data:    convertToSwaggerBracket(bracket),
		Message: "ブラケットを取得しました",
	})
}

// SwitchTournamentFormat はトーナメント形式切り替えエンドポイントハンドラー
// @Summary トーナメント形式切り替え
// @Description 指定されたスポーツのトーナメント形式を切り替える（卓球の天候条件用、管理者のみ）
// @Tags tournaments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "トーナメントID"
// @Param request body SwitchFormatRequest true "形式切り替え情報"
// @Success 200 {object} map[string]string "切り替え成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 409 {object} ErrorResponse "競合エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{id}/format [put]
func (h *TournamentHandler) SwitchTournamentFormat(c *gin.Context) {
	idStr := c.Param("id")
	var req SwitchFormatRequest

	// IDの変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なトーナメントIDです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なリクエスト形式です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 入力値の検証
	if strings.TrimSpace(req.Format) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "フォーマットは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// トーナメント取得
	tournament, err := h.tournamentService.GetTournament(context.Background(), uint(id))
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}
	
	// 既に同じ形式の場合
	if tournament.Format == req.Format {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Conflict",
			Message: "既に指定された形式です",
			Code:    http.StatusConflict,
		})
		return
	}
	
	// トーナメント形式更新
	tournament.Format = req.Format
	err = h.tournamentService.UpdateTournament(context.Background(), uint(tournament.ID), tournament)
	if err != nil {
		if strings.Contains(err.Error(), "サポートしていません") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if strings.Contains(err.Error(), "既に") || strings.Contains(err.Error(), "変更できません") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Conflict",
				Message: err.Error(),
				Code:    http.StatusConflict,
			})
			return
		}

		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		if strings.Contains(err.Error(), "無効な") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメント形式の切り替えに失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, TournamentResponse{
		Success: true,
		Data:    convertToSwaggerTournament(tournament),
		Message: "トーナメント形式を切り替えました",
	})
}

// GetActiveTournaments はアクティブトーナメント取得エンドポイントハンドラー
// @Summary アクティブトーナメント取得
// @Description アクティブなトーナメントのみを取得する
// @Tags tournaments
// @Produce json
// @Success 200 {object} TournamentListResponse "取得成功"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/active [get]
func (h *TournamentHandler) GetActiveTournaments(c *gin.Context) {
	tournaments, err := h.tournamentService.GetTournamentBySport(context.Background(), "active", 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "アクティブトーナメント一覧の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, TournamentListResponse{
		Success: true,
		Data:    convertToSwaggerTournaments(tournaments),
		Count:   len(tournaments),
		Message: "アクティブトーナメント一覧を取得しました",
	})
}

// GetTournamentProgress はトーナメント進行状況取得エンドポイントハンドラー
// @Summary トーナメント進行状況取得
// @Description 指定されたスポーツのトーナメント進行状況を取得する
// @Tags tournaments
// @Produce json
// @Param sport path string true "スポーツ名" Enums(volleyball,table_tennis,soccer)
// @Success 200 {object} ProgressResponse "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{sport}/progress [get]
func (h *TournamentHandler) GetTournamentProgress(c *gin.Context) {
	sport := c.Param("sport")

	// 入力値の検証
	if strings.TrimSpace(sport) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "スポーツパラメータは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 進行状況取得
	progress, err := h.tournamentService.GetTournamentProgress(sport)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		if strings.Contains(err.Error(), "無効な") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメント進行状況の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, ProgressResponse{
		Success: true,
		Data:    convertToSwaggerTournamentProgress(progress),
		Message: "トーナメント進行状況を取得しました",
	})
}

// CompleteTournament はトーナメント完了エンドポイントハンドラー
// @Summary トーナメント完了
// @Description 指定されたスポーツのトーナメントを完了状態にする（管理者のみ）
// @Tags tournaments
// @Produce json
// @Security BearerAuth
// @Param sport path string true "スポーツ名" Enums(volleyball,table_tennis,soccer)
// @Success 200 {object} map[string]string "完了成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 409 {object} ErrorResponse "競合エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{sport}/complete [put]
func (h *TournamentHandler) CompleteTournament(c *gin.Context) {
	sport := c.Param("sport")

	// 入力値の検証
	if strings.TrimSpace(sport) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "スポーツパラメータは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// スポーツ別トーナメント取得
	tournaments, err := h.tournamentService.GetTournamentBySport(context.Background(), sport, 1, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}
	
	if len(tournaments) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Not Found",
			Message: "指定されたスポーツのトーナメントが見つかりません",
			Code:    http.StatusNotFound,
		})
		return
	}
	
	tournament := tournaments[0]
	
	// 既に完了している場合
	if tournament.Status == "completed" {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Conflict",
			Message: "既に完了しています",
			Code:    http.StatusConflict,
		})
		return
	}
	
	// トーナメント完了
	tournament.Status = "completed"
	err = h.tournamentService.UpdateTournament(context.Background(), uint(tournament.ID), tournament)
	if err != nil {
		if strings.Contains(err.Error(), "既に完了") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Conflict",
				Message: err.Error(),
				Code:    http.StatusConflict,
			})
			return
		}

		if strings.Contains(err.Error(), "完了していない") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Conflict",
				Message: err.Error(),
				Code:    http.StatusConflict,
			})
			return
		}

		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		if strings.Contains(err.Error(), "無効な") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの完了に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, TournamentResponse{
		Success: true,
		Data:    convertToSwaggerTournament(tournament),
		Message: "トーナメントを完了しました",
	})
}

// GetAvailableFormats は利用可能な形式一覧取得エンドポイントハンドラー
// @Summary 利用可能な形式一覧取得
// @Description 指定されたスポーツで利用可能なトーナメント形式一覧を取得する
// @Tags tournaments
// @Produce json
// @Param sport path string true "スポーツ名" Enums(volleyball,table_tennis,soccer)
// @Success 200 {object} map[string]interface{} "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{sport}/formats [get]
func (h *TournamentHandler) GetAvailableFormats(c *gin.Context) {
	sport := c.Param("sport")

	// 入力値の検証
	if strings.TrimSpace(sport) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "スポーツパラメータは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// スポーツ別の利用可能な形式を定義
	var formats []string
	switch sport {
	case "volleyball":
		formats = []string{"standard", "single_elimination", "double_elimination"}
	case "table_tennis":
		formats = []string{"sunny", "rainy", "standard"}
	case "soccer":
		formats = []string{"standard", "group_stage", "knockout"}
	default:
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "サポートされていないスポーツです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    formats,
		"message": fmt.Sprintf("%sの利用可能な形式一覧を取得しました", sport),
	})
}

// UpdateTournamentFormat はトーナメント形式更新エンドポイントハンドラー
// @Summary トーナメント形式更新
// @Description 指定されたスポーツのトーナメント形式を更新する（管理者のみ）
// @Tags tournaments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param sport path string true "スポーツ名" Enums(volleyball,table_tennis,soccer)
// @Param request body SwitchFormatRequest true "形式更新情報"
// @Success 200 {object} TournamentResponse "更新成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{sport}/format [put]
func (h *TournamentHandler) UpdateTournamentFormat(c *gin.Context) {
	sport := c.Param("sport")
	var req SwitchFormatRequest

	// 入力値の検証
	if strings.TrimSpace(sport) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "スポーツパラメータは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なリクエスト形式です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 入力値の検証
	if strings.TrimSpace(req.Format) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "フォーマットは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// スポーツ別トーナメント取得
	tournaments, err := h.tournamentService.GetTournamentBySport(context.Background(), sport, 1, 0)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") || strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: "指定されたスポーツのトーナメントが見つかりません",
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}
	
	if len(tournaments) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Not Found",
			Message: "指定されたスポーツのトーナメントが見つかりません",
			Code:    http.StatusNotFound,
		})
		return
	}
	
	tournament := tournaments[0]
	
	// 既に同じ形式の場合
	if tournament.Format == req.Format {
		c.JSON(http.StatusOK, TournamentResponse{
			Success: true,
			Data:    convertToSwaggerTournament(tournament),
			Message: "既に指定された形式です",
		})
		return
	}
	
	// トーナメント形式更新
	tournament.Format = req.Format
	err = h.tournamentService.UpdateTournament(context.Background(), uint(tournament.ID), tournament)
	if err != nil {
		if strings.Contains(err.Error(), "サポートしていません") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		if strings.Contains(err.Error(), "無効な") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメント形式の更新に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, TournamentResponse{
		Success: true,
		Data:    convertToSwaggerTournament(tournament),
		Message: "トーナメント形式を更新しました",
	})
}

// ActivateTournament はトーナメントアクティブ化エンドポイントハンドラー
// @Summary トーナメントアクティブ化
// @Description 指定されたスポーツのトーナメントをアクティブ状態にする（管理者のみ）
// @Tags tournaments
// @Produce json
// @Security BearerAuth
// @Param sport path string true "スポーツ名" Enums(volleyball,table_tennis,soccer)
// @Success 200 {object} map[string]string "アクティブ化成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 409 {object} ErrorResponse "競合エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/tournaments/{sport}/activate [put]
func (h *TournamentHandler) ActivateTournament(c *gin.Context) {
	sport := c.Param("sport")

	// 入力値の検証
	if strings.TrimSpace(sport) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "スポーツパラメータは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// スポーツ別トーナメント取得
	tournaments, err := h.tournamentService.GetTournamentBySport(context.Background(), sport, 1, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントの取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}
	
	if len(tournaments) == 0 {
		c.JSON(http.StatusNotFound, ErrorResponse{
			Error:   "Not Found",
			Message: "指定されたスポーツのトーナメントが見つかりません",
			Code:    http.StatusNotFound,
		})
		return
	}
	
	tournament := tournaments[0]
	
	// 既にアクティブの場合
	if tournament.Status == "active" {
		c.JSON(http.StatusConflict, ErrorResponse{
			Error:   "Conflict",
			Message: "既にアクティブです",
			Code:    http.StatusConflict,
		})
		return
	}
	
	// トーナメントアクティブ化
	tournament.Status = "active"
	err = h.tournamentService.UpdateTournament(context.Background(), uint(tournament.ID), tournament)
	if err != nil {
		if strings.Contains(err.Error(), "既にアクティブ") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Conflict",
				Message: err.Error(),
				Code:    http.StatusConflict,
			})
			return
		}

		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		if strings.Contains(err.Error(), "無効な") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "トーナメントのアクティブ化に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, TournamentResponse{
		Success: true,
		Data:    convertToSwaggerTournament(tournament),
		Message: "トーナメントをアクティブ化しました",
	})
}


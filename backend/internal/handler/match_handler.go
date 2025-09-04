package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend/internal/models"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

// MatchHandler は試合関連のHTTPハンドラー
type MatchHandler struct {
	matchService service.MatchService
}

// NewMatchHandler は新しいMatchHandlerを作成する
func NewMatchHandler(matchService service.MatchService) *MatchHandler {
	return &MatchHandler{
		matchService: matchService,
	}
}

// CreateMatchRequest は試合作成リクエストの構造体
type CreateMatchRequest struct {
	TournamentID int    `json:"tournament_id" binding:"required"`
	Round        string `json:"round" binding:"required"`
	Team1        string `json:"team1" binding:"required"`
	Team2        string `json:"team2" binding:"required"`
	ScheduledAt  string `json:"scheduled_at" binding:"required"`
}

// UpdateMatchRequest は試合更新リクエストの構造体
type UpdateMatchRequest struct {
	Round       string `json:"round"`
	Team1       string `json:"team1"`
	Team2       string `json:"team2"`
	Status      string `json:"status"`
	ScheduledAt string `json:"scheduled_at"`
}

// SubmitMatchResultRequest は試合結果提出リクエストの構造体
type SubmitMatchResultRequest struct {
	Score1 int    `json:"score1" binding:"required,min=0"`
	Score2 int    `json:"score2" binding:"required,min=0"`
	Winner string `json:"winner" binding:"required"`
}

// MatchResponse は試合レスポンスの構造体
type MatchResponse struct {
	*models.Match
	Message string `json:"message,omitempty"`
}

// MatchListResponse は試合一覧レスポンスの構造体
type MatchListResponse struct {
	Matches []*models.Match `json:"matches"`
	Count   int             `json:"count"`
	Message string          `json:"message,omitempty"`
}

// MatchStatisticsResponse は試合統計レスポンスの構造体
type MatchStatisticsResponse struct {
	*service.MatchStatistics
	Message string `json:"message,omitempty"`
}

// CreateMatch は試合作成エンドポイントハンドラー
// @Summary 試合作成
// @Description 新しい試合を作成する（管理者のみ）
// @Tags matches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateMatchRequest true "試合作成情報"
// @Success 201 {object} MatchResponse "作成成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 409 {object} ErrorResponse "競合エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches [post]
func (h *MatchHandler) CreateMatch(c *gin.Context) {
	var req CreateMatchRequest

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
	if req.TournamentID <= 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "トーナメントIDは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if strings.TrimSpace(req.Round) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "ラウンドは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if strings.TrimSpace(req.Team1) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "チーム1は必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if strings.TrimSpace(req.Team2) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "チーム2は必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if req.Team1 == req.Team2 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "同じチーム同士の試合はできません",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 日時のパース
	scheduledAt, err := parseDateTime(req.ScheduledAt)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効な日時形式です（YYYY-MM-DD HH:MM:SS形式で入力してください）",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 試合モデルを作成
	match := &models.Match{
		TournamentID: req.TournamentID,
		Round:        req.Round,
		Team1:        req.Team1,
		Team2:        req.Team2,
		Status:       models.MatchStatusPending,
		ScheduledAt:  scheduledAt,
	}

	// 試合作成
	err = h.matchService.CreateMatch(match)
	if err != nil {
		if strings.Contains(err.Error(), "既に存在") {
			c.JSON(http.StatusConflict, ErrorResponse{
				Error:   "Conflict",
				Message: err.Error(),
				Code:    http.StatusConflict,
			})
			return
		}

		if strings.Contains(err.Error(), "無効な") || strings.Contains(err.Error(), "検証") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "試合の作成に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusCreated, MatchResponse{
		Match:   match,
		Message: "試合を作成しました",
	})
}

// GetMatches は全試合取得エンドポイントハンドラー
// @Summary 全試合取得
// @Description 全ての試合を取得する
// @Tags matches
// @Produce json
// @Success 200 {object} MatchListResponse "取得成功"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches [get]
func (h *MatchHandler) GetMatches(c *gin.Context) {
	// クエリパラメータから状態フィルターを取得
	status := c.Query("status")
	
	var matches []*models.Match
	var err error

	switch status {
	case "pending":
		matches, err = h.matchService.GetPendingMatches()
	case "completed":
		matches, err = h.matchService.GetCompletedMatches()
	default:
		// 全ての試合を取得（実装が必要な場合）
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "statusパラメータを指定してください（pending または completed）",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "試合一覧の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, MatchListResponse{
		Matches: matches,
		Count:   len(matches),
		Message: "試合一覧を取得しました",
	})
}

// GetMatchesBySport はスポーツ別試合取得エンドポイントハンドラー
// @Summary スポーツ別試合取得
// @Description 指定されたスポーツの試合を取得する
// @Tags matches
// @Produce json
// @Param sport path string true "スポーツ名" Enums(volleyball,table_tennis,soccer)
// @Success 200 {object} MatchListResponse "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches/{sport} [get]
func (h *MatchHandler) GetMatchesBySport(c *gin.Context) {
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

	// スポーツ別試合取得
	matches, err := h.matchService.GetMatchesBySport(sport)
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
			Message: "試合の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, MatchListResponse{
		Matches: matches,
		Count:   len(matches),
		Message: "スポーツ別試合一覧を取得しました",
	})
}

// GetMatch はID別試合取得エンドポイントハンドラー
// @Summary ID別試合取得
// @Description 指定されたIDの試合を取得する
// @Tags matches
// @Produce json
// @Param id path int true "試合ID"
// @Success 200 {object} MatchResponse "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches/id/{id} [get]
func (h *MatchHandler) GetMatch(c *gin.Context) {
	idStr := c.Param("id")

	// IDの変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効な試合IDです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 試合取得
	match, err := h.matchService.GetMatch(id)
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
			Message: "試合の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, MatchResponse{
		Match:   match,
		Message: "試合を取得しました",
	})
}

// UpdateMatch は試合更新エンドポイントハンドラー
// @Summary 試合更新
// @Description 指定されたIDの試合を更新する（管理者のみ）
// @Tags matches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "試合ID"
// @Param request body UpdateMatchRequest true "試合更新情報"
// @Success 200 {object} MatchResponse "更新成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches/{id} [put]
func (h *MatchHandler) UpdateMatch(c *gin.Context) {
	idStr := c.Param("id")
	var req UpdateMatchRequest

	// IDの変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効な試合IDです",
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

	// 既存の試合を取得
	match, err := h.matchService.GetMatch(id)
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
			Message: "試合の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 更新フィールドを適用
	if strings.TrimSpace(req.Round) != "" {
		match.Round = req.Round
	}
	if strings.TrimSpace(req.Team1) != "" {
		match.Team1 = req.Team1
	}
	if strings.TrimSpace(req.Team2) != "" {
		match.Team2 = req.Team2
	}
	if strings.TrimSpace(req.Status) != "" {
		match.Status = req.Status
	}
	if strings.TrimSpace(req.ScheduledAt) != "" {
		scheduledAt, err := parseDateTime(req.ScheduledAt)
		if err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "無効な日時形式です（YYYY-MM-DD HH:MM:SS形式で入力してください）",
				Code:    http.StatusBadRequest,
			})
			return
		}
		match.ScheduledAt = scheduledAt
	}

	// 試合更新
	err = h.matchService.UpdateMatch(match)
	if err != nil {
		if strings.Contains(err.Error(), "無効な") || strings.Contains(err.Error(), "検証") {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: err.Error(),
				Code:    http.StatusBadRequest,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "試合の更新に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, MatchResponse{
		Match:   match,
		Message: "試合を更新しました",
	})
}

// SubmitMatchResult は試合結果提出エンドポイントハンドラー
// @Summary 試合結果提出
// @Description 指定された試合の結果を提出する（管理者のみ）
// @Tags matches
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "試合ID"
// @Param request body SubmitMatchResultRequest true "試合結果情報"
// @Success 200 {object} MatchResponse "提出成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 422 {object} ErrorResponse "ビジネスロジックエラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches/{id}/result [put]
func (h *MatchHandler) SubmitMatchResult(c *gin.Context) {
	idStr := c.Param("id")
	var req SubmitMatchResultRequest

	// IDの変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効な試合IDです",
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
	if req.Score1 < 0 || req.Score2 < 0 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "スコアは0以上である必要があります",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if req.Score1 == req.Score2 {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "引き分けは許可されていません",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if strings.TrimSpace(req.Winner) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "勝者は必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 試合結果を作成
	result := models.MatchResult{
		Score1: req.Score1,
		Score2: req.Score2,
		Winner: req.Winner,
	}

	// 試合結果提出
	err = h.matchService.UpdateMatchResult(id, result)
	if err != nil {
		if strings.Contains(err.Error(), "見つかりません") {
			c.JSON(http.StatusNotFound, ErrorResponse{
				Error:   "Not Found",
				Message: err.Error(),
				Code:    http.StatusNotFound,
			})
			return
		}

		if strings.Contains(err.Error(), "検証") || strings.Contains(err.Error(), "無効な") || 
		   strings.Contains(err.Error(), "一致しません") || strings.Contains(err.Error(), "引き分け") ||
		   strings.Contains(err.Error(), "完了している") || strings.Contains(err.Error(), "参加チーム") {
			c.JSON(http.StatusUnprocessableEntity, ErrorResponse{
				Error:   "Unprocessable Entity",
				Message: err.Error(),
				Code:    http.StatusUnprocessableEntity,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "試合結果の提出に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 更新された試合を取得
	match, err := h.matchService.GetMatch(id)
	if err != nil {
		// 結果提出は成功したが、取得に失敗した場合
		c.JSON(http.StatusOK, gin.H{
			"message": "試合結果を提出しました",
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, MatchResponse{
		Match:   match,
		Message: "試合結果を提出しました",
	})
}

// DeleteMatch は試合削除エンドポイントハンドラー
// @Summary 試合削除
// @Description 指定されたIDの試合を削除する（管理者のみ）
// @Tags matches
// @Produce json
// @Security BearerAuth
// @Param id path int true "試合ID"
// @Success 200 {object} map[string]string "削除成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 404 {object} ErrorResponse "未発見エラー"
// @Failure 409 {object} ErrorResponse "競合エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches/{id} [delete]
func (h *MatchHandler) DeleteMatch(c *gin.Context) {
	idStr := c.Param("id")

	// IDの変換
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効な試合IDです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 試合削除
	err = h.matchService.DeleteMatch(id)
	if err != nil {
		if strings.Contains(err.Error(), "完了した試合は削除できません") {
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
				Message: "削除対象の試合が見つかりません",
				Code:    http.StatusNotFound,
			})
			return
		}

		c.JSON(http.StatusInternalServerError, ErrorResponse{
			Error:   "Internal Server Error",
			Message: "試合の削除に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, gin.H{
		"message": "試合を削除しました",
	})
}

// GetMatchesByTournament はトーナメント別試合取得エンドポイントハンドラー
// @Summary トーナメント別試合取得
// @Description 指定されたトーナメントの試合を取得する
// @Tags matches
// @Produce json
// @Param tournament_id path int true "トーナメントID"
// @Success 200 {object} MatchListResponse "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches/tournament/{tournament_id} [get]
func (h *MatchHandler) GetMatchesByTournament(c *gin.Context) {
	tournamentIDStr := c.Param("tournament_id")

	// トーナメントIDの変換
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なトーナメントIDです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// トーナメント別試合取得
	matches, err := h.matchService.GetMatchesByTournament(tournamentID)
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
			Message: "試合の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, MatchListResponse{
		Matches: matches,
		Count:   len(matches),
		Message: "トーナメント別試合一覧を取得しました",
	})
}

// GetMatchStatistics は試合統計取得エンドポイントハンドラー
// @Summary 試合統計取得
// @Description 指定されたトーナメントの試合統計を取得する
// @Tags matches
// @Produce json
// @Param tournament_id path int true "トーナメントID"
// @Success 200 {object} MatchStatisticsResponse "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches/tournament/{tournament_id}/statistics [get]
func (h *MatchHandler) GetMatchStatistics(c *gin.Context) {
	tournamentIDStr := c.Param("tournament_id")

	// トーナメントIDの変換
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なトーナメントIDです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 試合統計取得
	statistics, err := h.matchService.GetMatchStatistics(tournamentID)
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
			Message: "試合統計の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, MatchStatisticsResponse{
		MatchStatistics: statistics,
		Message:         "試合統計を取得しました",
	})
}

// GetNextMatches は次の試合取得エンドポイントハンドラー
// @Summary 次の試合取得
// @Description 指定されたトーナメントの次に実施予定の試合を取得する
// @Tags matches
// @Produce json
// @Param tournament_id path int true "トーナメントID"
// @Success 200 {object} MatchListResponse "取得成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/matches/tournament/{tournament_id}/next [get]
func (h *MatchHandler) GetNextMatches(c *gin.Context) {
	tournamentIDStr := c.Param("tournament_id")

	// トーナメントIDの変換
	tournamentID, err := strconv.Atoi(tournamentIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なトーナメントIDです",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 次の試合取得
	matches, err := h.matchService.GetNextMatches(tournamentID)
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
			Message: "次の試合の取得に失敗しました",
			Code:    http.StatusInternalServerError,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, MatchListResponse{
		Matches: matches,
		Count:   len(matches),
		Message: "次の試合一覧を取得しました",
	})
}

// parseDateTime は日時文字列をtime.Timeに変換する
func parseDateTime(dateTimeStr string) (time.Time, error) {
	// 複数の日時フォーマットを試行
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, dateTimeStr); err == nil {
			return t, nil
		}
	}
	
	return time.Time{}, errors.New("サポートされていない日時形式です")
}
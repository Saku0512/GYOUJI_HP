package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/models"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMatchService はMatchServiceのモック
type MockMatchService struct {
	mock.Mock
}

func (m *MockMatchService) CreateMatch(match *models.Match) error {
	args := m.Called(match)
	return args.Error(0)
}

func (m *MockMatchService) GetMatch(id int) (*models.Match, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Match), args.Error(1)
}

func (m *MockMatchService) UpdateMatch(match *models.Match) error {
	args := m.Called(match)
	return args.Error(0)
}

func (m *MockMatchService) DeleteMatch(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockMatchService) UpdateMatchResult(matchID int, result models.MatchResult) error {
	args := m.Called(matchID, result)
	return args.Error(0)
}

func (m *MockMatchService) SubmitMatchResult(matchID int, score1, score2 int, winner string) error {
	args := m.Called(matchID, score1, score2, winner)
	return args.Error(0)
}

func (m *MockMatchService) AdvanceWinner(matchID int) error {
	args := m.Called(matchID)
	return args.Error(0)
}

func (m *MockMatchService) ProcessTournamentAdvancement(tournamentID int, round string) error {
	args := m.Called(tournamentID, round)
	return args.Error(0)
}

func (m *MockMatchService) GetMatchesBySport(sport string) ([]*models.Match, error) {
	args := m.Called(sport)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Match), args.Error(1)
}

func (m *MockMatchService) GetMatchesByTournament(tournamentID int) ([]*models.Match, error) {
	args := m.Called(tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Match), args.Error(1)
}

func (m *MockMatchService) GetMatchesByRound(tournamentID int, round string) ([]*models.Match, error) {
	args := m.Called(tournamentID, round)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Match), args.Error(1)
}

func (m *MockMatchService) GetPendingMatches() ([]*models.Match, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Match), args.Error(1)
}

func (m *MockMatchService) GetCompletedMatches() ([]*models.Match, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Match), args.Error(1)
}

func (m *MockMatchService) ValidateMatchResult(matchID int, result models.MatchResult) error {
	args := m.Called(matchID, result)
	return args.Error(0)
}

func (m *MockMatchService) ValidateMatchAdvancement(matchID int) error {
	args := m.Called(matchID)
	return args.Error(0)
}

func (m *MockMatchService) EnforceTournamentRules(tournamentID int) error {
	args := m.Called(tournamentID)
	return args.Error(0)
}

func (m *MockMatchService) GetMatchStatistics(tournamentID int) (*service.MatchStatistics, error) {
	args := m.Called(tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.MatchStatistics), args.Error(1)
}

func (m *MockMatchService) GetNextMatches(tournamentID int) ([]*models.Match, error) {
	args := m.Called(tournamentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Match), args.Error(1)
}

// setupMatchHandler はテスト用のMatchHandlerをセットアップする
func setupMatchHandler() (*MatchHandler, *MockMatchService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	
	mockService := new(MockMatchService)
	handler := NewMatchHandler(mockService)
	
	router := gin.New()
	
	// ルートを設定
	api := router.Group("/api")
	{
		matches := api.Group("/matches")
		{
			matches.POST("", handler.CreateMatch)
			matches.GET("", handler.GetMatches)
			matches.GET("/:sport", handler.GetMatchesBySport)
			matches.GET("/id/:id", handler.GetMatch)
			matches.PUT("/:id", handler.UpdateMatch)
			matches.DELETE("/:id", handler.DeleteMatch)
			matches.PUT("/:id/result", handler.SubmitMatchResult)
			matches.GET("/tournament/:tournament_id", handler.GetMatchesByTournament)
			matches.GET("/tournament/:tournament_id/statistics", handler.GetMatchStatistics)
			matches.GET("/tournament/:tournament_id/next", handler.GetNextMatches)
		}
	}
	
	return handler, mockService, router
}

// createSampleMatch はテスト用のサンプル試合を作成する
func createSampleMatch() *models.Match {
	now := time.Now()
	return &models.Match{
		ID:           1,
		TournamentID: 1,
		Round:        models.Round1stRound,
		Team1:        "チームA",
		Team2:        "チームB",
		Status:       models.MatchStatusPending,
		ScheduledAt:  now,
	}
}

// TestCreateMatch は試合作成のテスト
func TestCreateMatch(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		requestBody    CreateMatchRequest
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "正常な試合作成",
			requestBody: CreateMatchRequest{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "チームA",
				Team2:        "チームB",
				ScheduledAt:  "2024-01-01 10:00:00",
			},
			mockSetup: func() {
				mockService.On("CreateMatch", mock.AnythingOfType("*models.Match")).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "無効なリクエスト形式",
			requestBody: CreateMatchRequest{
				// TournamentIDが不足
				Round:       models.Round1stRound,
				Team1:       "チームA",
				Team2:       "チームB",
				ScheduledAt: "2024-01-01 10:00:00",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効なリクエスト形式です",
		},
		{
			name: "同じチーム同士の試合",
			requestBody: CreateMatchRequest{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "チームA",
				Team2:        "チームA",
				ScheduledAt:  "2024-01-01 10:00:00",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "同じチーム同士の試合はできません",
		},
		{
			name: "無効な日時形式",
			requestBody: CreateMatchRequest{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "チームA",
				Team2:        "チームB",
				ScheduledAt:  "invalid-date",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効な日時形式です",
		},
		{
			name: "試合作成サービスエラー",
			requestBody: CreateMatchRequest{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "チームA",
				Team2:        "チームB",
				ScheduledAt:  "2024-01-01 10:00:00",
			},
			mockSetup: func() {
				mockService.On("CreateMatch", mock.AnythingOfType("*models.Match")).Return(errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "試合の作成に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストボディを作成
			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/api/matches", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestGetMatches は試合一覧取得のテスト
func TestGetMatches(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		queryParam     string
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:       "未実施試合取得",
			queryParam: "?status=pending",
			mockSetup: func() {
				matches := []*models.Match{createSampleMatch()}
				mockService.On("GetPendingMatches").Return(matches, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:       "完了試合取得",
			queryParam: "?status=completed",
			mockSetup: func() {
				matches := []*models.Match{createSampleMatch()}
				mockService.On("GetCompletedMatches").Return(matches, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効なステータスパラメータ",
			queryParam:     "?status=invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "statusパラメータを指定してください",
		},
		{
			name:       "サービスエラー",
			queryParam: "?status=pending",
			mockSetup: func() {
				mockService.On("GetPendingMatches").Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "試合一覧の取得に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストを作成
			req, _ := http.NewRequest("GET", "/api/matches"+tt.queryParam, nil)

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestGetMatchesBySport はスポーツ別試合取得のテスト
func TestGetMatchesBySport(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		sport          string
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:  "正常なスポーツ別試合取得",
			sport: "volleyball",
			mockSetup: func() {
				matches := []*models.Match{createSampleMatch()}
				mockService.On("GetMatchesBySport", "volleyball").Return(matches, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "空のスポーツパラメータ",
			sport:          "",
			mockSetup:      func() {},
			expectedStatus: http.StatusMovedPermanently, // Ginのルーティングで301になる
		},
		{
			name:  "無効なスポーツ",
			sport: "invalid_sport",
			mockSetup: func() {
				mockService.On("GetMatchesBySport", "invalid_sport").Return(nil, errors.New("無効なスポーツです"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効なスポーツです",
		},
		{
			name:  "サービスエラー",
			sport: "volleyball",
			mockSetup: func() {
				mockService.On("GetMatchesBySport", "volleyball").Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "試合の取得に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストを作成
			url := "/api/matches/" + tt.sport
			req, _ := http.NewRequest("GET", url, nil)

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestGetMatch はID別試合取得のテスト
func TestGetMatch(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		matchID        string
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "正常な試合取得",
			matchID: "1",
			mockSetup: func() {
				match := createSampleMatch()
				mockService.On("GetMatch", 1).Return(match, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効な試合ID",
			matchID:        "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効な試合IDです",
		},
		{
			name:    "試合が見つからない",
			matchID: "999",
			mockSetup: func() {
				mockService.On("GetMatch", 999).Return(nil, errors.New("試合が見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "試合が見つかりません",
		},
		{
			name:    "サービスエラー",
			matchID: "1",
			mockSetup: func() {
				mockService.On("GetMatch", 1).Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "試合の取得に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストを作成
			url := "/api/matches/id/" + tt.matchID
			req, _ := http.NewRequest("GET", url, nil)

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestSubmitMatchResult は試合結果提出のテスト
func TestSubmitMatchResult(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		matchID        string
		requestBody    SubmitMatchResultRequest
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "正常な試合結果提出",
			matchID: "1",
			requestBody: SubmitMatchResultRequest{
				Score1: 3,
				Score2: 1,
				Winner: "チームA",
			},
			mockSetup: func() {
				result := models.MatchResult{Score1: 3, Score2: 1, Winner: "チームA"}
				mockService.On("UpdateMatchResult", 1, result).Return(nil)
				match := createSampleMatch()
				mockService.On("GetMatch", 1).Return(match, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効な試合ID",
			matchID:        "invalid",
			requestBody:    SubmitMatchResultRequest{Score1: 3, Score2: 1, Winner: "チームA"},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効な試合IDです",
		},
		{
			name:    "引き分けスコア",
			matchID: "1",
			requestBody: SubmitMatchResultRequest{
				Score1: 2,
				Score2: 2,
				Winner: "チームA",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "引き分けは許可されていません",
		},
		{
			name:    "負のスコア",
			matchID: "1",
			requestBody: SubmitMatchResultRequest{
				Score1: -1,
				Score2: 2,
				Winner: "チームA",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:    "空の勝者",
			matchID: "1",
			requestBody: SubmitMatchResultRequest{
				Score1: 3,
				Score2: 1,
				Winner: "",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効なリクエスト形式です", // Ginのバインディング検証でキャッチされる
		},
		{
			name:    "試合が見つからない",
			matchID: "999",
			requestBody: SubmitMatchResultRequest{
				Score1: 3,
				Score2: 1,
				Winner: "チームA",
			},
			mockSetup: func() {
				result := models.MatchResult{Score1: 3, Score2: 1, Winner: "チームA"}
				mockService.On("UpdateMatchResult", 999, result).Return(errors.New("試合が見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "試合が見つかりません",
		},
		{
			name:    "検証エラー",
			matchID: "1",
			requestBody: SubmitMatchResultRequest{
				Score1: 3,
				Score2: 1,
				Winner: "チームC",
			},
			mockSetup: func() {
				result := models.MatchResult{Score1: 3, Score2: 1, Winner: "チームC"}
				mockService.On("UpdateMatchResult", 1, result).Return(errors.New("勝者は参加チームのいずれかである必要があります"))
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "勝者は参加チームのいずれかである必要があります",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストボディを作成
			body, _ := json.Marshal(tt.requestBody)
			url := "/api/matches/" + tt.matchID + "/result"
			req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestUpdateMatch は試合更新のテスト
func TestUpdateMatch(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		matchID        string
		requestBody    UpdateMatchRequest
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "正常な試合更新",
			matchID: "1",
			requestBody: UpdateMatchRequest{
				Round:  models.RoundQuarterfinal,
				Status: models.MatchStatusCompleted,
			},
			mockSetup: func() {
				match := createSampleMatch()
				mockService.On("GetMatch", 1).Return(match, nil)
				mockService.On("UpdateMatch", mock.AnythingOfType("*models.Match")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効な試合ID",
			matchID:        "invalid",
			requestBody:    UpdateMatchRequest{Round: models.RoundQuarterfinal},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効な試合IDです",
		},
		{
			name:    "試合が見つからない",
			matchID: "999",
			requestBody: UpdateMatchRequest{
				Round: models.RoundQuarterfinal,
			},
			mockSetup: func() {
				mockService.On("GetMatch", 999).Return(nil, errors.New("試合が見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "試合が見つかりません",
		},
		{
			name:    "更新サービスエラー",
			matchID: "1",
			requestBody: UpdateMatchRequest{
				Round: models.RoundQuarterfinal,
			},
			mockSetup: func() {
				match := createSampleMatch()
				mockService.On("GetMatch", 1).Return(match, nil)
				mockService.On("UpdateMatch", mock.AnythingOfType("*models.Match")).Return(errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "試合の更新に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストボディを作成
			body, _ := json.Marshal(tt.requestBody)
			url := "/api/matches/" + tt.matchID
			req, _ := http.NewRequest("PUT", url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestDeleteMatch は試合削除のテスト
func TestDeleteMatch(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		matchID        string
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:    "正常な試合削除",
			matchID: "1",
			mockSetup: func() {
				mockService.On("DeleteMatch", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効な試合ID",
			matchID:        "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効な試合IDです",
		},
		{
			name:    "完了した試合の削除",
			matchID: "1",
			mockSetup: func() {
				mockService.On("DeleteMatch", 1).Return(errors.New("完了した試合は削除できません"))
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "完了した試合は削除できません",
		},
		{
			name:    "試合が見つからない",
			matchID: "999",
			mockSetup: func() {
				mockService.On("DeleteMatch", 999).Return(errors.New("削除対象の試合が見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "削除対象の試合が見つかりません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストを作成
			url := "/api/matches/" + tt.matchID
			req, _ := http.NewRequest("DELETE", url, nil)

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestGetMatchesByTournament はトーナメント別試合取得のテスト
func TestGetMatchesByTournament(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		tournamentID   string
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:         "正常なトーナメント別試合取得",
			tournamentID: "1",
			mockSetup: func() {
				matches := []*models.Match{createSampleMatch()}
				mockService.On("GetMatchesByTournament", 1).Return(matches, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効なトーナメントID",
			tournamentID:   "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効なトーナメントIDです",
		},
		{
			name:         "サービスエラー",
			tournamentID: "1",
			mockSetup: func() {
				mockService.On("GetMatchesByTournament", 1).Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "試合の取得に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストを作成
			url := "/api/matches/tournament/" + tt.tournamentID
			req, _ := http.NewRequest("GET", url, nil)

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestGetMatchStatistics は試合統計取得のテスト
func TestGetMatchStatistics(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		tournamentID   string
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:         "正常な試合統計取得",
			tournamentID: "1",
			mockSetup: func() {
				stats := &service.MatchStatistics{
					TournamentID:     1,
					TotalMatches:     10,
					CompletedMatches: 5,
					PendingMatches:   5,
					CompletionRate:   50.0,
				}
				mockService.On("GetMatchStatistics", 1).Return(stats, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効なトーナメントID",
			tournamentID:   "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効なトーナメントIDです",
		},
		{
			name:         "サービスエラー",
			tournamentID: "1",
			mockSetup: func() {
				mockService.On("GetMatchStatistics", 1).Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "試合統計の取得に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストを作成
			url := "/api/matches/tournament/" + tt.tournamentID + "/statistics"
			req, _ := http.NewRequest("GET", url, nil)

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestGetNextMatches は次の試合取得のテスト
func TestGetNextMatches(t *testing.T) {
	handler, mockService, router := setupMatchHandler()
	_ = handler

	tests := []struct {
		name           string
		tournamentID   string
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name:         "正常な次の試合取得",
			tournamentID: "1",
			mockSetup: func() {
				matches := []*models.Match{createSampleMatch()}
				mockService.On("GetNextMatches", 1).Return(matches, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効なトーナメントID",
			tournamentID:   "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効なトーナメントIDです",
		},
		{
			name:         "サービスエラー",
			tournamentID: "1",
			mockSetup: func() {
				mockService.On("GetNextMatches", 1).Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "次の試合の取得に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			mockService.Calls = nil
			
			// モックをセットアップ
			tt.mockSetup()

			// リクエストを作成
			url := "/api/matches/tournament/" + tt.tournamentID + "/next"
			req, _ := http.NewRequest("GET", url, nil)

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			// エラーメッセージを検証
			if tt.expectedError != "" {
				var response ErrorResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Contains(t, response.Message, tt.expectedError)
			}

			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestParseDateTime は日時パース関数のテスト
func TestParseDateTime(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "標準形式",
			input:       "2024-01-01 10:00:00",
			expectError: false,
		},
		{
			name:        "ISO形式",
			input:       "2024-01-01T10:00:00Z",
			expectError: false,
		},
		{
			name:        "ISO形式（タイムゾーンなし）",
			input:       "2024-01-01T10:00:00",
			expectError: false,
		},
		{
			name:        "分まで",
			input:       "2024-01-01 10:00",
			expectError: false,
		},
		{
			name:        "日付のみ",
			input:       "2024-01-01",
			expectError: false,
		},
		{
			name:        "無効な形式",
			input:       "invalid-date",
			expectError: true,
		},
		{
			name:        "空文字",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseDateTime(tt.input)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
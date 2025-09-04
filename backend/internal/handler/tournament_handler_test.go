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

// MockTournamentService はTournamentServiceのモック
type MockTournamentService struct {
	mock.Mock
}

func (m *MockTournamentService) CreateTournament(sport, format string) (*models.Tournament, error) {
	args := m.Called(sport, format)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tournament), args.Error(1)
}

func (m *MockTournamentService) GetTournament(sport string) (*models.Tournament, error) {
	args := m.Called(sport)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tournament), args.Error(1)
}

func (m *MockTournamentService) GetTournamentByID(id int) (*models.Tournament, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Tournament), args.Error(1)
}

func (m *MockTournamentService) UpdateTournament(tournament *models.Tournament) error {
	args := m.Called(tournament)
	return args.Error(0)
}

func (m *MockTournamentService) DeleteTournament(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTournamentService) GetTournamentBracket(sport string) (*models.Bracket, error) {
	args := m.Called(sport)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bracket), args.Error(1)
}

func (m *MockTournamentService) GenerateBracket(sport, format string, teams []string) (*models.Bracket, error) {
	args := m.Called(sport, format, teams)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Bracket), args.Error(1)
}

func (m *MockTournamentService) InitializeTournament(sport string, teams []string) error {
	args := m.Called(sport, teams)
	return args.Error(0)
}

func (m *MockTournamentService) SwitchTournamentFormat(sport, newFormat string) error {
	args := m.Called(sport, newFormat)
	return args.Error(0)
}

func (m *MockTournamentService) CompleteTournament(sport string) error {
	args := m.Called(sport)
	return args.Error(0)
}

func (m *MockTournamentService) ActivateTournament(sport string) error {
	args := m.Called(sport)
	return args.Error(0)
}

func (m *MockTournamentService) GetAllTournaments() ([]*models.Tournament, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Tournament), args.Error(1)
}

func (m *MockTournamentService) GetActiveTournaments() ([]*models.Tournament, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.Tournament), args.Error(1)
}

func (m *MockTournamentService) GetTournamentProgress(sport string) (*service.TournamentProgress, error) {
	args := m.Called(sport)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.TournamentProgress), args.Error(1)
}

// setupTournamentHandler はテスト用のハンドラーとルーターをセットアップする
func setupTournamentHandler() (*TournamentHandler, *MockTournamentService, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	
	mockService := &MockTournamentService{}
	handler := NewTournamentHandler(mockService)
	
	router := gin.New()
	
	// ルートを設定
	api := router.Group("/api")
	tournaments := api.Group("/tournaments")
	{
		tournaments.POST("", handler.CreateTournament)
		tournaments.GET("", handler.GetTournaments)
		tournaments.GET("/active", handler.GetActiveTournaments)
		tournaments.GET("/id/:id", handler.GetTournamentByID)
		tournaments.PUT("/id/:id", handler.UpdateTournament)
		tournaments.DELETE("/id/:id", handler.DeleteTournament)
		tournaments.GET("/:sport", handler.GetTournamentBySport)
		tournaments.GET("/:sport/bracket", handler.GetTournamentBracket)
		tournaments.PUT("/:sport/format", handler.SwitchTournamentFormat)
		tournaments.GET("/:sport/progress", handler.GetTournamentProgress)
		tournaments.PUT("/:sport/complete", handler.CompleteTournament)
		tournaments.PUT("/:sport/activate", handler.ActivateTournament)
	}
	
	return handler, mockService, router
}

// createSampleTournament はテスト用のサンプルトーナメントを作成する
func createSampleTournament() *models.Tournament {
	return &models.Tournament{
		ID:        1,
		Sport:     models.SportVolleyball,
		Format:    models.FormatStandard,
		Status:    models.TournamentStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// createSampleBracket はテスト用のサンプルブラケットを作成する
func createSampleBracket() *models.Bracket {
	return &models.Bracket{
		TournamentID: 1,
		Sport:        models.SportVolleyball,
		Format:       models.FormatStandard,
		Rounds: []models.Round{
			{
				Name: models.Round1stRound,
				Matches: []models.Match{
					{
						ID:           1,
						TournamentID: 1,
						Round:        models.Round1stRound,
						Team1:        "チームA",
						Team2:        "チームB",
						Status:       models.MatchStatusPending,
						ScheduledAt:  time.Now().Add(time.Hour),
					},
				},
			},
		},
	}
}

// createSampleProgress はテスト用のサンプル進行状況を作成する
func createSampleProgress() *service.TournamentProgress {
	return &service.TournamentProgress{
		TournamentID:     1,
		Sport:            models.SportVolleyball,
		Format:           models.FormatStandard,
		Status:           models.TournamentStatusActive,
		TotalMatches:     8,
		CompletedMatches: 4,
		PendingMatches:   4,
		ProgressPercent:  50.0,
		CurrentRound:     models.RoundQuarterfinal,
	}
}

func TestTournamentHandler_CreateTournament(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		requestBody    CreateTournamentRequest
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "正常なトーナメント作成",
			requestBody: CreateTournamentRequest{
				Sport:  models.SportVolleyball,
				Format: models.FormatStandard,
			},
			mockSetup: func() {
				tournament := createSampleTournament()
				mockService.On("CreateTournament", models.SportVolleyball, models.FormatStandard).Return(tournament, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "無効なリクエスト形式",
			requestBody: CreateTournamentRequest{
				Sport: "", // 空のスポーツ
				Format: models.FormatStandard,
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "無効なリクエスト形式です",
		},
		{
			name: "既存のアクティブトーナメント",
			requestBody: CreateTournamentRequest{
				Sport:  models.SportVolleyball,
				Format: models.FormatStandard,
			},
			mockSetup: func() {
				mockService.On("CreateTournament", models.SportVolleyball, models.FormatStandard).Return(nil, errors.New("既にアクティブなトーナメントが存在します"))
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "無効なスポーツ",
			requestBody: CreateTournamentRequest{
				Sport:  models.SportVolleyball,
				Format: models.FormatStandard,
			},
			mockSetup: func() {
				mockService.On("CreateTournament", models.SportVolleyball, models.FormatStandard).Return(nil, errors.New("無効なスポーツです"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "サーバーエラー",
			requestBody: CreateTournamentRequest{
				Sport:  models.SportVolleyball,
				Format: models.FormatStandard,
			},
			mockSetup: func() {
				mockService.On("CreateTournament", models.SportVolleyball, models.FormatStandard).Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストボディを作成
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/tournaments", bytes.NewBuffer(body))
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

func TestTournamentHandler_GetTournaments(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "正常なトーナメント一覧取得",
			mockSetup: func() {
				tournaments := []*models.Tournament{
					createSampleTournament(),
					{
						ID:        2,
						Sport:     models.SportTableTennis,
						Format:    models.FormatStandard,
						Status:    models.TournamentStatusActive,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}
				mockService.On("GetAllTournaments").Return(tournaments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "空のトーナメント一覧",
			mockSetup: func() {
				tournaments := []*models.Tournament{}
				mockService.On("GetAllTournaments").Return(tournaments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "サーバーエラー",
			mockSetup: func() {
				mockService.On("GetAllTournaments").Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストを作成
			req := httptest.NewRequest(http.MethodGet, "/api/tournaments", nil)
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// 成功の場合、レスポンスの内容を検証
			if tt.expectedStatus == http.StatusOK {
				var response TournamentListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, response.Count)
				assert.Equal(t, tt.expectedCount, len(response.Tournaments))
			}
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_GetTournamentBySport(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		sport          string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:  "正常なスポーツ別トーナメント取得",
			sport: models.SportVolleyball,
			mockSetup: func() {
				tournament := createSampleTournament()
				mockService.On("GetTournament", models.SportVolleyball).Return(tournament, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "無効なスポーツパラメータ",
			sport: "invalid_sport",
			mockSetup: func() {
				mockService.On("GetTournament", "invalid_sport").Return(nil, errors.New("無効なスポーツです"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "トーナメントが見つからない",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("GetTournament", models.SportVolleyball).Return(nil, errors.New("見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "無効なスポーツ",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("GetTournament", models.SportVolleyball).Return(nil, errors.New("無効なスポーツです"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "サーバーエラー",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("GetTournament", models.SportVolleyball).Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストを作成
			url := "/api/tournaments/" + tt.sport
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// 成功の場合、レスポンスの内容を検証
			if tt.expectedStatus == http.StatusOK {
				var response TournamentResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.SportVolleyball, response.Sport)
			}
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_GetTournamentByID(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		id             string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "正常なID別トーナメント取得",
			id:   "1",
			mockSetup: func() {
				tournament := createSampleTournament()
				mockService.On("GetTournamentByID", 1).Return(tournament, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効なID",
			id:             "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "トーナメントが見つからない",
			id:   "999",
			mockSetup: func() {
				mockService.On("GetTournamentByID", 999).Return(nil, errors.New("見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "サーバーエラー",
			id:   "1",
			mockSetup: func() {
				mockService.On("GetTournamentByID", 1).Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストを作成
			url := "/api/tournaments/id/" + tt.id
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_UpdateTournament(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		id             string
		requestBody    UpdateTournamentRequest
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "正常なトーナメント更新",
			id:   "1",
			requestBody: UpdateTournamentRequest{
				Format: models.FormatRainy,
				Status: models.TournamentStatusCompleted,
			},
			mockSetup: func() {
				tournament := createSampleTournament()
				mockService.On("GetTournamentByID", 1).Return(tournament, nil)
				
				// 更新されたトーナメントを期待
				updatedTournament := *tournament
				updatedTournament.Format = models.FormatRainy
				updatedTournament.Status = models.TournamentStatusCompleted
				mockService.On("UpdateTournament", &updatedTournament).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効なID",
			id:             "invalid",
			requestBody:    UpdateTournamentRequest{},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "トーナメントが見つからない",
			id:   "999",
			requestBody: UpdateTournamentRequest{
				Format: models.FormatRainy,
			},
			mockSetup: func() {
				mockService.On("GetTournamentByID", 999).Return(nil, errors.New("見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "更新エラー",
			id:   "1",
			requestBody: UpdateTournamentRequest{
				Format: "invalid_format",
			},
			mockSetup: func() {
				tournament := createSampleTournament()
				mockService.On("GetTournamentByID", 1).Return(tournament, nil)
				
				updatedTournament := *tournament
				updatedTournament.Format = "invalid_format"
				mockService.On("UpdateTournament", &updatedTournament).Return(errors.New("無効なフォーマットです"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストボディを作成
			body, _ := json.Marshal(tt.requestBody)
			url := "/api/tournaments/id/" + tt.id
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_DeleteTournament(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		id             string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name: "正常なトーナメント削除",
			id:   "1",
			mockSetup: func() {
				mockService.On("DeleteTournament", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "無効なID",
			id:             "invalid",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "試合が存在するトーナメント",
			id:   "1",
			mockSetup: func() {
				mockService.On("DeleteTournament", 1).Return(errors.New("試合が存在するトーナメントは削除できません"))
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "トーナメントが見つからない",
			id:   "999",
			mockSetup: func() {
				mockService.On("DeleteTournament", 999).Return(errors.New("見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストを作成
			url := "/api/tournaments/id/" + tt.id
			req := httptest.NewRequest(http.MethodDelete, url, nil)
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_GetTournamentBracket(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		sport          string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:  "正常なブラケット取得",
			sport: models.SportVolleyball,
			mockSetup: func() {
				bracket := createSampleBracket()
				mockService.On("GetTournamentBracket", models.SportVolleyball).Return(bracket, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "空のスポーツパラメータ",
			sport:          "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "ブラケットが見つからない",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("GetTournamentBracket", models.SportVolleyball).Return(nil, errors.New("見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:  "無効なスポーツ",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("GetTournamentBracket", models.SportVolleyball).Return(nil, errors.New("無効なスポーツです"))
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストを作成
			url := "/api/tournaments/" + tt.sport + "/bracket"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// 成功の場合、レスポンスの内容を検証
			if tt.expectedStatus == http.StatusOK {
				var response BracketResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.SportVolleyball, response.Sport)
			}
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_SwitchTournamentFormat(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		sport          string
		requestBody    SwitchFormatRequest
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:  "正常な形式切り替え",
			sport: models.SportTableTennis,
			requestBody: SwitchFormatRequest{
				Format: models.FormatRainy,
			},
			mockSetup: func() {
				mockService.On("SwitchTournamentFormat", models.SportTableTennis, models.FormatRainy).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:  "空のスポーツパラメータ",
			sport: "",
			requestBody: SwitchFormatRequest{
				Format: models.FormatRainy,
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "空のフォーマット",
			sport: models.SportTableTennis,
			requestBody: SwitchFormatRequest{
				Format: "",
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "サポートされていないスポーツ",
			sport: models.SportVolleyball,
			requestBody: SwitchFormatRequest{
				Format: models.FormatRainy,
			},
			mockSetup: func() {
				mockService.On("SwitchTournamentFormat", models.SportVolleyball, models.FormatRainy).Return(errors.New("サポートしていません"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "既に同じフォーマット",
			sport: models.SportTableTennis,
			requestBody: SwitchFormatRequest{
				Format: models.FormatStandard,
			},
			mockSetup: func() {
				mockService.On("SwitchTournamentFormat", models.SportTableTennis, models.FormatStandard).Return(errors.New("既に standard フォーマットです"))
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストボディを作成
			body, _ := json.Marshal(tt.requestBody)
			url := "/api/tournaments/" + tt.sport + "/format"
			req := httptest.NewRequest(http.MethodPut, url, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_GetActiveTournaments(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "正常なアクティブトーナメント取得",
			mockSetup: func() {
				tournaments := []*models.Tournament{
					createSampleTournament(),
				}
				mockService.On("GetActiveTournaments").Return(tournaments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name: "空のアクティブトーナメント一覧",
			mockSetup: func() {
				tournaments := []*models.Tournament{}
				mockService.On("GetActiveTournaments").Return(tournaments, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
		{
			name: "サーバーエラー",
			mockSetup: func() {
				mockService.On("GetActiveTournaments").Return(nil, errors.New("データベースエラー"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストを作成
			req := httptest.NewRequest(http.MethodGet, "/api/tournaments/active", nil)
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// 成功の場合、レスポンスの内容を検証
			if tt.expectedStatus == http.StatusOK {
				var response TournamentListResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, response.Count)
			}
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_GetTournamentProgress(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		sport          string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:  "正常な進行状況取得",
			sport: models.SportVolleyball,
			mockSetup: func() {
				progress := createSampleProgress()
				mockService.On("GetTournamentProgress", models.SportVolleyball).Return(progress, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "空のスポーツパラメータ",
			sport:          "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "トーナメントが見つからない",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("GetTournamentProgress", models.SportVolleyball).Return(nil, errors.New("見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストを作成
			url := "/api/tournaments/" + tt.sport + "/progress"
			req := httptest.NewRequest(http.MethodGet, url, nil)
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// 成功の場合、レスポンスの内容を検証
			if tt.expectedStatus == http.StatusOK {
				var response ProgressResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, models.SportVolleyball, response.Sport)
				assert.Equal(t, 50.0, response.ProgressPercent)
			}
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_CompleteTournament(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		sport          string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:  "正常なトーナメント完了",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("CompleteTournament", models.SportVolleyball).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "空のスポーツパラメータ",
			sport:          "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "既に完了済み",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("CompleteTournament", models.SportVolleyball).Return(errors.New("既に完了しています"))
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:  "試合が未完了",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("CompleteTournament", models.SportVolleyball).Return(errors.New("全ての試合が完了していないため、トーナメントを完了できません"))
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストを作成
			url := "/api/tournaments/" + tt.sport + "/complete"
			req := httptest.NewRequest(http.MethodPut, url, nil)
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

func TestTournamentHandler_ActivateTournament(t *testing.T) {
	_, mockService, router := setupTournamentHandler()

	tests := []struct {
		name           string
		sport          string
		mockSetup      func()
		expectedStatus int
	}{
		{
			name:  "正常なトーナメントアクティブ化",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("ActivateTournament", models.SportVolleyball).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "空のスポーツパラメータ",
			sport:          "",
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:  "既にアクティブ",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("ActivateTournament", models.SportVolleyball).Return(errors.New("既にアクティブです"))
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:  "トーナメントが見つからない",
			sport: models.SportVolleyball,
			mockSetup: func() {
				mockService.On("ActivateTournament", models.SportVolleyball).Return(errors.New("見つかりません"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックをリセット
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			// リクエストを作成
			url := "/api/tournaments/" + tt.sport + "/activate"
			req := httptest.NewRequest(http.MethodPut, url, nil)
			w := httptest.NewRecorder()
			
			// リクエストを実行
			router.ServeHTTP(w, req)
			
			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// モックの呼び出しを検証
			mockService.AssertExpectations(t)
		})
	}
}

// TestNewTournamentHandler はコンストラクタのテスト
func TestNewTournamentHandler(t *testing.T) {
	mockService := &MockTournamentService{}
	handler := NewTournamentHandler(mockService)
	
	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.tournamentService)
}
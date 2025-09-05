// Package testutil は統合テスト用のHTTPユーティリティ関数を提供する
package testutil

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/config"
	"backend/internal/repository"
	"backend/internal/router"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// TestServer はテスト用のHTTPサーバーを管理する
type TestServer struct {
	Router   *router.Router
	TestDB   *TestDB
	Services *TestServices
}

// TestServices はテスト用のサービス群を管理する
type TestServices struct {
	AuthService       service.AuthService
	TournamentService service.TournamentService
	MatchService      service.MatchService
}

// SetupTestServer はテスト用HTTPサーバーをセットアップする
func SetupTestServer(t *testing.T) *TestServer {
	// Ginをテストモードに設定
	gin.SetMode(gin.TestMode)

	// テストデータベースをセットアップ
	testDB := SetupTestDatabase(t)

	// サービス層を初期化
	services := initializeTestServices(t, testDB)

	// ルーターを初期化
	testRouter := router.NewRouter(
		services.AuthService,
		services.TournamentService,
		services.MatchService,
	)

	return &TestServer{
		Router:   testRouter,
		TestDB:   testDB,
		Services: services,
	}
}

// TeardownTestServer はテスト用HTTPサーバーをクリーンアップする
func (ts *TestServer) TeardownTestServer(t *testing.T) {
	if ts.TestDB != nil {
		ts.TestDB.TeardownTestDatabase(t)
	}
}

// MakeRequest はHTTPリクエストを作成して実行する
func (ts *TestServer) MakeRequest(t *testing.T, method, path string, body interface{}, headers map[string]string) *httptest.ResponseRecorder {
	var requestBody *bytes.Buffer
	
	if body != nil {
		jsonBody, err := json.Marshal(body)
		require.NoError(t, err, "リクエストボディのJSONエンコードに失敗しました")
		requestBody = bytes.NewBuffer(jsonBody)
	} else {
		requestBody = bytes.NewBuffer([]byte{})
	}

	req, err := http.NewRequest(method, path, requestBody)
	require.NoError(t, err, "HTTPリクエストの作成に失敗しました")

	// デフォルトヘッダーを設定
	req.Header.Set("Content-Type", "application/json")

	// カスタムヘッダーを設定
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// レスポンスレコーダーを作成
	w := httptest.NewRecorder()

	// リクエストを実行
	ts.Router.GetEngine().ServeHTTP(w, req)

	return w
}

// MakeAuthenticatedRequest は認証付きHTTPリクエストを作成して実行する
func (ts *TestServer) MakeAuthenticatedRequest(t *testing.T, method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	headers := map[string]string{
		"Authorization": "Bearer " + token,
	}
	return ts.MakeRequest(t, method, path, body, headers)
}

// LoginAndGetToken は管理者でログインしてJWTトークンを取得する
func (ts *TestServer) LoginAndGetToken(t *testing.T) string {
	loginRequest := map[string]string{
		"username": "admin",
		"password": "password",
	}

	w := ts.MakeRequest(t, "POST", "/api/auth/login", loginRequest, nil)
	require.Equal(t, http.StatusOK, w.Code, "ログインに失敗しました")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "ログインレスポンスのパースに失敗しました")

	token, exists := response["token"].(string)
	require.True(t, exists, "レスポンスにトークンが含まれていません")
	require.NotEmpty(t, token, "トークンが空です")

	return token
}

// AssertJSONResponse はJSONレスポンスをアサートする
func (ts *TestServer) AssertJSONResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedFields map[string]interface{}) {
	require.Equal(t, expectedStatus, w.Code, "HTTPステータスコードが期待値と異なります")
	require.Equal(t, "application/json; charset=utf-8", w.Header().Get("Content-Type"), "Content-Typeが期待値と異なります")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "レスポンスのJSONパースに失敗しました")

	for key, expectedValue := range expectedFields {
		actualValue, exists := response[key]
		require.True(t, exists, fmt.Sprintf("レスポンスに%sフィールドが含まれていません", key))
		require.Equal(t, expectedValue, actualValue, fmt.Sprintf("%sフィールドの値が期待値と異なります", key))
	}
}

// AssertErrorResponse はエラーレスポンスをアサートする
func (ts *TestServer) AssertErrorResponse(t *testing.T, w *httptest.ResponseRecorder, expectedStatus int, expectedMessage string) {
	require.Equal(t, expectedStatus, w.Code, "HTTPステータスコードが期待値と異なります")

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err, "エラーレスポンスのJSONパースに失敗しました")

	message, exists := response["message"].(string)
	require.True(t, exists, "エラーレスポンスにmessageフィールドが含まれていません")
	require.Contains(t, message, expectedMessage, "エラーメッセージが期待値を含んでいません")
}

// initializeTestServices はテスト用サービス群を初期化する
func initializeTestServices(t *testing.T, testDB *TestDB) *TestServices {
	// 設定を読み込み
	cfg, err := config.Load()
	require.NoError(t, err, "設定の読み込みに失敗しました")

	// リポジトリ層を初期化
	userRepo := repository.NewUserRepository(testDB.DB)
	tournamentRepo := repository.NewTournamentRepository(testDB.DB)
	matchRepo := repository.NewMatchRepository(testDB.DB)

	// サービス層を初期化
	authService := service.NewAuthService(userRepo, cfg)
	tournamentService := service.NewTournamentService(tournamentRepo, matchRepo)
	matchService := service.NewMatchService(matchRepo, tournamentRepo)

	return &TestServices{
		AuthService:       authService,
		TournamentService: tournamentService,
		MatchService:      matchService,
	}
}
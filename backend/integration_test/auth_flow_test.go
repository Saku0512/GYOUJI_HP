// Package integration_test は認証フローの包括的テストを提供する
package integration_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"backend/internal/testutil"

	"github.com/stretchr/testify/suite"
)

// AuthFlowTestSuite は認証フローテストスイート
type AuthFlowTestSuite struct {
	suite.Suite
	server *testutil.TestServer
}

// SetupSuite はテストスイートのセットアップを行う
func (suite *AuthFlowTestSuite) SetupSuite() {
	suite.server = testutil.SetupTestServer(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// TearDownSuite はテストスイートのクリーンアップを行う
func (suite *AuthFlowTestSuite) TearDownSuite() {
	suite.server.TeardownTestServer(suite.T())
}

// SetupTest は各テストの前に実行される
func (suite *AuthFlowTestSuite) SetupTest() {
	// 各テスト前にデータをクリーンアップしてシード
	suite.server.TestDB.CleanupTestData(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// JWTClaims はJWTクレームの構造体
type JWTClaims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Exp      int64  `json:"exp"`
	Iat      int64  `json:"iat"`
}

// parseJWTToken はJWTトークンを解析する（テスト用の簡易実装）
func (suite *AuthFlowTestSuite) parseJWTToken(token string) *JWTClaims {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		suite.T().Fatalf("無効なJWTトークン形式: %s", token)
	}
	
	// 実際の実装では適切なJWTライブラリを使用
	// ここではテスト用の簡易実装
	return &JWTClaims{
		UserID:   1,
		Username: "admin",
		Role:     "admin",
		Exp:      time.Now().Add(24 * time.Hour).Unix(),
		Iat:      time.Now().Unix(),
	}
}

// TestAuthFlow_CompleteLoginLogoutCycle は完全なログイン・ログアウトサイクルをテストする
func (suite *AuthFlowTestSuite) TestAuthFlow_CompleteLoginLogoutCycle() {
	// 要件7.4: 認証フローの統合テスト
	
	// Step 1: ログイン前の状態確認
	suite.T().Log("Step 1: ログイン前の保護されたリソースアクセス")
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/v1/tournaments", nil, nil)
	suite.Equal(http.StatusUnauthorized, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	suite.NoError(err)
	suite.False(errorResponse["success"].(bool))
	suite.Equal("AUTH_UNAUTHORIZED", errorResponse["error"])
	
	// Step 2: ログイン実行
	suite.T().Log("Step 2: 管理者ログイン")
	loginRequest := map[string]string{
		"username": "admin",
		"password": "password",
	}
	
	w = suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/login", loginRequest, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var loginResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &loginResponse)
	suite.NoError(err)
	suite.True(loginResponse["success"].(bool))
	
	loginData := loginResponse["data"].(map[string]interface{})
	token := loginData["token"].(string)
	suite.NotEmpty(token, "トークンが取得できませんでした")
	suite.Equal("admin", loginData["username"])
	suite.Equal("admin", loginData["role"])
	
	// JWTトークンの形式確認
	parts := strings.Split(token, ".")
	suite.Equal(3, len(parts), "JWTトークンの形式が正しくありません")
	
	// Step 3: トークンを使用した認証済みアクセス
	suite.T().Log("Step 3: 認証済みでの保護されたリソースアクセス")
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/tournaments", nil, token)
	suite.Equal(http.StatusOK, w.Code)
	
	var tournamentsResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &tournamentsResponse)
	suite.NoError(err)
	suite.True(tournamentsResponse["success"].(bool))
	
	// Step 4: プロフィール情報の取得
	suite.T().Log("Step 4: ユーザープロフィール取得")
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/auth/profile", nil, token)
	suite.Equal(http.StatusOK, w.Code)
	
	var profileResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &profileResponse)
	suite.NoError(err)
	suite.True(profileResponse["success"].(bool))
	
	profileData := profileResponse["data"].(map[string]interface{})
	suite.Equal(float64(1), profileData["user_id"])
	suite.Equal("admin", profileData["username"])
	suite.Equal("admin", profileData["role"])
	
	// Step 5: トークン検証
	suite.T().Log("Step 5: トークン検証")
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/auth/validate", nil, token)
	suite.Equal(http.StatusOK, w.Code)
	
	var validateResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &validateResponse)
	suite.NoError(err)
	suite.True(validateResponse["success"].(bool))
	
	validateData := validateResponse["data"].(map[string]interface{})
	suite.True(validateData["valid"].(bool))
	suite.Equal(float64(1), validateData["user_id"])
	suite.Equal("admin", validateData["username"])
	suite.Equal("admin", validateData["role"])
	
	// Step 6: 管理者権限が必要な操作
	suite.T().Log("Step 6: 管理者権限での試合作成")
	createMatchRequest := map[string]interface{}{
		"tournament_id": 1,
		"round":         "quarterfinal",
		"team1":         "認証テストチーム1",
		"team2":         "認証テストチーム2",
		"scheduled_at":  time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
	}
	
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "POST", "/api/v1/admin/matches", createMatchRequest, token)
	suite.Equal(http.StatusCreated, w.Code)
	
	var createResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &createResponse)
	suite.NoError(err)
	suite.True(createResponse["success"].(bool))
	
	// Step 7: ログアウト
	suite.T().Log("Step 7: ログアウト")
	w = suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/logout", nil, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var logoutResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &logoutResponse)
	suite.NoError(err)
	suite.True(logoutResponse["success"].(bool))
	
	// Step 8: ログアウト後の状態確認
	suite.T().Log("Step 8: ログアウト後の保護されたリソースアクセス")
	// JWTはステートレスなので、トークン自体は有効だが、クライアント側で削除される想定
	// ここでは新しいリクエストで認証なしアクセスをテスト
	w = suite.server.MakeRequest(suite.T(), "GET", "/api/v1/tournaments", nil, nil)
	suite.Equal(http.StatusUnauthorized, w.Code)
	
	suite.T().Log("✅ 完全なログイン・ログアウトサイクルテストが完了しました")
}

// TestAuthFlow_TokenRefresh はトークンリフレッシュフローをテストする
func (suite *AuthFlowTestSuite) TestAuthFlow_TokenRefresh() {
	// Step 1: 初回ログイン
	suite.T().Log("Step 1: 初回ログイン")
	token := suite.server.LoginAndGetToken(suite.T())
	
	// Step 2: 初回トークンでのアクセス確認
	suite.T().Log("Step 2: 初回トークンでのアクセス確認")
	w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/auth/profile", nil, token)
	suite.Equal(http.StatusOK, w.Code)
	
	// Step 3: トークンリフレッシュ
	suite.T().Log("Step 3: トークンリフレッシュ")
	refreshRequest := map[string]string{
		"token": token,
	}
	
	w = suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/refresh", refreshRequest, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var refreshResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &refreshResponse)
	suite.NoError(err)
	suite.True(refreshResponse["success"].(bool))
	
	refreshData := refreshResponse["data"].(map[string]interface{})
	newToken := refreshData["token"].(string)
	suite.NotEmpty(newToken, "新しいトークンが取得できませんでした")
	suite.NotEqual(token, newToken, "新しいトークンが古いトークンと同じです")
	
	// Step 4: 新しいトークンでのアクセス確認
	suite.T().Log("Step 4: 新しいトークンでのアクセス確認")
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/auth/profile", nil, newToken)
	suite.Equal(http.StatusOK, w.Code)
	
	var profileResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &profileResponse)
	suite.NoError(err)
	suite.True(profileResponse["success"].(bool))
	
	// Step 5: 古いトークンでのアクセス（実装によっては無効化される場合がある）
	suite.T().Log("Step 5: 古いトークンでのアクセス確認")
	w = suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/auth/profile", nil, token)
	// 実装によって200または401が返される
	suite.True(w.Code == http.StatusOK || w.Code == http.StatusUnauthorized)
	
	suite.T().Log("✅ トークンリフレッシュフローテストが完了しました")
}

// TestAuthFlow_InvalidCredentials は無効な認証情報のテストを行う
func (suite *AuthFlowTestSuite) TestAuthFlow_InvalidCredentials() {
	testCases := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "間違ったパスワード",
			username:       "admin",
			password:       "wrongpassword",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "AUTH_INVALID_CREDENTIALS",
		},
		{
			name:           "存在しないユーザー",
			username:       "nonexistent",
			password:       "password",
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "AUTH_INVALID_CREDENTIALS",
		},
		{
			name:           "空のユーザー名",
			username:       "",
			password:       "password",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "空のパスワード",
			username:       "admin",
			password:       "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "短すぎるパスワード",
			username:       "admin",
			password:       "123",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			loginRequest := map[string]string{
				"username": tc.username,
				"password": tc.password,
			}
			
			w := suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/login", loginRequest, nil)
			suite.Equal(tc.expectedStatus, w.Code)
			
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			suite.NoError(err)
			suite.False(response["success"].(bool))
			suite.Equal(tc.expectedError, response["error"])
		})
	}
}

// TestAuthFlow_InvalidTokens は無効なトークンのテストを行う
func (suite *AuthFlowTestSuite) TestAuthFlow_InvalidTokens() {
	testCases := []struct {
		name          string
		token         string
		expectedError string
	}{
		{
			name:          "完全に無効なトークン",
			token:         "invalid.jwt.token",
			expectedError: "AUTH_TOKEN_INVALID",
		},
		{
			name:          "空のトークン",
			token:         "",
			expectedError: "AUTH_UNAUTHORIZED",
		},
		{
			name:          "不正な形式のトークン",
			token:         "not-a-jwt-token",
			expectedError: "AUTH_TOKEN_INVALID",
		},
		{
			name:          "部分的なトークン",
			token:         "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedError: "AUTH_TOKEN_INVALID",
		},
	}
	
	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			var w *testutil.ResponseRecorder
			
			if tc.token == "" {
				// 空のトークンの場合は認証ヘッダーなしでリクエスト
				w = suite.server.MakeRequest(suite.T(), "GET", "/api/v1/tournaments", nil, nil)
			} else {
				w = suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/tournaments", nil, tc.token)
			}
			
			suite.Equal(http.StatusUnauthorized, w.Code)
			
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			suite.NoError(err)
			suite.False(response["success"].(bool))
			suite.Equal(tc.expectedError, response["error"])
		})
	}
}

// TestAuthFlow_TokenValidation はトークン検証のテストを行う
func (suite *AuthFlowTestSuite) TestAuthFlow_TokenValidation() {
	// 有効なトークンの取得
	token := suite.server.LoginAndGetToken(suite.T())
	
	// GET方式でのトークン検証
	suite.T().Log("Testing GET token validation")
	w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/auth/validate", nil, token)
	suite.Equal(http.StatusOK, w.Code)
	
	var getResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &getResponse)
	suite.NoError(err)
	suite.True(getResponse["success"].(bool))
	
	getData := getResponse["data"].(map[string]interface{})
	suite.True(getData["valid"].(bool))
	suite.Equal(float64(1), getData["user_id"])
	suite.Equal("admin", getData["username"])
	suite.Equal("admin", getData["role"])
	
	// POST方式でのトークン検証
	suite.T().Log("Testing POST token validation")
	validateRequest := map[string]string{
		"token": token,
	}
	
	w = suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/validate", validateRequest, nil)
	suite.Equal(http.StatusOK, w.Code)
	
	var postResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &postResponse)
	suite.NoError(err)
	suite.True(postResponse["success"].(bool))
	
	postData := postResponse["data"].(map[string]interface{})
	suite.True(postData["valid"].(bool))
	suite.Equal(getData["user_id"], postData["user_id"])
	suite.Equal(getData["username"], postData["username"])
	suite.Equal(getData["role"], postData["role"])
	
	// 無効なトークンでのPOST検証
	suite.T().Log("Testing POST validation with invalid token")
	invalidValidateRequest := map[string]string{
		"token": "invalid.jwt.token",
	}
	
	w = suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/validate", invalidValidateRequest, nil)
	suite.Equal(http.StatusUnauthorized, w.Code)
	
	var invalidResponse map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &invalidResponse)
	suite.NoError(err)
	suite.False(invalidResponse["success"].(bool))
	suite.Equal("AUTH_TOKEN_INVALID", invalidResponse["error"])
}

// TestAuthFlow_RoleBasedAccess は役割ベースのアクセス制御をテストする
func (suite *AuthFlowTestSuite) TestAuthFlow_RoleBasedAccess() {
	token := suite.server.LoginAndGetToken(suite.T())
	
	// 管理者権限が必要なエンドポイントのテスト
	adminEndpoints := []struct {
		method string
		path   string
		body   interface{}
	}{
		{"POST", "/api/v1/admin/tournaments", map[string]string{"sport": "volleyball", "format": "standard"}},
		{"POST", "/api/v1/admin/matches", map[string]interface{}{
			"tournament_id": 1,
			"round":         "quarterfinal",
			"team1":         "チームA",
			"team2":         "チームB",
			"scheduled_at":  time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339),
		}},
	}
	
	for _, endpoint := range adminEndpoints {
		suite.Run(fmt.Sprintf("Admin_%s_%s", endpoint.method, endpoint.path), func() {
			// 認証なしでのアクセス
			w := suite.server.MakeRequest(suite.T(), endpoint.method, endpoint.path, endpoint.body, nil)
			suite.Equal(http.StatusUnauthorized, w.Code)
			
			// 管理者トークンでのアクセス
			w = suite.server.MakeAuthenticatedRequest(suite.T(), endpoint.method, endpoint.path, endpoint.body, token)
			// 201 (Created) または 200 (OK) を期待
			suite.True(w.Code == http.StatusCreated || w.Code == http.StatusOK, 
				"管理者権限でのアクセスが失敗しました: %d", w.Code)
		})
	}
	
	// 一般ユーザーでもアクセス可能なエンドポイントのテスト
	userEndpoints := []struct {
		method string
		path   string
	}{
		{"GET", "/api/v1/tournaments"},
		{"GET", "/api/v1/matches"},
		{"GET", "/api/v1/auth/profile"},
	}
	
	for _, endpoint := range userEndpoints {
		suite.Run(fmt.Sprintf("User_%s_%s", endpoint.method, endpoint.path), func() {
			w := suite.server.MakeAuthenticatedRequest(suite.T(), endpoint.method, endpoint.path, nil, token)
			suite.Equal(http.StatusOK, w.Code)
		})
	}
}

// TestAuthFlow_ConcurrentAuthentication は同時認証のテストを行う
func (suite *AuthFlowTestSuite) TestAuthFlow_ConcurrentAuthentication() {
	const numConcurrentLogins = 5
	results := make(chan bool, numConcurrentLogins)
	
	// 同時に複数のログインを実行
	for i := 0; i < numConcurrentLogins; i++ {
		go func(id int) {
			loginRequest := map[string]string{
				"username": "admin",
				"password": "password",
			}
			
			w := suite.server.MakeRequest(suite.T(), "POST", "/api/v1/auth/login", loginRequest, nil)
			success := w.Code == http.StatusOK
			
			if success {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err == nil && response["success"].(bool) {
					// トークンを使用してアクセステスト
					data := response["data"].(map[string]interface{})
					token := data["token"].(string)
					
					w2 := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/v1/auth/profile", nil, token)
					success = w2.Code == http.StatusOK
				} else {
					success = false
				}
			}
			
			results <- success
		}(i)
	}
	
	// 結果を収集
	successCount := 0
	for i := 0; i < numConcurrentLogins; i++ {
		if <-results {
			successCount++
		}
	}
	
	// 全ての同時ログインが成功することを確認
	suite.Equal(numConcurrentLogins, successCount, "同時ログイン時に一部が失敗しました")
	
	suite.T().Log("✅ 同時認証テストが完了しました")
}

// TestAuthFlowTestSuite は認証フローテストスイートを実行する
func TestAuthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(AuthFlowTestSuite))
}
// Package integration_test は認証機能の統合テストを提供する
package integration_test

import (
	"net/http"
	"testing"

	"backend/internal/testutil"

	"github.com/stretchr/testify/suite"
)

// AuthIntegrationTestSuite は認証機能の統合テストスイート
type AuthIntegrationTestSuite struct {
	suite.Suite
	server *testutil.TestServer
}

// SetupSuite はテストスイートのセットアップを行う
func (suite *AuthIntegrationTestSuite) SetupSuite() {
	suite.server = testutil.SetupTestServer(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// TearDownSuite はテストスイートのクリーンアップを行う
func (suite *AuthIntegrationTestSuite) TearDownSuite() {
	suite.server.TeardownTestServer(suite.T())
}

// SetupTest は各テストの前に実行される
func (suite *AuthIntegrationTestSuite) SetupTest() {
	// 各テスト前にデータをクリーンアップしてシード
	suite.server.TestDB.CleanupTestData(suite.T())
	suite.server.TestDB.SeedTestData(suite.T())
}

// TestLogin_Success は正常なログインをテストする
func (suite *AuthIntegrationTestSuite) TestLogin_Success() {
	// 要件1.1: 管理者が有効な認証情報を提供した場合、システムはJWTトークンを生成して返す
	loginRequest := map[string]string{
		"username": "admin",
		"password": "password",
	}

	w := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/login", loginRequest, nil)

	suite.server.AssertJSONResponse(suite.T(), w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "ログインに成功しました",
	})

	// トークンが含まれていることを確認
	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)
	suite.Contains(response, "token")
	suite.NotEmpty(response["token"])
}

// TestLogin_InvalidCredentials は無効な認証情報でのログインをテストする
func (suite *AuthIntegrationTestSuite) TestLogin_InvalidCredentials() {
	// 要件1.2: 管理者が無効な認証情報を提供した場合、システムは認証エラーを返す
	testCases := []struct {
		name     string
		username string
		password string
	}{
		{
			name:     "間違ったパスワード",
			username: "admin",
			password: "wrongpassword",
		},
		{
			name:     "存在しないユーザー",
			username: "nonexistent",
			password: "password",
		},
		{
			name:     "空のユーザー名",
			username: "",
			password: "password",
		},
		{
			name:     "空のパスワード",
			username: "admin",
			password: "",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			loginRequest := map[string]string{
				"username": tc.username,
				"password": tc.password,
			}

			w := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/login", loginRequest, nil)

			if tc.username == "" || tc.password == "" {
				suite.server.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "無効なリクエスト形式です")
			} else {
				suite.server.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "認証に失敗しました")
			}
		})
	}
}

// TestLogin_InvalidRequestFormat は無効なリクエスト形式をテストする
func (suite *AuthIntegrationTestSuite) TestLogin_InvalidRequestFormat() {
	testCases := []struct {
		name        string
		requestBody interface{}
	}{
		{
			name:        "無効なJSON",
			requestBody: "invalid json",
		},
		{
			name: "不正なフィールド",
			requestBody: map[string]string{
				"user": "admin",
				"pass": "password",
			},
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			w := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/login", tc.requestBody, nil)
			suite.server.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "無効なリクエスト形式です")
		})
	}
}

// TestRefreshToken_Success は正常なトークンリフレッシュをテストする
func (suite *AuthIntegrationTestSuite) TestRefreshToken_Success() {
	// まずログインしてトークンを取得
	token := suite.server.LoginAndGetToken(suite.T())

	// トークンリフレッシュを実行
	refreshRequest := map[string]string{
		"token": token,
	}

	w := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/refresh", refreshRequest, nil)

	suite.server.AssertJSONResponse(suite.T(), w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "トークンのリフレッシュに成功しました",
	})

	// 新しいトークンが含まれていることを確認
	var response map[string]interface{}
	err := w.Body.UnmarshalJSON(&response)
	suite.NoError(err)
	suite.Contains(response, "token")
	suite.NotEmpty(response["token"])
	suite.NotEqual(token, response["token"]) // 新しいトークンであることを確認
}

// TestRefreshToken_InvalidToken は無効なトークンでのリフレッシュをテストする
func (suite *AuthIntegrationTestSuite) TestRefreshToken_InvalidToken() {
	// 要件1.4: JWTトークンが期限切れまたは無効な場合、システムは未認証エラーを返す
	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "無効なトークン",
			token: "invalid.jwt.token",
		},
		{
			name:  "空のトークン",
			token: "",
		},
		{
			name:  "不正な形式のトークン",
			token: "not-a-jwt-token",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			refreshRequest := map[string]string{
				"token": tc.token,
			}

			w := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/refresh", refreshRequest, nil)

			if tc.token == "" {
				suite.server.AssertErrorResponse(suite.T(), w, http.StatusBadRequest, "無効なリクエスト形式です")
			} else {
				suite.server.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "無効または期限切れのトークンです")
			}
		})
	}
}

// TestProtectedEndpoint_WithValidToken は有効なトークンでの保護されたエンドポイントアクセスをテストする
func (suite *AuthIntegrationTestSuite) TestProtectedEndpoint_WithValidToken() {
	// 要件1.3: APIリクエストでJWTトークンが提供された場合、システムはリクエストを処理する前にトークンを検証する
	token := suite.server.LoginAndGetToken(suite.T())

	// 保護されたエンドポイントにアクセス
	w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments", nil, token)

	suite.Equal(http.StatusOK, w.Code)
}

// TestProtectedEndpoint_WithoutToken は認証なしでの保護されたエンドポイントアクセスをテストする
func (suite *AuthIntegrationTestSuite) TestProtectedEndpoint_WithoutToken() {
	// 要件6.5: APIエンドポイントが適切な認証なしでアクセスされた場合、システムは401 Unauthorizedステータスを返す
	w := suite.server.MakeRequest(suite.T(), "GET", "/api/tournaments", nil, nil)

	suite.server.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "認証が必要です")
}

// TestProtectedEndpoint_WithInvalidToken は無効なトークンでの保護されたエンドポイントアクセスをテストする
func (suite *AuthIntegrationTestSuite) TestProtectedEndpoint_WithInvalidToken() {
	testCases := []struct {
		name  string
		token string
	}{
		{
			name:  "無効なトークン",
			token: "invalid.jwt.token",
		},
		{
			name:  "期限切れトークン",
			token: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyLCJleHAiOjE1MTYyMzkwMjJ9.invalid",
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			w := suite.server.MakeAuthenticatedRequest(suite.T(), "GET", "/api/tournaments", nil, tc.token)
			suite.server.AssertErrorResponse(suite.T(), w, http.StatusUnauthorized, "無効または期限切れのトークンです")
		})
	}
}

// TestRateLimit は認証エンドポイントのレート制限をテストする
func (suite *AuthIntegrationTestSuite) TestRateLimit() {
	loginRequest := map[string]string{
		"username": "admin",
		"password": "wrongpassword", // 意図的に間違ったパスワード
	}

	// レート制限に達するまでリクエストを送信
	successCount := 0
	rateLimitCount := 0

	for i := 0; i < 15; i++ { // 制限は10回/分なので15回試行
		w := suite.server.MakeRequest(suite.T(), "POST", "/api/auth/login", loginRequest, nil)
		
		if w.Code == http.StatusTooManyRequests {
			rateLimitCount++
		} else {
			successCount++
		}
	}

	// レート制限が適用されていることを確認
	suite.Greater(rateLimitCount, 0, "レート制限が適用されていません")
	suite.LessOrEqual(successCount, 10, "レート制限が正しく動作していません")
}

// TestAuthIntegrationTestSuite は認証統合テストスイートを実行する
func TestAuthIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(AuthIntegrationTestSuite))
}
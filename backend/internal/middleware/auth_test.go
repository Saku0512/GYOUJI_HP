package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService はテスト用のAuthServiceモック
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*service.JWTClaims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.JWTClaims), args.Error(1)
}

func (m *MockAuthService) GenerateToken(userID int, username string) (string, error) {
	args := m.Called(userID, username)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) RefreshToken(tokenString string) (string, error) {
	args := m.Called(tokenString)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) HashPassword(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) VerifyPassword(hashedPassword, password string) error {
	args := m.Called(hashedPassword, password)
	return args.Error(0)
}

// テスト用のJWTクレームを作成
func createTestClaims() *service.JWTClaims {
	return &service.JWTClaims{
		UserID:   1,
		Username: "admin",
		Role:     models.RoleAdmin,
	}
}

// テスト用のGinエンジンを作成
func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "認証ヘッダーなし",
			authHeader:     "",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  models.ErrorAuthUnauthorized,
		},
		{
			name:           "無効な認証ヘッダー形式",
			authHeader:     "InvalidFormat",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  models.ErrorAuthTokenInvalid,
		},
		{
			name:           "Bearerプレフィックスなし",
			authHeader:     "Token abc123",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  models.ErrorAuthTokenInvalid,
		},
		{
			name:           "空のトークン",
			authHeader:     "Bearer ",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  models.ErrorAuthTokenInvalid,
		},
		{
			name:       "無効なトークン",
			authHeader: "Bearer invalid-token",
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  models.ErrorAuthTokenInvalid,
		},
		{
			name:       "有効なトークン",
			authHeader: "Bearer valid-token",
			mockSetup: func(m *MockAuthService) {
				claims := createTestClaims()
				m.On("ValidateToken", "valid-token").Return(claims, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスを作成
			mockAuthService := new(MockAuthService)
			tt.mockSetup(mockAuthService)

			// ミドルウェアを作成
			authMiddleware := NewAuthMiddleware(mockAuthService)

			// テスト用ルーターを設定
			router := setupTestRouter()
			router.GET("/protected", authMiddleware.RequireAuth(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// テストリクエストを作成
			req := httptest.NewRequest("GET", "/protected", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus != http.StatusOK {
				// エラーレスポンスを検証
				var response models.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, tt.expectedError, response.Error)
			}

			// モックの呼び出しを検証
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthMiddleware_RequireAdmin(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "認証情報なし",
			setupContext: func(c *gin.Context) {
				// 何も設定しない
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  models.ErrorAuthUnauthorized,
		},
		{
			name: "管理者権限なし",
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  models.ErrorAuthForbidden,
		},
		{
			name: "管理者権限あり",
			setupContext: func(c *gin.Context) {
				c.Set("role", models.RoleAdmin)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスを作成
			mockAuthService := new(MockAuthService)

			// ミドルウェアを作成
			authMiddleware := NewAuthMiddleware(mockAuthService)

			// テスト用ルーターを設定
			router := setupTestRouter()
			router.GET("/admin", func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			}, authMiddleware.RequireAdmin(), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin access"})
			})

			// テストリクエストを作成
			req := httptest.NewRequest("GET", "/admin", nil)
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus != http.StatusOK {
				// エラーレスポンスを検証
				var response models.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
}

func TestAuthMiddleware_OptionalAuth(t *testing.T) {
	tests := []struct {
		name         string
		authHeader   string
		mockSetup    func(*MockAuthService)
		expectClaims bool
	}{
		{
			name:         "認証ヘッダーなし",
			authHeader:   "",
			mockSetup:    func(m *MockAuthService) {},
			expectClaims: false,
		},
		{
			name:         "無効な認証ヘッダー形式",
			authHeader:   "InvalidFormat",
			mockSetup:    func(m *MockAuthService) {},
			expectClaims: false,
		},
		{
			name:       "無効なトークン",
			authHeader: "Bearer invalid-token",
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)
			},
			expectClaims: false,
		},
		{
			name:       "有効なトークン",
			authHeader: "Bearer valid-token",
			mockSetup: func(m *MockAuthService) {
				claims := createTestClaims()
				m.On("ValidateToken", "valid-token").Return(claims, nil)
			},
			expectClaims: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスを作成
			mockAuthService := new(MockAuthService)
			tt.mockSetup(mockAuthService)

			// ミドルウェアを作成
			authMiddleware := NewAuthMiddleware(mockAuthService)

			// テスト用ルーターを設定
			router := setupTestRouter()
			router.GET("/optional", authMiddleware.OptionalAuth(), func(c *gin.Context) {
				isAuth := IsAuthenticated(c)
				c.JSON(http.StatusOK, gin.H{
					"authenticated": isAuth,
					"user_id":       c.GetInt("user_id"),
				})
			})

			// テストリクエストを作成
			req := httptest.NewRequest("GET", "/optional", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証（常に200であるべき）
			assert.Equal(t, http.StatusOK, w.Code)

			// レスポンスを検証
			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectClaims, response["authenticated"])

			// モックの呼び出しを検証
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthMiddleware_RequireRole(t *testing.T) {
	tests := []struct {
		name           string
		requiredRole   string
		userRole       string
		setupContext   func(*gin.Context, string)
		expectedStatus int
		expectedError  string
	}{
		{
			name:         "認証情報なし",
			requiredRole: models.RoleAdmin,
			setupContext: func(c *gin.Context, role string) {
				// 何も設定しない
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  models.ErrorAuthUnauthorized,
		},
		{
			name:         "ロール不一致",
			requiredRole: models.RoleAdmin,
			userRole:     "user",
			setupContext: func(c *gin.Context, role string) {
				c.Set("role", role)
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  models.ErrorAuthForbidden,
		},
		{
			name:         "ロール一致",
			requiredRole: models.RoleAdmin,
			userRole:     models.RoleAdmin,
			setupContext: func(c *gin.Context, role string) {
				c.Set("role", role)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスを作成
			mockAuthService := new(MockAuthService)

			// ミドルウェアを作成
			authMiddleware := NewAuthMiddleware(mockAuthService)

			// テスト用ルーターを設定
			router := setupTestRouter()
			router.GET("/role", func(c *gin.Context) {
				tt.setupContext(c, tt.userRole)
				c.Next()
			}, authMiddleware.RequireRole(tt.requiredRole), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "role access"})
			})

			// テストリクエストを作成
			req := httptest.NewRequest("GET", "/role", nil)
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus != http.StatusOK {
				// エラーレスポンスを検証
				var response models.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
}

func TestAuthMiddleware_HelperFunctions(t *testing.T) {
	// テスト用のコンテキストを作成
	router := setupTestRouter()
	router.GET("/test", func(c *gin.Context) {
		// テストデータを設定
		claims := createTestClaims()
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		// ヘルパー関数をテスト
		userID, hasUserID := GetUserID(c)
		username, hasUsername := GetUsername(c)
		role, hasRole := GetUserRole(c)
		jwtClaims, hasClaims := GetClaims(c)
		isAuth := IsAuthenticated(c)
		isAdminUser := IsAdmin(c)

		c.JSON(http.StatusOK, gin.H{
			"user_id":      userID,
			"has_user_id":  hasUserID,
			"username":     username,
			"has_username": hasUsername,
			"role":         role,
			"has_role":     hasRole,
			"has_claims":   hasClaims,
			"claims_valid": jwtClaims != nil,
			"is_auth":      isAuth,
			"is_admin":     isAdminUser,
		})
	})

	// テストリクエストを実行
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// レスポンスを検証
	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// 各ヘルパー関数の結果を検証
	assert.Equal(t, float64(1), response["user_id"])
	assert.True(t, response["has_user_id"].(bool))
	assert.Equal(t, "admin", response["username"])
	assert.True(t, response["has_username"].(bool))
	assert.Equal(t, models.RoleAdmin, response["role"])
	assert.True(t, response["has_role"].(bool))
	assert.True(t, response["has_claims"].(bool))
	assert.True(t, response["claims_valid"].(bool))
	assert.True(t, response["is_auth"].(bool))
	assert.True(t, response["is_admin"].(bool))
}

func TestAuthMiddleware_RequireUser(t *testing.T) {
	tests := []struct {
		name           string
		targetUserID   int
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedError  string
	}{
		{
			name:         "認証情報なし",
			targetUserID: 1,
			setupContext: func(c *gin.Context) {
				// 何も設定しない
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  models.ErrorAuthUnauthorized,
		},
		{
			name:         "ロール情報なし",
			targetUserID: 1,
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 1)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  models.ErrorAuthUnauthorized,
		},
		{
			name:         "管理者アクセス",
			targetUserID: 2,
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("role", models.RoleAdmin)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "本人アクセス",
			targetUserID: 1,
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("role", "user")
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:         "他人アクセス（権限なし）",
			targetUserID: 2,
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("role", "user")
			},
			expectedStatus: http.StatusForbidden,
			expectedError:  models.ErrorAuthForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスを作成
			mockAuthService := new(MockAuthService)

			// ミドルウェアを作成
			authMiddleware := NewAuthMiddleware(mockAuthService)

			// テスト用ルーターを設定
			router := setupTestRouter()
			router.GET("/user", func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			}, authMiddleware.RequireUser(tt.targetUserID), func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "user access"})
			})

			// テストリクエストを作成
			req := httptest.NewRequest("GET", "/user", nil)
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// ステータスコードを検証
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus != http.StatusOK {
				// エラーレスポンスを検証
				var response models.APIResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.False(t, response.Success)
				assert.Equal(t, tt.expectedError, response.Error)
			}
		})
	}
}
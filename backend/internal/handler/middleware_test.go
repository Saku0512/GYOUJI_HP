package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService はAuthServiceのモック
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Login(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

func (m *MockAuthService) ValidateToken(tokenString string) (*service.Claims, error) {
	args := m.Called(tokenString)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.Claims), args.Error(1)
}

func (m *MockAuthService) GenerateToken(userID int, username string) (string, error) {
	args := m.Called(userID, username)
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

func TestAuthMiddleware_RequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "認証ヘッダーなし",
			authHeader:     "",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "認証トークンが必要です",
		},
		{
			name:           "無効な認証ヘッダー形式",
			authHeader:     "InvalidFormat",
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "無効な認証トークン形式です",
		},
		{
			name:       "無効なトークン",
			authHeader: "Bearer invalid-token",
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "無効または期限切れのトークンです",
		},
		{
			name:       "有効なトークン",
			authHeader: "Bearer valid-token",
			mockSetup: func(m *MockAuthService) {
				claims := &service.Claims{
					UserID:   1,
					Username: "admin",
					Role:     "admin",
				}
				m.On("ValidateToken", "valid-token").Return(claims, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスをセットアップ
			mockAuthService := new(MockAuthService)
			tt.mockSetup(mockAuthService)

			// ミドルウェアを作成
			middleware := NewAuthMiddleware(mockAuthService)

			// Ginルーターをセットアップ
			router := gin.New()
			router.Use(middleware.RequireAuth())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// テストリクエストを作成
			req, _ := http.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// アサーション
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}

			// モックの検証
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthMiddleware_RequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "認証情報なし",
			setupContext: func(c *gin.Context) {
				// 何も設定しない
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "認証情報が見つかりません",
		},
		{
			name: "管理者権限なし",
			setupContext: func(c *gin.Context) {
				c.Set("role", "user")
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "管理者権限が必要です",
		},
		{
			name: "管理者権限あり",
			setupContext: func(c *gin.Context) {
				c.Set("role", "admin")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスを作成
			mockAuthService := new(MockAuthService)
			middleware := NewAuthMiddleware(mockAuthService)

			// Ginルーターをセットアップ
			router := gin.New()
			router.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			router.Use(middleware.RequireAdmin())
			router.GET("/admin", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "admin access"})
			})

			// テストリクエストを作成
			req, _ := http.NewRequest("GET", "/admin", nil)
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// アサーション
			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestCORSMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "OPTIONS リクエスト",
			method:         "OPTIONS",
			expectedStatus: http.StatusNoContent,
			checkHeaders:   true,
		},
		{
			name:           "GET リクエスト",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ginルーターをセットアップ
			router := gin.New()
			router.Use(CORSMiddleware())
			router.GET("/test", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "success"})
			})

			// テストリクエストを作成
			req, _ := http.NewRequest(tt.method, "/test", nil)
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// アサーション
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders {
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Headers"), "Authorization")
				assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "GET")
			}
		})
	}
}

func TestErrorHandlerMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Ginルーターをセットアップ
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	router.GET("/panic", func(c *gin.Context) {
		panic("テストパニック")
	})

	// テストリクエストを作成
	req, _ := http.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()

	// リクエストを実行
	router.ServeHTTP(w, req)

	// アサーション
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "サーバー内部エラーが発生しました")
}
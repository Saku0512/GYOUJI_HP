package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "成功ケース",
			requestBody: LoginRequest{
				Username: "admin",
				Password: "password123",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("Login", "admin", "password123").Return("valid-jwt-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "ログインに成功しました",
		},
		{
			name: "無効なリクエスト形式",
			requestBody: map[string]interface{}{
				"invalid": "request",
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "無効なリクエスト形式です",
		},
		{
			name: "ユーザー名が空",
			requestBody: map[string]interface{}{
				"username": "",
				"password": "password123",
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "無効なリクエスト形式です",
		},
		{
			name: "パスワードが空",
			requestBody: map[string]interface{}{
				"username": "admin",
				"password": "",
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "無効なリクエスト形式です",
		},
		{
			name: "認証失敗",
			requestBody: LoginRequest{
				Username: "admin",
				Password: "wrongpassword",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("Login", "admin", "wrongpassword").Return("", assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "認証に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスをセットアップ
			mockAuthService := new(MockAuthService)
			tt.mockSetup(mockAuthService)

			// ハンドラーを作成
			handler := NewAuthHandler(mockAuthService)

			// Ginルーターをセットアップ
			router := gin.New()
			router.POST("/login", handler.Login)

			// リクエストボディを作成
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// アサーション
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			// モックの検証
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "成功ケース",
			requestBody: RefreshTokenRequest{
				Token: "valid-token",
			},
			mockSetup: func(m *MockAuthService) {
				claims := &service.Claims{
					UserID:   1,
					Username: "admin",
					Role:     "admin",
				}
				m.On("ValidateToken", "valid-token").Return(claims, nil)
				m.On("GenerateToken", 1, "admin").Return("new-jwt-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "トークンのリフレッシュに成功しました",
		},
		{
			name: "無効なリクエスト形式",
			requestBody: map[string]interface{}{
				"invalid": "request",
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "無効なリクエスト形式です",
		},
		{
			name: "トークンが空",
			requestBody: map[string]interface{}{
				"token": "",
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "無効なリクエスト形式です",
		},
		{
			name: "無効なトークン",
			requestBody: RefreshTokenRequest{
				Token: "invalid-token",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "無効または期限切れのトークンです",
		},
		{
			name: "新しいトークン生成失敗",
			requestBody: RefreshTokenRequest{
				Token: "valid-token",
			},
			mockSetup: func(m *MockAuthService) {
				claims := &service.Claims{
					UserID:   1,
					Username: "admin",
					Role:     "admin",
				}
				m.On("ValidateToken", "valid-token").Return(claims, nil)
				m.On("GenerateToken", 1, "admin").Return("", assert.AnError)
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "新しいトークンの生成に失敗しました",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスをセットアップ
			mockAuthService := new(MockAuthService)
			tt.mockSetup(mockAuthService)

			// ハンドラーを作成
			handler := NewAuthHandler(mockAuthService)

			// Ginルーターをセットアップ
			router := gin.New()
			router.POST("/refresh", handler.RefreshToken)

			// リクエストボディを作成
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/refresh", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// アサーション
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			// モックの検証
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// モックサービスを作成
	mockAuthService := new(MockAuthService)

	// ハンドラーを作成
	handler := NewAuthHandler(mockAuthService)

	// Ginルーターをセットアップ
	router := gin.New()
	router.POST("/logout", handler.Logout)

	// テストリクエストを作成
	req, _ := http.NewRequest("POST", "/logout", nil)
	w := httptest.NewRecorder()

	// リクエストを実行
	router.ServeHTTP(w, req)

	// アサーション
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "ログアウトしました")
}

func TestAuthHandler_GetProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		setupContext   func(*gin.Context)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "成功ケース",
			setupContext: func(c *gin.Context) {
				c.Set("user_id", 1)
				c.Set("username", "admin")
				c.Set("role", "admin")
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "ユーザー情報を取得しました",
		},
		{
			name: "認証情報なし",
			setupContext: func(c *gin.Context) {
				// 何も設定しない
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "認証情報が見つかりません",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスを作成
			mockAuthService := new(MockAuthService)

			// ハンドラーを作成
			handler := NewAuthHandler(mockAuthService)

			// Ginルーターをセットアップ
			router := gin.New()
			router.Use(func(c *gin.Context) {
				tt.setupContext(c)
				c.Next()
			})
			router.GET("/profile", handler.GetProfile)

			// テストリクエストを作成
			req, _ := http.NewRequest("GET", "/profile", nil)
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// アサーション
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)
		})
	}
}

func TestAuthHandler_ValidateToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockAuthService)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "成功ケース",
			requestBody: RefreshTokenRequest{
				Token: "valid-token",
			},
			mockSetup: func(m *MockAuthService) {
				claims := &service.Claims{
					UserID:   1,
					Username: "admin",
					Role:     "admin",
					RegisteredClaims: jwt.RegisteredClaims{
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
					},
				}
				m.On("ValidateToken", "valid-token").Return(claims, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "トークンは有効です",
		},
		{
			name: "無効なリクエスト形式",
			requestBody: map[string]interface{}{
				"invalid": "request",
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "無効なリクエスト形式です",
		},
		{
			name: "トークンが空",
			requestBody: map[string]interface{}{
				"token": "",
			},
			mockSetup:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "無効なリクエスト形式です",
		},
		{
			name: "無効なトークン",
			requestBody: RefreshTokenRequest{
				Token: "invalid-token",
			},
			mockSetup: func(m *MockAuthService) {
				m.On("ValidateToken", "invalid-token").Return(nil, assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "無効または期限切れのトークンです",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックサービスをセットアップ
			mockAuthService := new(MockAuthService)
			tt.mockSetup(mockAuthService)

			// ハンドラーを作成
			handler := NewAuthHandler(mockAuthService)

			// Ginルーターをセットアップ
			router := gin.New()
			router.POST("/validate", handler.ValidateToken)

			// リクエストボディを作成
			requestBody, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/validate", bytes.NewBuffer(requestBody))
			req.Header.Set("Content-Type", "application/json")

			// レスポンスレコーダーを作成
			w := httptest.NewRecorder()

			// リクエストを実行
			router.ServeHTTP(w, req)

			// アサーション
			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.expectedBody)

			// モックの検証
			mockAuthService.AssertExpectations(t)
		})
	}
}
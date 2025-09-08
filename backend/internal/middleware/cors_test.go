package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGetDefaultCORSConfig(t *testing.T) {
	config := GetDefaultCORSConfig()
	
	assert.NotNil(t, config)
	assert.Contains(t, config.AllowOrigins, "http://localhost:3000")
	assert.Contains(t, config.AllowOrigins, "http://localhost:5173")
	assert.Contains(t, config.AllowMethods, http.MethodGet)
	assert.Contains(t, config.AllowMethods, http.MethodPost)
	assert.Contains(t, config.AllowHeaders, "Authorization")
	assert.Contains(t, config.ExposeHeaders, "X-Request-ID")
	assert.True(t, config.AllowCredentials)
}

func TestGetProductionCORSConfig(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("CORS_ALLOWED_ORIGINS")
	
	config := GetProductionCORSConfig()
	
	assert.NotNil(t, config)
	assert.Contains(t, config.AllowOrigins, "https://tournament.example.com")
	assert.NotContains(t, config.AllowOrigins, "http://localhost:3000")
}

func TestGetProductionCORSConfigWithEnv(t *testing.T) {
	// 環境変数を設定
	os.Setenv("CORS_ALLOWED_ORIGINS", "https://example.com,https://api.example.com")
	defer os.Unsetenv("CORS_ALLOWED_ORIGINS")
	
	config := GetProductionCORSConfig()
	
	assert.NotNil(t, config)
	assert.Contains(t, config.AllowOrigins, "https://example.com")
	assert.Contains(t, config.AllowOrigins, "https://api.example.com")
}

func TestGetDevelopmentCORSConfig(t *testing.T) {
	config := GetDevelopmentCORSConfig()
	
	assert.NotNil(t, config)
	assert.Contains(t, config.AllowOrigins, "http://localhost:3000")
	assert.Contains(t, config.AllowOrigins, "http://localhost:4000")
	assert.Contains(t, config.AllowOrigins, "http://localhost:5000")
}

func TestNewCORSMiddleware(t *testing.T) {
	// 開発環境でテスト
	os.Setenv("GIN_MODE", "debug")
	defer os.Unsetenv("GIN_MODE")
	
	middleware := NewCORSMiddleware()
	assert.NotNil(t, middleware)
	
	// テスト用ルーターを設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// テストリクエストを作成
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	
	// リクエストを実行
	router.ServeHTTP(w, req)
	
	// CORSヘッダーを確認
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	// 開発環境でテスト
	os.Setenv("GIN_MODE", "debug")
	defer os.Unsetenv("GIN_MODE")
	
	middleware := SecurityHeadersMiddleware()
	assert.NotNil(t, middleware)
	
	// テスト用ルーターを設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// テストリクエストを作成
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// リクエストを実行
	router.ServeHTTP(w, req)
	
	// セキュリティヘッダーを確認
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.Equal(t, "geolocation=(), microphone=(), camera=()", w.Header().Get("Permissions-Policy"))
}

func TestSecurityHeadersMiddleware_Production(t *testing.T) {
	// 本番環境でテスト
	os.Setenv("GIN_MODE", "release")
	defer os.Unsetenv("GIN_MODE")
	
	middleware := SecurityHeadersMiddleware()
	assert.NotNil(t, middleware)
	
	// テスト用ルーターを設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "test"})
	})
	
	// テストリクエストを作成
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	
	// リクエストを実行
	router.ServeHTTP(w, req)
	
	// 本番環境固有のセキュリティヘッダーを確認
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Header().Get("Strict-Transport-Security"), "max-age=31536000")
	assert.Contains(t, w.Header().Get("Expect-CT"), "max-age=86400")
}

func TestPreflightMiddleware(t *testing.T) {
	middleware := PreflightMiddleware()
	assert.NotNil(t, middleware)
	
	// テスト用ルーターを設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.OPTIONS("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "get request"})
	})
	
	// OPTIONSリクエストのテスト
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Empty(t, w.Body.String())
	
	// 通常のGETリクエストのテスト
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "get request")
}

func TestRequestIDMiddleware(t *testing.T) {
	middleware := RequestIDMiddleware()
	assert.NotNil(t, middleware)
	
	// テスト用ルーターを設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		requestID, exists := c.Get("request_id")
		assert.True(t, exists)
		assert.NotEmpty(t, requestID)
		c.JSON(http.StatusOK, gin.H{"request_id": requestID})
	})
	
	// リクエストIDなしのテスト
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
	
	// 既存のリクエストIDありのテスト
	req = httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "existing-request-id")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "existing-request-id", w.Header().Get("X-Request-ID"))
}

func TestCombinedMiddleware(t *testing.T) {
	middlewares := CombinedMiddleware()
	assert.NotEmpty(t, middlewares)
	assert.Len(t, middlewares, 4) // RequestID, Security, CORS, Preflight
	
	// テスト用ルーターを設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	for _, middleware := range middlewares {
		router.Use(middleware)
	}
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "combined test"})
	})
	
	// テストリクエストを作成
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()
	
	// リクエストを実行
	router.ServeHTTP(w, req)
	
	// 各ミドルウェアの効果を確認
	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, w.Header().Get("X-Request-ID"))           // RequestID
	assert.NotEmpty(t, w.Header().Get("X-Content-Type-Options")) // Security
	assert.NotEmpty(t, w.Header().Get("Access-Control-Allow-Origin")) // CORS
}

func TestLegacyCORSMiddleware(t *testing.T) {
	middleware := LegacyCORSMiddleware()
	assert.NotNil(t, middleware)
	
	// テスト用ルーターを設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(middleware)
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "legacy test"})
	})
	router.OPTIONS("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "should not reach here"})
	})
	
	// 通常のリクエストのテスト
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
	
	// OPTIONSリクエストのテスト
	req = httptest.NewRequest("OPTIONS", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestGetContentSecurityPolicy(t *testing.T) {
	// 開発環境でのテスト
	os.Setenv("GIN_MODE", "debug")
	defer os.Unsetenv("GIN_MODE")
	
	csp := getContentSecurityPolicy()
	assert.NotEmpty(t, csp)
	assert.Contains(t, csp, "default-src 'self'")
	assert.Contains(t, csp, "'unsafe-inline'") // 開発環境では許可
	
	// 本番環境でのテスト
	os.Setenv("GIN_MODE", "release")
	csp = getContentSecurityPolicy()
	assert.NotEmpty(t, csp)
	assert.Contains(t, csp, "default-src 'self'")
	assert.NotContains(t, csp, "'unsafe-eval'") // 本番環境では禁止
}

func TestIsProduction(t *testing.T) {
	// デフォルト（開発環境）
	os.Unsetenv("GIN_MODE")
	assert.False(t, isProduction())
	
	// 開発環境
	os.Setenv("GIN_MODE", "debug")
	assert.False(t, isProduction())
	
	// 本番環境（release）
	os.Setenv("GIN_MODE", "release")
	assert.True(t, isProduction())
	
	// 本番環境（production）
	os.Setenv("GIN_MODE", "production")
	assert.True(t, isProduction())
	
	// クリーンアップ
	os.Unsetenv("GIN_MODE")
}
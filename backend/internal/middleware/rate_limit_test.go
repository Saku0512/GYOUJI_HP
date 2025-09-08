package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimiter_EndpointLimit(t *testing.T) {
	// テスト用の設定
	config := &RateLimitConfig{
		EndpointLimits: map[string]*EndpointLimit{
			"/api/v1/test": {
				RequestsPerSecond: 2,
				BurstSize:         3,
				WindowDuration:    time.Minute,
			},
		},
	}

	rateLimiter := NewRateLimiter(config)
	defer rateLimiter.Stop()

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.RateLimitMiddleware())
	router.GET("/api/v1/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 最初の3回のリクエストは成功するはず（バーストサイズ）
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/api/v1/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "リクエスト %d は成功するべき", i+1)
	}

	// 4回目のリクエストは制限に引っかかるはず
	req := httptest.NewRequest("GET", "/api/v1/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "4回目のリクエストは制限されるべき")

	// レスポンスの内容をチェック
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "RATE_LIMIT_ENDPOINT_EXCEEDED", response.Error)
	assert.Equal(t, http.StatusTooManyRequests, response.Code)

	// ヘッダーのチェック
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-Reset"))
	assert.Equal(t, "60", w.Header().Get("Retry-After"))
}

func TestRateLimiter_IPLimit(t *testing.T) {
	// テスト用の設定
	config := &RateLimitConfig{
		IPLimits: &IPLimitConfig{
			RequestsPerMinute: 2,
			BurstSize:         3,
			WindowDuration:    time.Minute,
		},
	}

	rateLimiter := NewRateLimiter(config)
	defer rateLimiter.Stop()

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 最初の3回のリクエストは成功するはず
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "リクエスト %d は成功するべき", i+1)
	}

	// 4回目のリクエストは制限に引っかかるはず
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "4回目のリクエストは制限されるべき")

	// レスポンスの内容をチェック
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "RATE_LIMIT_IP_EXCEEDED", response.Error)

	// IPヘッダーのチェック
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-IP-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-IP-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-IP-Reset"))
}

func TestRateLimiter_UserLimit(t *testing.T) {
	// テスト用の設定
	config := &RateLimitConfig{
		UserLimits: &UserLimitConfig{
			RequestsPerMinute: 2,
			BurstSize:         3,
			WindowDuration:    time.Minute,
		},
	}

	rateLimiter := NewRateLimiter(config)
	defer rateLimiter.Stop()

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(func(c *gin.Context) {
		// テスト用にユーザーIDを設定
		c.Set("user_id", 123)
		c.Next()
	})
	router.Use(rateLimiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 最初の3回のリクエストは成功するはず
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "リクエスト %d は成功するべき", i+1)
	}

	// 4回目のリクエストは制限に引っかかるはず
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusTooManyRequests, w.Code, "4回目のリクエストは制限されるべき")

	// レスポンスの内容をチェック
	var response models.APIResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.False(t, response.Success)
	assert.Equal(t, "RATE_LIMIT_USER_EXCEEDED", response.Error)

	// ユーザーヘッダーのチェック
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-User-Limit"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-User-Remaining"))
	assert.NotEmpty(t, w.Header().Get("X-RateLimit-User-Reset"))
}

func TestRateLimiter_ExcludedIP(t *testing.T) {
	// テスト用の設定（127.0.0.1を除外）
	config := &RateLimitConfig{
		EndpointLimits: map[string]*EndpointLimit{
			"/test": {
				RequestsPerSecond: 1,
				BurstSize:         1,
				WindowDuration:    time.Minute,
			},
		},
		ExcludedIPs: []string{"127.0.0.1"},
	}

	rateLimiter := NewRateLimiter(config)
	defer rateLimiter.Stop()

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 除外IPからの複数リクエストは制限されないはず
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "127.0.0.1:12345"
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code, "除外IPからのリクエスト %d は成功するべき", i+1)
	}
}

func TestRateLimiter_XForwardedFor(t *testing.T) {
	// テスト用の設定
	config := &RateLimitConfig{
		IPLimits: &IPLimitConfig{
			RequestsPerMinute: 1,
			BurstSize:         1,
			WindowDuration:    time.Minute,
		},
	}

	rateLimiter := NewRateLimiter(config)
	defer rateLimiter.Stop()

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.RateLimitMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// X-Forwarded-Forヘッダーを使用したテスト
	req1 := httptest.NewRequest("GET", "/test", nil)
	req1.Header.Set("X-Forwarded-For", "192.168.1.100")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// 同じIPからの2回目のリクエストは制限されるはず
	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Forwarded-For", "192.168.1.100")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusTooManyRequests, w2.Code)
}

func TestRateLimiter_EndpointPatternMatching(t *testing.T) {
	rateLimiter := NewRateLimiter(&RateLimitConfig{})

	tests := []struct {
		path     string
		pattern  string
		expected bool
	}{
		{"/api/v1/auth/login", "/api/v1/auth/login", true},
		{"/api/v1/auth/logout", "/api/v1/auth/login", false},
		{"/api/v1/auth/login", "/api/v1/auth", true},
		{"/api/v1/auth/login/test", "/api/v1/auth", true},
		{"/api/v1/tournaments", "/api/v1/tournaments", true},
		{"/api/v1/tournaments/volleyball", "/api/v1/tournaments", true},
		{"/api/v1/tournaments/", "/api/v1/tournaments/", true},
		{"/api/v1/tournaments/volleyball", "/api/v1/tournaments/", true},
		{"/different/path", "/api/v1/auth", false},
	}

	for _, tt := range tests {
		result := rateLimiter.matchEndpointPattern(tt.path, tt.pattern)
		assert.Equal(t, tt.expected, result, "パス %s とパターン %s のマッチング結果が期待値と異なる", tt.path, tt.pattern)
	}
}

func TestRateLimiter_GetClientIP(t *testing.T) {
	rateLimiter := NewRateLimiter(&RateLimitConfig{})

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		ip := rateLimiter.getClientIP(c)
		c.JSON(http.StatusOK, gin.H{"ip": ip})
	})

	tests := []struct {
		name           string
		remoteAddr     string
		xForwardedFor  string
		xRealIP        string
		expectedIP     string
	}{
		{
			name:       "RemoteAddr only",
			remoteAddr: "192.168.1.1:12345",
			expectedIP: "192.168.1.1",
		},
		{
			name:          "X-Forwarded-For header",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1, 192.168.1.1",
			expectedIP:    "203.0.113.1",
		},
		{
			name:       "X-Real-IP header",
			remoteAddr: "10.0.0.1:12345",
			xRealIP:    "203.0.113.2",
			expectedIP: "203.0.113.2",
		},
		{
			name:          "X-Forwarded-For takes precedence",
			remoteAddr:    "10.0.0.1:12345",
			xForwardedFor: "203.0.113.1",
			xRealIP:       "203.0.113.2",
			expectedIP:    "203.0.113.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			req.RemoteAddr = tt.remoteAddr
			if tt.xForwardedFor != "" {
				req.Header.Set("X-Forwarded-For", tt.xForwardedFor)
			}
			if tt.xRealIP != "" {
				req.Header.Set("X-Real-IP", tt.xRealIP)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			var response map[string]string
			err := json.Unmarshal(w.Body.Bytes(), &response)
			require.NoError(t, err)

			assert.Equal(t, tt.expectedIP, response["ip"])
		})
	}
}

func TestRateLimiter_GetStats(t *testing.T) {
	config := GetDefaultRateLimitConfig()
	rateLimiter := NewRateLimiter(config)
	defer rateLimiter.Stop()

	// 初期状態の統計情報
	stats := rateLimiter.GetStats()
	assert.Equal(t, 0, stats["ip_limiters_count"])
	assert.Equal(t, 0, stats["user_limiters_count"])
	assert.Equal(t, 0, stats["endpoint_limiters_count"])

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(rateLimiter.RateLimitMiddleware())
	router.GET("/api/v1/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// リクエストを送信してリミッターを作成
	req := httptest.NewRequest("GET", "/api/v1/auth/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// 統計情報を再取得
	stats = rateLimiter.GetStats()
	assert.Equal(t, 1, stats["ip_limiters_count"])
	assert.Equal(t, 1, stats["endpoint_limiters_count"])
}

func TestGetDefaultRateLimitConfig(t *testing.T) {
	config := GetDefaultRateLimitConfig()

	// デフォルト設定の検証
	assert.NotNil(t, config.EndpointLimits)
	assert.NotNil(t, config.IPLimits)
	assert.NotNil(t, config.UserLimits)
	assert.NotEmpty(t, config.ExcludedIPs)

	// 認証エンドポイントの制限設定
	authLimit, exists := config.EndpointLimits["/api/v1/auth/login"]
	assert.True(t, exists)
	assert.Equal(t, 5, authLimit.RequestsPerSecond)
	assert.Equal(t, 10, authLimit.BurstSize)

	// 除外IPの確認
	assert.Contains(t, config.ExcludedIPs, "127.0.0.1")
	assert.Contains(t, config.ExcludedIPs, "::1")
}

func TestRateLimiter_UpdateConfig(t *testing.T) {
	// 初期設定
	initialConfig := &RateLimitConfig{
		EndpointLimits: map[string]*EndpointLimit{
			"/test": {
				RequestsPerSecond: 1,
				BurstSize:         1,
				WindowDuration:    time.Minute,
			},
		},
	}

	rateLimiter := NewRateLimiter(initialConfig)
	defer rateLimiter.Stop()

	// 新しい設定
	newConfig := &RateLimitConfig{
		EndpointLimits: map[string]*EndpointLimit{
			"/test": {
				RequestsPerSecond: 10,
				BurstSize:         20,
				WindowDuration:    time.Minute,
			},
		},
	}

	// 設定を更新
	rateLimiter.UpdateConfig(newConfig)

	// 設定が更新されたことを確認
	assert.Equal(t, newConfig, rateLimiter.config)

	// 既存のリミッターがクリアされたことを確認
	stats := rateLimiter.GetStats()
	assert.Equal(t, 0, stats["ip_limiters_count"])
	assert.Equal(t, 0, stats["user_limiters_count"])
	assert.Equal(t, 0, stats["endpoint_limiters_count"])
}
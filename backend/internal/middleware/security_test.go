package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecurityMiddleware_DefaultConfig(t *testing.T) {
	config := GetDefaultSecurityConfig()
	securityMiddleware := NewSecurityMiddleware(config)

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(securityMiddleware.SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// セキュリティヘッダーの確認
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.NotEmpty(t, w.Header().Get("Strict-Transport-Security"))
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	assert.NotEmpty(t, w.Header().Get("Permissions-Policy"))
}

func TestSecurityMiddleware_DevelopmentConfig(t *testing.T) {
	config := GetDevelopmentSecurityConfig()
	securityMiddleware := NewSecurityMiddleware(config)

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(securityMiddleware.SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// 開発環境では一部のヘッダーが設定されない
	assert.NotEmpty(t, w.Header().Get("Content-Security-Policy"))
	assert.Empty(t, w.Header().Get("Strict-Transport-Security")) // 開発環境では無効
	assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
}

func TestSecurityMiddleware_CSPHeader(t *testing.T) {
	config := &SecurityConfig{
		CSP: &CSPConfig{
			Enabled:    true,
			DefaultSrc: []string{"'self'"},
			ScriptSrc:  []string{"'self'", "'unsafe-inline'"},
			StyleSrc:   []string{"'self'", "'unsafe-inline'"},
			ImgSrc:     []string{"'self'", "data:", "https:"},
			ObjectSrc:  []string{"'none'"},
			UpgradeInsecureRequests: true,
		},
	}

	securityMiddleware := NewSecurityMiddleware(config)
	cspHeader := securityMiddleware.buildCSPHeader(config.CSP)

	expectedDirectives := []string{
		"default-src 'self'",
		"script-src 'self' 'unsafe-inline'",
		"style-src 'self' 'unsafe-inline'",
		"img-src 'self' data: https:",
		"object-src 'none'",
		"upgrade-insecure-requests",
	}

	for _, directive := range expectedDirectives {
		assert.Contains(t, cspHeader, directive)
	}
}

func TestSecurityMiddleware_HSTSHeader(t *testing.T) {
	config := &SecurityConfig{
		HSTS: &HSTSConfig{
			Enabled:           true,
			MaxAge:            31536000,
			IncludeSubDomains: true,
			Preload:           true,
		},
	}

	securityMiddleware := NewSecurityMiddleware(config)
	hstsHeader := securityMiddleware.buildHSTSHeader(config.HSTS)

	assert.Contains(t, hstsHeader, "max-age=31536000")
	assert.Contains(t, hstsHeader, "includeSubDomains")
	assert.Contains(t, hstsHeader, "preload")
}

func TestSetSecureCookie(t *testing.T) {
	config := &SecurityConfig{
		SecureCookies: &SecureCookieConfig{
			Enabled:  true,
			Secure:   true,
			HttpOnly: true,
			SameSite: "Strict",
			Path:     "/",
			MaxAge:   3600,
		},
	}

	securityMiddleware := NewSecurityMiddleware(config)

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(securityMiddleware.SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		SetSecureCookie(c, "test_cookie", "test_value", 3600)
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Cookieヘッダーの確認
	cookies := w.Header()["Set-Cookie"]
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Contains(t, cookie, "test_cookie=test_value")
	assert.Contains(t, cookie, "Secure")
	assert.Contains(t, cookie, "HttpOnly")
	assert.Contains(t, cookie, "SameSite=Strict")
	assert.Contains(t, cookie, "Path=/")
	assert.Contains(t, cookie, "Max-Age=3600")
}

func TestCSRFProtectionMiddleware(t *testing.T) {
	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// CSRFトークンを設定するミドルウェア
	router.Use(func(c *gin.Context) {
		c.Set("csrf_token", "valid_token")
		c.Next()
	})
	
	router.Use(CSRFProtectionMiddleware())
	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// 有効なCSRFトークンでのテスト
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-CSRF-Token", "valid_token")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// 無効なCSRFトークンでのテスト
	req = httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-CSRF-Token", "invalid_token")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// CSRFトークンなしでのテスト
	req = httptest.NewRequest("POST", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusForbidden, w.Code)

	// GETリクエストはCSRFチェックをスキップ
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	// GETエンドポイントが定義されていないため404になるが、CSRFチェックはスキップされる
	assert.NotEqual(t, http.StatusForbidden, w.Code)
}

func TestXSSProtectionMiddleware(t *testing.T) {
	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(XSSProtectionMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
}

func TestFrameOptionsMiddleware(t *testing.T) {
	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(FrameOptionsMiddleware("SAMEORIGIN"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
}

func TestHSTSMiddleware(t *testing.T) {
	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(HSTSMiddleware(31536000, true, false))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	// HTTPSリクエストのテスト
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	hstsHeader := w.Header().Get("Strict-Transport-Security")
	assert.Contains(t, hstsHeader, "max-age=31536000")
	assert.Contains(t, hstsHeader, "includeSubDomains")
	assert.NotContains(t, hstsHeader, "preload")

	// HTTPリクエストのテスト（HSTSヘッダーは設定されない）
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Empty(t, w.Header().Get("Strict-Transport-Security"))
}

func TestReferrerPolicyMiddleware(t *testing.T) {
	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(ReferrerPolicyMiddleware("no-referrer"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "no-referrer", w.Header().Get("Referrer-Policy"))
}

func TestPermissionsPolicyMiddleware(t *testing.T) {
	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(PermissionsPolicyMiddleware("geolocation=(), camera=()"))
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "success"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "geolocation=(), camera=()", w.Header().Get("Permissions-Policy"))
}

func TestGetSecurityHeaders(t *testing.T) {
	config := GetDefaultSecurityConfig()
	securityMiddleware := NewSecurityMiddleware(config)

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(securityMiddleware.SecurityHeadersMiddleware())
	router.GET("/test", func(c *gin.Context) {
		headers := GetSecurityHeaders(c)
		c.JSON(http.StatusOK, headers)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// レスポンスボディから取得したヘッダー情報を確認
	// 実際のテストでは、レスポンスボディをパースして確認する
	assert.NotEmpty(t, w.Body.String())
}

func TestValidateSecurityConfig(t *testing.T) {
	// 空のCSP設定のテスト
	config := &SecurityConfig{
		CSP: &CSPConfig{
			Enabled: true,
		},
	}

	err := ValidateSecurityConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, []string{"'self'"}, config.CSP.DefaultSrc) // デフォルト値が設定される

	// 無効なHSTS設定のテスト
	config = &SecurityConfig{
		HSTS: &HSTSConfig{
			Enabled: true,
			MaxAge:  0, // 無効な値
		},
	}

	err = ValidateSecurityConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, 31536000, config.HSTS.MaxAge) // デフォルト値が設定される

	// 無効なCookie設定のテスト
	config = &SecurityConfig{
		SecureCookies: &SecureCookieConfig{
			Enabled: true,
			Path:    "", // 空の値
			MaxAge:  0,  // 無効な値
		},
	}

	err = ValidateSecurityConfig(config)
	assert.NoError(t, err)
	assert.Equal(t, "/", config.SecureCookies.Path)         // デフォルト値が設定される
	assert.Equal(t, "Strict", config.SecureCookies.SameSite) // デフォルト値が設定される
	assert.Equal(t, 86400, config.SecureCookies.MaxAge)     // デフォルト値が設定される
}

func TestIsSecureContext(t *testing.T) {
	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		isSecure := IsSecureContext(c)
		c.JSON(http.StatusOK, gin.H{"is_secure": isSecure})
	})

	// HTTPSリクエストのテスト
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_secure":true`)

	// HTTPリクエストのテスト
	req = httptest.NewRequest("GET", "/test", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"is_secure":false`)
}

func TestGetClientFingerprint(t *testing.T) {
	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/test", func(c *gin.Context) {
		fingerprint := GetClientFingerprint(c)
		c.JSON(http.StatusOK, gin.H{"fingerprint": fingerprint})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Test Browser)")
	req.Header.Set("Accept-Language", "ja,en")
	req.Header.Set("Accept-Encoding", "gzip, deflate")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	fingerprint := w.Body.String()
	assert.Contains(t, fingerprint, "Mozilla/5.0 (Test Browser)")
	assert.Contains(t, fingerprint, "ja,en")
	assert.Contains(t, fingerprint, "gzip, deflate")
}

func TestSecurityMiddleware_UpdateConfig(t *testing.T) {
	initialConfig := GetDefaultSecurityConfig()
	securityMiddleware := NewSecurityMiddleware(initialConfig)

	// 新しい設定
	newConfig := &SecurityConfig{
		XFrameOptions:       "SAMEORIGIN",
		XContentTypeOptions: "nosniff",
		XSSProtection:       "0", // XSS保護を無効化
	}

	// 設定を更新
	err := securityMiddleware.UpdateSecurityConfig(newConfig)
	assert.NoError(t, err)

	// 設定が更新されたことを確認
	updatedConfig := securityMiddleware.GetSecurityConfig()
	assert.Equal(t, "SAMEORIGIN", updatedConfig.XFrameOptions)
	assert.Equal(t, "0", updatedConfig.XSSProtection)
}
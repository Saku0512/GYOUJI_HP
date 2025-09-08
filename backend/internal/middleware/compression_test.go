package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGzipMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("gzip圧縮が適用される", func(t *testing.T) {
		router := gin.New()
		router.Use(GzipMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test response with long content to trigger compression"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "gzip", w.Header().Get("Content-Encoding"))
		assert.Equal(t, "Accept-Encoding", w.Header().Get("Vary"))

		// レスポンスが圧縮されているかチェック
		reader, err := gzip.NewReader(w.Body)
		assert.NoError(t, err)
		defer reader.Close()

		decompressed, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Contains(t, string(decompressed), "test response")
	})

	t.Run("Accept-Encodingがない場合は圧縮されない", func(t *testing.T) {
		router := gin.New()
		router.Use(GzipMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
	})

	t.Run("除外パスは圧縮されない", func(t *testing.T) {
		config := CompressionConfig{
			Level:         gzip.DefaultCompression,
			MinLength:     1024,
			ExcludedPaths: []string{"/health"},
		}

		router := gin.New()
		router.Use(GzipMiddleware(config))
		
		router.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		req := httptest.NewRequest("GET", "/health", nil)
		req.Header.Set("Accept-Encoding", "gzip")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Empty(t, w.Header().Get("Content-Encoding"))
	})
}

func TestCacheMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("静的リソースにキャッシュヘッダーが設定される", func(t *testing.T) {
		router := gin.New()
		router.Use(CacheMiddleware())
		
		router.GET("/static/test.css", func(c *gin.Context) {
			c.String(200, "body { color: red; }")
		})

		req := httptest.NewRequest("GET", "/static/test.css", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		cacheControl := w.Header().Get("Cache-Control")
		assert.Contains(t, cacheControl, "public")
		assert.Contains(t, cacheControl, "max-age=86400") // 24時間
	})

	t.Run("APIレスポンスに適切なキャッシュヘッダーが設定される", func(t *testing.T) {
		router := gin.New()
		router.Use(CacheMiddleware())
		
		router.GET("/api/v1/tournaments", func(c *gin.Context) {
			c.JSON(200, gin.H{"data": []string{}})
		})

		req := httptest.NewRequest("GET", "/api/v1/tournaments", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		cacheControl := w.Header().Get("Cache-Control")
		assert.Contains(t, cacheControl, "public")
		assert.Contains(t, cacheControl, "max-age=300") // 5分
	})

	t.Run("キャッシュ無効パスにno-cacheが設定される", func(t *testing.T) {
		router := gin.New()
		router.Use(CacheMiddleware())
		
		router.POST("/api/v1/auth/login", func(c *gin.Context) {
			c.JSON(200, gin.H{"token": "test"})
		})

		req := httptest.NewRequest("POST", "/api/v1/auth/login", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "no-cache, no-store, must-revalidate", w.Header().Get("Cache-Control"))
		assert.Equal(t, "no-cache", w.Header().Get("Pragma"))
		assert.Equal(t, "0", w.Header().Get("Expires"))
	})
}

func TestSecurityHeadersMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("セキュリティヘッダーが設定される", func(t *testing.T) {
		router := gin.New()
		router.Use(SecurityHeadersMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Contains(t, w.Header().Get("Content-Security-Policy"), "default-src 'self'")
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
		assert.Equal(t, "strict-origin-when-cross-origin", w.Header().Get("Referrer-Policy"))
	})

	t.Run("カスタム設定が適用される", func(t *testing.T) {
		config := SecurityHeadersConfig{
			XFrameOptions: "SAMEORIGIN",
			ContentSecurityPolicy: "default-src 'none'",
		}

		router := gin.New()
		router.Use(SecurityHeadersMiddleware(config))
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "SAMEORIGIN", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "default-src 'none'", w.Header().Get("Content-Security-Policy"))
	})
}

func TestResponseOptimizationMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("レスポンス最適化ヘッダーが設定される", func(t *testing.T) {
		router := gin.New()
		router.Use(ResponseOptimizationMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
		assert.NotEmpty(t, w.Header().Get("X-Response-Time"))
		assert.Equal(t, "nosniff", w.Header().Get("X-Content-Type-Options"))
		assert.Equal(t, "DENY", w.Header().Get("X-Frame-Options"))
		assert.Equal(t, "1; mode=block", w.Header().Get("X-XSS-Protection"))
	})

	t.Run("既存のリクエストIDが保持される", func(t *testing.T) {
		router := gin.New()
		router.Use(ResponseOptimizationMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("X-Request-ID", "existing-id")
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		assert.Equal(t, "existing-id", w.Header().Get("X-Request-ID"))
	})
}

func TestConditionalGetMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("ETagが一致する場合304を返す", func(t *testing.T) {
		router := gin.New()
		router.Use(ConditionalGetMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			c.Header("ETag", `"test-etag"`)
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("If-None-Match", `"test-etag"`)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 304, w.Code)
	})

	t.Run("Last-Modifiedが古い場合304を返す", func(t *testing.T) {
		router := gin.New()
		router.Use(ConditionalGetMiddleware())
		
		lastModified := time.Now().Add(-1 * time.Hour)
		
		router.GET("/test", func(c *gin.Context) {
			c.Header("Last-Modified", lastModified.UTC().Format(http.TimeFormat))
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("If-Modified-Since", time.Now().UTC().Format(http.TimeFormat))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 304, w.Code)
	})

	t.Run("条件が一致しない場合は通常のレスポンスを返す", func(t *testing.T) {
		router := gin.New()
		router.Use(ConditionalGetMiddleware())
		
		router.GET("/test", func(c *gin.Context) {
			c.Header("ETag", `"different-etag"`)
			c.JSON(200, gin.H{"message": "test"})
		})

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("If-None-Match", `"test-etag"`)
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, 200, w.Code)
		body := w.Body.String()
		assert.Contains(t, body, "test")
	})
}

func TestGenerateRequestID(t *testing.T) {
	t.Run("リクエストIDが生成される", func(t *testing.T) {
		id1 := generateRequestID()
		id2 := generateRequestID()

		assert.NotEmpty(t, id1)
		assert.NotEmpty(t, id2)
		assert.NotEqual(t, id1, id2)
		assert.True(t, strings.HasPrefix(id1, "req_"))
		assert.True(t, strings.HasPrefix(id2, "req_"))
	})
}
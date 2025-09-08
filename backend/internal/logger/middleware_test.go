package logger

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		existingHeader string
		wantNewID      bool
	}{
		{
			name:           "新しいリクエストIDを生成",
			existingHeader: "",
			wantNewID:      true,
		},
		{
			name:           "既存のリクエストIDを使用",
			existingHeader: "existing-request-id",
			wantNewID:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestIDMiddleware())
			
			var capturedRequestID string
			router.GET("/test", func(c *gin.Context) {
				capturedRequestID = c.Request.Context().Value("request_id").(string)
				c.JSON(200, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.existingHeader != "" {
				req.Header.Set("X-Request-ID", tt.existingHeader)
			}
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// レスポンスヘッダーにリクエストIDが設定されていることを確認
			responseRequestID := w.Header().Get("X-Request-ID")
			assert.NotEmpty(t, responseRequestID)

			// コンテキストにリクエストIDが設定されていることを確認
			assert.NotEmpty(t, capturedRequestID)
			assert.Equal(t, responseRequestID, capturedRequestID)

			if tt.existingHeader != "" {
				// 既存のヘッダーが使用されることを確認
				assert.Equal(t, tt.existingHeader, capturedRequestID)
			} else {
				// 新しいUUIDが生成されることを確認（36文字のUUID形式）
				assert.Len(t, capturedRequestID, 36)
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		path           string
		method         string
		statusCode     int
		expectedStatus int
	}{
		{
			name:           "正常なGETリクエスト",
			path:           "/test",
			method:         "GET",
			statusCode:     200,
			expectedStatus: 200,
		},
		{
			name:           "404エラー",
			path:           "/notfound",
			method:         "GET",
			statusCode:     404,
			expectedStatus: 404,
		},
		{
			name:           "500エラー",
			path:           "/error",
			method:         "POST",
			statusCode:     500,
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestIDMiddleware())
			router.Use(LoggingMiddleware())
			
			router.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"status": "ok"})
			})
			
			router.POST("/error", func(c *gin.Context) {
				c.JSON(500, gin.H{"error": "internal server error"})
			})

			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.Header.Set("User-Agent", "test-agent")
			
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// リクエストIDがレスポンスヘッダーに含まれていることを確認
			requestID := w.Header().Get("X-Request-ID")
			assert.NotEmpty(t, requestID)
		})
	}
}

func TestLoggingMiddlewareWithQuery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(LoggingMiddleware())
	
	router.GET("/search", func(c *gin.Context) {
		c.JSON(200, gin.H{"query": c.Query("q")})
	})

	req := httptest.NewRequest("GET", "/search?q=test&limit=10", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	
	// リクエストIDがレスポンスヘッダーに含まれていることを確認
	requestID := w.Header().Get("X-Request-ID")
	assert.NotEmpty(t, requestID)
}
f
unc TestLoggingMiddlewareWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name   string
		config LoggingConfig
		path   string
		skip   bool
	}{
		{
			name: "通常のパス",
			config: LoggingConfig{
				SkipPaths: []string{"/health"},
			},
			path: "/api/test",
			skip: false,
		},
		{
			name: "スキップパス",
			config: LoggingConfig{
				SkipPaths: []string{"/health", "/metrics"},
			},
			path: "/health",
			skip: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(RequestIDMiddleware())
			router.Use(LoggingMiddlewareWithConfig(tt.config))
			
			var requestIDCaptured string
			router.GET("/*path", func(c *gin.Context) {
				requestIDCaptured = c.Request.Context().Value("request_id").(string)
				c.JSON(200, gin.H{"status": "ok"})
			})

			req := httptest.NewRequest("GET", tt.path, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, 200, w.Code)
			
			if !tt.skip {
				// スキップされない場合はリクエストIDが設定される
				assert.NotEmpty(t, requestIDCaptured)
			}
		})
	}
}

func TestLoggingMiddlewareWithRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := LoggingConfig{
		LogRequestBody: true,
		MaxBodySize:    1024,
	}

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(LoggingMiddlewareWithConfig(config))
	
	router.POST("/api/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	body := `{"test": "data"}`
	req := httptest.NewRequest("POST", "/api/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
}

func TestErrorLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(ErrorLoggingMiddleware())
	
	router.GET("/error", func(c *gin.Context) {
		c.Error(assert.AnError)
		c.JSON(500, gin.H{"error": "test error"})
	})

	req := httptest.NewRequest("GET", "/error", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestPanicRecoveryMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	router.Use(PanicRecoveryMiddleware())
	
	router.GET("/panic", func(c *gin.Context) {
		panic("test panic")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 500, w.Code)
}

func TestResponseWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// responseWriterのテスト
	w := httptest.NewRecorder()
	body := &bytes.Buffer{}
	rw := &responseWriter{
		ResponseWriter: w,
		body:          body,
	}

	testData := []byte("test response")
	n, err := rw.Write(testData)

	assert.NoError(t, err)
	assert.Equal(t, len(testData), n)
	assert.Equal(t, string(testData), body.String())
	assert.Equal(t, string(testData), w.Body.String())
}

func TestLoggerWithContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(RequestIDMiddleware())
	
	var contextLogger Logger
	router.GET("/test", func(c *gin.Context) {
		// コンテキストにユーザー情報を設定
		ctx := context.WithValue(c.Request.Context(), "user_id", 123)
		ctx = context.WithValue(ctx, "username", "testuser")
		ctx = context.WithValue(ctx, "role", "admin")
		c.Request = c.Request.WithContext(ctx)
		
		contextLogger = GetLogger().WithContext(c.Request.Context())
		c.JSON(200, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.NotNil(t, contextLogger)
}
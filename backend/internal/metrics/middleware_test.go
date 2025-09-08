package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestMetricsMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	collector := NewDefaultCollector()
	router := gin.New()
	router.Use(MetricsMiddlewareWithCollector(collector))

	router.GET("/test", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	router.POST("/users", func(c *gin.Context) {
		c.JSON(201, gin.H{"id": 1})
	})

	router.GET("/error", func(c *gin.Context) {
		c.JSON(500, gin.H{"error": "internal error"})
	})

	tests := []struct {
		name       string
		method     string
		path       string
		body       string
		wantStatus int
	}{
		{
			name:       "GET request",
			method:     "GET",
			path:       "/test",
			wantStatus: 200,
		},
		{
			name:       "POST request with body",
			method:     "POST",
			path:       "/users",
			body:       `{"name": "test"}`,
			wantStatus: 201,
		},
		{
			name:       "Error request",
			method:     "GET",
			path:       "/error",
			wantStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var req *http.Request
			if tt.body != "" {
				req = httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tt.method, tt.path, nil)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}

	// メトリクスが記録されていることを確認
	metrics := collector.GetMetrics()
	assert.True(t, len(metrics) > 0)

	// HTTPリクエスト総数メトリクスを確認
	found := false
	for _, metric := range metrics {
		if metric.Name == "http_requests_total" {
			found = true
			assert.True(t, metric.Value >= 3.0) // 3つのリクエストが記録されている
			break
		}
	}
	assert.True(t, found, "http_requests_total メトリクスが見つかりません")
}

func TestMetricsMiddlewareWithConfig(t *testing.T) {
	gin.SetMode(gin.TestMode)

	collector := NewDefaultCollector()
	config := MetricsConfig{
		SkipPaths: []string{"/health", "/metrics"},
		PathNormalizer: func(path string) string {
			// /users/:id を /users/{id} に正規化
			if strings.HasPrefix(path, "/users/") && len(path) > 7 {
				return "/users/{id}"
			}
			return path
		},
	}

	router := gin.New()
	router.Use(MetricsMiddlewareWithConfig(config, collector))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	router.GET("/users/:id", func(c *gin.Context) {
		c.JSON(200, gin.H{"id": c.Param("id")})
	})

	// ヘルスチェックエンドポイント（スキップされる）
	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// ユーザーエンドポイント（正規化される）
	req = httptest.NewRequest("GET", "/users/123", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	req = httptest.NewRequest("GET", "/users/456", nil)
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, 200, w.Code)

	// メトリクスを確認
	metrics := collector.GetMetrics()

	// /health エンドポイントのメトリクスは記録されていないはず
	healthFound := false
	usersFound := false
	for _, metric := range metrics {
		if metric.Name == "http_requests_total" {
			if path, exists := metric.Labels["path"]; exists {
				if path == "/health" {
					healthFound = true
				}
				if path == "/users/{id}" {
					usersFound = true
				}
			}
		}
	}

	assert.False(t, healthFound, "/health エンドポイントのメトリクスはスキップされるべき")
	assert.True(t, usersFound, "/users/{id} エンドポイントのメトリクスが正規化されて記録されるべき")
}

func TestDatabaseMetricsWrapper(t *testing.T) {
	collector := NewDefaultCollector()
	wrapper := NewDatabaseMetricsWrapper(collector)

	// 成功するクエリ
	err := wrapper.WrapQuery("SELECT", func() error {
		time.Sleep(10 * time.Millisecond)
		return nil
	})
	assert.NoError(t, err)

	// 失敗するクエリ
	err = wrapper.WrapQuery("INSERT", func() error {
		time.Sleep(5 * time.Millisecond)
		return assert.AnError
	})
	assert.Error(t, err)

	// 結果付きクエリ
	result, err := wrapper.WrapQueryWithResult("SELECT", func() (interface{}, error) {
		time.Sleep(15 * time.Millisecond)
		return "test result", nil
	})
	assert.NoError(t, err)
	assert.Equal(t, "test result", result)

	// メトリクスが記録されていることを確認
	metrics := collector.GetMetrics()
	found := false
	for _, metric := range metrics {
		if metric.Name == "db_queries_total" {
			found = true
			assert.True(t, metric.Value >= 3.0) // 3つのクエリが記録されている
			break
		}
	}
	assert.True(t, found, "db_queries_total メトリクスが見つかりません")
}

func TestWebSocketMetricsWrapper(t *testing.T) {
	collector := NewDefaultCollector()
	wrapper := NewWebSocketMetricsWrapper(collector)

	// 接続
	wrapper.OnConnect()
	wrapper.OnConnect()

	// 切断
	wrapper.OnDisconnect()

	// メトリクスを確認
	metrics := collector.GetMetrics()
	found := false
	for _, metric := range metrics {
		if metric.Name == "websocket_connections" {
			found = true
			assert.Equal(t, 1.0, metric.Value) // 2接続 - 1切断 = 1
			break
		}
	}
	assert.True(t, found, "websocket_connections メトリクスが見つかりません")
}

func TestMetricsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	collector := NewDefaultCollector()
	
	// テストデータを追加
	collector.RecordHTTPRequest("GET", "/test", 200, 100*time.Millisecond, 1024, 2048)
	collector.SetActiveUsers(50)

	router := gin.New()
	router.GET("/metrics", MetricsHandlerWithCollector(collector))

	tests := []struct {
		name        string
		query       string
		wantType    string
		wantStatus  int
	}{
		{
			name:       "JSON format (default)",
			query:      "",
			wantType:   "application/json",
			wantStatus: 200,
		},
		{
			name:       "JSON format (explicit)",
			query:      "?format=json",
			wantType:   "application/json",
			wantStatus: 200,
		},
		{
			name:       "Prometheus format",
			query:      "?format=prometheus",
			wantType:   "text/plain",
			wantStatus: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/metrics"+tt.query, nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Header().Get("Content-Type"), tt.wantType)
			assert.NotEmpty(t, w.Body.String())
		})
	}
}

func TestFormatPrometheus(t *testing.T) {
	metrics := []Metric{
		{
			Name:      "test_counter",
			Value:     42.0,
			Labels:    Labels{"method": "GET", "status": "200"},
			Timestamp: time.Unix(1640995200, 0), // 2022-01-01 00:00:00 UTC
			Unit:      "counter",
			Help:      "Test counter metric",
		},
		{
			Name:      "test_gauge",
			Value:     100.5,
			Labels:    Labels{},
			Timestamp: time.Unix(1640995200, 0),
			Unit:      "gauge",
			Help:      "Test gauge metric",
		},
	}

	result := formatPrometheus(metrics)

	// 基本的な形式チェック
	assert.Contains(t, result, "# HELP test_counter Test counter metric")
	assert.Contains(t, result, "# TYPE test_counter counter")
	assert.Contains(t, result, "test_counter{method=\"GET\",status=\"200\"} 42")
	assert.Contains(t, result, "# HELP test_gauge Test gauge metric")
	assert.Contains(t, result, "# TYPE test_gauge gauge")
	assert.Contains(t, result, "test_gauge 100.5")
}

func TestPathNormalizers(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Remove query parameters",
			input:    "/api/users?page=1&limit=10",
			expected: "/api/users",
		},
		{
			name:     "Path without query",
			input:    "/api/users",
			expected: "/api/users",
		},
		{
			name:     "Empty path",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathNormalizers.RemoveQuery(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestMetricsMiddlewareContextValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	collector := NewDefaultCollector()
	router := gin.New()
	router.Use(MetricsMiddlewareWithCollector(collector))

	var capturedDuration time.Duration
	var capturedStatusCode int
	var capturedRequestSize int64
	var capturedResponseSize int64

	router.POST("/test", func(c *gin.Context) {
		// ミドルウェアが設定した値を取得
		if duration, exists := c.Get("metrics_duration"); exists {
			capturedDuration = duration.(time.Duration)
		}
		if statusCode, exists := c.Get("metrics_status_code"); exists {
			capturedStatusCode = statusCode.(int)
		}
		if requestSize, exists := c.Get("metrics_request_size"); exists {
			capturedRequestSize = requestSize.(int64)
		}
		if responseSize, exists := c.Get("metrics_response_size"); exists {
			capturedResponseSize = responseSize.(int64)
		}

		c.JSON(200, gin.H{"status": "ok"})
	})

	body := `{"test": "data"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.True(t, capturedDuration > 0)
	assert.Equal(t, 200, capturedStatusCode)
	assert.Equal(t, int64(len(body)), capturedRequestSize)
	assert.True(t, capturedResponseSize > 0)
}
package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSanitizationMiddleware_HTMLSanitization(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(sanitizationMiddleware.SanitizeInputMiddleware())
	router.POST("/test", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	// XSSを含むJSONデータ
	maliciousData := map[string]interface{}{
		"name":    "<script>alert('XSS')</script>John",
		"email":   "test@example.com",
		"comment": "Hello <script>alert('XSS')</script> World",
	}

	jsonData, _ := json.Marshal(maliciousData)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// スクリプトタグが削除されていることを確認
	assert.NotContains(t, response["name"], "<script>")
	assert.NotContains(t, response["comment"], "<script>")
	assert.Equal(t, "test@example.com", response["email"]) // 正常な値は変更されない
}

func TestSanitizationMiddleware_SQLInjectionProtection(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(sanitizationMiddleware.SanitizeInputMiddleware())
	router.POST("/test", func(c *gin.Context) {
		var data map[string]interface{}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, data)
	})

	// SQLインジェクションを含むデータ
	maliciousData := map[string]interface{}{
		"username": "admin' OR '1'='1",
		"password": "password'; DROP TABLE users; --",
		"search":   "test UNION SELECT * FROM passwords",
	}

	jsonData, _ := json.Marshal(maliciousData)
	req := httptest.NewRequest("POST", "/test", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// SQLインジェクションパターンが削除されていることを確認
	assert.Equal(t, "", response["username"]) // 危険なパターンは空文字列になる
	assert.Equal(t, "", response["password"])
	assert.Equal(t, "", response["search"])
}

func TestSanitizationMiddleware_QueryParameterSanitization(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(sanitizationMiddleware.SanitizeInputMiddleware())
	router.GET("/test", func(c *gin.Context) {
		search := c.Query("search")
		name := c.Query("name")
		c.JSON(http.StatusOK, gin.H{
			"search": search,
			"name":   name,
		})
	})

	// 悪意のあるクエリパラメータ
	req := httptest.NewRequest("GET", "/test?search=<script>alert('XSS')</script>&name=John' OR '1'='1", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// 悪意のあるパターンがサニタイズされていることを確認
	assert.NotContains(t, response["search"], "<script>")
	assert.Equal(t, "", response["name"]) // SQLインジェクションパターンは空文字列
}

func TestSanitizationMiddleware_FormDataSanitization(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(sanitizationMiddleware.SanitizeInputMiddleware())
	router.POST("/test", func(c *gin.Context) {
		name := c.PostForm("name")
		email := c.PostForm("email")
		c.JSON(http.StatusOK, gin.H{
			"name":  name,
			"email": email,
		})
	})

	// フォームデータ
	formData := url.Values{}
	formData.Set("name", "<script>alert('XSS')</script>John")
	formData.Set("email", "test@example.com")

	req := httptest.NewRequest("POST", "/test", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// スクリプトタグが削除されていることを確認
	assert.NotContains(t, response["name"], "<script>")
	assert.Equal(t, "test@example.com", response["email"])
}

func TestSanitizationMiddleware_ExcludedPaths(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	config.ExcludedPaths = []string{"/api/v1/health"}
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	// Ginエンジンの設定
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(sanitizationMiddleware.SanitizeInputMiddleware())
	router.GET("/api/v1/health", func(c *gin.Context) {
		search := c.Query("search")
		c.JSON(http.StatusOK, gin.H{"search": search})
	})

	// 除外パスでは悪意のあるデータもそのまま通る
	req := httptest.NewRequest("GET", "/api/v1/health?search=<script>alert('XSS')</script>", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)

	// 除外パスではサニタイゼーションが行われない
	assert.Contains(t, response["search"], "<script>")
}

func TestSanitizeString(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "Script tag removal",
			input:    "<script>alert('XSS')</script>Hello",
			expected: "Hello",
		},
		{
			name:     "Event handler removal",
			input:    "<div onclick='alert(1)'>Hello</div>",
			expected: "&lt;div&gt;Hello&lt;/div&gt;",
		},
		{
			name:     "SQL injection",
			input:    "admin' OR '1'='1",
			expected: "",
		},
		{
			name:     "UNION attack",
			input:    "test UNION SELECT * FROM users",
			expected: "",
		},
		{
			name:     "Whitespace trimming",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizationMiddleware.sanitizeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDetectSQLInjection(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "SQL keywords",
			input:    "SELECT * FROM users",
			expected: true,
		},
		{
			name:     "SQL comments",
			input:    "admin'--",
			expected: true,
		},
		{
			name:     "UNION attack",
			input:    "1 UNION SELECT password FROM users",
			expected: true,
		},
		{
			name:     "Boolean attack",
			input:    "admin' OR '1'='1",
			expected: true,
		},
		{
			name:     "Time-based attack",
			input:    "admin'; WAITFOR DELAY '00:00:05'--",
			expected: true,
		},
		{
			name:     "Normal email",
			input:    "user@example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizationMiddleware.detectSQLInjection(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestContainsSuspiciousPattern(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "Script tag",
			input:    "<script>alert(1)</script>",
			expected: true,
		},
		{
			name:     "JavaScript URL",
			input:    "javascript:alert(1)",
			expected: true,
		},
		{
			name:     "VBScript URL",
			input:    "vbscript:msgbox(1)",
			expected: true,
		},
		{
			name:     "Event handler",
			input:    "onload=alert(1)",
			expected: true,
		},
		{
			name:     "Eval function",
			input:    "eval(maliciousCode)",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizationMiddleware.containsSuspiciousPattern(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNormalizeInput(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Trim whitespace",
			input:    "  Hello World  ",
			expected: "Hello World",
		},
		{
			name:     "Normalize line breaks",
			input:    "Line1\r\nLine2\rLine3\nLine4",
			expected: "Line1\nLine2\nLine3\nLine4",
		},
		{
			name:     "Remove control characters",
			input:    "Hello\x00\x01World",
			expected: "HelloWorld",
		},
		{
			name:     "Max length truncation",
			input:    strings.Repeat("a", 1500), // 1500文字
			expected: strings.Repeat("a", 1000), // 1000文字に切り詰め
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitizationMiddleware.normalizeInput(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestValidateInput(t *testing.T) {
	config := GetDefaultSanitizationConfig()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "XSS attempt",
			input:    "<script>alert('XSS')</script>",
			expected: "",
		},
		{
			name:     "SQL injection",
			input:    "admin' OR '1'='1",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ValidateInput(tt.input, config)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEscapeHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "HTML tags",
			input:    "<script>alert('XSS')</script>",
			expected: "&lt;script&gt;alert(&#39;XSS&#39;)&lt;/script&gt;",
		},
		{
			name:     "Ampersand",
			input:    "Tom & Jerry",
			expected: "Tom &amp; Jerry",
		},
		{
			name:     "Quotes",
			input:    `He said "Hello"`,
			expected: "He said &#34;Hello&#34;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EscapeHTML(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsSafeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "Normal string",
			input:    "Hello World",
			expected: true,
		},
		{
			name:     "Script tag",
			input:    "<script>alert(1)</script>",
			expected: false,
		},
		{
			name:     "JavaScript URL",
			input:    "javascript:alert(1)",
			expected: false,
		},
		{
			name:     "SQL injection",
			input:    "admin' OR '1'='1",
			expected: false,
		},
		{
			name:     "UNION attack",
			input:    "1 UNION SELECT * FROM users",
			expected: false,
		},
		{
			name:     "Normal email",
			input:    "user@example.com",
			expected: true,
		},
		{
			name:     "Normal URL",
			input:    "https://example.com",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsSafeString(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetSanitizationStats(t *testing.T) {
	config := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(config)

	stats := sanitizationMiddleware.GetSanitizationStats()

	assert.True(t, stats["html_sanitization_enabled"].(bool))
	assert.True(t, stats["sql_injection_protection_enabled"].(bool))
	assert.True(t, stats["input_normalization_enabled"].(bool))
	assert.True(t, stats["file_upload_protection_enabled"].(bool))
	assert.Equal(t, 2, stats["excluded_paths_count"].(int))
}

func TestUpdateSanitizationConfig(t *testing.T) {
	initialConfig := GetDefaultSanitizationConfig()
	sanitizationMiddleware := NewSanitizationMiddleware(initialConfig)

	// 新しい設定
	newConfig := &SanitizationConfig{
		HTMLSanitization: &HTMLSanitizationConfig{
			Enabled:    false, // HTMLサニタイゼーションを無効化
			EscapeHTML: false,
		},
		SQLInjectionProtection: &SQLInjectionConfig{
			Enabled: true,
		},
		InputNormalization: &InputNormalizationConfig{
			Enabled: true,
		},
		ExcludedPaths: []string{"/api/v1/test"},
	}

	// 設定を更新
	sanitizationMiddleware.UpdateSanitizationConfig(newConfig)

	// 設定が更新されたことを確認
	stats := sanitizationMiddleware.GetSanitizationStats()
	assert.False(t, stats["html_sanitization_enabled"].(bool))
	assert.True(t, stats["sql_injection_protection_enabled"].(bool))
	assert.Equal(t, 1, stats["excluded_paths_count"].(int))
}
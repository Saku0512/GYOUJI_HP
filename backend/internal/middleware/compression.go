// Package middleware はHTTP圧縮とヘッダー最適化ミドルウェアを提供する
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// CompressionConfig は圧縮設定
type CompressionConfig struct {
	Level            int      // 圧縮レベル (1-9)
	MinLength        int      // 最小圧縮サイズ（バイト）
	ExcludedPaths    []string // 圧縮対象外のパス
	ExcludedMimeTypes []string // 圧縮対象外のMIMEタイプ
}

// DefaultCompressionConfig はデフォルト圧縮設定
var DefaultCompressionConfig = CompressionConfig{
	Level:     gzip.DefaultCompression,
	MinLength: 1024, // 1KB以上のレスポンスを圧縮
	ExcludedPaths: []string{
		"/health",
		"/metrics",
	},
	ExcludedMimeTypes: []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"video/",
		"audio/",
		"application/zip",
		"application/gzip",
		"application/x-gzip",
	},
}

// gzipWriter はgzip圧縮ライター
type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

// Write はデータを圧縮して書き込む
func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

// WriteString は文字列を圧縮して書き込む
func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

// Close は圧縮ライターを閉じる
func (g *gzipWriter) Close() error {
	return g.writer.Close()
}

// gzipWriterPool はgzipライターのプール
var gzipWriterPool = sync.Pool{
	New: func() interface{} {
		gz, _ := gzip.NewWriterLevel(io.Discard, gzip.DefaultCompression)
		return gz
	},
}

// getGzipWriter はプールからgzipライターを取得する
func getGzipWriter(w io.Writer, level int) *gzip.Writer {
	gz := gzipWriterPool.Get().(*gzip.Writer)
	gz.Reset(w)
	return gz
}

// putGzipWriter はgzipライターをプールに戻す
func putGzipWriter(gz *gzip.Writer) {
	gz.Close()
	gzipWriterPool.Put(gz)
}

// GzipMiddleware はgzip圧縮ミドルウェアを作成する
func GzipMiddleware(config ...CompressionConfig) gin.HandlerFunc {
	cfg := DefaultCompressionConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// 圧縮対象外のパスをチェック
		for _, path := range cfg.ExcludedPaths {
			if strings.HasPrefix(c.Request.URL.Path, path) {
				c.Next()
				return
			}
		}

		// クライアントがgzip圧縮をサポートしているかチェック
		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		// レスポンスヘッダーを設定
		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		// gzipライターを作成
		gz := getGzipWriter(c.Writer, cfg.Level)
		defer putGzipWriter(gz)

		// カスタムレスポンスライターを設定
		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		// 次のハンドラーを実行
		c.Next()

		// gzipライターを閉じる
		gz.Close()
	}
}

// CacheConfig はキャッシュ設定
type CacheConfig struct {
	// 静的リソース用キャッシュ設定
	StaticMaxAge    time.Duration
	StaticPaths     []string
	
	// API用キャッシュ設定
	APIMaxAge       time.Duration
	APIPaths        []string
	
	// キャッシュ無効設定
	NoCachePaths    []string
	
	// ETag有効化
	EnableETag      bool
	
	// Last-Modified有効化
	EnableLastModified bool
}

// DefaultCacheConfig はデフォルトキャッシュ設定
var DefaultCacheConfig = CacheConfig{
	StaticMaxAge: 24 * time.Hour, // 静的リソースは24時間キャッシュ
	StaticPaths: []string{
		"/static/",
		"/assets/",
		"/images/",
		"/css/",
		"/js/",
	},
	APIMaxAge: 5 * time.Minute, // APIレスポンスは5分キャッシュ
	APIPaths: []string{
		"/api/v1/tournaments",
		"/api/v1/matches",
	},
	NoCachePaths: []string{
		"/api/v1/auth/",
		"/api/v1/admin/",
		"/health",
		"/metrics",
	},
	EnableETag:         true,
	EnableLastModified: true,
}

// CacheMiddleware はキャッシュヘッダー設定ミドルウェアを作成する
func CacheMiddleware(config ...CacheConfig) gin.HandlerFunc {
	cfg := DefaultCacheConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// キャッシュ無効パスをチェック
		for _, noCache := range cfg.NoCachePaths {
			if strings.HasPrefix(path, noCache) {
				c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
				c.Header("Pragma", "no-cache")
				c.Header("Expires", "0")
				c.Next()
				return
			}
		}

		// 静的リソースのキャッシュ設定
		for _, staticPath := range cfg.StaticPaths {
			if strings.HasPrefix(path, staticPath) {
				maxAge := int(cfg.StaticMaxAge.Seconds())
				c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
				
				if cfg.EnableLastModified {
					c.Header("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
				}
				
				c.Next()
				return
			}
		}

		// APIレスポンスのキャッシュ設定
		for _, apiPath := range cfg.APIPaths {
			if strings.HasPrefix(path, apiPath) {
				maxAge := int(cfg.APIMaxAge.Seconds())
				c.Header("Cache-Control", "public, max-age="+strconv.Itoa(maxAge))
				
				if cfg.EnableETag {
					// 簡単なETag生成（実際の実装では内容ベースのハッシュを使用）
					etag := `"` + strconv.FormatInt(time.Now().Unix(), 36) + `"`
					c.Header("ETag", etag)
				}
				
				c.Next()
				return
			}
		}

		// デフォルト設定（短時間キャッシュ）
		c.Header("Cache-Control", "public, max-age=300") // 5分
		c.Next()
	}
}

// SecurityHeadersConfig はセキュリティヘッダー設定
type SecurityHeadersConfig struct {
	ContentSecurityPolicy string
	StrictTransportSecurity string
	XFrameOptions         string
	XContentTypeOptions   string
	ReferrerPolicy        string
	PermissionsPolicy     string
}

// DefaultSecurityHeadersConfig はデフォルトセキュリティヘッダー設定
var DefaultSecurityHeadersConfig = SecurityHeadersConfig{
	ContentSecurityPolicy: "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self' ws: wss:",
	StrictTransportSecurity: "max-age=31536000; includeSubDomains",
	XFrameOptions:         "DENY",
	XContentTypeOptions:   "nosniff",
	ReferrerPolicy:        "strict-origin-when-cross-origin",
	PermissionsPolicy:     "geolocation=(), microphone=(), camera=()",
}

// SecurityHeadersMiddleware はセキュリティヘッダー設定ミドルウェアを作成する
func SecurityHeadersMiddleware(config ...SecurityHeadersConfig) gin.HandlerFunc {
	cfg := DefaultSecurityHeadersConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *gin.Context) {
		// セキュリティヘッダーを設定
		if cfg.ContentSecurityPolicy != "" {
			c.Header("Content-Security-Policy", cfg.ContentSecurityPolicy)
		}
		
		if cfg.StrictTransportSecurity != "" && c.Request.TLS != nil {
			c.Header("Strict-Transport-Security", cfg.StrictTransportSecurity)
		}
		
		if cfg.XFrameOptions != "" {
			c.Header("X-Frame-Options", cfg.XFrameOptions)
		}
		
		if cfg.XContentTypeOptions != "" {
			c.Header("X-Content-Type-Options", cfg.XContentTypeOptions)
		}
		
		if cfg.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", cfg.ReferrerPolicy)
		}
		
		if cfg.PermissionsPolicy != "" {
			c.Header("Permissions-Policy", cfg.PermissionsPolicy)
		}

		// XSS保護
		c.Header("X-XSS-Protection", "1; mode=block")
		
		c.Next()
	}
}

// ResponseOptimizationMiddleware はレスポンス最適化ミドルウェアを作成する
func ResponseOptimizationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// レスポンス時間測定開始
		start := time.Now()

		// リクエストIDを設定（ログ追跡用）
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
			c.Header("X-Request-ID", requestID)
		}
		c.Set("request_id", requestID)

		// レスポンスヘッダーを設定
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")

		// 次のハンドラーを実行
		c.Next()

		// レスポンス時間をヘッダーに追加
		duration := time.Since(start)
		c.Header("X-Response-Time", duration.String())

		// パフォーマンス情報をログ出力
		if duration > 1*time.Second {
			// 1秒以上かかった場合は警告ログ
			c.Set("slow_request", true)
		}
	})
}

// generateRequestID はリクエストIDを生成する
func generateRequestID() string {
	return "req_" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// ConditionalGetMiddleware は条件付きGETリクエスト処理ミドルウェア
func ConditionalGetMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// If-None-Matchヘッダーをチェック
		ifNoneMatch := c.GetHeader("If-None-Match")
		if ifNoneMatch != "" && c.Request.Method == "GET" {
			// ETagが一致する場合は304を返す
			c.Next()
			
			etag := c.GetHeader("ETag")
			if etag != "" && ifNoneMatch == etag {
				c.Status(http.StatusNotModified)
				return
			}
		}

		// If-Modified-Sinceヘッダーをチェック
		ifModifiedSince := c.GetHeader("If-Modified-Since")
		if ifModifiedSince != "" && c.Request.Method == "GET" {
			if t, err := time.Parse(http.TimeFormat, ifModifiedSince); err == nil {
				c.Next()
				
				lastModified := c.GetHeader("Last-Modified")
				if lastModified != "" {
					if lastModTime, err := time.Parse(http.TimeFormat, lastModified); err == nil {
						if !lastModTime.After(t) {
							c.Status(http.StatusNotModified)
							return
						}
					}
				}
			}
		}

		c.Next()
	}
}
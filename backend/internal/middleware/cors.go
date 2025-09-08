package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSConfig はCORS設定の構造体
type CORSConfig struct {
	AllowOrigins     []string      // 許可するオリジン
	AllowMethods     []string      // 許可するHTTPメソッド
	AllowHeaders     []string      // 許可するヘッダー
	ExposeHeaders    []string      // 公開するヘッダー
	AllowCredentials bool          // 認証情報の送信を許可するか
	MaxAge           time.Duration // プリフライトリクエストのキャッシュ時間
}

// GetDefaultCORSConfig はデフォルトのCORS設定を取得する
func GetDefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{
			"http://localhost:3000",     // React開発サーバー
			"http://localhost:5173",     // Vite開発サーバー
			"http://localhost:8080",     // 開発用フロントエンド
			"http://127.0.0.1:3000",
			"http://127.0.0.1:5173",
			"http://127.0.0.1:8080",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"Cache-Control",
			"Pragma",
			"Expires",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-Request-ID",
			"X-Response-Time",
		},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
}

// GetProductionCORSConfig は本番環境用のCORS設定を取得する
func GetProductionCORSConfig() *CORSConfig {
	config := GetDefaultCORSConfig()
	
	// 本番環境のオリジンを環境変数から取得
	allowedOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if allowedOrigins != "" {
		config.AllowOrigins = strings.Split(allowedOrigins, ",")
		// 各オリジンの前後の空白を削除
		for i, origin := range config.AllowOrigins {
			config.AllowOrigins[i] = strings.TrimSpace(origin)
		}
	} else {
		// デフォルトの本番環境オリジン
		config.AllowOrigins = []string{
			"https://tournament.example.com",
			"https://www.tournament.example.com",
		}
	}
	
	return config
}

// GetDevelopmentCORSConfig は開発環境用のCORS設定を取得する
func GetDevelopmentCORSConfig() *CORSConfig {
	config := GetDefaultCORSConfig()
	
	// 開発環境では追加のオリジンを許可
	additionalOrigins := []string{
		"http://localhost:4000",
		"http://localhost:5000",
		"http://127.0.0.1:4000",
		"http://127.0.0.1:5000",
	}
	
	config.AllowOrigins = append(config.AllowOrigins, additionalOrigins...)
	
	return config
}

// NewCORSMiddleware は統一されたCORSミドルウェアを作成する
func NewCORSMiddleware() gin.HandlerFunc {
	var config *CORSConfig
	
	// 環境に応じて設定を選択
	env := os.Getenv("GIN_MODE")
	if env == "release" || env == "production" {
		config = GetProductionCORSConfig()
	} else {
		config = GetDevelopmentCORSConfig()
	}
	
	return cors.New(cors.Config{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	})
}

// NewCustomCORSMiddleware はカスタム設定でCORSミドルウェアを作成する
func NewCustomCORSMiddleware(config *CORSConfig) gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:     config.AllowOrigins,
		AllowMethods:     config.AllowMethods,
		AllowHeaders:     config.AllowHeaders,
		ExposeHeaders:    config.ExposeHeaders,
		AllowCredentials: config.AllowCredentials,
		MaxAge:           config.MaxAge,
	})
}

// SecurityHeadersMiddleware はセキュリティヘッダーを設定するミドルウェア
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// セキュリティヘッダーを設定
		setSecurityHeaders(c)
		c.Next()
	}
}

// setSecurityHeaders はセキュリティヘッダーを設定する
func setSecurityHeaders(c *gin.Context) {
	// X-Content-Type-Options: MIMEタイプスニッフィングを防ぐ
	c.Header("X-Content-Type-Options", "nosniff")
	
	// X-Frame-Options: クリックジャッキング攻撃を防ぐ
	c.Header("X-Frame-Options", "DENY")
	
	// X-XSS-Protection: XSS攻撃を防ぐ（古いブラウザ用）
	c.Header("X-XSS-Protection", "1; mode=block")
	
	// Referrer-Policy: リファラー情報の送信を制御
	c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
	
	// Content-Security-Policy: XSS攻撃を防ぐ
	csp := getContentSecurityPolicy()
	c.Header("Content-Security-Policy", csp)
	
	// 本番環境でのみHTTPS関連のヘッダーを設定
	if isProduction() {
		// Strict-Transport-Security: HTTPS接続を強制
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
		
		// Expect-CT: Certificate Transparencyを要求
		c.Header("Expect-CT", "max-age=86400, enforce")
	}
	
	// Permissions-Policy: ブラウザ機能の使用を制御
	c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")
}

// getContentSecurityPolicy はContent Security Policyを取得する
func getContentSecurityPolicy() string {
	if isProduction() {
		// 本番環境用の厳格なCSP
		return "default-src 'self'; " +
			"script-src 'self' 'unsafe-inline'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: https:; " +
			"font-src 'self'; " +
			"connect-src 'self'; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'; " +
			"form-action 'self'"
	} else {
		// 開発環境用の緩いCSP
		return "default-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"script-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
			"style-src 'self' 'unsafe-inline'; " +
			"img-src 'self' data: blob: https: http:; " +
			"font-src 'self' data:; " +
			"connect-src 'self' ws: wss: http: https:; " +
			"frame-ancestors 'none'; " +
			"base-uri 'self'"
	}
}

// isProduction は本番環境かどうかを判定する
func isProduction() bool {
	env := os.Getenv("GIN_MODE")
	return env == "release" || env == "production"
}

// PreflightMiddleware はプリフライトリクエストを処理するミドルウェア
func PreflightMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			// プリフライトリクエストの場合は204を返す
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

// RequestIDMiddleware はリクエストIDを設定するミドルウェア
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 既存のリクエストIDがあるかチェック
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			// リクエストIDを生成
			requestID = generateRequestID()
		}
		
		// コンテキストとレスポンスヘッダーに設定
		c.Set("request_id", requestID)
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// generateRequestID はリクエストIDを生成する
func generateRequestID() string {
	// 簡易的なリクエストID生成（本番環境ではより堅牢な実装を推奨）
	return "req_" + generateRandomString(16)
}

// generateRandomString はランダム文字列を生成する
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[len(charset)/2] // 簡易実装
	}
	return string(b)
}

// CombinedMiddleware は複数のミドルウェアを組み合わせる
func CombinedMiddleware() []gin.HandlerFunc {
	return []gin.HandlerFunc{
		RequestIDMiddleware(),    // リクエストID設定
		SecurityHeadersMiddleware(), // セキュリティヘッダー設定
		NewCORSMiddleware(),     // CORS設定
		PreflightMiddleware(),   // プリフライト処理
	}
}

// LegacyCORSMiddleware は旧式のCORSミドルウェア（廃止予定）
// 新しいコードでは NewCORSMiddleware を使用してください
func LegacyCORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
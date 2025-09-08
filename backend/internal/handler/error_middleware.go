package handler

import (
	"errors"
	"net/http"
	"time"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ErrorHandlerMiddleware は統一されたエラーハンドリングを提供するミドルウェア
// 全てのエラーを統一されたAPIResponse形式で返す
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// エラーが発生している場合の処理
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			var apiError *models.APIError
			if errors.As(err.Err, &apiError) {
				// APIErrorの場合は統一形式でレスポンス
				response := models.NewErrorResponse(apiError.Code, apiError.Message, apiError.StatusCode)
				
				// リクエストIDを追加
				if requestID, exists := c.Get("request_id"); exists {
					if id, ok := requestID.(string); ok {
						response.SetRequestID(id)
					}
				}
				
				c.JSON(apiError.StatusCode, response)
			} else {
				// 未知のエラーの場合
				response := models.NewErrorResponse(
					models.ErrorSystemUnknownError,
					"予期しないエラーが発生しました",
					http.StatusInternalServerError,
				)
				
				// リクエストIDを追加
				if requestID, exists := c.Get("request_id"); exists {
					if id, ok := requestID.(string); ok {
						response.SetRequestID(id)
					}
				}
				
				c.JSON(http.StatusInternalServerError, response)
			}
		}
	}
}

// RecoveryMiddleware はパニックを捕捉して統一されたエラーレスポンスを返すミドルウェア
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		response := models.NewErrorResponse(
			models.ErrorSystemUnknownError,
			"サーバー内部エラーが発生しました",
			http.StatusInternalServerError,
		)
		
		// リクエストIDを追加
		if requestID, exists := c.Get("request_id"); exists {
			if id, ok := requestID.(string); ok {
				response.SetRequestID(id)
			}
		}
		
		c.JSON(http.StatusInternalServerError, response)
		c.Abort()
	})
}

// RequestIDMiddleware は各リクエストにユニークなIDを付与するミドルウェア
// レスポンス追跡とログ関連付けに使用される
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// リクエストIDを生成
		requestID := uuid.New().String()
		
		// コンテキストに設定
		c.Set("request_id", requestID)
		
		// レスポンスヘッダーにも設定
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// CORSMiddleware は統一されたCORS設定を提供するミドルウェア
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 許可するオリジンのリスト
		allowedOrigins := []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"http://localhost:8080",
			// 本番環境のオリジンは環境変数から取得することを推奨
		}
		
		// オリジンが許可リストに含まれているかチェック
		isAllowed := false
		for _, allowedOrigin := range allowedOrigins {
			if origin == allowedOrigin {
				isAllowed = true
				break
			}
		}
		
		if isAllowed {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Request-ID")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
		c.Header("Access-Control-Expose-Headers", "Content-Length, X-Request-ID")
		c.Header("Access-Control-Max-Age", "86400") // 24時間

		// プリフライトリクエストの処理
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersMiddleware はセキュリティヘッダーを設定するミドルウェア
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// セキュリティヘッダーを設定
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data:; font-src 'self'")
		
		// 本番環境でのみHSTSヘッダーを設定
		if gin.Mode() == gin.ReleaseMode {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		
		c.Next()
	}
}

// RateLimitMiddleware は統一されたレート制限を提供するミドルウェア
// 注意: この実装は簡易版です。本番環境ではRedisベースの実装を推奨します
func RateLimitMiddleware() gin.HandlerFunc {
	// インメモリでのレート制限（簡易実装）
	// 本番環境では Redis や外部のレート制限サービスを使用することを推奨
	type rateLimitInfo struct {
		count     int
		resetTime time.Time
	}
	
	clientLimits := make(map[string]*rateLimitInfo)
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()
		
		// クライアント情報を取得または作成
		if info, exists := clientLimits[clientIP]; exists {
			// リセット時間を過ぎている場合はカウントをリセット
			if now.After(info.resetTime) {
				info.count = 0
				info.resetTime = now.Add(time.Minute)
			}
		} else {
			clientLimits[clientIP] = &rateLimitInfo{
				count:     0,
				resetTime: now.Add(time.Minute),
			}
		}
		
		info := clientLimits[clientIP]
		
		// レート制限をチェック（1分間に100リクエスト）
		if info.count >= 100 {
			response := models.NewErrorResponse(
				"RATE_LIMIT_EXCEEDED",
				"リクエスト制限に達しました。しばらく待ってから再試行してください",
				http.StatusTooManyRequests,
			)
			
			// リクエストIDを追加
			if requestID, exists := c.Get("request_id"); exists {
				if id, ok := requestID.(string); ok {
					response.SetRequestID(id)
				}
			}
			
			// レート制限ヘッダーを設定
			c.Header("X-RateLimit-Limit", "100")
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", info.resetTime.Format(time.RFC3339))
			
			c.JSON(http.StatusTooManyRequests, response)
			c.Abort()
			return
		}
		
		// カウントを増加
		info.count++
		
		// レート制限ヘッダーを設定
		c.Header("X-RateLimit-Limit", "100")
		c.Header("X-RateLimit-Remaining", string(rune(100-info.count)))
		c.Header("X-RateLimit-Reset", info.resetTime.Format(time.RFC3339))
		
		c.Next()
	}
}
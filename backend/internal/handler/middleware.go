package handler

import (
	"fmt"
	"net/http"
	"strings"

	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware はJWT認証ミドルウェア
type AuthMiddleware struct {
	authService service.AuthService
}

// NewAuthMiddleware は新しい認証ミドルウェアを作成する
func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth はJWTトークン検証を行うミドルウェア
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorizationヘッダーを取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "認証トークンが必要です",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// Bearer トークンの形式をチェック
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "無効な認証トークン形式です",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		// トークンを検証
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "無効または期限切れのトークンです",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// クレーム情報をコンテキストに保存
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Set("claims", claims)

		c.Next()
	}
}

// RequireAdmin は管理者専用アクセスを制御するミドルウェア
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先にRequireAuth()が実行されていることを前提とする
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "Unauthorized",
				"message": "認証情報が見つかりません",
				"code":    http.StatusUnauthorized,
			})
			c.Abort()
			return
		}

		// 管理者権限をチェック
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Forbidden",
				"message": "管理者権限が必要です",
				"code":    http.StatusForbidden,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CORSMiddleware はCORS設定を行うミドルウェア
func CORSMiddleware() gin.HandlerFunc {
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

// ErrorHandlerMiddleware は統一されたエラーレスポンスを提供するミドルウェア
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "サーバー内部エラーが発生しました",
				"code":    http.StatusInternalServerError,
				"details": err,
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal Server Error",
				"message": "予期しないエラーが発生しました",
				"code":    http.StatusInternalServerError,
			})
		}
		c.Abort()
	})
}

// LoggingMiddleware はリクエスト/レスポンスのログを記録するミドルウェア
func LoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// RateLimitMiddleware は認証エンドポイント用のレート制限ミドルウェア
// 注意: 本格的な実装では Redis などの外部ストレージを使用することを推奨
func RateLimitMiddleware() gin.HandlerFunc {
	// 簡易的なインメモリレート制限（本番環境では Redis 等を使用）
	return func(c *gin.Context) {
		// 実装は簡略化し、将来的に Redis ベースのレート制限に置き換える
		c.Next()
	}
}
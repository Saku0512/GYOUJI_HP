package handler

import (
	"fmt"
	"net/http"
	"strings"

	"backend/internal/models"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware はJWT認証ミドルウェア（廃止予定）
// 新しいコードでは backend/internal/middleware/auth.go の AuthMiddleware を使用してください
type AuthMiddleware struct {
	authService service.AuthService
}

// NewAuthMiddleware は新しい認証ミドルウェアを作成する（廃止予定）
// 新しいコードでは middleware.NewAuthMiddleware を使用してください
func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// RequireAuth はJWTトークン検証を行うミドルウェア（廃止予定）
// 新しいコードでは middleware.AuthMiddleware.RequireAuth を使用してください
func (m *AuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorizationヘッダーを取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Error(models.NewAPIError(models.ErrorAuthUnauthorized, "認証トークンが必要です", http.StatusUnauthorized))
			c.Abort()
			return
		}

		// Bearer トークンの形式をチェック
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.Error(models.NewAPIError(models.ErrorAuthTokenInvalid, "無効な認証トークン形式です", http.StatusUnauthorized))
			c.Abort()
			return
		}

		token := tokenParts[1]

		// トークンを検証
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			c.Error(models.NewAPIError(models.ErrorAuthTokenInvalid, "無効または期限切れのトークンです", http.StatusUnauthorized))
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

// RequireAdmin は管理者専用アクセスを制御するミドルウェア（廃止予定）
// 新しいコードでは middleware.AuthMiddleware.RequireAdmin を使用してください
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先にRequireAuth()が実行されていることを前提とする
		role, exists := c.Get("role")
		if !exists {
			c.Error(models.NewAPIError(models.ErrorAuthUnauthorized, "認証情報が見つかりません", http.StatusUnauthorized))
			c.Abort()
			return
		}

		// 管理者権限をチェック
		if role != models.RoleAdmin {
			c.Error(models.NewAPIError(models.ErrorAuthForbidden, "管理者権限が必要です", http.StatusForbidden))
			c.Abort()
			return
		}

		c.Next()
	}
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


package middleware

import (
	"net/http"
	"strings"

	"backend/internal/models"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware は統一された認証ミドルウェア
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
			m.sendAuthError(c, models.ErrorAuthUnauthorized, "認証トークンが必要です", http.StatusUnauthorized)
			return
		}

		// Bearer トークンの形式をチェック
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			m.sendAuthError(c, models.ErrorAuthTokenInvalid, "無効な認証トークン形式です", http.StatusUnauthorized)
			return
		}

		token := tokenParts[1]
		if strings.TrimSpace(token) == "" {
			m.sendAuthError(c, models.ErrorAuthTokenInvalid, "認証トークンが空です", http.StatusUnauthorized)
			return
		}

		// トークンを検証
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			m.sendAuthError(c, models.ErrorAuthTokenInvalid, "無効または期限切れのトークンです", http.StatusUnauthorized)
			return
		}

		// クレーム情報をコンテキストに保存
		m.setUserContext(c, claims)

		c.Next()
	}
}

// RequireAdmin は管理者専用アクセスを制御するミドルウェア
func (m *AuthMiddleware) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先にRequireAuth()が実行されていることを前提とする
		role, exists := c.Get("role")
		if !exists {
			m.sendAuthError(c, models.ErrorAuthUnauthorized, "認証情報が見つかりません", http.StatusUnauthorized)
			return
		}

		// 管理者権限をチェック
		if role != models.RoleAdmin {
			m.sendAuthError(c, models.ErrorAuthForbidden, "管理者権限が必要です", http.StatusForbidden)
			return
		}

		c.Next()
	}
}

// OptionalAuth は任意の認証を行うミドルウェア（認証エラーでも処理を継続）
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Authorizationヘッダーを取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 認証情報がない場合はそのまま継続
			c.Next()
			return
		}

		// Bearer トークンの形式をチェック
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// 無効な形式の場合はそのまま継続
			c.Next()
			return
		}

		token := tokenParts[1]
		if strings.TrimSpace(token) == "" {
			// 空のトークンの場合はそのまま継続
			c.Next()
			return
		}

		// トークンを検証
		claims, err := m.authService.ValidateToken(token)
		if err != nil {
			// 検証エラーの場合はそのまま継続
			c.Next()
			return
		}

		// 有効な認証情報をコンテキストに保存
		m.setUserContext(c, claims)

		c.Next()
	}
}

// RequireRole は指定されたロールを要求するミドルウェア
func (m *AuthMiddleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先にRequireAuth()が実行されていることを前提とする
		role, exists := c.Get("role")
		if !exists {
			m.sendAuthError(c, models.ErrorAuthUnauthorized, "認証情報が見つかりません", http.StatusUnauthorized)
			return
		}

		// 指定されたロールをチェック
		if role != requiredRole {
			m.sendAuthError(c, models.ErrorAuthForbidden, "必要な権限がありません", http.StatusForbidden)
			return
		}

		c.Next()
	}
}

// RequireUser は特定のユーザーまたは管理者のアクセスを制御するミドルウェア
func (m *AuthMiddleware) RequireUser(userID int) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 先にRequireAuth()が実行されていることを前提とする
		contextUserID, exists := c.Get("user_id")
		if !exists {
			m.sendAuthError(c, models.ErrorAuthUnauthorized, "認証情報が見つかりません", http.StatusUnauthorized)
			return
		}

		role, roleExists := c.Get("role")
		if !roleExists {
			m.sendAuthError(c, models.ErrorAuthUnauthorized, "ロール情報が見つかりません", http.StatusUnauthorized)
			return
		}

		// 管理者または本人の場合はアクセス許可
		if role == models.RoleAdmin || contextUserID == userID {
			c.Next()
			return
		}

		m.sendAuthError(c, models.ErrorAuthForbidden, "アクセス権限がありません", http.StatusForbidden)
	}
}

// sendAuthError は統一された認証エラーレスポンスを送信する
func (m *AuthMiddleware) sendAuthError(c *gin.Context, errorCode string, message string, statusCode int) {
	// 統一されたエラーレスポンスを作成
	response := models.NewErrorResponse(errorCode, message, statusCode)
	
	// リクエストIDを追加
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			response.RequestID = id
		}
	}
	
	c.JSON(statusCode, response)
	c.Abort()
}

// setUserContext はユーザー情報をコンテキストに設定する
func (m *AuthMiddleware) setUserContext(c *gin.Context, claims *service.JWTClaims) {
	c.Set("user_id", claims.UserID)
	c.Set("username", claims.Username)
	c.Set("role", claims.Role)
	c.Set("claims", claims)
	
	// 追加のメタデータ
	if claims.ExpiresAt != nil {
		c.Set("token_expires_at", claims.ExpiresAt.Time)
	}
	if claims.IssuedAt != nil {
		c.Set("token_issued_at", claims.IssuedAt.Time)
	}
}

// GetUserID はコンテキストからユーザーIDを取得する
func GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	
	if id, ok := userID.(int); ok {
		return id, true
	}
	
	return 0, false
}

// GetUsername はコンテキストからユーザー名を取得する
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	
	if name, ok := username.(string); ok {
		return name, true
	}
	
	return "", false
}

// GetUserRole はコンテキストからユーザーロールを取得する
func GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	
	if r, ok := role.(string); ok {
		return r, true
	}
	
	return "", false
}

// GetClaims はコンテキストからJWTクレームを取得する
func GetClaims(c *gin.Context) (*service.JWTClaims, bool) {
	claims, exists := c.Get("claims")
	if !exists {
		return nil, false
	}
	
	if jwtClaims, ok := claims.(*service.JWTClaims); ok {
		return jwtClaims, true
	}
	
	return nil, false
}

// IsAuthenticated はユーザーが認証されているかどうかを確認する
func IsAuthenticated(c *gin.Context) bool {
	_, exists := c.Get("user_id")
	return exists
}

// IsAdmin はユーザーが管理者かどうかを確認する
func IsAdmin(c *gin.Context) bool {
	role, exists := GetUserRole(c)
	return exists && role == models.RoleAdmin
}
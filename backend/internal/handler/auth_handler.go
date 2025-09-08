package handler

import (
	"net/http"
	"strings"

	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler は認証関連のHTTPハンドラー
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler は新しいAuthHandlerを作成する
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// LoginRequest はログインリクエストの構造体
type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"admin"`        // ユーザー名
	Password string `json:"password" binding:"required" example:"password"`     // パスワード
}

// LoginResponse はログインレスポンスの構造体
type LoginResponse struct {
	Success  bool   `json:"success" example:"true"`                                                    // 成功フラグ
	Token    string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`                  // JWTトークン
	Username string `json:"username" example:"admin"`                                                  // ユーザー名
	Role     string `json:"role" example:"admin"`                                                      // ユーザーロール
	Message  string `json:"message" example:"ログインに成功しました"`                                          // メッセージ
}

// RefreshTokenRequest はトークンリフレッシュリクエストの構造体
type RefreshTokenRequest struct {
	Token string `json:"token" binding:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."` // 既存のJWTトークン
}

// RefreshTokenResponse はトークンリフレッシュレスポンスの構造体
type RefreshTokenResponse struct {
	Success bool   `json:"success" example:"true"`                                                    // 成功フラグ
	Token   string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`                  // 新しいJWTトークン
	Message string `json:"message" example:"トークンのリフレッシュに成功しました"`                                // メッセージ
}

// ErrorResponse は統一されたエラーレスポンスの構造体
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// Login はログインエンドポイントハンドラー
// @Summary ユーザーログイン
// @Description 管理者認証情報でログインし、JWTトークンを取得する
// @Tags auth
// @Accept json
// @Produce json
// @Param request body LoginRequest true "ログイン情報"
// @Success 200 {object} LoginResponse "ログイン成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なリクエスト形式です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 入力値の検証
	if strings.TrimSpace(req.Username) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "ユーザー名は必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	if strings.TrimSpace(req.Password) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "パスワードは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 認証処理
	token, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		// 認証エラーの場合は401を返す
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "認証に失敗しました。ユーザー名またはパスワードが正しくありません",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, LoginResponse{
		Token:    token,
		Username: req.Username,
		Role:     "admin", // 現在は管理者のみサポート
		Message:  "ログインに成功しました",
	})
}

// RefreshToken はトークンリフレッシュエンドポイントハンドラー
// @Summary JWTトークンリフレッシュ
// @Description 既存のJWTトークンを使用して新しいトークンを生成する
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest true "リフレッシュ情報"
// @Success 200 {object} RefreshTokenResponse "リフレッシュ成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Failure 500 {object} ErrorResponse "サーバーエラー"
// @Router /api/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req RefreshTokenRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "無効なリクエスト形式です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 入力値の検証
	if strings.TrimSpace(req.Token) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "トークンは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// 既存のトークンを検証してリフレッシュ
	newToken, err := h.authService.RefreshToken(req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "無効または期限切れのトークンです",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// 成功レスポンス
	c.JSON(http.StatusOK, RefreshTokenResponse{
		Token:   newToken,
		Message: "トークンのリフレッシュに成功しました",
	})
}

// Logout はログアウトエンドポイントハンドラー
// @Summary ユーザーログアウト
// @Description ログアウト処理（JWTはステートレスのため、クライアント側でトークンを削除）
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]string "ログアウト成功"
// @Router /api/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWTはステートレスなので、サーバー側では特別な処理は不要
	// クライアント側でトークンを削除することでログアウトが完了する
	c.JSON(http.StatusOK, gin.H{
		"message": "ログアウトしました。クライアント側でトークンを削除してください",
	})
}

// GetProfile は現在のユーザー情報を取得するエンドポイントハンドラー
// @Summary ユーザープロフィール取得
// @Description 現在認証されているユーザーの情報を取得する
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "ユーザー情報"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Router /api/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// ミドルウェアで設定されたユーザー情報を取得
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "認証情報が見つかりません",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	username, _ := c.Get("username")
	role, _ := c.Get("role")

	// ユーザー情報を返す
	c.JSON(http.StatusOK, gin.H{
		"user_id":  userID,
		"username": username,
		"role":     role,
		"message":  "ユーザー情報を取得しました",
	})
}

// ValidateToken はトークン検証エンドポイントハンドラー
// @Summary トークン検証
// @Description JWTトークンの有効性を検証する（POSTリクエストボディまたはGETのAuthorizationヘッダーから）
// @Tags auth
// @Accept json
// @Produce json
// @Param request body RefreshTokenRequest false "検証するトークン（POSTの場合）"
// @Param Authorization header string false "Bearer トークン（GETの場合）"
// @Success 200 {object} map[string]interface{} "検証成功"
// @Failure 400 {object} ErrorResponse "リクエストエラー"
// @Failure 401 {object} ErrorResponse "認証エラー"
// @Router /api/auth/validate [post]
// @Router /api/auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var token string

	// リクエストメソッドに応じてトークンを取得
	if c.Request.Method == "POST" {
		// POSTリクエストの場合：リクエストボディから取得
		var req RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "無効なリクエスト形式です",
				Code:    http.StatusBadRequest,
			})
			return
		}
		token = req.Token
	} else {
		// GETリクエストの場合：Authorizationヘッダーから取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "Authorizationヘッダーが必要です",
				Code:    http.StatusBadRequest,
			})
			return
		}

		// "Bearer "プレフィックスを削除
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			c.JSON(http.StatusBadRequest, ErrorResponse{
				Error:   "Bad Request",
				Message: "無効なAuthorizationヘッダー形式です",
				Code:    http.StatusBadRequest,
			})
			return
		}
	}

	// 入力値の検証
	if strings.TrimSpace(token) == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{
			Error:   "Bad Request",
			Message: "トークンは必須です",
			Code:    http.StatusBadRequest,
		})
		return
	}

	// トークンを検証
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, ErrorResponse{
			Error:   "Unauthorized",
			Message: "無効または期限切れのトークンです",
			Code:    http.StatusUnauthorized,
		})
		return
	}

	// 検証成功レスポンス
	c.JSON(http.StatusOK, gin.H{
		"valid":     true,
		"user_id":   claims.UserID,
		"username":  claims.Username,
		"role":      claims.Role,
		"expires_at": claims.ExpiresAt,
		"message":   "トークンは有効です",
	})
}
package handler

import (
	"net/http"
	"strings"

	"backend/internal/models"
	"backend/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler は認証関連のHTTPハンドラー
type AuthHandler struct {
	*BaseHandler
	authService service.AuthService
}

// NewAuthHandler は新しいAuthHandlerを作成する
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		BaseHandler: NewBaseHandler(),
		authService: authService,
	}
}



// Login はログインエンドポイントハンドラー
// @Summary ユーザーログイン
// @Description 管理者認証情報でログインし、JWTトークンを取得する
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "ログイン情報"
// @Success 200 {object} models.DataResponse[models.LoginResponse] "ログイン成功"
// @Failure 400 {object} models.ValidationErrorResponse "バリデーションエラー"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req models.LoginRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		h.SendBindingError(c, err)
		return
	}

	// 統一されたバリデーション実行
	if !h.ValidateRequest(c, func() models.ValidationErrors {
		return models.ValidateLoginRequest(&req)
	}) {
		return
	}

	// 認証処理
	token, claims, err := h.authService.Login(req.Username, req.Password)
	if err != nil {
		// 認証エラーの場合は401を返す
		h.SendError(c, models.ErrInvalidCredentials)
		return
	}

	// 成功レスポンス
	loginResponse := &models.LoginResponse{
		Token:     token,
		Username:  claims.Username,
		Role:      claims.Role,
		ExpiresAt: models.NewDateTime(claims.ExpiresAt.Time),
	}

	h.SendSuccess(c, loginResponse, "ログインに成功しました")
}

// RefreshToken はトークンリフレッシュエンドポイントハンドラー
// @Summary JWTトークンリフレッシュ
// @Description 既存のJWTトークンを使用して新しいトークンを生成する
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest true "リフレッシュ情報"
// @Success 200 {object} models.DataResponse[models.RefreshTokenResponse] "リフレッシュ成功"
// @Failure 400 {object} models.ValidationErrorResponse "バリデーションエラー"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Failure 500 {object} models.ErrorResponse "サーバーエラー"
// @Router /api/v1/auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req models.RefreshTokenRequest

	// リクエストボディをバインド
	if err := c.ShouldBindJSON(&req); err != nil {
		h.SendBindingError(c, err)
		return
	}

	// 統一されたバリデーション実行
	if !h.ValidateRequest(c, func() models.ValidationErrors {
		validator := models.NewValidator()
		var errors models.ValidationErrors
		
		if err := validator.ValidateRequired(req.Token, "token"); err != nil {
			errors.AddError(*err)
		}
		
		return errors
	}) {
		return
	}

	// 既存のトークンを検証してリフレッシュ
	newToken, claims, err := h.authService.RefreshToken(req.Token)
	if err != nil {
		h.SendError(c, models.ErrTokenInvalid)
		return
	}

	// 成功レスポンス
	refreshResponse := &models.RefreshTokenResponse{
		Token:     newToken,
		ExpiresAt: models.NewDateTime(claims.ExpiresAt.Time),
	}

	h.SendSuccess(c, refreshResponse, "トークンのリフレッシュに成功しました")
}

// Logout はログアウトエンドポイントハンドラー
// @Summary ユーザーログアウト
// @Description ログアウト処理（JWTはステートレスのため、クライアント側でトークンを削除）
// @Tags auth
// @Produce json
// @Success 200 {object} models.DataResponse[interface{}] "ログアウト成功"
// @Router /api/v1/auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	// JWTはステートレスなので、サーバー側では特別な処理は不要
	// クライアント側でトークンを削除することでログアウトが完了する
	h.SendSuccess(c, nil, "ログアウトしました。クライアント側でトークンを削除してください")
}

// GetProfile は現在のユーザー情報を取得するエンドポイントハンドラー
// @Summary ユーザープロフィール取得
// @Description 現在認証されているユーザーの情報を取得する
// @Tags auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.DataResponse[models.UserProfileResponse] "ユーザー情報"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Router /api/v1/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	// ミドルウェアで設定されたユーザー情報を取得
	userID, exists := h.GetUserID(c)
	if !exists {
		h.SendUnauthorized(c, "認証情報が見つかりません")
		return
	}

	username, _ := h.GetUsername(c)
	role, _ := h.GetUserRole(c)

	// ユーザー情報を返す
	profileResponse := &models.UserProfileResponse{
		UserID:   userID,
		Username: username,
		Role:     role,
	}

	h.SendSuccess(c, profileResponse, "ユーザー情報を取得しました")
}

// ValidateToken はトークン検証エンドポイントハンドラー
// @Summary トークン検証
// @Description JWTトークンの有効性を検証する（POSTリクエストボディまたはGETのAuthorizationヘッダーから）
// @Tags auth
// @Accept json
// @Produce json
// @Param request body models.RefreshTokenRequest false "検証するトークン（POSTの場合）"
// @Param Authorization header string false "Bearer トークン（GETの場合）"
// @Success 200 {object} models.DataResponse[models.TokenValidationResponse] "検証成功"
// @Failure 400 {object} models.ValidationErrorResponse "バリデーションエラー"
// @Failure 401 {object} models.ErrorResponse "認証エラー"
// @Router /api/v1/auth/validate [post]
// @Router /api/v1/auth/validate [get]
func (h *AuthHandler) ValidateToken(c *gin.Context) {
	var token string

	// リクエストメソッドに応じてトークンを取得
	if c.Request.Method == "POST" {
		// POSTリクエストの場合：リクエストボディから取得
		var req models.RefreshTokenRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			h.SendBindingError(c, err)
			return
		}
		
		if err := req.Validate(); err != nil {
			h.SendErrorWithCode(c, models.ErrorValidationInvalidFormat, err.Error(), http.StatusBadRequest)
			return
		}
		
		token = req.Token
	} else {
		// GETリクエストの場合：Authorizationヘッダーから取得
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			h.SendErrorWithCode(c, models.ErrorValidationRequiredField, "Authorizationヘッダーが必要です", http.StatusBadRequest)
			return
		}

		// "Bearer "プレフィックスを削除
		if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
			token = authHeader[7:]
		} else {
			h.SendErrorWithCode(c, models.ErrorValidationInvalidFormat, "無効なAuthorizationヘッダー形式です", http.StatusBadRequest)
			return
		}
	}

	// 入力値の検証
	if strings.TrimSpace(token) == "" {
		h.SendErrorWithCode(c, models.ErrorValidationRequiredField, "トークンは必須です", http.StatusBadRequest)
		return
	}

	// トークンを検証
	claims, err := h.authService.ValidateToken(token)
	if err != nil {
		h.SendError(c, models.ErrTokenInvalid)
		return
	}

	// 検証成功レスポンス
	validationResponse := &models.TokenValidationResponse{
		Valid:     true,
		UserID:    claims.UserID,
		Username:  claims.Username,
		Role:      claims.Role,
		ExpiresAt: models.NewDateTime(claims.ExpiresAt.Time),
	}

	h.SendSuccess(c, validationResponse, "トークンは有効です")
}
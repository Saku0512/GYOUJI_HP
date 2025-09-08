package handler

import (
	"net/http"
	"strconv"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// BaseHandler は全てのハンドラーで共通して使用される基底ハンドラー
// 統一されたレスポンス送信メソッドを提供する
type BaseHandler struct {
	validator *validator.Validate
}

// NewBaseHandler は新しいBaseHandlerを作成する
func NewBaseHandler() *BaseHandler {
	return &BaseHandler{
		validator: validator.New(),
	}
}

// SendSuccess は成功レスポンスを送信する
// data: レスポンスデータ
// message: 成功メッセージ
// statusCode: HTTPステータスコード（省略時は200）
func (h *BaseHandler) SendSuccess(c *gin.Context, data interface{}, message string, statusCode ...int) {
	code := http.StatusOK
	if len(statusCode) > 0 {
		code = statusCode[0]
	}

	response := models.NewDataResponse(data, message, code)
	
	// リクエストIDが設定されている場合は追加
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			response.SetRequestID(id)
		}
	}

	c.JSON(code, response)
}

// SendError はエラーレスポンスを送信する
// apiError: APIErrorオブジェクト
func (h *BaseHandler) SendError(c *gin.Context, apiError *models.APIError) {
	response := models.NewErrorResponseUnified(apiError.Code, apiError.Message, apiError.StatusCode)
	
	// リクエストIDが設定されている場合は追加
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			response.SetRequestID(id)
		}
	}

	c.JSON(apiError.StatusCode, response)
}

// SendErrorWithCode はエラーコードとメッセージでエラーレスポンスを送信する
// errorCode: エラーコード
// message: エラーメッセージ
// statusCode: HTTPステータスコード
func (h *BaseHandler) SendErrorWithCode(c *gin.Context, errorCode string, message string, statusCode int) {
	apiError := models.NewAPIError(errorCode, message, statusCode)
	h.SendError(c, apiError)
}

// SendValidationError はバリデーションエラーレスポンスを送信する
// message: 全体的なエラーメッセージ
// details: 詳細なバリデーションエラー情報
func (h *BaseHandler) SendValidationError(c *gin.Context, message string, details []models.ValidationErrorDetail) {
	response := models.NewValidationErrorResponse(message, details)
	
	// リクエストIDが設定されている場合は追加
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			response.SetRequestID(id)
		}
	}

	c.JSON(http.StatusBadRequest, response)
}

// SendValidationErrors は統一されたValidationErrorsからレスポンスを送信する
func (h *BaseHandler) SendValidationErrors(c *gin.Context, errors models.ValidationErrors) {
	if !errors.HasErrors() {
		return
	}
	
	details := errors.ToValidationErrorDetails()
	h.SendValidationError(c, "入力データが無効です", details)
}

// ValidateRequest はリクエスト構造体のバリデーションを実行し、エラーがあればレスポンスを送信する
func (h *BaseHandler) ValidateRequest(c *gin.Context, validator func() models.ValidationErrors) bool {
	errors := validator()
	if errors.HasErrors() {
		h.SendValidationErrors(c, errors)
		return false
	}
	return true
}

// SendBindingError はリクエストバインディングエラーを処理してレスポンスを送信する
// err: バインディングエラー
func (h *BaseHandler) SendBindingError(c *gin.Context, err error) {
	var details []models.ValidationErrorDetail

	// validator.ValidationErrorsの場合は詳細情報を抽出
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, fieldError := range validationErrors {
			detail := models.ValidationErrorDetail{
				Field:   fieldError.Field(),
				Message: h.getValidationErrorMessage(fieldError),
				Value:   fieldError.Value().(string),
			}
			details = append(details, detail)
		}
	} else {
		// その他のバインディングエラーの場合
		details = append(details, models.ValidationErrorDetail{
			Field:   "request",
			Message: "リクエスト形式が無効です",
			Value:   "",
		})
	}

	h.SendValidationError(c, "入力データが無効です", details)
}

// SendUnauthorized は認証エラーレスポンスを送信する
// message: エラーメッセージ（省略時はデフォルトメッセージ）
func (h *BaseHandler) SendUnauthorized(c *gin.Context, message ...string) {
	msg := "認証が必要です"
	if len(message) > 0 {
		msg = message[0]
	}
	h.SendError(c, models.NewAPIError(models.ErrorAuthUnauthorized, msg, http.StatusUnauthorized))
}

// SendForbidden は認可エラーレスポンスを送信する
// message: エラーメッセージ（省略時はデフォルトメッセージ）
func (h *BaseHandler) SendForbidden(c *gin.Context, message ...string) {
	msg := "アクセス権限がありません"
	if len(message) > 0 {
		msg = message[0]
	}
	h.SendError(c, models.NewAPIError(models.ErrorAuthForbidden, msg, http.StatusForbidden))
}

// SendNotFound はリソースが見つからないエラーレスポンスを送信する
// message: エラーメッセージ（省略時はデフォルトメッセージ）
func (h *BaseHandler) SendNotFound(c *gin.Context, message ...string) {
	msg := "指定されたリソースが見つかりません"
	if len(message) > 0 {
		msg = message[0]
	}
	h.SendError(c, models.NewAPIError(models.ErrorResourceNotFound, msg, http.StatusNotFound))
}

// SendInternalServerError はサーバーエラーレスポンスを送信する
// message: エラーメッセージ（省略時はデフォルトメッセージ）
func (h *BaseHandler) SendInternalServerError(c *gin.Context, message ...string) {
	msg := "内部サーバーエラーが発生しました"
	if len(message) > 0 {
		msg = message[0]
	}
	h.SendError(c, models.NewAPIError(models.ErrorSystemUnknownError, msg, http.StatusInternalServerError))
}

// GetUserID はコンテキストからユーザーIDを取得する
func (h *BaseHandler) GetUserID(c *gin.Context) (int, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	
	switch v := userID.(type) {
	case int:
		return v, true
	case string:
		if id, err := strconv.Atoi(v); err == nil {
			return id, true
		}
	}
	
	return 0, false
}

// GetUsername はコンテキストからユーザー名を取得する
func (h *BaseHandler) GetUsername(c *gin.Context) (string, bool) {
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
func (h *BaseHandler) GetUserRole(c *gin.Context) (string, bool) {
	role, exists := c.Get("role")
	if !exists {
		return "", false
	}
	
	if r, ok := role.(string); ok {
		return r, true
	}
	
	return "", false
}

// ValidateStruct は構造体のバリデーションを実行する
func (h *BaseHandler) ValidateStruct(s interface{}) error {
	return h.validator.Struct(s)
}

// getValidationErrorMessage はバリデーションエラーから適切なメッセージを生成する
func (h *BaseHandler) getValidationErrorMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "必須項目です"
	case "email":
		return "有効なメールアドレスを入力してください"
	case "min":
		return "最小長は" + fe.Param() + "文字です"
	case "max":
		return "最大長は" + fe.Param() + "文字です"
	case "len":
		return "長さは" + fe.Param() + "文字である必要があります"
	case "numeric":
		return "数値を入力してください"
	case "alpha":
		return "英字のみ入力可能です"
	case "alphanum":
		return "英数字のみ入力可能です"
	case "oneof":
		return "許可された値のいずれかを選択してください: " + fe.Param()
	default:
		return "入力値が無効です"
	}
}
package errors

import (
	"fmt"
	"net/http"
)

// ErrorType はエラーの種類を表す
type ErrorType string

const (
	// ValidationError は入力検証エラー
	ValidationError ErrorType = "validation_error"
	// AuthenticationError は認証エラー
	AuthenticationError ErrorType = "authentication_error"
	// AuthorizationError は認可エラー
	AuthorizationError ErrorType = "authorization_error"
	// NotFoundError はリソースが見つからないエラー
	NotFoundError ErrorType = "not_found_error"
	// ConflictError はリソースの競合エラー
	ConflictError ErrorType = "conflict_error"
	// DatabaseError はデータベース関連エラー
	DatabaseError ErrorType = "database_error"
	// BusinessLogicError はビジネスロジックエラー
	BusinessLogicError ErrorType = "business_logic_error"
	// InternalError は内部サーバーエラー
	InternalError ErrorType = "internal_error"
)

// AppError はアプリケーション固有のエラー構造体
type AppError struct {
	Type       ErrorType              `json:"type"`
	Message    string                 `json:"message"`
	Details    string                 `json:"details,omitempty"`
	StatusCode int                    `json:"status_code"`
	Fields     map[string]interface{} `json:"fields,omitempty"`
	Cause      error                  `json:"-"`
}

// Error はerrorインターフェースを実装
func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Type, e.Message)
}

// Unwrap は元のエラーを返す
func (e *AppError) Unwrap() error {
	return e.Cause
}

// WithField はフィールドを追加したエラーを返す
func (e *AppError) WithField(key string, value interface{}) *AppError {
	if e.Fields == nil {
		e.Fields = make(map[string]interface{})
	}
	e.Fields[key] = value
	return e
}

// WithDetails は詳細情報を追加したエラーを返す
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// WithCause は原因エラーを追加したエラーを返す
func (e *AppError) WithCause(cause error) *AppError {
	e.Cause = cause
	return e
}

// NewValidationError は検証エラーを作成
func NewValidationError(message string) *AppError {
	return &AppError{
		Type:       ValidationError,
		Message:    message,
		StatusCode: http.StatusBadRequest,
	}
}

// NewAuthenticationError は認証エラーを作成
func NewAuthenticationError(message string) *AppError {
	return &AppError{
		Type:       AuthenticationError,
		Message:    message,
		StatusCode: http.StatusUnauthorized,
	}
}

// NewAuthorizationError は認可エラーを作成
func NewAuthorizationError(message string) *AppError {
	return &AppError{
		Type:       AuthorizationError,
		Message:    message,
		StatusCode: http.StatusForbidden,
	}
}

// NewNotFoundError はリソース未発見エラーを作成
func NewNotFoundError(resource string) *AppError {
	return &AppError{
		Type:       NotFoundError,
		Message:    fmt.Sprintf("%sが見つかりません", resource),
		StatusCode: http.StatusNotFound,
	}
}

// NewConflictError は競合エラーを作成
func NewConflictError(message string) *AppError {
	return &AppError{
		Type:       ConflictError,
		Message:    message,
		StatusCode: http.StatusConflict,
	}
}

// NewDatabaseError はデータベースエラーを作成
func NewDatabaseError(message string, cause error) *AppError {
	return &AppError{
		Type:       DatabaseError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      cause,
	}
}

// NewBusinessLogicError はビジネスロジックエラーを作成
func NewBusinessLogicError(message string) *AppError {
	return &AppError{
		Type:       BusinessLogicError,
		Message:    message,
		StatusCode: http.StatusUnprocessableEntity,
	}
}

// NewInternalError は内部エラーを作成
func NewInternalError(message string, cause error) *AppError {
	return &AppError{
		Type:       InternalError,
		Message:    message,
		StatusCode: http.StatusInternalServerError,
		Cause:      cause,
	}
}

// ErrorResponse はAPIエラーレスポンスの構造体
type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Details string                 `json:"details,omitempty"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// ToErrorResponse はAppErrorをErrorResponseに変換
func (e *AppError) ToErrorResponse() *ErrorResponse {
	return &ErrorResponse{
		Error:   string(e.Type),
		Message: e.Message,
		Code:    e.StatusCode,
		Details: e.Details,
		Fields:  e.Fields,
	}
}

// IsType は指定されたエラータイプかどうかを判定
func IsType(err error, errorType ErrorType) bool {
	if appErr, ok := err.(*AppError); ok {
		return appErr.Type == errorType
	}
	return false
}

// GetStatusCode はエラーからHTTPステータスコードを取得
func GetStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
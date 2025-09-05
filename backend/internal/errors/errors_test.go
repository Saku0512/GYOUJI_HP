package errors

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppError_Error(t *testing.T) {
	tests := []struct {
		name     string
		appError *AppError
		want     string
	}{
		{
			name: "原因エラーなし",
			appError: &AppError{
				Type:    ValidationError,
				Message: "検証エラー",
			},
			want: "validation_error: 検証エラー",
		},
		{
			name: "原因エラーあり",
			appError: &AppError{
				Type:    DatabaseError,
				Message: "データベースエラー",
				Cause:   fmt.Errorf("connection failed"),
			},
			want: "database_error: データベースエラー (caused by: connection failed)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appError.Error()
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestAppError_Unwrap(t *testing.T) {
	cause := fmt.Errorf("original error")
	appErr := &AppError{
		Type:    InternalError,
		Message: "内部エラー",
		Cause:   cause,
	}

	unwrapped := appErr.Unwrap()
	assert.Equal(t, cause, unwrapped)
}

func TestAppError_WithField(t *testing.T) {
	appErr := &AppError{
		Type:    ValidationError,
		Message: "検証エラー",
	}

	result := appErr.WithField("field", "value")
	
	assert.Equal(t, appErr, result) // 同じインスタンスが返される
	assert.Equal(t, "value", appErr.Fields["field"])
}

func TestAppError_WithDetails(t *testing.T) {
	appErr := &AppError{
		Type:    ValidationError,
		Message: "検証エラー",
	}

	details := "詳細な説明"
	result := appErr.WithDetails(details)
	
	assert.Equal(t, appErr, result)
	assert.Equal(t, details, appErr.Details)
}

func TestAppError_WithCause(t *testing.T) {
	appErr := &AppError{
		Type:    InternalError,
		Message: "内部エラー",
	}

	cause := fmt.Errorf("original error")
	result := appErr.WithCause(cause)
	
	assert.Equal(t, appErr, result)
	assert.Equal(t, cause, appErr.Cause)
}

func TestNewValidationError(t *testing.T) {
	message := "検証に失敗しました"
	err := NewValidationError(message)

	assert.Equal(t, ValidationError, err.Type)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusBadRequest, err.StatusCode)
}

func TestNewAuthenticationError(t *testing.T) {
	message := "認証に失敗しました"
	err := NewAuthenticationError(message)

	assert.Equal(t, AuthenticationError, err.Type)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusUnauthorized, err.StatusCode)
}

func TestNewAuthorizationError(t *testing.T) {
	message := "権限がありません"
	err := NewAuthorizationError(message)

	assert.Equal(t, AuthorizationError, err.Type)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusForbidden, err.StatusCode)
}

func TestNewNotFoundError(t *testing.T) {
	resource := "ユーザー"
	err := NewNotFoundError(resource)

	assert.Equal(t, NotFoundError, err.Type)
	assert.Equal(t, "ユーザーが見つかりません", err.Message)
	assert.Equal(t, http.StatusNotFound, err.StatusCode)
}

func TestNewConflictError(t *testing.T) {
	message := "データが競合しています"
	err := NewConflictError(message)

	assert.Equal(t, ConflictError, err.Type)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusConflict, err.StatusCode)
}

func TestNewDatabaseError(t *testing.T) {
	message := "データベースエラー"
	cause := fmt.Errorf("connection failed")
	err := NewDatabaseError(message, cause)

	assert.Equal(t, DatabaseError, err.Type)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, cause, err.Cause)
}

func TestNewBusinessLogicError(t *testing.T) {
	message := "ビジネスルール違反"
	err := NewBusinessLogicError(message)

	assert.Equal(t, BusinessLogicError, err.Type)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusUnprocessableEntity, err.StatusCode)
}

func TestNewInternalError(t *testing.T) {
	message := "内部エラー"
	cause := fmt.Errorf("unexpected error")
	err := NewInternalError(message, cause)

	assert.Equal(t, InternalError, err.Type)
	assert.Equal(t, message, err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, cause, err.Cause)
}

func TestAppError_ToErrorResponse(t *testing.T) {
	appErr := &AppError{
		Type:       ValidationError,
		Message:    "検証エラー",
		StatusCode: http.StatusBadRequest,
		Details:    "詳細情報",
		Fields:     map[string]interface{}{"field": "value"},
	}

	response := appErr.ToErrorResponse()

	assert.Equal(t, "validation_error", response.Error)
	assert.Equal(t, "検証エラー", response.Message)
	assert.Equal(t, http.StatusBadRequest, response.Code)
	assert.Equal(t, "詳細情報", response.Details)
	assert.Equal(t, map[string]interface{}{"field": "value"}, response.Fields)
}

func TestIsType(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		errorType ErrorType
		want      bool
	}{
		{
			name:      "AppErrorで一致",
			err:       NewValidationError("test"),
			errorType: ValidationError,
			want:      true,
		},
		{
			name:      "AppErrorで不一致",
			err:       NewValidationError("test"),
			errorType: AuthenticationError,
			want:      false,
		},
		{
			name:      "通常のエラー",
			err:       fmt.Errorf("normal error"),
			errorType: ValidationError,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsType(tt.err, tt.errorType)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{
			name: "AppError",
			err:  NewValidationError("test"),
			want: http.StatusBadRequest,
		},
		{
			name: "通常のエラー",
			err:  fmt.Errorf("normal error"),
			want: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStatusCode(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}
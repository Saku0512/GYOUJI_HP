package models

import (
	"time"
)

// APIResponse は統一されたAPIレスポンス構造体（後方互換性のため維持）
// 新しいコードではDataResponse[T]やListResponse[T]を使用することを推奨
type APIResponse struct {
	Success   bool        `json:"success"`             // 成功フラグ
	Data      interface{} `json:"data,omitempty"`      // レスポンスデータ（成功時のみ）
	Error     string      `json:"error,omitempty"`     // エラーコード（エラー時のみ）
	Message   string      `json:"message"`             // メッセージ
	Code      int         `json:"code"`                // HTTPステータスコード
	Timestamp string      `json:"timestamp"`           // タイムスタンプ（ISO 8601形式）
	RequestID string      `json:"request_id,omitempty"` // リクエストID（追跡用）
}

// NewSuccessResponse は成功レスポンスを作成する（後方互換性のため維持）
func NewSuccessResponse(data interface{}, message string, code int) *APIResponse {
	return &APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Code:      code,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewErrorResponse はエラーレスポンスを作成する（後方互換性のため維持）
func NewErrorResponse(errorCode string, message string, statusCode int) *APIResponse {
	return &APIResponse{
		Success:   false,
		Error:     errorCode,
		Message:   message,
		Code:      statusCode,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// SetRequestID はリクエストIDを設定する（後方互換性のため維持）
func (r *APIResponse) SetRequestID(requestID string) *APIResponse {
	r.RequestID = requestID
	return r
}

// 新しい統一レスポンス作成関数

// NewDataResponse は単一データレスポンスを作成する
func NewDataResponse[T any](data T, message string, code int) *DataResponse[T] {
	return &DataResponse[T]{
		BaseResponse: BaseResponse{
			Success:   true,
			Message:   message,
			Code:      code,
			Timestamp: Now().String(),
		},
		Data: data,
	}
}

// NewListResponse はリストレスポンスを作成する
func NewListResponse[T any](data []T, message string, code int) *ListResponse[T] {
	return &ListResponse[T]{
		BaseResponse: BaseResponse{
			Success:   true,
			Message:   message,
			Code:      code,
			Timestamp: Now().String(),
		},
		Data:  data,
		Count: len(data),
	}
}

// NewPaginatedResponse はページネーション付きレスポンスを作成する
func NewPaginatedResponse[T any](data []T, pagination *PaginationResponse, message string, code int) *PaginatedResponse[T] {
	return &PaginatedResponse[T]{
		BaseResponse: BaseResponse{
			Success:   true,
			Message:   message,
			Code:      code,
			Timestamp: Now().String(),
		},
		Data:       data,
		Pagination: pagination,
	}
}

// NewErrorResponseUnified は統一エラーレスポンスを作成する
func NewErrorResponseUnified(errorCode string, message string, statusCode int) *ErrorResponse {
	return &ErrorResponse{
		BaseResponse: BaseResponse{
			Success:   false,
			Message:   message,
			Code:      statusCode,
			Timestamp: Now().String(),
		},
		Error: errorCode,
	}
}

// ValidationErrorDetail はバリデーションエラーの詳細情報（後方互換性のため維持）
type ValidationErrorDetail struct {
	Field   string `json:"field"`   // エラーが発生したフィールド名
	Message string `json:"message"` // エラーメッセージ
	Value   string `json:"value"`   // 入力された値
}

// NewValidationErrorResponse はバリデーションエラーレスポンスを作成する（後方互換性のため維持）
func NewValidationErrorResponse(message string, details []ValidationErrorDetail) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		BaseResponse: BaseResponse{
			Success:   false,
			Message:   message,
			Code:      400,
			Timestamp: Now().String(),
		},
		Error:   "VALIDATION_ERROR",
		Details: details,
	}
}
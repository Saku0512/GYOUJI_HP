package models

import (
	"time"
)

// APIResponse は統一されたAPIレスポンス構造体
// 全てのAPIエンドポイントで一貫したレスポンス形式を提供する
type APIResponse struct {
	Success   bool        `json:"success"`             // 成功フラグ
	Data      interface{} `json:"data,omitempty"`      // レスポンスデータ（成功時のみ）
	Error     string      `json:"error,omitempty"`     // エラーコード（エラー時のみ）
	Message   string      `json:"message"`             // メッセージ
	Code      int         `json:"code"`                // HTTPステータスコード
	Timestamp string      `json:"timestamp"`           // タイムスタンプ（ISO 8601形式）
	RequestID string      `json:"request_id,omitempty"` // リクエストID（追跡用）
}

// NewSuccessResponse は成功レスポンスを作成する
func NewSuccessResponse(data interface{}, message string, code int) *APIResponse {
	return &APIResponse{
		Success:   true,
		Data:      data,
		Message:   message,
		Code:      code,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// NewErrorResponse はエラーレスポンスを作成する
func NewErrorResponse(errorCode string, message string, statusCode int) *APIResponse {
	return &APIResponse{
		Success:   false,
		Error:     errorCode,
		Message:   message,
		Code:      statusCode,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// SetRequestID はリクエストIDを設定する
func (r *APIResponse) SetRequestID(requestID string) *APIResponse {
	r.RequestID = requestID
	return r
}

// ValidationErrorDetail はバリデーションエラーの詳細情報
type ValidationErrorDetail struct {
	Field   string `json:"field"`   // エラーが発生したフィールド名
	Message string `json:"message"` // エラーメッセージ
	Value   string `json:"value"`   // 入力された値
}

// ValidationErrorResponse はバリデーションエラー専用のレスポンス構造体
type ValidationErrorResponse struct {
	Success   bool                    `json:"success"`             // 成功フラグ（常にfalse）
	Error     string                  `json:"error"`               // エラーコード
	Message   string                  `json:"message"`             // 全体的なエラーメッセージ
	Details   []ValidationErrorDetail `json:"details"`             // 詳細なバリデーションエラー情報
	Code      int                     `json:"code"`                // HTTPステータスコード
	Timestamp string                  `json:"timestamp"`           // タイムスタンプ
	RequestID string                  `json:"request_id,omitempty"` // リクエストID
}

// NewValidationErrorResponse はバリデーションエラーレスポンスを作成する
func NewValidationErrorResponse(message string, details []ValidationErrorDetail) *ValidationErrorResponse {
	return &ValidationErrorResponse{
		Success:   false,
		Error:     "VALIDATION_ERROR",
		Message:   message,
		Details:   details,
		Code:      400,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
}

// SetRequestID はリクエストIDを設定する
func (r *ValidationErrorResponse) SetRequestID(requestID string) *ValidationErrorResponse {
	r.RequestID = requestID
	return r
}
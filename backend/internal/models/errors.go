package models

// APIError は統一されたAPIエラー型
type APIError struct {
	Code       string                 `json:"code"`                 // エラーコード
	Message    string                 `json:"message"`              // エラーメッセージ
	StatusCode int                    `json:"status_code"`          // HTTPステータスコード
	Details    map[string]interface{} `json:"details,omitempty"`    // 追加の詳細情報
}

// Error はerrorインターフェースを実装する
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError は新しいAPIErrorを作成する
func NewAPIError(code string, message string, statusCode int) *APIError {
	return &APIError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Details:    make(map[string]interface{}),
	}
}

// WithDetails は詳細情報を追加する
func (e *APIError) WithDetails(key string, value interface{}) *APIError {
	if e.Details == nil {
		e.Details = make(map[string]interface{})
	}
	e.Details[key] = value
	return e
}

// 認証関連エラーコード
const (
	// AUTH_* - 認証関連エラー
	ErrorAuthInvalidCredentials = "AUTH_INVALID_CREDENTIALS" // 認証情報が無効
	ErrorAuthTokenExpired       = "AUTH_TOKEN_EXPIRED"       // トークンが期限切れ
	ErrorAuthTokenInvalid       = "AUTH_TOKEN_INVALID"       // トークンが無効
	ErrorAuthUnauthorized       = "AUTH_UNAUTHORIZED"        // 認証が必要
	ErrorAuthForbidden          = "AUTH_FORBIDDEN"           // アクセス権限なし
)

// バリデーション関連エラーコード
const (
	// VALIDATION_* - バリデーション関連エラー
	ErrorValidationRequiredField = "VALIDATION_REQUIRED_FIELD" // 必須フィールドが未入力
	ErrorValidationInvalidFormat = "VALIDATION_INVALID_FORMAT" // 形式が無効
	ErrorValidationOutOfRange    = "VALIDATION_OUT_OF_RANGE"   // 値が範囲外
	ErrorValidationDuplicateValue = "VALIDATION_DUPLICATE_VALUE" // 重複した値
)

// リソース関連エラーコード
const (
	// RESOURCE_* - リソース関連エラー
	ErrorResourceNotFound      = "RESOURCE_NOT_FOUND"       // リソースが見つからない
	ErrorResourceAlreadyExists = "RESOURCE_ALREADY_EXISTS"  // リソースが既に存在
	ErrorResourceConflict      = "RESOURCE_CONFLICT"        // リソースの競合
)

// ビジネスロジック関連エラーコード
const (
	// BUSINESS_* - ビジネスロジック関連エラー
	ErrorBusinessTournamentCompleted    = "BUSINESS_TOURNAMENT_COMPLETED"     // トーナメントが既に完了
	ErrorBusinessMatchAlreadyCompleted  = "BUSINESS_MATCH_ALREADY_COMPLETED"  // 試合が既に完了
	ErrorBusinessInvalidMatchResult     = "BUSINESS_INVALID_MATCH_RESULT"     // 無効な試合結果
)

// システム関連エラーコード
const (
	// SYSTEM_* - システム関連エラー
	ErrorSystemDatabaseError = "SYSTEM_DATABASE_ERROR" // データベースエラー
	ErrorSystemNetworkError  = "SYSTEM_NETWORK_ERROR"  // ネットワークエラー
	ErrorSystemTimeout       = "SYSTEM_TIMEOUT"        // タイムアウト
	ErrorSystemUnknownError  = "SYSTEM_UNKNOWN_ERROR"  // 不明なエラー
)

// 事前定義されたAPIエラー

// 認証関連エラー
var (
	ErrInvalidCredentials = NewAPIError(ErrorAuthInvalidCredentials, "認証情報が無効です", 401)
	ErrTokenExpired       = NewAPIError(ErrorAuthTokenExpired, "トークンが期限切れです", 401)
	ErrTokenInvalid       = NewAPIError(ErrorAuthTokenInvalid, "無効なトークンです", 401)
	ErrUnauthorized       = NewAPIError(ErrorAuthUnauthorized, "認証が必要です", 401)
	ErrForbidden          = NewAPIError(ErrorAuthForbidden, "アクセス権限がありません", 403)
)

// バリデーション関連エラー
var (
	ErrRequiredField    = NewAPIError(ErrorValidationRequiredField, "必須フィールドが未入力です", 400)
	ErrInvalidFormat    = NewAPIError(ErrorValidationInvalidFormat, "入力形式が無効です", 400)
	ErrOutOfRange       = NewAPIError(ErrorValidationOutOfRange, "値が許可された範囲外です", 400)
	ErrDuplicateValue   = NewAPIError(ErrorValidationDuplicateValue, "重複した値が入力されています", 400)
)

// リソース関連エラー
var (
	ErrResourceNotFound      = NewAPIError(ErrorResourceNotFound, "指定されたリソースが見つかりません", 404)
	ErrResourceAlreadyExists = NewAPIError(ErrorResourceAlreadyExists, "リソースが既に存在します", 409)
	ErrResourceConflict      = NewAPIError(ErrorResourceConflict, "リソースの競合が発生しました", 409)
)

// ビジネスロジック関連エラー
var (
	ErrTournamentCompleted   = NewAPIError(ErrorBusinessTournamentCompleted, "トーナメントは既に完了しています", 400)
	ErrMatchAlreadyCompleted = NewAPIError(ErrorBusinessMatchAlreadyCompleted, "試合は既に完了しています", 400)
	ErrInvalidMatchResult    = NewAPIError(ErrorBusinessInvalidMatchResult, "無効な試合結果です", 400)
)

// システム関連エラー
var (
	ErrDatabaseError = NewAPIError(ErrorSystemDatabaseError, "データベースエラーが発生しました", 500)
	ErrNetworkError  = NewAPIError(ErrorSystemNetworkError, "ネットワークエラーが発生しました", 500)
	ErrTimeout       = NewAPIError(ErrorSystemTimeout, "処理がタイムアウトしました", 500)
	ErrUnknownError  = NewAPIError(ErrorSystemUnknownError, "予期しないエラーが発生しました", 500)
)
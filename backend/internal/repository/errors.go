package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// ErrorType はリポジトリエラーの種類を表す
type ErrorType string

const (
	// ErrTypeConnection はデータベース接続関連のエラー
	ErrTypeConnection ErrorType = "connection"
	// ErrTypeQuery はクエリ実行関連のエラー
	ErrTypeQuery ErrorType = "query"
	// ErrTypeTransaction はトランザクション関連のエラー
	ErrTypeTransaction ErrorType = "transaction"
	// ErrTypeValidation はデータ検証関連のエラー
	ErrTypeValidation ErrorType = "validation"
	// ErrTypeNotFound はデータが見つからない場合のエラー
	ErrTypeNotFound ErrorType = "not_found"
	// ErrTypeDuplicate は重複データ関連のエラー
	ErrTypeDuplicate ErrorType = "duplicate"
	// ErrTypeConstraint は制約違反関連のエラー
	ErrTypeConstraint ErrorType = "constraint"
)

// RepositoryError はリポジトリ層で発生するエラーを表す
type RepositoryError struct {
	Type    ErrorType
	Message string
	Err     error
}

// Error はerrorインターフェースの実装
func (e *RepositoryError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Err)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap は元のエラーを返す
func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError は新しいRepositoryErrorを作成する
func NewRepositoryError(errType ErrorType, message string, err error) *RepositoryError {
	return &RepositoryError{
		Type:    errType,
		Message: message,
		Err:     err,
	}
}

// IsRepositoryError は指定されたエラーがRepositoryErrorかどうかを判定する
func IsRepositoryError(err error) bool {
	var repoErr *RepositoryError
	return errors.As(err, &repoErr)
}

// GetRepositoryErrorType はRepositoryErrorの種類を取得する
func GetRepositoryErrorType(err error) ErrorType {
	var repoErr *RepositoryError
	if errors.As(err, &repoErr) {
		return repoErr.Type
	}
	return ""
}

// HandleSQLError はSQLエラーを適切なRepositoryErrorに変換する
func HandleSQLError(err error, operation string) error {
	if err == nil {
		return nil
	}

	// sql.ErrNoRowsの場合
	if errors.Is(err, sql.ErrNoRows) {
		return NewRepositoryError(ErrTypeNotFound, fmt.Sprintf("%s: データが見つかりません", operation), err)
	}

	// MySQLエラーコードによる分類
	errStr := err.Error()
	
	// 重複キーエラー (MySQL Error 1062)
	if strings.Contains(errStr, "Error 1062") || strings.Contains(errStr, "Duplicate entry") {
		return NewRepositoryError(ErrTypeDuplicate, fmt.Sprintf("%s: 重複するデータが存在します", operation), err)
	}
	
	// 外部キー制約エラー (MySQL Error 1452)
	if strings.Contains(errStr, "Error 1452") || strings.Contains(errStr, "foreign key constraint") {
		return NewRepositoryError(ErrTypeConstraint, fmt.Sprintf("%s: 外部キー制約違反です", operation), err)
	}
	
	// NOT NULL制約エラー (MySQL Error 1048)
	if strings.Contains(errStr, "Error 1048") || strings.Contains(errStr, "cannot be null") {
		return NewRepositoryError(ErrTypeValidation, fmt.Sprintf("%s: 必須フィールドが空です", operation), err)
	}
	
	// データが長すぎるエラー (MySQL Error 1406)
	if strings.Contains(errStr, "Error 1406") || strings.Contains(errStr, "Data too long") {
		return NewRepositoryError(ErrTypeValidation, fmt.Sprintf("%s: データが長すぎます", operation), err)
	}
	
	// 接続関連エラー
	if strings.Contains(errStr, "connection") || strings.Contains(errStr, "timeout") {
		return NewRepositoryError(ErrTypeConnection, fmt.Sprintf("%s: データベース接続エラー", operation), err)
	}
	
	// その他のクエリエラー
	return NewRepositoryError(ErrTypeQuery, fmt.Sprintf("%s: クエリ実行エラー", operation), err)
}

// ValidateNotNil は値がnilでないことを検証する
func ValidateNotNil(value interface{}, fieldName string) error {
	if value == nil {
		return NewRepositoryError(ErrTypeValidation, fmt.Sprintf("%s は必須です", fieldName), nil)
	}
	return nil
}

// ValidateNotEmpty は文字列が空でないことを検証する
func ValidateNotEmpty(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return NewRepositoryError(ErrTypeValidation, fmt.Sprintf("%s は空にできません", fieldName), nil)
	}
	return nil
}

// ValidatePositiveInt は整数が正の値であることを検証する
func ValidatePositiveInt(value int, fieldName string) error {
	if value <= 0 {
		return NewRepositoryError(ErrTypeValidation, fmt.Sprintf("%s は正の値である必要があります", fieldName), nil)
	}
	return nil
}

// ValidateMaxLength は文字列の最大長を検証する
func ValidateMaxLength(value string, maxLength int, fieldName string) error {
	if len(value) > maxLength {
		return NewRepositoryError(ErrTypeValidation, 
			fmt.Sprintf("%s は%d文字以下である必要があります", fieldName, maxLength), nil)
	}
	return nil
}

// ExecuteWithRetry は指定された回数だけリトライしながら操作を実行する
func ExecuteWithRetry(operation func() error, maxRetries int, operationName string) error {
	var lastErr error
	
	for i := 0; i <= maxRetries; i++ {
		err := operation()
		if err == nil {
			return nil
		}
		
		lastErr = err
		
		// 接続エラーの場合のみリトライする
		if IsRepositoryError(err) && GetRepositoryErrorType(err) == ErrTypeConnection {
			if i < maxRetries {
				fmt.Printf("%s: リトライ %d/%d - %v\n", operationName, i+1, maxRetries, err)
				continue
			}
		} else {
			// 接続エラー以外はリトライしない
			break
		}
	}
	
	return NewRepositoryError(ErrTypeQuery, 
		fmt.Sprintf("%s: %d回のリトライ後も失敗しました", operationName, maxRetries), lastErr)
}

// LogError はエラーをログに出力する
func LogError(err error, context string) {
	if err == nil {
		return
	}
	
	if IsRepositoryError(err) {
		repoErr := err.(*RepositoryError)
		fmt.Printf("Repository Error [%s] %s: %s\n", repoErr.Type, context, repoErr.Message)
		if repoErr.Err != nil {
			fmt.Printf("Underlying error: %v\n", repoErr.Err)
		}
	} else {
		fmt.Printf("Error [%s]: %v\n", context, err)
	}
}
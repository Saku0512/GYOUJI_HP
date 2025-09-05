package errors

import (
	"database/sql"

	"backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
)

// ErrorHandlerMiddleware はエラーハンドリングミドルウェア
func ErrorHandlerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// エラーが発生した場合の処理
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			handleError(c, err)
		}
	}
}

// handleError はエラーを適切に処理してレスポンスを返す
func handleError(c *gin.Context, err error) {
	log := logger.GetLogger().WithContext(c.Request.Context())

	// AppErrorの場合
	if appErr, ok := err.(*AppError); ok {
		log.Error("アプリケーションエラーが発生しました",
			logger.String("error_type", string(appErr.Type)),
			logger.String("message", appErr.Message),
			logger.Int("status_code", appErr.StatusCode),
			logger.Err(appErr.Cause),
		)

		c.JSON(appErr.StatusCode, appErr.ToErrorResponse())
		return
	}

	// データベースエラーの処理
	if dbErr := handleDatabaseError(err); dbErr != nil {
		log.Error("データベースエラーが発生しました",
			logger.String("error_type", string(dbErr.Type)),
			logger.String("message", dbErr.Message),
			logger.Err(err),
		)

		c.JSON(dbErr.StatusCode, dbErr.ToErrorResponse())
		return
	}

	// その他の予期しないエラー
	log.Error("予期しないエラーが発生しました",
		logger.Err(err),
	)

	internalErr := NewInternalError("内部サーバーエラーが発生しました", err)
	c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse())
}

// handleDatabaseError はデータベース固有のエラーを処理
func handleDatabaseError(err error) *AppError {
	// MySQL固有のエラー処理
	if mysqlErr, ok := err.(*mysql.MySQLError); ok {
		switch mysqlErr.Number {
		case 1062: // Duplicate entry
			return NewConflictError("データが既に存在します")
		case 1452: // Foreign key constraint fails
			return NewValidationError("関連するデータが存在しません")
		case 1406: // Data too long
			return NewValidationError("データが長すぎます")
		case 1048: // Column cannot be null
			return NewValidationError("必須フィールドが入力されていません")
		default:
			return NewDatabaseError("データベースエラーが発生しました", err)
		}
	}

	// sql.ErrNoRowsの処理
	if err == sql.ErrNoRows {
		return NewNotFoundError("データ")
	}

	// その他のデータベース関連エラー
	if isDatabaseError(err) {
		return NewDatabaseError("データベース操作でエラーが発生しました", err)
	}

	return nil
}

// isDatabaseError はデータベース関連のエラーかどうかを判定
func isDatabaseError(err error) bool {
	// よくあるデータベースエラーのパターンをチェック
	errorMsg := err.Error()
	
	databaseErrorPatterns := []string{
		"connection refused",
		"connection reset",
		"connection timeout",
		"database is locked",
		"no such table",
		"syntax error",
		"constraint failed",
	}

	for _, pattern := range databaseErrorPatterns {
		if contains(errorMsg, pattern) {
			return true
		}
	}

	return false
}

// contains は文字列に部分文字列が含まれているかチェック
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || 
		(len(s) > len(substr) && 
			(s[:len(substr)] == substr || 
			 s[len(s)-len(substr):] == substr || 
			 containsSubstring(s, substr))))
}

// containsSubstring は文字列内の部分文字列を検索
func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// RecoveryMiddleware はパニックを捕捉してエラーに変換するミドルウェア
func RecoveryMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log := logger.GetLogger().WithContext(c.Request.Context())
				
				log.Error("パニックが発生しました",
					logger.Any("panic", r),
				)

				internalErr := NewInternalError("内部サーバーエラーが発生しました", nil)
				c.JSON(internalErr.StatusCode, internalErr.ToErrorResponse())
				c.Abort()
			}
		}()

		c.Next()
	}
}
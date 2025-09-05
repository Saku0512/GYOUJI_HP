package errors

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/logger"

	"github.com/gin-gonic/gin"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func init() {
	// テスト用にロガーを初期化
	logger.Init()
	gin.SetMode(gin.TestMode)
}

func TestErrorHandlerMiddleware_AppError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	
	router.GET("/test", func(c *gin.Context) {
		err := NewValidationError("テスト検証エラー")
		c.Error(err)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "validation_error", response.Error)
	assert.Equal(t, "テスト検証エラー", response.Message)
	assert.Equal(t, http.StatusBadRequest, response.Code)
}

func TestErrorHandlerMiddleware_DatabaseError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	
	router.GET("/test", func(c *gin.Context) {
		// sql.ErrNoRowsをシミュレート
		c.Error(sql.ErrNoRows)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "not_found_error", response.Error)
	assert.Equal(t, "データが見つかりません", response.Message)
}

func TestErrorHandlerMiddleware_MySQLError(t *testing.T) {
	tests := []struct {
		name           string
		mysqlError     *mysql.MySQLError
		expectedStatus int
		expectedType   string
	}{
		{
			name:           "重複エントリエラー",
			mysqlError:     &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"},
			expectedStatus: http.StatusConflict,
			expectedType:   "conflict_error",
		},
		{
			name:           "外部キー制約エラー",
			mysqlError:     &mysql.MySQLError{Number: 1452, Message: "Foreign key constraint fails"},
			expectedStatus: http.StatusBadRequest,
			expectedType:   "validation_error",
		},
		{
			name:           "データ長エラー",
			mysqlError:     &mysql.MySQLError{Number: 1406, Message: "Data too long"},
			expectedStatus: http.StatusBadRequest,
			expectedType:   "validation_error",
		},
		{
			name:           "NULL制約エラー",
			mysqlError:     &mysql.MySQLError{Number: 1048, Message: "Column cannot be null"},
			expectedStatus: http.StatusBadRequest,
			expectedType:   "validation_error",
		},
		{
			name:           "その他のMySQLエラー",
			mysqlError:     &mysql.MySQLError{Number: 9999, Message: "Unknown error"},
			expectedStatus: http.StatusInternalServerError,
			expectedType:   "database_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Use(ErrorHandlerMiddleware())
			
			router.GET("/test", func(c *gin.Context) {
				c.Error(tt.mysqlError)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response ErrorResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedType, response.Error)
		})
	}
}

func TestErrorHandlerMiddleware_UnexpectedError(t *testing.T) {
	router := gin.New()
	router.Use(ErrorHandlerMiddleware())
	
	router.GET("/test", func(c *gin.Context) {
		err := fmt.Errorf("予期しないエラー")
		c.Error(err)
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "internal_error", response.Error)
	assert.Equal(t, "内部サーバーエラーが発生しました", response.Message)
}

func TestRecoveryMiddleware(t *testing.T) {
	router := gin.New()
	router.Use(RecoveryMiddleware())
	
	router.GET("/panic", func(c *gin.Context) {
		panic("テストパニック")
	})

	req := httptest.NewRequest("GET", "/panic", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)

	var response ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "internal_error", response.Error)
	assert.Equal(t, "内部サーバーエラーが発生しました", response.Message)
}

func TestHandleDatabaseError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantNil     bool
		wantType    ErrorType
		wantStatus  int
	}{
		{
			name:       "sql.ErrNoRows",
			err:        sql.ErrNoRows,
			wantNil:    false,
			wantType:   NotFoundError,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "MySQL重複エラー",
			err:        &mysql.MySQLError{Number: 1062},
			wantNil:    false,
			wantType:   ConflictError,
			wantStatus: http.StatusConflict,
		},
		{
			name:    "非データベースエラー",
			err:     fmt.Errorf("通常のエラー"),
			wantNil: true,
		},
		{
			name:       "データベース接続エラー",
			err:        fmt.Errorf("connection refused"),
			wantNil:    false,
			wantType:   DatabaseError,
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := handleDatabaseError(tt.err)
			
			if tt.wantNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.Equal(t, tt.wantType, result.Type)
				assert.Equal(t, tt.wantStatus, result.StatusCode)
			}
		})
	}
}

func TestIsDatabaseError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "接続拒否エラー",
			err:  fmt.Errorf("connection refused"),
			want: true,
		},
		{
			name: "接続リセットエラー",
			err:  fmt.Errorf("connection reset"),
			want: true,
		},
		{
			name: "接続タイムアウトエラー",
			err:  fmt.Errorf("connection timeout"),
			want: true,
		},
		{
			name: "テーブル不存在エラー",
			err:  fmt.Errorf("no such table"),
			want: true,
		},
		{
			name: "構文エラー",
			err:  fmt.Errorf("syntax error"),
			want: true,
		},
		{
			name: "制約エラー",
			err:  fmt.Errorf("constraint failed"),
			want: true,
		},
		{
			name: "通常のエラー",
			err:  fmt.Errorf("normal error"),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isDatabaseError(tt.err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name   string
		s      string
		substr string
		want   bool
	}{
		{
			name:   "完全一致",
			s:      "test",
			substr: "test",
			want:   true,
		},
		{
			name:   "前方一致",
			s:      "testing",
			substr: "test",
			want:   true,
		},
		{
			name:   "後方一致",
			s:      "unittest",
			substr: "test",
			want:   true,
		},
		{
			name:   "中間一致",
			s:      "this is a test case",
			substr: "test",
			want:   true,
		},
		{
			name:   "不一致",
			s:      "example",
			substr: "test",
			want:   false,
		},
		{
			name:   "空文字列",
			s:      "test",
			substr: "",
			want:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := contains(tt.s, tt.substr)
			assert.Equal(t, tt.want, got)
		})
	}
}
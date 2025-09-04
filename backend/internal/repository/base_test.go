package repository

import (
	"database/sql"
	"testing"

	"backend/internal/database"
)

// TestNewBaseRepository はNewBaseRepositoryのテスト
func TestNewBaseRepository(t *testing.T) {
	tests := []struct {
		name   string
		db     *database.DB
		panics bool
	}{
		{
			name:   "正常なデータベース接続",
			db:     &database.DB{},
			panics: false,
		},
		{
			name:   "nilデータベース接続",
			db:     nil,
			panics: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if (r != nil) != tt.panics {
					t.Errorf("NewBaseRepository() panic = %v, want %v", r != nil, tt.panics)
				}
			}()

			if !tt.panics {
				repo := NewBaseRepository(tt.db)
				if repo == nil {
					t.Error("NewBaseRepository() returned nil")
				}
			} else {
				NewBaseRepository(tt.db)
			}
		})
	}
}

// TestBaseRepository_GetDB はGetDBメソッドのテスト（データベース接続不要）
func TestBaseRepository_GetDB(t *testing.T) {
	// データベース接続が必要ないテストのみ実行
	t.Skip("データベース接続が必要なため、統合テスト時に実行")
}

// TestValidateNotNil はValidateNotNilのテスト
func TestValidateNotNil(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		fieldName string
		wantErr   bool
	}{
		{
			name:      "非nil値",
			value:     "test",
			fieldName: "testField",
			wantErr:   false,
		},
		{
			name:      "nil値",
			value:     nil,
			fieldName: "testField",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotNil(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotNil() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !IsRepositoryError(err) {
				t.Error("ValidateNotNil() should return RepositoryError")
			}
		})
	}
}

// TestValidateNotEmpty はValidateNotEmptyのテスト
func TestValidateNotEmpty(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantErr   bool
	}{
		{
			name:      "非空文字列",
			value:     "test",
			fieldName: "testField",
			wantErr:   false,
		},
		{
			name:      "空文字列",
			value:     "",
			fieldName: "testField",
			wantErr:   true,
		},
		{
			name:      "空白のみの文字列",
			value:     "   ",
			fieldName: "testField",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotEmpty(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNotEmpty() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !IsRepositoryError(err) {
				t.Error("ValidateNotEmpty() should return RepositoryError")
			}
		})
	}
}

// TestValidatePositiveInt はValidatePositiveIntのテスト
func TestValidatePositiveInt(t *testing.T) {
	tests := []struct {
		name      string
		value     int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "正の値",
			value:     1,
			fieldName: "testField",
			wantErr:   false,
		},
		{
			name:      "ゼロ",
			value:     0,
			fieldName: "testField",
			wantErr:   true,
		},
		{
			name:      "負の値",
			value:     -1,
			fieldName: "testField",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePositiveInt(tt.value, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePositiveInt() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !IsRepositoryError(err) {
				t.Error("ValidatePositiveInt() should return RepositoryError")
			}
		})
	}
}

// TestValidateMaxLength はValidateMaxLengthのテスト
func TestValidateMaxLength(t *testing.T) {
	tests := []struct {
		name      string
		value     string
		maxLength int
		fieldName string
		wantErr   bool
	}{
		{
			name:      "最大長以下",
			value:     "test",
			maxLength: 10,
			fieldName: "testField",
			wantErr:   false,
		},
		{
			name:      "最大長と同じ",
			value:     "test",
			maxLength: 4,
			fieldName: "testField",
			wantErr:   false,
		},
		{
			name:      "最大長を超過",
			value:     "test",
			maxLength: 3,
			fieldName: "testField",
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMaxLength(tt.value, tt.maxLength, tt.fieldName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateMaxLength() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !IsRepositoryError(err) {
				t.Error("ValidateMaxLength() should return RepositoryError")
			}
		})
	}
}

// TestHandleSQLError はHandleSQLErrorのテスト
func TestHandleSQLError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		operation     string
		wantErrType   ErrorType
		wantNil       bool
	}{
		{
			name:      "nilエラー",
			err:       nil,
			operation: "test",
			wantNil:   true,
		},
		{
			name:        "sql.ErrNoRows",
			err:         sql.ErrNoRows,
			operation:   "test",
			wantErrType: ErrTypeNotFound,
			wantNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := HandleSQLError(tt.err, tt.operation)
			if tt.wantNil {
				if err != nil {
					t.Errorf("HandleSQLError() error = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Error("HandleSQLError() error = nil, want error")
				return
			}

			if !IsRepositoryError(err) {
				t.Error("HandleSQLError() should return RepositoryError")
				return
			}

			errType := GetRepositoryErrorType(err)
			if errType != tt.wantErrType {
				t.Errorf("HandleSQLError() error type = %v, want %v", errType, tt.wantErrType)
			}
		})
	}
}

// TestRepositoryError はRepositoryErrorのテスト
func TestRepositoryError(t *testing.T) {
	originalErr := sql.ErrNoRows
	repoErr := NewRepositoryError(ErrTypeNotFound, "test message", originalErr)

	// Error()メソッドのテスト
	errorStr := repoErr.Error()
	expectedStr := "[not_found] test message: sql: no rows in result set"
	if errorStr != expectedStr {
		t.Errorf("Error() = %v, want %v", errorStr, expectedStr)
	}

	// Unwrap()メソッドのテスト
	unwrapped := repoErr.Unwrap()
	if unwrapped != originalErr {
		t.Errorf("Unwrap() = %v, want %v", unwrapped, originalErr)
	}

	// IsRepositoryError()のテスト
	if !IsRepositoryError(repoErr) {
		t.Error("IsRepositoryError() should return true for RepositoryError")
	}

	// GetRepositoryErrorType()のテスト
	errType := GetRepositoryErrorType(repoErr)
	if errType != ErrTypeNotFound {
		t.Errorf("GetRepositoryErrorType() = %v, want %v", errType, ErrTypeNotFound)
	}
}
package models

import (
	"strings"
	"testing"
	"time"
)

func TestValidator_ValidateRequired(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name      string
		value     string
		fieldName string
		wantError bool
	}{
		{
			name:      "有効な値",
			value:     "test",
			fieldName: "field",
			wantError: false,
		},
		{
			name:      "空文字列",
			value:     "",
			fieldName: "field",
			wantError: true,
		},
		{
			name:      "空白のみ",
			value:     "   ",
			fieldName: "field",
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateRequired(tt.value, tt.fieldName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateRequired() error = %v, wantError %v", err, tt.wantError)
			}
			if err != nil {
				if err.Field != tt.fieldName {
					t.Errorf("ValidateRequired() field = %v, want %v", err.Field, tt.fieldName)
				}
				if err.Code != ErrorValidationRequiredField {
					t.Errorf("ValidateRequired() code = %v, want %v", err.Code, ErrorValidationRequiredField)
				}
			}
		})
	}
}

func TestValidator_ValidateStringLength(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name      string
		value     string
		fieldName string
		min       int
		max       int
		wantError bool
	}{
		{
			name:      "有効な長さ",
			value:     "test",
			fieldName: "field",
			min:       1,
			max:       10,
			wantError: false,
		},
		{
			name:      "最小長未満",
			value:     "a",
			fieldName: "field",
			min:       3,
			max:       10,
			wantError: true,
		},
		{
			name:      "最大長超過",
			value:     "very long string",
			fieldName: "field",
			min:       1,
			max:       5,
			wantError: true,
		},
		{
			name:      "日本語文字列",
			value:     "テスト",
			fieldName: "field",
			min:       1,
			max:       10,
			wantError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateStringLength(tt.value, tt.fieldName, tt.min, tt.max)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateStringLength() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateEmail(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name      string
		email     string
		fieldName string
		wantError bool
	}{
		{
			name:      "有効なメールアドレス",
			email:     "test@example.com",
			fieldName: "email",
			wantError: false,
		},
		{
			name:      "無効なメールアドレス",
			email:     "invalid-email",
			fieldName: "email",
			wantError: true,
		},
		{
			name:      "空文字列（スキップ）",
			email:     "",
			fieldName: "email",
			wantError: false,
		},
		{
			name:      "@マークなし",
			email:     "testexample.com",
			fieldName: "email",
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateEmail(tt.email, tt.fieldName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateEmail() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidatePassword(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name      string
		password  string
		fieldName string
		wantError bool
	}{
		{
			name:      "有効なパスワード",
			password:  "password123",
			fieldName: "password",
			wantError: false,
		},
		{
			name:      "短すぎるパスワード",
			password:  "pass1",
			fieldName: "password",
			wantError: true,
		},
		{
			name:      "数字なし",
			password:  "password",
			fieldName: "password",
			wantError: true,
		},
		{
			name:      "英字なし",
			password:  "12345678",
			fieldName: "password",
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidatePassword(tt.password, tt.fieldName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidatePassword() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateSportType(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name      string
		sport     SportType
		fieldName string
		wantError bool
	}{
		{
			name:      "有効なスポーツ",
			sport:     SportTypeVolleyball,
			fieldName: "sport",
			wantError: false,
		},
		{
			name:      "無効なスポーツ",
			sport:     SportType("invalid"),
			fieldName: "sport",
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateSportType(tt.sport, tt.fieldName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateSportType() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateMatchScore(t *testing.T) {
	validator := NewValidator()
	
	tests := []struct {
		name      string
		score1    int
		score2    int
		winner    string
		team1     string
		team2     string
		wantError bool
	}{
		{
			name:      "有効なスコア",
			score1:    3,
			score2:    1,
			winner:    "チームA",
			team1:     "チームA",
			team2:     "チームB",
			wantError: false,
		},
		{
			name:      "負のスコア",
			score1:    -1,
			score2:    1,
			winner:    "チームB",
			team1:     "チームA",
			team2:     "チームB",
			wantError: true,
		},
		{
			name:      "引き分け",
			score1:    1,
			score2:    1,
			winner:    "チームA",
			team1:     "チームA",
			team2:     "チームB",
			wantError: true,
		},
		{
			name:      "無効な勝者",
			score1:    3,
			score2:    1,
			winner:    "チームC",
			team1:     "チームA",
			team2:     "チームB",
			wantError: true,
		},
		{
			name:      "スコアと勝者の不一致",
			score1:    1,
			score2:    3,
			winner:    "チームA",
			team1:     "チームA",
			team2:     "チームB",
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateMatchScore(tt.score1, tt.score2, tt.winner, tt.team1, tt.team2)
			if errors.HasErrors() != tt.wantError {
				t.Errorf("ValidateMatchScore() hasErrors = %v, wantError %v", errors.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidator_ValidateFutureDateTime(t *testing.T) {
	validator := NewValidator()
	
	futureTime := time.Now().Add(1 * time.Hour)
	pastTime := time.Now().Add(-1 * time.Hour)
	
	tests := []struct {
		name      string
		value     time.Time
		fieldName string
		wantError bool
	}{
		{
			name:      "未来の日時",
			value:     futureTime,
			fieldName: "scheduled_at",
			wantError: false,
		},
		{
			name:      "過去の日時",
			value:     pastTime,
			fieldName: "scheduled_at",
			wantError: true,
		},
		{
			name:      "ゼロ値（スキップ）",
			value:     time.Time{},
			fieldName: "scheduled_at",
			wantError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateFutureDateTime(tt.value, tt.fieldName)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateFutureDateTime() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestValidationErrors_Add(t *testing.T) {
	var errors ValidationErrors
	
	errors.Add("field1", "message1", "value1", "CODE1", "rule1")
	errors.Add("field2", "message2", "value2", "CODE2", "rule2")
	
	if len(errors) != 2 {
		t.Errorf("Expected 2 errors, got %d", len(errors))
	}
	
	if errors[0].Field != "field1" {
		t.Errorf("Expected field1, got %s", errors[0].Field)
	}
	
	if errors[1].Code != "CODE2" {
		t.Errorf("Expected CODE2, got %s", errors[1].Code)
	}
}

func TestValidationErrors_GetFieldErrors(t *testing.T) {
	var errors ValidationErrors
	
	errors.Add("field1", "message1", "value1", "CODE1", "rule1")
	errors.Add("field2", "message2", "value2", "CODE2", "rule2")
	errors.Add("field1", "message3", "value3", "CODE3", "rule3")
	
	field1Errors := errors.GetFieldErrors("field1")
	if len(field1Errors) != 2 {
		t.Errorf("Expected 2 errors for field1, got %d", len(field1Errors))
	}
	
	field2Errors := errors.GetFieldErrors("field2")
	if len(field2Errors) != 1 {
		t.Errorf("Expected 1 error for field2, got %d", len(field2Errors))
	}
}

func TestValidateLoginRequest(t *testing.T) {
	tests := []struct {
		name      string
		request   *LoginRequest
		wantError bool
	}{
		{
			name: "有効なリクエスト",
			request: &LoginRequest{
				Username: "admin",
				Password: "password123",
			},
			wantError: false,
		},
		{
			name: "ユーザー名が空",
			request: &LoginRequest{
				Username: "",
				Password: "password123",
			},
			wantError: true,
		},
		{
			name: "パスワードが短い",
			request: &LoginRequest{
				Username: "admin",
				Password: "pass",
			},
			wantError: true,
		},
		{
			name: "ユーザー名に無効文字",
			request: &LoginRequest{
				Username: "admin@test",
				Password: "password123",
			},
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateLoginRequest(tt.request)
			if errors.HasErrors() != tt.wantError {
				t.Errorf("ValidateLoginRequest() hasErrors = %v, wantError %v", errors.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidateCreateTournamentRequest(t *testing.T) {
	tests := []struct {
		name      string
		request   *CreateTournamentRequest
		wantError bool
	}{
		{
			name: "有効なリクエスト",
			request: &CreateTournamentRequest{
				Sport:  SportTypeVolleyball,
				Format: TournamentFormatStandard,
			},
			wantError: false,
		},
		{
			name: "無効なスポーツ",
			request: &CreateTournamentRequest{
				Sport:  SportType("invalid"),
				Format: TournamentFormatStandard,
			},
			wantError: true,
		},
		{
			name: "無効なフォーマット",
			request: &CreateTournamentRequest{
				Sport:  SportTypeVolleyball,
				Format: TournamentFormat("invalid"),
			},
			wantError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateCreateTournamentRequest(tt.request)
			if errors.HasErrors() != tt.wantError {
				t.Errorf("ValidateCreateTournamentRequest() hasErrors = %v, wantError %v", errors.HasErrors(), tt.wantError)
			}
		})
	}
}

func TestValidationContext(t *testing.T) {
	ctx := NewValidationContext()
	
	// デフォルト言語の確認
	if ctx.Language != "ja" {
		t.Errorf("Expected default language 'ja', got %s", ctx.Language)
	}
	
	// 言語設定の変更
	ctx.SetLanguage("en")
	if ctx.Language != "en" {
		t.Errorf("Expected language 'en', got %s", ctx.Language)
	}
	
	// データの設定と取得
	ctx.SetData("key1", "value1")
	value, exists := ctx.GetData("key1")
	if !exists {
		t.Error("Expected data to exist")
	}
	if value != "value1" {
		t.Errorf("Expected 'value1', got %v", value)
	}
	
	// 存在しないデータの取得
	_, exists = ctx.GetData("nonexistent")
	if exists {
		t.Error("Expected data not to exist")
	}
}

func TestValidator_getLocalizedMessage(t *testing.T) {
	validator := NewValidator()
	
	// 日本語メッセージのテスト
	msg := validator.getLocalizedMessage("required", "フィールド名")
	expected := "フィールド名は必須です"
	if msg != expected {
		t.Errorf("Expected '%s', got '%s'", expected, msg)
	}
	
	// 英語メッセージのテスト
	validator.context.SetLanguage("en")
	msg = validator.getLocalizedMessage("required", "field_name")
	expected = "field_name is required"
	if msg != expected {
		t.Errorf("Expected '%s', got '%s'", expected, msg)
	}
	
	// 存在しないメッセージキーのテスト
	msg = validator.getLocalizedMessage("nonexistent", "field")
	if !strings.Contains(msg, "バリデーションエラー") {
		t.Errorf("Expected fallback message, got '%s'", msg)
	}
}
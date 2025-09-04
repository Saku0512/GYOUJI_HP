package repository

import (
	"testing"

	"backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func TestNewUserRepository(t *testing.T) {
	// NewUserRepositoryはデータベース接続が必要なため、統合テストで実行
	t.Skip("NewUserRepositoryはデータベース接続が必要なため、統合テストで実行")
}

// TestUserValidation はユーザーモデルの検証ロジックをテストする（データベース接続不要）
func TestUserValidation(t *testing.T) {
	tests := []struct {
		name        string
		user        *models.User
		expectError bool
	}{
		{
			name: "有効なユーザー",
			user: &models.User{
				Username: "testuser",
				Password: "password123",
				Role:     models.RoleAdmin,
			},
			expectError: false,
		},
		{
			name: "空のユーザー名",
			user: &models.User{
				Username: "",
				Password: "password123",
				Role:     models.RoleAdmin,
			},
			expectError: true,
		},
		{
			name: "空のパスワード",
			user: &models.User{
				Username: "testuser",
				Password: "",
				Role:     models.RoleAdmin,
			},
			expectError: true,
		},
		{
			name: "無効な役割",
			user: &models.User{
				Username: "testuser",
				Password: "password123",
				Role:     "invalid",
			},
			expectError: true,
		},
		{
			name: "短すぎるユーザー名",
			user: &models.User{
				Username: "ab",
				Password: "password123",
				Role:     models.RoleAdmin,
			},
			expectError: true,
		},
		{
			name: "長すぎるユーザー名",
			user: &models.User{
				Username: "a" + string(make([]byte, 50)), // 51文字
				Password: "password123",
				Role:     models.RoleAdmin,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.user.Validate()

			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				}
				return
			}

			if err != nil {
				t.Errorf("予期しないエラーが発生しました: %v", err)
			}
		})
	}
}

// TestValidateCredentialsInputs は認証情報の入力検証をテストする（データベース接続不要）
func TestValidateCredentialsInputs(t *testing.T) {
	tests := []struct {
		name     string
		username string
		password string
		valid    bool
	}{
		{
			name:     "有効な入力",
			username: "admin",
			password: "password123",
			valid:    true,
		},
		{
			name:     "空のユーザー名",
			username: "",
			password: "password123",
			valid:    false,
		},
		{
			name:     "空のパスワード",
			username: "admin",
			password: "",
			valid:    false,
		},
		{
			name:     "両方とも空",
			username: "",
			password: "",
			valid:    false,
		},
		{
			name:     "空白のみのユーザー名",
			username: "   ",
			password: "password123",
			valid:    false,
		},
		{
			name:     "空白のみのパスワード",
			username: "admin",
			password: "   ",
			valid:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// ValidateNotEmptyを直接テストして入力検証ロジックを確認
			usernameErr := ValidateNotEmpty(tt.username, "ユーザー名")
			passwordErr := ValidateNotEmpty(tt.password, "パスワード")
			
			hasError := usernameErr != nil || passwordErr != nil
			
			if tt.valid && hasError {
				t.Errorf("有効な入力でエラーが発生しました: username=%v, password=%v", usernameErr, passwordErr)
			}
			
			if !tt.valid && !hasError {
				t.Error("無効な入力でエラーが発生しませんでした")
			}
		})
	}
}

// TestUsernameValidation はユーザー名の検証をテストする（データベース接続不要）
func TestUsernameValidation(t *testing.T) {
	tests := []struct {
		name        string
		username    string
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "有効なユーザー名",
			username:    "admin",
			expectError: false,
		},
		{
			name:        "空のユーザー名",
			username:    "",
			expectError: true,
			errorType:   ErrTypeValidation,
		},
		{
			name:        "空白のみのユーザー名",
			username:    "   ",
			expectError: true,
			errorType:   ErrTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNotEmpty(tt.username, "ユーザー名")

			if tt.expectError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
					return
				}

				if IsRepositoryError(err) {
					repoErr := err.(*RepositoryError)
					if repoErr.Type != tt.errorType {
						t.Errorf("期待されたエラータイプ: %s, 実際: %s", tt.errorType, repoErr.Type)
					}
				}
				return
			}

			if err != nil {
				t.Errorf("予期しないエラーが発生しました: %v", err)
			}
		})
	}
}

// TestPasswordHashing はパスワードハッシュ化のテスト
func TestPasswordHashing(t *testing.T) {
	password := "testpassword123"
	
	// パスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("パスワードハッシュ化に失敗しました: %v", err)
	}

	// ハッシュ化されたパスワードが元のパスワードと異なることを確認
	if string(hashedPassword) == password {
		t.Error("ハッシュ化されたパスワードが元のパスワードと同じです")
	}

	// 正しいパスワードで検証
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		t.Errorf("正しいパスワードの検証に失敗しました: %v", err)
	}

	// 間違ったパスワードで検証
	err = bcrypt.CompareHashAndPassword(hashedPassword, []byte("wrongpassword"))
	if err == nil {
		t.Error("間違ったパスワードの検証が成功してしまいました")
	}
}

// TestUserCredentialsValidation はユーザー認証情報の検証をテストする
func TestUserCredentialsValidation(t *testing.T) {
	user := &models.User{
		Username: "admin",
		Password: "password123",
		Role:     models.RoleAdmin,
	}

	// ValidateCredentialsメソッドのテスト
	err := user.ValidateCredentials()
	if err != nil {
		t.Errorf("有効な認証情報でエラーが発生しました: %v", err)
	}

	// 空のユーザー名でのテスト
	emptyUsernameUser := &models.User{
		Username: "",
		Password: "password123",
		Role:     models.RoleAdmin,
	}
	err = emptyUsernameUser.ValidateCredentials()
	if err == nil {
		t.Error("空のユーザー名でエラーが発生しませんでした")
	}

	// 空のパスワードでのテスト
	emptyPasswordUser := &models.User{
		Username: "admin",
		Password: "",
		Role:     models.RoleAdmin,
	}
	err = emptyPasswordUser.ValidateCredentials()
	if err == nil {
		t.Error("空のパスワードでエラーが発生しませんでした")
	}
}

// 統合テスト用のスキップメッセージ
func TestUserRepository_IntegrationTests(t *testing.T) {
	t.Skip("統合テストはデータベース接続が必要なため、別途実行してください")
	
	// 以下のテストは実際のデータベース接続が必要です:
	// - GetAdminUser の実際のデータベースクエリ
	// - ValidateCredentials の実際のユーザー検索とパスワード検証
	// - CreateUser の実際のデータベース挿入
	// - GetUserByUsername の実際のデータベースクエリ
	// - 重複ユーザー作成のテスト
}
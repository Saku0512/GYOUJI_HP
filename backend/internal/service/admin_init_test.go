package service

import (
	"errors"
	"testing"

	"backend/internal/config"
	"backend/internal/models"
)

func TestNewAdminInitService(t *testing.T) {
	mockRepo := NewMockUserRepository()
	cfg := createTestConfig()
	
	service := NewAdminInitService(mockRepo, cfg)
	
	if service == nil {
		t.Error("NewAdminInitService should return a non-nil service")
	}
}

func TestAdminInitService_InitializeAdmin(t *testing.T) {
	tests := []struct {
		name           string
		adminUsername  string
		adminPassword  string
		existingUser   *models.User
		repoError      error
		expectedError  bool
		errorMessage   string
	}{
		{
			name:          "新しい管理者ユーザーの作成成功",
			adminUsername: "admin",
			adminPassword: "admin123",
			existingUser:  nil,
			repoError:     nil,
			expectedError: false,
		},
		{
			name:          "既存管理者ユーザーの更新成功",
			adminUsername: "admin",
			adminPassword: "newpassword",
			existingUser: &models.User{
				ID:       1,
				Username: "admin",
				Password: "oldhashedpassword",
				Role:     models.RoleAdmin,
			},
			repoError:     nil,
			expectedError: false,
		},
		{
			name:          "管理者ユーザー名が空",
			adminUsername: "",
			adminPassword: "admin123",
			existingUser:  nil,
			repoError:     nil,
			expectedError: true,
			errorMessage:  "管理者ユーザー名が設定されていません",
		},
		{
			name:          "管理者パスワードが空",
			adminUsername: "admin",
			adminPassword: "",
			existingUser:  nil,
			repoError:     nil,
			expectedError: true,
			errorMessage:  "管理者パスワードが設定されていません",
		},
		{
			name:          "ユーザー作成エラー",
			adminUsername: "admin",
			adminPassword: "admin123",
			existingUser:  nil,
			repoError:     errors.New("データベースエラー"),
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository()
			cfg := createTestConfig()
			cfg.Admin.Username = tt.adminUsername
			cfg.Admin.Password = tt.adminPassword
			
			// 既存ユーザーがある場合は追加
			if tt.existingUser != nil {
				mockRepo.users[tt.existingUser.Username] = tt.existingUser
			}
			
			// エラーを設定
			if tt.repoError != nil {
				mockRepo.SetError(tt.repoError)
			}
			
			service := NewAdminInitService(mockRepo, cfg)
			err := service.InitializeAdmin()
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				
				// パスワードハッシュが設定されているかチェック
				if cfg.Admin.PasswordHash == "" {
					t.Error("Admin password hash should be set after initialization")
				}
			}
		})
	}
}

func TestAdminInitService_hashPassword(t *testing.T) {
	mockRepo := NewMockUserRepository()
	cfg := createTestConfig()
	service := &adminInitServiceImpl{
		userRepo: mockRepo,
		config:   cfg,
	}

	tests := []struct {
		name          string
		password      string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "正常なパスワードハッシュ化",
			password:      "admin123",
			expectedError: false,
		},
		{
			name:          "空のパスワード",
			password:      "",
			expectedError: true,
			errorMessage:  "パスワードは必須です",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := service.hashPassword(tt.password)
			
			if tt.expectedError {
				if err == nil {
					t.Errorf("Expected error but got none")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error but got: %v", err)
				}
				
				if hash == "" {
					t.Error("Hash should not be empty")
				}
				
				if hash == tt.password {
					t.Error("Hash should be different from original password")
				}
			}
		})
	}
}
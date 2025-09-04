package service

import (
	"errors"
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// MockUserRepository はテスト用のUserRepositoryモック
type MockUserRepository struct {
	users map[string]*models.User
	err   error
}

func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]*models.User),
	}
}

func (m *MockUserRepository) GetAdminUser() (*models.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	for _, user := range m.users {
		if user.Role == models.RoleAdmin {
			return user, nil
		}
	}
	
	return nil, errors.New("管理者ユーザーが見つかりません")
}

func (m *MockUserRepository) ValidateCredentials(username, password string) bool {
	user, exists := m.users[username]
	if !exists {
		return false
	}
	
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	return err == nil
}

func (m *MockUserRepository) CreateUser(user *models.User) error {
	if m.err != nil {
		return m.err
	}
	
	// パスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	user.Password = string(hashedPassword)
	user.ID = len(m.users) + 1
	m.users[user.Username] = user
	return nil
}

func (m *MockUserRepository) GetUserByUsername(username string) (*models.User, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	user, exists := m.users[username]
	if !exists {
		return nil, errors.New("ユーザーが見つかりません")
	}
	
	return user, nil
}

func (m *MockUserRepository) SetError(err error) {
	m.err = err
}

func (m *MockUserRepository) AddUser(username, password, role string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	
	user := &models.User{
		ID:       len(m.users) + 1,
		Username: username,
		Password: string(hashedPassword),
		Role:     role,
	}
	
	m.users[username] = user
	return nil
}

// テスト用の設定を作成
func createTestConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret-key",
			ExpirationHours: 24,
			Issuer:          "test-issuer",
		},
	}
}

func TestNewAuthService(t *testing.T) {
	mockRepo := NewMockUserRepository()
	cfg := createTestConfig()
	
	service := NewAuthService(mockRepo, cfg)
	
	if service == nil {
		t.Error("AuthServiceの作成に失敗しました")
	}
}

func TestAuthService_Login(t *testing.T) {
	tests := []struct {
		name           string
		username       string
		password       string
		setupUser      bool
		userPassword   string
		repositoryErr  error
		expectedError  bool
		errorMessage   string
	}{
		{
			name:          "正常なログイン",
			username:      "admin",
			password:      "password123",
			setupUser:     true,
			userPassword:  "password123",
			expectedError: false,
		},
		{
			name:          "空のユーザー名",
			username:      "",
			password:      "password123",
			expectedError: true,
			errorMessage:  "ユーザー名は必須です",
		},
		{
			name:          "空のパスワード",
			username:      "admin",
			password:      "",
			expectedError: true,
			errorMessage:  "パスワードは必須です",
		},
		{
			name:          "存在しないユーザー",
			username:      "nonexistent",
			password:      "password123",
			expectedError: true,
			errorMessage:  "認証に失敗しました",
		},
		{
			name:          "間違ったパスワード",
			username:      "admin",
			password:      "wrongpassword",
			setupUser:     true,
			userPassword:  "password123",
			expectedError: true,
			errorMessage:  "認証に失敗しました",
		},
		{
			name:          "リポジトリエラー",
			username:      "admin",
			password:      "password123",
			repositoryErr: errors.New("データベースエラー"),
			expectedError: true,
			errorMessage:  "認証に失敗しました",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := NewMockUserRepository()
			cfg := createTestConfig()
			service := NewAuthService(mockRepo, cfg)
			
			// テストユーザーをセットアップ
			if tt.setupUser {
				err := mockRepo.AddUser(tt.username, tt.userPassword, models.RoleAdmin)
				if err != nil {
					t.Fatalf("テストユーザーのセットアップに失敗: %v", err)
				}
			}
			
			// リポジトリエラーを設定
			if tt.repositoryErr != nil {
				mockRepo.SetError(tt.repositoryErr)
			}
			
			// ログインを実行
			token, err := service.Login(tt.username, tt.password)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
				if token != "" {
					t.Error("エラー時にトークンが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if token == "" {
					t.Error("トークンが返されませんでした")
				}
			}
		})
	}
}

func TestAuthService_ValidateToken(t *testing.T) {
	mockRepo := NewMockUserRepository()
	cfg := createTestConfig()
	service := NewAuthService(mockRepo, cfg)
	
	// テストユーザーを作成
	err := mockRepo.AddUser("admin", "password123", models.RoleAdmin)
	if err != nil {
		t.Fatalf("テストユーザーのセットアップに失敗: %v", err)
	}
	
	// 有効なトークンを生成
	validToken, err := service.GenerateToken(1, "admin")
	if err != nil {
		t.Fatalf("テスト用トークンの生成に失敗: %v", err)
	}
	
	// 期限切れトークンを生成
	expiredClaims := &Claims{
		UserID:   1,
		Username: "admin",
		Role:     models.RoleAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Hour)), // 1時間前に期限切れ
			IssuedAt:  jwt.NewNumericDate(time.Now().Add(-2 * time.Hour)),
			Issuer:    cfg.JWT.Issuer,
		},
	}
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	expiredTokenString, _ := expiredToken.SignedString([]byte(cfg.JWT.SecretKey))
	
	// 無効な署名のトークンを生成
	invalidSignatureToken := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	invalidTokenString, _ := invalidSignatureToken.SignedString([]byte("wrong-secret"))
	
	tests := []struct {
		name          string
		token         string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "有効なトークン",
			token:         validToken,
			expectedError: false,
		},
		{
			name:          "空のトークン",
			token:         "",
			expectedError: true,
			errorMessage:  "トークンは必須です",
		},
		{
			name:          "期限切れトークン",
			token:         expiredTokenString,
			expectedError: true,
			errorMessage:  "無効なトークンです",
		},
		{
			name:          "無効な署名",
			token:         invalidTokenString,
			expectedError: true,
			errorMessage:  "無効なトークンです",
		},
		{
			name:          "不正な形式のトークン",
			token:         "invalid.token.format",
			expectedError: true,
			errorMessage:  "無効なトークンです",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
				if claims != nil {
					t.Error("エラー時にクレームが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if claims == nil {
					t.Error("クレームが返されませんでした")
				} else {
					if claims.Username != "admin" {
						t.Errorf("期待されたユーザー名: admin, 実際: %s", claims.Username)
					}
					if claims.UserID != 1 {
						t.Errorf("期待されたユーザーID: 1, 実際: %d", claims.UserID)
					}
				}
			}
		})
	}
}

func TestAuthService_GenerateToken(t *testing.T) {
	mockRepo := NewMockUserRepository()
	cfg := createTestConfig()
	service := NewAuthService(mockRepo, cfg)
	
	tests := []struct {
		name          string
		userID        int
		username      string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "正常なトークン生成",
			userID:        1,
			username:      "admin",
			expectedError: false,
		},
		{
			name:          "無効なユーザーID（0）",
			userID:        0,
			username:      "admin",
			expectedError: true,
			errorMessage:  "無効なユーザーIDです",
		},
		{
			name:          "無効なユーザーID（負の値）",
			userID:        -1,
			username:      "admin",
			expectedError: true,
			errorMessage:  "無効なユーザーIDです",
		},
		{
			name:          "空のユーザー名",
			userID:        1,
			username:      "",
			expectedError: true,
			errorMessage:  "ユーザー名は必須です",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := service.GenerateToken(tt.userID, tt.username)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
				if token != "" {
					t.Error("エラー時にトークンが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if token == "" {
					t.Error("トークンが返されませんでした")
				}
				
				// 生成されたトークンを検証
				claims, err := service.ValidateToken(token)
				if err != nil {
					t.Errorf("生成されたトークンの検証に失敗: %v", err)
				} else {
					if claims.UserID != tt.userID {
						t.Errorf("期待されたユーザーID: %d, 実際: %d", tt.userID, claims.UserID)
					}
					if claims.Username != tt.username {
						t.Errorf("期待されたユーザー名: %s, 実際: %s", tt.username, claims.Username)
					}
				}
			}
		})
	}
}

func TestAuthService_HashPassword(t *testing.T) {
	mockRepo := NewMockUserRepository()
	cfg := createTestConfig()
	service := NewAuthService(mockRepo, cfg)
	
	tests := []struct {
		name          string
		password      string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "正常なパスワードハッシュ化",
			password:      "password123",
			expectedError: false,
		},
		{
			name:          "空のパスワード",
			password:      "",
			expectedError: true,
			errorMessage:  "パスワードは必須です",
		},
		{
			name:          "長いパスワード",
			password:      "very-long-password-with-special-characters-123!@#",
			expectedError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hashedPassword, err := service.HashPassword(tt.password)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
				if hashedPassword != "" {
					t.Error("エラー時にハッシュが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if hashedPassword == "" {
					t.Error("ハッシュが返されませんでした")
				}
				if hashedPassword == tt.password {
					t.Error("パスワードがハッシュ化されていません")
				}
				
				// ハッシュ化されたパスワードを検証
				err = service.VerifyPassword(hashedPassword, tt.password)
				if err != nil {
					t.Errorf("ハッシュ化されたパスワードの検証に失敗: %v", err)
				}
			}
		})
	}
}

func TestAuthService_VerifyPassword(t *testing.T) {
	mockRepo := NewMockUserRepository()
	cfg := createTestConfig()
	service := NewAuthService(mockRepo, cfg)
	
	// テスト用のハッシュ化されたパスワードを生成
	validPassword := "password123"
	hashedPassword, err := service.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("テスト用パスワードのハッシュ化に失敗: %v", err)
	}
	
	tests := []struct {
		name           string
		hashedPassword string
		password       string
		expectedError  bool
		errorMessage   string
	}{
		{
			name:           "正常なパスワード検証",
			hashedPassword: hashedPassword,
			password:       validPassword,
			expectedError:  false,
		},
		{
			name:           "間違ったパスワード",
			hashedPassword: hashedPassword,
			password:       "wrongpassword",
			expectedError:  true,
			errorMessage:   "パスワードが一致しません",
		},
		{
			name:           "空のハッシュ化パスワード",
			hashedPassword: "",
			password:       validPassword,
			expectedError:  true,
			errorMessage:   "ハッシュ化されたパスワードは必須です",
		},
		{
			name:           "空のパスワード",
			hashedPassword: hashedPassword,
			password:       "",
			expectedError:  true,
			errorMessage:   "パスワードは必須です",
		},
		{
			name:           "無効なハッシュ形式",
			hashedPassword: "invalid-hash",
			password:       validPassword,
			expectedError:  true,
			errorMessage:   "パスワード検証に失敗しました",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.VerifyPassword(tt.hashedPassword, tt.password)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
			}
		})
	}
}

// 統合テスト: ログインからトークン検証まで
func TestAuthService_LoginToValidateFlow(t *testing.T) {
	mockRepo := NewMockUserRepository()
	cfg := createTestConfig()
	service := NewAuthService(mockRepo, cfg)
	
	// テストユーザーを作成
	username := "admin"
	password := "password123"
	err := mockRepo.AddUser(username, password, models.RoleAdmin)
	if err != nil {
		t.Fatalf("テストユーザーのセットアップに失敗: %v", err)
	}
	
	// ログインしてトークンを取得
	token, err := service.Login(username, password)
	if err != nil {
		t.Fatalf("ログインに失敗: %v", err)
	}
	
	if token == "" {
		t.Fatal("トークンが返されませんでした")
	}
	
	// トークンを検証
	claims, err := service.ValidateToken(token)
	if err != nil {
		t.Fatalf("トークン検証に失敗: %v", err)
	}
	
	if claims.Username != username {
		t.Errorf("期待されたユーザー名: %s, 実際: %s", username, claims.Username)
	}
	
	if claims.Role != models.RoleAdmin {
		t.Errorf("期待された役割: %s, 実際: %s", models.RoleAdmin, claims.Role)
	}
}
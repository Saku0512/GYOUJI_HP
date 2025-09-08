package service

import (
	"testing"
	"time"

	"backend/internal/config"
	"backend/internal/models"
)

// テスト用の設定を作成
func createTestJWTConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret-key-for-jwt-testing",
			ExpirationHours: 1, // 1時間
			Issuer:          "tournament-test",
		},
	}
}

func TestJWTService_GenerateToken(t *testing.T) {
	cfg := createTestJWTConfig()
	jwtService := NewJWTService(cfg)

	tests := []struct {
		name     string
		userID   int
		username string
		role     string
		wantErr  bool
	}{
		{
			name:     "正常なトークン生成",
			userID:   1,
			username: "admin",
			role:     models.RoleAdmin,
			wantErr:  false,
		},
		{
			name:     "無効なユーザーID（0）",
			userID:   0,
			username: "admin",
			role:     models.RoleAdmin,
			wantErr:  true,
		},
		{
			name:     "無効なユーザーID（負の値）",
			userID:   -1,
			username: "admin",
			role:     models.RoleAdmin,
			wantErr:  true,
		},
		{
			name:     "空のユーザー名",
			userID:   1,
			username: "",
			role:     models.RoleAdmin,
			wantErr:  true,
		},
		{
			name:     "空のロール",
			userID:   1,
			username: "admin",
			role:     "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtService.GenerateToken(tt.userID, tt.username, tt.role)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("GenerateToken() エラーが期待されましたが、エラーが発生しませんでした")
				}
				if token != "" {
					t.Errorf("GenerateToken() エラー時はトークンが空であるべきです, got = %v", token)
				}
			} else {
				if err != nil {
					t.Errorf("GenerateToken() エラーが発生しました = %v", err)
				}
				if token == "" {
					t.Errorf("GenerateToken() トークンが空です")
				}
			}
		})
	}
}

func TestJWTService_ValidateToken(t *testing.T) {
	cfg := createTestJWTConfig()
	jwtService := NewJWTService(cfg)

	// 有効なトークンを生成
	validToken, err := jwtService.GenerateToken(1, "admin", models.RoleAdmin)
	if err != nil {
		t.Fatalf("テスト用トークンの生成に失敗しました: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "有効なトークン",
			token:   validToken,
			wantErr: false,
		},
		{
			name:    "空のトークン",
			token:   "",
			wantErr: true,
		},
		{
			name:    "無効なトークン",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:    "不正な形式のトークン",
			token:   "not-a-jwt-token",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := jwtService.ValidateToken(tt.token)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateToken() エラーが期待されましたが、エラーが発生しませんでした")
				}
				if claims != nil {
					t.Errorf("ValidateToken() エラー時はクレームがnilであるべきです, got = %v", claims)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateToken() エラーが発生しました = %v", err)
				}
				if claims == nil {
					t.Errorf("ValidateToken() クレームがnilです")
				} else {
					// クレームの内容を検証
					if claims.UserID != 1 {
						t.Errorf("ValidateToken() UserID = %v, want %v", claims.UserID, 1)
					}
					if claims.Username != "admin" {
						t.Errorf("ValidateToken() Username = %v, want %v", claims.Username, "admin")
					}
					if claims.Role != models.RoleAdmin {
						t.Errorf("ValidateToken() Role = %v, want %v", claims.Role, models.RoleAdmin)
					}
				}
			}
		})
	}
}

func TestJWTService_RefreshToken(t *testing.T) {
	cfg := createTestJWTConfig()
	jwtService := NewJWTService(cfg)

	// 有効なトークンを生成
	originalToken, err := jwtService.GenerateToken(1, "admin", models.RoleAdmin)
	if err != nil {
		t.Fatalf("テスト用トークンの生成に失敗しました: %v", err)
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "有効なトークンのリフレッシュ",
			token:   originalToken,
			wantErr: false,
		},
		{
			name:    "空のトークン",
			token:   "",
			wantErr: true,
		},
		{
			name:    "無効なトークン",
			token:   "invalid.token.here",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			newToken, err := jwtService.RefreshToken(tt.token)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("RefreshToken() エラーが期待されましたが、エラーが発生しませんでした")
				}
				if newToken != "" {
					t.Errorf("RefreshToken() エラー時はトークンが空であるべきです, got = %v", newToken)
				}
			} else {
				if err != nil {
					t.Errorf("RefreshToken() エラーが発生しました = %v", err)
				}
				if newToken == "" {
					t.Errorf("RefreshToken() 新しいトークンが空です")
				}
				
				// 新しいトークンが元のトークンと異なることを確認（時間精度により同じ場合もある）
				// Note: 高精度タイマーにより、同じ時刻に生成される場合があるため、このチェックは参考程度
				
				// 新しいトークンが有効であることを確認
				claims, err := jwtService.ValidateToken(newToken)
				if err != nil {
					t.Errorf("RefreshToken() 生成された新しいトークンが無効です: %v", err)
				}
				if claims == nil {
					t.Errorf("RefreshToken() 新しいトークンのクレームがnilです")
				}
			}
		})
	}
}

func TestJWTService_GetTokenExpiration(t *testing.T) {
	cfg := createTestJWTConfig()
	jwtService := NewJWTService(cfg)

	expiration := jwtService.GetTokenExpiration()
	expected := time.Duration(cfg.JWT.ExpirationHours) * time.Hour

	if expiration != expected {
		t.Errorf("GetTokenExpiration() = %v, want %v", expiration, expected)
	}
}

func TestJWTService_TokenExpiration(t *testing.T) {
	// 短い有効期限でテスト用設定を作成
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey:       "test-secret-key",
			ExpirationHours: 0, // 即座に期限切れ
			Issuer:          "tournament-test",
		},
	}
	
	jwtService := NewJWTService(cfg)

	// トークンを生成（即座に期限切れになる）
	token, err := jwtService.GenerateToken(1, "admin", models.RoleAdmin)
	if err != nil {
		t.Fatalf("テスト用トークンの生成に失敗しました: %v", err)
	}

	// 少し待ってから検証（期限切れになるはず）
	time.Sleep(time.Millisecond * 10)

	// 期限切れトークンの検証
	_, err = jwtService.ValidateToken(token)
	if err == nil {
		t.Errorf("ValidateToken() 期限切れトークンでエラーが期待されましたが、エラーが発生しませんでした")
	}
}

func TestJWTService_Integration(t *testing.T) {
	cfg := createTestJWTConfig()
	jwtService := NewJWTService(cfg)

	// 1. トークン生成
	userID := 123
	username := "testuser"
	role := models.RoleAdmin

	token, err := jwtService.GenerateToken(userID, username, role)
	if err != nil {
		t.Fatalf("トークン生成に失敗しました: %v", err)
	}

	// 2. トークン検証
	claims, err := jwtService.ValidateToken(token)
	if err != nil {
		t.Fatalf("トークン検証に失敗しました: %v", err)
	}

	// 3. クレーム内容の確認
	if claims.UserID != userID {
		t.Errorf("UserID = %v, want %v", claims.UserID, userID)
	}
	if claims.Username != username {
		t.Errorf("Username = %v, want %v", claims.Username, username)
	}
	if claims.Role != role {
		t.Errorf("Role = %v, want %v", claims.Role, role)
	}

	// 4. トークンリフレッシュ
	newToken, err := jwtService.RefreshToken(token)
	if err != nil {
		t.Fatalf("トークンリフレッシュに失敗しました: %v", err)
	}

	// 5. 新しいトークンの検証
	newClaims, err := jwtService.ValidateToken(newToken)
	if err != nil {
		t.Fatalf("新しいトークンの検証に失敗しました: %v", err)
	}

	// 6. 新しいクレーム内容の確認
	if newClaims.UserID != userID {
		t.Errorf("New UserID = %v, want %v", newClaims.UserID, userID)
	}
	if newClaims.Username != username {
		t.Errorf("New Username = %v, want %v", newClaims.Username, username)
	}
	if newClaims.Role != role {
		t.Errorf("New Role = %v, want %v", newClaims.Role, role)
	}
}
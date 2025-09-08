package service

import (
	"errors"
	"log"

	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// AuthService は認証関連のビジネスロジックを提供するインターフェース
type AuthService interface {
	// Login はユーザー認証を行い、JWTトークンとクレームを生成する
	Login(username, password string) (string, *JWTClaims, error)
	
	// ValidateToken はJWTトークンを検証し、クレームを返す
	ValidateToken(tokenString string) (*JWTClaims, error)
	
	// GenerateToken はユーザーIDに基づいてJWTトークンを生成する
	GenerateToken(userID int, username string) (string, error)
	
	// RefreshToken は既存のトークンから新しいトークンとクレームを生成する
	RefreshToken(tokenString string) (string, *JWTClaims, error)
	
	// HashPassword はパスワードをbcryptでハッシュ化する
	HashPassword(password string) (string, error)
	
	// VerifyPassword はプレーンテキストパスワードとハッシュを比較する
	VerifyPassword(hashedPassword, password string) error
}

// authServiceImpl はAuthServiceの実装
type authServiceImpl struct {
	userRepo   repository.UserRepository
	config     *config.Config
	jwtService JWTService
}

// NewAuthService は新しいAuthServiceインスタンスを作成する
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authServiceImpl{
		userRepo:   userRepo,
		config:     cfg,
		jwtService: NewJWTService(cfg),
	}
}

// Login はユーザー認証を行い、JWTトークンとクレームを生成する
func (s *authServiceImpl) Login(username, password string) (string, *JWTClaims, error) {
	// 入力値の検証
	if username == "" {
		return "", nil, errors.New("ユーザー名は必須です")
	}
	
	if password == "" {
		return "", nil, errors.New("パスワードは必須です")
	}
	
	log.Printf("ログイン試行: %s", username)
	
	// 管理者認証を優先的にチェック
	log.Printf("管理者認証チェック: username=%s, config.Admin.Username=%s", username, s.config.Admin.Username)
	if username == s.config.Admin.Username {
		// ハッシュ化されたパスワードで検証
		if err := s.VerifyPassword(s.config.Admin.PasswordHash, password); err == nil {
			// 管理者認証成功 - 固定のユーザーIDを使用
			token, err := s.jwtService.GenerateToken(1, username, models.RoleAdmin)
			if err != nil {
				log.Printf("管理者トークン生成エラー: %v", err)
				return "", nil, errors.New("トークン生成に失敗しました")
			}
			
			// トークンを検証してクレームを取得
			claims, err := s.jwtService.ValidateToken(token)
			if err != nil {
				log.Printf("管理者トークン検証エラー: %v", err)
				return "", nil, errors.New("トークン検証に失敗しました")
			}
			
			log.Printf("管理者ログイン成功: %s", username)
			return token, claims, nil
		} else {
			log.Printf("管理者パスワード検証失敗: %s", username)
			return "", nil, errors.New("認証に失敗しました")
		}
	}
	
	// 通常ユーザーの認証（データベースから取得）
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		log.Printf("ユーザー取得エラー: %v", err)
		return "", nil, errors.New("認証に失敗しました")
	}
	
	// パスワード検証
	if err := s.VerifyPassword(user.Password, password); err != nil {
		log.Printf("パスワード検証失敗: %s", username)
		return "", nil, errors.New("認証に失敗しました")
	}
	
	// JWTトークン生成（通常ユーザーは管理者ロールを付与）
	token, err := s.jwtService.GenerateToken(user.ID, user.Username, models.RoleAdmin)
	if err != nil {
		log.Printf("トークン生成エラー: %v", err)
		return "", nil, errors.New("トークン生成に失敗しました")
	}
	
	// トークンを検証してクレームを取得
	claims, err := s.jwtService.ValidateToken(token)
	if err != nil {
		log.Printf("トークン検証エラー: %v", err)
		return "", nil, errors.New("トークン検証に失敗しました")
	}
	
	log.Printf("ログイン成功: %s", username)
	return token, claims, nil
}

// ValidateToken はJWTトークンを検証し、クレームを返す
func (s *authServiceImpl) ValidateToken(tokenString string) (*JWTClaims, error) {
	return s.jwtService.ValidateToken(tokenString)
}

// GenerateToken はユーザーIDに基づいてJWTトークンを生成する
func (s *authServiceImpl) GenerateToken(userID int, username string) (string, error) {
	// デフォルトで管理者ロールを設定（後方互換性のため）
	return s.jwtService.GenerateToken(userID, username, models.RoleAdmin)
}

// RefreshToken は既存のトークンから新しいトークンとクレームを生成する
func (s *authServiceImpl) RefreshToken(tokenString string) (string, *JWTClaims, error) {
	// 既存のトークンを検証（期限切れでも構造が正しければOK）
	claims, err := s.jwtService.ParseTokenIgnoreExpiration(tokenString)
	if err != nil {
		return "", nil, err
	}
	
	// 新しいトークンを生成
	newToken, err := s.jwtService.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		log.Printf("リフレッシュトークン生成エラー: %v", err)
		return "", nil, errors.New("新しいトークンの生成に失敗しました")
	}
	
	// 新しいトークンを検証してクレームを取得
	newClaims, err := s.jwtService.ValidateToken(newToken)
	if err != nil {
		log.Printf("リフレッシュトークン検証エラー: %v", err)
		return "", nil, errors.New("新しいトークンの検証に失敗しました")
	}
	
	log.Printf("JWTリフレッシュ成功: user_id=%d, username=%s", claims.UserID, claims.Username)
	
	return newToken, newClaims, nil
}

// HashPassword はパスワードをbcryptでハッシュ化する
func (s *authServiceImpl) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("パスワードは必須です")
	}
	
	// bcryptでパスワードをハッシュ化
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("パスワードハッシュ化エラー: %v", err)
		return "", errors.New("パスワードハッシュ化に失敗しました")
	}
	
	return string(hashedBytes), nil
}

// VerifyPassword はプレーンテキストパスワードとハッシュを比較する
func (s *authServiceImpl) VerifyPassword(hashedPassword, password string) error {
	if hashedPassword == "" {
		return errors.New("ハッシュ化されたパスワードは必須です")
	}
	
	if password == "" {
		return errors.New("パスワードは必須です")
	}
	
	// bcryptでパスワードを比較
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return errors.New("パスワードが一致しません")
		}
		log.Printf("パスワード検証エラー: %v", err)
		return errors.New("パスワード検証に失敗しました")
	}
	
	return nil
}
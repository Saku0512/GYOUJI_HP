package service

import (
	"errors"
	"log"
	"time"

	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthService は認証関連のビジネスロジックを提供するインターフェース
type AuthService interface {
	// Login はユーザー認証を行い、JWTトークンを生成する
	Login(username, password string) (string, error)
	
	// ValidateToken はJWTトークンを検証し、クレームを返す
	ValidateToken(tokenString string) (*Claims, error)
	
	// GenerateToken はユーザーIDに基づいてJWTトークンを生成する
	GenerateToken(userID int, username string) (string, error)
	
	// HashPassword はパスワードをbcryptでハッシュ化する
	HashPassword(password string) (string, error)
	
	// VerifyPassword はプレーンテキストパスワードとハッシュを比較する
	VerifyPassword(hashedPassword, password string) error
}

// Claims はJWTトークンのクレームを表す構造体
type Claims struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// authServiceImpl はAuthServiceの実装
type authServiceImpl struct {
	userRepo repository.UserRepository
	config   *config.Config
}

// NewAuthService は新しいAuthServiceインスタンスを作成する
func NewAuthService(userRepo repository.UserRepository, cfg *config.Config) AuthService {
	return &authServiceImpl{
		userRepo: userRepo,
		config:   cfg,
	}
}

// Login はユーザー認証を行い、JWTトークンを生成する
func (s *authServiceImpl) Login(username, password string) (string, error) {
	// 入力値の検証
	if username == "" {
		return "", errors.New("ユーザー名は必須です")
	}
	
	if password == "" {
		return "", errors.New("パスワードは必須です")
	}
	
	log.Printf("ログイン試行: %s", username)
	
	// ユーザーの存在確認と認証情報検証
	user, err := s.userRepo.GetUserByUsername(username)
	if err != nil {
		log.Printf("ユーザー取得エラー: %v", err)
		return "", errors.New("認証に失敗しました")
	}
	
	// パスワード検証
	if err := s.VerifyPassword(user.Password, password); err != nil {
		log.Printf("パスワード検証失敗: %s", username)
		return "", errors.New("認証に失敗しました")
	}
	
	// JWTトークン生成
	token, err := s.GenerateToken(user.ID, user.Username)
	if err != nil {
		log.Printf("トークン生成エラー: %v", err)
		return "", errors.New("トークン生成に失敗しました")
	}
	
	log.Printf("ログイン成功: %s", username)
	return token, nil
}

// ValidateToken はJWTトークンを検証し、クレームを返す
func (s *authServiceImpl) ValidateToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, errors.New("トークンは必須です")
	}
	
	// JWTトークンをパース
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名方法の検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("無効な署名方法です")
		}
		return []byte(s.config.JWT.SecretKey), nil
	})
	
	if err != nil {
		log.Printf("トークンパースエラー: %v", err)
		return nil, errors.New("無効なトークンです")
	}
	
	// クレームの取得と検証
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		log.Printf("無効なクレーム")
		return nil, errors.New("無効なトークンです")
	}
	
	// トークンの有効期限確認
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		log.Printf("トークンが期限切れです")
		return nil, errors.New("トークンが期限切れです")
	}
	
	log.Printf("トークン検証成功: %s", claims.Username)
	return claims, nil
}

// GenerateToken はユーザーIDに基づいてJWTトークンを生成する
func (s *authServiceImpl) GenerateToken(userID int, username string) (string, error) {
	if userID <= 0 {
		return "", errors.New("無効なユーザーIDです")
	}
	
	if username == "" {
		return "", errors.New("ユーザー名は必須です")
	}
	
	// トークンの有効期限を設定
	expirationTime := time.Now().Add(s.config.GetJWTExpiration())
	
	// クレームを作成
	claims := &Claims{
		UserID:   userID,
		Username: username,
		Role:     models.RoleAdmin, // 現在は管理者のみサポート
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    s.config.JWT.Issuer,
			Subject:   username,
		},
	}
	
	// トークンを作成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// トークンに署名
	tokenString, err := token.SignedString([]byte(s.config.JWT.SecretKey))
	if err != nil {
		log.Printf("トークン署名エラー: %v", err)
		return "", errors.New("トークン生成に失敗しました")
	}
	
	log.Printf("トークン生成成功: %s (有効期限: %v)", username, expirationTime)
	return tokenString, nil
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
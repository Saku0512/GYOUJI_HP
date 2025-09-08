package service

import (
	"errors"
	"log"
	"strconv"
	"time"

	"backend/internal/config"

	"github.com/golang-jwt/jwt/v5"
)

// JWTService は統一されたJWT管理サービス
type JWTService interface {
	// GenerateToken はユーザー情報からJWTトークンを生成する
	GenerateToken(userID int, username string, role string) (string, error)
	
	// ValidateToken はJWTトークンを検証し、クレームを返す
	ValidateToken(tokenString string) (*JWTClaims, error)
	
	// RefreshToken は既存のトークンから新しいトークンを生成する
	RefreshToken(tokenString string) (string, error)
	
	// ParseTokenIgnoreExpiration は有効期限を無視してトークンをパースする（リフレッシュ用）
	ParseTokenIgnoreExpiration(tokenString string) (*JWTClaims, error)
	
	// GetTokenExpiration はトークンの有効期限を取得する
	GetTokenExpiration() time.Duration
}

// JWTClaims は統一されたJWTクレーム構造体
type JWTClaims struct {
	UserID   int    `json:"user_id"`   // ユーザーID
	Username string `json:"username"`  // ユーザー名
	Role     string `json:"role"`      // ユーザーロール
	jwt.RegisteredClaims
}

// jwtServiceImpl はJWTServiceの実装
type jwtServiceImpl struct {
	config *config.Config
}

// NewJWTService は新しいJWTServiceインスタンスを作成する
func NewJWTService(cfg *config.Config) JWTService {
	return &jwtServiceImpl{
		config: cfg,
	}
}

// GenerateToken はユーザー情報からJWTトークンを生成する
func (s *jwtServiceImpl) GenerateToken(userID int, username string, role string) (string, error) {
	// 入力値の検証
	if userID <= 0 {
		return "", errors.New("無効なユーザーIDです")
	}
	
	if username == "" {
		return "", errors.New("ユーザー名は必須です")
	}
	
	if role == "" {
		return "", errors.New("ユーザーロールは必須です")
	}
	
	// 現在時刻（ナノ秒精度で一意性を保証）
	now := time.Now()
	
	// トークンの有効期限を設定
	expirationTime := now.Add(s.GetTokenExpiration())
	
	// 統一されたクレームを作成
	claims := &JWTClaims{
		UserID:   userID,
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.config.JWT.Issuer,
			Subject:   strconv.Itoa(userID),
			ID:        generateJTI(userID, now), // JWT ID for uniqueness
		},
	}
	
	// トークンを作成
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	
	// トークンに署名
	tokenString, err := token.SignedString([]byte(s.config.JWT.SecretKey))
	if err != nil {
		log.Printf("JWT署名エラー: %v", err)
		return "", errors.New("トークン生成に失敗しました")
	}
	
	log.Printf("JWT生成成功: user_id=%d, username=%s, role=%s, expires_at=%v", 
		userID, username, role, expirationTime)
	
	return tokenString, nil
}

// ValidateToken はJWTトークンを検証し、クレームを返す
func (s *jwtServiceImpl) ValidateToken(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, errors.New("トークンは必須です")
	}
	
	// JWTトークンをパース
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名方法の検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("無効な署名方法です")
		}
		return []byte(s.config.JWT.SecretKey), nil
	})
	
	if err != nil {
		log.Printf("JWTパースエラー: %v", err)
		return nil, errors.New("無効なトークンです")
	}
	
	// クレームの取得と検証
	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		log.Printf("無効なJWTクレーム")
		return nil, errors.New("無効なトークンです")
	}
	
	// トークンの有効期限確認
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		log.Printf("JWTトークンが期限切れです: user_id=%d, expires_at=%v", 
			claims.UserID, claims.ExpiresAt.Time)
		return nil, errors.New("トークンが期限切れです")
	}
	
	// 必須フィールドの検証
	if claims.UserID <= 0 {
		return nil, errors.New("無効なユーザーIDです")
	}
	
	if claims.Username == "" {
		return nil, errors.New("ユーザー名が設定されていません")
	}
	
	if claims.Role == "" {
		return nil, errors.New("ユーザーロールが設定されていません")
	}
	
	log.Printf("JWT検証成功: user_id=%d, username=%s, role=%s", 
		claims.UserID, claims.Username, claims.Role)
	
	return claims, nil
}

// RefreshToken は既存のトークンから新しいトークンを生成する
func (s *jwtServiceImpl) RefreshToken(tokenString string) (string, error) {
	// 既存のトークンを検証（期限切れでも構造が正しければOK）
	claims, err := s.ParseTokenIgnoreExpiration(tokenString)
	if err != nil {
		return "", err
	}
	
	// 新しいトークンを生成
	newToken, err := s.GenerateToken(claims.UserID, claims.Username, claims.Role)
	if err != nil {
		log.Printf("リフレッシュトークン生成エラー: %v", err)
		return "", errors.New("新しいトークンの生成に失敗しました")
	}
	
	log.Printf("JWTリフレッシュ成功: user_id=%d, username=%s", claims.UserID, claims.Username)
	
	return newToken, nil
}

// GetTokenExpiration はトークンの有効期限を取得する
func (s *jwtServiceImpl) GetTokenExpiration() time.Duration {
	return s.config.GetJWTExpiration()
}

// ParseTokenIgnoreExpiration は有効期限を無視してトークンをパースする（リフレッシュ用）
func (s *jwtServiceImpl) ParseTokenIgnoreExpiration(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, errors.New("トークンは必須です")
	}
	
	// 有効期限チェックを無効にしたパーサーを作成
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	
	token, err := parser.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 署名方法の検証
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("無効な署名方法です")
		}
		return []byte(s.config.JWT.SecretKey), nil
	})
	
	if err != nil {
		log.Printf("リフレッシュ用JWTパースエラー: %v", err)
		return nil, errors.New("無効なトークンです")
	}
	
	// クレームの取得
	claims, ok := token.Claims.(*JWTClaims)
	if !ok {
		log.Printf("無効なJWTクレーム（リフレッシュ用）")
		return nil, errors.New("無効なトークンです")
	}
	
	// 基本的なフィールド検証（有効期限は除く）
	if claims.UserID <= 0 {
		return nil, errors.New("無効なユーザーIDです")
	}
	
	if claims.Username == "" {
		return nil, errors.New("ユーザー名が設定されていません")
	}
	
	if claims.Role == "" {
		return nil, errors.New("ユーザーロールが設定されていません")
	}
	
	return claims, nil
}

// generateJTI はJWT IDを生成する（トークンの一意性を保証）
func generateJTI(userID int, issuedAt time.Time) string {
	return strconv.Itoa(userID) + "_" + strconv.FormatInt(issuedAt.UnixNano(), 10)
}
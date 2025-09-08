package service

import (
	"errors"
	"log"

	"backend/internal/config"
	"backend/internal/models"
	"backend/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

// AdminInitService は管理者初期化関連のビジネスロジックを提供するインターフェース
type AdminInitService interface {
	// InitializeAdmin は管理者ユーザーを初期化する
	InitializeAdmin() error
}

// adminInitServiceImpl はAdminInitServiceの実装
type adminInitServiceImpl struct {
	userRepo repository.UserRepository
	config   *config.Config
}

// NewAdminInitService は新しいAdminInitServiceインスタンスを作成する
func NewAdminInitService(userRepo repository.UserRepository, cfg *config.Config) AdminInitService {
	return &adminInitServiceImpl{
		userRepo: userRepo,
		config:   cfg,
	}
}

// InitializeAdmin は管理者ユーザーを初期化する
func (s *adminInitServiceImpl) InitializeAdmin() error {
	// 設定値の検証
	if s.config.Admin.Username == "" {
		return errors.New("管理者ユーザー名が設定されていません")
	}
	
	if s.config.Admin.Password == "" {
		return errors.New("管理者パスワードが設定されていません")
	}
	
	log.Printf("管理者ユーザーの初期化を開始します: %s", s.config.Admin.Username)
	
	// パスワードをハッシュ化
	hashedPassword, err := s.hashPassword(s.config.Admin.Password)
	if err != nil {
		log.Printf("管理者パスワードのハッシュ化に失敗しました: %v", err)
		return err
	}
	
	// ハッシュ化されたパスワードを設定に保存（認証時に使用）
	s.config.Admin.PasswordHash = hashedPassword
	
	// 既存の管理者ユーザーをチェック
	existingUser, err := s.userRepo.GetUserByUsername(s.config.Admin.Username)
	if err == nil && existingUser != nil {
		// 既存ユーザーが存在する場合、パスワードを更新
		log.Printf("既存の管理者ユーザーが見つかりました。パスワードを更新します: %s", s.config.Admin.Username)
		
		existingUser.Password = hashedPassword
		if err := s.userRepo.UpdateUser(existingUser); err != nil {
			log.Printf("管理者ユーザーの更新に失敗しました: %v", err)
			return err
		}
		
		log.Printf("管理者ユーザーのパスワードを更新しました: %s", s.config.Admin.Username)
		return nil
	}
	
	// 新しい管理者ユーザーを作成
	log.Printf("新しい管理者ユーザーを作成します: %s", s.config.Admin.Username)
	
	adminUser := &models.User{
		Username: s.config.Admin.Username,
		Password: hashedPassword,
		Role:     models.RoleAdmin,
	}
	
	if err := s.userRepo.CreateUser(adminUser); err != nil {
		log.Printf("管理者ユーザーの作成に失敗しました: %v", err)
		return err
	}
	
	log.Printf("管理者ユーザーを作成しました: %s (ID: %d)", adminUser.Username, adminUser.ID)
	return nil
}

// hashPassword はパスワードをbcryptでハッシュ化する
func (s *adminInitServiceImpl) hashPassword(password string) (string, error) {
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
package repository

import (
	"log"

	"backend/internal/database"
	"backend/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserRepository はユーザー関連のデータアクセスを提供するインターフェース
type UserRepository interface {
	// GetAdminUser は管理者ユーザーを取得する
	GetAdminUser() (*models.User, error)
	
	// ValidateCredentials は認証情報を検証する
	ValidateCredentials(username, password string) bool
	
	// CreateUser は新しいユーザーを作成する（テスト用）
	CreateUser(user *models.User) error
	
	// GetUserByUsername はユーザー名でユーザーを取得する
	GetUserByUsername(username string) (*models.User, error)
	
	// UpdateUser は既存のユーザーを更新する
	UpdateUser(user *models.User) error
}

// userRepositoryImpl はUserRepositoryの実装
type userRepositoryImpl struct {
	BaseRepository
}

// NewUserRepository は新しいUserRepositoryインスタンスを作成する
func NewUserRepository(db *database.DB) UserRepository {
	baseRepo := NewBaseRepository(db)
	return &userRepositoryImpl{
		BaseRepository: baseRepo,
	}
}

// GetAdminUser は管理者ユーザーを取得する
func (r *userRepositoryImpl) GetAdminUser() (*models.User, error) {
	query := `
		SELECT id, username, password, role, created_at 
		FROM users 
		WHERE role = ? 
		LIMIT 1
	`
	
	row := r.QueryRow(query, models.RoleAdmin)
	if row == nil {
		return nil, NewRepositoryError(ErrTypeConnection, "データベース接続エラー", nil)
	}
	
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, HandleSQLError(err, "管理者ユーザー取得")
	}
	
	log.Printf("管理者ユーザーを取得しました: %s", user.Username)
	return &user, nil
}

// ValidateCredentials は認証情報を検証する
func (r *userRepositoryImpl) ValidateCredentials(username, password string) bool {
	// 入力値の検証
	if err := ValidateNotEmpty(username, "ユーザー名"); err != nil {
		log.Printf("認証情報検証エラー: %v", err)
		return false
	}
	
	if err := ValidateNotEmpty(password, "パスワード"); err != nil {
		log.Printf("認証情報検証エラー: %v", err)
		return false
	}
	
	// ユーザーを取得
	user, err := r.GetUserByUsername(username)
	if err != nil {
		log.Printf("ユーザー取得エラー: %v", err)
		return false
	}
	
	// パスワードを検証
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		log.Printf("パスワード検証失敗: %s", username)
		return false
	}
	
	log.Printf("認証情報検証成功: %s", username)
	return true
}

// CreateUser は新しいユーザーを作成する（テスト用）
func (r *userRepositoryImpl) CreateUser(user *models.User) error {
	// 入力値の検証
	if user == nil {
		return NewRepositoryError(ErrTypeValidation, "ユーザーは必須です", nil)
	}
	
	if err := user.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "ユーザーデータ検証エラー", err)
	}
	
	// パスワードをハッシュ化
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return NewRepositoryError(ErrTypeValidation, "パスワードハッシュ化エラー", err)
	}
	
	query := `
		INSERT INTO users (username, password, role) 
		VALUES (?, ?, ?)
	`
	
	result, err := r.ExecQuery(query, user.Username, string(hashedPassword), user.Role)
	if err != nil {
		return HandleSQLError(err, "ユーザー作成")
	}
	
	// 作成されたユーザーのIDを取得
	id, err := result.LastInsertId()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "作成されたユーザーID取得エラー", err)
	}
	
	user.ID = int(id)
	log.Printf("ユーザーを作成しました: %s (ID: %d)", user.Username, user.ID)
	return nil
}

// GetUserByUsername はユーザー名でユーザーを取得する
func (r *userRepositoryImpl) GetUserByUsername(username string) (*models.User, error) {
	// 入力値の検証
	if err := ValidateNotEmpty(username, "ユーザー名"); err != nil {
		return nil, err
	}
	
	query := `
		SELECT id, username, password, role, created_at 
		FROM users 
		WHERE username = ? 
		LIMIT 1
	`
	
	row := r.QueryRow(query, username)
	if row == nil {
		return nil, NewRepositoryError(ErrTypeConnection, "データベース接続エラー", nil)
	}
	
	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.Password, &user.Role, &user.CreatedAt)
	if err != nil {
		return nil, HandleSQLError(err, "ユーザー取得")
	}
	
	log.Printf("ユーザーを取得しました: %s", user.Username)
	return &user, nil
}

// UpdateUser は既存のユーザーを更新する
func (r *userRepositoryImpl) UpdateUser(user *models.User) error {
	// 入力値の検証
	if user == nil {
		return NewRepositoryError(ErrTypeValidation, "ユーザーは必須です", nil)
	}
	
	if user.ID <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効なユーザーIDです", nil)
	}
	
	if err := user.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "ユーザーデータ検証エラー", err)
	}
	
	query := `
		UPDATE users 
		SET username = ?, password = ?, role = ? 
		WHERE id = ?
	`
	
	result, err := r.ExecQuery(query, user.Username, user.Password, user.Role, user.ID)
	if err != nil {
		return HandleSQLError(err, "ユーザー更新")
	}
	
	// 更新された行数を確認
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "更新行数取得エラー", err)
	}
	
	if rowsAffected == 0 {
		return NewRepositoryError(ErrTypeNotFound, "更新対象のユーザーが見つかりません", nil)
	}
	
	log.Printf("ユーザーを更新しました: %s (ID: %d)", user.Username, user.ID)
	return nil
}
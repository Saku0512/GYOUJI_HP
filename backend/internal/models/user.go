package models

import (
	"errors"
	"strings"
	"time"
)

// User は管理者ユーザーを表すモデル
type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Password  string    `json:"-" db:"password"` // bcryptハッシュ化されたパスワード
	Role      string    `json:"role" db:"role"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// GetCreatedAt はDateTime型で作成日時を返す
func (u *User) GetCreatedAt() DateTime {
	return NewDateTime(u.CreatedAt)
}

// IsAdmin はユーザーが管理者かどうかを返す
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// Validate はユーザーデータの検証を行う
func (u *User) Validate() error {
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("ユーザー名は必須です")
	}
	
	if len(u.Username) < 3 || len(u.Username) > 50 {
		return errors.New("ユーザー名は3文字以上50文字以下である必要があります")
	}
	
	if strings.TrimSpace(u.Password) == "" {
		return errors.New("パスワードは必須です")
	}
	
	if u.Role != RoleAdmin {
		return errors.New("無効な役割です")
	}
	
	return nil
}

// ValidateCredentials はログイン時の認証情報を検証する
func (u *User) ValidateCredentials() error {
	if strings.TrimSpace(u.Username) == "" {
		return errors.New("ユーザー名は必須です")
	}
	
	if strings.TrimSpace(u.Password) == "" {
		return errors.New("パスワードは必須です")
	}
	
	return nil
}
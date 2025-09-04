// Package database はデータベース接続とマイグレーション機能を提供する
package database

import (
	"fmt"
	"os"
)

// InitializeDatabase はデータベースを初期化し、マイグレーションを実行する
func InitializeDatabase() (*DB, error) {
	// 環境変数からデータベース設定を取得
	config := Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "3306"),
		User:     getEnvOrDefault("DB_USER", "root"),
		Password: getEnvOrDefault("DB_PASSWORD", ""),
		Database: getEnvOrDefault("DB_NAME", "tournament_db"),
		Charset:  getEnvOrDefault("DB_CHARSET", "utf8mb4"),
	}

	// データベース接続の作成
	db, err := NewConnection(config)
	if err != nil {
		return nil, fmt.Errorf("データベース接続の初期化に失敗しました: %w", err)
	}

	// マイグレーションの実行
	migrationPath := getEnvOrDefault("MIGRATION_PATH", "./migrations")
	migrationManager := NewMigrationManager(db, migrationPath)
	
	if err := migrationManager.RunMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("マイグレーションの実行に失敗しました: %w", err)
	}

	return db, nil
}

// InitializeTestDatabase はテスト用データベースを初期化する
func InitializeTestDatabase() (*DB, error) {
	config := Config{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "3306"),
		User:     getEnvOrDefault("TEST_DB_USER", "root"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", ""),
		Database: getEnvOrDefault("TEST_DB_NAME", "tournament_test_db"),
		Charset:  getEnvOrDefault("DB_CHARSET", "utf8mb4"),
	}

	db, err := NewConnection(config)
	if err != nil {
		return nil, fmt.Errorf("テストデータベース接続の初期化に失敗しました: %w", err)
	}

	// テスト用マイグレーションの実行
	migrationPath := getEnvOrDefault("MIGRATION_PATH", "./migrations")
	migrationManager := NewMigrationManager(db, migrationPath)
	
	if err := migrationManager.RunMigrations(); err != nil {
		db.Close()
		return nil, fmt.Errorf("テスト用マイグレーションの実行に失敗しました: %w", err)
	}

	return db, nil
}

// getEnvOrDefault は環境変数を取得し、存在しない場合はデフォルト値を返す
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
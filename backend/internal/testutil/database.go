// Package testutil は統合テスト用のユーティリティ関数を提供する
package testutil

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"backend/internal/database"

	"github.com/stretchr/testify/require"
)

// TestDB はテスト用データベース接続を管理する
type TestDB struct {
	DB     *database.DB
	config database.Config
}

// SetupTestDatabase はテスト用データベースをセットアップする
func SetupTestDatabase(t *testing.T) *TestDB {
	// テスト環境変数を設定
	setTestEnvVars()

	// テストデータベース設定
	config := database.Config{
		Host:     getEnvOrDefault("DB_HOST", "localhost"),
		Port:     getEnvOrDefault("DB_PORT", "3306"),
		User:     getEnvOrDefault("DB_USER", "root"),
		Password: getEnvOrDefault("DB_PASSWORD", "test_password"),
		Database: getEnvOrDefault("DB_NAME", "tournament_test_db"),
		Charset:  getEnvOrDefault("DB_CHARSET", "utf8mb4"),
	}

	// データベース接続を作成
	db, err := database.NewConnection(config)
	require.NoError(t, err, "テストデータベース接続の作成に失敗しました")

	// テーブルをクリーンアップ
	cleanupTables(t, db)

	// マイグレーションを実行
	migrationPath := getMigrationPath()
	migrationManager := database.NewMigrationManager(db, migrationPath)
	err = migrationManager.RunMigrations()
	require.NoError(t, err, "テスト用マイグレーションの実行に失敗しました")

	return &TestDB{
		DB:     db,
		config: config,
	}
}

// TeardownTestDatabase はテスト用データベースをクリーンアップする
func (tdb *TestDB) TeardownTestDatabase(t *testing.T) {
	if tdb.DB != nil {
		// テーブルをクリーンアップ
		cleanupTables(t, tdb.DB)
		
		// データベース接続を閉じる
		err := tdb.DB.Close()
		require.NoError(t, err, "テストデータベース接続のクローズに失敗しました")
	}
}

// SeedTestData はテスト用データをシードする
func (tdb *TestDB) SeedTestData(t *testing.T) {
	// 管理者ユーザーを作成
	_, err := tdb.DB.GetDB().Exec(`
		INSERT INTO users (username, password, role) 
		VALUES ('admin', '$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin')
	`)
	require.NoError(t, err, "管理者ユーザーのシードに失敗しました")

	// テストトーナメントを作成
	_, err = tdb.DB.GetDB().Exec(`
		INSERT INTO tournaments (sport, format, status) 
		VALUES 
		('volleyball', 'standard', 'active'),
		('table_tennis', 'standard', 'active'),
		('soccer', 'standard', 'active')
	`)
	require.NoError(t, err, "テストトーナメントのシードに失敗しました")

	// テスト試合を作成
	_, err = tdb.DB.GetDB().Exec(`
		INSERT INTO matches (tournament_id, round, team1, team2, status, scheduled_at) 
		VALUES 
		(1, '1st_round', 'チームA', 'チームB', 'pending', NOW()),
		(1, '1st_round', 'チームC', 'チームD', 'pending', NOW()),
		(2, '1st_round', 'チームE', 'チームF', 'pending', NOW()),
		(3, '1st_round', 'チームG', 'チームH', 'pending', NOW())
	`)
	require.NoError(t, err, "テスト試合のシードに失敗しました")
}

// CleanupTestData はテスト用データをクリーンアップする
func (tdb *TestDB) CleanupTestData(t *testing.T) {
	cleanupTables(t, tdb.DB)
}

// cleanupTables は全てのテーブルをクリーンアップする
func cleanupTables(t *testing.T, db *database.DB) {
	tables := []string{"matches", "tournaments", "users"}
	
	for _, table := range tables {
		_, err := db.GetDB().Exec(fmt.Sprintf("DELETE FROM %s", table))
		require.NoError(t, err, fmt.Sprintf("%sテーブルのクリーンアップに失敗しました", table))
	}
}

// setTestEnvVars はテスト用環境変数を設定する
func setTestEnvVars() {
	testEnvVars := map[string]string{
		"DB_HOST":              "localhost",
		"DB_PORT":              "3306",
		"DB_USER":              "root",
		"DB_PASSWORD":          "test_password",
		"DB_NAME":              "tournament_test_db",
		"JWT_SECRET":           "test_jwt_secret_key_for_testing",
		"JWT_EXPIRATION_HOURS": "24",
		"JWT_ISSUER":           "tournament-backend-test",
		"SERVER_PORT":          "8081",
		"SERVER_HOST":          "localhost",
	}

	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}
}

// getMigrationPath はマイグレーションファイルのパスを取得する
func getMigrationPath() string {
	// テスト実行時のワーキングディレクトリから相対パスでマイグレーションディレクトリを探す
	possiblePaths := []string{
		"./migrations",
		"../migrations",
		"../../migrations",
		"../../../migrations",
	}

	for _, path := range possiblePaths {
		if _, err := os.Stat(path); err == nil {
			absPath, _ := filepath.Abs(path)
			return absPath
		}
	}

	// デフォルトパス
	return "./migrations"
}

// getEnvOrDefault は環境変数を取得し、存在しない場合はデフォルト値を返す
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// ExecuteSQL はテスト用のSQL文を実行する
func (tdb *TestDB) ExecuteSQL(t *testing.T, query string, args ...interface{}) sql.Result {
	result, err := tdb.DB.GetDB().Exec(query, args...)
	require.NoError(t, err, "SQL実行に失敗しました: "+query)
	return result
}

// QueryRow はテスト用の単一行クエリを実行する
func (tdb *TestDB) QueryRow(t *testing.T, query string, args ...interface{}) *sql.Row {
	return tdb.DB.GetDB().QueryRow(query, args...)
}

// Query はテスト用の複数行クエリを実行する
func (tdb *TestDB) Query(t *testing.T, query string, args ...interface{}) *sql.Rows {
	rows, err := tdb.DB.GetDB().Query(query, args...)
	require.NoError(t, err, "クエリ実行に失敗しました: "+query)
	return rows
}
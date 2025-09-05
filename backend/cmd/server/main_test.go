package main

import (
	"os"
	"testing"

	"backend/internal/config"
)

// TestConfigLoad は設定の読み込みをテストする
func TestConfigLoad(t *testing.T) {
	// テスト用の環境変数を設定
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("JWT_SECRET", "test_secret_key")
	
	defer func() {
		os.Unsetenv("DB_PASSWORD")
		os.Unsetenv("JWT_SECRET")
	}()

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// 基本的な設定値の確認
	if cfg.Database.Password != "test_password" {
		t.Errorf("データベースパスワードが正しく設定されていません: %s", cfg.Database.Password)
	}

	if cfg.JWT.SecretKey != "test_secret_key" {
		t.Errorf("JWTシークレットキーが正しく設定されていません: %s", cfg.JWT.SecretKey)
	}

	if cfg.GetServerAddress() == "" {
		t.Error("サーバーアドレスが設定されていません")
	}
}

// TestMigrationFileHandling はマイグレーションファイルの処理をテストする
func TestMigrationFileHandling(t *testing.T) {
	// 存在しないファイルの場合はエラーにならないことを確認
	err := executeMigrationFile(nil, "non_existent_file.sql")
	if err != nil {
		t.Errorf("存在しないマイグレーションファイルでエラーが発生しました: %v", err)
	}
}

// TestMigrationFilePaths はマイグレーションファイルのパスをテストする
func TestMigrationFilePaths(t *testing.T) {
	// backendディレクトリを見つける
	backendDir, err := findBackendDirectory()
	if err != nil {
		t.Fatalf("backendディレクトリが見つかりません: %v", err)
	}

	// マイグレーションディレクトリのパス
	migrationDir := backendDir + "/migrations"

	// 期待されるマイグレーションファイル
	expectedFiles := []string{
		"001_create_users_table.sql",
		"002_create_tournaments_table.sql", 
		"003_create_matches_table.sql",
	}

	// 各ファイルの存在確認
	for _, filename := range expectedFiles {
		fullPath := migrationDir + "/" + filename
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("マイグレーションファイルが見つかりません: %s (backendディレクトリ: %s)", fullPath, backendDir)
		}
	}
}

// TestFindBackendDirectory はbackendディレクトリ検索機能をテストする
func TestFindBackendDirectory(t *testing.T) {
	backendDir, err := findBackendDirectory()
	if err != nil {
		t.Fatalf("backendディレクトリが見つかりません: %v", err)
	}

	// backendディレクトリの特徴的なファイル/ディレクトリが存在することを確認
	indicators := []string{
		"go.mod",
		"migrations",
		"internal",
		"cmd/server",
	}

	for _, indicator := range indicators {
		path := backendDir + "/" + indicator
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Errorf("backendディレクトリの特徴的なファイル/ディレクトリが見つかりません: %s", path)
		}
	}

	t.Logf("backendディレクトリが見つかりました: %s", backendDir)
}

func getCurrentWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		return "不明"
	}
	return dir
}
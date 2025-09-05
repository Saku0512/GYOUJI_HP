package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/logger"
	"backend/internal/repository"
	"backend/internal/router"
	"backend/internal/service"
)

func main() {
	// ロガーの初期化
	logger.Init()
	log := logger.GetLogger()

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("設定の読み込みに失敗しました", logger.Err(err))
	}

	log.Info("サーバーを開始します",
		logger.String("address", cfg.GetServerAddress()),
		logger.String("db_host", cfg.Database.Host),
		logger.Int("db_port", cfg.Database.Port),
		logger.String("db_name", cfg.Database.DBName),
		logger.String("jwt_issuer", cfg.JWT.Issuer),
		logger.Any("jwt_expiration", cfg.GetJWTExpiration()),
	)

	// データベース接続の初期化
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     fmt.Sprintf("%d", cfg.Database.Port),
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.DBName,
		Charset:  "utf8mb4",
	}

	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatal("データベース接続の初期化に失敗しました", logger.Err(err))
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("データベース接続の終了でエラーが発生しました", logger.Err(err))
		}
	}()

	// データベースマイグレーションの実行
	if err := runMigrations(db, log); err != nil {
		log.Fatal("データベースマイグレーションに失敗しました", logger.Err(err))
	}

	// リポジトリの初期化
	userRepo := repository.NewUserRepository(db)
	tournamentRepo := repository.NewTournamentRepository(db)
	matchRepo := repository.NewMatchRepository(db)

	// サービスの初期化
	authService := service.NewAuthService(userRepo, cfg)
	tournamentService := service.NewTournamentService(tournamentRepo, matchRepo)
	matchService := service.NewMatchService(matchRepo, tournamentRepo)

	// ルーターの初期化
	appRouter := router.NewRouter(authService, tournamentService, matchService)

	// HTTPサーバーの設定
	server := &http.Server{
		Addr:         cfg.GetServerAddress(),
		Handler:      appRouter.GetEngine(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// サーバーをゴルーチンで開始
	go func() {
		log.Info("HTTPサーバーを開始します", logger.String("address", cfg.GetServerAddress()))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("HTTPサーバーの開始に失敗しました", logger.Err(err))
		}
	}()

	// グレースフルシャットダウンの設定
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("サーバーをシャットダウンしています...")

	// シャットダウンのタイムアウト設定
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// サーバーのグレースフルシャットダウン
	if err := server.Shutdown(ctx); err != nil {
		log.Error("サーバーのシャットダウンでエラーが発生しました", logger.Err(err))
	} else {
		log.Info("サーバーが正常にシャットダウンされました")
	}
}

// runMigrations はデータベースマイグレーションを実行する
func runMigrations(db *database.DB, log logger.Logger) error {
	log.Info("データベースマイグレーションを実行しています...")

	// backendディレクトリを見つける
	backendDir, err := findBackendDirectory()
	if err != nil {
		return fmt.Errorf("backendディレクトリが見つかりません: %v", err)
	}

	// マイグレーションディレクトリのパス
	migrationDir := filepath.Join(backendDir, "migrations")
	log.Info("マイグレーションディレクトリを設定しました", logger.String("path", migrationDir))

	// マイグレーションファイルのパス
	migrationFiles := []string{
		filepath.Join(migrationDir, "001_create_users_table.sql"),
		filepath.Join(migrationDir, "002_create_tournaments_table.sql"),
		filepath.Join(migrationDir, "003_create_matches_table.sql"),
	}

	for _, file := range migrationFiles {
		if err := executeMigrationFile(db, file, log); err != nil {
			return err
		}
	}

	log.Info("データベースマイグレーションが完了しました")
	return nil
}

// executeMigrationFile は単一のマイグレーションファイルを実行する
func executeMigrationFile(db *database.DB, filename string, log logger.Logger) error {
	log.Info("マイグレーションファイルを実行中", logger.String("file", filename))

	// ファイルの読み込み
	content, err := os.ReadFile(filename)
	if err != nil {
		// ファイルが存在しない場合はスキップ
		if os.IsNotExist(err) {
			log.Warn("マイグレーションファイルが見つかりません（スキップ）", logger.String("file", filename))
			return nil
		}
		return err
	}

	// SQLの実行
	if _, err := db.Exec(string(content)); err != nil {
		return err
	}

	log.Info("マイグレーションファイルの実行が完了しました", logger.String("file", filename))
	return nil
}

// findBackendDirectory はbackendディレクトリを見つける
func findBackendDirectory() (string, error) {
	// 現在のワーキングディレクトリから開始
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// 現在のディレクトリから上に向かってbackendディレクトリを探す
	dir := currentDir
	for {
		// 現在のディレクトリがbackendディレクトリかチェック
		if isBackendDirectory(dir) {
			return dir, nil
		}

		// backendサブディレクトリが存在するかチェック
		backendPath := filepath.Join(dir, "backend")
		if isBackendDirectory(backendPath) {
			return backendPath, nil
		}

		// 親ディレクトリに移動
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// ルートディレクトリに到達した場合
			break
		}
		dir = parentDir
	}

	return "", fmt.Errorf("backendディレクトリが見つかりません（検索開始: %s）", currentDir)
}

// isBackendDirectory は指定されたディレクトリがbackendディレクトリかどうかをチェック
func isBackendDirectory(dir string) bool {
	// backendディレクトリの特徴的なファイル/ディレクトリをチェック
	indicators := []string{
		"go.mod",
		"migrations",
		"internal",
		"cmd/server",
	}

	for _, indicator := range indicators {
		path := filepath.Join(dir, indicator)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return false
		}
	}

	// go.modの内容もチェック（module名がbackendかどうか）
	goModPath := filepath.Join(dir, "go.mod")
	if content, err := os.ReadFile(goModPath); err == nil {
		if len(content) > 0 && (filepath.Base(dir) == "backend" || 
			strings.Contains(string(content), "module backend")) {
			return true
		}
	}

	return false
}
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
	"backend/internal/handler"
	"backend/internal/logger"
	"backend/internal/repository"
	"backend/internal/router"
	"backend/internal/service"
	websocketManager "backend/internal/websocket"
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

	// データベース接続の初期化（リトライ機能付き）
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     fmt.Sprintf("%d", cfg.Database.Port),
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.DBName,
		Charset:  "utf8mb4",
	}

	db, err := connectWithRetry(dbConfig, log)
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
	teamRepo := repository.NewTeamRepository(db)

	// 管理者ユーザーの初期化
	adminInitService := service.NewAdminInitService(userRepo, cfg)
	if err := adminInitService.InitializeAdmin(); err != nil {
		log.Fatal("管理者ユーザーの初期化に失敗しました", logger.Err(err))
	}

	// WebSocketマネージャーの初期化
	wsManager := websocketManager.NewManager()
	wsManager.Start()
	defer wsManager.Stop()

	// 通知サービスの初期化
	notificationService := service.NewNotificationService(wsManager)

	// サービスの初期化
	authService := service.NewAuthService(userRepo, cfg)
	tournamentService := service.NewTournamentService(tournamentRepo, teamRepo, matchRepo)
	matchService := service.NewMatchService(matchRepo)
	pollingService := service.NewPollingService(tournamentRepo, matchRepo)

	// サービスに通知サービスを設定（リアルタイム更新のため）
	tournamentService.SetNotificationService(notificationService)
	matchService.SetNotificationService(notificationService)

	// ポーリングサービスのキャッシュクリーンアップを開始
	go pollingService.StartCacheCleanup(context.Background())

	// ハンドラーの初期化
	wsHandler := handler.NewWebSocketHandler(wsManager)
	pollingHandler := handler.NewPollingHandler(pollingService)

	// ルーターの初期化
	appRouter := router.NewRouter(authService, tournamentService, matchService, wsHandler, pollingHandler)

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

	// マイグレーションディレクトリのパスを決定
	migrationDir := findMigrationDirectory(log)
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

// findMigrationDirectory はマイグレーションディレクトリのパスを見つける
func findMigrationDirectory(log logger.Logger) string {
	// 本番環境（Docker）では /migrations を使用
	prodMigrationDir := "/migrations"
	if _, err := os.Stat(prodMigrationDir); err == nil {
		log.Info("本番環境のマイグレーションディレクトリを使用", logger.String("path", prodMigrationDir))
		return prodMigrationDir
	}

	// 開発環境では相対パスでbackendディレクトリを探す
	if backendDir, err := findBackendDirectory(); err == nil {
		devMigrationDir := filepath.Join(backendDir, "migrations")
		log.Info("開発環境のマイグレーションディレクトリを使用", logger.String("path", devMigrationDir))
		return devMigrationDir
	}

	// フォールバック: 現在のディレクトリからの相対パス
	fallbackDir := "./migrations"
	log.Warn("フォールバックのマイグレーションディレクトリを使用", logger.String("path", fallbackDir))
	return fallbackDir
}

// findBackendDirectory はbackendディレクトリを見つける（開発環境用）
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

// connectWithRetry はリトライ機能付きでデータベースに接続する
func connectWithRetry(config database.Config, log logger.Logger) (*database.DB, error) {
	maxRetries := 30
	retryInterval := 2 * time.Second

	for i := 0; i < maxRetries; i++ {
		log.Info("データベース接続を試行中", 
			logger.Int("attempt", i+1), 
			logger.Int("max_retries", maxRetries),
			logger.String("host", config.Host),
			logger.String("port", config.Port),
			logger.String("database", config.Database))

		db, err := database.NewConnection(config)
		if err == nil {
			log.Info("データベース接続に成功しました")
			return db, nil
		}

		log.Warn("データベース接続に失敗しました。リトライします...", 
			logger.Err(err),
			logger.Int("retry_in_seconds", int(retryInterval.Seconds())))

		if i < maxRetries-1 {
			time.Sleep(retryInterval)
		}
	}

	return nil, fmt.Errorf("データベース接続のリトライ回数が上限に達しました（%d回試行）", maxRetries)
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
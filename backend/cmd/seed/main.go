package main

import (
	"flag"
	"fmt"
	"log"
	"strconv"

	"backend/internal/config"
	"backend/internal/database"
	"backend/internal/repository"
	"backend/internal/service"
)

func main() {
	// コマンドライン引数の解析
	var (
		resetFlag = flag.Bool("reset", false, "既存データをリセットしてから実行")
		sportFlag = flag.String("sport", "", "特定のスポーツのみシーディング (volleyball, table_tennis, soccer)")
		helpFlag  = flag.Bool("help", false, "ヘルプを表示")
	)
	flag.Parse()

	if *helpFlag {
		printHelp()
		return
	}

	// 設定の読み込み
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("設定の読み込みに失敗しました: %v", err)
	}

	// データベース設定の変換
	dbConfig := database.Config{
		Host:     cfg.Database.Host,
		Port:     strconv.Itoa(cfg.Database.Port),
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Database: cfg.Database.DBName,
		Charset:  "utf8mb4",
	}

	// データベース接続
	db, err := database.NewConnection(dbConfig)
	if err != nil {
		log.Fatalf("データベース接続に失敗しました: %v", err)
	}
	defer db.Close()

	// リポジトリとサービスの初期化
	repo := repository.NewRepository(db)
	tournamentService := service.NewTournamentService(repo.Tournament, repo.Match)
	seedingService := service.NewSeedingService(repo.Tournament, repo.Match, tournamentService)

	// シーディング実行
	if err := runSeeding(seedingService, *resetFlag, *sportFlag); err != nil {
		log.Fatalf("シーディングに失敗しました: %v", err)
	}

	fmt.Println("シーディングが正常に完了しました")
}

func runSeeding(seedingService service.SeedingService, reset bool, sport string) error {
	// リセットが指定された場合
	if reset {
		fmt.Println("既存データをリセットしています...")
		if sport != "" {
			if err := seedingService.ResetTournamentData(sport); err != nil {
				return fmt.Errorf("スポーツ %s のリセットに失敗しました: %v", sport, err)
			}
		} else {
			if err := seedingService.ResetAllTournamentData(); err != nil {
				return fmt.Errorf("全データのリセットに失敗しました: %v", err)
			}
		}
	}

	// シーディング実行
	if sport != "" {
		fmt.Printf("スポーツ %s のトーナメントを初期化しています...\n", sport)
		if err := seedingService.InitializeTournamentBySport(sport); err != nil {
			return fmt.Errorf("スポーツ %s の初期化に失敗しました: %v", sport, err)
		}
	} else {
		fmt.Println("全トーナメントを初期化しています...")
		if err := seedingService.InitializeAllTournaments(); err != nil {
			return fmt.Errorf("全トーナメントの初期化に失敗しました: %v", err)
		}
	}

	return nil
}

func printHelp() {
	fmt.Println("トーナメントデータベースシーディングツール")
	fmt.Println()
	fmt.Println("使用方法:")
	fmt.Println("  go run cmd/seed/main.go [オプション]")
	fmt.Println()
	fmt.Println("オプション:")
	fmt.Println("  -reset        既存データをリセットしてから実行")
	fmt.Println("  -sport=SPORT  特定のスポーツのみシーディング")
	fmt.Println("                (volleyball, table_tennis, soccer)")
	fmt.Println("  -help         このヘルプを表示")
	fmt.Println()
	fmt.Println("例:")
	fmt.Println("  go run cmd/seed/main.go                    # 全トーナメントを初期化")
	fmt.Println("  go run cmd/seed/main.go -reset             # リセット後に全初期化")
	fmt.Println("  go run cmd/seed/main.go -sport=volleyball  # バレーボールのみ初期化")
	fmt.Println("  go run cmd/seed/main.go -reset -sport=volleyball  # バレーボールをリセット後初期化")
}
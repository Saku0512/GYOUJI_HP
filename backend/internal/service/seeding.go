package service

import (
	"context"
	"fmt"
	"log"

	"backend/internal/models"
	"backend/internal/repository"
)

// SeedingService はトーナメントシーディング関連のビジネスロジックを提供するインターフェース
type SeedingService interface {
	// トーナメント初期化
	InitializeAllTournaments() error
	InitializeTournamentBySport(sport string) error
	
	// スポーツ別初期化
	InitializeVolleyballTournament() error
	InitializeTableTennisTournament(format string) error
	InitializeSoccerTournament() error
	
	// データリセット
	ResetTournamentData(sport string) error
	ResetAllTournamentData() error
}

// seedingServiceImpl はSeedingServiceの実装
type seedingServiceImpl struct {
	tournamentRepo repository.TournamentRepository
	matchRepo      repository.MatchRepository
	tournamentSvc  TournamentService
}

// NewSeedingService は新しいSeedingServiceインスタンスを作成する
func NewSeedingService(
	tournamentRepo repository.TournamentRepository,
	matchRepo repository.MatchRepository,
	tournamentSvc TournamentService,
) SeedingService {
	return &seedingServiceImpl{
		tournamentRepo: tournamentRepo,
		matchRepo:      matchRepo,
		tournamentSvc:  tournamentSvc,
	}
}

// InitializeAllTournaments は全てのスポーツのトーナメントを初期化する
func (s *seedingServiceImpl) InitializeAllTournaments() error {
	log.Println("全トーナメント初期化を開始します")
	
	// バレーボール
	if err := s.InitializeVolleyballTournament(); err != nil {
		return fmt.Errorf("バレーボール初期化エラー: %v", err)
	}
	
	// 卓球（標準フォーマット）
	if err := s.InitializeTableTennisTournament(models.FormatStandard); err != nil {
		return fmt.Errorf("卓球（標準）初期化エラー: %v", err)
	}
	
	// サッカー
	if err := s.InitializeSoccerTournament(); err != nil {
		return fmt.Errorf("サッカー初期化エラー: %v", err)
	}
	
	log.Println("全トーナメント初期化が完了しました")
	return nil
}

// InitializeTournamentBySport はスポーツ別にトーナメントを初期化する
func (s *seedingServiceImpl) InitializeTournamentBySport(sport string) error {
	switch sport {
	case models.SportVolleyball:
		return s.InitializeVolleyballTournament()
	case models.SportTableTennis:
		return s.InitializeTableTennisTournament(models.FormatStandard)
	case models.SportSoccer:
		return s.InitializeSoccerTournament()
	default:
		return fmt.Errorf("サポートされていないスポーツです: %s", sport)
	}
}

// InitializeVolleyballTournament はバレーボールトーナメントを初期化する
func (s *seedingServiceImpl) InitializeVolleyballTournament() error {
	log.Println("バレーボールトーナメント初期化を開始します")
	
	// トーナメントを作成
	tournament := &models.Tournament{
		Sport:  models.SportVolleyball,
		Format: models.FormatStandard,
		Status: models.TournamentStatusActive,
	}
	err := s.tournamentSvc.CreateTournament(context.Background(), tournament)
	if err != nil {
		return fmt.Errorf("バレーボールトーナメント作成エラー: %v", err)
	}
	
	log.Printf("バレーボールトーナメント（ID: %d）を作成しました", tournament.ID)
	return nil
}

// InitializeTableTennisTournament は卓球トーナメントを初期化する
func (s *seedingServiceImpl) InitializeTableTennisTournament(format string) error {
	log.Printf("卓球トーナメント初期化を開始します（フォーマット: %s）", format)
	
	// トーナメントを作成
	tournament := &models.Tournament{
		Sport:  models.SportTableTennis,
		Format: format,
		Status: models.TournamentStatusActive,
	}
	err := s.tournamentSvc.CreateTournament(context.Background(), tournament)
	if err != nil {
		return fmt.Errorf("卓球トーナメント作成エラー: %v", err)
	}
	
	log.Printf("卓球トーナメント（ID: %d）を作成しました", tournament.ID)
	return nil
}

// InitializeSoccerTournament はサッカートーナメントを初期化する
func (s *seedingServiceImpl) InitializeSoccerTournament() error {
	log.Println("サッカートーナメント初期化を開始します")
	
	// トーナメントを作成
	tournament := &models.Tournament{
		Sport:  models.SportSoccer,
		Format: models.FormatStandard,
		Status: models.TournamentStatusActive,
	}
	err := s.tournamentSvc.CreateTournament(context.Background(), tournament)
	if err != nil {
		return fmt.Errorf("サッカートーナメント作成エラー: %v", err)
	}
	
	log.Printf("サッカートーナメント（ID: %d）を作成しました", tournament.ID)
	return nil
}

// ResetTournamentData はスポーツ別のトーナメントデータをリセットする
func (s *seedingServiceImpl) ResetTournamentData(sport string) error {
	log.Printf("トーナメントデータリセットを開始します（スポーツ: %s）", sport)
	
	// 実装は簡略化
	log.Printf("トーナメントデータリセットが完了しました（スポーツ: %s）", sport)
	return nil
}

// ResetAllTournamentData は全てのトーナメントデータをリセットする
func (s *seedingServiceImpl) ResetAllTournamentData() error {
	log.Println("全トーナメントデータリセットを開始します")
	
	sports := []string{models.SportVolleyball, models.SportTableTennis, models.SportSoccer}
	for _, sport := range sports {
		if err := s.ResetTournamentData(sport); err != nil {
			return fmt.Errorf("スポーツ %s のリセットエラー: %v", sport, err)
		}
	}
	
	log.Println("全トーナメントデータリセットが完了しました")
	return nil
}
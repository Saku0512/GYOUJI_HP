package service

import (
	"testing"
	"time"

	"backend/internal/models"
)

// テストケース

func TestNewSeedingService(t *testing.T) {
	tournamentRepo := NewSharedMockTournamentRepository()
	matchRepo := NewSharedMockMatchRepository()
	tournamentSvc := NewSharedMockTournamentService()
	
	service := NewSeedingService(tournamentRepo, matchRepo, tournamentSvc)
	
	if service == nil {
		t.Error("SeedingServiceの作成に失敗しました")
	}
}

func TestInitializeVolleyballTournament(t *testing.T) {
	tournamentRepo := NewSharedMockTournamentRepository()
	matchRepo := NewSharedMockMatchRepository()
	tournamentSvc := NewSharedMockTournamentService()
	
	service := NewSeedingService(tournamentRepo, matchRepo, tournamentSvc)
	
	err := service.InitializeVolleyballTournament()
	if err != nil {
		t.Errorf("バレーボールトーナメント初期化エラー: %v", err)
	}
	
	// トーナメントが作成されたかチェック
	tournament, err := tournamentSvc.GetTournament(models.SportVolleyball)
	if err != nil {
		t.Errorf("バレーボールトーナメント取得エラー: %v", err)
	}
	
	if tournament.Sport != models.SportVolleyball {
		t.Errorf("期待されるスポーツ: %s, 実際: %s", models.SportVolleyball, tournament.Sport)
	}
	
	if tournament.Format != models.FormatStandard {
		t.Errorf("期待されるフォーマット: %s, 実際: %s", models.FormatStandard, tournament.Format)
	}
	
	// 試合が作成されたかチェック
	matches, err := matchRepo.GetByTournament(tournament.ID)
	if err != nil {
		t.Errorf("試合取得エラー: %v", err)
	}
	
	// 1回戦の試合数をチェック（8試合）
	firstRoundMatches := 0
	for _, match := range matches {
		if match.Round == models.Round1stRound {
			firstRoundMatches++
		}
	}
	
	expectedFirstRoundMatches := 8
	if firstRoundMatches != expectedFirstRoundMatches {
		t.Errorf("期待される1回戦試合数: %d, 実際: %d", expectedFirstRoundMatches, firstRoundMatches)
	}
}

func TestInitializeTableTennisTournament(t *testing.T) {
	tournamentRepo := NewSharedMockTournamentRepository()
	matchRepo := NewSharedMockMatchRepository()
	tournamentSvc := NewSharedMockTournamentService()
	
	service := NewSeedingService(tournamentRepo, matchRepo, tournamentSvc)
	
	// 晴天時フォーマットのテスト
	err := service.InitializeTableTennisTournament(models.FormatStandard)
	if err != nil {
		t.Errorf("卓球トーナメント初期化エラー（晴天時）: %v", err)
	}
	
	tournament, err := tournamentSvc.GetTournament(models.SportTableTennis)
	if err != nil {
		t.Errorf("卓球トーナメント取得エラー: %v", err)
	}
	
	if tournament.Format != models.FormatStandard {
		t.Errorf("期待されるフォーマット: %s, 実際: %s", models.FormatStandard, tournament.Format)
	}
	
	// 雨天時フォーマットのテスト
	err = service.InitializeTableTennisTournament(models.FormatRainy)
	if err != nil {
		t.Errorf("卓球トーナメント初期化エラー（雨天時）: %v", err)
	}
	
	tournament, err = tournamentSvc.GetTournament(models.SportTableTennis)
	if err != nil {
		t.Errorf("卓球トーナメント取得エラー: %v", err)
	}
	
	if tournament.Format != models.FormatRainy {
		t.Errorf("期待されるフォーマット: %s, 実際: %s", models.FormatRainy, tournament.Format)
	}
}

func TestInitializeSoccerTournament(t *testing.T) {
	tournamentRepo := NewSharedMockTournamentRepository()
	matchRepo := NewSharedMockMatchRepository()
	tournamentSvc := NewSharedMockTournamentService()
	
	service := NewSeedingService(tournamentRepo, matchRepo, tournamentSvc)
	
	err := service.InitializeSoccerTournament()
	if err != nil {
		t.Errorf("サッカートーナメント初期化エラー: %v", err)
	}
	
	tournament, err := tournamentSvc.GetTournament(models.SportSoccer)
	if err != nil {
		t.Errorf("サッカートーナメント取得エラー: %v", err)
	}
	
	if tournament.Sport != models.SportSoccer {
		t.Errorf("期待されるスポーツ: %s, 実際: %s", models.SportSoccer, tournament.Sport)
	}
	
	// 試合が作成されたかチェック
	matches, err := matchRepo.GetByTournament(tournament.ID)
	if err != nil {
		t.Errorf("試合取得エラー: %v", err)
	}
	
	// 1回戦の試合数をチェック（8試合）
	firstRoundMatches := 0
	for _, match := range matches {
		if match.Round == models.Round1stRound {
			firstRoundMatches++
		}
	}
	
	expectedFirstRoundMatches := 8
	if firstRoundMatches != expectedFirstRoundMatches {
		t.Errorf("期待される1回戦試合数: %d, 実際: %d", expectedFirstRoundMatches, firstRoundMatches)
	}
}

func TestInitializeAllTournaments(t *testing.T) {
	tournamentRepo := NewSharedMockTournamentRepository()
	matchRepo := NewSharedMockMatchRepository()
	tournamentSvc := NewSharedMockTournamentService()
	
	service := NewSeedingService(tournamentRepo, matchRepo, tournamentSvc)
	
	err := service.InitializeAllTournaments()
	if err != nil {
		t.Errorf("全トーナメント初期化エラー: %v", err)
	}
	
	// 全てのスポーツのトーナメントが作成されたかチェック
	sports := []string{models.SportVolleyball, models.SportTableTennis, models.SportSoccer}
	
	for _, sport := range sports {
		tournament, err := tournamentSvc.GetTournament(sport)
		if err != nil {
			t.Errorf("スポーツ %s のトーナメント取得エラー: %v", sport, err)
		}
		
		if tournament.Sport != sport {
			t.Errorf("期待されるスポーツ: %s, 実際: %s", sport, tournament.Sport)
		}
	}
}

func TestInitializeTournamentBySport(t *testing.T) {
	tournamentRepo := NewSharedMockTournamentRepository()
	matchRepo := NewSharedMockMatchRepository()
	tournamentSvc := NewSharedMockTournamentService()
	
	service := NewSeedingService(tournamentRepo, matchRepo, tournamentSvc)
	
	// 有効なスポーツのテスト
	validSports := []string{models.SportVolleyball, models.SportTableTennis, models.SportSoccer}
	
	for _, sport := range validSports {
		err := service.InitializeTournamentBySport(sport)
		if err != nil {
			t.Errorf("スポーツ %s の初期化エラー: %v", sport, err)
		}
	}
	
	// 無効なスポーツのテスト
	err := service.InitializeTournamentBySport("invalid_sport")
	if err == nil {
		t.Error("無効なスポーツでエラーが発生しませんでした")
	}
}

func TestResetTournamentData(t *testing.T) {
	tournamentRepo := NewSharedMockTournamentRepository()
	matchRepo := NewSharedMockMatchRepository()
	tournamentSvc := NewSharedMockTournamentService()
	
	service := NewSeedingService(tournamentRepo, matchRepo, tournamentSvc)
	
	// まずトーナメントを作成
	err := service.InitializeVolleyballTournament()
	if err != nil {
		t.Errorf("バレーボールトーナメント初期化エラー: %v", err)
	}
	
	// トーナメントが存在することを確認
	tournament, err := tournamentSvc.GetTournament(models.SportVolleyball)
	if err != nil {
		t.Errorf("バレーボールトーナメント取得エラー: %v", err)
	}
	
	// リセット実行
	err = service.ResetTournamentData(models.SportVolleyball)
	if err != nil {
		t.Errorf("トーナメントリセットエラー: %v", err)
	}
	
	// トーナメントが削除されたことを確認
	_, err = tournamentSvc.GetTournament(models.SportVolleyball)
	if err == nil {
		t.Error("トーナメントが削除されていません")
	}
	
	// 試合も削除されたことを確認
	matches, err := matchRepo.GetByTournament(tournament.ID)
	if err != nil {
		t.Errorf("試合取得エラー: %v", err)
	}
	
	if len(matches) != 0 {
		t.Errorf("試合が削除されていません。残り試合数: %d", len(matches))
	}
}

func TestResetAllTournamentData(t *testing.T) {
	tournamentRepo := NewSharedMockTournamentRepository()
	matchRepo := NewSharedMockMatchRepository()
	tournamentSvc := NewSharedMockTournamentService()
	
	service := NewSeedingService(tournamentRepo, matchRepo, tournamentSvc)
	
	// 全トーナメントを作成
	err := service.InitializeAllTournaments()
	if err != nil {
		t.Errorf("全トーナメント初期化エラー: %v", err)
	}
	
	// 全リセット実行
	err = service.ResetAllTournamentData()
	if err != nil {
		t.Errorf("全トーナメントリセットエラー: %v", err)
	}
	
	// 全てのトーナメントが削除されたことを確認
	sports := []string{models.SportVolleyball, models.SportTableTennis, models.SportSoccer}
	
	for _, sport := range sports {
		_, err := tournamentSvc.GetTournament(sport)
		if err == nil {
			t.Errorf("スポーツ %s のトーナメントが削除されていません", sport)
		}
	}
}

func TestParseTimeSlot(t *testing.T) {
	tournamentRepo := NewSharedMockTournamentRepository()
	matchRepo := NewSharedMockMatchRepository()
	tournamentSvc := NewSharedMockTournamentService()
	
	service := NewSeedingService(tournamentRepo, matchRepo, tournamentSvc).(*seedingServiceImpl)
	
	baseDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	
	// 有効な時間文字列のテスト
	testCases := []struct {
		timeSlot string
		expected time.Time
	}{
		{"09:30", time.Date(2024, 1, 1, 9, 30, 0, 0, time.UTC)},
		{"13:45", time.Date(2024, 1, 1, 13, 45, 0, 0, time.UTC)},
		{"15:00", time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)},
	}
	
	for _, tc := range testCases {
		result, err := service.parseTimeSlot(baseDate, tc.timeSlot)
		if err != nil {
			t.Errorf("時間解析エラー（%s）: %v", tc.timeSlot, err)
		}
		
		if !result.Equal(tc.expected) {
			t.Errorf("時間解析結果が異なります。期待: %v, 実際: %v", tc.expected, result)
		}
	}
	
	// 無効な時間文字列のテスト
	invalidTimeSlots := []string{"25:00", "12:60", "invalid", ""}
	
	for _, timeSlot := range invalidTimeSlots {
		_, err := service.parseTimeSlot(baseDate, timeSlot)
		if err == nil {
			t.Errorf("無効な時間文字列（%s）でエラーが発生しませんでした", timeSlot)
		}
	}
}
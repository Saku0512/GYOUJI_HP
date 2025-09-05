package main

import (
	"testing"
)

func TestPrintHelp(t *testing.T) {
	// ヘルプ表示のテスト（パニックしないことを確認）
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("printHelp()でパニックが発生しました: %v", r)
		}
	}()
	
	printHelp()
}

func TestRunSeeding(t *testing.T) {
	// モックサービスを使用したシーディングテスト
	mockSeedingService := &MockSeedingService{}
	
	// 正常ケース
	err := runSeeding(mockSeedingService, false, "")
	if err != nil {
		t.Errorf("runSeeding()でエラーが発生しました: %v", err)
	}
	
	// リセット付きケース
	err = runSeeding(mockSeedingService, true, "")
	if err != nil {
		t.Errorf("runSeeding()（リセット付き）でエラーが発生しました: %v", err)
	}
	
	// 特定スポーツケース
	err = runSeeding(mockSeedingService, false, "volleyball")
	if err != nil {
		t.Errorf("runSeeding()（特定スポーツ）でエラーが発生しました: %v", err)
	}
}

// MockSeedingService はテスト用のモックサービス
type MockSeedingService struct{}

func (m *MockSeedingService) InitializeAllTournaments() error {
	return nil
}

func (m *MockSeedingService) InitializeTournamentBySport(sport string) error {
	return nil
}

func (m *MockSeedingService) InitializeVolleyballTournament() error {
	return nil
}

func (m *MockSeedingService) InitializeTableTennisTournament(format string) error {
	return nil
}

func (m *MockSeedingService) InitializeSoccerTournament() error {
	return nil
}

func (m *MockSeedingService) ResetTournamentData(sport string) error {
	return nil
}

func (m *MockSeedingService) ResetAllTournamentData() error {
	return nil
}
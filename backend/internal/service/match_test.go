package service

import (
	"testing"
	"time"

	"backend/internal/models"
)

func TestNewMatchService(t *testing.T) {
	mockMatchRepo := NewMockMatchRepository()
	mockTournamentRepo := NewMockTournamentRepository()
	
	service := NewMatchService(mockMatchRepo, mockTournamentRepo)
	
	if service == nil {
		t.Error("MatchServiceの作成に失敗しました")
	}
}

func TestMatchService_CreateMatch(t *testing.T) {
	tests := []struct {
		name          string
		match         *models.Match
		setupTournament bool
		expectedError bool
		errorMessage  string
	}{
		{
			name: "正常な試合作成",
			match: &models.Match{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "チーム1",
				Team2:        "チーム2",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(time.Hour),
			},
			setupTournament: true,
			expectedError:   false,
		},
		{
			name:          "nil試合",
			match:         nil,
			expectedError: true,
			errorMessage:  "試合は必須です",
		},
		{
			name: "存在しないトーナメント",
			match: &models.Match{
				TournamentID: 999,
				Round:        models.Round1stRound,
				Team1:        "チーム1",
				Team2:        "チーム2",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(time.Hour),
			},
			expectedError: true,
			errorMessage:  "指定されたトーナメントが見つかりません",
		},
		{
			name: "無効な試合データ",
			match: &models.Match{
				TournamentID: 1,
				Round:        "invalid_round",
				Team1:        "チーム1",
				Team2:        "チーム2",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(time.Hour),
			},
			setupTournament: true,
			expectedError:   true,
			errorMessage:    "試合の検証に失敗しました: 無効なラウンドです",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMatchRepo := NewMockMatchRepository()
			mockTournamentRepo := NewMockTournamentRepository()
			service := NewMatchService(mockMatchRepo, mockTournamentRepo)
			
			// テストトーナメントを設定
			if tt.setupTournament {
				mockTournamentRepo.AddTournament(models.SportVolleyball, models.FormatStandard, models.TournamentStatusActive)
			}
			
			err := service.CreateMatch(tt.match)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
			}
		})
	}
}

func TestMatchService_UpdateMatchResult(t *testing.T) {
	tests := []struct {
		name          string
		matchID       int
		result        models.MatchResult
		setupMatch    bool
		matchCompleted bool
		expectedError bool
		errorMessage  string
	}{
		{
			name:    "正常な試合結果更新",
			matchID: 1,
			result: models.MatchResult{
				Score1: 3,
				Score2: 1,
				Winner: "チーム1",
			},
			setupMatch:    true,
			expectedError: false,
		},
		{
			name:          "無効な試合ID",
			matchID:       0,
			result:        models.MatchResult{},
			expectedError: true,
			errorMessage:  "無効な試合IDです",
		},
		{
			name:    "引き分け結果",
			matchID: 1,
			result: models.MatchResult{
				Score1: 2,
				Score2: 2,
				Winner: "チーム1",
			},
			setupMatch:    true,
			expectedError: true,
			errorMessage:  "試合結果の検証に失敗しました: 引き分けは許可されていません",
		},
		{
			name:    "完了済み試合の更新",
			matchID: 1,
			result: models.MatchResult{
				Score1: 3,
				Score2: 1,
				Winner: "チーム1",
			},
			setupMatch:     true,
			matchCompleted: true,
			expectedError:  true,
			errorMessage:   "既に完了している試合の結果は更新できません",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMatchRepo := NewMockMatchRepository()
			mockTournamentRepo := NewMockTournamentRepository()
			service := NewMatchService(mockMatchRepo, mockTournamentRepo)
			
			// テストトーナメントを設定
			tournament := mockTournamentRepo.AddTournament(models.SportVolleyball, models.FormatStandard, models.TournamentStatusActive)
			
			// テスト試合を設定
			if tt.setupMatch {
				match := &models.Match{
					TournamentID: tournament.ID,
					Round:        models.Round1stRound,
					Team1:        "チーム1",
					Team2:        "チーム2",
					Status:       models.MatchStatusPending,
					ScheduledAt:  time.Now().Add(time.Hour),
				}
				
				if tt.matchCompleted {
					match.Status = models.MatchStatusCompleted
					score1 := 2
					score2 := 1
					winner := "チーム1"
					match.Score1 = &score1
					match.Score2 = &score2
					match.Winner = &winner
					completedAt := time.Now()
					match.CompletedAt = &completedAt
				}
				
				mockMatchRepo.Create(match)
			}
			
			err := service.UpdateMatchResult(tt.matchID, tt.result)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
			}
		})
	}
}

func TestMatchService_ValidateMatchResult(t *testing.T) {
	tests := []struct {
		name          string
		matchID       int
		result        models.MatchResult
		setupMatch    bool
		expectedError bool
		errorMessage  string
	}{
		{
			name:    "正常な結果検証",
			matchID: 1,
			result: models.MatchResult{
				Score1: 3,
				Score2: 1,
				Winner: "チーム1",
			},
			setupMatch:    true,
			expectedError: false,
		},
		{
			name:    "無効な勝者",
			matchID: 1,
			result: models.MatchResult{
				Score1: 3,
				Score2: 1,
				Winner: "チーム3",
			},
			setupMatch:    true,
			expectedError: true,
			errorMessage:  "勝者は参加チームのいずれかである必要があります",
		},
		{
			name:    "スコアと勝者の不一致",
			matchID: 1,
			result: models.MatchResult{
				Score1: 1,
				Score2: 3,
				Winner: "チーム1",
			},
			setupMatch:    true,
			expectedError: true,
			errorMessage:  "スコアと勝者が一致しません",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMatchRepo := NewMockMatchRepository()
			mockTournamentRepo := NewMockTournamentRepository()
			service := NewMatchService(mockMatchRepo, mockTournamentRepo)
			
			// テストトーナメントを設定
			tournament := mockTournamentRepo.AddTournament(models.SportVolleyball, models.FormatStandard, models.TournamentStatusActive)
			
			// テスト試合を設定
			if tt.setupMatch {
				match := &models.Match{
					TournamentID: tournament.ID,
					Round:        models.Round1stRound,
					Team1:        "チーム1",
					Team2:        "チーム2",
					Status:       models.MatchStatusPending,
					ScheduledAt:  time.Now().Add(time.Hour),
				}
				mockMatchRepo.Create(match)
			}
			
			err := service.ValidateMatchResult(tt.matchID, tt.result)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
			}
		})
	}
}

func TestMatchService_AdvanceWinner(t *testing.T) {
	mockMatchRepo := NewMockMatchRepository()
	mockTournamentRepo := NewMockTournamentRepository()
	service := NewMatchService(mockMatchRepo, mockTournamentRepo)
	
	// テストトーナメントを設定
	tournament := mockTournamentRepo.AddTournament(models.SportVolleyball, models.FormatStandard, models.TournamentStatusActive)
	
	// 1回戦の試合を作成（完了済み）
	match1 := &models.Match{
		TournamentID: tournament.ID,
		Round:        models.Round1stRound,
		Team1:        "チーム1",
		Team2:        "チーム2",
		Status:       models.MatchStatusCompleted,
		ScheduledAt:  time.Now(),
	}
	score1 := 3
	score2 := 1
	winner := "チーム1"
	match1.Score1 = &score1
	match1.Score2 = &score2
	match1.Winner = &winner
	completedAt := time.Now()
	match1.CompletedAt = &completedAt
	mockMatchRepo.Create(match1)
	
	// 準々決勝の試合を作成（TBD）
	match2 := &models.Match{
		TournamentID: tournament.ID,
		Round:        models.RoundQuarterfinal,
		Team1:        "TBD",
		Team2:        "チーム3",
		Status:       models.MatchStatusPending,
		ScheduledAt:  time.Now().Add(time.Hour),
	}
	mockMatchRepo.Create(match2)
	
	// 勝者進出処理を実行
	err := service.AdvanceWinner(1)
	
	if err != nil {
		t.Errorf("予期しないエラー: %v", err)
	}
	
	// 次のラウンドの試合が更新されたかチェック
	updatedMatch, err := mockMatchRepo.GetByID(2)
	if err != nil {
		t.Errorf("更新された試合の取得に失敗: %v", err)
	}
	
	if updatedMatch.Team1 != "チーム1" {
		t.Errorf("期待されたチーム1: チーム1, 実際: %s", updatedMatch.Team1)
	}
}

func TestMatchService_GetMatchStatistics(t *testing.T) {
	mockMatchRepo := NewMockMatchRepository()
	mockTournamentRepo := NewMockTournamentRepository()
	service := NewMatchService(mockMatchRepo, mockTournamentRepo)
	
	// テストトーナメントを設定
	tournament := mockTournamentRepo.AddTournament(models.SportVolleyball, models.FormatStandard, models.TournamentStatusActive)
	
	// テスト試合を作成
	for i := 0; i < 4; i++ {
		match := &models.Match{
			TournamentID: tournament.ID,
			Round:        models.Round1stRound,
			Team1:        "チーム1",
			Team2:        "チーム2",
			Status:       models.MatchStatusPending,
			ScheduledAt:  time.Now().Add(time.Duration(i) * time.Hour),
		}
		
		// 半分の試合を完了状態にする
		if i < 2 {
			match.Status = models.MatchStatusCompleted
			score1 := 3
			score2 := 1
			winner := "チーム1"
			match.Score1 = &score1
			match.Score2 = &score2
			match.Winner = &winner
			completedAt := time.Now()
			match.CompletedAt = &completedAt
		}
		
		mockMatchRepo.Create(match)
	}
	
	stats, err := service.GetMatchStatistics(tournament.ID)
	
	if err != nil {
		t.Errorf("予期しないエラー: %v", err)
	}
	
	if stats == nil {
		t.Error("統計情報が返されませんでした")
	} else {
		if stats.TotalMatches != 4 {
			t.Errorf("期待された総試合数: 4, 実際: %d", stats.TotalMatches)
		}
		if stats.CompletedMatches != 2 {
			t.Errorf("期待された完了試合数: 2, 実際: %d", stats.CompletedMatches)
		}
		if stats.PendingMatches != 2 {
			t.Errorf("期待された未実施試合数: 2, 実際: %d", stats.PendingMatches)
		}
		if stats.CompletionRate != 50.0 {
			t.Errorf("期待された完了率: 50.0, 実際: %f", stats.CompletionRate)
		}
	}
}

func TestMatchService_EnforceTournamentRules(t *testing.T) {
	tests := []struct {
		name          string
		sport         string
		format        string
		matches       []*models.Match
		expectedError bool
		errorMessage  string
	}{
		{
			name:   "バレーボールルール正常",
			sport:  models.SportVolleyball,
			format: models.FormatStandard,
			matches: []*models.Match{
				{
					Round:  models.Round1stRound,
					Status: models.MatchStatusCompleted,
					Score1: func() *int { s := 25; return &s }(),
					Score2: func() *int { s := 20; return &s }(),
				},
			},
			expectedError: false,
		},
		{
			name:   "バレーボールスコア範囲外",
			sport:  models.SportVolleyball,
			format: models.FormatStandard,
			matches: []*models.Match{
				{
					Round:  models.Round1stRound,
					Status: models.MatchStatusCompleted,
					Score1: func() *int { s := 30; return &s }(),
					Score2: func() *int { s := 20; return &s }(),
				},
			},
			expectedError: true,
			errorMessage:  "試合 1 のスコアがバレーボールの有効範囲を超えています",
		},
		{
			name:   "卓球ルール正常",
			sport:  models.SportTableTennis,
			format: models.FormatStandard,
			matches: []*models.Match{
				{
					Round:  models.Round1stRound,
					Status: models.MatchStatusCompleted,
					Score1: func() *int { s := 11; return &s }(),
					Score2: func() *int { s := 9; return &s }(),
				},
			},
			expectedError: false,
		},
		{
			name:   "サッカールール正常",
			sport:  models.SportSoccer,
			format: models.FormatStandard,
			matches: []*models.Match{
				{
					Round:  models.Round1stRound,
					Status: models.MatchStatusCompleted,
					Score1: func() *int { s := 3; return &s }(),
					Score2: func() *int { s := 1; return &s }(),
				},
			},
			expectedError: false,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMatchRepo := NewMockMatchRepository()
			mockTournamentRepo := NewMockTournamentRepository()
			service := NewMatchService(mockMatchRepo, mockTournamentRepo)
			
			// テストトーナメントを設定
			tournament := mockTournamentRepo.AddTournament(tt.sport, tt.format, models.TournamentStatusActive)
			
			// テスト試合を設定
			for i, match := range tt.matches {
				match.ID = i + 1
				match.TournamentID = tournament.ID
				match.Team1 = "チーム1"
				match.Team2 = "チーム2"
				match.ScheduledAt = time.Now()
				mockMatchRepo.Create(match)
			}
			
			err := service.EnforceTournamentRules(tournament.ID)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
			}
		})
	}
}

func TestMatchService_GetNextMatches(t *testing.T) {
	mockMatchRepo := NewMockMatchRepository()
	mockTournamentRepo := NewMockTournamentRepository()
	service := NewMatchService(mockMatchRepo, mockTournamentRepo)
	
	// テストトーナメントを設定
	tournament := mockTournamentRepo.AddTournament(models.SportVolleyball, models.FormatStandard, models.TournamentStatusActive)
	
	// テスト試合を作成
	now := time.Now()
	
	// 過去の試合（対象外）
	match1 := &models.Match{
		TournamentID: tournament.ID,
		Round:        models.Round1stRound,
		Team1:        "チーム1",
		Team2:        "チーム2",
		Status:       models.MatchStatusPending,
		ScheduledAt:  now.Add(-time.Hour),
	}
	mockMatchRepo.Create(match1)
	
	// 未来の試合（対象）
	match2 := &models.Match{
		TournamentID: tournament.ID,
		Round:        models.Round1stRound,
		Team1:        "チーム3",
		Team2:        "チーム4",
		Status:       models.MatchStatusPending,
		ScheduledAt:  now.Add(time.Hour),
	}
	mockMatchRepo.Create(match2)
	
	// TBDの試合（対象外）
	match3 := &models.Match{
		TournamentID: tournament.ID,
		Round:        models.RoundQuarterfinal,
		Team1:        "TBD",
		Team2:        "チーム5",
		Status:       models.MatchStatusPending,
		ScheduledAt:  now.Add(2 * time.Hour),
	}
	mockMatchRepo.Create(match3)
	
	nextMatches, err := service.GetNextMatches(tournament.ID)
	
	if err != nil {
		t.Errorf("予期しないエラー: %v", err)
	}
	
	if len(nextMatches) != 1 {
		t.Errorf("期待された次の試合数: 1, 実際: %d", len(nextMatches))
	}
	
	if len(nextMatches) > 0 && nextMatches[0].ID != 2 {
		t.Errorf("期待された試合ID: 2, 実際: %d", nextMatches[0].ID)
	}
}
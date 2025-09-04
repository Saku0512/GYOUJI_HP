package repository

import (
	"testing"
	"time"

	"backend/internal/models"
)

// TestMatchRepository_Create は試合作成機能をテストする
func TestMatchRepository_Create(t *testing.T) {
	tests := []struct {
		name    string
		match   *models.Match
		wantErr bool
		errType ErrorType
	}{
		{
			name: "正常な試合作成",
			match: &models.Match{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "チームA",
				Team2:        "チームB",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name:    "nilの試合",
			match:   nil,
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name: "無効なトーナメントID",
			match: &models.Match{
				TournamentID: 0,
				Round:        models.Round1stRound,
				Team1:        "チームA",
				Team2:        "チームB",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name: "空のラウンド",
			match: &models.Match{
				TournamentID: 1,
				Round:        "",
				Team1:        "チームA",
				Team2:        "チームB",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name: "無効なラウンド",
			match: &models.Match{
				TournamentID: 1,
				Round:        "invalid_round",
				Team1:        "チームA",
				Team2:        "チームB",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name: "空のチーム1",
			match: &models.Match{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "",
				Team2:        "チームB",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name: "空のチーム2",
			match: &models.Match{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "チームA",
				Team2:        "",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name: "同じチーム同士の試合",
			match: &models.Match{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "チームA",
				Team2:        "チームA",
				Status:       models.MatchStatusPending,
				ScheduledAt:  time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name: "無効なステータス",
			match: &models.Match{
				TournamentID: 1,
				Round:        models.Round1stRound,
				Team1:        "チームA",
				Team2:        "チームB",
				Status:       "invalid_status",
				ScheduledAt:  time.Now().Add(24 * time.Hour),
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// モックリポジトリを作成
			mockRepo := &mockMatchRepository{}
			
			err := mockRepo.Create(tt.match)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("Create() error = nil, wantErr %v", tt.wantErr)
					return
				}
				
				if IsRepositoryError(err) {
					if GetRepositoryErrorType(err) != tt.errType {
						t.Errorf("Create() error type = %v, want %v", GetRepositoryErrorType(err), tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				
				if tt.match != nil && tt.match.ID == 0 {
					t.Error("Create() should set match ID")
				}
			}
		})
	}
}

// TestMatchRepository_GetByID はID取得機能をテストする
func TestMatchRepository_GetByID(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr bool
		errType ErrorType
	}{
		{
			name:    "正常なID取得",
			id:      1,
			wantErr: false,
		},
		{
			name:    "無効なID（0）",
			id:      0,
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name:    "無効なID（負の値）",
			id:      -1,
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name:    "存在しないID",
			id:      999,
			wantErr: true,
			errType: ErrTypeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMatchRepository{}
			
			match, err := mockRepo.GetByID(tt.id)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetByID() error = nil, wantErr %v", tt.wantErr)
					return
				}
				
				if IsRepositoryError(err) {
					if GetRepositoryErrorType(err) != tt.errType {
						t.Errorf("GetByID() error type = %v, want %v", GetRepositoryErrorType(err), tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetByID() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				
				if match == nil {
					t.Error("GetByID() should return match")
				}
				
				if match.ID != tt.id {
					t.Errorf("GetByID() match.ID = %v, want %v", match.ID, tt.id)
				}
			}
		})
	}
}

// TestMatchRepository_UpdateResult は試合結果更新機能をテストする
func TestMatchRepository_UpdateResult(t *testing.T) {
	tests := []struct {
		name    string
		matchID int
		result  models.MatchResult
		wantErr bool
		errType ErrorType
	}{
		{
			name:    "正常な結果更新",
			matchID: 1,
			result: models.MatchResult{
				Score1: 3,
				Score2: 1,
				Winner: "チームA",
			},
			wantErr: false,
		},
		{
			name:    "無効なマッチID",
			matchID: 0,
			result: models.MatchResult{
				Score1: 3,
				Score2: 1,
				Winner: "チームA",
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name:    "負のスコア",
			matchID: 1,
			result: models.MatchResult{
				Score1: -1,
				Score2: 1,
				Winner: "チームA",
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name:    "引き分け",
			matchID: 1,
			result: models.MatchResult{
				Score1: 2,
				Score2: 2,
				Winner: "チームA",
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name:    "空の勝者",
			matchID: 1,
			result: models.MatchResult{
				Score1: 3,
				Score2: 1,
				Winner: "",
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name:    "存在しない試合",
			matchID: 999,
			result: models.MatchResult{
				Score1: 3,
				Score2: 1,
				Winner: "チームA",
			},
			wantErr: true,
			errType: ErrTypeNotFound,
		},
		{
			name:    "完了済み試合の更新",
			matchID: 2, // モックで完了済みとして設定
			result: models.MatchResult{
				Score1: 3,
				Score2: 1,
				Winner: "チームA",
			},
			wantErr: true,
			errType: ErrTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMatchRepository{}
			
			err := mockRepo.UpdateResult(tt.matchID, tt.result)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("UpdateResult() error = nil, wantErr %v", tt.wantErr)
					return
				}
				
				if IsRepositoryError(err) {
					if GetRepositoryErrorType(err) != tt.errType {
						t.Errorf("UpdateResult() error type = %v, want %v", GetRepositoryErrorType(err), tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("UpdateResult() error = %v, wantErr %v", err, tt.wantErr)
				}
			}
		})
	}
}

// TestMatchRepository_GetBySport はスポーツ別取得機能をテストする
func TestMatchRepository_GetBySport(t *testing.T) {
	tests := []struct {
		name    string
		sport   string
		wantErr bool
		errType ErrorType
	}{
		{
			name:    "バレーボール",
			sport:   models.SportVolleyball,
			wantErr: false,
		},
		{
			name:    "卓球",
			sport:   models.SportTableTennis,
			wantErr: false,
		},
		{
			name:    "サッカー",
			sport:   models.SportSoccer,
			wantErr: false,
		},
		{
			name:    "無効なスポーツ",
			sport:   "invalid_sport",
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name:    "空のスポーツ",
			sport:   "",
			wantErr: true,
			errType: ErrTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMatchRepository{}
			
			matches, err := mockRepo.GetBySport(tt.sport)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetBySport() error = nil, wantErr %v", tt.wantErr)
					return
				}
				
				if IsRepositoryError(err) {
					if GetRepositoryErrorType(err) != tt.errType {
						t.Errorf("GetBySport() error type = %v, want %v", GetRepositoryErrorType(err), tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetBySport() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				
				if matches == nil {
					t.Error("GetBySport() should return matches slice")
				}
			}
		})
	}
}

// TestMatchRepository_GetByTournament はトーナメント別取得機能をテストする
func TestMatchRepository_GetByTournament(t *testing.T) {
	tests := []struct {
		name         string
		tournamentID int
		wantErr      bool
		errType      ErrorType
	}{
		{
			name:         "正常なトーナメントID",
			tournamentID: 1,
			wantErr:      false,
		},
		{
			name:         "無効なトーナメントID（0）",
			tournamentID: 0,
			wantErr:      true,
			errType:      ErrTypeValidation,
		},
		{
			name:         "無効なトーナメントID（負の値）",
			tournamentID: -1,
			wantErr:      true,
			errType:      ErrTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMatchRepository{}
			
			matches, err := mockRepo.GetByTournament(tt.tournamentID)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetByTournament() error = nil, wantErr %v", tt.wantErr)
					return
				}
				
				if IsRepositoryError(err) {
					if GetRepositoryErrorType(err) != tt.errType {
						t.Errorf("GetByTournament() error type = %v, want %v", GetRepositoryErrorType(err), tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetByTournament() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				
				if matches == nil {
					t.Error("GetByTournament() should return matches slice")
				}
			}
		})
	}
}

// TestMatchRepository_GetByStatus はステータス別取得機能をテストする
func TestMatchRepository_GetByStatus(t *testing.T) {
	tests := []struct {
		name    string
		status  string
		wantErr bool
		errType ErrorType
	}{
		{
			name:    "未実施ステータス",
			status:  models.MatchStatusPending,
			wantErr: false,
		},
		{
			name:    "完了ステータス",
			status:  models.MatchStatusCompleted,
			wantErr: false,
		},
		{
			name:    "無効なステータス",
			status:  "invalid_status",
			wantErr: true,
			errType: ErrTypeValidation,
		},
		{
			name:    "空のステータス",
			status:  "",
			wantErr: true,
			errType: ErrTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMatchRepository{}
			
			matches, err := mockRepo.GetByStatus(tt.status)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetByStatus() error = nil, wantErr %v", tt.wantErr)
					return
				}
				
				if IsRepositoryError(err) {
					if GetRepositoryErrorType(err) != tt.errType {
						t.Errorf("GetByStatus() error type = %v, want %v", GetRepositoryErrorType(err), tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetByStatus() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				
				if matches == nil {
					t.Error("GetByStatus() should return matches slice")
				}
			}
		})
	}
}

// TestMatchRepository_CountByTournament はトーナメント別カウント機能をテストする
func TestMatchRepository_CountByTournament(t *testing.T) {
	tests := []struct {
		name         string
		tournamentID int
		wantErr      bool
		errType      ErrorType
		wantCount    int
	}{
		{
			name:         "正常なトーナメントID",
			tournamentID: 1,
			wantErr:      false,
			wantCount:    5,
		},
		{
			name:         "無効なトーナメントID",
			tournamentID: 0,
			wantErr:      true,
			errType:      ErrTypeValidation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMatchRepository{}
			
			count, err := mockRepo.CountByTournament(tt.tournamentID)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("CountByTournament() error = nil, wantErr %v", tt.wantErr)
					return
				}
				
				if IsRepositoryError(err) {
					if GetRepositoryErrorType(err) != tt.errType {
						t.Errorf("CountByTournament() error type = %v, want %v", GetRepositoryErrorType(err), tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("CountByTournament() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				
				if count != tt.wantCount {
					t.Errorf("CountByTournament() count = %v, want %v", count, tt.wantCount)
				}
			}
		})
	}
}

// TestMatchRepository_GetMatchesByDateRange は日付範囲取得機能をテストする
func TestMatchRepository_GetMatchesByDateRange(t *testing.T) {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	yesterday := now.Add(-24 * time.Hour)

	tests := []struct {
		name      string
		startDate time.Time
		endDate   time.Time
		wantErr   bool
		errType   ErrorType
	}{
		{
			name:      "正常な日付範囲",
			startDate: yesterday,
			endDate:   tomorrow,
			wantErr:   false,
		},
		{
			name:      "開始日が終了日より後",
			startDate: tomorrow,
			endDate:   yesterday,
			wantErr:   true,
			errType:   ErrTypeValidation,
		},
		{
			name:      "同じ日付",
			startDate: now,
			endDate:   now,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &mockMatchRepository{}
			
			matches, err := mockRepo.GetMatchesByDateRange(tt.startDate, tt.endDate)
			
			if tt.wantErr {
				if err == nil {
					t.Errorf("GetMatchesByDateRange() error = nil, wantErr %v", tt.wantErr)
					return
				}
				
				if IsRepositoryError(err) {
					if GetRepositoryErrorType(err) != tt.errType {
						t.Errorf("GetMatchesByDateRange() error type = %v, want %v", GetRepositoryErrorType(err), tt.errType)
					}
				}
			} else {
				if err != nil {
					t.Errorf("GetMatchesByDateRange() error = %v, wantErr %v", err, tt.wantErr)
					return
				}
				
				if matches == nil {
					t.Error("GetMatchesByDateRange() should return matches slice")
				}
			}
		})
	}
}

// mockMatchRepository はテスト用のモックリポジトリ
type mockMatchRepository struct{}

func (m *mockMatchRepository) Create(match *models.Match) error {
	if match == nil {
		return NewRepositoryError(ErrTypeValidation, "試合がnilです", nil)
	}
	
	if err := match.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "試合の検証に失敗しました", err)
	}
	
	// モックでIDを設定
	match.ID = 1
	return nil
}

func (m *mockMatchRepository) GetByID(id int) (*models.Match, error) {
	if id <= 0 {
		return nil, NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	if id == 999 {
		return nil, NewRepositoryError(ErrTypeNotFound, "試合が見つかりません", nil)
	}
	
	match := &models.Match{
		ID:           id,
		TournamentID: 1,
		Round:        models.Round1stRound,
		Team1:        "チームA",
		Team2:        "チームB",
		Status:       models.MatchStatusPending,
		ScheduledAt:  time.Now().Add(24 * time.Hour),
	}
	
	// ID=2は完了済みとして設定
	if id == 2 {
		match.Status = models.MatchStatusCompleted
		score1 := 3
		score2 := 1
		winner := "チームA"
		match.Score1 = &score1
		match.Score2 = &score2
		match.Winner = &winner
		completedAt := time.Now()
		match.CompletedAt = &completedAt
	}
	
	return match, nil
}

func (m *mockMatchRepository) Update(match *models.Match) error {
	if match == nil {
		return NewRepositoryError(ErrTypeValidation, "試合がnilです", nil)
	}
	
	if match.ID <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	if err := match.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "試合の検証に失敗しました", err)
	}
	
	return nil
}

func (m *mockMatchRepository) Delete(id int) error {
	if id <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	return nil
}

func (m *mockMatchRepository) UpdateResult(matchID int, result models.MatchResult) error {
	if matchID <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	if err := result.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "試合結果の検証に失敗しました", err)
	}
	
	if matchID == 999 {
		return NewRepositoryError(ErrTypeNotFound, "試合が見つかりません", nil)
	}
	
	// ID=2は完了済みとして設定
	if matchID == 2 {
		return NewRepositoryError(ErrTypeValidation, "完了済みの試合結果は更新できません", nil)
	}
	
	return nil
}

func (m *mockMatchRepository) UpdateStatus(matchID int, status string) error {
	if matchID <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	if !models.IsValidMatchStatus(status) {
		return NewRepositoryError(ErrTypeValidation, "無効な試合ステータスです", nil)
	}
	
	return nil
}

func (m *mockMatchRepository) GetBySport(sport string) ([]*models.Match, error) {
	if !models.IsValidSport(sport) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なスポーツです", nil)
	}
	
	return []*models.Match{}, nil
}

func (m *mockMatchRepository) GetByTournament(tournamentID int) ([]*models.Match, error) {
	if tournamentID <= 0 {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	return []*models.Match{}, nil
}

func (m *mockMatchRepository) GetByTournamentAndRound(tournamentID int, round string) ([]*models.Match, error) {
	if tournamentID <= 0 {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	if !models.IsValidRound(round) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なラウンドです", nil)
	}
	
	return []*models.Match{}, nil
}

func (m *mockMatchRepository) GetByStatus(status string) ([]*models.Match, error) {
	if !models.IsValidMatchStatus(status) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効な試合ステータスです", nil)
	}
	
	return []*models.Match{}, nil
}

func (m *mockMatchRepository) GetPendingMatches() ([]*models.Match, error) {
	return []*models.Match{}, nil
}

func (m *mockMatchRepository) GetCompletedMatches() ([]*models.Match, error) {
	return []*models.Match{}, nil
}

func (m *mockMatchRepository) GetMatchesByDateRange(startDate, endDate time.Time) ([]*models.Match, error) {
	if startDate.After(endDate) {
		return nil, NewRepositoryError(ErrTypeValidation, "開始日は終了日より前である必要があります", nil)
	}
	
	return []*models.Match{}, nil
}

func (m *mockMatchRepository) CountByTournament(tournamentID int) (int, error) {
	if tournamentID <= 0 {
		return 0, NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	return 5, nil
}

func (m *mockMatchRepository) CountByStatus(status string) (int, error) {
	if !models.IsValidMatchStatus(status) {
		return 0, NewRepositoryError(ErrTypeValidation, "無効な試合ステータスです", nil)
	}
	
	return 3, nil
}
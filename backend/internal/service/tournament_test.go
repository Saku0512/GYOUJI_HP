package service

import (
	"errors"
	"testing"
	"time"

	"backend/internal/models"
)

// MockTournamentRepository はテスト用のTournamentRepositoryモック
type MockTournamentRepository struct {
	tournaments map[int]*models.Tournament
	sportIndex  map[string]*models.Tournament
	err         error
	nextID      int
}

func NewMockTournamentRepository() *MockTournamentRepository {
	return &MockTournamentRepository{
		tournaments: make(map[int]*models.Tournament),
		sportIndex:  make(map[string]*models.Tournament),
		nextID:      1,
	}
}

func (m *MockTournamentRepository) Create(tournament *models.Tournament) error {
	if m.err != nil {
		return m.err
	}
	
	tournament.ID = m.nextID
	m.nextID++
	tournament.CreatedAt = time.Now()
	tournament.UpdatedAt = time.Now()
	
	m.tournaments[tournament.ID] = tournament
	m.sportIndex[tournament.Sport] = tournament
	return nil
}

func (m *MockTournamentRepository) GetByID(id int) (*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return nil, errors.New("トーナメントが見つかりません")
	}
	
	return tournament, nil
}

func (m *MockTournamentRepository) GetBySport(sport string) (*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, exists := m.sportIndex[sport]
	if !exists {
		return nil, errors.New("指定されたスポーツのトーナメントが見つかりません")
	}
	
	return tournament, nil
}

func (m *MockTournamentRepository) GetAll() ([]*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var tournaments []*models.Tournament
	for _, tournament := range m.tournaments {
		tournaments = append(tournaments, tournament)
	}
	
	return tournaments, nil
}

func (m *MockTournamentRepository) Update(tournament *models.Tournament) error {
	if m.err != nil {
		return m.err
	}
	
	_, exists := m.tournaments[tournament.ID]
	if !exists {
		return errors.New("更新対象のトーナメントが見つかりません")
	}
	
	tournament.UpdatedAt = time.Now()
	m.tournaments[tournament.ID] = tournament
	m.sportIndex[tournament.Sport] = tournament
	return nil
}

func (m *MockTournamentRepository) Delete(id int) error {
	if m.err != nil {
		return m.err
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return errors.New("削除対象のトーナメントが見つかりません")
	}
	
	delete(m.tournaments, id)
	delete(m.sportIndex, tournament.Sport)
	return nil
}

func (m *MockTournamentRepository) GetTournamentBracket(sport string) (*models.Bracket, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, err := m.GetBySport(sport)
	if err != nil {
		return nil, err
	}
	
	// モック用の簡単なブラケットを作成
	bracket := &models.Bracket{
		TournamentID: tournament.ID,
		Sport:        tournament.Sport,
		Format:       tournament.Format,
		Rounds:       []models.Round{},
	}
	
	// 1回戦のラウンドを追加（テスト用）
	round := models.Round{
		Name:    models.Round1stRound,
		Matches: []models.Match{},
	}
	
	// テスト用の試合を追加
	for i := 0; i < 4; i++ {
		match := models.Match{
			ID:           i + 1,
			TournamentID: tournament.ID,
			Round:        models.Round1stRound,
			Team1:        "チーム1",
			Team2:        "チーム2",
			Status:       models.MatchStatusPending,
			ScheduledAt:  time.Now(),
		}
		round.Matches = append(round.Matches, match)
	}
	
	bracket.Rounds = append(bracket.Rounds, round)
	return bracket, nil
}

func (m *MockTournamentRepository) GetTournamentBracketByID(tournamentID int) (*models.Bracket, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, err := m.GetByID(tournamentID)
	if err != nil {
		return nil, err
	}
	
	return &models.Bracket{
		TournamentID: tournament.ID,
		Sport:        tournament.Sport,
		Format:       tournament.Format,
		Rounds:       []models.Round{},
	}, nil
}

func (m *MockTournamentRepository) UpdateStatus(id int, status string) error {
	if m.err != nil {
		return m.err
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return errors.New("更新対象のトーナメントが見つかりません")
	}
	
	tournament.Status = status
	tournament.UpdatedAt = time.Now()
	return nil
}

func (m *MockTournamentRepository) UpdateFormat(id int, format string) error {
	if m.err != nil {
		return m.err
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return errors.New("更新対象のトーナメントが見つかりません")
	}
	
	tournament.Format = format
	tournament.UpdatedAt = time.Now()
	return nil
}

func (m *MockTournamentRepository) GetByStatus(status string) ([]*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var tournaments []*models.Tournament
	for _, tournament := range m.tournaments {
		if tournament.Status == status {
			tournaments = append(tournaments, tournament)
		}
	}
	
	return tournaments, nil
}

func (m *MockTournamentRepository) GetActiveByFormat(format string) ([]*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var tournaments []*models.Tournament
	for _, tournament := range m.tournaments {
		if tournament.Format == format && tournament.Status == models.TournamentStatusActive {
			tournaments = append(tournaments, tournament)
		}
	}
	
	return tournaments, nil
}

func (m *MockTournamentRepository) SetError(err error) {
	m.err = err
}

func (m *MockTournamentRepository) AddTournament(sport, format, status string) *models.Tournament {
	tournament := &models.Tournament{
		ID:        m.nextID,
		Sport:     sport,
		Format:    format,
		Status:    status,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	m.nextID++
	m.tournaments[tournament.ID] = tournament
	m.sportIndex[tournament.Sport] = tournament
	
	return tournament
}

// MockMatchRepository はテスト用のMatchRepositoryモック
type MockMatchRepository struct {
	matches    map[int]*models.Match
	err        error
	nextID     int
	matchCount map[int]int // tournamentID -> match count
}

func NewMockMatchRepository() *MockMatchRepository {
	return &MockMatchRepository{
		matches:    make(map[int]*models.Match),
		matchCount: make(map[int]int),
		nextID:     1,
	}
}

func (m *MockMatchRepository) Create(match *models.Match) error {
	if m.err != nil {
		return m.err
	}
	
	match.ID = m.nextID
	m.nextID++
	
	m.matches[match.ID] = match
	m.matchCount[match.TournamentID]++
	return nil
}

func (m *MockMatchRepository) GetByID(id int) (*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	match, exists := m.matches[id]
	if !exists {
		return nil, errors.New("試合が見つかりません")
	}
	
	return match, nil
}

func (m *MockMatchRepository) Update(match *models.Match) error {
	if m.err != nil {
		return m.err
	}
	
	_, exists := m.matches[match.ID]
	if !exists {
		return errors.New("更新対象の試合が見つかりません")
	}
	
	m.matches[match.ID] = match
	return nil
}

func (m *MockMatchRepository) Delete(id int) error {
	if m.err != nil {
		return m.err
	}
	
	match, exists := m.matches[id]
	if !exists {
		return errors.New("削除対象の試合が見つかりません")
	}
	
	delete(m.matches, id)
	m.matchCount[match.TournamentID]--
	return nil
}

func (m *MockMatchRepository) UpdateResult(matchID int, result models.MatchResult) error {
	if m.err != nil {
		return m.err
	}
	
	match, exists := m.matches[matchID]
	if !exists {
		return errors.New("試合が見つかりません")
	}
	
	match.Score1 = &result.Score1
	match.Score2 = &result.Score2
	match.Winner = &result.Winner
	match.Status = models.MatchStatusCompleted
	completedAt := time.Now()
	match.CompletedAt = &completedAt
	
	return nil
}

func (m *MockMatchRepository) UpdateStatus(matchID int, status string) error {
	if m.err != nil {
		return m.err
	}
	
	match, exists := m.matches[matchID]
	if !exists {
		return errors.New("試合が見つかりません")
	}
	
	match.Status = status
	return nil
}

func (m *MockMatchRepository) GetBySport(sport string) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var matches []*models.Match
	for _, match := range m.matches {
		matches = append(matches, match)
	}
	
	return matches, nil
}

func (m *MockMatchRepository) GetByTournament(tournamentID int) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var matches []*models.Match
	for _, match := range m.matches {
		if match.TournamentID == tournamentID {
			matches = append(matches, match)
		}
	}
	
	return matches, nil
}

func (m *MockMatchRepository) GetByTournamentAndRound(tournamentID int, round string) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var matches []*models.Match
	for _, match := range m.matches {
		if match.TournamentID == tournamentID && match.Round == round {
			matches = append(matches, match)
		}
	}
	
	return matches, nil
}

func (m *MockMatchRepository) GetByStatus(status string) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var matches []*models.Match
	for _, match := range m.matches {
		if match.Status == status {
			matches = append(matches, match)
		}
	}
	
	return matches, nil
}

func (m *MockMatchRepository) GetPendingMatches() ([]*models.Match, error) {
	return m.GetByStatus(models.MatchStatusPending)
}

func (m *MockMatchRepository) GetCompletedMatches() ([]*models.Match, error) {
	return m.GetByStatus(models.MatchStatusCompleted)
}

func (m *MockMatchRepository) GetMatchesByDateRange(startDate, endDate time.Time) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	var matches []*models.Match
	for _, match := range m.matches {
		if match.ScheduledAt.After(startDate) && match.ScheduledAt.Before(endDate) {
			matches = append(matches, match)
		}
	}
	
	return matches, nil
}

func (m *MockMatchRepository) CountByTournament(tournamentID int) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	
	return m.matchCount[tournamentID], nil
}

func (m *MockMatchRepository) CountByStatus(status string) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	
	count := 0
	for _, match := range m.matches {
		if match.Status == status {
			count++
		}
	}
	
	return count, nil
}

func (m *MockMatchRepository) SetError(err error) {
	m.err = err
}

func TestNewTournamentService(t *testing.T) {
	mockTournamentRepo := NewMockTournamentRepository()
	mockMatchRepo := NewMockMatchRepository()
	
	service := NewTournamentService(mockTournamentRepo, mockMatchRepo)
	
	if service == nil {
		t.Error("TournamentServiceの作成に失敗しました")
	}
}

func TestTournamentService_CreateTournament(t *testing.T) {
	tests := []struct {
		name          string
		sport         string
		format        string
		existingTournament bool
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "正常なトーナメント作成",
			sport:         models.SportVolleyball,
			format:        models.FormatStandard,
			expectedError: false,
		},
		{
			name:          "無効なスポーツ",
			sport:         "invalid_sport",
			format:        models.FormatStandard,
			expectedError: true,
			errorMessage:  "無効なスポーツです",
		},
		{
			name:          "無効なフォーマット",
			sport:         models.SportVolleyball,
			format:        "invalid_format",
			expectedError: true,
			errorMessage:  "無効なトーナメントフォーマットです",
		},
		{
			name:               "既存のアクティブトーナメント",
			sport:              models.SportVolleyball,
			format:             models.FormatStandard,
			existingTournament: true,
			expectedError:      true,
			errorMessage:       "スポーツ volleyball には既にアクティブなトーナメントが存在します",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTournamentRepo := NewMockTournamentRepository()
			mockMatchRepo := NewMockMatchRepository()
			service := NewTournamentService(mockTournamentRepo, mockMatchRepo)
			
			// 既存のトーナメントを設定
			if tt.existingTournament {
				mockTournamentRepo.AddTournament(tt.sport, tt.format, models.TournamentStatusActive)
			}
			
			tournament, err := service.CreateTournament(tt.sport, tt.format)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
				if tournament != nil {
					t.Error("エラー時にトーナメントが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if tournament == nil {
					t.Error("トーナメントが返されませんでした")
				} else {
					if tournament.Sport != tt.sport {
						t.Errorf("期待されたスポーツ: %s, 実際: %s", tt.sport, tournament.Sport)
					}
					if tournament.Format != tt.format {
						t.Errorf("期待されたフォーマット: %s, 実際: %s", tt.format, tournament.Format)
					}
					if tournament.Status != models.TournamentStatusActive {
						t.Errorf("期待されたステータス: %s, 実際: %s", models.TournamentStatusActive, tournament.Status)
					}
				}
			}
		})
	}
}

func TestTournamentService_GetTournament(t *testing.T) {
	tests := []struct {
		name          string
		sport         string
		setupTournament bool
		expectedError bool
		errorMessage  string
	}{
		{
			name:            "正常なトーナメント取得",
			sport:           models.SportVolleyball,
			setupTournament: true,
			expectedError:   false,
		},
		{
			name:          "無効なスポーツ",
			sport:         "invalid_sport",
			expectedError: true,
			errorMessage:  "無効なスポーツです",
		},
		{
			name:          "存在しないトーナメント",
			sport:         models.SportVolleyball,
			expectedError: true,
			errorMessage:  "スポーツ volleyball のトーナメントが見つかりません",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTournamentRepo := NewMockTournamentRepository()
			mockMatchRepo := NewMockMatchRepository()
			service := NewTournamentService(mockTournamentRepo, mockMatchRepo)
			
			// テストトーナメントを設定
			if tt.setupTournament {
				mockTournamentRepo.AddTournament(tt.sport, models.FormatStandard, models.TournamentStatusActive)
			}
			
			tournament, err := service.GetTournament(tt.sport)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
				if tournament != nil {
					t.Error("エラー時にトーナメントが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if tournament == nil {
					t.Error("トーナメントが返されませんでした")
				} else {
					if tournament.Sport != tt.sport {
						t.Errorf("期待されたスポーツ: %s, 実際: %s", tt.sport, tournament.Sport)
					}
				}
			}
		})
	}
}

func TestTournamentService_GenerateBracket(t *testing.T) {
	tests := []struct {
		name          string
		sport         string
		format        string
		teams         []string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "バレーボール正常ブラケット生成",
			sport:         models.SportVolleyball,
			format:        models.FormatStandard,
			teams:         []string{"チーム1", "チーム2", "チーム3", "チーム4", "チーム5", "チーム6", "チーム7", "チーム8"},
			expectedError: false,
		},
		{
			name:          "卓球標準フォーマット",
			sport:         models.SportTableTennis,
			format:        models.FormatStandard,
			teams:         []string{"チーム1", "チーム2", "チーム3", "チーム4", "チーム5", "チーム6", "チーム7", "チーム8"},
			expectedError: false,
		},
		{
			name:          "卓球雨天フォーマット",
			sport:         models.SportTableTennis,
			format:        models.FormatRainy,
			teams:         []string{"チーム1", "チーム2", "チーム3", "チーム4", "チーム5", "チーム6", "チーム7", "チーム8"},
			expectedError: false,
		},
		{
			name:          "サッカー正常ブラケット生成",
			sport:         models.SportSoccer,
			format:        models.FormatStandard,
			teams:         []string{"チーム1", "チーム2", "チーム3", "チーム4", "チーム5", "チーム6", "チーム7", "チーム8"},
			expectedError: false,
		},
		{
			name:          "無効なスポーツ",
			sport:         "invalid_sport",
			format:        models.FormatStandard,
			teams:         []string{"チーム1", "チーム2"},
			expectedError: true,
			errorMessage:  "無効なスポーツです",
		},
		{
			name:          "チーム不足",
			sport:         models.SportVolleyball,
			format:        models.FormatStandard,
			teams:         []string{"チーム1", "チーム2"},
			expectedError: true,
			errorMessage:  "スポーツ volleyball には最低 8 チームが必要です",
		},
		{
			name:          "空のチーム",
			sport:         models.SportVolleyball,
			format:        models.FormatStandard,
			teams:         []string{},
			expectedError: true,
			errorMessage:  "チームが指定されていません",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTournamentRepo := NewMockTournamentRepository()
			mockMatchRepo := NewMockMatchRepository()
			service := NewTournamentService(mockTournamentRepo, mockMatchRepo)
			
			bracket, err := service.GenerateBracket(tt.sport, tt.format, tt.teams)
			
			if tt.expectedError {
				if err == nil {
					t.Error("エラーが期待されましたが、エラーが発生しませんでした")
				} else if tt.errorMessage != "" && err.Error() != tt.errorMessage {
					t.Errorf("期待されたエラーメッセージ: %s, 実際: %s", tt.errorMessage, err.Error())
				}
				if bracket != nil {
					t.Error("エラー時にブラケットが返されました")
				}
			} else {
				if err != nil {
					t.Errorf("予期しないエラー: %v", err)
				}
				if bracket == nil {
					t.Error("ブラケットが返されませんでした")
				} else {
					if bracket.Sport != tt.sport {
						t.Errorf("期待されたスポーツ: %s, 実際: %s", tt.sport, bracket.Sport)
					}
					if bracket.Format != tt.format {
						t.Errorf("期待されたフォーマット: %s, 実際: %s", tt.format, bracket.Format)
					}
					if len(bracket.Rounds) == 0 {
						t.Error("ラウンドが生成されませんでした")
					}
					
					// 1回戦の試合数をチェック
					if len(bracket.Rounds) > 0 {
						firstRound := bracket.Rounds[0]
						expectedMatches := len(tt.teams) / 2
						if len(firstRound.Matches) != expectedMatches {
							t.Errorf("期待された1回戦試合数: %d, 実際: %d", expectedMatches, len(firstRound.Matches))
						}
					}
				}
			}
		})
	}
}

func TestTournamentService_SwitchTournamentFormat(t *testing.T) {
	tests := []struct {
		name          string
		sport         string
		currentFormat string
		newFormat     string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "卓球フォーマット切り替え成功",
			sport:         models.SportTableTennis,
			currentFormat: models.FormatStandard,
			newFormat:     models.FormatRainy,
			expectedError: false,
		},
		{
			name:          "無効なスポーツ",
			sport:         "invalid_sport",
			currentFormat: models.FormatStandard,
			newFormat:     models.FormatRainy,
			expectedError: true,
			errorMessage:  "無効なスポーツです",
		},
		{
			name:          "サポートされていないスポーツ",
			sport:         models.SportVolleyball,
			currentFormat: models.FormatStandard,
			newFormat:     models.FormatRainy,
			expectedError: true,
			errorMessage:  "スポーツ volleyball はフォーマット切り替えをサポートしていません",
		},
		{
			name:          "同じフォーマット",
			sport:         models.SportTableTennis,
			currentFormat: models.FormatStandard,
			newFormat:     models.FormatStandard,
			expectedError: true,
			errorMessage:  "トーナメントは既に standard フォーマットです",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockTournamentRepo := NewMockTournamentRepository()
			mockMatchRepo := NewMockMatchRepository()
			service := NewTournamentService(mockTournamentRepo, mockMatchRepo)
			
			// テストトーナメントを設定
			if models.IsValidSport(tt.sport) {
				mockTournamentRepo.AddTournament(tt.sport, tt.currentFormat, models.TournamentStatusActive)
			}
			
			err := service.SwitchTournamentFormat(tt.sport, tt.newFormat)
			
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
				
				// フォーマットが更新されたかチェック
				tournament, err := service.GetTournament(tt.sport)
				if err != nil {
					t.Errorf("更新後のトーナメント取得エラー: %v", err)
				} else if tournament.Format != tt.newFormat {
					t.Errorf("期待されたフォーマット: %s, 実際: %s", tt.newFormat, tournament.Format)
				}
			}
		})
	}
}

func TestTournamentService_GetTournamentProgress(t *testing.T) {
	mockTournamentRepo := NewMockTournamentRepository()
	mockMatchRepo := NewMockMatchRepository()
	service := NewTournamentService(mockTournamentRepo, mockMatchRepo)
	
	// テストトーナメントを設定
	tournament := mockTournamentRepo.AddTournament(models.SportVolleyball, models.FormatStandard, models.TournamentStatusActive)
	
	// テスト試合を追加
	for i := 0; i < 4; i++ {
		match := &models.Match{
			TournamentID: tournament.ID,
			Round:        models.Round1stRound,
			Team1:        "チーム1",
			Team2:        "チーム2",
			Status:       models.MatchStatusPending,
			ScheduledAt:  time.Now(),
		}
		mockMatchRepo.Create(match)
	}
	
	progress, err := service.GetTournamentProgress(models.SportVolleyball)
	
	if err != nil {
		t.Errorf("予期しないエラー: %v", err)
	}
	
	if progress == nil {
		t.Error("進行状況が返されませんでした")
	} else {
		if progress.Sport != models.SportVolleyball {
			t.Errorf("期待されたスポーツ: %s, 実際: %s", models.SportVolleyball, progress.Sport)
		}
		if progress.TotalMatches != 4 {
			t.Errorf("期待された総試合数: 4, 実際: %d", progress.TotalMatches)
		}
		if progress.CompletedMatches != 0 {
			t.Errorf("期待された完了試合数: 0, 実際: %d", progress.CompletedMatches)
		}
		if progress.PendingMatches != 4 {
			t.Errorf("期待された未実施試合数: 4, 実際: %d", progress.PendingMatches)
		}
	}
}
package repository

import (
	"testing"
	"time"

	"backend/internal/models"
)

// MockTournamentRepository はTournamentRepositoryのモック実装
type MockTournamentRepository struct {
	tournaments map[int]*models.Tournament
	nextID      int
	matches     map[int][]models.Match // tournamentID -> matches
}

// NewMockTournamentRepository は新しいモックリポジトリを作成する
func NewMockTournamentRepository() *MockTournamentRepository {
	return &MockTournamentRepository{
		tournaments: make(map[int]*models.Tournament),
		nextID:      1,
		matches:     make(map[int][]models.Match),
	}
}

// Create はトーナメントを作成する（モック実装）
func (m *MockTournamentRepository) Create(tournament *models.Tournament) error {
	if tournament == nil {
		return NewRepositoryError(ErrTypeValidation, "トーナメントがnilです", nil)
	}
	
	if err := tournament.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "トーナメントの検証に失敗しました", err)
	}
	
	tournament.ID = m.nextID
	m.nextID++
	
	now := time.Now()
	tournament.CreatedAt = now
	tournament.UpdatedAt = now
	
	// ディープコピーを作成
	tournamentCopy := *tournament
	m.tournaments[tournament.ID] = &tournamentCopy
	
	return nil
}

// GetByID はIDでトーナメントを取得する（モック実装）
func (m *MockTournamentRepository) GetByID(id int) (*models.Tournament, error) {
	if id <= 0 {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return nil, NewRepositoryError(ErrTypeNotFound, "トーナメントが見つかりません", nil)
	}
	
	// ディープコピーを返す
	tournamentCopy := *tournament
	return &tournamentCopy, nil
}

// GetBySport はスポーツでトーナメントを取得する（モック実装）
func (m *MockTournamentRepository) GetBySport(sport string) (*models.Tournament, error) {
	if !models.IsValidSport(sport) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なスポーツです", nil)
	}
	
	var latestTournament *models.Tournament
	var latestTime time.Time
	
	for _, tournament := range m.tournaments {
		if tournament.Sport == sport {
			if latestTournament == nil || tournament.CreatedAt.After(latestTime) {
				latestTournament = tournament
				latestTime = tournament.CreatedAt
			}
		}
	}
	
	if latestTournament == nil {
		return nil, NewRepositoryError(ErrTypeNotFound, "指定されたスポーツのトーナメントが見つかりません", nil)
	}
	
	// ディープコピーを返す
	tournamentCopy := *latestTournament
	return &tournamentCopy, nil
}

// GetAll は全てのトーナメントを取得する（モック実装）
func (m *MockTournamentRepository) GetAll() ([]*models.Tournament, error) {
	var tournaments []*models.Tournament
	
	for _, tournament := range m.tournaments {
		// ディープコピーを作成
		tournamentCopy := *tournament
		tournaments = append(tournaments, &tournamentCopy)
	}
	
	return tournaments, nil
}

// Update はトーナメントを更新する（モック実装）
func (m *MockTournamentRepository) Update(tournament *models.Tournament) error {
	if tournament == nil {
		return NewRepositoryError(ErrTypeValidation, "トーナメントがnilです", nil)
	}
	
	if tournament.ID <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	if err := tournament.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "トーナメントの検証に失敗しました", err)
	}
	
	_, exists := m.tournaments[tournament.ID]
	if !exists {
		return NewRepositoryError(ErrTypeNotFound, "更新対象のトーナメントが見つかりません", nil)
	}
	
	tournament.UpdatedAt = time.Now()
	
	// ディープコピーを保存
	tournamentCopy := *tournament
	m.tournaments[tournament.ID] = &tournamentCopy
	
	return nil
}

// Delete はトーナメントを削除する（モック実装）
func (m *MockTournamentRepository) Delete(id int) error {
	if id <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	_, exists := m.tournaments[id]
	if !exists {
		return NewRepositoryError(ErrTypeNotFound, "削除対象のトーナメントが見つかりません", nil)
	}
	
	delete(m.tournaments, id)
	delete(m.matches, id)
	
	return nil
}

// GetTournamentBracket はスポーツに基づいてトーナメントブラケットを取得する（モック実装）
func (m *MockTournamentRepository) GetTournamentBracket(sport string) (*models.Bracket, error) {
	tournament, err := m.GetBySport(sport)
	if err != nil {
		return nil, err
	}
	
	return m.GetTournamentBracketByID(tournament.ID)
}

// GetTournamentBracketByID はトーナメントIDに基づいてブラケットを取得する（モック実装）
func (m *MockTournamentRepository) GetTournamentBracketByID(tournamentID int) (*models.Bracket, error) {
	tournament, err := m.GetByID(tournamentID)
	if err != nil {
		return nil, err
	}
	
	matches, exists := m.matches[tournamentID]
	if !exists {
		matches = []models.Match{}
	}
	
	// ラウンド別に試合を整理
	roundMatches := make(map[string][]models.Match)
	for _, match := range matches {
		roundMatches[match.Round] = append(roundMatches[match.Round], match)
	}
	
	// ブラケット構造を構築
	bracket := &models.Bracket{
		TournamentID: tournament.ID,
		Sport:        tournament.Sport,
		Format:       tournament.Format,
		Rounds:       []models.Round{},
	}
	
	// スポーツに応じた有効なラウンドを取得
	validRounds := models.GetValidRoundsForSport(tournament.Sport)
	
	// 各ラウンドのデータを構築
	for _, roundName := range validRounds {
		matches, exists := roundMatches[roundName]
		if !exists {
			matches = []models.Match{}
		}
		
		round := models.Round{
			Name:    roundName,
			Matches: matches,
		}
		
		bracket.Rounds = append(bracket.Rounds, round)
	}
	
	return bracket, nil
}

// UpdateStatus はトーナメントのステータスを更新する（モック実装）
func (m *MockTournamentRepository) UpdateStatus(id int, status string) error {
	if id <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	if !models.IsValidTournamentStatus(status) {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントステータスです", nil)
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return NewRepositoryError(ErrTypeNotFound, "更新対象のトーナメントが見つかりません", nil)
	}
	
	tournament.Status = status
	tournament.UpdatedAt = time.Now()
	
	return nil
}

// UpdateFormat はトーナメントのフォーマットを更新する（モック実装）
func (m *MockTournamentRepository) UpdateFormat(id int, format string) error {
	if id <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	if !models.IsValidTournamentFormat(format) {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントフォーマットです", nil)
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return NewRepositoryError(ErrTypeNotFound, "更新対象のトーナメントが見つかりません", nil)
	}
	
	tournament.Format = format
	tournament.UpdatedAt = time.Now()
	
	return nil
}

// GetByStatus はステータスでトーナメントを取得する（モック実装）
func (m *MockTournamentRepository) GetByStatus(status string) ([]*models.Tournament, error) {
	if !models.IsValidTournamentStatus(status) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントステータスです", nil)
	}
	
	var tournaments []*models.Tournament
	
	for _, tournament := range m.tournaments {
		if tournament.Status == status {
			// ディープコピーを作成
			tournamentCopy := *tournament
			tournaments = append(tournaments, &tournamentCopy)
		}
	}
	
	return tournaments, nil
}

// GetActiveByFormat はフォーマットでアクティブなトーナメントを取得する（モック実装）
func (m *MockTournamentRepository) GetActiveByFormat(format string) ([]*models.Tournament, error) {
	if !models.IsValidTournamentFormat(format) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントフォーマットです", nil)
	}
	
	var tournaments []*models.Tournament
	
	for _, tournament := range m.tournaments {
		if tournament.Format == format && tournament.Status == models.TournamentStatusActive {
			// ディープコピーを作成
			tournamentCopy := *tournament
			tournaments = append(tournaments, &tournamentCopy)
		}
	}
	
	return tournaments, nil
}

// AddMatchesToTournament はテスト用にトーナメントに試合を追加する
func (m *MockTournamentRepository) AddMatchesToTournament(tournamentID int, matches []models.Match) {
	m.matches[tournamentID] = matches
}

// TestTournamentRepository_Create はトーナメント作成のテスト
func TestTournamentRepository_Create(t *testing.T) {
	repo := NewMockTournamentRepository()
	
	tests := []struct {
		name        string
		tournament  *models.Tournament
		expectError bool
		errorType   ErrorType
	}{
		{
			name: "正常なトーナメント作成",
			tournament: &models.Tournament{
				Sport:  models.SportVolleyball,
				Format: models.FormatStandard,
				Status: models.TournamentStatusActive,
			},
			expectError: false,
		},
		{
			name:        "nilトーナメント",
			tournament:  nil,
			expectError: true,
			errorType:   ErrTypeValidation,
		},
		{
			name: "無効なスポーツ",
			tournament: &models.Tournament{
				Sport:  "invalid_sport",
				Format: models.FormatStandard,
				Status: models.TournamentStatusActive,
			},
			expectError: true,
			errorType:   ErrTypeValidation,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(tt.tournament)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、エラーが発生しませんでした")
					return
				}
				
				if repoErr, ok := err.(*RepositoryError); ok {
					if repoErr.Type != tt.errorType {
						t.Errorf("期待されたエラータイプ %s, 実際のエラータイプ %s", tt.errorType, repoErr.Type)
					}
				} else {
					t.Errorf("RepositoryErrorが期待されましたが、異なるエラータイプが返されました: %T", err)
				}
			} else {
				if err != nil {
					t.Errorf("エラーが発生しました: %v", err)
					return
				}
				
				if tt.tournament.ID <= 0 {
					t.Errorf("トーナメントIDが設定されていません")
				}
			}
		})
	}
}

// TestTournamentRepository_GetByID はID取得のテスト
func TestTournamentRepository_GetByID(t *testing.T) {
	repo := NewMockTournamentRepository()
	
	// テストデータを作成
	tournament := &models.Tournament{
		Sport:  models.SportVolleyball,
		Format: models.FormatStandard,
		Status: models.TournamentStatusActive,
	}
	
	err := repo.Create(tournament)
	if err != nil {
		t.Fatalf("トーナメント作成に失敗しました: %v", err)
	}
	
	tests := []struct {
		name        string
		id          int
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "正常なID取得",
			id:          tournament.ID,
			expectError: false,
		},
		{
			name:        "無効なID（0）",
			id:          0,
			expectError: true,
			errorType:   ErrTypeValidation,
		},
		{
			name:        "存在しないID",
			id:          999,
			expectError: true,
			errorType:   ErrTypeNotFound,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(tt.id)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、エラーが発生しませんでした")
					return
				}
				
				if repoErr, ok := err.(*RepositoryError); ok {
					if repoErr.Type != tt.errorType {
						t.Errorf("期待されたエラータイプ %s, 実際のエラータイプ %s", tt.errorType, repoErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("エラーが発生しました: %v", err)
					return
				}
				
				if result == nil {
					t.Errorf("結果がnilです")
					return
				}
				
				if result.ID != tt.id {
					t.Errorf("期待されたID %d, 実際のID %d", tt.id, result.ID)
				}
			}
		})
	}
}

// TestTournamentRepository_GetBySport はスポーツ別取得のテスト
func TestTournamentRepository_GetBySport(t *testing.T) {
	repo := NewMockTournamentRepository()
	
	// テストデータを作成
	volleyball := &models.Tournament{
		Sport:  models.SportVolleyball,
		Format: models.FormatStandard,
		Status: models.TournamentStatusActive,
	}
	
	err := repo.Create(volleyball)
	if err != nil {
		t.Fatalf("バレーボールトーナメント作成に失敗しました: %v", err)
	}
	
	tests := []struct {
		name        string
		sport       string
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "バレーボール取得",
			sport:       models.SportVolleyball,
			expectError: false,
		},
		{
			name:        "無効なスポーツ",
			sport:       "invalid_sport",
			expectError: true,
			errorType:   ErrTypeValidation,
		},
		{
			name:        "存在しないスポーツ",
			sport:       models.SportSoccer,
			expectError: true,
			errorType:   ErrTypeNotFound,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetBySport(tt.sport)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、エラーが発生しませんでした")
					return
				}
				
				if repoErr, ok := err.(*RepositoryError); ok {
					if repoErr.Type != tt.errorType {
						t.Errorf("期待されたエラータイプ %s, 実際のエラータイプ %s", tt.errorType, repoErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("エラーが発生しました: %v", err)
					return
				}
				
				if result == nil {
					t.Errorf("結果がnilです")
					return
				}
				
				if result.Sport != tt.sport {
					t.Errorf("期待されたスポーツ %s, 実際のスポーツ %s", tt.sport, result.Sport)
				}
			}
		})
	}
}

// TestTournamentRepository_GetTournamentBracket はブラケット取得のテスト
func TestTournamentRepository_GetTournamentBracket(t *testing.T) {
	repo := NewMockTournamentRepository()
	
	// テストデータを作成
	tournament := &models.Tournament{
		Sport:  models.SportVolleyball,
		Format: models.FormatStandard,
		Status: models.TournamentStatusActive,
	}
	
	err := repo.Create(tournament)
	if err != nil {
		t.Fatalf("トーナメント作成に失敗しました: %v", err)
	}
	
	// テスト用の試合データを追加
	matches := []models.Match{
		{
			ID:           1,
			TournamentID: tournament.ID,
			Round:        models.Round1stRound,
			Team1:        "チームA",
			Team2:        "チームB",
			Status:       models.MatchStatusPending,
			ScheduledAt:  time.Now(),
		},
	}
	
	repo.AddMatchesToTournament(tournament.ID, matches)
	
	tests := []struct {
		name        string
		sport       string
		expectError bool
		errorType   ErrorType
	}{
		{
			name:        "正常なブラケット取得",
			sport:       models.SportVolleyball,
			expectError: false,
		},
		{
			name:        "無効なスポーツ",
			sport:       "invalid_sport",
			expectError: true,
			errorType:   ErrTypeValidation,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetTournamentBracket(tt.sport)
			
			if tt.expectError {
				if err == nil {
					t.Errorf("エラーが期待されましたが、エラーが発生しませんでした")
					return
				}
				
				if repoErr, ok := err.(*RepositoryError); ok {
					if repoErr.Type != tt.errorType {
						t.Errorf("期待されたエラータイプ %s, 実際のエラータイプ %s", tt.errorType, repoErr.Type)
					}
				}
			} else {
				if err != nil {
					t.Errorf("エラーが発生しました: %v", err)
					return
				}
				
				if result == nil {
					t.Errorf("結果がnilです")
					return
				}
				
				if result.Sport != tt.sport {
					t.Errorf("期待されたスポーツ %s, 実際のスポーツ %s", tt.sport, result.Sport)
				}
				
				if len(result.Rounds) == 0 {
					t.Errorf("ラウンドが設定されていません")
				}
			}
		})
	}
}
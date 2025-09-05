package service

import (
	"errors"
	"time"

	"backend/internal/models"
)

// SharedMockTournamentRepository はテスト用の共有TournamentRepositoryモック
type SharedMockTournamentRepository struct {
	tournaments map[int]*models.Tournament
	sportIndex  map[string]*models.Tournament
	err         error
	nextID      int
}

func NewSharedMockTournamentRepository() *SharedMockTournamentRepository {
	return &SharedMockTournamentRepository{
		tournaments: make(map[int]*models.Tournament),
		sportIndex:  make(map[string]*models.Tournament),
		nextID:      1,
	}
}

func (m *SharedMockTournamentRepository) Create(tournament *models.Tournament) error {
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

func (m *SharedMockTournamentRepository) GetByID(id int) (*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return nil, errors.New("トーナメントが見つかりません")
	}
	return tournament, nil
}

func (m *SharedMockTournamentRepository) GetBySport(sport string) (*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, exists := m.sportIndex[sport]
	if !exists {
		return nil, errors.New("トーナメントが見つかりません")
	}
	return tournament, nil
}

func (m *SharedMockTournamentRepository) GetAll() ([]*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournaments := make([]*models.Tournament, 0, len(m.tournaments))
	for _, tournament := range m.tournaments {
		tournaments = append(tournaments, tournament)
	}
	return tournaments, nil
}

func (m *SharedMockTournamentRepository) GetByStatus(status string) ([]*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournaments := make([]*models.Tournament, 0)
	for _, tournament := range m.tournaments {
		if tournament.Status == status {
			tournaments = append(tournaments, tournament)
		}
	}
	return tournaments, nil
}

func (m *SharedMockTournamentRepository) Update(tournament *models.Tournament) error {
	if m.err != nil {
		return m.err
	}
	
	existing, exists := m.tournaments[tournament.ID]
	if !exists {
		return errors.New("トーナメントが見つかりません")
	}
	
	// スポーツインデックスを更新
	delete(m.sportIndex, existing.Sport)
	
	tournament.UpdatedAt = time.Now()
	m.tournaments[tournament.ID] = tournament
	m.sportIndex[tournament.Sport] = tournament
	return nil
}

func (m *SharedMockTournamentRepository) Delete(id int) error {
	if m.err != nil {
		return m.err
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return errors.New("トーナメントが見つかりません")
	}
	
	delete(m.tournaments, id)
	delete(m.sportIndex, tournament.Sport)
	return nil
}

func (m *SharedMockTournamentRepository) GetTournamentBracket(sport string) (*models.Bracket, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, exists := m.sportIndex[sport]
	if !exists {
		return nil, errors.New("トーナメントが見つかりません")
	}
	
	bracket := &models.Bracket{
		TournamentID: tournament.ID,
		Sport:        sport,
		Format:       tournament.Format,
		Rounds:       []models.Round{},
	}
	return bracket, nil
}

func (m *SharedMockTournamentRepository) UpdateFormat(id int, format string) error {
	if m.err != nil {
		return m.err
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return errors.New("トーナメントが見つかりません")
	}
	
	tournament.Format = format
	tournament.UpdatedAt = time.Now()
	return nil
}

func (m *SharedMockTournamentRepository) UpdateStatus(id int, status string) error {
	if m.err != nil {
		return m.err
	}
	
	tournament, exists := m.tournaments[id]
	if !exists {
		return errors.New("トーナメントが見つかりません")
	}
	
	tournament.Status = status
	tournament.UpdatedAt = time.Now()
	return nil
}

func (m *SharedMockTournamentRepository) SetError(err error) {
	m.err = err
}

// SharedMockMatchRepository はテスト用の共有MatchRepositoryモック
type SharedMockMatchRepository struct {
	matches map[int]*models.Match
	tournamentMatches map[int][]*models.Match
	err     error
	nextID  int
}

func NewSharedMockMatchRepository() *SharedMockMatchRepository {
	return &SharedMockMatchRepository{
		matches: make(map[int]*models.Match),
		tournamentMatches: make(map[int][]*models.Match),
		nextID:  1,
	}
}

func (m *SharedMockMatchRepository) Create(match *models.Match) error {
	if m.err != nil {
		return m.err
	}
	
	match.ID = m.nextID
	m.nextID++
	
	m.matches[match.ID] = match
	
	if m.tournamentMatches[match.TournamentID] == nil {
		m.tournamentMatches[match.TournamentID] = make([]*models.Match, 0)
	}
	m.tournamentMatches[match.TournamentID] = append(m.tournamentMatches[match.TournamentID], match)
	
	return nil
}

func (m *SharedMockMatchRepository) GetByID(id int) (*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	match, exists := m.matches[id]
	if !exists {
		return nil, errors.New("試合が見つかりません")
	}
	return match, nil
}

func (m *SharedMockMatchRepository) Update(match *models.Match) error {
	if m.err != nil {
		return m.err
	}
	
	existing, exists := m.matches[match.ID]
	if !exists {
		return errors.New("試合が見つかりません")
	}
	
	// トーナメント別リストも更新
	tournamentMatches := m.tournamentMatches[existing.TournamentID]
	for i, tm := range tournamentMatches {
		if tm.ID == match.ID {
			tournamentMatches[i] = match
			break
		}
	}
	
	m.matches[match.ID] = match
	return nil
}

func (m *SharedMockMatchRepository) Delete(id int) error {
	if m.err != nil {
		return m.err
	}
	
	match, exists := m.matches[id]
	if !exists {
		return errors.New("試合が見つかりません")
	}
	
	// トーナメント別リストからも削除
	tournamentMatches := m.tournamentMatches[match.TournamentID]
	for i, tm := range tournamentMatches {
		if tm.ID == id {
			m.tournamentMatches[match.TournamentID] = append(tournamentMatches[:i], tournamentMatches[i+1:]...)
			break
		}
	}
	
	delete(m.matches, id)
	return nil
}

func (m *SharedMockMatchRepository) GetByTournament(tournamentID int) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	matches, exists := m.tournamentMatches[tournamentID]
	if !exists {
		return []*models.Match{}, nil
	}
	return matches, nil
}

func (m *SharedMockMatchRepository) GetBySport(sport string) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	// 簡略化: 全ての試合を返す
	matches := make([]*models.Match, 0, len(m.matches))
	for _, match := range m.matches {
		matches = append(matches, match)
	}
	return matches, nil
}

func (m *SharedMockMatchRepository) CountByTournament(tournamentID int) (int, error) {
	if m.err != nil {
		return 0, m.err
	}
	
	matches, exists := m.tournamentMatches[tournamentID]
	if !exists {
		return 0, nil
	}
	return len(matches), nil
}

func (m *SharedMockMatchRepository) DeleteByTournament(tournamentID int) error {
	if m.err != nil {
		return m.err
	}
	
	matches, exists := m.tournamentMatches[tournamentID]
	if !exists {
		return nil
	}
	
	// 個別の試合マップからも削除
	for _, match := range matches {
		delete(m.matches, match.ID)
	}
	
	delete(m.tournamentMatches, tournamentID)
	return nil
}

func (m *SharedMockMatchRepository) SetError(err error) {
	m.err = err
}

// SharedMockTournamentService はテスト用の共有TournamentServiceモック
type SharedMockTournamentService struct {
	tournaments map[string]*models.Tournament
	nextID      int
	err         error
}

func NewSharedMockTournamentService() *SharedMockTournamentService {
	return &SharedMockTournamentService{
		tournaments: make(map[string]*models.Tournament),
		nextID:      1,
	}
}

func (m *SharedMockTournamentService) CreateTournament(sport, format string) (*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	if !models.IsValidSport(sport) {
		return nil, errors.New("無効なスポーツです")
	}
	if !models.IsValidTournamentFormat(format) {
		return nil, errors.New("無効なフォーマットです")
	}
	
	tournament := &models.Tournament{
		ID:        m.nextID,
		Sport:     sport,
		Format:    format,
		Status:    models.TournamentStatusActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.nextID++
	m.tournaments[sport] = tournament
	return tournament, nil
}

func (m *SharedMockTournamentService) GetTournament(sport string) (*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, exists := m.tournaments[sport]
	if !exists {
		return nil, errors.New("トーナメントが見つかりません")
	}
	return tournament, nil
}

func (m *SharedMockTournamentService) GetTournamentByID(id int) (*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	for _, tournament := range m.tournaments {
		if tournament.ID == id {
			return tournament, nil
		}
	}
	return nil, errors.New("トーナメントが見つかりません")
}

func (m *SharedMockTournamentService) UpdateTournament(tournament *models.Tournament) error {
	if m.err != nil {
		return m.err
	}
	
	if tournament == nil {
		return errors.New("トーナメントがnilです")
	}
	existing, exists := m.tournaments[tournament.Sport]
	if !exists || existing.ID != tournament.ID {
		return errors.New("トーナメントが見つかりません")
	}
	m.tournaments[tournament.Sport] = tournament
	return nil
}

func (m *SharedMockTournamentService) DeleteTournament(id int) error {
	if m.err != nil {
		return m.err
	}
	
	for sport, tournament := range m.tournaments {
		if tournament.ID == id {
			delete(m.tournaments, sport)
			return nil
		}
	}
	return errors.New("トーナメントが見つかりません")
}

func (m *SharedMockTournamentService) GetTournamentBracket(sport string) (*models.Bracket, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, exists := m.tournaments[sport]
	if !exists {
		return nil, errors.New("トーナメントが見つかりません")
	}
	
	bracket := &models.Bracket{
		TournamentID: tournament.ID,
		Sport:        sport,
		Format:       tournament.Format,
		Rounds:       []models.Round{},
	}
	return bracket, nil
}

func (m *SharedMockTournamentService) GenerateBracket(sport, format string, teams []string) (*models.Bracket, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &models.Bracket{}, nil
}

func (m *SharedMockTournamentService) InitializeTournament(sport string, teams []string) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *SharedMockTournamentService) SwitchTournamentFormat(sport, newFormat string) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *SharedMockTournamentService) CompleteTournament(sport string) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *SharedMockTournamentService) ActivateTournament(sport string) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func (m *SharedMockTournamentService) GetAllTournaments() ([]*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournaments := make([]*models.Tournament, 0, len(m.tournaments))
	for _, tournament := range m.tournaments {
		tournaments = append(tournaments, tournament)
	}
	return tournaments, nil
}

func (m *SharedMockTournamentService) GetActiveTournaments() ([]*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournaments := make([]*models.Tournament, 0)
	for _, tournament := range m.tournaments {
		if tournament.Status == models.TournamentStatusActive {
			tournaments = append(tournaments, tournament)
		}
	}
	return tournaments, nil
}

func (m *SharedMockTournamentService) GetTournamentProgress(sport string) (*TournamentProgress, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &TournamentProgress{}, nil
}

func (m *SharedMockTournamentService) SetError(err error) {
	m.err = err
}

func (m *SharedMockTournamentRepository) GetActiveByFormat(format string) ([]*models.Tournament, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournaments := make([]*models.Tournament, 0)
	for _, tournament := range m.tournaments {
		if tournament.Status == models.TournamentStatusActive && tournament.Format == format {
			tournaments = append(tournaments, tournament)
		}
	}
	return tournaments, nil
}

func (m *SharedMockTournamentRepository) GetTournamentBracketByID(tournamentID int) (*models.Bracket, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	tournament, exists := m.tournaments[tournamentID]
	if !exists {
		return nil, errors.New("トーナメントが見つかりません")
	}
	
	bracket := &models.Bracket{
		TournamentID: tournament.ID,
		Sport:        tournament.Sport,
		Format:       tournament.Format,
		Rounds:       []models.Round{},
	}
	return bracket, nil
}

func (m *SharedMockMatchRepository) UpdateResult(matchID int, result models.MatchResult) error {
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
	now := time.Now()
	match.CompletedAt = &now
	
	return nil
}

func (m *SharedMockMatchRepository) UpdateStatus(matchID int, status string) error {
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

func (m *SharedMockMatchRepository) GetByTournamentAndRound(tournamentID int, round string) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	matches, exists := m.tournamentMatches[tournamentID]
	if !exists {
		return []*models.Match{}, nil
	}
	
	roundMatches := make([]*models.Match, 0)
	for _, match := range matches {
		if match.Round == round {
			roundMatches = append(roundMatches, match)
		}
	}
	
	return roundMatches, nil
}

func (m *SharedMockMatchRepository) GetByStatus(status string) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	matches := make([]*models.Match, 0)
	for _, match := range m.matches {
		if match.Status == status {
			matches = append(matches, match)
		}
	}
	return matches, nil
}

func (m *SharedMockMatchRepository) GetPendingMatches() ([]*models.Match, error) {
	return m.GetByStatus(models.MatchStatusPending)
}

func (m *SharedMockMatchRepository) GetCompletedMatches() ([]*models.Match, error) {
	return m.GetByStatus(models.MatchStatusCompleted)
}

func (m *SharedMockMatchRepository) GetMatchesByDateRange(startDate, endDate time.Time) ([]*models.Match, error) {
	if m.err != nil {
		return nil, m.err
	}
	
	matches := make([]*models.Match, 0)
	for _, match := range m.matches {
		if match.ScheduledAt.After(startDate) && match.ScheduledAt.Before(endDate) {
			matches = append(matches, match)
		}
	}
	return matches, nil
}

func (m *SharedMockMatchRepository) CountByStatus(status string) (int, error) {
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
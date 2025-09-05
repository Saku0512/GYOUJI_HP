package service

import (
	"context"
	"backend/internal/models"
	"backend/internal/repository"
)

// MatchService defines the interface for match operations
type MatchService interface {
	// Basic CRUD operations
	CreateMatch(match *models.Match) error
	GetMatch(id int) (*models.Match, error)
	GetMatches(ctx context.Context, limit, offset int) ([]*models.Match, error)
	UpdateMatch(match *models.Match) error
	DeleteMatch(id int) error
	
	// Query operations
	GetMatchesByTournament(tournamentID int) ([]*models.Match, error)
	GetMatchesBySport(sport string) ([]*models.Match, error)
	GetPendingMatches() ([]*models.Match, error)
	GetCompletedMatches() ([]*models.Match, error)
	GetNextMatches(tournamentID int) ([]*models.Match, error)
	
	// Match result operations
	UpdateMatchResult(matchID int, result models.MatchResult) error
	
	// Statistics
	GetMatchStatistics(tournamentID int) (*MatchStatistics, error)
}

// MatchStatistics represents match statistics
type MatchStatistics struct {
	TournamentID     int                    `json:"tournament_id"`
	TotalMatches     int                    `json:"total_matches"`
	CompletedMatches int                    `json:"completed_matches"`
	PendingMatches   int                    `json:"pending_matches"`
	MatchesByRound   map[string]int         `json:"matches_by_round"`
	CompletionRate   float64                `json:"completion_rate"`
	AverageScore     map[string]float64     `json:"average_score"`
	TeamStats        map[string]*TeamStats  `json:"team_stats"`
}

// TeamStats represents team statistics
type TeamStats struct {
	TeamName      string  `json:"team_name"`
	MatchesPlayed int     `json:"matches_played"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	TotalScore    int     `json:"total_score"`
	AverageScore  float64 `json:"average_score"`
}

// matchService implements MatchService
type matchService struct {
	matchRepo repository.MatchRepository
}

// NewMatchService creates a new match service
func NewMatchService(matchRepo repository.MatchRepository) MatchService {
	return &matchService{
		matchRepo: matchRepo,
	}
}

// CreateMatch creates a new match
func (s *matchService) CreateMatch(match *models.Match) error {
	if err := s.matchRepo.Create(context.Background(), match); err != nil {
		logger.Error("Failed to create match", "error", err)
		return NewDatabaseError("failed to create match")
	}
	return nil
}

// GetMatch retrieves a match by ID
func (s *matchService) GetMatch(id int) (*models.Match, error) {
	match, err := s.matchRepo.GetByID(context.Background(), uint(id))
	if err != nil {
		logger.Error("Failed to get match", "id", id, "error", err)
		return nil, NewDatabaseError("failed to get match")
	}
	if match == nil {
		return nil, NewNotFoundError("match not found")
	}
	return match, nil
}

// GetMatches retrieves matches with pagination
func (s *matchService) GetMatches(ctx context.Context, limit, offset int) ([]*models.Match, error) {
	matches, err := s.matchRepo.GetAll(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to get matches", "error", err)
		return nil, NewDatabaseError("failed to get matches")
	}
	return matches, nil
}

// UpdateMatch updates an existing match
func (s *matchService) UpdateMatch(match *models.Match) error {
	if err := s.matchRepo.Update(context.Background(), match); err != nil {
		logger.Error("Failed to update match", "id", match.ID, "error", err)
		return NewDatabaseError("failed to update match")
	}
	return nil
}

// DeleteMatch deletes a match
func (s *matchService) DeleteMatch(id int) error {
	if err := s.matchRepo.Delete(context.Background(), uint(id)); err != nil {
		logger.Error("Failed to delete match", "id", id, "error", err)
		return NewDatabaseError("failed to delete match")
	}
	return nil
}

// GetMatchesByTournament retrieves matches for a specific tournament
func (s *matchService) GetMatchesByTournament(tournamentID int) ([]*models.Match, error) {
	matches, err := s.matchRepo.GetByTournamentID(context.Background(), uint(tournamentID))
	if err != nil {
		logger.Error("Failed to get matches by tournament", "tournamentID", tournamentID, "error", err)
		return nil, NewDatabaseError("failed to get matches by tournament")
	}
	return matches, nil
}

// GetMatchesBySport retrieves matches by sport (simplified implementation)
func (s *matchService) GetMatchesBySport(sport string) ([]*models.Match, error) {
	// This is a simplified implementation
	// In a real system, you would need to join with tournaments table
	matches, err := s.matchRepo.GetAll(context.Background(), 100, 0)
	if err != nil {
		logger.Error("Failed to get matches by sport", "sport", sport, "error", err)
		return nil, NewDatabaseError("failed to get matches by sport")
	}
	return matches, nil
}

// GetPendingMatches retrieves pending matches
func (s *matchService) GetPendingMatches() ([]*models.Match, error) {
	matches, err := s.matchRepo.GetAll(context.Background(), 100, 0)
	if err != nil {
		logger.Error("Failed to get pending matches", "error", err)
		return nil, NewDatabaseError("failed to get pending matches")
	}
	
	var pendingMatches []*models.Match
	for _, match := range matches {
		if match.Status == "pending" {
			pendingMatches = append(pendingMatches, match)
		}
	}
	return pendingMatches, nil
}

// GetCompletedMatches retrieves completed matches
func (s *matchService) GetCompletedMatches() ([]*models.Match, error) {
	matches, err := s.matchRepo.GetAll(context.Background(), 100, 0)
	if err != nil {
		logger.Error("Failed to get completed matches", "error", err)
		return nil, NewDatabaseError("failed to get completed matches")
	}
	
	var completedMatches []*models.Match
	for _, match := range matches {
		if match.Status == "completed" {
			completedMatches = append(completedMatches, match)
		}
	}
	return completedMatches, nil
}

// GetNextMatches retrieves next matches for a tournament
func (s *matchService) GetNextMatches(tournamentID int) ([]*models.Match, error) {
	matches, err := s.GetMatchesByTournament(tournamentID)
	if err != nil {
		return nil, err
	}
	
	var nextMatches []*models.Match
	for _, match := range matches {
		if match.Status == "pending" {
			nextMatches = append(nextMatches, match)
		}
	}
	return nextMatches, nil
}

// UpdateMatchResult updates match result
func (s *matchService) UpdateMatchResult(matchID int, result models.MatchResult) error {
	match, err := s.GetMatch(matchID)
	if err != nil {
		return err
	}
	
	match.Score1 = &result.Score1
	match.Score2 = &result.Score2
	match.Winner = &result.Winner
	match.Status = "completed"
	
	return s.UpdateMatch(match)
}

// GetMatchStatistics retrieves match statistics for a tournament
func (s *matchService) GetMatchStatistics(tournamentID int) (*MatchStatistics, error) {
	matches, err := s.GetMatchesByTournament(tournamentID)
	if err != nil {
		return nil, err
	}
	
	stats := &MatchStatistics{
		TournamentID:   tournamentID,
		TotalMatches:   len(matches),
		MatchesByRound: make(map[string]int),
		AverageScore:   make(map[string]float64),
		TeamStats:      make(map[string]*TeamStats),
	}
	
	completedCount := 0
	for _, match := range matches {
		stats.MatchesByRound[match.Round]++
		
		if match.Status == "completed" {
			completedCount++
		}
	}
	
	stats.CompletedMatches = completedCount
	stats.PendingMatches = stats.TotalMatches - completedCount
	
	if stats.TotalMatches > 0 {
		stats.CompletionRate = float64(completedCount) / float64(stats.TotalMatches) * 100
	}
	
	return stats, nil
}
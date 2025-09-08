package service

import (
	"context"
	"backend/internal/models"
	"backend/internal/repository"
)

// TournamentService defines the interface for tournament operations
type TournamentService interface {
	CreateTournament(ctx context.Context, tournament *models.Tournament) error
	GetTournament(ctx context.Context, id uint) (*models.Tournament, error)
	GetTournaments(ctx context.Context, limit, offset int) ([]*models.Tournament, error)
	GetTournamentBySport(ctx context.Context, sport string, limit, offset int) ([]*models.Tournament, error)
	UpdateTournament(ctx context.Context, id uint, tournament *models.Tournament) error
	DeleteTournament(ctx context.Context, id uint) error
	GenerateBracket(ctx context.Context, tournamentID uint) error
	GetBracket(ctx context.Context, tournamentID uint) ([]*models.Match, error)
	UpdateMatchResult(ctx context.Context, matchID uint, team1Score, team2Score int, winnerID uint) error
	AdvanceWinner(ctx context.Context, matchID uint) error
	GetTournamentProgress(sport string) (*TournamentProgress, error)
}

// TournamentProgress represents tournament progress information
type TournamentProgress struct {
	TournamentID     int     `json:"tournament_id"`
	Sport            string  `json:"sport"`
	Format           string  `json:"format"`
	Status           string  `json:"status"`
	TotalMatches     int     `json:"total_matches"`
	CompletedMatches int     `json:"completed_matches"`
	PendingMatches   int     `json:"pending_matches"`
	CompletionRate   float64 `json:"completion_rate"`
	ProgressPercent  float64 `json:"progress_percent"`
	CurrentRound     string  `json:"current_round"`
	NextMatches      int     `json:"next_matches"`
}

// tournamentService implements TournamentService
type tournamentService struct {
	tournamentRepo      repository.TournamentRepository
	teamRepo            repository.TeamRepository
	matchRepo           repository.MatchRepository
	notificationService *NotificationService
}

// NewTournamentService creates a new tournament service
func NewTournamentService(
	tournamentRepo repository.TournamentRepository,
	teamRepo repository.TeamRepository,
	matchRepo repository.MatchRepository,
) TournamentService {
	return &tournamentService{
		tournamentRepo: tournamentRepo,
		teamRepo:       teamRepo,
		matchRepo:      matchRepo,
	}
}

// CreateTournament creates a new tournament
func (s *tournamentService) CreateTournament(ctx context.Context, tournament *models.Tournament) error {
	if tournament.Sport == "" {
		return NewValidationError("tournament sport is required")
	}
	if tournament.Format == "" {
		return NewValidationError("tournament format is required")
	}

	// Check if tournament with same sport exists
	existing, err := s.tournamentRepo.GetByName(ctx, tournament.Sport)
	if err != nil {
		logger.Error("Failed to check existing tournament", "error", err)
		return NewDatabaseError("failed to check existing tournament")
	}
	if existing != nil {
		return NewConflictError("tournament with this sport already exists")
	}

	if err := s.tournamentRepo.Create(ctx, tournament); err != nil {
		logger.Error("Failed to create tournament", "error", err)
		return NewDatabaseError("failed to create tournament")
	}

	// Send real-time notification
	if s.notificationService != nil {
		s.notificationService.NotifyTournamentUpdate(tournament, "created")
	}

	return nil
}

// GetTournament retrieves a tournament by ID
func (s *tournamentService) GetTournament(ctx context.Context, id uint) (*models.Tournament, error) {
	tournament, err := s.tournamentRepo.GetByID(ctx, id)
	if err != nil {
		logger.Error("Failed to get tournament", "id", id, "error", err)
		return nil, NewDatabaseError("failed to get tournament")
	}
	if tournament == nil {
		return nil, NewNotFoundError("tournament not found")
	}
	return tournament, nil
}

// GetTournaments retrieves tournaments with pagination
func (s *tournamentService) GetTournaments(ctx context.Context, limit, offset int) ([]*models.Tournament, error) {
	tournaments, err := s.tournamentRepo.GetAll(ctx, limit, offset)
	if err != nil {
		logger.Error("Failed to get tournaments", "error", err)
		return nil, NewDatabaseError("failed to get tournaments")
	}
	return tournaments, nil
}

// GetTournamentBySport retrieves tournaments by sport
func (s *tournamentService) GetTournamentBySport(ctx context.Context, sport string, limit, offset int) ([]*models.Tournament, error) {
	tournaments, err := s.tournamentRepo.GetBySport(ctx, sport, limit, offset)
	if err != nil {
		logger.Error("Failed to get tournaments by sport", "sport", sport, "error", err)
		return nil, NewDatabaseError("failed to get tournaments by sport")
	}
	return tournaments, nil
}

// UpdateTournament updates an existing tournament
func (s *tournamentService) UpdateTournament(ctx context.Context, id uint, tournament *models.Tournament) error {
	existing, err := s.GetTournament(ctx, id)
	if err != nil {
		return err
	}

	if tournament.Sport != "" {
		existing.Sport = tournament.Sport
	}
	if tournament.Format != "" {
		existing.Format = tournament.Format
	}
	if tournament.Status != "" {
		existing.Status = tournament.Status
	}

	if err := s.tournamentRepo.Update(ctx, existing); err != nil {
		logger.Error("Failed to update tournament", "id", id, "error", err)
		return NewDatabaseError("failed to update tournament")
	}

	// Send real-time notification
	if s.notificationService != nil {
		s.notificationService.NotifyTournamentUpdate(existing, "updated")
	}

	return nil
}

// DeleteTournament deletes a tournament
func (s *tournamentService) DeleteTournament(ctx context.Context, id uint) error {
	tournament, err := s.GetTournament(ctx, id)
	if err != nil {
		return err
	}

	if err := s.tournamentRepo.Delete(ctx, uint(tournament.ID)); err != nil {
		logger.Error("Failed to delete tournament", "id", id, "error", err)
		return NewDatabaseError("failed to delete tournament")
	}
	return nil
}

// GenerateBracket generates tournament bracket
func (s *tournamentService) GenerateBracket(ctx context.Context, tournamentID uint) error {
	tournament, err := s.GetTournament(ctx, tournamentID)
	if err != nil {
		return err
	}

	if tournament.Status != "registration" {
		return NewValidationError("can only generate bracket for tournaments in registration status")
	}

	// Get registered teams
	teams, err := s.teamRepo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		logger.Error("Failed to get teams for tournament", "tournamentID", tournamentID, "error", err)
		return NewDatabaseError("failed to get teams")
	}

	if len(teams) < 2 {
		return NewValidationError("need at least 2 teams to generate bracket")
	}

	// Check if teams count is power of 2
	teamCount := len(teams)
	if !isPowerOfTwo(teamCount) {
		return NewValidationError("number of teams must be a power of 2")
	}

	// Generate matches for first round
	matches := make([]*models.Match, 0)
	for i := 0; i < teamCount; i += 2 {
		match := &models.Match{
			TournamentID: int(tournamentID),
			Team1:        teams[i].Name,
			Team2:        teams[i+1].Name,
			Round:        "1st_round",
			Status:       "pending",
		}
		matches = append(matches, match)
	}

	// Create matches in database
	for _, match := range matches {
		if err := s.matchRepo.Create(ctx, match); err != nil {
			logger.Error("Failed to create match", "error", err)
			return NewDatabaseError("failed to create match")
		}
	}

	// Update tournament status
	tournament.Status = "active"
	if err := s.tournamentRepo.Update(ctx, tournament); err != nil {
		logger.Error("Failed to update tournament status", "error", err)
		return NewDatabaseError("failed to update tournament status")
	}

	// Send real-time notification for tournament status change
	if s.notificationService != nil {
		s.notificationService.NotifyTournamentUpdate(tournament, "status_changed")
	}

	return nil
}

// GetBracket retrieves tournament bracket
func (s *tournamentService) GetBracket(ctx context.Context, tournamentID uint) ([]*models.Match, error) {
	matches, err := s.matchRepo.GetByTournamentID(ctx, tournamentID)
	if err != nil {
		logger.Error("Failed to get matches for tournament", "tournamentID", tournamentID, "error", err)
		return nil, NewDatabaseError("failed to get matches")
	}
	return matches, nil
}

// UpdateMatchResult updates match result and advances winner
func (s *tournamentService) UpdateMatchResult(ctx context.Context, matchID uint, team1Score, team2Score int, winnerID uint) error {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		logger.Error("Failed to get match", "matchID", matchID, "error", err)
		return NewDatabaseError("failed to get match")
	}
	if match == nil {
		return NewNotFoundError("match not found")
	}

	if match.Status != "pending" && match.Status != "in_progress" {
		return NewValidationError("can only update result for pending or in-progress matches")
	}

	// Update match result
	match.Score1 = &team1Score
	match.Score2 = &team2Score
	
	// Determine winner based on scores
	if team1Score > team2Score {
		match.Winner = &match.Team1
	} else {
		match.Winner = &match.Team2
	}
	
	match.Status = "completed"

	if err := s.matchRepo.Update(ctx, match); err != nil {
		logger.Error("Failed to update match", "matchID", matchID, "error", err)
		return NewDatabaseError("failed to update match")
	}

	// Advance winner to next round
	return s.AdvanceWinner(ctx, matchID)
}

// AdvanceWinner advances the winner to the next round
func (s *tournamentService) AdvanceWinner(ctx context.Context, matchID uint) error {
	match, err := s.matchRepo.GetByID(ctx, matchID)
	if err != nil {
		return NewDatabaseError("failed to get match")
	}
	if match.Winner == nil {
		return NewValidationError("match has no winner")
	}

	// This is a simplified implementation
	// In a real system, you would need more complex logic to determine the next match
	logger.Info("Winner advanced", "winner", *match.Winner, "matchID", matchID)
	
	return nil
}

// GetTournamentProgress retrieves tournament progress information
func (s *tournamentService) GetTournamentProgress(sport string) (*TournamentProgress, error) {
	// Get tournament by sport
	tournaments, err := s.tournamentRepo.GetBySport(context.Background(), sport, 1, 0)
	if err != nil {
		logger.Error("Failed to get tournament by sport", "sport", sport, "error", err)
		return nil, NewDatabaseError("failed to get tournament")
	}
	
	if len(tournaments) == 0 {
		return nil, NewNotFoundError("tournament not found")
	}
	
	tournament := tournaments[0]
	
	// Get matches for this tournament
	matches, err := s.matchRepo.GetByTournamentID(context.Background(), uint(tournament.ID))
	if err != nil {
		logger.Error("Failed to get matches for tournament", "tournamentID", tournament.ID, "error", err)
		return nil, NewDatabaseError("failed to get matches")
	}
	
	// Calculate progress statistics
	totalMatches := len(matches)
	completedMatches := 0
	currentRound := ""
	
	for _, match := range matches {
		if match.Status == "completed" {
			completedMatches++
		}
		if match.Status == "pending" && currentRound == "" {
			currentRound = match.Round
		}
	}
	
	pendingMatches := totalMatches - completedMatches
	completionRate := 0.0
	if totalMatches > 0 {
		completionRate = float64(completedMatches) / float64(totalMatches) * 100
	}
	
	progress := &TournamentProgress{
		TournamentID:     tournament.ID,
		Sport:            tournament.Sport,
		Format:           tournament.Format,
		Status:           tournament.Status,
		TotalMatches:     totalMatches,
		CompletedMatches: completedMatches,
		PendingMatches:   pendingMatches,
		CompletionRate:   completionRate,
		ProgressPercent:  completionRate,
		CurrentRound:     currentRound,
		NextMatches:      pendingMatches,
	}
	
	return progress, nil
}

// Helper function to check if number is power of 2
func isPowerOfTwo(n int) bool {
	return n > 0 && (n&(n-1)) == 0
}

// SetNotificationService sets the notification service for real-time updates
func (s *tournamentService) SetNotificationService(notificationService *NotificationService) {
	s.notificationService = notificationService
}
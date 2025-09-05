package repository

import (
	"context"
	"database/sql"
	"log"

	"backend/internal/database"
	"backend/internal/models"
)

// MatchRepository defines the interface for match data operations
type MatchRepository interface {
	Create(ctx context.Context, match *models.Match) error
	GetByID(ctx context.Context, id uint) (*models.Match, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.Match, error)
	GetByTournamentID(ctx context.Context, tournamentID uint) ([]*models.Match, error)
	GetByRoundAndPosition(ctx context.Context, tournamentID uint, round, position int) (*models.Match, error)
	Update(ctx context.Context, match *models.Match) error
	Delete(ctx context.Context, id uint) error
}

// matchRepository implements MatchRepository
type matchRepository struct {
	base BaseRepository
}

// NewMatchRepository creates a new match repository
func NewMatchRepository(db *database.DB) MatchRepository {
	if db == nil {
		log.Fatal("データベース接続がnilです")
	}
	
	baseRepo := NewBaseRepository(db)
	return &matchRepository{
		base: baseRepo,
	}
}

// Create creates a new match
func (r *matchRepository) Create(ctx context.Context, match *models.Match) error {
	query := `
		INSERT INTO matches (tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	
	result, err := r.base.ExecQuery(query,
		match.TournamentID,
		match.Round,
		match.Team1,
		match.Team2,
		match.Score1,
		match.Score2,
		match.Winner,
		match.Status,
		match.ScheduledAt,
		match.CompletedAt,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	match.ID = int(id)
	return nil
}

// GetByID retrieves a match by ID
func (r *matchRepository) GetByID(ctx context.Context, id uint) (*models.Match, error) {
	query := `
		SELECT id, tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at, created_at, updated_at
		FROM matches
		WHERE id = ?
	`
	
	row := r.base.QueryRow(query, id)
	
	match := &models.Match{}
	err := row.Scan(
		&match.ID,
		&match.TournamentID,
		&match.Round,
		&match.Team1,
		&match.Team2,
		&match.Score1,
		&match.Score2,
		&match.Winner,
		&match.Status,
		&match.ScheduledAt,
		&match.CompletedAt,
		&match.CreatedAt,
		&match.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return match, nil
}

// GetAll retrieves all matches with pagination
func (r *matchRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Match, error) {
	query := `
		SELECT id, tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at, created_at, updated_at
		FROM matches
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.base.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanMatches(rows)
}

// GetByTournamentID retrieves matches by tournament ID
func (r *matchRepository) GetByTournamentID(ctx context.Context, tournamentID uint) ([]*models.Match, error) {
	query := `
		SELECT id, tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at, created_at, updated_at
		FROM matches
		WHERE tournament_id = ?
		ORDER BY scheduled_at ASC
	`
	
	rows, err := r.base.Query(query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanMatches(rows)
}

// GetByRoundAndPosition retrieves a match by tournament, round and position
// Note: This method is kept for interface compatibility but position is not used in the current model
func (r *matchRepository) GetByRoundAndPosition(ctx context.Context, tournamentID uint, round, position int) (*models.Match, error) {
	query := `
		SELECT id, tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at, created_at, updated_at
		FROM matches
		WHERE tournament_id = ? AND round = ?
		LIMIT 1
	`
	
	row := r.base.QueryRow(query, tournamentID, round)
	
	match := &models.Match{}
	err := row.Scan(
		&match.ID,
		&match.TournamentID,
		&match.Round,
		&match.Team1,
		&match.Team2,
		&match.Score1,
		&match.Score2,
		&match.Winner,
		&match.Status,
		&match.ScheduledAt,
		&match.CompletedAt,
		&match.CreatedAt,
		&match.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return match, nil
}

// Update updates an existing match
func (r *matchRepository) Update(ctx context.Context, match *models.Match) error {
	query := `
		UPDATE matches
		SET tournament_id = ?, round = ?, team1 = ?, team2 = ?, score1 = ?, score2 = ?, winner = ?, status = ?, scheduled_at = ?, completed_at = ?, updated_at = NOW()
		WHERE id = ?
	`
	
	_, err := r.base.ExecQuery(query,
		match.TournamentID,
		match.Round,
		match.Team1,
		match.Team2,
		match.Score1,
		match.Score2,
		match.Winner,
		match.Status,
		match.ScheduledAt,
		match.CompletedAt,
		match.ID,
	)
	
	return err
}

// Delete deletes a match
func (r *matchRepository) Delete(ctx context.Context, id uint) error {
	query := `DELETE FROM matches WHERE id = ?`
	
	_, err := r.base.ExecQuery(query, id)
	return err
}

// scanMatches scans multiple match rows
func (r *matchRepository) scanMatches(rows *sql.Rows) ([]*models.Match, error) {
	var matches []*models.Match
	
	for rows.Next() {
		match := &models.Match{}
		err := rows.Scan(
			&match.ID,
			&match.TournamentID,
			&match.Round,
			&match.Team1,
			&match.Team2,
			&match.Score1,
			&match.Score2,
			&match.Winner,
			&match.Status,
			&match.ScheduledAt,
			&match.CompletedAt,
			&match.CreatedAt,
			&match.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		matches = append(matches, match)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return matches, nil
}
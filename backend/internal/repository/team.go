package repository

import (
	"context"
	"database/sql"
	"log"

	"backend/internal/database"
	"backend/internal/models"
)

// TeamRepository defines the interface for team data operations
type TeamRepository interface {
	Create(ctx context.Context, team *models.Team) error
	GetByID(ctx context.Context, id uint) (*models.Team, error)
	GetByName(ctx context.Context, name string) (*models.Team, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.Team, error)
	GetByTournamentID(ctx context.Context, tournamentID uint) ([]*models.Team, error)
	Update(ctx context.Context, team *models.Team) error
	Delete(ctx context.Context, id uint) error
}

// teamRepository implements TeamRepository
type teamRepository struct {
	base BaseRepository
}

// NewTeamRepository creates a new team repository
func NewTeamRepository(db *database.DB) TeamRepository {
	if db == nil {
		log.Fatal("データベース接続がnilです")
	}
	
	baseRepo := NewBaseRepository(db)
	return &teamRepository{
		base: baseRepo,
	}
}

// Create creates a new team
func (r *teamRepository) Create(ctx context.Context, team *models.Team) error {
	query := `
		INSERT INTO teams (name, description, tournament_id, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`
	
	result, err := r.base.ExecQuery(query,
		team.Name,
		team.Description,
		team.TournamentID,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	team.ID = uint(id)
	return nil
}

// GetByID retrieves a team by ID
func (r *teamRepository) GetByID(ctx context.Context, id uint) (*models.Team, error) {
	query := `
		SELECT id, name, description, tournament_id, created_at, updated_at
		FROM teams
		WHERE id = ?
	`
	
	row := r.base.QueryRow(query, id)
	
	team := &models.Team{}
	err := row.Scan(
		&team.ID,
		&team.Name,
		&team.Description,
		&team.TournamentID,
		&team.CreatedAt,
		&team.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return team, nil
}

// GetByName retrieves a team by name
func (r *teamRepository) GetByName(ctx context.Context, name string) (*models.Team, error) {
	query := `
		SELECT id, name, description, tournament_id, created_at, updated_at
		FROM teams
		WHERE name = ?
	`
	
	row := r.base.QueryRow(query, name)
	
	team := &models.Team{}
	err := row.Scan(
		&team.ID,
		&team.Name,
		&team.Description,
		&team.TournamentID,
		&team.CreatedAt,
		&team.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return team, nil
}

// GetAll retrieves all teams with pagination
func (r *teamRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Team, error) {
	query := `
		SELECT id, name, description, tournament_id, created_at, updated_at
		FROM teams
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.base.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanTeams(rows)
}

// GetByTournamentID retrieves teams by tournament ID
func (r *teamRepository) GetByTournamentID(ctx context.Context, tournamentID uint) ([]*models.Team, error) {
	query := `
		SELECT id, name, description, tournament_id, created_at, updated_at
		FROM teams
		WHERE tournament_id = ?
		ORDER BY name ASC
	`
	
	rows, err := r.base.Query(query, tournamentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanTeams(rows)
}

// Update updates an existing team
func (r *teamRepository) Update(ctx context.Context, team *models.Team) error {
	query := `
		UPDATE teams
		SET name = ?, description = ?, tournament_id = ?, updated_at = NOW()
		WHERE id = ?
	`
	
	_, err := r.base.ExecQuery(query,
		team.Name,
		team.Description,
		team.TournamentID,
		team.ID,
	)
	
	return err
}

// Delete deletes a team
func (r *teamRepository) Delete(ctx context.Context, id uint) error {
	query := `DELETE FROM teams WHERE id = ?`
	
	_, err := r.base.ExecQuery(query, id)
	return err
}

// scanTeams scans multiple team rows
func (r *teamRepository) scanTeams(rows *sql.Rows) ([]*models.Team, error) {
	var teams []*models.Team
	
	for rows.Next() {
		team := &models.Team{}
		err := rows.Scan(
			&team.ID,
			&team.Name,
			&team.Description,
			&team.TournamentID,
			&team.CreatedAt,
			&team.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		teams = append(teams, team)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return teams, nil
}
package repository

import (
	"context"
	"database/sql"
	"log"

	"backend/internal/database"
	"backend/internal/models"
)

// TournamentRepository defines the interface for tournament data operations
type TournamentRepository interface {
	Create(ctx context.Context, tournament *models.Tournament) error
	GetByID(ctx context.Context, id uint) (*models.Tournament, error)
	GetByName(ctx context.Context, name string) (*models.Tournament, error)
	GetAll(ctx context.Context, limit, offset int) ([]*models.Tournament, error)
	GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Tournament, error)
	GetBySport(ctx context.Context, sport string, limit, offset int) ([]*models.Tournament, error)
	Update(ctx context.Context, tournament *models.Tournament) error
	Delete(ctx context.Context, id uint) error
}

// tournamentRepository implements TournamentRepository
type tournamentRepository struct {
	base BaseRepository
}

// NewTournamentRepository creates a new tournament repository
func NewTournamentRepository(db *database.DB) TournamentRepository {
	if db == nil {
		log.Fatal("データベース接続がnilです")
	}
	
	baseRepo := NewBaseRepository(db)
	return &tournamentRepository{
		base: baseRepo,
	}
}

// Create creates a new tournament
func (r *tournamentRepository) Create(ctx context.Context, tournament *models.Tournament) error {
	query := `
		INSERT INTO tournaments (sport, format, status, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`
	
	result, err := r.base.ExecQuery(query,
		tournament.Sport,
		tournament.Format,
		tournament.Status,
	)
	
	if err != nil {
		return err
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	
	tournament.ID = int(id)
	return nil
}

// GetByID retrieves a tournament by ID
func (r *tournamentRepository) GetByID(ctx context.Context, id uint) (*models.Tournament, error) {
	query := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		WHERE id = ?
	`
	
	row := r.base.QueryRow(query, id)
	
	tournament := &models.Tournament{}
	err := row.Scan(
		&tournament.ID,
		&tournament.Sport,
		&tournament.Format,
		&tournament.Status,
		&tournament.CreatedAt,
		&tournament.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return tournament, nil
}

// GetByName retrieves a tournament by name (using sport as identifier)
func (r *tournamentRepository) GetByName(ctx context.Context, name string) (*models.Tournament, error) {
	query := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		WHERE sport = ?
		LIMIT 1
	`
	
	row := r.base.QueryRow(query, name)
	
	tournament := &models.Tournament{}
	err := row.Scan(
		&tournament.ID,
		&tournament.Sport,
		&tournament.Format,
		&tournament.Status,
		&tournament.CreatedAt,
		&tournament.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	
	return tournament, nil
}

// GetAll retrieves all tournaments with pagination
func (r *tournamentRepository) GetAll(ctx context.Context, limit, offset int) ([]*models.Tournament, error) {
	query := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.base.Query(query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanTournaments(rows)
}

// GetByStatus retrieves tournaments by status with pagination
func (r *tournamentRepository) GetByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Tournament, error) {
	query := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		WHERE status = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.base.Query(query, status, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanTournaments(rows)
}

// Update updates an existing tournament
func (r *tournamentRepository) Update(ctx context.Context, tournament *models.Tournament) error {
	query := `
		UPDATE tournaments
		SET sport = ?, format = ?, status = ?, updated_at = NOW()
		WHERE id = ?
	`
	
	_, err := r.base.ExecQuery(query,
		tournament.Sport,
		tournament.Format,
		tournament.Status,
		tournament.ID,
	)
	
	return err
}

// GetBySport retrieves tournaments by sport with pagination
func (r *tournamentRepository) GetBySport(ctx context.Context, sport string, limit, offset int) ([]*models.Tournament, error) {
	query := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		WHERE sport = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`
	
	rows, err := r.base.Query(query, sport, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	return r.scanTournaments(rows)
}

// Delete deletes a tournament
func (r *tournamentRepository) Delete(ctx context.Context, id uint) error {
	query := `DELETE FROM tournaments WHERE id = ?`
	
	_, err := r.base.ExecQuery(query, id)
	return err
}

// scanTournaments scans multiple tournament rows
func (r *tournamentRepository) scanTournaments(rows *sql.Rows) ([]*models.Tournament, error) {
	var tournaments []*models.Tournament
	
	for rows.Next() {
		tournament := &models.Tournament{}
		err := rows.Scan(
			&tournament.ID,
			&tournament.Sport,
			&tournament.Format,
			&tournament.Status,
			&tournament.CreatedAt,
			&tournament.UpdatedAt,
		)
		
		if err != nil {
			return nil, err
		}
		
		tournaments = append(tournaments, tournament)
	}
	
	if err := rows.Err(); err != nil {
		return nil, err
	}
	
	return tournaments, nil
}
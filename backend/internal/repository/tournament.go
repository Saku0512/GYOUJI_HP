// Package repository はデータアクセス層の実装を提供する
package repository

import (
	"database/sql"
	"log"
	"time"

	"backend/internal/database"
	"backend/internal/models"
)

// TournamentRepository はトーナメント関連のデータアクセスを提供するインターフェース
type TournamentRepository interface {
	// トーナメントのCRUD操作
	Create(tournament *models.Tournament) error
	GetByID(id int) (*models.Tournament, error)
	GetBySport(sport string) (*models.Tournament, error)
	GetAll() ([]*models.Tournament, error)
	Update(tournament *models.Tournament) error
	Delete(id int) error
	
	// ブラケット関連操作
	GetTournamentBracket(sport string) (*models.Bracket, error)
	GetTournamentBracketByID(tournamentID int) (*models.Bracket, error)
	
	// トーナメント状態管理
	UpdateStatus(id int, status string) error
	UpdateFormat(id int, format string) error
	
	// 検索・フィルタリング
	GetByStatus(status string) ([]*models.Tournament, error)
	GetActiveByFormat(format string) ([]*models.Tournament, error)
}

// tournamentRepositoryImpl はTournamentRepositoryの実装
type tournamentRepositoryImpl struct {
	base BaseRepository
}

// NewTournamentRepository は新しいTournamentRepositoryインスタンスを作成する
func NewTournamentRepository(db *database.DB) TournamentRepository {
	if db == nil {
		log.Fatal("データベース接続がnilです")
	}
	
	baseRepo := NewBaseRepository(db)
	return &tournamentRepositoryImpl{
		base: baseRepo,
	}
}

// Create は新しいトーナメントを作成する
func (r *tournamentRepositoryImpl) Create(tournament *models.Tournament) error {
	if tournament == nil {
		return NewRepositoryError(ErrTypeValidation, "トーナメントがnilです", nil)
	}
	
	if err := tournament.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "トーナメントの検証に失敗しました", err)
	}
	
	query := `
		INSERT INTO tournaments (sport, format, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
	`
	
	now := time.Now()
	tournament.CreatedAt = now
	tournament.UpdatedAt = now
	
	result, err := r.base.ExecQuery(query, 
		tournament.Sport, 
		tournament.Format, 
		tournament.Status, 
		tournament.CreatedAt, 
		tournament.UpdatedAt,
	)
	
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "トーナメントの作成に失敗しました", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "作成されたトーナメントのIDの取得に失敗しました", err)
	}
	
	tournament.ID = int(id)
	log.Printf("トーナメントを作成しました: ID=%d, Sport=%s", tournament.ID, tournament.Sport)
	
	return nil
}

// GetByID はIDでトーナメントを取得する
func (r *tournamentRepositoryImpl) GetByID(id int) (*models.Tournament, error) {
	if id <= 0 {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
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
			return nil, NewRepositoryError(ErrTypeNotFound, "トーナメントが見つかりません", err)
		}
		return nil, NewRepositoryError(ErrTypeQuery, "トーナメントの取得に失敗しました", err)
	}
	
	return tournament, nil
}

// GetBySport はスポーツでトーナメントを取得する
func (r *tournamentRepositoryImpl) GetBySport(sport string) (*models.Tournament, error) {
	if !models.IsValidSport(sport) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なスポーツです", nil)
	}
	
	query := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		WHERE sport = ?
		ORDER BY created_at DESC
		LIMIT 1
	`
	
	row := r.base.QueryRow(query, sport)
	
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
			return nil, NewRepositoryError(ErrTypeNotFound, "指定されたスポーツのトーナメントが見つかりません", err)
		}
		return nil, NewRepositoryError(ErrTypeQuery, "トーナメントの取得に失敗しました", err)
	}
	
	return tournament, nil
}

// GetAll は全てのトーナメントを取得する
func (r *tournamentRepositoryImpl) GetAll() ([]*models.Tournament, error) {
	query := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		ORDER BY created_at DESC
	`
	
	rows, err := r.base.Query(query)
	if err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "トーナメント一覧の取得に失敗しました", err)
	}
	defer rows.Close()
	
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
			return nil, NewRepositoryError(ErrTypeQuery, "トーナメントデータの読み取りに失敗しました", err)
		}
		
		tournaments = append(tournaments, tournament)
	}
	
	if err = rows.Err(); err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "トーナメント一覧の処理中にエラーが発生しました", err)
	}
	
	return tournaments, nil
}

// Update はトーナメントを更新する
func (r *tournamentRepositoryImpl) Update(tournament *models.Tournament) error {
	if tournament == nil {
		return NewRepositoryError(ErrTypeValidation, "トーナメントがnilです", nil)
	}
	
	if tournament.ID <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	if err := tournament.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "トーナメントの検証に失敗しました", err)
	}
	
	query := `
		UPDATE tournaments
		SET sport = ?, format = ?, status = ?, updated_at = ?
		WHERE id = ?
	`
	
	tournament.UpdatedAt = time.Now()
	
	result, err := r.base.ExecQuery(query,
		tournament.Sport,
		tournament.Format,
		tournament.Status,
		tournament.UpdatedAt,
		tournament.ID,
	)
	
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "トーナメントの更新に失敗しました", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "更新された行数の取得に失敗しました", err)
	}
	
	if rowsAffected == 0 {
		return NewRepositoryError(ErrTypeNotFound, "更新対象のトーナメントが見つかりません", nil)
	}
	
	log.Printf("トーナメントを更新しました: ID=%d", tournament.ID)
	return nil
}

// Delete はトーナメントを削除する
func (r *tournamentRepositoryImpl) Delete(id int) error {
	if id <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	query := `DELETE FROM tournaments WHERE id = ?`
	
	result, err := r.base.ExecQuery(query, id)
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "トーナメントの削除に失敗しました", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "削除された行数の取得に失敗しました", err)
	}
	
	if rowsAffected == 0 {
		return NewRepositoryError(ErrTypeNotFound, "削除対象のトーナメントが見つかりません", nil)
	}
	
	log.Printf("トーナメントを削除しました: ID=%d", id)
	return nil
}

// GetTournamentBracket はスポーツに基づいてトーナメントブラケットを取得する
func (r *tournamentRepositoryImpl) GetTournamentBracket(sport string) (*models.Bracket, error) {
	if !models.IsValidSport(sport) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なスポーツです", nil)
	}
	
	// まずトーナメントを取得
	tournament, err := r.GetBySport(sport)
	if err != nil {
		return nil, err
	}
	
	return r.GetTournamentBracketByID(tournament.ID)
}

// GetTournamentBracketByID はトーナメントIDに基づいてブラケットを取得する
func (r *tournamentRepositoryImpl) GetTournamentBracketByID(tournamentID int) (*models.Bracket, error) {
	if tournamentID <= 0 {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	// トーナメント情報を取得
	tournament, err := r.GetByID(tournamentID)
	if err != nil {
		return nil, err
	}
	
	// 試合データを取得
	query := `
		SELECT id, tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at
		FROM matches
		WHERE tournament_id = ?
		ORDER BY 
			CASE round
				WHEN '1st_round' THEN 1
				WHEN 'quarterfinal' THEN 2
				WHEN 'semifinal' THEN 3
				WHEN 'third_place' THEN 4
				WHEN 'final' THEN 5
				WHEN 'loser_bracket' THEN 6
				ELSE 7
			END,
			id
	`
	
	rows, err := r.base.Query(query, tournamentID)
	if err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "試合データの取得に失敗しました", err)
	}
	defer rows.Close()
	
	// ラウンド別に試合を整理
	roundMatches := make(map[string][]models.Match)
	
	for rows.Next() {
		match := models.Match{}
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
		)
		
		if err != nil {
			return nil, NewRepositoryError(ErrTypeQuery, "試合データの読み取りに失敗しました", err)
		}
		
		roundMatches[match.Round] = append(roundMatches[match.Round], match)
	}
	
	if err = rows.Err(); err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "試合データの処理中にエラーが発生しました", err)
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
	
	// 卓球の雨天時フォーマットの場合、敗者復活戦を含める
	if tournament.Sport == models.SportTableTennis && tournament.Format == models.FormatRainy {
		// 敗者復活戦が含まれていない場合は追加
		hasLoserBracket := false
		for _, round := range validRounds {
			if round == models.RoundLoserBracket {
				hasLoserBracket = true
				break
			}
		}
		if !hasLoserBracket {
			validRounds = append(validRounds, models.RoundLoserBracket)
		}
	}
	
	// 各ラウンドのデータを構築
	for _, roundName := range validRounds {
		matches, exists := roundMatches[roundName]
		if !exists {
			matches = []models.Match{} // 空のスライスを作成
		}
		
		round := models.Round{
			Name:    roundName,
			Matches: matches,
		}
		
		bracket.Rounds = append(bracket.Rounds, round)
	}
	
	return bracket, nil
}

// UpdateStatus はトーナメントのステータスを更新する
func (r *tournamentRepositoryImpl) UpdateStatus(id int, status string) error {
	if id <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	if !models.IsValidTournamentStatus(status) {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントステータスです", nil)
	}
	
	query := `
		UPDATE tournaments
		SET status = ?, updated_at = ?
		WHERE id = ?
	`
	
	result, err := r.base.ExecQuery(query, status, time.Now(), id)
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "トーナメントステータスの更新に失敗しました", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "更新された行数の取得に失敗しました", err)
	}
	
	if rowsAffected == 0 {
		return NewRepositoryError(ErrTypeNotFound, "更新対象のトーナメントが見つかりません", nil)
	}
	
	log.Printf("トーナメントステータスを更新しました: ID=%d, Status=%s", id, status)
	return nil
}

// UpdateFormat はトーナメントのフォーマットを更新する
func (r *tournamentRepositoryImpl) UpdateFormat(id int, format string) error {
	if id <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	if !models.IsValidTournamentFormat(format) {
		return NewRepositoryError(ErrTypeValidation, "無効なトーナメントフォーマットです", nil)
	}
	
	query := `
		UPDATE tournaments
		SET format = ?, updated_at = ?
		WHERE id = ?
	`
	
	result, err := r.base.ExecQuery(query, format, time.Now(), id)
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "トーナメントフォーマットの更新に失敗しました", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "更新された行数の取得に失敗しました", err)
	}
	
	if rowsAffected == 0 {
		return NewRepositoryError(ErrTypeNotFound, "更新対象のトーナメントが見つかりません", nil)
	}
	
	log.Printf("トーナメントフォーマットを更新しました: ID=%d, Format=%s", id, format)
	return nil
}

// GetByStatus はステータスでトーナメントを取得する
func (r *tournamentRepositoryImpl) GetByStatus(status string) ([]*models.Tournament, error) {
	if !models.IsValidTournamentStatus(status) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントステータスです", nil)
	}
	
	query := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		WHERE status = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.base.Query(query, status)
	if err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "ステータス別トーナメント一覧の取得に失敗しました", err)
	}
	defer rows.Close()
	
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
			return nil, NewRepositoryError(ErrTypeQuery, "トーナメントデータの読み取りに失敗しました", err)
		}
		
		tournaments = append(tournaments, tournament)
	}
	
	if err = rows.Err(); err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "ステータス別トーナメント一覧の処理中にエラーが発生しました", err)
	}
	
	return tournaments, nil
}

// GetActiveByFormat はフォーマットでアクティブなトーナメントを取得する
func (r *tournamentRepositoryImpl) GetActiveByFormat(format string) ([]*models.Tournament, error) {
	if !models.IsValidTournamentFormat(format) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントフォーマットです", nil)
	}
	
	query := `
		SELECT id, sport, format, status, created_at, updated_at
		FROM tournaments
		WHERE format = ? AND status = ?
		ORDER BY created_at DESC
	`
	
	rows, err := r.base.Query(query, format, models.TournamentStatusActive)
	if err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "フォーマット別アクティブトーナメント一覧の取得に失敗しました", err)
	}
	defer rows.Close()
	
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
			return nil, NewRepositoryError(ErrTypeQuery, "トーナメントデータの読み取りに失敗しました", err)
		}
		
		tournaments = append(tournaments, tournament)
	}
	
	if err = rows.Err(); err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "フォーマット別アクティブトーナメント一覧の処理中にエラーが発生しました", err)
	}
	
	return tournaments, nil
}
// Package repository はデータアクセス層の実装を提供する
package repository

import (
	"database/sql"
	"log"
	"time"

	"backend/internal/database"
	"backend/internal/models"
)

// MatchRepository は試合関連のデータアクセスを提供するインターフェース
type MatchRepository interface {
	// 試合のCRUD操作
	Create(match *models.Match) error
	GetByID(id int) (*models.Match, error)
	Update(match *models.Match) error
	Delete(id int) error
	
	// 試合結果更新機能
	UpdateResult(matchID int, result models.MatchResult) error
	UpdateStatus(matchID int, status string) error
	
	// スポーツ/トーナメント別クエリ
	GetBySport(sport string) ([]*models.Match, error)
	GetByTournament(tournamentID int) ([]*models.Match, error)
	GetByTournamentAndRound(tournamentID int, round string) ([]*models.Match, error)
	GetByStatus(status string) ([]*models.Match, error)
	
	// 検索・フィルタリング
	GetPendingMatches() ([]*models.Match, error)
	GetCompletedMatches() ([]*models.Match, error)
	GetMatchesByDateRange(startDate, endDate time.Time) ([]*models.Match, error)
	
	// 統計・集計
	CountByTournament(tournamentID int) (int, error)
	CountByStatus(status string) (int, error)
}

// matchRepositoryImpl はMatchRepositoryの実装
type matchRepositoryImpl struct {
	base BaseRepository
}

// NewMatchRepository は新しいMatchRepositoryインスタンスを作成する
func NewMatchRepository(db *database.DB) MatchRepository {
	if db == nil {
		log.Fatal("データベース接続がnilです")
	}
	
	baseRepo := NewBaseRepository(db)
	return &matchRepositoryImpl{
		base: baseRepo,
	}
}

// Create は新しい試合を作成する
func (r *matchRepositoryImpl) Create(match *models.Match) error {
	if match == nil {
		return NewRepositoryError(ErrTypeValidation, "試合がnilです", nil)
	}
	
	if err := match.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "試合の検証に失敗しました", err)
	}
	
	query := `
		INSERT INTO matches (tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		return NewRepositoryError(ErrTypeQuery, "試合の作成に失敗しました", err)
	}
	
	id, err := result.LastInsertId()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "作成された試合のIDの取得に失敗しました", err)
	}
	
	match.ID = int(id)
	log.Printf("試合を作成しました: ID=%d, Tournament=%d, %s vs %s", 
		match.ID, match.TournamentID, match.Team1, match.Team2)
	
	return nil
}

// GetByID はIDで試合を取得する
func (r *matchRepositoryImpl) GetByID(id int) (*models.Match, error) {
	if id <= 0 {
		return nil, NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	query := `
		SELECT id, tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at
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
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, NewRepositoryError(ErrTypeNotFound, "試合が見つかりません", err)
		}
		return nil, NewRepositoryError(ErrTypeQuery, "試合の取得に失敗しました", err)
	}
	
	return match, nil
}

// Update は試合を更新する
func (r *matchRepositoryImpl) Update(match *models.Match) error {
	if match == nil {
		return NewRepositoryError(ErrTypeValidation, "試合がnilです", nil)
	}
	
	if match.ID <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	if err := match.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "試合の検証に失敗しました", err)
	}
	
	query := `
		UPDATE matches
		SET tournament_id = ?, round = ?, team1 = ?, team2 = ?, score1 = ?, score2 = ?, winner = ?, status = ?, scheduled_at = ?, completed_at = ?
		WHERE id = ?
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
		match.ID,
	)
	
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "試合の更新に失敗しました", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "更新された行数の取得に失敗しました", err)
	}
	
	if rowsAffected == 0 {
		return NewRepositoryError(ErrTypeNotFound, "更新対象の試合が見つかりません", nil)
	}
	
	log.Printf("試合を更新しました: ID=%d", match.ID)
	return nil
}

// Delete は試合を削除する
func (r *matchRepositoryImpl) Delete(id int) error {
	if id <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	query := `DELETE FROM matches WHERE id = ?`
	
	result, err := r.base.ExecQuery(query, id)
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "試合の削除に失敗しました", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "削除された行数の取得に失敗しました", err)
	}
	
	if rowsAffected == 0 {
		return NewRepositoryError(ErrTypeNotFound, "削除対象の試合が見つかりません", nil)
	}
	
	log.Printf("試合を削除しました: ID=%d", id)
	return nil
}

// UpdateResult は試合結果を更新する
func (r *matchRepositoryImpl) UpdateResult(matchID int, result models.MatchResult) error {
	if matchID <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	if err := result.Validate(); err != nil {
		return NewRepositoryError(ErrTypeValidation, "試合結果の検証に失敗しました", err)
	}
	
	// 試合が存在するかチェック
	existingMatch, err := r.GetByID(matchID)
	if err != nil {
		return err
	}
	
	// 既に完了している試合の結果は更新できない
	if existingMatch.IsCompleted() {
		return NewRepositoryError(ErrTypeValidation, "完了済みの試合結果は更新できません", nil)
	}
	
	query := `
		UPDATE matches
		SET score1 = ?, score2 = ?, winner = ?, status = ?, completed_at = ?
		WHERE id = ?
	`
	
	completedAt := time.Now()
	
	result_exec, err := r.base.ExecQuery(query,
		result.Score1,
		result.Score2,
		result.Winner,
		models.MatchStatusCompleted,
		completedAt,
		matchID,
	)
	
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "試合結果の更新に失敗しました", err)
	}
	
	rowsAffected, err := result_exec.RowsAffected()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "更新された行数の取得に失敗しました", err)
	}
	
	if rowsAffected == 0 {
		return NewRepositoryError(ErrTypeNotFound, "更新対象の試合が見つかりません", nil)
	}
	
	log.Printf("試合結果を更新しました: ID=%d, Winner=%s, Score=%d-%d", 
		matchID, result.Winner, result.Score1, result.Score2)
	return nil
}

// UpdateStatus は試合のステータスを更新する
func (r *matchRepositoryImpl) UpdateStatus(matchID int, status string) error {
	if matchID <= 0 {
		return NewRepositoryError(ErrTypeValidation, "無効な試合IDです", nil)
	}
	
	if !models.IsValidMatchStatus(status) {
		return NewRepositoryError(ErrTypeValidation, "無効な試合ステータスです", nil)
	}
	
	query := `
		UPDATE matches
		SET status = ?
		WHERE id = ?
	`
	
	result, err := r.base.ExecQuery(query, status, matchID)
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "試合ステータスの更新に失敗しました", err)
	}
	
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return NewRepositoryError(ErrTypeQuery, "更新された行数の取得に失敗しました", err)
	}
	
	if rowsAffected == 0 {
		return NewRepositoryError(ErrTypeNotFound, "更新対象の試合が見つかりません", nil)
	}
	
	log.Printf("試合ステータスを更新しました: ID=%d, Status=%s", matchID, status)
	return nil
}

// GetBySport はスポーツで試合を取得する
func (r *matchRepositoryImpl) GetBySport(sport string) ([]*models.Match, error) {
	if !models.IsValidSport(sport) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なスポーツです", nil)
	}
	
	query := `
		SELECT m.id, m.tournament_id, m.round, m.team1, m.team2, m.score1, m.score2, m.winner, m.status, m.scheduled_at, m.completed_at
		FROM matches m
		INNER JOIN tournaments t ON m.tournament_id = t.id
		WHERE t.sport = ?
		ORDER BY m.scheduled_at ASC, m.id ASC
	`
	
	rows, err := r.base.Query(query, sport)
	if err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "スポーツ別試合一覧の取得に失敗しました", err)
	}
	defer rows.Close()
	
	return r.scanMatches(rows)
}

// GetByTournament はトーナメントで試合を取得する
func (r *matchRepositoryImpl) GetByTournament(tournamentID int) ([]*models.Match, error) {
	if tournamentID <= 0 {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
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
			scheduled_at ASC,
			id ASC
	`
	
	rows, err := r.base.Query(query, tournamentID)
	if err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "トーナメント別試合一覧の取得に失敗しました", err)
	}
	defer rows.Close()
	
	return r.scanMatches(rows)
}

// GetByTournamentAndRound はトーナメントとラウンドで試合を取得する
func (r *matchRepositoryImpl) GetByTournamentAndRound(tournamentID int, round string) ([]*models.Match, error) {
	if tournamentID <= 0 {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	if !models.IsValidRound(round) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効なラウンドです", nil)
	}
	
	query := `
		SELECT id, tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at
		FROM matches
		WHERE tournament_id = ? AND round = ?
		ORDER BY scheduled_at ASC, id ASC
	`
	
	rows, err := r.base.Query(query, tournamentID, round)
	if err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "トーナメント・ラウンド別試合一覧の取得に失敗しました", err)
	}
	defer rows.Close()
	
	return r.scanMatches(rows)
}

// GetByStatus はステータスで試合を取得する
func (r *matchRepositoryImpl) GetByStatus(status string) ([]*models.Match, error) {
	if !models.IsValidMatchStatus(status) {
		return nil, NewRepositoryError(ErrTypeValidation, "無効な試合ステータスです", nil)
	}
	
	query := `
		SELECT id, tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at
		FROM matches
		WHERE status = ?
		ORDER BY scheduled_at ASC, id ASC
	`
	
	rows, err := r.base.Query(query, status)
	if err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "ステータス別試合一覧の取得に失敗しました", err)
	}
	defer rows.Close()
	
	return r.scanMatches(rows)
}

// GetPendingMatches は未実施の試合を取得する
func (r *matchRepositoryImpl) GetPendingMatches() ([]*models.Match, error) {
	return r.GetByStatus(models.MatchStatusPending)
}

// GetCompletedMatches は完了した試合を取得する
func (r *matchRepositoryImpl) GetCompletedMatches() ([]*models.Match, error) {
	return r.GetByStatus(models.MatchStatusCompleted)
}

// GetMatchesByDateRange は日付範囲で試合を取得する
func (r *matchRepositoryImpl) GetMatchesByDateRange(startDate, endDate time.Time) ([]*models.Match, error) {
	if startDate.After(endDate) {
		return nil, NewRepositoryError(ErrTypeValidation, "開始日は終了日より前である必要があります", nil)
	}
	
	query := `
		SELECT id, tournament_id, round, team1, team2, score1, score2, winner, status, scheduled_at, completed_at
		FROM matches
		WHERE scheduled_at BETWEEN ? AND ?
		ORDER BY scheduled_at ASC, id ASC
	`
	
	rows, err := r.base.Query(query, startDate, endDate)
	if err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "日付範囲別試合一覧の取得に失敗しました", err)
	}
	defer rows.Close()
	
	return r.scanMatches(rows)
}

// CountByTournament はトーナメント別の試合数を取得する
func (r *matchRepositoryImpl) CountByTournament(tournamentID int) (int, error) {
	if tournamentID <= 0 {
		return 0, NewRepositoryError(ErrTypeValidation, "無効なトーナメントIDです", nil)
	}
	
	query := `SELECT COUNT(*) FROM matches WHERE tournament_id = ?`
	
	row := r.base.QueryRow(query, tournamentID)
	
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, NewRepositoryError(ErrTypeQuery, "トーナメント別試合数の取得に失敗しました", err)
	}
	
	return count, nil
}

// CountByStatus はステータス別の試合数を取得する
func (r *matchRepositoryImpl) CountByStatus(status string) (int, error) {
	if !models.IsValidMatchStatus(status) {
		return 0, NewRepositoryError(ErrTypeValidation, "無効な試合ステータスです", nil)
	}
	
	query := `SELECT COUNT(*) FROM matches WHERE status = ?`
	
	row := r.base.QueryRow(query, status)
	
	var count int
	err := row.Scan(&count)
	if err != nil {
		return 0, NewRepositoryError(ErrTypeQuery, "ステータス別試合数の取得に失敗しました", err)
	}
	
	return count, nil
}

// scanMatches は複数の試合データをスキャンする共通メソッド
func (r *matchRepositoryImpl) scanMatches(rows *sql.Rows) ([]*models.Match, error) {
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
		)
		
		if err != nil {
			return nil, NewRepositoryError(ErrTypeQuery, "試合データの読み取りに失敗しました", err)
		}
		
		matches = append(matches, match)
	}
	
	if err := rows.Err(); err != nil {
		return nil, NewRepositoryError(ErrTypeQuery, "試合一覧の処理中にエラーが発生しました", err)
	}
	
	return matches, nil
}
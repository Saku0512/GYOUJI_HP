package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
)

// TournamentService はトーナメント関連のビジネスロジックを提供するインターフェース
type TournamentService interface {
	// トーナメント管理
	CreateTournament(sport, format string) (*models.Tournament, error)
	GetTournament(sport string) (*models.Tournament, error)
	GetTournamentByID(id int) (*models.Tournament, error)
	UpdateTournament(tournament *models.Tournament) error
	DeleteTournament(id int) error
	
	// ブラケット生成と管理
	GetTournamentBracket(sport string) (*models.Bracket, error)
	GenerateBracket(sport, format string, teams []string) (*models.Bracket, error)
	InitializeTournament(sport string, teams []string) error
	
	// トーナメント形式切り替え（卓球の天候条件用）
	SwitchTournamentFormat(sport, newFormat string) error
	
	// トーナメント状態管理
	CompleteTournament(sport string) error
	ActivateTournament(sport string) error
	
	// 統計・情報取得
	GetAllTournaments() ([]*models.Tournament, error)
	GetActiveTournaments() ([]*models.Tournament, error)
	GetTournamentProgress(sport string) (*TournamentProgress, error)
}

// TournamentProgress はトーナメントの進行状況を表す構造体
type TournamentProgress struct {
	TournamentID     int     `json:"tournament_id"`
	Sport            string  `json:"sport"`
	Format           string  `json:"format"`
	Status           string  `json:"status"`
	TotalMatches     int     `json:"total_matches"`
	CompletedMatches int     `json:"completed_matches"`
	PendingMatches   int     `json:"pending_matches"`
	ProgressPercent  float64 `json:"progress_percent"`
	CurrentRound     string  `json:"current_round"`
}

// tournamentServiceImpl はTournamentServiceの実装
type tournamentServiceImpl struct {
	tournamentRepo repository.TournamentRepository
	matchRepo      repository.MatchRepository
}

// NewTournamentService は新しいTournamentServiceインスタンスを作成する
func NewTournamentService(tournamentRepo repository.TournamentRepository, matchRepo repository.MatchRepository) TournamentService {
	return &tournamentServiceImpl{
		tournamentRepo: tournamentRepo,
		matchRepo:      matchRepo,
	}
}

// CreateTournament は新しいトーナメントを作成する
func (s *tournamentServiceImpl) CreateTournament(sport, format string) (*models.Tournament, error) {
	log := logger.GetLogger()
	
	// 入力値の検証
	if !models.IsValidSport(sport) {
		return nil, errors.NewValidationError("無効なスポーツです")
	}
	
	if !models.IsValidTournamentFormat(format) {
		return nil, errors.NewValidationError("無効なトーナメントフォーマットです")
	}
	
	// 既存のアクティブなトーナメントをチェック
	existingTournament, err := s.tournamentRepo.GetBySport(sport)
	if err == nil && existingTournament.IsActive() {
		return nil, errors.NewConflictError(fmt.Sprintf("スポーツ %s には既にアクティブなトーナメントが存在します", sport))
	}
	
	// 新しいトーナメントを作成
	tournament := &models.Tournament{
		Sport:  sport,
		Format: format,
		Status: models.TournamentStatusActive,
	}
	
	err = s.tournamentRepo.Create(tournament)
	if err != nil {
		log.Error("トーナメント作成エラー", 
			logger.Err(err),
			logger.String("sport", sport),
			logger.String("format", format),
		)
		return nil, errors.NewDatabaseError("トーナメントの作成に失敗しました", err)
	}
	
	log.Info("トーナメントを作成しました", 
		logger.Int("id", tournament.ID),
		logger.String("sport", sport),
		logger.String("format", format),
	)
	return tournament, nil
}

// GetTournament はスポーツに基づいてトーナメントを取得する
func (s *tournamentServiceImpl) GetTournament(sport string) (*models.Tournament, error) {
	log := logger.GetLogger()
	
	if !models.IsValidSport(sport) {
		return nil, errors.NewValidationError("無効なスポーツです")
	}
	
	tournament, err := s.tournamentRepo.GetBySport(sport)
	if err != nil {
		log.Error("トーナメント取得エラー", 
			logger.Err(err),
			logger.String("sport", sport),
		)
		return nil, errors.NewNotFoundError(fmt.Sprintf("スポーツ %s のトーナメント", sport))
	}
	
	return tournament, nil
}

// GetTournamentByID はIDに基づいてトーナメントを取得する
func (s *tournamentServiceImpl) GetTournamentByID(id int) (*models.Tournament, error) {
	log := logger.GetLogger()
	
	if id <= 0 {
		return nil, errors.NewValidationError("無効なトーナメントIDです")
	}
	
	tournament, err := s.tournamentRepo.GetByID(id)
	if err != nil {
		log.Error("トーナメント取得エラー", 
			logger.Err(err),
			logger.Int("id", id),
		)
		return nil, errors.NewNotFoundError("トーナメント")
	}
	
	return tournament, nil
}

// UpdateTournament はトーナメントを更新する
func (s *tournamentServiceImpl) UpdateTournament(tournament *models.Tournament) error {
	log := logger.GetLogger()
	
	if tournament == nil {
		return errors.NewValidationError("トーナメントは必須です")
	}
	
	if tournament.ID <= 0 {
		return errors.NewValidationError("無効なトーナメントIDです")
	}
	
	// 既存のトーナメントが存在するかチェック
	_, err := s.tournamentRepo.GetByID(tournament.ID)
	if err != nil {
		return errors.NewNotFoundError("更新対象のトーナメント")
	}
	
	err = s.tournamentRepo.Update(tournament)
	if err != nil {
		log.Error("トーナメント更新エラー", 
			logger.Err(err),
			logger.Int("id", tournament.ID),
		)
		return errors.NewDatabaseError("トーナメントの更新に失敗しました", err)
	}
	
	log.Info("トーナメントを更新しました", logger.Int("id", tournament.ID))
	return nil
}

// DeleteTournament はトーナメントを削除する
func (s *tournamentServiceImpl) DeleteTournament(id int) error {
	if id <= 0 {
		return errors.NewValidationError("無効なトーナメントIDです")
	}
	
	// 関連する試合があるかチェック
	matchCount, err := s.matchRepo.CountByTournament(id)
	if err != nil {
		logger.GetLogger().Error("試合数取得エラー", logger.Err(err))
		return errors.NewDatabaseError("トーナメントの削除チェックに失敗しました", err)
	}
	
	if matchCount > 0 {
		return errors.NewBusinessLogicError("試合が存在するトーナメントは削除できません")
	}
	
	err = s.tournamentRepo.Delete(id)
	if err != nil {
		logger.GetLogger().Error("トーナメント削除エラー", logger.Err(err))
		return errors.NewDatabaseError("トーナメントの削除に失敗しました", err)
	}
	
	logger.GetLogger().Info("トーナメントを削除しました", logger.Int("id", id))
	return nil
}

// GetTournamentBracket はトーナメントブラケットを取得する
func (s *tournamentServiceImpl) GetTournamentBracket(sport string) (*models.Bracket, error) {
	log := logger.GetLogger()
	
	if !models.IsValidSport(sport) {
		return nil, errors.NewValidationError("無効なスポーツです")
	}
	
	bracket, err := s.tournamentRepo.GetTournamentBracket(sport)
	if err != nil {
		log.Error("ブラケット取得エラー", 
			logger.Err(err),
			logger.String("sport", sport),
		)
		return nil, errors.NewNotFoundError(fmt.Sprintf("スポーツ %s のブラケット", sport))
	}
	
	return bracket, nil
}

// GenerateBracket は指定されたスポーツとチームでブラケットを生成する
func (s *tournamentServiceImpl) GenerateBracket(sport, format string, teams []string) (*models.Bracket, error) {
	// 入力値の検証
	if !models.IsValidSport(sport) {
		return nil, errors.New("無効なスポーツです")
	}
	
	if !models.IsValidTournamentFormat(format) {
		return nil, errors.New("無効なトーナメントフォーマットです")
	}
	
	if len(teams) == 0 {
		return nil, errors.New("チームが指定されていません")
	}
	
	// スポーツに応じた最小チーム数をチェック
	minTeams := getMinimumTeamsForSport(sport)
	if len(teams) < minTeams {
		return nil, fmt.Errorf("スポーツ %s には最低 %d チームが必要です", sport, minTeams)
	}
	
	// トーナメントを取得または作成
	tournament, err := s.GetTournament(sport)
	if err != nil {
		// トーナメントが存在しない場合は作成
		tournament, err = s.CreateTournament(sport, format)
		if err != nil {
			return nil, err
		}
	}
	
	// ブラケット構造を生成
	bracket := &models.Bracket{
		TournamentID: tournament.ID,
		Sport:        sport,
		Format:       format,
		Rounds:       []models.Round{},
	}
	
	// スポーツに応じたブラケット生成
	switch sport {
	case models.SportVolleyball:
		err = s.generateVolleyballBracket(bracket, teams)
	case models.SportTableTennis:
		err = s.generateTableTennisBracket(bracket, teams, format)
	case models.SportSoccer:
		err = s.generateSoccerBracket(bracket, teams)
	default:
		return nil, fmt.Errorf("サポートされていないスポーツです: %s", sport)
	}
	
	if err != nil {
		log.Printf("ブラケット生成エラー: %v", err)
		return nil, errors.New("ブラケットの生成に失敗しました")
	}
	
	log.Printf("ブラケットを生成しました: Sport=%s, Format=%s, Teams=%d", sport, format, len(teams))
	return bracket, nil
}

// InitializeTournament はトーナメントを初期化し、試合を作成する
func (s *tournamentServiceImpl) InitializeTournament(sport string, teams []string) error {
	if !models.IsValidSport(sport) {
		return errors.New("無効なスポーツです")
	}
	
	if len(teams) == 0 {
		return errors.New("チームが指定されていません")
	}
	
	// デフォルトフォーマットでブラケットを生成
	bracket, err := s.GenerateBracket(sport, models.FormatStandard, teams)
	if err != nil {
		return err
	}
	
	// ブラケットから試合を作成
	for _, round := range bracket.Rounds {
		for _, match := range round.Matches {
			err = s.matchRepo.Create(&match)
			if err != nil {
				log.Printf("試合作成エラー: %v", err)
				return errors.New("試合の作成に失敗しました")
			}
		}
	}
	
	log.Printf("トーナメントを初期化しました: Sport=%s, Teams=%d", sport, len(teams))
	return nil
}

// SwitchTournamentFormat はトーナメント形式を切り替える（卓球の天候条件用）
func (s *tournamentServiceImpl) SwitchTournamentFormat(sport, newFormat string) error {
	if !models.IsValidSport(sport) {
		return errors.New("無効なスポーツです")
	}
	
	if !models.IsValidTournamentFormat(newFormat) {
		return errors.New("無効なトーナメントフォーマットです")
	}
	
	// 卓球のみフォーマット切り替えをサポート
	if sport != models.SportTableTennis {
		return fmt.Errorf("スポーツ %s はフォーマット切り替えをサポートしていません", sport)
	}
	
	// 現在のトーナメントを取得
	tournament, err := s.GetTournament(sport)
	if err != nil {
		return err
	}
	
	// 既に同じフォーマットの場合はエラー
	if tournament.Format == newFormat {
		return fmt.Errorf("トーナメントは既に %s フォーマットです", newFormat)
	}
	
	// 完了した試合があるかチェック
	completedMatches, err := s.matchRepo.GetByTournament(tournament.ID)
	if err != nil {
		return errors.New("試合データの取得に失敗しました")
	}
	
	hasCompletedMatches := false
	for _, match := range completedMatches {
		if match.IsCompleted() {
			hasCompletedMatches = true
			break
		}
	}
	
	if hasCompletedMatches {
		return errors.New("試合が開始されているトーナメントのフォーマットは変更できません")
	}
	
	// フォーマットを更新
	err = s.tournamentRepo.UpdateFormat(tournament.ID, newFormat)
	if err != nil {
		log.Printf("フォーマット更新エラー: %v", err)
		return errors.New("トーナメントフォーマットの更新に失敗しました")
	}
	
	log.Printf("トーナメントフォーマットを切り替えました: Sport=%s, NewFormat=%s", sport, newFormat)
	return nil
}

// CompleteTournament はトーナメントを完了状態にする
func (s *tournamentServiceImpl) CompleteTournament(sport string) error {
	if !models.IsValidSport(sport) {
		return errors.New("無効なスポーツです")
	}
	
	tournament, err := s.GetTournament(sport)
	if err != nil {
		return err
	}
	
	if tournament.IsCompleted() {
		return errors.New("トーナメントは既に完了しています")
	}
	
	// 全ての試合が完了しているかチェック
	bracket, err := s.GetTournamentBracket(sport)
	if err != nil {
		return err
	}
	
	if !bracket.IsCompleted() {
		return errors.New("全ての試合が完了していないため、トーナメントを完了できません")
	}
	
	err = s.tournamentRepo.UpdateStatus(tournament.ID, models.TournamentStatusCompleted)
	if err != nil {
		log.Printf("トーナメント完了エラー: %v", err)
		return errors.New("トーナメントの完了に失敗しました")
	}
	
	log.Printf("トーナメントを完了しました: Sport=%s", sport)
	return nil
}

// ActivateTournament はトーナメントをアクティブ状態にする
func (s *tournamentServiceImpl) ActivateTournament(sport string) error {
	if !models.IsValidSport(sport) {
		return errors.New("無効なスポーツです")
	}
	
	tournament, err := s.GetTournament(sport)
	if err != nil {
		return err
	}
	
	if tournament.IsActive() {
		return errors.New("トーナメントは既にアクティブです")
	}
	
	err = s.tournamentRepo.UpdateStatus(tournament.ID, models.TournamentStatusActive)
	if err != nil {
		log.Printf("トーナメントアクティブ化エラー: %v", err)
		return errors.New("トーナメントのアクティブ化に失敗しました")
	}
	
	log.Printf("トーナメントをアクティブ化しました: Sport=%s", sport)
	return nil
}

// GetAllTournaments は全てのトーナメントを取得する
func (s *tournamentServiceImpl) GetAllTournaments() ([]*models.Tournament, error) {
	tournaments, err := s.tournamentRepo.GetAll()
	if err != nil {
		log.Printf("全トーナメント取得エラー: %v", err)
		return nil, errors.New("トーナメント一覧の取得に失敗しました")
	}
	
	return tournaments, nil
}

// GetActiveTournaments はアクティブなトーナメントを取得する
func (s *tournamentServiceImpl) GetActiveTournaments() ([]*models.Tournament, error) {
	tournaments, err := s.tournamentRepo.GetByStatus(models.TournamentStatusActive)
	if err != nil {
		log.Printf("アクティブトーナメント取得エラー: %v", err)
		return nil, errors.New("アクティブトーナメント一覧の取得に失敗しました")
	}
	
	return tournaments, nil
}

// GetTournamentProgress はトーナメントの進行状況を取得する
func (s *tournamentServiceImpl) GetTournamentProgress(sport string) (*TournamentProgress, error) {
	if !models.IsValidSport(sport) {
		return nil, errors.New("無効なスポーツです")
	}
	
	tournament, err := s.GetTournament(sport)
	if err != nil {
		return nil, err
	}
	
	bracket, err := s.GetTournamentBracket(sport)
	if err != nil {
		return nil, err
	}
	
	totalMatches := bracket.GetTotalMatches()
	completedMatches := bracket.GetCompletedMatches()
	pendingMatches := totalMatches - completedMatches
	
	var progressPercent float64
	if totalMatches > 0 {
		progressPercent = float64(completedMatches) / float64(totalMatches) * 100
	}
	
	// 現在のラウンドを決定
	currentRound := s.determineCurrentRound(bracket)
	
	progress := &TournamentProgress{
		TournamentID:     tournament.ID,
		Sport:            tournament.Sport,
		Format:           tournament.Format,
		Status:           tournament.Status,
		TotalMatches:     totalMatches,
		CompletedMatches: completedMatches,
		PendingMatches:   pendingMatches,
		ProgressPercent:  progressPercent,
		CurrentRound:     currentRound,
	}
	
	return progress, nil
}

// getMinimumTeamsForSport はスポーツに必要な最小チーム数を返す
func getMinimumTeamsForSport(sport string) int {
	switch sport {
	case models.SportVolleyball:
		return 8 // 1回戦、準々決勝、準決勝、決勝
	case models.SportTableTennis:
		return 8 // 同様の構造
	case models.SportSoccer:
		return 8 // 8人制サッカー
	default:
		return 4 // デフォルト
	}
}

// determineCurrentRound は現在のラウンドを決定する
func (s *tournamentServiceImpl) determineCurrentRound(bracket *models.Bracket) string {
	for _, round := range bracket.Rounds {
		for _, match := range round.Matches {
			if match.IsPending() {
				return round.Name
			}
		}
	}
	
	// 全ての試合が完了している場合
	if len(bracket.Rounds) > 0 {
		return bracket.Rounds[len(bracket.Rounds)-1].Name
	}
	
	return ""
}

// generateVolleyballBracket はバレーボールのブラケットを生成する
func (s *tournamentServiceImpl) generateVolleyballBracket(bracket *models.Bracket, teams []string) error {
	if len(teams) < 8 {
		return errors.New("バレーボールには最低8チームが必要です")
	}
	
	// 8チームトーナメント: 1回戦(4試合) -> 準々決勝(2試合) -> 準決勝(2試合) -> 3位決定戦・決勝(2試合)
	rounds := []models.Round{
		{Name: models.Round1stRound, Matches: []models.Match{}},
		{Name: models.RoundQuarterfinal, Matches: []models.Match{}},
		{Name: models.RoundSemifinal, Matches: []models.Match{}},
		{Name: models.RoundThirdPlace, Matches: []models.Match{}},
		{Name: models.RoundFinal, Matches: []models.Match{}},
	}
	
	// 1回戦の試合を生成（8チーム -> 4試合）
	baseTime := time.Now().Add(time.Hour) // 1時間後から開始
	for i := 0; i < len(teams) && i < 8; i += 2 {
		if i+1 < len(teams) {
			match := models.Match{
				TournamentID: bracket.TournamentID,
				Round:        models.Round1stRound,
				Team1:        teams[i],
				Team2:        teams[i+1],
				Status:       models.MatchStatusPending,
				ScheduledAt:  baseTime.Add(time.Duration(i/2) * 30 * time.Minute),
			}
			rounds[0].Matches = append(rounds[0].Matches, match)
		}
	}
	
	// 準々決勝以降はプレースホルダーで生成
	s.generatePlaceholderMatches(&rounds[1], models.RoundQuarterfinal, bracket.TournamentID, 2, baseTime.Add(2*time.Hour))
	s.generatePlaceholderMatches(&rounds[2], models.RoundSemifinal, bracket.TournamentID, 2, baseTime.Add(4*time.Hour))
	s.generatePlaceholderMatches(&rounds[3], models.RoundThirdPlace, bracket.TournamentID, 1, baseTime.Add(6*time.Hour))
	s.generatePlaceholderMatches(&rounds[4], models.RoundFinal, bracket.TournamentID, 1, baseTime.Add(6*time.Hour+30*time.Minute))
	
	bracket.Rounds = rounds
	return nil
}

// generateTableTennisBracket は卓球のブラケットを生成する
func (s *tournamentServiceImpl) generateTableTennisBracket(bracket *models.Bracket, teams []string, format string) error {
	if len(teams) < 8 {
		return errors.New("卓球には最低8チームが必要です")
	}
	
	var rounds []models.Round
	
	if format == models.FormatRainy {
		// 雨天時フォーマット（敗者復活戦付き）
		rounds = []models.Round{
			{Name: models.Round1stRound, Matches: []models.Match{}},
			{Name: models.RoundQuarterfinal, Matches: []models.Match{}},
			{Name: models.RoundSemifinal, Matches: []models.Match{}},
			{Name: models.RoundLoserBracket, Matches: []models.Match{}}, // 敗者復活戦
			{Name: models.RoundThirdPlace, Matches: []models.Match{}},
			{Name: models.RoundFinal, Matches: []models.Match{}},
		}
	} else {
		// 標準フォーマット
		rounds = []models.Round{
			{Name: models.Round1stRound, Matches: []models.Match{}},
			{Name: models.RoundQuarterfinal, Matches: []models.Match{}},
			{Name: models.RoundSemifinal, Matches: []models.Match{}},
			{Name: models.RoundThirdPlace, Matches: []models.Match{}},
			{Name: models.RoundFinal, Matches: []models.Match{}},
		}
	}
	
	// 1回戦の試合を生成
	baseTime := time.Now().Add(time.Hour)
	for i := 0; i < len(teams) && i < 8; i += 2 {
		if i+1 < len(teams) {
			match := models.Match{
				TournamentID: bracket.TournamentID,
				Round:        models.Round1stRound,
				Team1:        teams[i],
				Team2:        teams[i+1],
				Status:       models.MatchStatusPending,
				ScheduledAt:  baseTime.Add(time.Duration(i/2) * 20 * time.Minute), // 卓球は20分間隔
			}
			rounds[0].Matches = append(rounds[0].Matches, match)
		}
	}
	
	// 後続のラウンドを生成
	if format == models.FormatRainy {
		s.generatePlaceholderMatches(&rounds[1], models.RoundQuarterfinal, bracket.TournamentID, 2, baseTime.Add(2*time.Hour))
		s.generatePlaceholderMatches(&rounds[2], models.RoundSemifinal, bracket.TournamentID, 2, baseTime.Add(3*time.Hour))
		s.generatePlaceholderMatches(&rounds[3], models.RoundLoserBracket, bracket.TournamentID, 1, baseTime.Add(4*time.Hour))
		s.generatePlaceholderMatches(&rounds[4], models.RoundThirdPlace, bracket.TournamentID, 1, baseTime.Add(5*time.Hour))
		s.generatePlaceholderMatches(&rounds[5], models.RoundFinal, bracket.TournamentID, 1, baseTime.Add(5*time.Hour+30*time.Minute))
	} else {
		s.generatePlaceholderMatches(&rounds[1], models.RoundQuarterfinal, bracket.TournamentID, 2, baseTime.Add(2*time.Hour))
		s.generatePlaceholderMatches(&rounds[2], models.RoundSemifinal, bracket.TournamentID, 2, baseTime.Add(3*time.Hour))
		s.generatePlaceholderMatches(&rounds[3], models.RoundThirdPlace, bracket.TournamentID, 1, baseTime.Add(4*time.Hour))
		s.generatePlaceholderMatches(&rounds[4], models.RoundFinal, bracket.TournamentID, 1, baseTime.Add(4*time.Hour+30*time.Minute))
	}
	
	bracket.Rounds = rounds
	return nil
}

// generateSoccerBracket は8人制サッカーのブラケットを生成する
func (s *tournamentServiceImpl) generateSoccerBracket(bracket *models.Bracket, teams []string) error {
	if len(teams) < 8 {
		return errors.New("8人制サッカーには最低8チームが必要です")
	}
	
	// 8チームトーナメント構造
	rounds := []models.Round{
		{Name: models.Round1stRound, Matches: []models.Match{}},
		{Name: models.RoundQuarterfinal, Matches: []models.Match{}},
		{Name: models.RoundSemifinal, Matches: []models.Match{}},
		{Name: models.RoundThirdPlace, Matches: []models.Match{}},
		{Name: models.RoundFinal, Matches: []models.Match{}},
	}
	
	// 1回戦の試合を生成
	baseTime := time.Now().Add(time.Hour)
	for i := 0; i < len(teams) && i < 8; i += 2 {
		if i+1 < len(teams) {
			match := models.Match{
				TournamentID: bracket.TournamentID,
				Round:        models.Round1stRound,
				Team1:        teams[i],
				Team2:        teams[i+1],
				Status:       models.MatchStatusPending,
				ScheduledAt:  baseTime.Add(time.Duration(i/2) * 45 * time.Minute), // サッカーは45分間隔
			}
			rounds[0].Matches = append(rounds[0].Matches, match)
		}
	}
	
	// 後続のラウンドを生成
	s.generatePlaceholderMatches(&rounds[1], models.RoundQuarterfinal, bracket.TournamentID, 2, baseTime.Add(3*time.Hour))
	s.generatePlaceholderMatches(&rounds[2], models.RoundSemifinal, bracket.TournamentID, 2, baseTime.Add(5*time.Hour))
	s.generatePlaceholderMatches(&rounds[3], models.RoundThirdPlace, bracket.TournamentID, 1, baseTime.Add(7*time.Hour))
	s.generatePlaceholderMatches(&rounds[4], models.RoundFinal, bracket.TournamentID, 1, baseTime.Add(7*time.Hour+45*time.Minute))
	
	bracket.Rounds = rounds
	return nil
}

// generatePlaceholderMatches はプレースホルダーの試合を生成する
func (s *tournamentServiceImpl) generatePlaceholderMatches(round *models.Round, roundName string, tournamentID int, matchCount int, baseTime time.Time) {
	for i := 0; i < matchCount; i++ {
		match := models.Match{
			TournamentID: tournamentID,
			Round:        roundName,
			Team1:        "TBD", // To Be Determined
			Team2:        "TBD",
			Status:       models.MatchStatusPending,
			ScheduledAt:  baseTime.Add(time.Duration(i) * 30 * time.Minute),
		}
		round.Matches = append(round.Matches, match)
	}
}
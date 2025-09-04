package service

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
)

// MatchService は試合関連のビジネスロジックを提供するインターフェース
type MatchService interface {
	// 試合作成と管理
	CreateMatch(match *models.Match) error
	GetMatch(id int) (*models.Match, error)
	UpdateMatch(match *models.Match) error
	DeleteMatch(id int) error
	
	// 試合結果処理
	UpdateMatchResult(matchID int, result models.MatchResult) error
	SubmitMatchResult(matchID int, score1, score2 int, winner string) error
	
	// トーナメント進行管理
	AdvanceWinner(matchID int) error
	ProcessTournamentAdvancement(tournamentID int, round string) error
	
	// 試合検索・フィルタリング
	GetMatchesBySport(sport string) ([]*models.Match, error)
	GetMatchesByTournament(tournamentID int) ([]*models.Match, error)
	GetMatchesByRound(tournamentID int, round string) ([]*models.Match, error)
	GetPendingMatches() ([]*models.Match, error)
	GetCompletedMatches() ([]*models.Match, error)
	
	// 試合検証とルール強制
	ValidateMatchResult(matchID int, result models.MatchResult) error
	ValidateMatchAdvancement(matchID int) error
	EnforceTournamentRules(tournamentID int) error
	
	// 統計・情報取得
	GetMatchStatistics(tournamentID int) (*MatchStatistics, error)
	GetNextMatches(tournamentID int) ([]*models.Match, error)
}

// MatchStatistics は試合統計情報を表す構造体
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

// TeamStats はチーム統計情報を表す構造体
type TeamStats struct {
	TeamName      string  `json:"team_name"`
	MatchesPlayed int     `json:"matches_played"`
	Wins          int     `json:"wins"`
	Losses        int     `json:"losses"`
	TotalScore    int     `json:"total_score"`
	AverageScore  float64 `json:"average_score"`
}

// matchServiceImpl はMatchServiceの実装
type matchServiceImpl struct {
	matchRepo      repository.MatchRepository
	tournamentRepo repository.TournamentRepository
}

// NewMatchService は新しいMatchServiceインスタンスを作成する
func NewMatchService(matchRepo repository.MatchRepository, tournamentRepo repository.TournamentRepository) MatchService {
	return &matchServiceImpl{
		matchRepo:      matchRepo,
		tournamentRepo: tournamentRepo,
	}
}

// CreateMatch は新しい試合を作成する
func (s *matchServiceImpl) CreateMatch(match *models.Match) error {
	if match == nil {
		return errors.New("試合は必須です")
	}
	
	// トーナメントの存在確認
	_, err := s.tournamentRepo.GetByID(match.TournamentID)
	if err != nil {
		return errors.New("指定されたトーナメントが見つかりません")
	}
	
	// 試合の検証
	if err := match.Validate(); err != nil {
		return fmt.Errorf("試合の検証に失敗しました: %v", err)
	}
	
	// 同じラウンドで同じチーム同士の試合が既に存在しないかチェック
	existingMatches, err := s.matchRepo.GetByTournamentAndRound(match.TournamentID, match.Round)
	if err == nil {
		for _, existingMatch := range existingMatches {
			if (existingMatch.Team1 == match.Team1 && existingMatch.Team2 == match.Team2) ||
			   (existingMatch.Team1 == match.Team2 && existingMatch.Team2 == match.Team1) {
				return errors.New("同じラウンドで同じチーム同士の試合が既に存在します")
			}
		}
	}
	
	err = s.matchRepo.Create(match)
	if err != nil {
		log.Printf("試合作成エラー: %v", err)
		return errors.New("試合の作成に失敗しました")
	}
	
	log.Printf("試合を作成しました: ID=%d, %s vs %s", match.ID, match.Team1, match.Team2)
	return nil
}

// GetMatch はIDで試合を取得する
func (s *matchServiceImpl) GetMatch(id int) (*models.Match, error) {
	if id <= 0 {
		return nil, errors.New("無効な試合IDです")
	}
	
	match, err := s.matchRepo.GetByID(id)
	if err != nil {
		log.Printf("試合取得エラー: %v", err)
		return nil, errors.New("試合が見つかりません")
	}
	
	return match, nil
}

// UpdateMatch は試合を更新する
func (s *matchServiceImpl) UpdateMatch(match *models.Match) error {
	if match == nil {
		return errors.New("試合は必須です")
	}
	
	if match.ID <= 0 {
		return errors.New("無効な試合IDです")
	}
	
	// 既存の試合が存在するかチェック
	_, err := s.matchRepo.GetByID(match.ID)
	if err != nil {
		return errors.New("更新対象の試合が見つかりません")
	}
	
	err = s.matchRepo.Update(match)
	if err != nil {
		log.Printf("試合更新エラー: %v", err)
		return errors.New("試合の更新に失敗しました")
	}
	
	log.Printf("試合を更新しました: ID=%d", match.ID)
	return nil
}

// DeleteMatch は試合を削除する
func (s *matchServiceImpl) DeleteMatch(id int) error {
	if id <= 0 {
		return errors.New("無効な試合IDです")
	}
	
	// 試合が完了している場合は削除を禁止
	match, err := s.matchRepo.GetByID(id)
	if err != nil {
		return errors.New("削除対象の試合が見つかりません")
	}
	
	if match.IsCompleted() {
		return errors.New("完了した試合は削除できません")
	}
	
	err = s.matchRepo.Delete(id)
	if err != nil {
		log.Printf("試合削除エラー: %v", err)
		return errors.New("試合の削除に失敗しました")
	}
	
	log.Printf("試合を削除しました: ID=%d", id)
	return nil
}

// UpdateMatchResult は試合結果を更新する
func (s *matchServiceImpl) UpdateMatchResult(matchID int, result models.MatchResult) error {
	if matchID <= 0 {
		return errors.New("無効な試合IDです")
	}
	
	// 試合結果の検証
	if err := s.ValidateMatchResult(matchID, result); err != nil {
		return err
	}
	
	err := s.matchRepo.UpdateResult(matchID, result)
	if err != nil {
		log.Printf("試合結果更新エラー: %v", err)
		return errors.New("試合結果の更新に失敗しました")
	}
	
	// 勝者を次のラウンドに進出させる
	err = s.AdvanceWinner(matchID)
	if err != nil {
		log.Printf("勝者進出処理エラー: %v", err)
		// 進出処理のエラーは警告として扱い、試合結果更新は成功とする
	}
	
	log.Printf("試合結果を更新しました: ID=%d, Winner=%s", matchID, result.Winner)
	return nil
}

// SubmitMatchResult は試合結果を提出する（簡易版）
func (s *matchServiceImpl) SubmitMatchResult(matchID int, score1, score2 int, winner string) error {
	result := models.MatchResult{
		Score1: score1,
		Score2: score2,
		Winner: winner,
	}
	
	return s.UpdateMatchResult(matchID, result)
}

// AdvanceWinner は勝者を次のラウンドに進出させる
func (s *matchServiceImpl) AdvanceWinner(matchID int) error {
	if matchID <= 0 {
		return errors.New("無効な試合IDです")
	}
	
	// 試合を取得
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return errors.New("試合が見つかりません")
	}
	
	// 試合が完了していない場合はエラー
	if !match.IsCompleted() || match.Winner == nil {
		return errors.New("試合が完了していないため、勝者を進出させることができません")
	}
	
	// 次のラウンドを決定
	nextRound := s.getNextRound(match.Round)
	if nextRound == "" {
		// 決勝戦の場合は進出先がない
		log.Printf("試合 %d は最終ラウンドのため、進出処理をスキップします", matchID)
		return nil
	}
	
	// 次のラウンドの試合を検索して勝者を設定
	nextMatches, err := s.matchRepo.GetByTournamentAndRound(match.TournamentID, nextRound)
	if err != nil {
		return fmt.Errorf("次のラウンドの試合取得に失敗しました: %v", err)
	}
	
	// 適切な次の試合を見つけて勝者を設定
	for _, nextMatch := range nextMatches {
		if nextMatch.Team1 == "TBD" || nextMatch.Team2 == "TBD" {
			// TBDを勝者で置き換え
			if nextMatch.Team1 == "TBD" {
				nextMatch.Team1 = *match.Winner
			} else {
				nextMatch.Team2 = *match.Winner
			}
			
			err = s.matchRepo.Update(nextMatch)
			if err != nil {
				return fmt.Errorf("次のラウンドの試合更新に失敗しました: %v", err)
			}
			
			log.Printf("勝者 %s を次のラウンド %s に進出させました", *match.Winner, nextRound)
			break
		}
	}
	
	return nil
}

// ProcessTournamentAdvancement はトーナメント全体の進出処理を行う
func (s *matchServiceImpl) ProcessTournamentAdvancement(tournamentID int, round string) error {
	if tournamentID <= 0 {
		return errors.New("無効なトーナメントIDです")
	}
	
	if !models.IsValidRound(round) {
		return errors.New("無効なラウンドです")
	}
	
	// 指定されたラウンドの全ての試合を取得
	matches, err := s.matchRepo.GetByTournamentAndRound(tournamentID, round)
	if err != nil {
		return fmt.Errorf("ラウンド %s の試合取得に失敗しました: %v", round, err)
	}
	
	// 全ての試合が完了しているかチェック
	for _, match := range matches {
		if !match.IsCompleted() {
			return fmt.Errorf("ラウンド %s にはまだ完了していない試合があります", round)
		}
	}
	
	// 各試合の勝者を次のラウンドに進出させる
	for _, match := range matches {
		err = s.AdvanceWinner(match.ID)
		if err != nil {
			log.Printf("試合 %d の勝者進出処理エラー: %v", match.ID, err)
		}
	}
	
	log.Printf("トーナメント %d のラウンド %s の進出処理を完了しました", tournamentID, round)
	return nil
}

// GetMatchesBySport はスポーツで試合を取得する
func (s *matchServiceImpl) GetMatchesBySport(sport string) ([]*models.Match, error) {
	if !models.IsValidSport(sport) {
		return nil, errors.New("無効なスポーツです")
	}
	
	matches, err := s.matchRepo.GetBySport(sport)
	if err != nil {
		log.Printf("スポーツ別試合取得エラー: %v", err)
		return nil, errors.New("試合の取得に失敗しました")
	}
	
	return matches, nil
}

// GetMatchesByTournament はトーナメントで試合を取得する
func (s *matchServiceImpl) GetMatchesByTournament(tournamentID int) ([]*models.Match, error) {
	if tournamentID <= 0 {
		return nil, errors.New("無効なトーナメントIDです")
	}
	
	matches, err := s.matchRepo.GetByTournament(tournamentID)
	if err != nil {
		log.Printf("トーナメント別試合取得エラー: %v", err)
		return nil, errors.New("試合の取得に失敗しました")
	}
	
	return matches, nil
}

// GetMatchesByRound はトーナメントとラウンドで試合を取得する
func (s *matchServiceImpl) GetMatchesByRound(tournamentID int, round string) ([]*models.Match, error) {
	if tournamentID <= 0 {
		return nil, errors.New("無効なトーナメントIDです")
	}
	
	if !models.IsValidRound(round) {
		return nil, errors.New("無効なラウンドです")
	}
	
	matches, err := s.matchRepo.GetByTournamentAndRound(tournamentID, round)
	if err != nil {
		log.Printf("ラウンド別試合取得エラー: %v", err)
		return nil, errors.New("試合の取得に失敗しました")
	}
	
	return matches, nil
}

// GetPendingMatches は未実施の試合を取得する
func (s *matchServiceImpl) GetPendingMatches() ([]*models.Match, error) {
	matches, err := s.matchRepo.GetPendingMatches()
	if err != nil {
		log.Printf("未実施試合取得エラー: %v", err)
		return nil, errors.New("未実施試合の取得に失敗しました")
	}
	
	return matches, nil
}

// GetCompletedMatches は完了した試合を取得する
func (s *matchServiceImpl) GetCompletedMatches() ([]*models.Match, error) {
	matches, err := s.matchRepo.GetCompletedMatches()
	if err != nil {
		log.Printf("完了試合取得エラー: %v", err)
		return nil, errors.New("完了試合の取得に失敗しました")
	}
	
	return matches, nil
}

// ValidateMatchResult は試合結果を検証する
func (s *matchServiceImpl) ValidateMatchResult(matchID int, result models.MatchResult) error {
	// 基本的な結果検証
	if err := result.Validate(); err != nil {
		return fmt.Errorf("試合結果の検証に失敗しました: %v", err)
	}
	
	// 試合を取得
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return errors.New("試合が見つかりません")
	}
	
	// 既に完了している試合の結果は更新できない
	if match.IsCompleted() {
		return errors.New("既に完了している試合の結果は更新できません")
	}
	
	// 勝者がチーム1またはチーム2のいずれかであることを確認
	if result.Winner != match.Team1 && result.Winner != match.Team2 {
		return errors.New("勝者は参加チームのいずれかである必要があります")
	}
	
	// スコアと勝者の整合性をチェック
	if result.Score1 > result.Score2 && result.Winner != match.Team1 {
		return errors.New("スコアと勝者が一致しません")
	}
	
	if result.Score2 > result.Score1 && result.Winner != match.Team2 {
		return errors.New("スコアと勝者が一致しません")
	}
	
	if result.Score1 == result.Score2 {
		return errors.New("引き分けは許可されていません")
	}
	
	return nil
}

// ValidateMatchAdvancement は試合進出の検証を行う
func (s *matchServiceImpl) ValidateMatchAdvancement(matchID int) error {
	match, err := s.matchRepo.GetByID(matchID)
	if err != nil {
		return errors.New("試合が見つかりません")
	}
	
	if !match.IsCompleted() {
		return errors.New("試合が完了していません")
	}
	
	if match.Winner == nil || strings.TrimSpace(*match.Winner) == "" {
		return errors.New("勝者が決定していません")
	}
	
	return nil
}

// EnforceTournamentRules はトーナメントルールを強制する
func (s *matchServiceImpl) EnforceTournamentRules(tournamentID int) error {
	if tournamentID <= 0 {
		return errors.New("無効なトーナメントIDです")
	}
	
	// トーナメントを取得
	tournament, err := s.tournamentRepo.GetByID(tournamentID)
	if err != nil {
		return errors.New("トーナメントが見つかりません")
	}
	
	// 全ての試合を取得
	matches, err := s.matchRepo.GetByTournament(tournamentID)
	if err != nil {
		return errors.New("試合の取得に失敗しました")
	}
	
	// スポーツ固有のルールを適用
	switch tournament.Sport {
	case models.SportVolleyball:
		return s.enforceVolleyballRules(matches)
	case models.SportTableTennis:
		return s.enforceTableTennisRules(matches, tournament.Format)
	case models.SportSoccer:
		return s.enforceSoccerRules(matches)
	default:
		return fmt.Errorf("サポートされていないスポーツです: %s", tournament.Sport)
	}
}

// GetMatchStatistics は試合統計情報を取得する
func (s *matchServiceImpl) GetMatchStatistics(tournamentID int) (*MatchStatistics, error) {
	if tournamentID <= 0 {
		return nil, errors.New("無効なトーナメントIDです")
	}
	
	matches, err := s.matchRepo.GetByTournament(tournamentID)
	if err != nil {
		return nil, errors.New("試合の取得に失敗しました")
	}
	
	stats := &MatchStatistics{
		TournamentID:   tournamentID,
		TotalMatches:   len(matches),
		MatchesByRound: make(map[string]int),
		AverageScore:   make(map[string]float64),
		TeamStats:      make(map[string]*TeamStats),
	}
	
	completedCount := 0
	totalScore1 := 0
	totalScore2 := 0
	scoreCount := 0
	
	for _, match := range matches {
		// ラウンド別統計
		stats.MatchesByRound[match.Round]++
		
		if match.IsCompleted() {
			completedCount++
			
			// スコア統計
			if match.Score1 != nil && match.Score2 != nil {
				totalScore1 += *match.Score1
				totalScore2 += *match.Score2
				scoreCount++
			}
			
			// チーム統計
			s.updateTeamStats(stats.TeamStats, match.Team1, match)
			s.updateTeamStats(stats.TeamStats, match.Team2, match)
		}
	}
	
	stats.CompletedMatches = completedCount
	stats.PendingMatches = stats.TotalMatches - completedCount
	
	if stats.TotalMatches > 0 {
		stats.CompletionRate = float64(completedCount) / float64(stats.TotalMatches) * 100
	}
	
	if scoreCount > 0 {
		stats.AverageScore["team1"] = float64(totalScore1) / float64(scoreCount)
		stats.AverageScore["team2"] = float64(totalScore2) / float64(scoreCount)
	}
	
	return stats, nil
}

// GetNextMatches は次に実施予定の試合を取得する
func (s *matchServiceImpl) GetNextMatches(tournamentID int) ([]*models.Match, error) {
	if tournamentID <= 0 {
		return nil, errors.New("無効なトーナメントIDです")
	}
	
	matches, err := s.matchRepo.GetByTournament(tournamentID)
	if err != nil {
		return nil, errors.New("試合の取得に失敗しました")
	}
	
	var nextMatches []*models.Match
	now := time.Now()
	
	for _, match := range matches {
		if match.IsPending() && match.ScheduledAt.After(now) {
			// チームが決定している試合のみ
			if match.Team1 != "TBD" && match.Team2 != "TBD" {
				nextMatches = append(nextMatches, match)
			}
		}
	}
	
	return nextMatches, nil
}

// getNextRound は現在のラウンドから次のラウンドを決定する
func (s *matchServiceImpl) getNextRound(currentRound string) string {
	switch currentRound {
	case models.Round1stRound:
		return models.RoundQuarterfinal
	case models.RoundQuarterfinal:
		return models.RoundSemifinal
	case models.RoundSemifinal:
		return models.RoundFinal
	case models.RoundLoserBracket:
		return models.RoundThirdPlace
	case models.RoundThirdPlace:
		return "" // 3位決定戦の後は進出先なし
	case models.RoundFinal:
		return "" // 決勝戦の後は進出先なし
	default:
		return ""
	}
}

// enforceVolleyballRules はバレーボール固有のルールを強制する
func (s *matchServiceImpl) enforceVolleyballRules(matches []*models.Match) error {
	// バレーボール固有のルール検証
	for _, match := range matches {
		if match.IsCompleted() && match.Score1 != nil && match.Score2 != nil {
			// バレーボールのスコア範囲チェック（例：0-25の範囲）
			if *match.Score1 < 0 || *match.Score1 > 25 || *match.Score2 < 0 || *match.Score2 > 25 {
				return fmt.Errorf("試合 %d のスコアがバレーボールの有効範囲を超えています", match.ID)
			}
		}
	}
	
	log.Printf("バレーボールルールの検証が完了しました")
	return nil
}

// enforceTableTennisRules は卓球固有のルールを強制する
func (s *matchServiceImpl) enforceTableTennisRules(matches []*models.Match, format string) error {
	// 卓球固有のルール検証
	for _, match := range matches {
		if match.IsCompleted() && match.Score1 != nil && match.Score2 != nil {
			// 卓球のスコア範囲チェック（例：0-11の範囲）
			if *match.Score1 < 0 || *match.Score1 > 11 || *match.Score2 < 0 || *match.Score2 > 11 {
				return fmt.Errorf("試合 %d のスコアが卓球の有効範囲を超えています", match.ID)
			}
		}
	}
	
	// 雨天フォーマット固有のルール
	if format == models.FormatRainy {
		// 敗者復活戦の存在確認
		hasLoserBracket := false
		for _, match := range matches {
			if match.Round == models.RoundLoserBracket {
				hasLoserBracket = true
				break
			}
		}
		
		if !hasLoserBracket {
			log.Printf("警告: 雨天フォーマットですが敗者復活戦が見つかりません")
		}
	}
	
	log.Printf("卓球ルールの検証が完了しました（フォーマット: %s）", format)
	return nil
}

// enforceSoccerRules はサッカー固有のルールを強制する
func (s *matchServiceImpl) enforceSoccerRules(matches []*models.Match) error {
	// サッカー固有のルール検証
	for _, match := range matches {
		if match.IsCompleted() && match.Score1 != nil && match.Score2 != nil {
			// サッカーのスコア範囲チェック（例：0-20の範囲）
			if *match.Score1 < 0 || *match.Score1 > 20 || *match.Score2 < 0 || *match.Score2 > 20 {
				return fmt.Errorf("試合 %d のスコアがサッカーの有効範囲を超えています", match.ID)
			}
		}
	}
	
	log.Printf("サッカールールの検証が完了しました")
	return nil
}

// updateTeamStats はチーム統計を更新する
func (s *matchServiceImpl) updateTeamStats(teamStats map[string]*TeamStats, teamName string, match *models.Match) {
	if teamStats[teamName] == nil {
		teamStats[teamName] = &TeamStats{
			TeamName: teamName,
		}
	}
	
	stats := teamStats[teamName]
	stats.MatchesPlayed++
	
	if match.Winner != nil && *match.Winner == teamName {
		stats.Wins++
	} else {
		stats.Losses++
	}
	
	// スコアを追加
	if match.Score1 != nil && match.Score2 != nil {
		if match.Team1 == teamName {
			stats.TotalScore += *match.Score1
		} else {
			stats.TotalScore += *match.Score2
		}
		
		if stats.MatchesPlayed > 0 {
			stats.AverageScore = float64(stats.TotalScore) / float64(stats.MatchesPlayed)
		}
	}
}
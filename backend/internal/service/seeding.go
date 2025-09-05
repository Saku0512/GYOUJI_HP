package service

import (
	"errors"
	"fmt"
	"log"
	"time"

	"backend/internal/models"
	"backend/internal/repository"
)

// SeedingService はトーナメントシーディング関連のビジネスロジックを提供するインターフェース
type SeedingService interface {
	// トーナメント初期化
	InitializeAllTournaments() error
	InitializeTournamentBySport(sport string) error
	
	// スポーツ別初期化
	InitializeVolleyballTournament() error
	InitializeTableTennisTournament(format string) error
	InitializeSoccerTournament() error
	
	// データリセット
	ResetTournamentData(sport string) error
	ResetAllTournamentData() error
}

// seedingServiceImpl はSeedingServiceの実装
type seedingServiceImpl struct {
	tournamentRepo repository.TournamentRepository
	matchRepo      repository.MatchRepository
	tournamentSvc  TournamentService
}

// NewSeedingService は新しいSeedingServiceインスタンスを作成する
func NewSeedingService(
	tournamentRepo repository.TournamentRepository,
	matchRepo repository.MatchRepository,
	tournamentSvc TournamentService,
) SeedingService {
	return &seedingServiceImpl{
		tournamentRepo: tournamentRepo,
		matchRepo:      matchRepo,
		tournamentSvc:  tournamentSvc,
	}
}

// InitializeAllTournaments は全てのスポーツのトーナメントを初期化する
func (s *seedingServiceImpl) InitializeAllTournaments() error {
	log.Println("全トーナメントの初期化を開始します")
	
	// バレーボールトーナメントの初期化
	if err := s.InitializeVolleyballTournament(); err != nil {
		log.Printf("バレーボールトーナメント初期化エラー: %v", err)
		return fmt.Errorf("バレーボールトーナメントの初期化に失敗しました: %v", err)
	}
	
	// 卓球トーナメントの初期化（標準フォーマット）
	if err := s.InitializeTableTennisTournament(models.FormatStandard); err != nil {
		log.Printf("卓球トーナメント初期化エラー: %v", err)
		return fmt.Errorf("卓球トーナメントの初期化に失敗しました: %v", err)
	}
	
	// サッカートーナメントの初期化
	if err := s.InitializeSoccerTournament(); err != nil {
		log.Printf("サッカートーナメント初期化エラー: %v", err)
		return fmt.Errorf("サッカートーナメントの初期化に失敗しました: %v", err)
	}
	
	log.Println("全トーナメントの初期化が完了しました")
	return nil
}

// InitializeTournamentBySport は指定されたスポーツのトーナメントを初期化する
func (s *seedingServiceImpl) InitializeTournamentBySport(sport string) error {
	if !models.IsValidSport(sport) {
		return errors.New("無効なスポーツです")
	}
	
	switch sport {
	case models.SportVolleyball:
		return s.InitializeVolleyballTournament()
	case models.SportTableTennis:
		return s.InitializeTableTennisTournament(models.FormatStandard)
	case models.SportSoccer:
		return s.InitializeSoccerTournament()
	default:
		return fmt.Errorf("サポートされていないスポーツです: %s", sport)
	}
}

// InitializeVolleyballTournament はバレーボールトーナメントを初期化する
func (s *seedingServiceImpl) InitializeVolleyballTournament() error {
	log.Println("バレーボールトーナメントの初期化を開始します")
	
	// 既存のトーナメントをリセット
	if err := s.ResetTournamentData(models.SportVolleyball); err != nil {
		log.Printf("バレーボールトーナメントリセットエラー: %v", err)
	}
	
	// トーナメントを作成
	tournament, err := s.tournamentSvc.CreateTournament(models.SportVolleyball, models.FormatStandard)
	if err != nil {
		return fmt.Errorf("バレーボールトーナメント作成エラー: %v", err)
	}
	
	// READMEに基づく1回戦の試合データ
	firstRoundMatches := []VolleyballMatch{
		{Team1: "専・教", Team2: "IE4", Time: "09:30"},
		{Team1: "IS5", Team2: "IT4", Time: "09:30"},
		{Team1: "IT3", Team2: "IT2", Time: "10:00"},
		{Team1: "1-1", Team2: "IE2", Time: "10:00"},
		{Team1: "IS3", Team2: "IS2", Time: "10:30"},
		{Team1: "IS4", Team2: "IE5", Time: "10:30"},
		{Team1: "1-2", Team2: "1-3", Time: "11:00"},
		{Team1: "IE3", Team2: "IT5", Time: "11:00"},
	}
	
	// 1回戦の試合を作成
	baseDate := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour) // 明日の日付
	for i, matchData := range firstRoundMatches {
		scheduledTime, err := s.parseTimeSlot(baseDate, matchData.Time)
		if err != nil {
			return fmt.Errorf("時間解析エラー: %v", err)
		}
		
		match := &models.Match{
			TournamentID: tournament.ID,
			Round:        models.Round1stRound,
			Team1:        matchData.Team1,
			Team2:        matchData.Team2,
			Status:       models.MatchStatusPending,
			ScheduledAt:  scheduledTime,
		}
		
		if err := s.matchRepo.Create(match); err != nil {
			return fmt.Errorf("1回戦試合%d作成エラー: %v", i+1, err)
		}
	}
	
	// 準々決勝以降のプレースホルダー試合を作成
	if err := s.createVolleyballPlaceholderMatches(tournament.ID, baseDate); err != nil {
		return fmt.Errorf("プレースホルダー試合作成エラー: %v", err)
	}
	
	log.Printf("バレーボールトーナメントの初期化が完了しました: ID=%d", tournament.ID)
	return nil
}

// InitializeTableTennisTournament は卓球トーナメントを初期化する
func (s *seedingServiceImpl) InitializeTableTennisTournament(format string) error {
	log.Printf("卓球トーナメントの初期化を開始します: Format=%s", format)
	
	// 既存のトーナメントをリセット
	if err := s.ResetTournamentData(models.SportTableTennis); err != nil {
		log.Printf("卓球トーナメントリセットエラー: %v", err)
	}
	
	// トーナメントを作成
	tournament, err := s.tournamentSvc.CreateTournament(models.SportTableTennis, format)
	if err != nil {
		return fmt.Errorf("卓球トーナメント作成エラー: %v", err)
	}
	
	// フォーマットに応じた1回戦の試合データ
	var firstRoundMatches []TableTennisMatch
	
	if format == models.FormatRainy {
		// 雨天時の試合データ
		firstRoundMatches = []TableTennisMatch{
			{Team1: "1-2", Team2: "IE5", Time: "09:30"},
			{Team1: "IS3", Team2: "IT2", Time: "09:45"},
			{Team1: "1-3", Team2: "IE4", Time: "10:00"},
			{Team1: "IT3", Team2: "IT4", Time: "10:15"},
			{Team1: "IS5", Team2: "IE2", Time: "10:30"},
			{Team1: "1-1", Team2: "IS3", Time: "10:45"},
			{Team1: "IS2", Team2: "IS4", Time: "11:00"},
			{Team1: "IT5", Team2: "専・教", Time: "11:15"},
		}
	} else {
		// 晴天時の試合データ
		firstRoundMatches = []TableTennisMatch{
			{Team1: "1-2", Team2: "IE5", Time: "09:30"},
			{Team1: "IE3", Team2: "IT2", Time: "09:50"},
			{Team1: "1-3", Team2: "IE4", Time: "10:10"},
			{Team1: "IT3", Team2: "IT4", Time: "10:30"},
			{Team1: "IS5", Team2: "IE2", Time: "10:50"},
			{Team1: "1-1", Team2: "IS3", Time: "11:10"},
			{Team1: "IS2", Team2: "IS4", Time: "11:30"},
			{Team1: "IT5", Team2: "専・教", Time: "11:50"},
		}
	}
	
	// 1回戦の試合を作成
	baseDate := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour) // 明日の日付
	for i, matchData := range firstRoundMatches {
		scheduledTime, err := s.parseTimeSlot(baseDate, matchData.Time)
		if err != nil {
			return fmt.Errorf("時間解析エラー: %v", err)
		}
		
		match := &models.Match{
			TournamentID: tournament.ID,
			Round:        models.Round1stRound,
			Team1:        matchData.Team1,
			Team2:        matchData.Team2,
			Status:       models.MatchStatusPending,
			ScheduledAt:  scheduledTime,
		}
		
		if err := s.matchRepo.Create(match); err != nil {
			return fmt.Errorf("1回戦試合%d作成エラー: %v", i+1, err)
		}
	}
	
	// 準々決勝以降のプレースホルダー試合を作成
	if err := s.createTableTennisPlaceholderMatches(tournament.ID, baseDate, format); err != nil {
		return fmt.Errorf("プレースホルダー試合作成エラー: %v", err)
	}
	
	log.Printf("卓球トーナメントの初期化が完了しました: ID=%d, Format=%s", tournament.ID, format)
	return nil
}

// InitializeSoccerTournament はサッカートーナメントを初期化する
func (s *seedingServiceImpl) InitializeSoccerTournament() error {
	log.Println("サッカートーナメントの初期化を開始します")
	
	// 既存のトーナメントをリセット
	if err := s.ResetTournamentData(models.SportSoccer); err != nil {
		log.Printf("サッカートーナメントリセットエラー: %v", err)
	}
	
	// トーナメントを作成
	tournament, err := s.tournamentSvc.CreateTournament(models.SportSoccer, models.FormatStandard)
	if err != nil {
		return fmt.Errorf("サッカートーナメント作成エラー: %v", err)
	}
	
	// READMEに基づく1回戦の試合データ
	firstRoundMatches := []SoccerMatch{
		{Team1: "IS3", Team2: "IE2", Time: "09:30"},
		{Team1: "1-1", Team2: "IS2", Time: "09:45"},
		{Team1: "IS4", Team2: "IT5", Time: "10:00"},
		{Team1: "IS5", Team2: "専・教", Time: "10:15"},
		{Team1: "1-2", Team2: "1-3", Time: "10:30"},
		{Team1: "IE3", Team2: "IT4", Time: "10:45"},
		{Team1: "IT3", Team2: "IE4", Time: "11:00"},
		{Team1: "IT2", Team2: "IE5", Time: "11:15"},
	}
	
	// 1回戦の試合を作成
	baseDate := time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour) // 明日の日付
	for i, matchData := range firstRoundMatches {
		scheduledTime, err := s.parseTimeSlot(baseDate, matchData.Time)
		if err != nil {
			return fmt.Errorf("時間解析エラー: %v", err)
		}
		
		match := &models.Match{
			TournamentID: tournament.ID,
			Round:        models.Round1stRound,
			Team1:        matchData.Team1,
			Team2:        matchData.Team2,
			Status:       models.MatchStatusPending,
			ScheduledAt:  scheduledTime,
		}
		
		if err := s.matchRepo.Create(match); err != nil {
			return fmt.Errorf("1回戦試合%d作成エラー: %v", i+1, err)
		}
	}
	
	// 準々決勝以降のプレースホルダー試合を作成
	if err := s.createSoccerPlaceholderMatches(tournament.ID, baseDate); err != nil {
		return fmt.Errorf("プレースホルダー試合作成エラー: %v", err)
	}
	
	log.Printf("サッカートーナメントの初期化が完了しました: ID=%d", tournament.ID)
	return nil
}

// ResetTournamentData は指定されたスポーツのトーナメントデータをリセットする
func (s *seedingServiceImpl) ResetTournamentData(sport string) error {
	if !models.IsValidSport(sport) {
		return errors.New("無効なスポーツです")
	}
	
	// 既存のトーナメントを取得
	tournament, err := s.tournamentRepo.GetBySport(sport)
	if err != nil {
		// トーナメントが存在しない場合は何もしない
		return nil
	}
	
	// 関連する試合を削除
	if err := s.matchRepo.DeleteByTournament(tournament.ID); err != nil {
		log.Printf("試合削除エラー: %v", err)
		return fmt.Errorf("試合データの削除に失敗しました: %v", err)
	}
	
	// トーナメントを削除
	if err := s.tournamentRepo.Delete(tournament.ID); err != nil {
		log.Printf("トーナメント削除エラー: %v", err)
		return fmt.Errorf("トーナメントの削除に失敗しました: %v", err)
	}
	
	// トーナメントサービスからも削除（テスト用）
	if err := s.tournamentSvc.DeleteTournament(tournament.ID); err != nil {
		log.Printf("トーナメントサービス削除エラー: %v", err)
		// エラーは無視（実装によっては存在しない場合がある）
	}
	
	log.Printf("トーナメントデータをリセットしました: Sport=%s", sport)
	return nil
}

// ResetAllTournamentData は全てのトーナメントデータをリセットする
func (s *seedingServiceImpl) ResetAllTournamentData() error {
	log.Println("全トーナメントデータのリセットを開始します")
	
	sports := []string{
		models.SportVolleyball,
		models.SportTableTennis,
		models.SportSoccer,
	}
	
	for _, sport := range sports {
		if err := s.ResetTournamentData(sport); err != nil {
			log.Printf("スポーツ %s のリセットエラー: %v", sport, err)
			return fmt.Errorf("スポーツ %s のリセットに失敗しました: %v", sport, err)
		}
	}
	
	log.Println("全トーナメントデータのリセットが完了しました")
	return nil
}

// parseTimeSlot は時間文字列（例: "09:30"）をtime.Timeに変換する
func (s *seedingServiceImpl) parseTimeSlot(baseDate time.Time, timeSlot string) (time.Time, error) {
	parsedTime, err := time.Parse("15:04", timeSlot)
	if err != nil {
		return time.Time{}, fmt.Errorf("時間解析エラー: %v", err)
	}
	
	// ベース日付に時間を設定
	result := time.Date(
		baseDate.Year(),
		baseDate.Month(),
		baseDate.Day(),
		parsedTime.Hour(),
		parsedTime.Minute(),
		0,
		0,
		baseDate.Location(),
	)
	
	return result, nil
}

// createVolleyballPlaceholderMatches はバレーボールのプレースホルダー試合を作成する
func (s *seedingServiceImpl) createVolleyballPlaceholderMatches(tournamentID int, baseDate time.Time) error {
	// 準々決勝（11:30, 13:00）
	quarterTimes := []string{"11:30", "13:00"}
	for i, timeSlot := range quarterTimes {
		scheduledTime, err := s.parseTimeSlot(baseDate, timeSlot)
		if err != nil {
			return err
		}
		
		match := &models.Match{
			TournamentID: tournamentID,
			Round:        models.RoundQuarterfinal,
			Team1:        "TBD",
			Team2:        "TBD",
			Status:       models.MatchStatusPending,
			ScheduledAt:  scheduledTime,
		}
		
		if err := s.matchRepo.Create(match); err != nil {
			return fmt.Errorf("準々決勝試合%d作成エラー: %v", i+1, err)
		}
	}
	
	// 準決勝（14:00, 14:30）
	semiTimes := []string{"14:00", "14:30"}
	for i, timeSlot := range semiTimes {
		scheduledTime, err := s.parseTimeSlot(baseDate, timeSlot)
		if err != nil {
			return err
		}
		
		match := &models.Match{
			TournamentID: tournamentID,
			Round:        models.RoundSemifinal,
			Team1:        "TBD",
			Team2:        "TBD",
			Status:       models.MatchStatusPending,
			ScheduledAt:  scheduledTime,
		}
		
		if err := s.matchRepo.Create(match); err != nil {
			return fmt.Errorf("準決勝試合%d作成エラー: %v", i+1, err)
		}
	}
	
	// 3位決定戦（15:00）
	thirdPlaceTime, err := s.parseTimeSlot(baseDate, "15:00")
	if err != nil {
		return err
	}
	
	thirdPlaceMatch := &models.Match{
		TournamentID: tournamentID,
		Round:        models.RoundThirdPlace,
		Team1:        "TBD",
		Team2:        "TBD",
		Status:       models.MatchStatusPending,
		ScheduledAt:  thirdPlaceTime,
	}
	
	if err := s.matchRepo.Create(thirdPlaceMatch); err != nil {
		return fmt.Errorf("3位決定戦作成エラー: %v", err)
	}
	
	// 決勝（15:30）
	finalTime, err := s.parseTimeSlot(baseDate, "15:30")
	if err != nil {
		return err
	}
	
	finalMatch := &models.Match{
		TournamentID: tournamentID,
		Round:        models.RoundFinal,
		Team1:        "TBD",
		Team2:        "TBD",
		Status:       models.MatchStatusPending,
		ScheduledAt:  finalTime,
	}
	
	if err := s.matchRepo.Create(finalMatch); err != nil {
		return fmt.Errorf("決勝作成エラー: %v", err)
	}
	
	return nil
}

// createTableTennisPlaceholderMatches は卓球のプレースホルダー試合を作成する
func (s *seedingServiceImpl) createTableTennisPlaceholderMatches(tournamentID int, baseDate time.Time, format string) error {
	if format == models.FormatRainy {
		// 雨天時フォーマット
		// 準々決勝（11:30, 11:45, 13:00, 13:15）
		quarterTimes := []string{"11:30", "11:45", "13:00", "13:15"}
		for i, timeSlot := range quarterTimes {
			scheduledTime, err := s.parseTimeSlot(baseDate, timeSlot)
			if err != nil {
				return err
			}
			
			match := &models.Match{
				TournamentID: tournamentID,
				Round:        models.RoundQuarterfinal,
				Team1:        "TBD",
				Team2:        "TBD",
				Status:       models.MatchStatusPending,
				ScheduledAt:  scheduledTime,
			}
			
			if err := s.matchRepo.Create(match); err != nil {
				return fmt.Errorf("準々決勝試合%d作成エラー: %v", i+1, err)
			}
		}
		
		// 敗者戦（13:30, 13:45, 14:00, 14:15）
		loserTimes := []string{"13:30", "13:45", "14:00", "14:15"}
		for i, timeSlot := range loserTimes {
			scheduledTime, err := s.parseTimeSlot(baseDate, timeSlot)
			if err != nil {
				return err
			}
			
			match := &models.Match{
				TournamentID: tournamentID,
				Round:        models.RoundLoserBracket,
				Team1:        "TBD",
				Team2:        "TBD",
				Status:       models.MatchStatusPending,
				ScheduledAt:  scheduledTime,
			}
			
			if err := s.matchRepo.Create(match); err != nil {
				return fmt.Errorf("敗者戦試合%d作成エラー: %v", i+1, err)
			}
		}
		
		// 準決勝（14:30, 14:45）
		semiTimes := []string{"14:30", "14:45"}
		for i, timeSlot := range semiTimes {
			scheduledTime, err := s.parseTimeSlot(baseDate, timeSlot)
			if err != nil {
				return err
			}
			
			match := &models.Match{
				TournamentID: tournamentID,
				Round:        models.RoundSemifinal,
				Team1:        "TBD",
				Team2:        "TBD",
				Status:       models.MatchStatusPending,
				ScheduledAt:  scheduledTime,
			}
			
			if err := s.matchRepo.Create(match); err != nil {
				return fmt.Errorf("準決勝試合%d作成エラー: %v", i+1, err)
			}
		}
		
		// 3位決定戦（15:00）
		thirdPlaceTime, err := s.parseTimeSlot(baseDate, "15:00")
		if err != nil {
			return err
		}
		
		thirdPlaceMatch := &models.Match{
			TournamentID: tournamentID,
			Round:        models.RoundThirdPlace,
			Team1:        "TBD",
			Team2:        "TBD",
			Status:       models.MatchStatusPending,
			ScheduledAt:  thirdPlaceTime,
		}
		
		if err := s.matchRepo.Create(thirdPlaceMatch); err != nil {
			return fmt.Errorf("3位決定戦作成エラー: %v", err)
		}
		
		// 決勝（15:30）
		finalTime, err := s.parseTimeSlot(baseDate, "15:30")
		if err != nil {
			return err
		}
		
		finalMatch := &models.Match{
			TournamentID: tournamentID,
			Round:        models.RoundFinal,
			Team1:        "TBD",
			Team2:        "TBD",
			Status:       models.MatchStatusPending,
			ScheduledAt:  finalTime,
		}
		
		if err := s.matchRepo.Create(finalMatch); err != nil {
			return fmt.Errorf("決勝作成エラー: %v", err)
		}
		
	} else {
		// 晴天時フォーマット
		// 準々決勝（13:00, 13:20, 13:40, 14:00）
		quarterTimes := []string{"13:00", "13:20", "13:40", "14:00"}
		for i, timeSlot := range quarterTimes {
			scheduledTime, err := s.parseTimeSlot(baseDate, timeSlot)
			if err != nil {
				return err
			}
			
			match := &models.Match{
				TournamentID: tournamentID,
				Round:        models.RoundQuarterfinal,
				Team1:        "TBD",
				Team2:        "TBD",
				Status:       models.MatchStatusPending,
				ScheduledAt:  scheduledTime,
			}
			
			if err := s.matchRepo.Create(match); err != nil {
				return fmt.Errorf("準々決勝試合%d作成エラー: %v", i+1, err)
			}
		}
		
		// 準決勝（14:30, 14:50）
		semiTimes := []string{"14:30", "14:50"}
		for i, timeSlot := range semiTimes {
			scheduledTime, err := s.parseTimeSlot(baseDate, timeSlot)
			if err != nil {
				return err
			}
			
			match := &models.Match{
				TournamentID: tournamentID,
				Round:        models.RoundSemifinal,
				Team1:        "TBD",
				Team2:        "TBD",
				Status:       models.MatchStatusPending,
				ScheduledAt:  scheduledTime,
			}
			
			if err := s.matchRepo.Create(match); err != nil {
				return fmt.Errorf("準決勝試合%d作成エラー: %v", i+1, err)
			}
		}
		
		// 3位決定戦（15:10）
		thirdPlaceTime, err := s.parseTimeSlot(baseDate, "15:10")
		if err != nil {
			return err
		}
		
		thirdPlaceMatch := &models.Match{
			TournamentID: tournamentID,
			Round:        models.RoundThirdPlace,
			Team1:        "TBD",
			Team2:        "TBD",
			Status:       models.MatchStatusPending,
			ScheduledAt:  thirdPlaceTime,
		}
		
		if err := s.matchRepo.Create(thirdPlaceMatch); err != nil {
			return fmt.Errorf("3位決定戦作成エラー: %v", err)
		}
		
		// 決勝（15:30）
		finalTime, err := s.parseTimeSlot(baseDate, "15:30")
		if err != nil {
			return err
		}
		
		finalMatch := &models.Match{
			TournamentID: tournamentID,
			Round:        models.RoundFinal,
			Team1:        "TBD",
			Team2:        "TBD",
			Status:       models.MatchStatusPending,
			ScheduledAt:  finalTime,
		}
		
		if err := s.matchRepo.Create(finalMatch); err != nil {
			return fmt.Errorf("決勝作成エラー: %v", err)
		}
	}
	
	return nil
}

// createSoccerPlaceholderMatches はサッカーのプレースホルダー試合を作成する
func (s *seedingServiceImpl) createSoccerPlaceholderMatches(tournamentID int, baseDate time.Time) error {
	// 準々決勝（11:30, 11:45, 13:00, 13:15）
	quarterTimes := []string{"11:30", "11:45", "13:00", "13:15"}
	for i, timeSlot := range quarterTimes {
		scheduledTime, err := s.parseTimeSlot(baseDate, timeSlot)
		if err != nil {
			return err
		}
		
		match := &models.Match{
			TournamentID: tournamentID,
			Round:        models.RoundQuarterfinal,
			Team1:        "TBD",
			Team2:        "TBD",
			Status:       models.MatchStatusPending,
			ScheduledAt:  scheduledTime,
		}
		
		if err := s.matchRepo.Create(match); err != nil {
			return fmt.Errorf("準々決勝試合%d作成エラー: %v", i+1, err)
		}
	}
	
	// 準決勝（13:30, 14:00）
	semiTimes := []string{"13:30", "14:00"}
	for i, timeSlot := range semiTimes {
		scheduledTime, err := s.parseTimeSlot(baseDate, timeSlot)
		if err != nil {
			return err
		}
		
		match := &models.Match{
			TournamentID: tournamentID,
			Round:        models.RoundSemifinal,
			Team1:        "TBD",
			Team2:        "TBD",
			Status:       models.MatchStatusPending,
			ScheduledAt:  scheduledTime,
		}
		
		if err := s.matchRepo.Create(match); err != nil {
			return fmt.Errorf("準決勝試合%d作成エラー: %v", i+1, err)
		}
	}
	
	// 3位決定戦（14:30）
	thirdPlaceTime, err := s.parseTimeSlot(baseDate, "14:30")
	if err != nil {
		return err
	}
	
	thirdPlaceMatch := &models.Match{
		TournamentID: tournamentID,
		Round:        models.RoundThirdPlace,
		Team1:        "TBD",
		Team2:        "TBD",
		Status:       models.MatchStatusPending,
		ScheduledAt:  thirdPlaceTime,
	}
	
	if err := s.matchRepo.Create(thirdPlaceMatch); err != nil {
		return fmt.Errorf("3位決定戦作成エラー: %v", err)
	}
	
	// 決勝（15:00）
	finalTime, err := s.parseTimeSlot(baseDate, "15:00")
	if err != nil {
		return err
	}
	
	finalMatch := &models.Match{
		TournamentID: tournamentID,
		Round:        models.RoundFinal,
		Team1:        "TBD",
		Team2:        "TBD",
		Status:       models.MatchStatusPending,
		ScheduledAt:  finalTime,
	}
	
	if err := s.matchRepo.Create(finalMatch); err != nil {
		return fmt.Errorf("決勝作成エラー: %v", err)
	}
	
	return nil
}

// VolleyballMatch はバレーボールの試合データ構造体
type VolleyballMatch struct {
	Team1 string
	Team2 string
	Time  string
}

// TableTennisMatch は卓球の試合データ構造体
type TableTennisMatch struct {
	Team1 string
	Team2 string
	Time  string
}

// SoccerMatch はサッカーの試合データ構造体
type SoccerMatch struct {
	Team1 string
	Team2 string
	Time  string
}
package models

// ユーザー役割の定数
const (
	RoleAdmin = "admin"
)

// スポーツタイプの定数
const (
	SportVolleyball  = "volleyball"
	SportTableTennis = "table_tennis"
	SportSoccer      = "soccer"
)

// トーナメントフォーマットの定数
const (
	FormatStandard = "standard"
	FormatRainy    = "rainy" // 卓球の雨天時フォーマット
)

// トーナメントステータスの定数
const (
	TournamentStatusActive    = "active"
	TournamentStatusCompleted = "completed"
)

// 試合ステータスの定数
const (
	MatchStatusPending   = "pending"
	MatchStatusCompleted = "completed"
)

// ラウンド名の定数
const (
	Round1stRound     = "1st_round"
	RoundQuarterfinal = "quarterfinal"
	RoundSemifinal    = "semifinal"
	RoundThirdPlace   = "third_place"
	RoundFinal        = "final"
	RoundLoserBracket = "loser_bracket" // 卓球の敗者復活戦用
)

// 有効なスポーツかどうかを判定する
func IsValidSport(sport string) bool {
	validSports := []string{
		SportVolleyball,
		SportTableTennis,
		SportSoccer,
	}
	
	for _, validSport := range validSports {
		if sport == validSport {
			return true
		}
	}
	return false
}

// 有効なトーナメントフォーマットかどうかを判定する
func IsValidTournamentFormat(format string) bool {
	validFormats := []string{
		FormatStandard,
		FormatRainy,
	}
	
	for _, validFormat := range validFormats {
		if format == validFormat {
			return true
		}
	}
	return false
}

// 有効なトーナメントステータスかどうかを判定する
func IsValidTournamentStatus(status string) bool {
	validStatuses := []string{
		TournamentStatusActive,
		TournamentStatusCompleted,
	}
	
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// 有効な試合ステータスかどうかを判定する
func IsValidMatchStatus(status string) bool {
	validStatuses := []string{
		MatchStatusPending,
		MatchStatusCompleted,
	}
	
	for _, validStatus := range validStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// 有効なラウンド名かどうかを判定する
func IsValidRound(round string) bool {
	validRounds := []string{
		Round1stRound,
		RoundQuarterfinal,
		RoundSemifinal,
		RoundThirdPlace,
		RoundFinal,
		RoundLoserBracket,
	}
	
	for _, validRound := range validRounds {
		if round == validRound {
			return true
		}
	}
	return false
}

// スポーツに応じた有効なラウンドを取得する
func GetValidRoundsForSport(sport string) []string {
	switch sport {
	case SportVolleyball:
		return []string{
			Round1stRound,
			RoundQuarterfinal,
			RoundSemifinal,
			RoundThirdPlace,
			RoundFinal,
		}
	case SportTableTennis:
		return []string{
			Round1stRound,
			RoundQuarterfinal,
			RoundSemifinal,
			RoundThirdPlace,
			RoundFinal,
			RoundLoserBracket, // 雨天時のみ
		}
	case SportSoccer:
		return []string{
			Round1stRound,
			RoundQuarterfinal,
			RoundSemifinal,
			RoundThirdPlace,
			RoundFinal,
		}
	default:
		return []string{}
	}
}
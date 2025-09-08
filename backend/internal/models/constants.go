package models

// ユーザー役割の定数
const (
	RoleAdmin = "admin"
)

// 後方互換性のための文字列定数（非推奨：新しいコードではenum型を使用）
const (
	// SportType用の文字列定数（非推奨）
	SportVolleyball  = "volleyball"
	SportTableTennis = "table_tennis"
	SportSoccer      = "soccer"
	
	// TournamentFormat用の文字列定数（非推奨）
	FormatStandard = "standard"
	FormatRainy    = "rainy"
	
	// TournamentStatus用の文字列定数（非推奨）
	TournamentStatusRegistration = "registration"
	TournamentStatusActive       = "active"
	TournamentStatusCompleted    = "completed"
	
	// MatchStatus用の文字列定数（非推奨）
	MatchStatusPending   = "pending"
	MatchStatusCompleted = "completed"
	
	// RoundType用の文字列定数（非推奨）
	Round1stRound     = "1st_round"
	RoundQuarterfinal = "quarterfinal"
	RoundSemifinal    = "semifinal"
	RoundThirdPlace   = "third_place"
	RoundFinal        = "final"
	RoundLoserBracket = "loser_bracket"
)

// 後方互換性のための関数（非推奨：新しいコードではenum型のメソッドを使用）

// IsValidSport は有効なスポーツかどうかを判定する（非推奨）
func IsValidSport(sport string) bool {
	return SportType(sport).IsValid()
}

// IsValidTournamentFormat は有効なトーナメントフォーマットかどうかを判定する（非推奨）
func IsValidTournamentFormat(format string) bool {
	return TournamentFormat(format).IsValid()
}

// IsValidTournamentStatus は有効なトーナメントステータスかどうかを判定する（非推奨）
func IsValidTournamentStatus(status string) bool {
	return TournamentStatus(status).IsValid()
}

// IsValidMatchStatus は有効な試合ステータスかどうかを判定する（非推奨）
func IsValidMatchStatus(status string) bool {
	return MatchStatus(status).IsValid()
}

// IsValidRound は有効なラウンド名かどうかを判定する（非推奨）
func IsValidRound(round string) bool {
	return RoundType(round).IsValid()
}

// GetValidRoundsForSport はスポーツに応じた有効なラウンドを取得する（非推奨）
// 新しいコードではtypes.goのGetValidRoundsForSport関数を使用してください
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
package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// SportType はスポーツ種目を表す列挙型
type SportType string

const (
	SportTypeVolleyball  SportType = "volleyball"
	SportTypeTableTennis SportType = "table_tennis"
	SportTypeSoccer      SportType = "soccer"
)

// String はSportTypeの文字列表現を返す
func (s SportType) String() string {
	return string(s)
}

// IsValid はSportTypeが有効かどうかを判定する
func (s SportType) IsValid() bool {
	switch s {
	case SportTypeVolleyball, SportTypeTableTennis, SportTypeSoccer:
		return true
	default:
		return false
	}
}

// Value はdatabase/sql/driverインターフェースを実装する
func (s SportType) Value() (driver.Value, error) {
	return string(s), nil
}

// Scan はdatabase/sql/driverインターフェースを実装する
func (s *SportType) Scan(value interface{}) error {
	if value == nil {
		*s = ""
		return nil
	}
	
	switch v := value.(type) {
	case string:
		*s = SportType(v)
	case []byte:
		*s = SportType(v)
	default:
		return fmt.Errorf("cannot scan %T into SportType", value)
	}
	
	return nil
}

// TournamentStatus はトーナメントステータスを表す列挙型
type TournamentStatus string

const (
	TournamentStatusRegistrationEnum TournamentStatus = "registration"
	TournamentStatusActiveEnum       TournamentStatus = "active"
	TournamentStatusCompletedEnum    TournamentStatus = "completed"
	TournamentStatusCancelledEnum    TournamentStatus = "cancelled"
)

// String はTournamentStatusの文字列表現を返す
func (t TournamentStatus) String() string {
	return string(t)
}

// IsValid はTournamentStatusが有効かどうかを判定する
func (t TournamentStatus) IsValid() bool {
	switch t {
	case TournamentStatusRegistrationEnum, TournamentStatusActiveEnum, TournamentStatusCompletedEnum, TournamentStatusCancelledEnum:
		return true
	default:
		return false
	}
}

// Value はdatabase/sql/driverインターフェースを実装する
func (t TournamentStatus) Value() (driver.Value, error) {
	return string(t), nil
}

// Scan はdatabase/sql/driverインターフェースを実装する
func (t *TournamentStatus) Scan(value interface{}) error {
	if value == nil {
		*t = ""
		return nil
	}
	
	switch v := value.(type) {
	case string:
		*t = TournamentStatus(v)
	case []byte:
		*t = TournamentStatus(v)
	default:
		return fmt.Errorf("cannot scan %T into TournamentStatus", value)
	}
	
	return nil
}

// MatchStatus は試合ステータスを表す列挙型
type MatchStatus string

const (
	MatchStatusPendingEnum    MatchStatus = "pending"
	MatchStatusInProgressEnum MatchStatus = "in_progress"
	MatchStatusCompletedEnum  MatchStatus = "completed"
	MatchStatusCancelledEnum  MatchStatus = "cancelled"
)

// String はMatchStatusの文字列表現を返す
func (m MatchStatus) String() string {
	return string(m)
}

// IsValid はMatchStatusが有効かどうかを判定する
func (m MatchStatus) IsValid() bool {
	switch m {
	case MatchStatusPendingEnum, MatchStatusInProgressEnum, MatchStatusCompletedEnum, MatchStatusCancelledEnum:
		return true
	default:
		return false
	}
}

// Value はdatabase/sql/driverインターフェースを実装する
func (m MatchStatus) Value() (driver.Value, error) {
	return string(m), nil
}

// Scan はdatabase/sql/driverインターフェースを実装する
func (m *MatchStatus) Scan(value interface{}) error {
	if value == nil {
		*m = ""
		return nil
	}
	
	switch v := value.(type) {
	case string:
		*m = MatchStatus(v)
	case []byte:
		*m = MatchStatus(v)
	default:
		return fmt.Errorf("cannot scan %T into MatchStatus", value)
	}
	
	return nil
}

// TournamentFormat はトーナメント形式を表す列挙型
type TournamentFormat string

const (
	TournamentFormatStandard TournamentFormat = "standard"
	TournamentFormatRainy    TournamentFormat = "rainy"
)

// String はTournamentFormatの文字列表現を返す
func (f TournamentFormat) String() string {
	return string(f)
}

// IsValid はTournamentFormatが有効かどうかを判定する
func (f TournamentFormat) IsValid() bool {
	switch f {
	case TournamentFormatStandard, TournamentFormatRainy:
		return true
	default:
		return false
	}
}

// Value はdatabase/sql/driverインターフェースを実装する
func (f TournamentFormat) Value() (driver.Value, error) {
	return string(f), nil
}

// Scan はdatabase/sql/driverインターフェースを実装する
func (f *TournamentFormat) Scan(value interface{}) error {
	if value == nil {
		*f = ""
		return nil
	}
	
	switch v := value.(type) {
	case string:
		*f = TournamentFormat(v)
	case []byte:
		*f = TournamentFormat(v)
	default:
		return fmt.Errorf("cannot scan %T into TournamentFormat", value)
	}
	
	return nil
}

// RoundType はラウンド種別を表す列挙型
type RoundType string

const (
	Round1stRoundEnum     RoundType = "1st_round"
	Round2ndRoundEnum     RoundType = "2nd_round"
	Round3rdRoundEnum     RoundType = "3rd_round"
	Round4thRoundEnum     RoundType = "4th_round"
	RoundQuarterfinalEnum RoundType = "quarterfinal"
	RoundSemifinalEnum    RoundType = "semifinal"
	RoundThirdPlaceEnum   RoundType = "third_place"
	RoundFinalEnum        RoundType = "final"
	RoundLoserBracketEnum RoundType = "loser_bracket"
)

// String はRoundTypeの文字列表現を返す
func (r RoundType) String() string {
	return string(r)
}

// IsValid はRoundTypeが有効かどうかを判定する
func (r RoundType) IsValid() bool {
	switch r {
	case Round1stRoundEnum, Round2ndRoundEnum, Round3rdRoundEnum, Round4thRoundEnum,
		 RoundQuarterfinalEnum, RoundSemifinalEnum, RoundThirdPlaceEnum, RoundFinalEnum, RoundLoserBracketEnum:
		return true
	default:
		return false
	}
}

// Value はdatabase/sql/driverインターフェースを実装する
func (r RoundType) Value() (driver.Value, error) {
	return string(r), nil
}

// Scan はdatabase/sql/driverインターフェースを実装する
func (r *RoundType) Scan(value interface{}) error {
	if value == nil {
		*r = ""
		return nil
	}
	
	switch v := value.(type) {
	case string:
		*r = RoundType(v)
	case []byte:
		*r = RoundType(v)
	default:
		return fmt.Errorf("cannot scan %T into RoundType", value)
	}
	
	return nil
}

// DateTime はISO 8601形式の日時を扱うカスタム型
type DateTime struct {
	time.Time
}

// NewDateTime は新しいDateTimeを作成する
func NewDateTime(t time.Time) DateTime {
	return DateTime{Time: t}
}

// Now は現在時刻のDateTimeを返す
func Now() DateTime {
	return DateTime{Time: time.Now().UTC()}
}

// MarshalJSON はJSONマーシャリング時にISO 8601形式で出力する
func (dt DateTime) MarshalJSON() ([]byte, error) {
	if dt.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(dt.Time.UTC().Format(time.RFC3339))
}

// UnmarshalJSON はJSONアンマーシャリング時にISO 8601形式から解析する
func (dt *DateTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	
	if str == "" || str == "null" {
		dt.Time = time.Time{}
		return nil
	}
	
	// 複数のフォーマットを試行
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			dt.Time = t.UTC()
			return nil
		}
	}
	
	return fmt.Errorf("invalid datetime format: %s", str)
}

// Value はdatabase/sql/driverインターフェースを実装する
func (dt DateTime) Value() (driver.Value, error) {
	if dt.Time.IsZero() {
		return nil, nil
	}
	return dt.Time.UTC(), nil
}

// Scan はdatabase/sql/driverインターフェースを実装する
func (dt *DateTime) Scan(value interface{}) error {
	if value == nil {
		dt.Time = time.Time{}
		return nil
	}
	
	switch v := value.(type) {
	case time.Time:
		dt.Time = v.UTC()
	case string:
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return err
		}
		dt.Time = t.UTC()
	default:
		return fmt.Errorf("cannot scan %T into DateTime", value)
	}
	
	return nil
}

// String はDateTimeの文字列表現を返す（ISO 8601形式）
func (dt DateTime) String() string {
	if dt.Time.IsZero() {
		return ""
	}
	return dt.Time.UTC().Format(time.RFC3339)
}

// IsZero は時刻がゼロ値かどうかを判定する
func (dt DateTime) IsZero() bool {
	return dt.Time.IsZero()
}

// NullableDateTime はnull許可のDateTime型
type NullableDateTime struct {
	DateTime DateTime
	Valid    bool
}

// NewNullableDateTime は新しいNullableDateTimeを作成する
func NewNullableDateTime(t time.Time) NullableDateTime {
	return NullableDateTime{
		DateTime: NewDateTime(t),
		Valid:    true,
	}
}

// MarshalJSON はJSONマーシャリング時の処理
func (ndt NullableDateTime) MarshalJSON() ([]byte, error) {
	if !ndt.Valid {
		return []byte("null"), nil
	}
	return ndt.DateTime.MarshalJSON()
}

// UnmarshalJSON はJSONアンマーシャリング時の処理
func (ndt *NullableDateTime) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	
	if str == "" || str == "null" {
		ndt.Valid = false
		ndt.DateTime = DateTime{}
		return nil
	}
	
	ndt.Valid = true
	return ndt.DateTime.UnmarshalJSON(data)
}

// Value はdatabase/sql/driverインターフェースを実装する
func (ndt NullableDateTime) Value() (driver.Value, error) {
	if !ndt.Valid {
		return nil, nil
	}
	return ndt.DateTime.Value()
}

// Scan はdatabase/sql/driverインターフェースを実装する
func (ndt *NullableDateTime) Scan(value interface{}) error {
	if value == nil {
		ndt.Valid = false
		ndt.DateTime = DateTime{}
		return nil
	}
	
	ndt.Valid = true
	return ndt.DateTime.Scan(value)
}

// String はNullableDateTimeの文字列表現を返す
func (ndt NullableDateTime) String() string {
	if !ndt.Valid {
		return ""
	}
	return ndt.DateTime.String()
}

// GetValidRoundsForSportType はスポーツに応じた有効なラウンドを取得する
func GetValidRoundsForSportType(sport SportType) []RoundType {
	switch sport {
	case SportTypeVolleyball:
		return []RoundType{
			Round1stRoundEnum,
			RoundQuarterfinalEnum,
			RoundSemifinalEnum,
			RoundThirdPlaceEnum,
			RoundFinalEnum,
		}
	case SportTypeTableTennis:
		return []RoundType{
			Round1stRoundEnum,
			RoundQuarterfinalEnum,
			RoundSemifinalEnum,
			RoundThirdPlaceEnum,
			RoundFinalEnum,
			RoundLoserBracketEnum, // 雨天時のみ
		}
	case SportTypeSoccer:
		return []RoundType{
			Round1stRoundEnum,
			RoundQuarterfinalEnum,
			RoundSemifinalEnum,
			RoundThirdPlaceEnum,
			RoundFinalEnum,
		}
	default:
		return []RoundType{}
	}
}

// IsValidRoundForSport は指定されたスポーツで有効なラウンドかどうかを判定する
func IsValidRoundForSport(sport SportType, round RoundType) bool {
	validRounds := GetValidRoundsForSportType(sport)
	for _, validRound := range validRounds {
		if round == validRound {
			return true
		}
	}
	return false
}
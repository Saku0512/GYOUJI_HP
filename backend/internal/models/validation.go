package models

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"
)

// ValidationError はバリデーションエラーの詳細情報を含む構造体
type ValidationError struct {
	Field   string `json:"field"`   // エラーが発生したフィールド名
	Message string `json:"message"` // エラーメッセージ
	Value   string `json:"value"`   // 入力された値
	Code    string `json:"code"`    // エラーコード
	Rule    string `json:"rule"`    // 違反したバリデーションルール
}

// Error はerrorインターフェースを実装する
func (ve ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", ve.Field, ve.Message)
}

// ValidationErrors は複数のバリデーションエラーを管理する
type ValidationErrors []ValidationError

// Error はerrorインターフェースを実装する
func (ves ValidationErrors) Error() string {
	if len(ves) == 0 {
		return "no validation errors"
	}
	
	var messages []string
	for _, ve := range ves {
		messages = append(messages, ve.Error())
	}
	
	return strings.Join(messages, "; ")
}

// Add はバリデーションエラーを追加する
func (ves *ValidationErrors) Add(field, message, value, code, rule string) {
	*ves = append(*ves, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
		Code:    code,
		Rule:    rule,
	})
}

// AddError はValidationErrorを直接追加する
func (ves *ValidationErrors) AddError(err ValidationError) {
	*ves = append(*ves, err)
}

// HasErrors はエラーが存在するかどうかを返す
func (ves ValidationErrors) HasErrors() bool {
	return len(ves) > 0
}

// GetFieldErrors は特定のフィールドのエラーを取得する
func (ves ValidationErrors) GetFieldErrors(field string) []ValidationError {
	var fieldErrors []ValidationError
	for _, ve := range ves {
		if ve.Field == field {
			fieldErrors = append(fieldErrors, ve)
		}
	}
	return fieldErrors
}

// ToValidationErrorDetails はValidationErrorDetailの配列に変換する
func (ves ValidationErrors) ToValidationErrorDetails() []ValidationErrorDetail {
	details := make([]ValidationErrorDetail, len(ves))
	for i, ve := range ves {
		details[i] = ValidationErrorDetail{
			Field:   ve.Field,
			Message: ve.Message,
			Value:   ve.Value,
		}
	}
	return details
}

// ValidationRule はバリデーションルールを表すインターフェース
type ValidationRule interface {
	Validate(value interface{}, fieldName string) *ValidationError
	GetRuleName() string
}

// ValidationContext はバリデーション実行時のコンテキスト情報
type ValidationContext struct {
	Language string                 // 言語設定（将来の多言語対応用）
	Data     map[string]interface{} // 追加のコンテキストデータ
}

// NewValidationContext は新しいValidationContextを作成する
func NewValidationContext() *ValidationContext {
	return &ValidationContext{
		Language: "ja", // デフォルトは日本語
		Data:     make(map[string]interface{}),
	}
}

// SetLanguage は言語設定を変更する
func (vc *ValidationContext) SetLanguage(lang string) *ValidationContext {
	vc.Language = lang
	return vc
}

// SetData はコンテキストデータを設定する
func (vc *ValidationContext) SetData(key string, value interface{}) *ValidationContext {
	vc.Data[key] = value
	return vc
}

// GetData はコンテキストデータを取得する
func (vc *ValidationContext) GetData(key string) (interface{}, bool) {
	value, exists := vc.Data[key]
	return value, exists
}

// Validator は統一されたバリデーション機能を提供する
type Validator struct {
	rules   map[string][]ValidationRule // フィールド別のバリデーションルール
	context *ValidationContext          // バリデーションコンテキスト
}

// NewValidator は新しいValidatorを作成する
func NewValidator() *Validator {
	return &Validator{
		rules:   make(map[string][]ValidationRule),
		context: NewValidationContext(),
	}
}

// WithContext はバリデーションコンテキストを設定する
func (v *Validator) WithContext(ctx *ValidationContext) *Validator {
	v.context = ctx
	return v
}

// AddRule はフィールドにバリデーションルールを追加する
func (v *Validator) AddRule(fieldName string, rule ValidationRule) *Validator {
	v.rules[fieldName] = append(v.rules[fieldName], rule)
	return v
}

// ValidateStruct は構造体全体のバリデーションを実行する
func (v *Validator) ValidateStruct(data interface{}) ValidationErrors {
	var errors ValidationErrors
	
	// リフレクションを使用して構造体のフィールドを検証
	// 実装は複雑になるため、ここでは基本的な検証メソッドを提供
	
	return errors
}

// ValidateRequired は必須フィールドの検証を行う
func (v *Validator) ValidateRequired(value string, fieldName string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("required", fieldName),
			Value:   value,
			Code:    ErrorValidationRequiredField,
			Rule:    "required",
		}
	}
	return nil
}

// ValidateRequiredInt は必須整数フィールドの検証を行う
func (v *Validator) ValidateRequiredInt(value *int, fieldName string) *ValidationError {
	if value == nil {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("required", fieldName),
			Value:   "",
			Code:    ErrorValidationRequiredField,
			Rule:    "required",
		}
	}
	return nil
}

// ValidateStringLength は文字列長の検証を行う
func (v *Validator) ValidateStringLength(value string, fieldName string, min, max int) *ValidationError {
	length := utf8.RuneCountInString(value)
	
	if min > 0 && length < min {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("min_length", fieldName, min),
			Value:   value,
			Code:    ErrorValidationOutOfRange,
			Rule:    "min_length",
		}
	}
	
	if max > 0 && length > max {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("max_length", fieldName, max),
			Value:   value,
			Code:    ErrorValidationOutOfRange,
			Rule:    "max_length",
		}
	}
	
	return nil
}

// ValidateIntRange は整数の範囲検証を行う
func (v *Validator) ValidateIntRange(value int, fieldName string, min, max int) *ValidationError {
	if value < min {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("min_value", fieldName, min),
			Value:   fmt.Sprintf("%d", value),
			Code:    ErrorValidationOutOfRange,
			Rule:    "min_value",
		}
	}
	
	if max > 0 && value > max {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("max_value", fieldName, max),
			Value:   fmt.Sprintf("%d", value),
			Code:    ErrorValidationOutOfRange,
			Rule:    "max_value",
		}
	}
	
	return nil
}

// ValidateEmail はメールアドレスの形式を検証する
func (v *Validator) ValidateEmail(email string, fieldName string) *ValidationError {
	if email == "" {
		return nil // 空の場合はスキップ（必須チェックは別途行う）
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("invalid_email", fieldName),
			Value:   email,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "email",
		}
	}
	
	return nil
}

// ValidateURL はURL形式を検証する
func (v *Validator) ValidateURL(url string, fieldName string) *ValidationError {
	if url == "" {
		return nil // 空の場合はスキップ
	}
	
	urlRegex := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)
	if !urlRegex.MatchString(url) {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("invalid_url", fieldName),
			Value:   url,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "url",
		}
	}
	
	return nil
}

// ValidateAlphanumeric は英数字のみかどうかを検証する
func (v *Validator) ValidateAlphanumeric(value string, fieldName string) *ValidationError {
	if value == "" {
		return nil // 空の場合はスキップ
	}
	
	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphanumericRegex.MatchString(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("alphanumeric_only", fieldName),
			Value:   value,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "alphanumeric",
		}
	}
	
	return nil
}

// ValidateAlpha は英字のみかどうかを検証する
func (v *Validator) ValidateAlpha(value string, fieldName string) *ValidationError {
	if value == "" {
		return nil // 空の場合はスキップ
	}
	
	alphaRegex := regexp.MustCompile(`^[a-zA-Z]+$`)
	if !alphaRegex.MatchString(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("alpha_only", fieldName),
			Value:   value,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "alpha",
		}
	}
	
	return nil
}

// ValidateNumeric は数字のみかどうかを検証する
func (v *Validator) ValidateNumeric(value string, fieldName string) *ValidationError {
	if value == "" {
		return nil // 空の場合はスキップ
	}
	
	numericRegex := regexp.MustCompile(`^[0-9]+$`)
	if !numericRegex.MatchString(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("numeric_only", fieldName),
			Value:   value,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "numeric",
		}
	}
	
	return nil
}

// ValidateDateTime は日時の検証を行う
func (v *Validator) ValidateDateTime(value time.Time, fieldName string, allowZero bool) *ValidationError {
	if value.IsZero() && !allowZero {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("required", fieldName),
			Value:   "",
			Code:    ErrorValidationRequiredField,
			Rule:    "required",
		}
	}
	
	return nil
}

// ValidateFutureDateTime は未来の日時かどうかを検証する
func (v *Validator) ValidateFutureDateTime(value time.Time, fieldName string) *ValidationError {
	if !value.IsZero() && value.Before(time.Now()) {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("future_datetime", fieldName),
			Value:   value.Format(time.RFC3339),
			Code:    ErrorValidationInvalidFormat,
			Rule:    "future_datetime",
		}
	}
	
	return nil
}

// ValidatePastDateTime は過去の日時かどうかを検証する
func (v *Validator) ValidatePastDateTime(value time.Time, fieldName string) *ValidationError {
	if !value.IsZero() && value.After(time.Now()) {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("past_datetime", fieldName),
			Value:   value.Format(time.RFC3339),
			Code:    ErrorValidationInvalidFormat,
			Rule:    "past_datetime",
		}
	}
	
	return nil
}

// ValidateEnum は列挙型の検証を行う
func (v *Validator) ValidateEnum(value string, validValues []string, fieldName string) *ValidationError {
	if value == "" {
		return nil // 空の場合はスキップ
	}
	
	for _, validValue := range validValues {
		if value == validValue {
			return nil
		}
	}
	
	return &ValidationError{
		Field:   fieldName,
		Message: v.getLocalizedMessage("invalid_enum", fieldName, strings.Join(validValues, ", ")),
		Value:   value,
		Code:    ErrorValidationInvalidFormat,
		Rule:    "enum",
	}
}

// ValidateSportType はスポーツタイプの検証を行う
func (v *Validator) ValidateSportType(value SportType, fieldName string) *ValidationError {
	if !value.IsValid() {
		validValues := []string{
			string(SportTypeVolleyball),
			string(SportTypeTableTennis),
			string(SportTypeSoccer),
		}
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("invalid_sport_type", fieldName, strings.Join(validValues, ", ")),
			Value:   string(value),
			Code:    ErrorValidationInvalidFormat,
			Rule:    "sport_type",
		}
	}
	return nil
}

// ValidateTournamentStatus はトーナメントステータスの検証を行う
func (v *Validator) ValidateTournamentStatus(value TournamentStatus, fieldName string) *ValidationError {
	if !value.IsValid() {
		validValues := []string{
			string(TournamentStatusRegistrationEnum),
			string(TournamentStatusActiveEnum),
			string(TournamentStatusCompletedEnum),
			string(TournamentStatusCancelledEnum),
		}
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("invalid_tournament_status", fieldName, strings.Join(validValues, ", ")),
			Value:   string(value),
			Code:    ErrorValidationInvalidFormat,
			Rule:    "tournament_status",
		}
	}
	return nil
}

// ValidateMatchStatus は試合ステータスの検証を行う
func (v *Validator) ValidateMatchStatus(value MatchStatus, fieldName string) *ValidationError {
	if !value.IsValid() {
		validValues := []string{
			string(MatchStatusPendingEnum),
			string(MatchStatusInProgressEnum),
			string(MatchStatusCompletedEnum),
			string(MatchStatusCancelledEnum),
		}
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("invalid_match_status", fieldName, strings.Join(validValues, ", ")),
			Value:   string(value),
			Code:    ErrorValidationInvalidFormat,
			Rule:    "match_status",
		}
	}
	return nil
}

// ValidateUniqueStrings は文字列配列の重複をチェックする
func (v *Validator) ValidateUniqueStrings(values []string, fieldName string) *ValidationError {
	seen := make(map[string]bool)
	
	for _, value := range values {
		if seen[value] {
			return &ValidationError{
				Field:   fieldName,
				Message: v.getLocalizedMessage("duplicate_value", fieldName, value),
				Value:   value,
				Code:    ErrorValidationDuplicateValue,
				Rule:    "unique",
			}
		}
		seen[value] = true
	}
	
	return nil
}

// ValidateNotEmpty は空でないことを検証する
func (v *Validator) ValidateNotEmpty(value string, fieldName string) *ValidationError {
	if strings.TrimSpace(value) == "" {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("not_empty", fieldName),
			Value:   value,
			Code:    ErrorValidationRequiredField,
			Rule:    "not_empty",
		}
	}
	return nil
}

// ValidatePattern は正規表現パターンマッチングを行う
func (v *Validator) ValidatePattern(value string, pattern string, fieldName string) *ValidationError {
	if value == "" {
		return nil // 空の場合はスキップ
	}
	
	regex, err := regexp.Compile(pattern)
	if err != nil {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("invalid_pattern", fieldName),
			Value:   value,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "pattern",
		}
	}
	
	if !regex.MatchString(value) {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("pattern_mismatch", fieldName),
			Value:   value,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "pattern",
		}
	}
	
	return nil
}

// ValidatePassword はパスワードの強度を検証する
func (v *Validator) ValidatePassword(password string, fieldName string) *ValidationError {
	if len(password) < 8 {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("password_min_length", fieldName, 8),
			Value:   password,
			Code:    ErrorValidationOutOfRange,
			Rule:    "password_min_length",
		}
	}
	
	if len(password) > 100 {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("password_max_length", fieldName, 100),
			Value:   password,
			Code:    ErrorValidationOutOfRange,
			Rule:    "password_max_length",
		}
	}
	
	// 英数字を含むかチェック
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	
	if !hasLetter || !hasNumber {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("password_complexity", fieldName),
			Value:   password,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "password_complexity",
		}
	}
	
	return nil
}

// ValidatePasswordStrength はより厳密なパスワード強度を検証する
func (v *Validator) ValidatePasswordStrength(password string, fieldName string) *ValidationError {
	if err := v.ValidatePassword(password, fieldName); err != nil {
		return err
	}
	
	// 特殊文字を含むかチェック
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)
	if !hasSpecial {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("password_special_char", fieldName),
			Value:   password,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "password_special_char",
		}
	}
	
	// 大文字小文字を含むかチェック
	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	
	if !hasUpper || !hasLower {
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("password_case_mix", fieldName),
			Value:   password,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "password_case_mix",
		}
	}
	
	return nil
}

// ValidateTeamNames はチーム名の重複をチェックする
func (v *Validator) ValidateTeamNames(team1, team2 string) *ValidationError {
	if strings.TrimSpace(team1) == strings.TrimSpace(team2) {
		return &ValidationError{
			Field:   "teams",
			Message: v.getLocalizedMessage("same_team_match", "チーム"),
			Value:   fmt.Sprintf("%s vs %s", team1, team2),
			Code:    ErrorBusinessInvalidMatchResult,
			Rule:    "different_teams",
		}
	}
	return nil
}

// ValidateMatchScore は試合スコアの整合性を検証する
func (v *Validator) ValidateMatchScore(score1, score2 int, winner, team1, team2 string) ValidationErrors {
	var errors ValidationErrors
	
	// スコアの基本検証
	if score1 < 0 {
		errors.Add("score1", v.getLocalizedMessage("score_non_negative", "チーム1のスコア"), 
			fmt.Sprintf("%d", score1), ErrorValidationOutOfRange, "min_value")
	}
	
	if score2 < 0 {
		errors.Add("score2", v.getLocalizedMessage("score_non_negative", "チーム2のスコア"), 
			fmt.Sprintf("%d", score2), ErrorValidationOutOfRange, "min_value")
	}
	
	// 引き分けチェック
	if score1 == score2 {
		errors.Add("scores", v.getLocalizedMessage("no_draw_allowed", "スコア"), 
			fmt.Sprintf("%d-%d", score1, score2), ErrorBusinessInvalidMatchResult, "no_draw")
	}
	
	// 勝者の検証
	if winner != team1 && winner != team2 {
		errors.Add("winner", v.getLocalizedMessage("winner_must_be_participant", "勝者"), 
			winner, ErrorBusinessInvalidMatchResult, "valid_winner")
	}
	
	// スコアと勝者の整合性チェック
	if score1 > score2 && winner != team1 {
		errors.Add("winner", v.getLocalizedMessage("score_winner_mismatch", "勝者"), 
			winner, ErrorBusinessInvalidMatchResult, "score_consistency")
	}
	
	if score2 > score1 && winner != team2 {
		errors.Add("winner", v.getLocalizedMessage("score_winner_mismatch", "勝者"), 
			winner, ErrorBusinessInvalidMatchResult, "score_consistency")
	}
	
	return errors
}

// ValidateBusinessRules はビジネスルールの検証を行う
func (v *Validator) ValidateBusinessRules(tournament *Tournament, match *Match) ValidationErrors {
	var errors ValidationErrors
	
	// トーナメントが完了している場合は試合を作成できない
	if tournament != nil && tournament.IsCompleted() {
		errors.Add("tournament", v.getLocalizedMessage("tournament_completed", "トーナメント"), 
			tournament.Status, ErrorBusinessTournamentCompleted, "tournament_status")
	}
	
	// キャンセルされたトーナメントには試合を作成できない
	if tournament != nil && tournament.IsCancelled() {
		errors.Add("tournament", v.getLocalizedMessage("tournament_cancelled", "トーナメント"), 
			tournament.Status, ErrorBusinessTournamentCompleted, "tournament_status")
	}
	
	// 完了した試合は更新できない
	if match != nil && match.IsCompleted() {
		errors.Add("match", v.getLocalizedMessage("match_completed", "試合"), 
			match.Status, ErrorBusinessMatchAlreadyCompleted, "match_status")
	}
	
	return errors
}

// ValidateRoundForSport はスポーツに対して有効なラウンドかを検証する
func (v *Validator) ValidateRoundForSport(sport SportType, round RoundType, fieldName string) *ValidationError {
	if !IsValidRoundForSport(sport, round) {
		validRounds := GetValidRoundsForSportType(sport)
		var validRoundStrings []string
		for _, r := range validRounds {
			validRoundStrings = append(validRoundStrings, string(r))
		}
		
		return &ValidationError{
			Field:   fieldName,
			Message: v.getLocalizedMessage("invalid_round_for_sport", fieldName, string(sport), strings.Join(validRoundStrings, ", ")),
			Value:   string(round),
			Code:    ErrorValidationInvalidFormat,
			Rule:    "valid_round_for_sport",
		}
	}
	return nil
}

// getLocalizedMessage はローカライズされたメッセージを取得する
func (v *Validator) getLocalizedMessage(messageKey string, args ...interface{}) string {
	// 日本語メッセージマップ
	messages := map[string]string{
		"required":                    "%sは必須です",
		"min_length":                  "%sは%d文字以上である必要があります",
		"max_length":                  "%sは%d文字以下である必要があります",
		"min_value":                   "%sは%d以上である必要があります",
		"max_value":                   "%sは%d以下である必要があります",
		"invalid_email":               "%sの形式が正しくありません",
		"invalid_url":                 "%sのURL形式が正しくありません",
		"alphanumeric_only":           "%sは英数字のみ使用可能です",
		"alpha_only":                  "%sは英字のみ使用可能です",
		"numeric_only":                "%sは数字のみ使用可能です",
		"future_datetime":             "%sは現在時刻より後である必要があります",
		"past_datetime":               "%sは現在時刻より前である必要があります",
		"invalid_enum":                "%sは無効な値です。有効な値: %s",
		"invalid_sport_type":          "%sは無効なスポーツです。有効な値: %s",
		"invalid_tournament_status":   "%sは無効なトーナメントステータスです。有効な値: %s",
		"invalid_match_status":        "%sは無効な試合ステータスです。有効な値: %s",
		"duplicate_value":             "%sに重複した値があります: %s",
		"not_empty":                   "%sは空にできません",
		"invalid_pattern":             "%sのパターンが無効です",
		"pattern_mismatch":            "%sの形式が正しくありません",
		"password_min_length":         "%sは%d文字以上である必要があります",
		"password_max_length":         "%sは%d文字以下である必要があります",
		"password_complexity":         "%sは英字と数字を含む必要があります",
		"password_special_char":       "%sは特殊文字を含む必要があります",
		"password_case_mix":           "%sは大文字と小文字を含む必要があります",
		"same_team_match":             "同じチーム同士の試合はできません",
		"score_non_negative":          "%sは0以上である必要があります",
		"no_draw_allowed":             "引き分けは許可されていません",
		"winner_must_be_participant":  "勝者は参加チームのいずれかである必要があります",
		"score_winner_mismatch":       "スコアと勝者が一致しません",
		"tournament_completed":        "完了したトーナメントには試合を追加できません",
		"tournament_cancelled":        "キャンセルされたトーナメントには試合を追加できません",
		"match_completed":             "完了した試合は更新できません",
		"invalid_round_for_sport":     "%sは%sで無効なラウンドです。有効な値: %s",
	}
	
	// 英語メッセージマップ（将来の多言語対応用）
	if v.context.Language == "en" {
		englishMessages := map[string]string{
			"required":                    "%s is required",
			"min_length":                  "%s must be at least %d characters",
			"max_length":                  "%s must be at most %d characters",
			"min_value":                   "%s must be at least %d",
			"max_value":                   "%s must be at most %d",
			"invalid_email":               "%s format is invalid",
			"invalid_url":                 "%s URL format is invalid",
			"alphanumeric_only":           "%s can only contain alphanumeric characters",
			"alpha_only":                  "%s can only contain alphabetic characters",
			"numeric_only":                "%s can only contain numeric characters",
			"future_datetime":             "%s must be in the future",
			"past_datetime":               "%s must be in the past",
			"invalid_enum":                "%s is invalid. Valid values: %s",
			"invalid_sport_type":          "%s is invalid sport. Valid values: %s",
			"invalid_tournament_status":   "%s is invalid tournament status. Valid values: %s",
			"invalid_match_status":        "%s is invalid match status. Valid values: %s",
			"duplicate_value":             "%s contains duplicate value: %s",
			"not_empty":                   "%s cannot be empty",
			"invalid_pattern":             "%s pattern is invalid",
			"pattern_mismatch":            "%s format is incorrect",
			"password_min_length":         "%s must be at least %d characters",
			"password_max_length":         "%s must be at most %d characters",
			"password_complexity":         "%s must contain letters and numbers",
			"password_special_char":       "%s must contain special characters",
			"password_case_mix":           "%s must contain uppercase and lowercase letters",
			"same_team_match":             "Cannot create match between same teams",
			"score_non_negative":          "%s must be non-negative",
			"no_draw_allowed":             "Draw is not allowed",
			"winner_must_be_participant":  "Winner must be one of the participating teams",
			"score_winner_mismatch":       "Score and winner do not match",
			"tournament_completed":        "Cannot add matches to completed tournament",
			"tournament_cancelled":        "Cannot add matches to cancelled tournament",
			"match_completed":             "Cannot update completed match",
			"invalid_round_for_sport":     "%s is invalid round for %s. Valid values: %s",
		}
		
		if msg, exists := englishMessages[messageKey]; exists {
			return fmt.Sprintf(msg, args...)
		}
	}
	
	// デフォルトは日本語
	if msg, exists := messages[messageKey]; exists {
		return fmt.Sprintf(msg, args...)
	}
	
	// メッセージが見つからない場合のフォールバック
	return fmt.Sprintf("バリデーションエラー: %s", messageKey)
}

// グローバルなValidator インスタンス
var DefaultValidator = NewValidator()

// ヘルパー関数（後方互換性のため維持、新しいコードでは直接Validatorを使用することを推奨）

// ValidateRequired は必須フィールドの検証を行う（グローバル関数）
func ValidateRequired(value string, fieldName string) error {
	if err := DefaultValidator.ValidateRequired(value, fieldName); err != nil {
		return errors.New(err.Message)
	}
	return nil
}

// ValidateStringLength は文字列長の検証を行う（グローバル関数）
func ValidateStringLength(value string, fieldName string, min, max int) error {
	if err := DefaultValidator.ValidateStringLength(value, fieldName, min, max); err != nil {
		return errors.New(err.Message)
	}
	return nil
}

// ValidateIntRange は整数の範囲検証を行う（グローバル関数）
func ValidateIntRange(value int, fieldName string, min, max int) error {
	if err := DefaultValidator.ValidateIntRange(value, fieldName, min, max); err != nil {
		return errors.New(err.Message)
	}
	return nil
}

// ValidateEnum は列挙型の検証を行う（グローバル関数）
func ValidateEnum(value string, validValues []string, fieldName string) error {
	if err := DefaultValidator.ValidateEnum(value, validValues, fieldName); err != nil {
		return errors.New(err.Message)
	}
	return nil
}

// ValidatePassword はパスワードの強度を検証する（グローバル関数）
func ValidatePassword(password string, fieldName string) error {
	if err := DefaultValidator.ValidatePassword(password, fieldName); err != nil {
		return errors.New(err.Message)
	}
	return nil
}

// ValidateMatchScore は試合スコアの整合性を検証する（グローバル関数）
func ValidateMatchScore(score1, score2 int, winner, team1, team2 string) error {
	errors := DefaultValidator.ValidateMatchScore(score1, score2, winner, team1, team2)
	if errors.HasErrors() {
		return fmt.Errorf("%s", errors.Error())
	}
	return nil
}

// 新しい統一バリデーション関数

// ValidateLoginRequest はログインリクエストの統一バリデーションを行う
func ValidateLoginRequest(req *LoginRequest) ValidationErrors {
	var errors ValidationErrors
	validator := NewValidator()
	
	if err := validator.ValidateRequired(req.Username, "username"); err != nil {
		errors.AddError(*err)
	}
	
	if err := validator.ValidateStringLength(req.Username, "username", 1, 50); err != nil {
		errors.AddError(*err)
	}
	
	if err := validator.ValidateAlphanumeric(req.Username, "username"); err != nil {
		errors.AddError(*err)
	}
	
	if err := validator.ValidateRequired(req.Password, "password"); err != nil {
		errors.AddError(*err)
	}
	
	if err := validator.ValidatePassword(req.Password, "password"); err != nil {
		errors.AddError(*err)
	}
	
	return errors
}

// ValidateCreateTournamentRequest はトーナメント作成リクエストの統一バリデーションを行う
func ValidateCreateTournamentRequest(req *CreateTournamentRequest) ValidationErrors {
	var errors ValidationErrors
	validator := NewValidator()
	
	if err := validator.ValidateSportType(req.Sport, "sport"); err != nil {
		errors.AddError(*err)
	}
	
	if err := validator.ValidateEnum(string(req.Format), []string{string(TournamentFormatStandard), string(TournamentFormatRainy)}, "format"); err != nil {
		errors.AddError(*err)
	}
	
	return errors
}

// ValidateSubmitMatchResultRequest は試合結果提出リクエストの統一バリデーションを行う
func ValidateSubmitMatchResultRequest(req *SubmitMatchResultRequest, team1, team2 string) ValidationErrors {
	var errors ValidationErrors
	validator := NewValidator()
	
	if err := validator.ValidateIntRange(req.Score1, "score1", 0, 1000); err != nil {
		errors.AddError(*err)
	}
	
	if err := validator.ValidateIntRange(req.Score2, "score2", 0, 1000); err != nil {
		errors.AddError(*err)
	}
	
	if err := validator.ValidateRequired(req.Winner, "winner"); err != nil {
		errors.AddError(*err)
	}
	
	// ビジネスロジック検証
	scoreErrors := validator.ValidateMatchScore(req.Score1, req.Score2, req.Winner, team1, team2)
	for _, scoreError := range scoreErrors {
		errors.AddError(scoreError)
	}
	
	return errors
}
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
	Field   string `json:"field"`
	Message string `json:"message"`
	Value   string `json:"value"`
	Code    string `json:"code"`
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
func (ves *ValidationErrors) Add(field, message, value, code string) {
	*ves = append(*ves, ValidationError{
		Field:   field,
		Message: message,
		Value:   value,
		Code:    code,
	})
}

// HasErrors はエラーが存在するかどうかを返す
func (ves ValidationErrors) HasErrors() bool {
	return len(ves) > 0
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

// Validator は統一されたバリデーション機能を提供する
type Validator struct{}

// NewValidator は新しいValidatorを作成する
func NewValidator() *Validator {
	return &Validator{}
}

// ValidateRequired は必須フィールドの検証を行う
func (v *Validator) ValidateRequired(value string, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%sは必須です", fieldName)
	}
	return nil
}

// ValidateStringLength は文字列長の検証を行う
func (v *Validator) ValidateStringLength(value string, fieldName string, min, max int) error {
	length := utf8.RuneCountInString(value)
	
	if min > 0 && length < min {
		return fmt.Errorf("%sは%d文字以上である必要があります", fieldName, min)
	}
	
	if max > 0 && length > max {
		return fmt.Errorf("%sは%d文字以下である必要があります", fieldName, max)
	}
	
	return nil
}

// ValidateIntRange は整数の範囲検証を行う
func (v *Validator) ValidateIntRange(value int, fieldName string, min, max int) error {
	if value < min {
		return fmt.Errorf("%sは%d以上である必要があります", fieldName, min)
	}
	
	if max > 0 && value > max {
		return fmt.Errorf("%sは%d以下である必要があります", fieldName, max)
	}
	
	return nil
}

// ValidateEmail はメールアドレスの形式を検証する
func (v *Validator) ValidateEmail(email string, fieldName string) error {
	if email == "" {
		return nil // 空の場合はスキップ（必須チェックは別途行う）
	}
	
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("%sの形式が正しくありません", fieldName)
	}
	
	return nil
}

// ValidateAlphanumeric は英数字のみかどうかを検証する
func (v *Validator) ValidateAlphanumeric(value string, fieldName string) error {
	if value == "" {
		return nil // 空の場合はスキップ
	}
	
	alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
	if !alphanumericRegex.MatchString(value) {
		return fmt.Errorf("%sは英数字のみ使用可能です", fieldName)
	}
	
	return nil
}

// ValidateDateTime は日時の検証を行う
func (v *Validator) ValidateDateTime(value time.Time, fieldName string, allowZero bool) error {
	if value.IsZero() && !allowZero {
		return fmt.Errorf("%sは必須です", fieldName)
	}
	
	// 未来の日時かどうかをチェック（必要に応じて）
	if !value.IsZero() && value.Before(time.Now()) {
		return fmt.Errorf("%sは現在時刻より後である必要があります", fieldName)
	}
	
	return nil
}

// ValidateEnum は列挙型の検証を行う
func (v *Validator) ValidateEnum(value string, validValues []string, fieldName string) error {
	if value == "" {
		return nil // 空の場合はスキップ
	}
	
	for _, validValue := range validValues {
		if value == validValue {
			return nil
		}
	}
	
	return fmt.Errorf("%sは無効な値です。有効な値: %s", fieldName, strings.Join(validValues, ", "))
}

// ValidateUniqueStrings は文字列配列の重複をチェックする
func (v *Validator) ValidateUniqueStrings(values []string, fieldName string) error {
	seen := make(map[string]bool)
	
	for _, value := range values {
		if seen[value] {
			return fmt.Errorf("%sに重複した値があります: %s", fieldName, value)
		}
		seen[value] = true
	}
	
	return nil
}

// ValidatePassword はパスワードの強度を検証する
func (v *Validator) ValidatePassword(password string, fieldName string) error {
	if len(password) < 8 {
		return fmt.Errorf("%sは8文字以上である必要があります", fieldName)
	}
	
	if len(password) > 100 {
		return fmt.Errorf("%sは100文字以下である必要があります", fieldName)
	}
	
	// 英数字を含むかチェック
	hasLetter := regexp.MustCompile(`[a-zA-Z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	
	if !hasLetter || !hasNumber {
		return fmt.Errorf("%sは英字と数字を含む必要があります", fieldName)
	}
	
	return nil
}

// ValidateTeamNames はチーム名の重複をチェックする
func (v *Validator) ValidateTeamNames(team1, team2 string) error {
	if strings.TrimSpace(team1) == strings.TrimSpace(team2) {
		return errors.New("同じチーム同士の試合はできません")
	}
	return nil
}

// ValidateMatchScore は試合スコアの整合性を検証する
func (v *Validator) ValidateMatchScore(score1, score2 int, winner, team1, team2 string) error {
	// スコアの基本検証
	if score1 < 0 || score2 < 0 {
		return errors.New("スコアは0以上である必要があります")
	}
	
	// 引き分けチェック
	if score1 == score2 {
		return errors.New("引き分けは許可されていません")
	}
	
	// 勝者の検証
	if winner != team1 && winner != team2 {
		return errors.New("勝者は参加チームのいずれかである必要があります")
	}
	
	// スコアと勝者の整合性チェック
	if score1 > score2 && winner != team1 {
		return errors.New("スコアと勝者が一致しません")
	}
	
	if score2 > score1 && winner != team2 {
		return errors.New("スコアと勝者が一致しません")
	}
	
	return nil
}

// ValidateBusinessRules はビジネスルールの検証を行う
func (v *Validator) ValidateBusinessRules(tournament *Tournament, match *Match) error {
	// トーナメントが完了している場合は試合を作成できない
	if tournament != nil && tournament.IsCompleted() {
		return errors.New("完了したトーナメントには試合を追加できません")
	}
	
	// キャンセルされたトーナメントには試合を作成できない
	if tournament != nil && tournament.IsCancelled() {
		return errors.New("キャンセルされたトーナメントには試合を追加できません")
	}
	
	// 完了した試合は更新できない
	if match != nil && match.IsCompleted() {
		return errors.New("完了した試合は更新できません")
	}
	
	return nil
}

// グローバルなValidator インスタンス
var DefaultValidator = NewValidator()

// ヘルパー関数

// ValidateRequired は必須フィールドの検証を行う（グローバル関数）
func ValidateRequired(value string, fieldName string) error {
	return DefaultValidator.ValidateRequired(value, fieldName)
}

// ValidateStringLength は文字列長の検証を行う（グローバル関数）
func ValidateStringLength(value string, fieldName string, min, max int) error {
	return DefaultValidator.ValidateStringLength(value, fieldName, min, max)
}

// ValidateIntRange は整数の範囲検証を行う（グローバル関数）
func ValidateIntRange(value int, fieldName string, min, max int) error {
	return DefaultValidator.ValidateIntRange(value, fieldName, min, max)
}

// ValidateEnum は列挙型の検証を行う（グローバル関数）
func ValidateEnum(value string, validValues []string, fieldName string) error {
	return DefaultValidator.ValidateEnum(value, validValues, fieldName)
}

// ValidatePassword はパスワードの強度を検証する（グローバル関数）
func ValidatePassword(password string, fieldName string) error {
	return DefaultValidator.ValidatePassword(password, fieldName)
}

// ValidateMatchScore は試合スコアの整合性を検証する（グローバル関数）
func ValidateMatchScore(score1, score2 int, winner, team1, team2 string) error {
	return DefaultValidator.ValidateMatchScore(score1, score2, winner, team1, team2)
}
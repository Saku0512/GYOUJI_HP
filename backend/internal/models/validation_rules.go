package models

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// RequiredRule は必須フィールドのバリデーションルール
type RequiredRule struct{}

func (r RequiredRule) Validate(value interface{}, fieldName string) *ValidationError {
	switch v := value.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは必須です", fieldName),
				Value:   v,
				Code:    ErrorValidationRequiredField,
				Rule:    "required",
			}
		}
	case *string:
		if v == nil || strings.TrimSpace(*v) == "" {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは必須です", fieldName),
				Value:   "",
				Code:    ErrorValidationRequiredField,
				Rule:    "required",
			}
		}
	case *int:
		if v == nil {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは必須です", fieldName),
				Value:   "",
				Code:    ErrorValidationRequiredField,
				Rule:    "required",
			}
		}
	}
	return nil
}

func (r RequiredRule) GetRuleName() string {
	return "required"
}

// MinLengthRule は最小文字数のバリデーションルール
type MinLengthRule struct {
	MinLength int
}

func (r MinLengthRule) Validate(value interface{}, fieldName string) *ValidationError {
	if str, ok := value.(string); ok {
		if len(str) < r.MinLength {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは%d文字以上である必要があります", fieldName, r.MinLength),
				Value:   str,
				Code:    ErrorValidationOutOfRange,
				Rule:    "min_length",
			}
		}
	}
	return nil
}

func (r MinLengthRule) GetRuleName() string {
	return "min_length"
}

// MaxLengthRule は最大文字数のバリデーションルール
type MaxLengthRule struct {
	MaxLength int
}

func (r MaxLengthRule) Validate(value interface{}, fieldName string) *ValidationError {
	if str, ok := value.(string); ok {
		if len(str) > r.MaxLength {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは%d文字以下である必要があります", fieldName, r.MaxLength),
				Value:   str,
				Code:    ErrorValidationOutOfRange,
				Rule:    "max_length",
			}
		}
	}
	return nil
}

func (r MaxLengthRule) GetRuleName() string {
	return "max_length"
}

// EmailRule はメールアドレス形式のバリデーションルール
type EmailRule struct{}

func (r EmailRule) Validate(value interface{}, fieldName string) *ValidationError {
	if str, ok := value.(string); ok && str != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(str) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sの形式が正しくありません", fieldName),
				Value:   str,
				Code:    ErrorValidationInvalidFormat,
				Rule:    "email",
			}
		}
	}
	return nil
}

func (r EmailRule) GetRuleName() string {
	return "email"
}

// AlphanumericRule は英数字のみのバリデーションルール
type AlphanumericRule struct{}

func (r AlphanumericRule) Validate(value interface{}, fieldName string) *ValidationError {
	if str, ok := value.(string); ok && str != "" {
		alphanumericRegex := regexp.MustCompile(`^[a-zA-Z0-9]+$`)
		if !alphanumericRegex.MatchString(str) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは英数字のみ使用可能です", fieldName),
				Value:   str,
				Code:    ErrorValidationInvalidFormat,
				Rule:    "alphanumeric",
			}
		}
	}
	return nil
}

func (r AlphanumericRule) GetRuleName() string {
	return "alphanumeric"
}

// MinValueRule は最小値のバリデーションルール
type MinValueRule struct {
	MinValue int
}

func (r MinValueRule) Validate(value interface{}, fieldName string) *ValidationError {
	switch v := value.(type) {
	case int:
		if v < r.MinValue {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは%d以上である必要があります", fieldName, r.MinValue),
				Value:   fmt.Sprintf("%d", v),
				Code:    ErrorValidationOutOfRange,
				Rule:    "min_value",
			}
		}
	case *int:
		if v != nil && *v < r.MinValue {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは%d以上である必要があります", fieldName, r.MinValue),
				Value:   fmt.Sprintf("%d", *v),
				Code:    ErrorValidationOutOfRange,
				Rule:    "min_value",
			}
		}
	}
	return nil
}

func (r MinValueRule) GetRuleName() string {
	return "min_value"
}

// MaxValueRule は最大値のバリデーションルール
type MaxValueRule struct {
	MaxValue int
}

func (r MaxValueRule) Validate(value interface{}, fieldName string) *ValidationError {
	switch v := value.(type) {
	case int:
		if v > r.MaxValue {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは%d以下である必要があります", fieldName, r.MaxValue),
				Value:   fmt.Sprintf("%d", v),
				Code:    ErrorValidationOutOfRange,
				Rule:    "max_value",
			}
		}
	case *int:
		if v != nil && *v > r.MaxValue {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは%d以下である必要があります", fieldName, r.MaxValue),
				Value:   fmt.Sprintf("%d", *v),
				Code:    ErrorValidationOutOfRange,
				Rule:    "max_value",
			}
		}
	}
	return nil
}

func (r MaxValueRule) GetRuleName() string {
	return "max_value"
}

// EnumRule は列挙型のバリデーションルール
type EnumRule struct {
	ValidValues []string
}

func (r EnumRule) Validate(value interface{}, fieldName string) *ValidationError {
	if str, ok := value.(string); ok && str != "" {
		for _, validValue := range r.ValidValues {
			if str == validValue {
				return nil
			}
		}
		return &ValidationError{
			Field:   fieldName,
			Message: fmt.Sprintf("%sは無効な値です。有効な値: %s", fieldName, strings.Join(r.ValidValues, ", ")),
			Value:   str,
			Code:    ErrorValidationInvalidFormat,
			Rule:    "enum",
		}
	}
	return nil
}

func (r EnumRule) GetRuleName() string {
	return "enum"
}

// FutureDateTimeRule は未来の日時のバリデーションルール
type FutureDateTimeRule struct{}

func (r FutureDateTimeRule) Validate(value interface{}, fieldName string) *ValidationError {
	switch v := value.(type) {
	case time.Time:
		if !v.IsZero() && v.Before(time.Now()) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは現在時刻より後である必要があります", fieldName),
				Value:   v.Format(time.RFC3339),
				Code:    ErrorValidationInvalidFormat,
				Rule:    "future_datetime",
			}
		}
	case DateTime:
		if !v.IsZero() && v.Time.Before(time.Now()) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは現在時刻より後である必要があります", fieldName),
				Value:   v.String(),
				Code:    ErrorValidationInvalidFormat,
				Rule:    "future_datetime",
			}
		}
	}
	return nil
}

func (r FutureDateTimeRule) GetRuleName() string {
	return "future_datetime"
}

// SportTypeRule はスポーツタイプのバリデーションルール
type SportTypeRule struct{}

func (r SportTypeRule) Validate(value interface{}, fieldName string) *ValidationError {
	if sport, ok := value.(SportType); ok {
		if !sport.IsValid() {
			validValues := []string{
				string(SportTypeVolleyball),
				string(SportTypeTableTennis),
				string(SportTypeSoccer),
			}
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは無効なスポーツです。有効な値: %s", fieldName, strings.Join(validValues, ", ")),
				Value:   string(sport),
				Code:    ErrorValidationInvalidFormat,
				Rule:    "sport_type",
			}
		}
	}
	return nil
}

func (r SportTypeRule) GetRuleName() string {
	return "sport_type"
}

// TournamentStatusRule はトーナメントステータスのバリデーションルール
type TournamentStatusRule struct{}

func (r TournamentStatusRule) Validate(value interface{}, fieldName string) *ValidationError {
	if status, ok := value.(TournamentStatus); ok {
		if !status.IsValid() {
			validValues := []string{
				string(TournamentStatusRegistrationEnum),
				string(TournamentStatusActiveEnum),
				string(TournamentStatusCompletedEnum),
				string(TournamentStatusCancelledEnum),
			}
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは無効なトーナメントステータスです。有効な値: %s", fieldName, strings.Join(validValues, ", ")),
				Value:   string(status),
				Code:    ErrorValidationInvalidFormat,
				Rule:    "tournament_status",
			}
		}
	}
	return nil
}

func (r TournamentStatusRule) GetRuleName() string {
	return "tournament_status"
}

// MatchStatusRule は試合ステータスのバリデーションルール
type MatchStatusRule struct{}

func (r MatchStatusRule) Validate(value interface{}, fieldName string) *ValidationError {
	if status, ok := value.(MatchStatus); ok {
		if !status.IsValid() {
			validValues := []string{
				string(MatchStatusPendingEnum),
				string(MatchStatusInProgressEnum),
				string(MatchStatusCompletedEnum),
				string(MatchStatusCancelledEnum),
			}
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは無効な試合ステータスです。有効な値: %s", fieldName, strings.Join(validValues, ", ")),
				Value:   string(status),
				Code:    ErrorValidationInvalidFormat,
				Rule:    "match_status",
			}
		}
	}
	return nil
}

func (r MatchStatusRule) GetRuleName() string {
	return "match_status"
}

// PatternRule は正規表現パターンのバリデーションルール
type PatternRule struct {
	Pattern string
	regex   *regexp.Regexp
}

func NewPatternRule(pattern string) *PatternRule {
	regex, _ := regexp.Compile(pattern)
	return &PatternRule{
		Pattern: pattern,
		regex:   regex,
	}
}

func (r PatternRule) Validate(value interface{}, fieldName string) *ValidationError {
	if str, ok := value.(string); ok && str != "" {
		if r.regex == nil {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sのパターンが無効です", fieldName),
				Value:   str,
				Code:    ErrorValidationInvalidFormat,
				Rule:    "pattern",
			}
		}
		
		if !r.regex.MatchString(str) {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sの形式が正しくありません", fieldName),
				Value:   str,
				Code:    ErrorValidationInvalidFormat,
				Rule:    "pattern",
			}
		}
	}
	return nil
}

func (r PatternRule) GetRuleName() string {
	return "pattern"
}

// PasswordRule はパスワード強度のバリデーションルール
type PasswordRule struct {
	MinLength        int
	MaxLength        int
	RequireAlpha     bool
	RequireNumeric   bool
	RequireSpecial   bool
	RequireMixedCase bool
}

func NewPasswordRule() *PasswordRule {
	return &PasswordRule{
		MinLength:        8,
		MaxLength:        100,
		RequireAlpha:     true,
		RequireNumeric:   true,
		RequireSpecial:   false,
		RequireMixedCase: false,
	}
}

func (r PasswordRule) Validate(value interface{}, fieldName string) *ValidationError {
	if str, ok := value.(string); ok {
		if len(str) < r.MinLength {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは%d文字以上である必要があります", fieldName, r.MinLength),
				Value:   str,
				Code:    ErrorValidationOutOfRange,
				Rule:    "password_min_length",
			}
		}
		
		if len(str) > r.MaxLength {
			return &ValidationError{
				Field:   fieldName,
				Message: fmt.Sprintf("%sは%d文字以下である必要があります", fieldName, r.MaxLength),
				Value:   str,
				Code:    ErrorValidationOutOfRange,
				Rule:    "password_max_length",
			}
		}
		
		if r.RequireAlpha {
			hasAlpha := regexp.MustCompile(`[a-zA-Z]`).MatchString(str)
			if !hasAlpha {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("%sは英字を含む必要があります", fieldName),
					Value:   str,
					Code:    ErrorValidationInvalidFormat,
					Rule:    "password_alpha",
				}
			}
		}
		
		if r.RequireNumeric {
			hasNumeric := regexp.MustCompile(`[0-9]`).MatchString(str)
			if !hasNumeric {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("%sは数字を含む必要があります", fieldName),
					Value:   str,
					Code:    ErrorValidationInvalidFormat,
					Rule:    "password_numeric",
				}
			}
		}
		
		if r.RequireSpecial {
			hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(str)
			if !hasSpecial {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("%sは特殊文字を含む必要があります", fieldName),
					Value:   str,
					Code:    ErrorValidationInvalidFormat,
					Rule:    "password_special",
				}
			}
		}
		
		if r.RequireMixedCase {
			hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(str)
			hasLower := regexp.MustCompile(`[a-z]`).MatchString(str)
			if !hasUpper || !hasLower {
				return &ValidationError{
					Field:   fieldName,
					Message: fmt.Sprintf("%sは大文字と小文字を含む必要があります", fieldName),
					Value:   str,
					Code:    ErrorValidationInvalidFormat,
					Rule:    "password_mixed_case",
				}
			}
		}
	}
	return nil
}

func (r PasswordRule) GetRuleName() string {
	return "password"
}

// 事前定義されたルールのファクトリー関数

// NewRequiredRule は必須フィールドルールを作成する
func NewRequiredRule() ValidationRule {
	return RequiredRule{}
}

// NewMinLengthRule は最小文字数ルールを作成する
func NewMinLengthRule(minLength int) ValidationRule {
	return MinLengthRule{MinLength: minLength}
}

// NewMaxLengthRule は最大文字数ルールを作成する
func NewMaxLengthRule(maxLength int) ValidationRule {
	return MaxLengthRule{MaxLength: maxLength}
}

// NewEmailRule はメールアドレスルールを作成する
func NewEmailRule() ValidationRule {
	return EmailRule{}
}

// NewAlphanumericRule は英数字ルールを作成する
func NewAlphanumericRule() ValidationRule {
	return AlphanumericRule{}
}

// NewMinValueRule は最小値ルールを作成する
func NewMinValueRule(minValue int) ValidationRule {
	return MinValueRule{MinValue: minValue}
}

// NewMaxValueRule は最大値ルールを作成する
func NewMaxValueRule(maxValue int) ValidationRule {
	return MaxValueRule{MaxValue: maxValue}
}

// NewEnumRule は列挙型ルールを作成する
func NewEnumRule(validValues []string) ValidationRule {
	return EnumRule{ValidValues: validValues}
}

// NewFutureDateTimeRule は未来日時ルールを作成する
func NewFutureDateTimeRule() ValidationRule {
	return FutureDateTimeRule{}
}

// NewSportTypeRule はスポーツタイプルールを作成する
func NewSportTypeRule() ValidationRule {
	return SportTypeRule{}
}

// NewTournamentStatusRule はトーナメントステータスルールを作成する
func NewTournamentStatusRule() ValidationRule {
	return TournamentStatusRule{}
}

// NewMatchStatusRule は試合ステータスルールを作成する
func NewMatchStatusRule() ValidationRule {
	return MatchStatusRule{}
}
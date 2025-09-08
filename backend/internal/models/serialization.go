package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"
)

// JSONSerializer は統一されたJSONシリアライゼーション機能を提供する
type JSONSerializer struct{}

// NewJSONSerializer は新しいJSONSerializerを作成する
func NewJSONSerializer() *JSONSerializer {
	return &JSONSerializer{}
}

// SerializeDateTime は日時をISO 8601形式でシリアライズする
func (js *JSONSerializer) SerializeDateTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

// DeserializeDateTime はISO 8601形式の文字列から日時をデシリアライズする
func (js *JSONSerializer) DeserializeDateTime(str string) (time.Time, error) {
	if str == "" {
		return time.Time{}, nil
	}
	
	// 複数のフォーマットを試行
	formats := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05Z",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	
	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			return t.UTC(), nil
		}
	}
	
	return time.Time{}, fmt.Errorf("invalid datetime format: %s", str)
}

// SerializeNullableDateTime はnull許可日時をシリアライズする
func (js *JSONSerializer) SerializeNullableDateTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return js.SerializeDateTime(*t)
}

// SerializeInt は整数をシリアライズする（null許可）
func (js *JSONSerializer) SerializeInt(i *int) interface{} {
	if i == nil {
		return nil
	}
	return *i
}

// SerializeString は文字列をシリアライズする（null許可）
func (js *JSONSerializer) SerializeString(s *string) interface{} {
	if s == nil {
		return nil
	}
	return *s
}

// DeserializeInt は文字列または数値から整数をデシリアライズする
func (js *JSONSerializer) DeserializeInt(value interface{}) (*int, error) {
	if value == nil {
		return nil, nil
	}
	
	switch v := value.(type) {
	case int:
		return &v, nil
	case int64:
		i := int(v)
		return &i, nil
	case float64:
		i := int(v)
		return &i, nil
	case string:
		if v == "" {
			return nil, nil
		}
		i, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("invalid integer format: %s", v)
		}
		return &i, nil
	default:
		return nil, fmt.Errorf("cannot convert %T to int", value)
	}
}

// DeserializeString は値から文字列をデシリアライズする
func (js *JSONSerializer) DeserializeString(value interface{}) (*string, error) {
	if value == nil {
		return nil, nil
	}
	
	switch v := value.(type) {
	case string:
		if v == "" {
			return nil, nil
		}
		return &v, nil
	default:
		str := fmt.Sprintf("%v", v)
		return &str, nil
	}
}

// NormalizeJSONNumbers はJSONの数値型を統一する
func (js *JSONSerializer) NormalizeJSONNumbers(data map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	
	for key, value := range data {
		switch v := value.(type) {
		case json.Number:
			// json.Numberを適切な型に変換
			if intVal, err := v.Int64(); err == nil {
				result[key] = int(intVal)
			} else if floatVal, err := v.Float64(); err == nil {
				result[key] = floatVal
			} else {
				result[key] = v.String()
			}
		case map[string]interface{}:
			// ネストされたオブジェクトも再帰的に処理
			result[key] = js.NormalizeJSONNumbers(v)
		case []interface{}:
			// 配列の要素も処理
			normalizedArray := make([]interface{}, len(v))
			for i, item := range v {
				if itemMap, ok := item.(map[string]interface{}); ok {
					normalizedArray[i] = js.NormalizeJSONNumbers(itemMap)
				} else {
					normalizedArray[i] = item
				}
			}
			result[key] = normalizedArray
		default:
			result[key] = value
		}
	}
	
	return result
}

// ValidateJSONStructure はJSONの構造を検証する
func (js *JSONSerializer) ValidateJSONStructure(data []byte, target interface{}) error {
	// まず構文チェック
	var temp interface{}
	if err := json.Unmarshal(data, &temp); err != nil {
		return fmt.Errorf("invalid JSON syntax: %w", err)
	}
	
	// 次に型チェック
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("JSON structure validation failed: %w", err)
	}
	
	return nil
}

// SanitizeJSONString はJSONエスケープを適用する
func (js *JSONSerializer) SanitizeJSONString(str string) string {
	// 危険な文字をエスケープ
	escaped, _ := json.Marshal(str)
	// 前後のダブルクォートを除去
	if len(escaped) >= 2 {
		return string(escaped[1 : len(escaped)-1])
	}
	return str
}

// ConvertToUnifiedResponse は既存のレスポンスを統一形式に変換する
func (js *JSONSerializer) ConvertToUnifiedResponse(data interface{}, message string, code int) map[string]interface{} {
	return map[string]interface{}{
		"success":   true,
		"data":      data,
		"message":   message,
		"code":      code,
		"timestamp": Now().String(),
	}
}

// ConvertToUnifiedErrorResponse はエラーレスポンスを統一形式に変換する
func (js *JSONSerializer) ConvertToUnifiedErrorResponse(errorCode, message string, statusCode int) map[string]interface{} {
	return map[string]interface{}{
		"success":   false,
		"error":     errorCode,
		"message":   message,
		"code":      statusCode,
		"timestamp": Now().String(),
	}
}

// グローバルなJSONSerializer インスタンス
var DefaultJSONSerializer = NewJSONSerializer()

// ヘルパー関数

// SerializeDateTime は日時をISO 8601形式でシリアライズする（グローバル関数）
func SerializeDateTime(t time.Time) string {
	return DefaultJSONSerializer.SerializeDateTime(t)
}

// DeserializeDateTime はISO 8601形式の文字列から日時をデシリアライズする（グローバル関数）
func DeserializeDateTime(str string) (time.Time, error) {
	return DefaultJSONSerializer.DeserializeDateTime(str)
}

// NormalizeJSONNumbers はJSONの数値型を統一する（グローバル関数）
func NormalizeJSONNumbers(data map[string]interface{}) map[string]interface{} {
	return DefaultJSONSerializer.NormalizeJSONNumbers(data)
}

// ValidateJSONStructure はJSONの構造を検証する（グローバル関数）
func ValidateJSONStructure(data []byte, target interface{}) error {
	return DefaultJSONSerializer.ValidateJSONStructure(data, target)
}

// SanitizeJSONString はJSONエスケープを適用する（グローバル関数）
func SanitizeJSONString(str string) string {
	return DefaultJSONSerializer.SanitizeJSONString(str)
}
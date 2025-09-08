package middleware

import (
	"bytes"
	"encoding/json"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"unicode"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
)

// SanitizationConfig は入力サニタイゼーションの設定
type SanitizationConfig struct {
	// HTML/XSS対策
	HTMLSanitization *HTMLSanitizationConfig `json:"html_sanitization"`
	
	// SQL インジェクション対策
	SQLInjectionProtection *SQLInjectionConfig `json:"sql_injection_protection"`
	
	// 入力値の正規化
	InputNormalization *InputNormalizationConfig `json:"input_normalization"`
	
	// ファイルアップロード対策
	FileUploadProtection *FileUploadConfig `json:"file_upload_protection"`
	
	// 除外するパス（サニタイゼーションをスキップ）
	ExcludedPaths []string `json:"excluded_paths"`
}

// HTMLSanitizationConfig はHTML/XSS対策の設定
type HTMLSanitizationConfig struct {
	Enabled                bool     `json:"enabled"`
	EscapeHTML             bool     `json:"escape_html"`              // HTMLエスケープを行う
	RemoveScriptTags       bool     `json:"remove_script_tags"`       // <script>タグを削除
	RemoveEventHandlers    bool     `json:"remove_event_handlers"`    // onclickなどのイベントハンドラーを削除
	AllowedTags            []string `json:"allowed_tags"`             // 許可するHTMLタグ
	AllowedAttributes      []string `json:"allowed_attributes"`       // 許可するHTML属性
	MaxInputLength         int      `json:"max_input_length"`         // 最大入力長
	BlockSuspiciousPatterns bool    `json:"block_suspicious_patterns"` // 疑わしいパターンをブロック
}

// SQLInjectionConfig はSQLインジェクション対策の設定
type SQLInjectionConfig struct {
	Enabled                 bool     `json:"enabled"`
	BlockSQLKeywords        bool     `json:"block_sql_keywords"`         // SQLキーワードをブロック
	BlockSQLComments        bool     `json:"block_sql_comments"`         // SQLコメントをブロック
	BlockUnionAttacks       bool     `json:"block_union_attacks"`        // UNION攻撃をブロック
	BlockBooleanAttacks     bool     `json:"block_boolean_attacks"`      // Boolean攻撃をブロック
	BlockTimeBasedAttacks   bool     `json:"block_time_based_attacks"`   // Time-based攻撃をブロック
	SuspiciousPatterns      []string `json:"suspicious_patterns"`        // 疑わしいパターンのリスト
	MaxQueryComplexity      int      `json:"max_query_complexity"`       // クエリの最大複雑度
}

// InputNormalizationConfig は入力値正規化の設定
type InputNormalizationConfig struct {
	Enabled              bool `json:"enabled"`
	TrimWhitespace       bool `json:"trim_whitespace"`        // 前後の空白を削除
	NormalizeUnicode     bool `json:"normalize_unicode"`      // Unicode正規化
	ConvertToLowerCase   bool `json:"convert_to_lower_case"`  // 小文字に変換（指定フィールドのみ）
	RemoveControlChars   bool `json:"remove_control_chars"`   // 制御文字を削除
	NormalizeLineBreaks  bool `json:"normalize_line_breaks"`  // 改行文字を統一
	MaxFieldLength       int  `json:"max_field_length"`       // フィールドの最大長
}

// FileUploadConfig はファイルアップロード対策の設定
type FileUploadConfig struct {
	Enabled           bool     `json:"enabled"`
	AllowedExtensions []string `json:"allowed_extensions"`    // 許可する拡張子
	AllowedMimeTypes  []string `json:"allowed_mime_types"`    // 許可するMIMEタイプ
	MaxFileSize       int64    `json:"max_file_size"`         // 最大ファイルサイズ（バイト）
	ScanForMalware    bool     `json:"scan_for_malware"`      // マルウェアスキャン
	ValidateFileHeader bool    `json:"validate_file_header"`  // ファイルヘッダーの検証
}

// SanitizationMiddleware は入力サニタイゼーションミドルウェア
type SanitizationMiddleware struct {
	config *SanitizationConfig
	
	// 正規表現パターン（コンパイル済み）
	scriptTagRegex      *regexp.Regexp
	eventHandlerRegex   *regexp.Regexp
	sqlKeywordRegex     *regexp.Regexp
	sqlCommentRegex     *regexp.Regexp
	unionAttackRegex    *regexp.Regexp
	booleanAttackRegex  *regexp.Regexp
	timeBasedAttackRegex *regexp.Regexp
	controlCharRegex    *regexp.Regexp
}

// NewSanitizationMiddleware は新しい入力サニタイゼーションミドルウェアを作成する
func NewSanitizationMiddleware(config *SanitizationConfig) *SanitizationMiddleware {
	sm := &SanitizationMiddleware{
		config: config,
	}
	
	// 正規表現パターンをコンパイル
	sm.compileRegexPatterns()
	
	return sm
}

// GetDefaultSanitizationConfig はデフォルトのサニタイゼーション設定を返す
func GetDefaultSanitizationConfig() *SanitizationConfig {
	return &SanitizationConfig{
		HTMLSanitization: &HTMLSanitizationConfig{
			Enabled:                 true,
			EscapeHTML:              true,
			RemoveScriptTags:        true,
			RemoveEventHandlers:     true,
			AllowedTags:             []string{"p", "br", "strong", "em", "u", "i", "b"},
			AllowedAttributes:       []string{"class", "id"},
			MaxInputLength:          10000,
			BlockSuspiciousPatterns: true,
		},
		SQLInjectionProtection: &SQLInjectionConfig{
			Enabled:               true,
			BlockSQLKeywords:      true,
			BlockSQLComments:      true,
			BlockUnionAttacks:     true,
			BlockBooleanAttacks:   true,
			BlockTimeBasedAttacks: true,
			SuspiciousPatterns: []string{
				`(?i)(union\s+select)`,
				`(?i)(drop\s+table)`,
				`(?i)(delete\s+from)`,
				`(?i)(insert\s+into)`,
				`(?i)(update\s+set)`,
				`(?i)(exec\s*\()`,
				`(?i)(script\s*>)`,
			},
			MaxQueryComplexity: 100,
		},
		InputNormalization: &InputNormalizationConfig{
			Enabled:             true,
			TrimWhitespace:      true,
			NormalizeUnicode:    true,
			ConvertToLowerCase:  false,
			RemoveControlChars:  true,
			NormalizeLineBreaks: true,
			MaxFieldLength:      1000,
		},
		FileUploadProtection: &FileUploadConfig{
			Enabled:           true,
			AllowedExtensions: []string{".jpg", ".jpeg", ".png", ".gif", ".pdf", ".txt", ".csv"},
			AllowedMimeTypes:  []string{"image/jpeg", "image/png", "image/gif", "application/pdf", "text/plain", "text/csv"},
			MaxFileSize:       10 * 1024 * 1024, // 10MB
			ScanForMalware:    false,
			ValidateFileHeader: true,
		},
		ExcludedPaths: []string{
			"/api/v1/health",
			"/api/v1/metrics",
		},
	}
}

// compileRegexPatterns は正規表現パターンをコンパイルする
func (sm *SanitizationMiddleware) compileRegexPatterns() {
	// HTMLサニタイゼーション用
	sm.scriptTagRegex = regexp.MustCompile(`(?i)<script[^>]*>.*?</script>`)
	sm.eventHandlerRegex = regexp.MustCompile(`(?i)\s*on\w+\s*=\s*["\'][^"\']*["\']`)
	
	// SQLインジェクション対策用
	sm.sqlKeywordRegex = regexp.MustCompile(`(?i)\b(select|insert|update|delete|drop|create|alter|exec|execute|union|or|and|where|from|into|values|set|table|database|schema)\b`)
	sm.sqlCommentRegex = regexp.MustCompile(`(--|/\*|\*/|#)`)
	sm.unionAttackRegex = regexp.MustCompile(`(?i)\bunion\s+(all\s+)?select\b`)
	sm.booleanAttackRegex = regexp.MustCompile(`(?i)(\'\s*(or|and)\s*\'\s*=\s*\'|\'\s*(or|and)\s*1\s*=\s*1)`)
	sm.timeBasedAttackRegex = regexp.MustCompile(`(?i)(sleep\s*\(|waitfor\s+delay|benchmark\s*\()`)
	
	// 制御文字削除用
	sm.controlCharRegex = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)
}

// SanitizeInputMiddleware は入力サニタイゼーションミドルウェアを返す
func (sm *SanitizationMiddleware) SanitizeInputMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 除外パスのチェック
		if sm.isExcludedPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		
		// リクエストボディのサニタイゼーション
		if c.Request.Body != nil && c.Request.ContentLength > 0 {
			if err := sm.sanitizeRequestBody(c); err != nil {
				sm.sendSanitizationError(c, "INPUT_SANITIZATION_ERROR", "入力データの処理中にエラーが発生しました")
				return
			}
		}
		
		// クエリパラメータのサニタイゼーション
		sm.sanitizeQueryParams(c)
		
		// フォームデータのサニタイゼーション
		sm.sanitizeFormData(c)
		
		c.Next()
	}
}

// sanitizeRequestBody はリクエストボディをサニタイゼーションする
func (sm *SanitizationMiddleware) sanitizeRequestBody(c *gin.Context) error {
	// リクエストボディを読み取り
	bodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return err
	}
	
	// ボディを復元
	c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	
	// Content-Typeに応じて処理
	contentType := c.GetHeader("Content-Type")
	
	if strings.Contains(contentType, "application/json") {
		return sm.sanitizeJSONBody(c, bodyBytes)
	} else if strings.Contains(contentType, "application/x-www-form-urlencoded") {
		return sm.sanitizeFormBody(c, bodyBytes)
	} else if strings.Contains(contentType, "multipart/form-data") {
		return sm.sanitizeMultipartBody(c)
	}
	
	return nil
}

// sanitizeJSONBody はJSONボディをサニタイゼーションする
func (sm *SanitizationMiddleware) sanitizeJSONBody(c *gin.Context, bodyBytes []byte) error {
	var data interface{}
	if err := json.Unmarshal(bodyBytes, &data); err != nil {
		return err
	}
	
	// データをサニタイゼーション
	sanitizedData := sm.sanitizeValue(data)
	
	// サニタイゼーション後のデータをJSONに変換
	sanitizedBytes, err := json.Marshal(sanitizedData)
	if err != nil {
		return err
	}
	
	// リクエストボディを更新
	c.Request.Body = io.NopCloser(bytes.NewBuffer(sanitizedBytes))
	c.Request.ContentLength = int64(len(sanitizedBytes))
	
	return nil
}

// sanitizeFormBody はフォームボディをサニタイゼーションする
func (sm *SanitizationMiddleware) sanitizeFormBody(c *gin.Context, bodyBytes []byte) error {
	// フォームデータをパース
	if err := c.Request.ParseForm(); err != nil {
		return err
	}
	
	// フォーム値をサニタイゼーション
	for key, values := range c.Request.Form {
		for i, value := range values {
			c.Request.Form[key][i] = sm.sanitizeString(value)
		}
	}
	
	return nil
}

// sanitizeMultipartBody はマルチパートボディをサニタイゼーションする
func (sm *SanitizationMiddleware) sanitizeMultipartBody(c *gin.Context) error {
	// マルチパートフォームをパース
	if err := c.Request.ParseMultipartForm(32 << 20); err != nil { // 32MB
		return err
	}
	
	// フォーム値をサニタイゼーション
	if c.Request.MultipartForm != nil {
		for key, values := range c.Request.MultipartForm.Value {
			for i, value := range values {
				c.Request.MultipartForm.Value[key][i] = sm.sanitizeString(value)
			}
		}
		
		// ファイルアップロードの検証
		if sm.config.FileUploadProtection != nil && sm.config.FileUploadProtection.Enabled {
			for _, fileHeaders := range c.Request.MultipartForm.File {
				for _, fileHeader := range fileHeaders {
					if err := sm.validateUploadedFile(fileHeader); err != nil {
						return err
					}
				}
			}
		}
	}
	
	return nil
}

// sanitizeQueryParams はクエリパラメータをサニタイゼーションする
func (sm *SanitizationMiddleware) sanitizeQueryParams(c *gin.Context) {
	query := c.Request.URL.Query()
	
	for key, values := range query {
		for i, value := range values {
			query[key][i] = sm.sanitizeString(value)
		}
	}
	
	c.Request.URL.RawQuery = query.Encode()
}

// sanitizeFormData はフォームデータをサニタイゼーションする
func (sm *SanitizationMiddleware) sanitizeFormData(c *gin.Context) {
	if c.Request.Form != nil {
		for key, values := range c.Request.Form {
			for i, value := range values {
				c.Request.Form[key][i] = sm.sanitizeString(value)
			}
		}
	}
}

// sanitizeValue は任意の値をサニタイゼーションする
func (sm *SanitizationMiddleware) sanitizeValue(value interface{}) interface{} {
	switch v := value.(type) {
	case string:
		return sm.sanitizeString(v)
	case map[string]interface{}:
		result := make(map[string]interface{})
		for key, val := range v {
			result[sm.sanitizeString(key)] = sm.sanitizeValue(val)
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, val := range v {
			result[i] = sm.sanitizeValue(val)
		}
		return result
	default:
		return value
	}
}

// sanitizeString は文字列をサニタイゼーションする
func (sm *SanitizationMiddleware) sanitizeString(input string) string {
	if input == "" {
		return input
	}
	
	result := input
	
	// 入力正規化
	if sm.config.InputNormalization != nil && sm.config.InputNormalization.Enabled {
		result = sm.normalizeInput(result)
	}
	
	// HTMLサニタイゼーション
	if sm.config.HTMLSanitization != nil && sm.config.HTMLSanitization.Enabled {
		result = sm.sanitizeHTML(result)
	}
	
	// SQLインジェクション対策
	if sm.config.SQLInjectionProtection != nil && sm.config.SQLInjectionProtection.Enabled {
		if sm.detectSQLInjection(result) {
			// SQLインジェクションの疑いがある場合は空文字列を返す
			return ""
		}
	}
	
	return result
}

// normalizeInput は入力値を正規化する
func (sm *SanitizationMiddleware) normalizeInput(input string) string {
	result := input
	config := sm.config.InputNormalization
	
	// 前後の空白を削除
	if config.TrimWhitespace {
		result = strings.TrimSpace(result)
	}
	
	// 制御文字を削除
	if config.RemoveControlChars {
		result = sm.controlCharRegex.ReplaceAllString(result, "")
	}
	
	// 改行文字を統一
	if config.NormalizeLineBreaks {
		result = strings.ReplaceAll(result, "\r\n", "\n")
		result = strings.ReplaceAll(result, "\r", "\n")
	}
	
	// 最大長チェック
	if config.MaxFieldLength > 0 && len(result) > config.MaxFieldLength {
		result = result[:config.MaxFieldLength]
	}
	
	// Unicode正規化（簡易版）
	if config.NormalizeUnicode {
		result = sm.normalizeUnicode(result)
	}
	
	return result
}

// sanitizeHTML はHTML/XSS対策を行う
func (sm *SanitizationMiddleware) sanitizeHTML(input string) string {
	result := input
	config := sm.config.HTMLSanitization
	
	// HTMLエスケープ
	if config.EscapeHTML {
		result = html.EscapeString(result)
	}
	
	// スクリプトタグを削除
	if config.RemoveScriptTags {
		result = sm.scriptTagRegex.ReplaceAllString(result, "")
	}
	
	// イベントハンドラーを削除
	if config.RemoveEventHandlers {
		result = sm.eventHandlerRegex.ReplaceAllString(result, "")
	}
	
	// 最大入力長チェック
	if config.MaxInputLength > 0 && len(result) > config.MaxInputLength {
		result = result[:config.MaxInputLength]
	}
	
	// 疑わしいパターンをブロック
	if config.BlockSuspiciousPatterns {
		if sm.containsSuspiciousPattern(result) {
			return ""
		}
	}
	
	return result
}

// detectSQLInjection はSQLインジェクションを検出する
func (sm *SanitizationMiddleware) detectSQLInjection(input string) bool {
	config := sm.config.SQLInjectionProtection
	
	// SQLキーワードをチェック
	if config.BlockSQLKeywords && sm.sqlKeywordRegex.MatchString(input) {
		return true
	}
	
	// SQLコメントをチェック
	if config.BlockSQLComments && sm.sqlCommentRegex.MatchString(input) {
		return true
	}
	
	// UNION攻撃をチェック
	if config.BlockUnionAttacks && sm.unionAttackRegex.MatchString(input) {
		return true
	}
	
	// Boolean攻撃をチェック
	if config.BlockBooleanAttacks && sm.booleanAttackRegex.MatchString(input) {
		return true
	}
	
	// Time-based攻撃をチェック
	if config.BlockTimeBasedAttacks && sm.timeBasedAttackRegex.MatchString(input) {
		return true
	}
	
	// カスタムパターンをチェック
	for _, pattern := range config.SuspiciousPatterns {
		if matched, _ := regexp.MatchString(pattern, input); matched {
			return true
		}
	}
	
	return false
}

// containsSuspiciousPattern は疑わしいパターンが含まれているかチェックする
func (sm *SanitizationMiddleware) containsSuspiciousPattern(input string) bool {
	suspiciousPatterns := []string{
		`<script`,
		`javascript:`,
		`vbscript:`,
		`onload=`,
		`onerror=`,
		`onclick=`,
		`onmouseover=`,
		`eval\s*\(`,
		`expression\s*\(`,
	}
	
	lowerInput := strings.ToLower(input)
	for _, pattern := range suspiciousPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerInput); matched {
			return true
		}
	}
	
	return false
}

// normalizeUnicode はUnicode正規化を行う（簡易版）
func (sm *SanitizationMiddleware) normalizeUnicode(input string) string {
	// 簡易的なUnicode正規化
	// 実際の実装では golang.org/x/text/unicode/norm を使用することを推奨
	result := strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' && r != '\t' {
			return -1 // 制御文字を削除
		}
		return r
	}, input)
	
	return result
}

// validateUploadedFile はアップロードされたファイルを検証する
func (sm *SanitizationMiddleware) validateUploadedFile(fileHeader *http.Request) error {
	// この実装は簡略化されています
	// 実際の実装では、ファイルヘッダーの検証、MIMEタイプの確認、
	// ファイルサイズの制限などを行います
	return nil
}

// isExcludedPath は除外パスかどうかをチェックする
func (sm *SanitizationMiddleware) isExcludedPath(path string) bool {
	for _, excludedPath := range sm.config.ExcludedPaths {
		if strings.HasPrefix(path, excludedPath) {
			return true
		}
	}
	return false
}

// sendSanitizationError はサニタイゼーションエラーレスポンスを送信する
func (sm *SanitizationMiddleware) sendSanitizationError(c *gin.Context, errorCode, message string) {
	response := models.NewErrorResponse(errorCode, message, http.StatusBadRequest)
	
	// リクエストIDを追加
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			response.RequestID = id
		}
	}
	
	c.JSON(http.StatusBadRequest, response)
	c.Abort()
}

// GetSanitizationStats はサニタイゼーション統計を取得する
func (sm *SanitizationMiddleware) GetSanitizationStats() map[string]interface{} {
	return map[string]interface{}{
		"html_sanitization_enabled":      sm.config.HTMLSanitization != nil && sm.config.HTMLSanitization.Enabled,
		"sql_injection_protection_enabled": sm.config.SQLInjectionProtection != nil && sm.config.SQLInjectionProtection.Enabled,
		"input_normalization_enabled":    sm.config.InputNormalization != nil && sm.config.InputNormalization.Enabled,
		"file_upload_protection_enabled": sm.config.FileUploadProtection != nil && sm.config.FileUploadProtection.Enabled,
		"excluded_paths_count":           len(sm.config.ExcludedPaths),
	}
}

// UpdateSanitizationConfig はサニタイゼーション設定を更新する
func (sm *SanitizationMiddleware) UpdateSanitizationConfig(config *SanitizationConfig) {
	sm.config = config
	sm.compileRegexPatterns() // 正規表現パターンを再コンパイル
}

// ValidateInput は単体で入力値を検証する関数
func ValidateInput(input string, config *SanitizationConfig) (string, error) {
	sm := NewSanitizationMiddleware(config)
	sanitized := sm.sanitizeString(input)
	
	if sanitized != input {
		return sanitized, nil
	}
	
	return input, nil
}

// EscapeHTML はHTMLエスケープを行う
func EscapeHTML(input string) string {
	return html.EscapeString(input)
}

// UnescapeHTML はHTMLアンエスケープを行う
func UnescapeHTML(input string) string {
	return html.UnescapeString(input)
}

// IsSafeString は文字列が安全かどうかをチェックする
func IsSafeString(input string) bool {
	// 基本的な安全性チェック
	suspiciousPatterns := []string{
		`<script`,
		`javascript:`,
		`vbscript:`,
		`onload=`,
		`onerror=`,
		`eval\s*\(`,
		`union\s+select`,
		`drop\s+table`,
		`delete\s+from`,
	}
	
	lowerInput := strings.ToLower(input)
	for _, pattern := range suspiciousPatterns {
		if matched, _ := regexp.MatchString(pattern, lowerInput); matched {
			return false
		}
	}
	
	return true
}
package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// SecurityConfig はセキュリティヘッダーの設定
type SecurityConfig struct {
	// Content Security Policy
	CSP *CSPConfig `json:"csp"`
	
	// HTTP Strict Transport Security
	HSTS *HSTSConfig `json:"hsts"`
	
	// その他のセキュリティヘッダー
	XFrameOptions           string `json:"x_frame_options"`           // X-Frame-Options
	XContentTypeOptions     string `json:"x_content_type_options"`    // X-Content-Type-Options
	XSSProtection           string `json:"xss_protection"`            // X-XSS-Protection
	ReferrerPolicy          string `json:"referrer_policy"`           // Referrer-Policy
	PermissionsPolicy       string `json:"permissions_policy"`        // Permissions-Policy
	CrossOriginEmbedderPolicy string `json:"cross_origin_embedder_policy"` // Cross-Origin-Embedder-Policy
	CrossOriginOpenerPolicy   string `json:"cross_origin_opener_policy"`   // Cross-Origin-Opener-Policy
	CrossOriginResourcePolicy string `json:"cross_origin_resource_policy"` // Cross-Origin-Resource-Policy
	
	// Cookie設定
	SecureCookies *SecureCookieConfig `json:"secure_cookies"`
}

// CSPConfig はContent Security Policyの設定
type CSPConfig struct {
	Enabled         bool     `json:"enabled"`
	DefaultSrc      []string `json:"default_src"`      // default-src
	ScriptSrc       []string `json:"script_src"`       // script-src
	StyleSrc        []string `json:"style_src"`        // style-src
	ImgSrc          []string `json:"img_src"`          // img-src
	ConnectSrc      []string `json:"connect_src"`      // connect-src
	FontSrc         []string `json:"font_src"`         // font-src
	ObjectSrc       []string `json:"object_src"`       // object-src
	MediaSrc        []string `json:"media_src"`        // media-src
	FrameSrc        []string `json:"frame_src"`        // frame-src
	ChildSrc        []string `json:"child_src"`        // child-src
	WorkerSrc       []string `json:"worker_src"`       // worker-src
	ManifestSrc     []string `json:"manifest_src"`     // manifest-src
	BaseURI         []string `json:"base_uri"`         // base-uri
	FormAction      []string `json:"form_action"`      // form-action
	FrameAncestors  []string `json:"frame_ancestors"`  // frame-ancestors
	UpgradeInsecureRequests bool `json:"upgrade_insecure_requests"` // upgrade-insecure-requests
	BlockAllMixedContent    bool `json:"block_all_mixed_content"`   // block-all-mixed-content
	ReportURI       string   `json:"report_uri"`       // report-uri
	ReportTo        string   `json:"report_to"`        // report-to
}

// HSTSConfig はHTTP Strict Transport Securityの設定
type HSTSConfig struct {
	Enabled           bool `json:"enabled"`
	MaxAge            int  `json:"max_age"`             // max-age (秒)
	IncludeSubDomains bool `json:"include_sub_domains"` // includeSubDomains
	Preload           bool `json:"preload"`             // preload
}

// SecureCookieConfig はセキュアなCookie設定
type SecureCookieConfig struct {
	Enabled    bool   `json:"enabled"`
	Secure     bool   `json:"secure"`      // Secure属性
	HttpOnly   bool   `json:"http_only"`   // HttpOnly属性
	SameSite   string `json:"same_site"`   // SameSite属性 (Strict, Lax, None)
	Domain     string `json:"domain"`      // Domain属性
	Path       string `json:"path"`        // Path属性
	MaxAge     int    `json:"max_age"`     // Max-Age属性 (秒)
}

// SecurityMiddleware はセキュリティヘッダーを設定するミドルウェア
type SecurityMiddleware struct {
	config *SecurityConfig
}

// NewSecurityMiddleware は新しいセキュリティミドルウェアを作成する
func NewSecurityMiddleware(config *SecurityConfig) *SecurityMiddleware {
	return &SecurityMiddleware{
		config: config,
	}
}

// GetDefaultSecurityConfig はデフォルトのセキュリティ設定を返す
func GetDefaultSecurityConfig() *SecurityConfig {
	return &SecurityConfig{
		CSP: &CSPConfig{
			Enabled:    true,
			DefaultSrc: []string{"'self'"},
			ScriptSrc:  []string{"'self'", "'unsafe-inline'", "'unsafe-eval'"},
			StyleSrc:   []string{"'self'", "'unsafe-inline'"},
			ImgSrc:     []string{"'self'", "data:", "https:"},
			ConnectSrc: []string{"'self'", "ws:", "wss:"},
			FontSrc:    []string{"'self'", "https:", "data:"},
			ObjectSrc:  []string{"'none'"},
			MediaSrc:   []string{"'self'"},
			FrameSrc:   []string{"'none'"},
			BaseURI:    []string{"'self'"},
			FormAction: []string{"'self'"},
			FrameAncestors: []string{"'none'"},
			UpgradeInsecureRequests: true,
			BlockAllMixedContent:    false,
		},
		HSTS: &HSTSConfig{
			Enabled:           true,
			MaxAge:            31536000, // 1年
			IncludeSubDomains: true,
			Preload:           false,
		},
		XFrameOptions:             "DENY",
		XContentTypeOptions:       "nosniff",
		XSSProtection:             "1; mode=block",
		ReferrerPolicy:            "strict-origin-when-cross-origin",
		PermissionsPolicy:         "geolocation=(), microphone=(), camera=()",
		CrossOriginEmbedderPolicy: "require-corp",
		CrossOriginOpenerPolicy:   "same-origin",
		CrossOriginResourcePolicy: "cross-origin",
		SecureCookies: &SecureCookieConfig{
			Enabled:  true,
			Secure:   true,
			HttpOnly: true,
			SameSite: "Strict",
			Path:     "/",
			MaxAge:   86400, // 24時間
		},
	}
}

// GetDevelopmentSecurityConfig は開発環境用のセキュリティ設定を返す
func GetDevelopmentSecurityConfig() *SecurityConfig {
	config := GetDefaultSecurityConfig()
	
	// 開発環境では一部の制限を緩和
	config.CSP.ScriptSrc = append(config.CSP.ScriptSrc, "localhost:*", "127.0.0.1:*")
	config.CSP.ConnectSrc = append(config.CSP.ConnectSrc, "localhost:*", "127.0.0.1:*")
	config.HSTS.Enabled = false // 開発環境ではHTTPSを使わない場合があるため
	config.SecureCookies.Secure = false // 開発環境ではHTTPを使う場合があるため
	config.SecureCookies.SameSite = "Lax" // 開発環境では緩和
	
	return config
}

// SecurityHeadersMiddleware はセキュリティヘッダーを設定するミドルウェアを返す
func (sm *SecurityMiddleware) SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy
		if sm.config.CSP != nil && sm.config.CSP.Enabled {
			cspValue := sm.buildCSPHeader(sm.config.CSP)
			c.Header("Content-Security-Policy", cspValue)
		}
		
		// HTTP Strict Transport Security
		if sm.config.HSTS != nil && sm.config.HSTS.Enabled {
			hstsValue := sm.buildHSTSHeader(sm.config.HSTS)
			c.Header("Strict-Transport-Security", hstsValue)
		}
		
		// X-Frame-Options
		if sm.config.XFrameOptions != "" {
			c.Header("X-Frame-Options", sm.config.XFrameOptions)
		}
		
		// X-Content-Type-Options
		if sm.config.XContentTypeOptions != "" {
			c.Header("X-Content-Type-Options", sm.config.XContentTypeOptions)
		}
		
		// X-XSS-Protection
		if sm.config.XSSProtection != "" {
			c.Header("X-XSS-Protection", sm.config.XSSProtection)
		}
		
		// Referrer-Policy
		if sm.config.ReferrerPolicy != "" {
			c.Header("Referrer-Policy", sm.config.ReferrerPolicy)
		}
		
		// Permissions-Policy
		if sm.config.PermissionsPolicy != "" {
			c.Header("Permissions-Policy", sm.config.PermissionsPolicy)
		}
		
		// Cross-Origin-Embedder-Policy
		if sm.config.CrossOriginEmbedderPolicy != "" {
			c.Header("Cross-Origin-Embedder-Policy", sm.config.CrossOriginEmbedderPolicy)
		}
		
		// Cross-Origin-Opener-Policy
		if sm.config.CrossOriginOpenerPolicy != "" {
			c.Header("Cross-Origin-Opener-Policy", sm.config.CrossOriginOpenerPolicy)
		}
		
		// Cross-Origin-Resource-Policy
		if sm.config.CrossOriginResourcePolicy != "" {
			c.Header("Cross-Origin-Resource-Policy", sm.config.CrossOriginResourcePolicy)
		}
		
		// セキュアなCookie設定
		if sm.config.SecureCookies != nil && sm.config.SecureCookies.Enabled {
			sm.setSecureCookieDefaults(c)
		}
		
		c.Next()
	}
}

// buildCSPHeader はCSPヘッダーの値を構築する
func (sm *SecurityMiddleware) buildCSPHeader(csp *CSPConfig) string {
	var directives []string
	
	if len(csp.DefaultSrc) > 0 {
		directives = append(directives, "default-src "+strings.Join(csp.DefaultSrc, " "))
	}
	
	if len(csp.ScriptSrc) > 0 {
		directives = append(directives, "script-src "+strings.Join(csp.ScriptSrc, " "))
	}
	
	if len(csp.StyleSrc) > 0 {
		directives = append(directives, "style-src "+strings.Join(csp.StyleSrc, " "))
	}
	
	if len(csp.ImgSrc) > 0 {
		directives = append(directives, "img-src "+strings.Join(csp.ImgSrc, " "))
	}
	
	if len(csp.ConnectSrc) > 0 {
		directives = append(directives, "connect-src "+strings.Join(csp.ConnectSrc, " "))
	}
	
	if len(csp.FontSrc) > 0 {
		directives = append(directives, "font-src "+strings.Join(csp.FontSrc, " "))
	}
	
	if len(csp.ObjectSrc) > 0 {
		directives = append(directives, "object-src "+strings.Join(csp.ObjectSrc, " "))
	}
	
	if len(csp.MediaSrc) > 0 {
		directives = append(directives, "media-src "+strings.Join(csp.MediaSrc, " "))
	}
	
	if len(csp.FrameSrc) > 0 {
		directives = append(directives, "frame-src "+strings.Join(csp.FrameSrc, " "))
	}
	
	if len(csp.ChildSrc) > 0 {
		directives = append(directives, "child-src "+strings.Join(csp.ChildSrc, " "))
	}
	
	if len(csp.WorkerSrc) > 0 {
		directives = append(directives, "worker-src "+strings.Join(csp.WorkerSrc, " "))
	}
	
	if len(csp.ManifestSrc) > 0 {
		directives = append(directives, "manifest-src "+strings.Join(csp.ManifestSrc, " "))
	}
	
	if len(csp.BaseURI) > 0 {
		directives = append(directives, "base-uri "+strings.Join(csp.BaseURI, " "))
	}
	
	if len(csp.FormAction) > 0 {
		directives = append(directives, "form-action "+strings.Join(csp.FormAction, " "))
	}
	
	if len(csp.FrameAncestors) > 0 {
		directives = append(directives, "frame-ancestors "+strings.Join(csp.FrameAncestors, " "))
	}
	
	if csp.UpgradeInsecureRequests {
		directives = append(directives, "upgrade-insecure-requests")
	}
	
	if csp.BlockAllMixedContent {
		directives = append(directives, "block-all-mixed-content")
	}
	
	if csp.ReportURI != "" {
		directives = append(directives, "report-uri "+csp.ReportURI)
	}
	
	if csp.ReportTo != "" {
		directives = append(directives, "report-to "+csp.ReportTo)
	}
	
	return strings.Join(directives, "; ")
}

// buildHSTSHeader はHSTSヘッダーの値を構築する
func (sm *SecurityMiddleware) buildHSTSHeader(hsts *HSTSConfig) string {
	var parts []string
	
	parts = append(parts, "max-age="+strconv.Itoa(hsts.MaxAge))
	
	if hsts.IncludeSubDomains {
		parts = append(parts, "includeSubDomains")
	}
	
	if hsts.Preload {
		parts = append(parts, "preload")
	}
	
	return strings.Join(parts, "; ")
}

// setSecureCookieDefaults はセキュアなCookieのデフォルト設定を適用する
func (sm *SecurityMiddleware) setSecureCookieDefaults(c *gin.Context) {
	// Ginのコンテキストにセキュアなクッキー設定を保存
	// 実際のCookie設定は各ハンドラーで使用される
	c.Set("secure_cookie_config", sm.config.SecureCookies)
}

// SetSecureCookie はセキュアなCookieを設定するヘルパー関数
func SetSecureCookie(c *gin.Context, name, value string, maxAge int) {
	// コンテキストからセキュアなCookie設定を取得
	if configInterface, exists := c.Get("secure_cookie_config"); exists {
		if config, ok := configInterface.(*SecureCookieConfig); ok {
			// セキュアなCookie設定を適用
			sameSite := http.SameSiteStrictMode
			switch strings.ToLower(config.SameSite) {
			case "lax":
				sameSite = http.SameSiteLaxMode
			case "none":
				sameSite = http.SameSiteNoneMode
			case "strict":
				sameSite = http.SameSiteStrictMode
			}
			
			// maxAgeが指定されていない場合は設定から取得
			if maxAge == 0 {
				maxAge = config.MaxAge
			}
			
			cookie := &http.Cookie{
				Name:     name,
				Value:    value,
				Path:     config.Path,
				Domain:   config.Domain,
				MaxAge:   maxAge,
				Secure:   config.Secure,
				HttpOnly: config.HttpOnly,
				SameSite: sameSite,
			}
			
			http.SetCookie(c.Writer, cookie)
			return
		}
	}
	
	// フォールバック: デフォルトのセキュアなCookie設定
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	
	http.SetCookie(c.Writer, cookie)
}

// CSRFProtectionMiddleware はCSRF攻撃対策のミドルウェア
func CSRFProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// POSTリクエストの場合はCSRFトークンをチェック
		if c.Request.Method == "POST" || c.Request.Method == "PUT" || c.Request.Method == "DELETE" || c.Request.Method == "PATCH" {
			// CSRFトークンの検証
			token := c.GetHeader("X-CSRF-Token")
			if token == "" {
				token = c.PostForm("_token")
			}
			
			// セッションからCSRFトークンを取得して比較
			sessionToken, exists := c.Get("csrf_token")
			if !exists || token == "" || token != sessionToken {
				c.JSON(http.StatusForbidden, gin.H{
					"success": false,
					"error":   "CSRF_TOKEN_INVALID",
					"message": "CSRFトークンが無効です",
					"code":    http.StatusForbidden,
				})
				c.Abort()
				return
			}
		}
		
		c.Next()
	}
}

// XSSProtectionMiddleware はXSS攻撃対策のミドルウェア
func XSSProtectionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// レスポンスヘッダーにXSS対策を設定
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("X-Content-Type-Options", "nosniff")
		
		c.Next()
	}
}

// NoSniffMiddleware はMIMEタイプスニッフィング対策のミドルウェア
func NoSniffMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Next()
	}
}

// FrameOptionsMiddleware はクリックジャッキング対策のミドルウェア
func FrameOptionsMiddleware(option string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if option == "" {
			option = "DENY"
		}
		c.Header("X-Frame-Options", option)
		c.Next()
	}
}

// HSTSMiddleware はHTTP Strict Transport Securityのミドルウェア
func HSTSMiddleware(maxAge int, includeSubDomains, preload bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// HTTPSの場合のみHSTSヘッダーを設定
		if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
			var hstsValue strings.Builder
			hstsValue.WriteString("max-age=")
			hstsValue.WriteString(strconv.Itoa(maxAge))
			
			if includeSubDomains {
				hstsValue.WriteString("; includeSubDomains")
			}
			
			if preload {
				hstsValue.WriteString("; preload")
			}
			
			c.Header("Strict-Transport-Security", hstsValue.String())
		}
		
		c.Next()
	}
}

// ReferrerPolicyMiddleware はReferrer-Policyヘッダーを設定するミドルウェア
func ReferrerPolicyMiddleware(policy string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == "" {
			policy = "strict-origin-when-cross-origin"
		}
		c.Header("Referrer-Policy", policy)
		c.Next()
	}
}

// PermissionsPolicyMiddleware はPermissions-Policyヘッダーを設定するミドルウェア
func PermissionsPolicyMiddleware(policy string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if policy == "" {
			policy = "geolocation=(), microphone=(), camera=()"
		}
		c.Header("Permissions-Policy", policy)
		c.Next()
	}
}

// GetSecurityHeaders は現在設定されているセキュリティヘッダーを取得する
func GetSecurityHeaders(c *gin.Context) map[string]string {
	headers := make(map[string]string)
	
	securityHeaderNames := []string{
		"Content-Security-Policy",
		"Strict-Transport-Security",
		"X-Frame-Options",
		"X-Content-Type-Options",
		"X-XSS-Protection",
		"Referrer-Policy",
		"Permissions-Policy",
		"Cross-Origin-Embedder-Policy",
		"Cross-Origin-Opener-Policy",
		"Cross-Origin-Resource-Policy",
	}
	
	for _, headerName := range securityHeaderNames {
		if value := c.Writer.Header().Get(headerName); value != "" {
			headers[headerName] = value
		}
	}
	
	return headers
}

// ValidateSecurityConfig はセキュリティ設定の妥当性を検証する
func ValidateSecurityConfig(config *SecurityConfig) error {
	// CSP設定の検証
	if config.CSP != nil && config.CSP.Enabled {
		if len(config.CSP.DefaultSrc) == 0 {
			config.CSP.DefaultSrc = []string{"'self'"}
		}
	}
	
	// HSTS設定の検証
	if config.HSTS != nil && config.HSTS.Enabled {
		if config.HSTS.MaxAge <= 0 {
			config.HSTS.MaxAge = 31536000 // デフォルト1年
		}
	}
	
	// Cookie設定の検証
	if config.SecureCookies != nil && config.SecureCookies.Enabled {
		if config.SecureCookies.Path == "" {
			config.SecureCookies.Path = "/"
		}
		if config.SecureCookies.SameSite == "" {
			config.SecureCookies.SameSite = "Strict"
		}
		if config.SecureCookies.MaxAge <= 0 {
			config.SecureCookies.MaxAge = 86400 // デフォルト24時間
		}
	}
	
	return nil
}

// UpdateSecurityConfig はセキュリティ設定を更新する
func (sm *SecurityMiddleware) UpdateSecurityConfig(config *SecurityConfig) error {
	if err := ValidateSecurityConfig(config); err != nil {
		return err
	}
	
	sm.config = config
	return nil
}

// GetSecurityConfig は現在のセキュリティ設定を取得する
func (sm *SecurityMiddleware) GetSecurityConfig() *SecurityConfig {
	return sm.config
}

// IsSecureContext はセキュアなコンテキスト（HTTPS）かどうかを判定する
func IsSecureContext(c *gin.Context) bool {
	return c.Request.TLS != nil || 
		   c.GetHeader("X-Forwarded-Proto") == "https" ||
		   c.GetHeader("X-Forwarded-Ssl") == "on" ||
		   c.GetHeader("X-Url-Scheme") == "https"
}

// GetClientFingerprint はクライアントのフィンガープリントを生成する（セキュリティ監視用）
func GetClientFingerprint(c *gin.Context) string {
	var fingerprint strings.Builder
	
	fingerprint.WriteString(c.ClientIP())
	fingerprint.WriteString("|")
	fingerprint.WriteString(c.GetHeader("User-Agent"))
	fingerprint.WriteString("|")
	fingerprint.WriteString(c.GetHeader("Accept-Language"))
	fingerprint.WriteString("|")
	fingerprint.WriteString(c.GetHeader("Accept-Encoding"))
	
	return fingerprint.String()
}
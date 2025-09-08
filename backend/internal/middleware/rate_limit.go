package middleware

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitConfig はレート制限の設定
type RateLimitConfig struct {
	// エンドポイント別の制限設定
	EndpointLimits map[string]*EndpointLimit `json:"endpoint_limits"`
	// IP別の制限設定
	IPLimits *IPLimitConfig `json:"ip_limits"`
	// ユーザー別の制限設定
	UserLimits *UserLimitConfig `json:"user_limits"`
	// 除外するIPアドレス（管理者用など）
	ExcludedIPs []string `json:"excluded_ips"`
}

// EndpointLimit はエンドポイント別の制限設定
type EndpointLimit struct {
	RequestsPerSecond int           `json:"requests_per_second"` // 1秒あたりのリクエスト数
	BurstSize         int           `json:"burst_size"`          // バーストサイズ
	WindowDuration    time.Duration `json:"window_duration"`     // ウィンドウ期間
}

// IPLimitConfig はIP別の制限設定
type IPLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute"` // 1分あたりのリクエスト数
	BurstSize         int           `json:"burst_size"`          // バーストサイズ
	WindowDuration    time.Duration `json:"window_duration"`     // ウィンドウ期間
}

// UserLimitConfig はユーザー別の制限設定
type UserLimitConfig struct {
	RequestsPerMinute int           `json:"requests_per_minute"` // 1分あたりのリクエスト数
	BurstSize         int           `json:"burst_size"`          // バーストサイズ
	WindowDuration    time.Duration `json:"window_duration"`     // ウィンドウ期間
}

// RateLimiter はレート制限を管理する構造体
type RateLimiter struct {
	config         *RateLimitConfig
	ipLimiters     map[string]*rate.Limiter
	userLimiters   map[int]*rate.Limiter
	endpointLimiters map[string]map[string]*rate.Limiter // [endpoint][identifier]*rate.Limiter
	mutex          sync.RWMutex
	cleanupTicker  *time.Ticker
	stopCleanup    chan bool
}

// NewRateLimiter は新しいレート制限器を作成する
func NewRateLimiter(config *RateLimitConfig) *RateLimiter {
	rl := &RateLimiter{
		config:           config,
		ipLimiters:       make(map[string]*rate.Limiter),
		userLimiters:     make(map[int]*rate.Limiter),
		endpointLimiters: make(map[string]map[string]*rate.Limiter),
		cleanupTicker:    time.NewTicker(5 * time.Minute), // 5分ごとにクリーンアップ
		stopCleanup:      make(chan bool),
	}

	// バックグラウンドでの古いリミッターのクリーンアップを開始
	go rl.startCleanup()

	return rl
}

// GetDefaultConfig はデフォルトのレート制限設定を返す
func GetDefaultRateLimitConfig() *RateLimitConfig {
	return &RateLimitConfig{
		EndpointLimits: map[string]*EndpointLimit{
			// 認証関連エンドポイント（厳しい制限）
			"/api/v1/auth/login": {
				RequestsPerSecond: 5,
				BurstSize:         10,
				WindowDuration:    time.Minute,
			},
			"/api/v1/auth/register": {
				RequestsPerSecond: 2,
				BurstSize:         5,
				WindowDuration:    time.Minute,
			},
			// トーナメント関連エンドポイント（中程度の制限）
			"/api/v1/tournaments": {
				RequestsPerSecond: 30,
				BurstSize:         50,
				WindowDuration:    time.Minute,
			},
			// 試合関連エンドポイント（中程度の制限）
			"/api/v1/matches": {
				RequestsPerSecond: 20,
				BurstSize:         40,
				WindowDuration:    time.Minute,
			},
			// 管理者エンドポイント（緩い制限）
			"/api/v1/admin": {
				RequestsPerSecond: 50,
				BurstSize:         100,
				WindowDuration:    time.Minute,
			},
		},
		IPLimits: &IPLimitConfig{
			RequestsPerMinute: 1000,
			BurstSize:         100,
			WindowDuration:    time.Minute,
		},
		UserLimits: &UserLimitConfig{
			RequestsPerMinute: 500,
			BurstSize:         50,
			WindowDuration:    time.Minute,
		},
		ExcludedIPs: []string{
			"127.0.0.1",
			"::1",
		},
	}
}

// RateLimitMiddleware はレート制限ミドルウェアを返す
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 除外IPのチェック
		clientIP := rl.getClientIP(c)
		if rl.isExcludedIP(clientIP) {
			c.Next()
			return
		}

		// エンドポイント別制限のチェック
		if !rl.checkEndpointLimit(c, clientIP) {
			rl.sendRateLimitError(c, "RATE_LIMIT_ENDPOINT_EXCEEDED", "エンドポイントのレート制限に達しました")
			return
		}

		// IP別制限のチェック
		if !rl.checkIPLimit(c, clientIP) {
			rl.sendRateLimitError(c, "RATE_LIMIT_IP_EXCEEDED", "IPアドレスのレート制限に達しました")
			return
		}

		// ユーザー別制限のチェック（認証済みの場合）
		if userID, exists := c.Get("user_id"); exists {
			if id, ok := userID.(int); ok {
				if !rl.checkUserLimit(c, id) {
					rl.sendRateLimitError(c, "RATE_LIMIT_USER_EXCEEDED", "ユーザーのレート制限に達しました")
					return
				}
			}
		}

		c.Next()
	}
}

// checkEndpointLimit はエンドポイント別の制限をチェックする
func (rl *RateLimiter) checkEndpointLimit(c *gin.Context, clientIP string) bool {
	path := c.Request.URL.Path

	// エンドポイントパターンにマッチする制限設定を検索
	var endpointLimit *EndpointLimit
	var endpointKey string

	for pattern, limit := range rl.config.EndpointLimits {
		if rl.matchEndpointPattern(path, pattern) {
			endpointLimit = limit
			endpointKey = pattern
			break
		}
	}

	// 制限設定がない場合はスキップ
	if endpointLimit == nil {
		return true
	}

	// リミッターを取得または作成
	limiter := rl.getOrCreateEndpointLimiter(endpointKey, clientIP, endpointLimit)

	// レート制限をチェック
	allowed := limiter.Allow()
	
	// レスポンスヘッダーを設定
	rl.setRateLimitHeaders(c, limiter, endpointLimit)

	return allowed
}

// checkIPLimit はIP別の制限をチェックする
func (rl *RateLimiter) checkIPLimit(c *gin.Context, clientIP string) bool {
	if rl.config.IPLimits == nil {
		return true
	}

	// リミッターを取得または作成
	limiter := rl.getOrCreateIPLimiter(clientIP, rl.config.IPLimits)

	// レート制限をチェック
	allowed := limiter.Allow()

	// レスポンスヘッダーを設定
	rl.setIPLimitHeaders(c, limiter, rl.config.IPLimits)

	return allowed
}

// checkUserLimit はユーザー別の制限をチェックする
func (rl *RateLimiter) checkUserLimit(c *gin.Context, userID int) bool {
	if rl.config.UserLimits == nil {
		return true
	}

	// リミッターを取得または作成
	limiter := rl.getOrCreateUserLimiter(userID, rl.config.UserLimits)

	// レート制限をチェック
	allowed := limiter.Allow()

	// レスポンスヘッダーを設定
	rl.setUserLimitHeaders(c, limiter, rl.config.UserLimits)

	return allowed
}

// getOrCreateEndpointLimiter はエンドポイント用のリミッターを取得または作成する
func (rl *RateLimiter) getOrCreateEndpointLimiter(endpoint, identifier string, config *EndpointLimit) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	if rl.endpointLimiters[endpoint] == nil {
		rl.endpointLimiters[endpoint] = make(map[string]*rate.Limiter)
	}

	limiter, exists := rl.endpointLimiters[endpoint][identifier]
	if !exists {
		// 1秒あたりのリクエスト数をレートに変換
		rateLimit := rate.Every(time.Second / time.Duration(config.RequestsPerSecond))
		limiter = rate.NewLimiter(rateLimit, config.BurstSize)
		rl.endpointLimiters[endpoint][identifier] = limiter
	}

	return limiter
}

// getOrCreateIPLimiter はIP用のリミッターを取得または作成する
func (rl *RateLimiter) getOrCreateIPLimiter(ip string, config *IPLimitConfig) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.ipLimiters[ip]
	if !exists {
		// 1分あたりのリクエスト数をレートに変換
		rateLimit := rate.Every(time.Minute / time.Duration(config.RequestsPerMinute))
		limiter = rate.NewLimiter(rateLimit, config.BurstSize)
		rl.ipLimiters[ip] = limiter
	}

	return limiter
}

// getOrCreateUserLimiter はユーザー用のリミッターを取得または作成する
func (rl *RateLimiter) getOrCreateUserLimiter(userID int, config *UserLimitConfig) *rate.Limiter {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	limiter, exists := rl.userLimiters[userID]
	if !exists {
		// 1分あたりのリクエスト数をレートに変換
		rateLimit := rate.Every(time.Minute / time.Duration(config.RequestsPerMinute))
		limiter = rate.NewLimiter(rateLimit, config.BurstSize)
		rl.userLimiters[userID] = limiter
	}

	return limiter
}

// matchEndpointPattern はエンドポイントパターンにマッチするかチェックする
func (rl *RateLimiter) matchEndpointPattern(path, pattern string) bool {
	// 完全一致
	if path == pattern {
		return true
	}

	// プレフィックスマッチ（パターンが/で終わる場合）
	if strings.HasSuffix(pattern, "/") && strings.HasPrefix(path, pattern) {
		return true
	}

	// プレフィックスマッチ（パスがパターンで始まる場合）
	if strings.HasPrefix(path, pattern) {
		// パターンの後に/があるかチェック
		if len(path) > len(pattern) && path[len(pattern)] == '/' {
			return true
		}
	}

	return false
}

// getClientIP はクライアントのIPアドレスを取得する
func (rl *RateLimiter) getClientIP(c *gin.Context) string {
	// X-Forwarded-Forヘッダーをチェック
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// X-Real-IPヘッダーをチェック
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return strings.TrimSpace(xri)
	}

	// RemoteAddrから取得
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return c.Request.RemoteAddr
	}

	return ip
}

// isExcludedIP は除外IPかどうかをチェックする
func (rl *RateLimiter) isExcludedIP(ip string) bool {
	for _, excludedIP := range rl.config.ExcludedIPs {
		if ip == excludedIP {
			return true
		}
	}
	return false
}

// setRateLimitHeaders はレート制限関連のヘッダーを設定する
func (rl *RateLimiter) setRateLimitHeaders(c *gin.Context, limiter *rate.Limiter, config *EndpointLimit) {
	// 残りリクエスト数を計算（概算）
	remaining := int(limiter.Tokens())
	if remaining < 0 {
		remaining = 0
	}

	c.Header("X-RateLimit-Limit", strconv.Itoa(config.RequestsPerSecond))
	c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(config.WindowDuration).Unix(), 10))
}

// setIPLimitHeaders はIP制限関連のヘッダーを設定する
func (rl *RateLimiter) setIPLimitHeaders(c *gin.Context, limiter *rate.Limiter, config *IPLimitConfig) {
	remaining := int(limiter.Tokens())
	if remaining < 0 {
		remaining = 0
	}

	c.Header("X-RateLimit-IP-Limit", strconv.Itoa(config.RequestsPerMinute))
	c.Header("X-RateLimit-IP-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-IP-Reset", strconv.FormatInt(time.Now().Add(config.WindowDuration).Unix(), 10))
}

// setUserLimitHeaders はユーザー制限関連のヘッダーを設定する
func (rl *RateLimiter) setUserLimitHeaders(c *gin.Context, limiter *rate.Limiter, config *UserLimitConfig) {
	remaining := int(limiter.Tokens())
	if remaining < 0 {
		remaining = 0
	}

	c.Header("X-RateLimit-User-Limit", strconv.Itoa(config.RequestsPerMinute))
	c.Header("X-RateLimit-User-Remaining", strconv.Itoa(remaining))
	c.Header("X-RateLimit-User-Reset", strconv.FormatInt(time.Now().Add(config.WindowDuration).Unix(), 10))
}

// sendRateLimitError はレート制限エラーレスポンスを送信する
func (rl *RateLimiter) sendRateLimitError(c *gin.Context, errorCode, message string) {
	// 統一されたエラーレスポンスを作成
	response := models.NewErrorResponse(errorCode, message, http.StatusTooManyRequests)
	
	// リクエストIDを追加
	if requestID, exists := c.Get("request_id"); exists {
		if id, ok := requestID.(string); ok {
			response.RequestID = id
		}
	}

	// Retry-Afterヘッダーを設定（60秒後に再試行を推奨）
	c.Header("Retry-After", "60")
	
	c.JSON(http.StatusTooManyRequests, response)
	c.Abort()
}

// startCleanup は古いリミッターのクリーンアップを開始する
func (rl *RateLimiter) startCleanup() {
	for {
		select {
		case <-rl.cleanupTicker.C:
			rl.cleanup()
		case <-rl.stopCleanup:
			rl.cleanupTicker.Stop()
			return
		}
	}
}

// cleanup は使用されていないリミッターを削除する
func (rl *RateLimiter) cleanup() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	// 5分以上使用されていないリミッターを削除
	// 実際の実装では、最後のアクセス時間を記録して判定する
	// ここでは簡略化のため、定期的にすべてクリアする
	
	// IPリミッターのクリーンアップ
	if len(rl.ipLimiters) > 1000 { // 閾値を超えた場合のみクリア
		rl.ipLimiters = make(map[string]*rate.Limiter)
	}

	// ユーザーリミッターのクリーンアップ
	if len(rl.userLimiters) > 1000 { // 閾値を超えた場合のみクリア
		rl.userLimiters = make(map[int]*rate.Limiter)
	}

	// エンドポイントリミッターのクリーンアップ
	for endpoint, limiters := range rl.endpointLimiters {
		if len(limiters) > 1000 { // 閾値を超えた場合のみクリア
			rl.endpointLimiters[endpoint] = make(map[string]*rate.Limiter)
		}
	}
}

// Stop はレート制限器を停止する
func (rl *RateLimiter) Stop() {
	close(rl.stopCleanup)
}

// GetStats はレート制限の統計情報を取得する
func (rl *RateLimiter) GetStats() map[string]interface{} {
	rl.mutex.RLock()
	defer rl.mutex.RUnlock()

	stats := map[string]interface{}{
		"ip_limiters_count":       len(rl.ipLimiters),
		"user_limiters_count":     len(rl.userLimiters),
		"endpoint_limiters_count": len(rl.endpointLimiters),
	}

	// エンドポイント別の詳細統計
	endpointStats := make(map[string]int)
	for endpoint, limiters := range rl.endpointLimiters {
		endpointStats[endpoint] = len(limiters)
	}
	stats["endpoint_details"] = endpointStats

	return stats
}

// UpdateConfig は設定を更新する
func (rl *RateLimiter) UpdateConfig(config *RateLimitConfig) {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	rl.config = config
	
	// 既存のリミッターをクリア（新しい設定を適用するため）
	rl.ipLimiters = make(map[string]*rate.Limiter)
	rl.userLimiters = make(map[int]*rate.Limiter)
	rl.endpointLimiters = make(map[string]map[string]*rate.Limiter)
}
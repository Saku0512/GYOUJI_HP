package router

import (
	"time"

	"backend/internal/errors"
	"backend/internal/handler"
	"backend/internal/logger"
	"backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// Router はアプリケーションのルーターを管理する
type Router struct {
	engine      *gin.Engine
	authService service.AuthService
	handlers    *Handlers
}

// Handlers は全てのハンドラーをまとめる構造体
type Handlers struct {
	AuthHandler       *handler.AuthHandler
	TournamentHandler *handler.TournamentHandler
	MatchHandler      *handler.MatchHandler
}

// NewRouter は新しいルーターを作成する
func NewRouter(
	authService service.AuthService,
	tournamentService service.TournamentService,
	matchService service.MatchService,
) *Router {
	// Ginエンジンを作成
	engine := gin.New()

	// ハンドラーを初期化
	handlers := &Handlers{
		AuthHandler:       handler.NewAuthHandler(authService),
		TournamentHandler: handler.NewTournamentHandler(tournamentService),
		MatchHandler:      handler.NewMatchHandler(matchService),
	}

	router := &Router{
		engine:      engine,
		authService: authService,
		handlers:    handlers,
	}

	// ミドルウェアとルートを設定
	router.setupMiddleware()
	router.setupRoutes()

	return router
}

// setupMiddleware は全てのミドルウェアを設定する
func (r *Router) setupMiddleware() {
	// パニック回復ミドルウェア（最初に設定）
	r.engine.Use(errors.RecoveryMiddleware())

	// リクエストIDミドルウェア
	r.engine.Use(logger.RequestIDMiddleware())

	// ログミドルウェア
	r.engine.Use(logger.LoggingMiddleware())

	// エラーハンドリングミドルウェア
	r.engine.Use(errors.ErrorHandlerMiddleware())

	// CORS設定
	r.engine.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:3300"}, // フロントエンドのURL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// レート制限ミドルウェア（認証エンドポイント用）
	r.engine.Use(r.rateLimitMiddleware())
}

// rateLimitMiddleware は認証エンドポイント用のレート制限を実装する
func (r *Router) rateLimitMiddleware() gin.HandlerFunc {
	// 認証エンドポイント用のレート制限（1分間に10回まで）
	limiter := rate.NewLimiter(rate.Every(time.Minute/10), 10)

	return func(c *gin.Context) {
		path := c.Request.URL.Path
		// v1と旧APIの両方の認証エンドポイントにレート制限を適用
		if path == "/api/auth/login" || path == "/api/auth/refresh" ||
		   path == "/api/v1/auth/login" || path == "/api/v1/auth/refresh" {
			if !limiter.Allow() {
				err := errors.NewValidationError("リクエスト制限に達しました。しばらく待ってから再試行してください。")
				err.StatusCode = 429
				c.Error(err)
				c.Abort()
				return
			}
		}
		c.Next()
	}
}

// setupRoutes は全てのルートを設定する
func (r *Router) setupRoutes() {
	// Swagger UIの設定
	r.setupSwaggerRoutes()

	// ヘルスチェックエンドポイント
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "サーバーは正常に動作しています",
		})
	})

	// APIバージョニング: v1エンドポイント
	apiV1 := r.engine.Group("/api/v1")

	// 認証不要の公開ルート
	r.setupPublicRoutes(apiV1)

	// 認証ルート（認証不要）
	r.setupAuthRoutes(apiV1)

	// 認証が必要なルート
	r.setupProtectedRoutes(apiV1)

	// 後方互換性のための旧APIエンドポイント
	r.setupLegacyRoutes()
}

// setupPublicRoutes は認証不要の公開ルートを設定する
func (r *Router) setupPublicRoutes(api *gin.RouterGroup) {
	// 公開トーナメント情報（認証不要）
	publicTournaments := api.Group("/public/tournaments")
	{
		publicTournaments.GET("", r.handlers.TournamentHandler.GetTournaments)                    // GET /public/tournaments
		publicTournaments.GET("/active", r.handlers.TournamentHandler.GetActiveTournaments)       // GET /public/tournaments/active
		publicTournaments.GET("/sport/:sport", r.handlers.TournamentHandler.GetTournamentBySport) // GET /public/tournaments/sport/{sport}
		publicTournaments.GET("/sport/:sport/bracket", r.handlers.TournamentHandler.GetTournamentBracket) // GET /public/tournaments/sport/{sport}/bracket
		publicTournaments.GET("/sport/:sport/progress", r.handlers.TournamentHandler.GetTournamentProgress) // GET /public/tournaments/sport/{sport}/progress
	}

	// 公開試合情報（認証不要）
	publicMatches := api.Group("/public/matches")
	{
		publicMatches.GET("/sport/:sport", r.handlers.MatchHandler.GetMatchesBySport)                // GET /public/matches/sport/{sport}
		publicMatches.GET("/tournament/:tournament_id", r.handlers.MatchHandler.GetMatchesByTournament) // GET /public/matches/tournament/{tournament_id}
		publicMatches.GET("/tournament/:tournament_id/next", r.handlers.MatchHandler.GetNextMatches) // GET /public/matches/tournament/{tournament_id}/next
	}
}

// setupAuthRoutes は認証関連のルートを設定する
func (r *Router) setupAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		// RESTful設計に従った統一パス構造
		auth.POST("/login", r.handlers.AuthHandler.Login)                    // POST /auth/login
		auth.POST("/logout", r.handlers.AuthHandler.Logout)                  // POST /auth/logout
		auth.POST("/refresh", r.handlers.AuthHandler.RefreshToken)           // POST /auth/refresh
		auth.POST("/validate", r.handlers.AuthHandler.ValidateToken)         // POST /auth/validate
		auth.GET("/validate", r.handlers.AuthHandler.ValidateToken)          // GET /auth/validate
		auth.GET("/profile", r.handlers.AuthHandler.GetProfile)              // GET /auth/profile
	}
}

// setupProtectedRoutes は認証が必要なルートを設定する
func (r *Router) setupProtectedRoutes(api *gin.RouterGroup) {
	// 認証ミドルウェアを作成
	authMiddleware := handler.NewAuthMiddleware(r.authService)

	// 認証が必要なルート（一般ユーザー）
	protected := api.Group("/")
	protected.Use(authMiddleware.RequireAuth())

	// 管理者専用ルート
	admin := api.Group("/admin")
	admin.Use(authMiddleware.RequireAuth())
	admin.Use(authMiddleware.RequireAdmin())

	// トーナメント関連ルート
	r.setupTournamentRoutes(protected, admin, authMiddleware)

	// 試合関連ルート
	r.setupMatchRoutes(protected, admin, authMiddleware)
}

// setupTournamentRoutes はトーナメント関連のルートを設定する
func (r *Router) setupTournamentRoutes(protected *gin.RouterGroup, admin *gin.RouterGroup, authMiddleware *handler.AuthMiddleware) {
	// 認証が必要なトーナメント関連ルート（読み取り専用）
	tournaments := protected.Group("/tournaments")
	{
		// 認証済みユーザーがアクセス可能 - RESTful設計に従った統一パス構造
		tournaments.GET("", r.handlers.TournamentHandler.GetTournaments)                    // GET /tournaments
		tournaments.GET("/:id", r.handlers.TournamentHandler.GetTournamentByID)             // GET /tournaments/{id}
		tournaments.GET("/sport/:sport", r.handlers.TournamentHandler.GetTournamentBySport) // GET /tournaments/sport/{sport}
		tournaments.GET("/sport/:sport/bracket", r.handlers.TournamentHandler.GetTournamentBracket) // GET /tournaments/sport/{sport}/bracket
		tournaments.GET("/sport/:sport/progress", r.handlers.TournamentHandler.GetTournamentProgress) // GET /tournaments/sport/{sport}/progress
		tournaments.GET("/active", r.handlers.TournamentHandler.GetActiveTournaments)       // GET /tournaments/active
	}

	// 管理者専用トーナメント関連ルート（作成・更新・削除）
	adminTournaments := admin.Group("/tournaments")
	{
		adminTournaments.POST("", r.handlers.TournamentHandler.CreateTournament)                    // POST /admin/tournaments
		adminTournaments.PUT("/:id", r.handlers.TournamentHandler.UpdateTournament)                // PUT /admin/tournaments/{id}
		adminTournaments.DELETE("/:id", r.handlers.TournamentHandler.DeleteTournament)             // DELETE /admin/tournaments/{id}
		adminTournaments.PUT("/:id/format", r.handlers.TournamentHandler.SwitchTournamentFormat)   // PUT /admin/tournaments/{id}/format
		adminTournaments.PUT("/sport/:sport/complete", r.handlers.TournamentHandler.CompleteTournament) // PUT /admin/tournaments/sport/{sport}/complete
	}
}

// setupMatchRoutes は試合関連のルートを設定する
func (r *Router) setupMatchRoutes(protected *gin.RouterGroup, admin *gin.RouterGroup, authMiddleware *handler.AuthMiddleware) {
	// 認証が必要な試合関連ルート（読み取り専用）
	matches := protected.Group("/matches")
	{
		// 認証済みユーザーがアクセス可能 - RESTful設計に従った統一パス構造
		matches.GET("", r.handlers.MatchHandler.GetMatches)                                    // GET /matches
		matches.GET("/:id", r.handlers.MatchHandler.GetMatch)                                  // GET /matches/{id}
		matches.GET("/sport/:sport", r.handlers.MatchHandler.GetMatchesBySport)                // GET /matches/sport/{sport}
		matches.GET("/tournament/:tournament_id", r.handlers.MatchHandler.GetMatchesByTournament) // GET /matches/tournament/{tournament_id}
		matches.GET("/tournament/:tournament_id/statistics", r.handlers.MatchHandler.GetMatchStatistics) // GET /matches/tournament/{tournament_id}/statistics
		matches.GET("/tournament/:tournament_id/next", r.handlers.MatchHandler.GetNextMatches) // GET /matches/tournament/{tournament_id}/next
	}

	// 管理者専用試合関連ルート（作成・更新・削除）
	adminMatches := admin.Group("/matches")
	{
		adminMatches.POST("", r.handlers.MatchHandler.CreateMatch)                         // POST /admin/matches
		adminMatches.PUT("/:id", r.handlers.MatchHandler.UpdateMatch)                      // PUT /admin/matches/{id}
		adminMatches.DELETE("/:id", r.handlers.MatchHandler.DeleteMatch)                   // DELETE /admin/matches/{id}
		adminMatches.PUT("/:id/result", r.handlers.MatchHandler.SubmitMatchResult)         // PUT /admin/matches/{id}/result
	}
}

// setupSwaggerRoutes はSwagger UIのルートを設定する
func (r *Router) setupSwaggerRoutes() {
	// 開発用のルートエンドポイント（Swagger UIは後で追加）
	r.engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":        "Tournament Backend API",
			"version":        "1.0.0",
			"health":         "/health",
			"api_prefix":     "/api/v1",
			"legacy_prefix":  "/api",
			"deprecation":    "旧API（/api）は廃止予定です。新しいAPI（/api/v1）をご利用ください",
			"migration_guide": "https://api-docs.example.com/migration",
		})
	})
}

// setupLegacyRoutes は後方互換性のための旧APIエンドポイントを設定する
func (r *Router) setupLegacyRoutes() {
	// 旧APIグループ（/api）
	legacyAPI := r.engine.Group("/api")

	// 廃止予定の警告ヘッダーを追加するミドルウェア
	legacyAPI.Use(func(c *gin.Context) {
		c.Header("X-API-Deprecated", "true")
		c.Header("X-API-Deprecation-Message", "このAPIは廃止予定です。/api/v1を使用してください")
		c.Header("X-API-Migration-Guide", "https://api-docs.example.com/migration")
		c.Next()
	})

	// 認証不要の公開ルート（旧形式）
	r.setupLegacyPublicRoutes(legacyAPI)

	// 認証ルート（認証不要）
	r.setupAuthRoutes(legacyAPI)

	// 認証が必要なルート（旧形式）
	r.setupLegacyProtectedRoutes(legacyAPI)
}

// setupLegacyPublicRoutes は旧形式の公開ルートを設定する
func (r *Router) setupLegacyPublicRoutes(api *gin.RouterGroup) {
	// 旧形式のトーナメント情報（認証不要）
	tournaments := api.Group("/tournaments")
	{
		tournaments.GET("", r.handlers.TournamentHandler.GetTournaments)
		tournaments.GET("/:sport", r.handlers.TournamentHandler.GetTournamentBySport)
		tournaments.GET("/:sport/bracket", r.handlers.TournamentHandler.GetTournamentBracket)
	}

	// 旧形式の試合情報（認証不要）
	matches := api.Group("/matches")
	{
		matches.GET("/:sport", r.handlers.MatchHandler.GetMatchesBySport)
	}
}

// setupLegacyProtectedRoutes は旧形式の認証が必要なルートを設定する
func (r *Router) setupLegacyProtectedRoutes(api *gin.RouterGroup) {
	// 認証ミドルウェアを作成
	authMiddleware := handler.NewAuthMiddleware(r.authService)

	// 認証が必要なルート（旧形式）
	protected := api.Group("/")
	protected.Use(authMiddleware.RequireAuth())

	// 旧形式のトーナメント関連ルート
	tournaments := protected.Group("/tournaments")
	tournaments.Use(authMiddleware.RequireAdmin())
	{
		tournaments.POST("", r.handlers.TournamentHandler.CreateTournament)
		tournaments.PUT("/:id", r.handlers.TournamentHandler.UpdateTournament)
		tournaments.PUT("/:id/format", r.handlers.TournamentHandler.SwitchTournamentFormat)
	}

	// 旧形式の試合関連ルート
	matches := protected.Group("/matches")
	matches.Use(authMiddleware.RequireAdmin())
	{
		matches.POST("", r.handlers.MatchHandler.CreateMatch)
		matches.PUT("/:id", r.handlers.MatchHandler.UpdateMatch)
		matches.PUT("/:id/result", r.handlers.MatchHandler.SubmitMatchResult)
	}
}

// GetEngine はGinエンジンを返す
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
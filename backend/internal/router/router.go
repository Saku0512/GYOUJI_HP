package router

import (
	"time"

	"backend/docs"
	"backend/internal/errors"
	"backend/internal/handler"
	"backend/internal/logger"
	"backend/internal/service"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
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
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:5173"}, // フロントエンドのURL
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
		// 認証エンドポイントのみレート制限を適用
		if c.Request.URL.Path == "/api/auth/login" || c.Request.URL.Path == "/api/auth/refresh" {
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

	// APIグループ
	api := r.engine.Group("/api")

	// 認証ルート（認証不要）
	r.setupAuthRoutes(api)

	// 認証が必要なルート
	r.setupProtectedRoutes(api)
}

// setupAuthRoutes は認証関連のルートを設定する
func (r *Router) setupAuthRoutes(api *gin.RouterGroup) {
	auth := api.Group("/auth")
	{
		auth.POST("/login", r.handlers.AuthHandler.Login)
		auth.POST("/refresh", r.handlers.AuthHandler.RefreshToken)
	}
}

// setupProtectedRoutes は認証が必要なルートを設定する
func (r *Router) setupProtectedRoutes(api *gin.RouterGroup) {
	// 認証ミドルウェアを作成
	authMiddleware := handler.NewAuthMiddleware(r.authService)

	// 認証が必要なルート
	protected := api.Group("/")
	protected.Use(authMiddleware.RequireAuth())

	// トーナメント関連ルート
	r.setupTournamentRoutes(protected, authMiddleware)

	// 試合関連ルート
	r.setupMatchRoutes(protected, authMiddleware)
}

// setupTournamentRoutes はトーナメント関連のルートを設定する
func (r *Router) setupTournamentRoutes(protected *gin.RouterGroup, authMiddleware *handler.AuthMiddleware) {
	tournaments := protected.Group("/tournaments")
	{
		// 全ユーザーがアクセス可能
		tournaments.GET("", r.handlers.TournamentHandler.GetTournaments)
		tournaments.GET("/:sport", r.handlers.TournamentHandler.GetTournamentBySport)
		tournaments.GET("/:sport/bracket", r.handlers.TournamentHandler.GetTournamentBracket)

		// 管理者のみアクセス可能
		adminTournaments := tournaments.Group("/")
		adminTournaments.Use(authMiddleware.RequireAdmin())
		{
			adminTournaments.POST("", r.handlers.TournamentHandler.CreateTournament)
			adminTournaments.PUT("/:id", r.handlers.TournamentHandler.UpdateTournament)
			adminTournaments.PUT("/:sport/format", r.handlers.TournamentHandler.SwitchTournamentFormat)
		}
	}
}

// setupMatchRoutes は試合関連のルートを設定する
func (r *Router) setupMatchRoutes(protected *gin.RouterGroup, authMiddleware *handler.AuthMiddleware) {
	matches := protected.Group("/matches")
	{
		// 全ユーザーがアクセス可能
		matches.GET("", r.handlers.MatchHandler.GetMatches)
		matches.GET("/:sport", r.handlers.MatchHandler.GetMatchesBySport)
		matches.GET("/match/:id", r.handlers.MatchHandler.GetMatch)

		// 管理者のみアクセス可能
		adminMatches := matches.Group("/")
		adminMatches.Use(authMiddleware.RequireAdmin())
		{
			adminMatches.POST("", r.handlers.MatchHandler.CreateMatch)
			adminMatches.PUT("/:id", r.handlers.MatchHandler.UpdateMatch)
			adminMatches.PUT("/:id/result", r.handlers.MatchHandler.SubmitMatchResult)
		}
	}
}

// setupSwaggerRoutes はSwagger UIのルートを設定する
func (r *Router) setupSwaggerRoutes() {
	// Swagger UIエンドポイント
	r.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	
	// APIドキュメントのルートリダイレクト
	r.engine.GET("/docs", func(c *gin.Context) {
		c.Redirect(302, "/swagger/index.html")
	})
	
	// ルートからドキュメントへのリダイレクト（開発用）
	r.engine.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message":     "Tournament Backend API",
			"version":     "1.0.0",
			"docs":        "/swagger/index.html",
			"health":      "/health",
			"api_prefix":  "/api",
		})
	})
}

// GetEngine はGinエンジンを返す
func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}
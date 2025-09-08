# 設計書

## 概要

API統合修正システムは、既存のGo言語バックエンドAPIとSvelteフロントエンド間の統合問題を解決するための包括的な設計です。現在のシステムでは、エンドポイントの不整合、データ形式の違い、エラーハンドリングの不統一などの問題が存在しており、これらを体系的に解決し、一貫性のあるAPI設計を実現します。

## アーキテクチャ

### 現在の問題点分析

```
現在の状況:
┌─────────────────┐    不整合なAPI    ┌─────────────────┐
│   Svelte        │ ←─────────────→ │   Go Backend    │
│   Frontend      │                 │   (Gin + MySQL) │
│                 │                 │                 │
│ ┌─────────────┐ │                 │ ┌─────────────┐ │
│ │ API Client  │ │  ❌ 不一致      │ │  Handlers   │ │
│ │ - auth.js   │ │  ❌ 重複       │ │  - auth     │ │
│ │ - tournament│ │  ❌ 形式違い    │ │  - tournament│ │
│ │ - matches   │ │                 │ │  - match    │ │
│ └─────────────┘ │                 │ └─────────────┘ │
└─────────────────┘                 └─────────────────┘
```

### 修正後のアーキテクチャ

```
修正後:
┌─────────────────┐    統一されたAPI   ┌─────────────────┐
│   Svelte        │ ←─────────────→ │   Go Backend    │
│   Frontend      │                 │   (Gin + MySQL) │
│                 │                 │                 │
│ ┌─────────────┐ │                 │ ┌─────────────┐ │
│ │ Unified API │ │  ✅ 一致       │ │ Standardized│ │
│ │ Client      │ │  ✅ 統一       │ │ Handlers    │ │
│ │ - Standard  │ │  ✅ 形式統一    │ │ - Consistent│ │
│ │ - Validated │ │                 │ │ - Validated │ │
│ └─────────────┘ │                 │ └─────────────┘ │
└─────────────────┘                 └─────────────────┘
```

## 問題点と解決策

### 1. エンドポイント不整合の問題

**現在の問題:**
- フロントエンド: `/tournaments/{sport}/bracket`
- バックエンド: `/api/tournaments/{sport}/bracket`
- 一部のエンドポイントでパス不一致

**解決策:**
```
統一されたエンドポイント構造:
/api/v1/{resource}/{identifier?}/{action?}

例:
- GET /api/v1/auth/login → POST /api/v1/auth/login
- GET /api/v1/tournaments → 全トーナメント取得
- GET /api/v1/tournaments/{sport} → スポーツ別取得
- GET /api/v1/tournaments/{sport}/bracket → ブラケット取得
- GET /api/v1/matches → 全試合取得
- GET /api/v1/matches/{sport} → スポーツ別試合取得
- PUT /api/v1/matches/{id}/result → 試合結果更新
```

### 2. レスポンス形式の不統一

**現在の問題:**
- 成功時: `{success: true, data: {...}, message: "..."}`
- エラー時: `{error: "...", message: "...", code: 400}`
- 一部で形式が異なる

**解決策:**
```typescript
// 統一されたレスポンス形式
interface APIResponse<T> {
  success: boolean;
  data?: T;
  error?: string;
  message: string;
  code: number;
  timestamp: string;
  request_id?: string;
}

// 成功レスポンス例
{
  "success": true,
  "data": {...},
  "message": "操作が成功しました",
  "code": 200,
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456"
}

// エラーレスポンス例
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "入力データが無効です",
  "code": 400,
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456"
}
```

### 3. データ型不整合の問題

**現在の問題:**
- 日時形式: バックエンド `2006-01-02T15:04:05Z` vs フロントエンド期待値
- ID型: 数値 vs 文字列の混在
- null値の扱いが不統一

**解決策:**
```typescript
// 統一されたデータ型定義
interface Tournament {
  id: number;                    // 常に数値
  sport: SportType;              // 列挙型
  format: string;
  status: TournamentStatus;      // 列挙型
  created_at: string;            // ISO 8601形式
  updated_at: string;            // ISO 8601形式
}

interface Match {
  id: number;
  tournament_id: number;
  round: string;
  team1: string;
  team2: string;
  score1: number | null;         // 明示的なnull許可
  score2: number | null;
  winner: string | null;
  status: MatchStatus;           // 列挙型
  scheduled_at: string;          // ISO 8601形式
  completed_at: string | null;   // ISO 8601形式またはnull
  created_at: string;
  updated_at: string;
}

// 列挙型定義
type SportType = 'volleyball' | 'table_tennis' | 'soccer';
type TournamentStatus = 'registration' | 'active' | 'completed' | 'cancelled';
type MatchStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled';
```

## コンポーネントとインターフェース

### 1. 統一APIクライアント

**新しいAPIClient設計:**
```typescript
class UnifiedAPIClient {
  private baseURL: string;
  private version: string = 'v1';
  private token: string | null = null;

  // 統一されたリクエストメソッド
  async request<T>(
    method: HTTPMethod,
    endpoint: string,
    data?: any,
    options?: RequestOptions
  ): Promise<APIResponse<T>>;

  // リソース別メソッド
  auth: AuthAPI;
  tournaments: TournamentAPI;
  matches: MatchAPI;
}

interface AuthAPI {
  login(credentials: LoginRequest): Promise<APIResponse<AuthResponse>>;
  logout(): Promise<APIResponse<void>>;
  refresh(token: string): Promise<APIResponse<AuthResponse>>;
  validate(token?: string): Promise<APIResponse<UserInfo>>;
}

interface TournamentAPI {
  getAll(): Promise<APIResponse<Tournament[]>>;
  getBySport(sport: SportType): Promise<APIResponse<Tournament>>;
  getBracket(sport: SportType): Promise<APIResponse<Bracket>>;
  updateFormat(sport: SportType, format: string): Promise<APIResponse<Tournament>>;
  create(data: CreateTournamentRequest): Promise<APIResponse<Tournament>>;
  update(id: number, data: UpdateTournamentRequest): Promise<APIResponse<Tournament>>;
  delete(id: number): Promise<APIResponse<void>>;
}

interface MatchAPI {
  getAll(filters?: MatchFilters): Promise<APIResponse<Match[]>>;
  getBySport(sport: SportType, filters?: MatchFilters): Promise<APIResponse<Match[]>>;
  getById(id: number): Promise<APIResponse<Match>>;
  create(data: CreateMatchRequest): Promise<APIResponse<Match>>;
  update(id: number, data: UpdateMatchRequest): Promise<APIResponse<Match>>;
  updateResult(id: number, result: MatchResult): Promise<APIResponse<Match>>;
  delete(id: number): Promise<APIResponse<void>>;
}
```

### 2. バックエンドハンドラー標準化

**統一されたハンドラー構造:**
```go
// 統一されたレスポンス構造体
type APIResponse struct {
    Success   bool        `json:"success"`
    Data      interface{} `json:"data,omitempty"`
    Error     string      `json:"error,omitempty"`
    Message   string      `json:"message"`
    Code      int         `json:"code"`
    Timestamp string      `json:"timestamp"`
    RequestID string      `json:"request_id,omitempty"`
}

// 統一されたエラーハンドリング
func (h *BaseHandler) SendSuccess(c *gin.Context, data interface{}, message string) {
    response := APIResponse{
        Success:   true,
        Data:      data,
        Message:   message,
        Code:      http.StatusOK,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: c.GetString("request_id"),
    }
    c.JSON(http.StatusOK, response)
}

func (h *BaseHandler) SendError(c *gin.Context, statusCode int, errorCode string, message string) {
    response := APIResponse{
        Success:   false,
        Error:     errorCode,
        Message:   message,
        Code:      statusCode,
        Timestamp: time.Now().UTC().Format(time.RFC3339),
        RequestID: c.GetString("request_id"),
    }
    c.JSON(statusCode, response)
}
```

### 3. データ検証とシリアライゼーション

**統一されたバリデーション:**
```go
// リクエストバリデーション
type LoginRequest struct {
    Username string `json:"username" binding:"required,min=1,max=50" validate:"alphanum"`
    Password string `json:"password" binding:"required,min=8,max=100"`
}

type CreateTournamentRequest struct {
    Sport  string `json:"sport" binding:"required,oneof=volleyball table_tennis soccer"`
    Format string `json:"format" binding:"required,min=1,max=50"`
}

type MatchResultRequest struct {
    Score1 int    `json:"score1" binding:"required,min=0"`
    Score2 int    `json:"score2" binding:"required,min=0"`
    Winner string `json:"winner" binding:"required,min=1,max=100"`
}

// レスポンス用構造体（統一された形式）
type TournamentResponse struct {
    ID        int       `json:"id"`
    Sport     string    `json:"sport"`
    Format    string    `json:"format"`
    Status    string    `json:"status"`
    CreatedAt string    `json:"created_at"`
    UpdatedAt string    `json:"updated_at"`
}

type MatchResponse struct {
    ID           int     `json:"id"`
    TournamentID int     `json:"tournament_id"`
    Round        string  `json:"round"`
    Team1        string  `json:"team1"`
    Team2        string  `json:"team2"`
    Score1       *int    `json:"score1"`
    Score2       *int    `json:"score2"`
    Winner       *string `json:"winner"`
    Status       string  `json:"status"`
    ScheduledAt  string  `json:"scheduled_at"`
    CompletedAt  *string `json:"completed_at"`
    CreatedAt    string  `json:"created_at"`
    UpdatedAt    string  `json:"updated_at"`
}
```

## エラーハンドリングの標準化

### 1. エラーコード体系

```typescript
// 統一されたエラーコード
enum ErrorCode {
  // 認証関連 (AUTH_*)
  AUTH_INVALID_CREDENTIALS = 'AUTH_INVALID_CREDENTIALS',
  AUTH_TOKEN_EXPIRED = 'AUTH_TOKEN_EXPIRED',
  AUTH_TOKEN_INVALID = 'AUTH_TOKEN_INVALID',
  AUTH_UNAUTHORIZED = 'AUTH_UNAUTHORIZED',
  AUTH_FORBIDDEN = 'AUTH_FORBIDDEN',

  // バリデーション関連 (VALIDATION_*)
  VALIDATION_REQUIRED_FIELD = 'VALIDATION_REQUIRED_FIELD',
  VALIDATION_INVALID_FORMAT = 'VALIDATION_INVALID_FORMAT',
  VALIDATION_OUT_OF_RANGE = 'VALIDATION_OUT_OF_RANGE',
  VALIDATION_DUPLICATE_VALUE = 'VALIDATION_DUPLICATE_VALUE',

  // リソース関連 (RESOURCE_*)
  RESOURCE_NOT_FOUND = 'RESOURCE_NOT_FOUND',
  RESOURCE_ALREADY_EXISTS = 'RESOURCE_ALREADY_EXISTS',
  RESOURCE_CONFLICT = 'RESOURCE_CONFLICT',

  // ビジネスロジック関連 (BUSINESS_*)
  BUSINESS_TOURNAMENT_COMPLETED = 'BUSINESS_TOURNAMENT_COMPLETED',
  BUSINESS_MATCH_ALREADY_COMPLETED = 'BUSINESS_MATCH_ALREADY_COMPLETED',
  BUSINESS_INVALID_MATCH_RESULT = 'BUSINESS_INVALID_MATCH_RESULT',

  // システム関連 (SYSTEM_*)
  SYSTEM_DATABASE_ERROR = 'SYSTEM_DATABASE_ERROR',
  SYSTEM_NETWORK_ERROR = 'SYSTEM_NETWORK_ERROR',
  SYSTEM_TIMEOUT = 'SYSTEM_TIMEOUT',
  SYSTEM_UNKNOWN_ERROR = 'SYSTEM_UNKNOWN_ERROR'
}
```

### 2. エラーハンドリングミドルウェア

```go
// 統一されたエラーハンドリングミドルウェア
func ErrorHandlerMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Next()

        if len(c.Errors) > 0 {
            err := c.Errors.Last()
            
            var apiError *APIError
            if errors.As(err.Err, &apiError) {
                c.JSON(apiError.StatusCode, APIResponse{
                    Success:   false,
                    Error:     apiError.Code,
                    Message:   apiError.Message,
                    Code:      apiError.StatusCode,
                    Timestamp: time.Now().UTC().Format(time.RFC3339),
                    RequestID: c.GetString("request_id"),
                })
            } else {
                // 未知のエラー
                c.JSON(http.StatusInternalServerError, APIResponse{
                    Success:   false,
                    Error:     "SYSTEM_UNKNOWN_ERROR",
                    Message:   "予期しないエラーが発生しました",
                    Code:      http.StatusInternalServerError,
                    Timestamp: time.Now().UTC().Format(time.RFC3339),
                    RequestID: c.GetString("request_id"),
                })
            }
        }
    }
}

// カスタムエラー型
type APIError struct {
    Code       string `json:"code"`
    Message    string `json:"message"`
    StatusCode int    `json:"status_code"`
    Details    map[string]interface{} `json:"details,omitempty"`
}

func (e *APIError) Error() string {
    return e.Message
}
```

## 認証とセキュリティの統合

### 1. JWT統一仕様

```go
// 統一されたJWTクレーム
type JWTClaims struct {
    UserID   int    `json:"user_id"`
    Username string `json:"username"`
    Role     string `json:"role"`
    jwt.RegisteredClaims
}

// 統一されたトークン生成
func GenerateToken(userID int, username string, role string) (string, error) {
    claims := JWTClaims{
        UserID:   userID,
        Username: username,
        Role:     role,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
            IssuedAt:  jwt.NewNumericDate(time.Now()),
            NotBefore: jwt.NewNumericDate(time.Now()),
            Issuer:    "tournament-api",
            Subject:   strconv.Itoa(userID),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(os.Getenv("JWT_SECRET")))
}
```

### 2. 認証ミドルウェア統一

```go
// 統一された認証ミドルウェア
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.Error(&APIError{
                Code:       "AUTH_UNAUTHORIZED",
                Message:    "認証が必要です",
                StatusCode: http.StatusUnauthorized,
            })
            c.Abort()
            return
        }

        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        claims, err := ValidateToken(tokenString)
        if err != nil {
            c.Error(&APIError{
                Code:       "AUTH_TOKEN_INVALID",
                Message:    "無効または期限切れのトークンです",
                StatusCode: http.StatusUnauthorized,
            })
            c.Abort()
            return
        }

        // コンテキストにユーザー情報を設定
        c.Set("user_id", claims.UserID)
        c.Set("username", claims.Username)
        c.Set("role", claims.Role)
        c.Next()
    }
}
```

## リアルタイム更新の統合

### 1. WebSocket統合設計

```go
// WebSocket接続管理
type ConnectionManager struct {
    connections map[string]*websocket.Conn
    broadcast   chan []byte
    register    chan *Client
    unregister  chan *Client
    mutex       sync.RWMutex
}

type Client struct {
    ID         string
    Connection *websocket.Conn
    UserID     int
    Sports     []string // 購読しているスポーツ
}

// リアルタイム更新通知
type UpdateNotification struct {
    Type      string      `json:"type"`      // "tournament_update", "match_result", etc.
    Sport     string      `json:"sport"`
    Data      interface{} `json:"data"`
    Timestamp string      `json:"timestamp"`
}
```

### 2. ポーリング代替機能

```typescript
// フロントエンド: ポーリング機能
class PollingManager {
  private intervals: Map<string, NodeJS.Timeout> = new Map();
  
  startPolling(
    key: string,
    callback: () => Promise<void>,
    intervalMs: number = 30000
  ): void {
    this.stopPolling(key);
    
    const interval = setInterval(async () => {
      try {
        await callback();
      } catch (error) {
        console.error(`Polling error for ${key}:`, error);
      }
    }, intervalMs);
    
    this.intervals.set(key, interval);
  }
  
  stopPolling(key: string): void {
    const interval = this.intervals.get(key);
    if (interval) {
      clearInterval(interval);
      this.intervals.delete(key);
    }
  }
}
```

## APIドキュメントとテストの統合

### 1. OpenAPI仕様統一

```yaml
# openapi.yaml
openapi: 3.0.3
info:
  title: Tournament Management API
  version: 1.0.0
  description: 統一されたトーナメント管理API

servers:
  - url: http://localhost:8080/api/v1
    description: 開発環境
  - url: https://api.tournament.example.com/api/v1
    description: 本番環境

components:
  schemas:
    APIResponse:
      type: object
      properties:
        success:
          type: boolean
        data:
          type: object
        error:
          type: string
        message:
          type: string
        code:
          type: integer
        timestamp:
          type: string
          format: date-time
        request_id:
          type: string

    Tournament:
      type: object
      properties:
        id:
          type: integer
        sport:
          type: string
          enum: [volleyball, table_tennis, soccer]
        format:
          type: string
        status:
          type: string
          enum: [registration, active, completed, cancelled]
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time

  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT

paths:
  /auth/login:
    post:
      summary: ユーザーログイン
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                username:
                  type: string
                password:
                  type: string
      responses:
        '200':
          description: ログイン成功
          content:
            application/json:
              schema:
                allOf:
                  - $ref: '#/components/schemas/APIResponse'
                  - type: object
                    properties:
                      data:
                        type: object
                        properties:
                          token:
                            type: string
                          user:
                            type: object
```

### 2. 契約テスト

```typescript
// 契約テスト例
describe('API Contract Tests', () => {
  test('POST /api/v1/auth/login should return valid response', async () => {
    const response = await apiClient.auth.login({
      username: 'admin',
      password: 'password'
    });

    expect(response).toMatchSchema({
      success: expect.any(Boolean),
      data: {
        token: expect.any(String),
        user: {
          id: expect.any(Number),
          username: expect.any(String),
          role: expect.any(String)
        }
      },
      message: expect.any(String),
      code: 200,
      timestamp: expect.stringMatching(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/)
    });
  });

  test('GET /api/v1/tournaments should return tournament list', async () => {
    const response = await apiClient.tournaments.getAll();

    expect(response.success).toBe(true);
    expect(Array.isArray(response.data)).toBe(true);
    
    if (response.data.length > 0) {
      expect(response.data[0]).toMatchSchema({
        id: expect.any(Number),
        sport: expect.stringMatching(/^(volleyball|table_tennis|soccer)$/),
        format: expect.any(String),
        status: expect.stringMatching(/^(registration|active|completed|cancelled)$/),
        created_at: expect.stringMatching(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/),
        updated_at: expect.stringMatching(/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$/)
      });
    }
  });
});
```

## パフォーマンス最適化

### 1. キャッシュ戦略

```go
// Redis統合キャッシュ
type CacheManager struct {
    client *redis.Client
}

func (c *CacheManager) GetTournament(sport string) (*models.Tournament, error) {
    key := fmt.Sprintf("tournament:%s", sport)
    
    // キャッシュから取得を試行
    cached, err := c.client.Get(context.Background(), key).Result()
    if err == nil {
        var tournament models.Tournament
        if err := json.Unmarshal([]byte(cached), &tournament); err == nil {
            return &tournament, nil
        }
    }
    
    // キャッシュミスの場合はDBから取得
    tournament, err := c.getTournamentFromDB(sport)
    if err != nil {
        return nil, err
    }
    
    // キャッシュに保存（5分間）
    data, _ := json.Marshal(tournament)
    c.client.Set(context.Background(), key, data, 5*time.Minute)
    
    return tournament, nil
}
```

### 2. ページネーション統一

```go
// 統一されたページネーション
type PaginationRequest struct {
    Page     int `form:"page" binding:"min=1"`
    PageSize int `form:"page_size" binding:"min=1,max=100"`
}

type PaginationResponse struct {
    Page       int `json:"page"`
    PageSize   int `json:"page_size"`
    TotalItems int `json:"total_items"`
    TotalPages int `json:"total_pages"`
    HasNext    bool `json:"has_next"`
    HasPrev    bool `json:"has_prev"`
}

type PaginatedAPIResponse struct {
    Success    bool                `json:"success"`
    Data       interface{}         `json:"data"`
    Pagination *PaginationResponse `json:"pagination,omitempty"`
    Message    string              `json:"message"`
    Code       int                 `json:"code"`
    Timestamp  string              `json:"timestamp"`
}
```

## セキュリティ考慮事項

### 1. CORS設定統一

```go
// 統一されたCORS設定
func CORSMiddleware() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins: []string{
            "http://localhost:3000",
            "http://localhost:5173",
            "https://tournament.example.com",
        },
        AllowMethods: []string{
            "GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH",
        },
        AllowHeaders: []string{
            "Origin", "Content-Type", "Accept", "Authorization",
            "X-Requested-With", "X-Request-ID",
        },
        ExposeHeaders: []string{
            "Content-Length", "X-Request-ID",
        },
        AllowCredentials: true,
        MaxAge:          12 * time.Hour,
    })
}
```

### 2. レート制限統一

```go
// 統一されたレート制限
func RateLimitMiddleware() gin.HandlerFunc {
    // エンドポイント別の制限設定
    limiters := map[string]*rate.Limiter{
        "auth":       rate.NewLimiter(rate.Every(time.Minute/10), 10), // 10回/分
        "tournament": rate.NewLimiter(rate.Every(time.Second), 30),     // 30回/秒
        "match":      rate.NewLimiter(rate.Every(time.Second), 20),     // 20回/秒
    }

    return func(c *gin.Context) {
        path := c.Request.URL.Path
        var limiter *rate.Limiter

        switch {
        case strings.HasPrefix(path, "/api/v1/auth"):
            limiter = limiters["auth"]
        case strings.HasPrefix(path, "/api/v1/tournaments"):
            limiter = limiters["tournament"]
        case strings.HasPrefix(path, "/api/v1/matches"):
            limiter = limiters["match"]
        default:
            c.Next()
            return
        }

        if !limiter.Allow() {
            c.Error(&APIError{
                Code:       "RATE_LIMIT_EXCEEDED",
                Message:    "リクエスト制限に達しました",
                StatusCode: http.StatusTooManyRequests,
            })
            c.Abort()
            return
        }

        c.Next()
    }
}
```

## 移行戦略

### 1. 段階的移行計画

```
Phase 1: レスポンス形式統一 (1週間)
- バックエンドのレスポンス形式を統一
- エラーハンドリングの標準化
- 基本的なバリデーション統一

Phase 2: エンドポイント整理 (1週間)
- 重複エンドポイントの統合
- パス構造の統一
- バージョニングの導入

Phase 3: データ型統一 (1週間)
- 日時形式の統一
- ID型の統一
- null値処理の統一

Phase 4: フロントエンド統合 (1週間)
- APIクライアントの統一
- エラーハンドリングの統合
- 型定義の統一

Phase 5: テストと最適化 (1週間)
- 契約テストの実装
- パフォーマンステスト
- セキュリティテスト
```

### 2. 後方互換性の維持

```go
// 旧APIとの互換性維持
func LegacyCompatibilityMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 旧形式のレスポンスが必要な場合
        if c.GetHeader("API-Version") == "legacy" {
            c.Set("legacy_mode", true)
        }
        c.Next()
    }
}

// レスポンス送信時に形式を調整
func SendResponse(c *gin.Context, data interface{}, message string) {
    if c.GetBool("legacy_mode") {
        // 旧形式でレスポンス
        c.JSON(http.StatusOK, gin.H{
            "data":    data,
            "message": message,
        })
    } else {
        // 新形式でレスポンス
        c.JSON(http.StatusOK, APIResponse{
            Success:   true,
            Data:      data,
            Message:   message,
            Code:      http.StatusOK,
            Timestamp: time.Now().UTC().Format(time.RFC3339),
        })
    }
}
```

この設計により、フロントエンドとバックエンド間のAPI統合問題を体系的に解決し、一貫性のある、保守しやすいシステムを構築できます。
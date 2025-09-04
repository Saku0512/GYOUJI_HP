# 設計書

## 概要

トーナメントバックエンドは、バレーボール、卓球、8人制サッカーのスポーツトーナメントデータを管理するGinフレームワークで構築されたGo言語ベースのREST APIです。システムはハンドラー、サービス、リポジトリ間の明確な分離を持つクリーンアーキテクチャパターンに従います。単一の管理者ユーザーのJWTベース認証を提供し、データ永続化のためにMySQLと統合します。

## アーキテクチャ

システムは階層アーキテクチャパターンに従います：

```
┌─────────────────┐
│   HTTP Layer    │  ← Ginハンドラー、ミドルウェア、ルーティング
├─────────────────┤
│  Service Layer  │  ← ビジネスロジック、トーナメントルール
├─────────────────┤
│Repository Layer │  ← データアクセス、MySQL操作
├─────────────────┤
│  Database Layer │  ← MySQLデータベース
└─────────────────┘
```

### 主要なアーキテクチャ決定

1. **クリーンアーキテクチャ**: HTTP処理、ビジネスロジック、データアクセスの明確な層分離による関心の分離
2. **リポジトリパターン**: テスト可能性と保守性を可能にするデータアクセスの抽象化
3. **JWT認証**: APIベースアーキテクチャに適したステートレス認証
4. **Ginフレームワーク**: 優れたミドルウェアサポートを持つ軽量で高速なHTTPフレームワーク
5. **MySQL統合**: ACID特性を持つトーナメントデータ用の信頼性の高いリレーショナルデータベース

## コンポーネントとインターフェース

### 1. HTTP層（ハンドラー）

**AuthHandler**
- ログインリクエストとJWTトークン生成を処理
- 設定された管理者ユーザーに対して認証情報を検証
- 認証成功時にJWTトークンを返す

**TournamentHandler**
- トーナメント関連のHTTPリクエストを管理
- トーナメントと試合のCRUD操作を処理
- 異なるスポーツとトーナメント形式のエンドポイントを提供

**MatchHandler**
- 試合結果提出を処理
- 試合結果に基づいてトーナメントブラケットを更新
- 試合データを検証し、トーナメントルールを強制

### 2. サービス層

**AuthService**
```go
type AuthService interface {
    Login(username, password string) (string, error)
    ValidateToken(token string) (*Claims, error)
    GenerateToken(userID string) (string, error)
}
```

**TournamentService**
```go
type TournamentService interface {
    GetTournament(sport string) (*Tournament, error)
    UpdateMatch(matchID int, result MatchResult) error
    GetTournamentBracket(sport string) (*Bracket, error)
    InitializeTournament(sport string) error
}
```

**MatchService**
```go
type MatchService interface {
    CreateMatch(match Match) error
    UpdateMatchResult(matchID int, result MatchResult) error
    GetMatchesBySport(sport string) ([]Match, error)
    AdvanceWinner(matchID int) error
}
```

### 3. リポジトリ層

**UserRepository**
```go
type UserRepository interface {
    GetAdminUser() (*User, error)
    ValidateCredentials(username, password string) bool
}
```

**TournamentRepository**
```go
type TournamentRepository interface {
    GetTournament(sport string) (*Tournament, error)
    CreateTournament(tournament Tournament) error
    UpdateTournament(tournament Tournament) error
    GetTournamentBracket(sport string) (*Bracket, error)
}
```

**MatchRepository**
```go
type MatchRepository interface {
    CreateMatch(match Match) error
    UpdateMatch(match Match) error
    GetMatch(id int) (*Match, error)
    GetMatchesBySport(sport string) ([]Match, error)
    GetMatchesByTournament(tournamentID int) ([]Match, error)
}
```

## データモデル

### コアモデル

**User**
```go
type User struct {
    ID       int    `json:"id" db:"id"`
    Username string `json:"username" db:"username"`
    Password string `json:"-" db:"password"` // bcryptハッシュ化
    Role     string `json:"role" db:"role"`
}
```

**Tournament**
```go
type Tournament struct {
    ID          int       `json:"id" db:"id"`
    Sport       string    `json:"sport" db:"sport"`
    Format      string    `json:"format" db:"format"` // 卓球の場合"standard", "rainy"
    Status      string    `json:"status" db:"status"` // "active", "completed"
    CreatedAt   time.Time `json:"created_at" db:"created_at"`
    UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
```

**Match**
```go
type Match struct {
    ID           int       `json:"id" db:"id"`
    TournamentID int       `json:"tournament_id" db:"tournament_id"`
    Round        string    `json:"round" db:"round"` // "1st_round", "quarterfinal"など
    Team1        string    `json:"team1" db:"team1"`
    Team2        string    `json:"team2" db:"team2"`
    Score1       *int      `json:"score1" db:"score1"` // 試合が行われるまでnull
    Score2       *int      `json:"score2" db:"score2"`
    Winner       *string   `json:"winner" db:"winner"`
    Status       string    `json:"status" db:"status"` // "pending", "completed"
    ScheduledAt  time.Time `json:"scheduled_at" db:"scheduled_at"`
    CompletedAt  *time.Time `json:"completed_at" db:"completed_at"`
}
```

**Bracket**
```go
type Bracket struct {
    TournamentID int     `json:"tournament_id"`
    Sport        string  `json:"sport"`
    Format       string  `json:"format"`
    Rounds       []Round `json:"rounds"`
}

type Round struct {
    Name    string  `json:"name"`
    Matches []Match `json:"matches"`
}
```

### データベーススキーマ

**usersテーブル**
```sql
CREATE TABLE users (
    id INT PRIMARY KEY AUTO_INCREMENT,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'admin',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**tournamentsテーブル**
```sql
CREATE TABLE tournaments (
    id INT PRIMARY KEY AUTO_INCREMENT,
    sport ENUM('volleyball', 'table_tennis', 'soccer') NOT NULL,
    format VARCHAR(20) DEFAULT 'standard',
    status ENUM('active', 'completed') DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
);
```

**matchesテーブル**
```sql
CREATE TABLE matches (
    id INT PRIMARY KEY AUTO_INCREMENT,
    tournament_id INT NOT NULL,
    round VARCHAR(50) NOT NULL,
    team1 VARCHAR(100) NOT NULL,
    team2 VARCHAR(100) NOT NULL,
    score1 INT NULL,
    score2 INT NULL,
    winner VARCHAR(100) NULL,
    status ENUM('pending', 'completed') DEFAULT 'pending',
    scheduled_at TIMESTAMP NOT NULL,
    completed_at TIMESTAMP NULL,
    FOREIGN KEY (tournament_id) REFERENCES tournaments(id)
);
```

## APIエンドポイント

### 認証エンドポイント
- `POST /api/auth/login` - 認証情報による管理者ログイン
- `POST /api/auth/refresh` - JWTトークンのリフレッシュ

### トーナメントエンドポイント
- `GET /api/tournaments` - 全トーナメントの取得
- `GET /api/tournaments/{sport}` - スポーツ別トーナメントの取得
- `GET /api/tournaments/{sport}/bracket` - トーナメントブラケットの取得
- `POST /api/tournaments` - 新規トーナメント作成（管理者のみ）
- `PUT /api/tournaments/{id}` - トーナメント更新（管理者のみ）

### 試合エンドポイント
- `GET /api/matches` - 全試合の取得
- `GET /api/matches/{sport}` - スポーツ別試合の取得
- `POST /api/matches` - 新規試合作成（管理者のみ）
- `PUT /api/matches/{id}` - 試合結果更新（管理者のみ）
- `GET /api/matches/{id}` - 特定試合の取得

### トーナメント形式エンドポイント
- `PUT /api/tournaments/{sport}/format` - トーナメント形式変更（卓球の天候条件用）

## エラーハンドリング

### エラーレスポンス形式
```go
type ErrorResponse struct {
    Error   string `json:"error"`
    Message string `json:"message"`
    Code    int    `json:"code"`
}
```

### エラーカテゴリ
1. **認証エラー** (401): 無効な認証情報、期限切れトークン
2. **認可エラー** (403): 権限不足
3. **検証エラー** (400): 無効な入力データ、不正なリクエスト
4. **未発見エラー** (404): リソースが見つからない
5. **データベースエラー** (500): データベース接続問題、クエリ失敗
6. **ビジネスロジックエラー** (422): トーナメントルール違反、無効な試合状態

### エラーハンドリング戦略
- 一貫したエラーレスポンス形式のためのミドルウェア使用
- 適切な重要度レベルでのすべてのエラーログ記録
- 詳細な技術情報をログに記録しながらユーザーフレンドリーなエラーメッセージを返す
- データベース接続のサーキットブレーカーパターンの実装

## テスト戦略

### 単体テスト
- すべてのサービス層ビジネスロジックのテスト
- リポジトリ依存関係のモック化
- JWTトークン生成と検証のテスト
- トーナメントルール強制のテスト

### 統合テスト
- 実際のデータベースでのAPIエンドポイントテスト
- 認証ミドルウェアのテスト
- トーナメントブラケット生成と更新のテスト
- 試合結果処理のテスト

### テストデータベース
- 統合テスト用の別テストデータベース使用
- 一貫したテストデータのためのデータベースシーディング実装
- 各テスト実行後のテストデータクリーンアップ

### テストツール
- Goの組み込みテストパッケージ
- アサーションとモック化のためのTestify
- HTTPハンドラーテスト用のGinテストコンテキスト
- テストデータセットアップ用のデータベースマイグレーション

## セキュリティ考慮事項

### 認証セキュリティ
- パスワードハッシュ化にbcryptを使用
- 適切な有効期限を持つJWTの実装
- セキュアなJWT署名キーの使用（環境変数に保存）
- トークンリフレッシュメカニズムの実装

### APIセキュリティ
- フロントエンド統合のためのCORS設定
- 認証エンドポイントでのレート制限
- 入力検証とサニタイゼーション
- パラメータ化クエリによるSQLインジェクション防止

### データベースセキュリティ
- 適切な制限を持つコネクションプールの使用
- 最小限の必要権限を持つデータベースユーザーの実装
- データベースと依存関係の定期的なセキュリティ更新
- データベース認証情報の環境ベース設定

## パフォーマンス考慮事項

### データベース最適化
- 頻繁にクエリされるカラムのインデックス作成（sport、tournament_id、status）
- データベース接続のコネクションプール使用
- ブラケット生成のクエリ最適化実装
- 高トラフィックシナリオでの読み取りレプリカの検討

### キャッシュ戦略
- 頻繁にアクセスされるデータのトーナメントブラケットのメモリキャッシュ
- 試合結果更新時のキャッシュ無効化実装
- 水平スケーリング時の分散キャッシュ用Redis使用

### APIパフォーマンス
- 大きな結果セットのページネーション実装
- 適切なHTTPステータスコードとヘッダーの使用
- 大きなトーナメントデータのレスポンス圧縮
- リクエストタイムアウトハンドリングの実装
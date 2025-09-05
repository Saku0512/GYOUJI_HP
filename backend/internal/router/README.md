# Router Package

このパッケージは、トーナメントバックエンドAPIのHTTPルーティングとミドルウェアを管理します。

## 機能

### ミドルウェア
- **CORS**: フロントエンドとの通信を許可するCORS設定
- **ログ**: リクエスト/レスポンスのログ記録
- **エラーハンドリング**: 統一されたエラーレスポンス形式
- **レート制限**: 認証エンドポイントのレート制限（1分間に10回まで）
- **認証**: JWT認証ミドルウェア
- **認可**: 管理者権限チェック

### エンドポイント

#### 認証関連
- `POST /api/auth/login` - ログイン
- `POST /api/auth/refresh` - トークンリフレッシュ

#### トーナメント関連
- `GET /api/tournaments` - 全トーナメント取得
- `GET /api/tournaments/{sport}` - スポーツ別トーナメント取得
- `GET /api/tournaments/{sport}/bracket` - トーナメントブラケット取得
- `POST /api/tournaments` - トーナメント作成（管理者のみ）
- `PUT /api/tournaments/{id}` - トーナメント更新（管理者のみ）
- `PUT /api/tournaments/{sport}/format` - トーナメント形式変更（管理者のみ）

#### 試合関連
- `GET /api/matches` - 全試合取得
- `GET /api/matches/{sport}` - スポーツ別試合取得
- `GET /api/matches/match/{id}` - 特定試合取得
- `POST /api/matches` - 試合作成（管理者のみ）
- `PUT /api/matches/{id}` - 試合更新（管理者のみ）
- `PUT /api/matches/{id}/result` - 試合結果提出（管理者のみ）

#### その他
- `GET /health` - ヘルスチェック

## 使用方法

```go
// サービスを初期化
authService := service.NewAuthService(userRepo, cfg)
tournamentService := service.NewTournamentService(tournamentRepo, matchRepo)
matchService := service.NewMatchService(matchRepo, tournamentRepo)

// ルーターを作成
router := router.NewRouter(authService, tournamentService, matchService)

// HTTPサーバーで使用
server := &http.Server{
    Addr:    ":8080",
    Handler: router.GetEngine(),
}
```

## セキュリティ

- JWT認証による保護されたエンドポイント
- 管理者権限が必要な操作の制限
- CORS設定によるオリジン制限
- レート制限による攻撃防止

## 設定

### CORS設定
- 許可オリジン: `http://localhost:3000`, `http://localhost:5173`
- 許可メソッド: `GET`, `POST`, `PUT`, `DELETE`, `OPTIONS`
- 認証情報の送信: 許可

### レート制限
- 認証エンドポイント: 1分間に10回まで
- その他のエンドポイント: 制限なし
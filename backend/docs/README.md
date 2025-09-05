# API ドキュメント

このディレクトリには、Tournament Backend APIの包括的なドキュメントが含まれています。

## 概要

Tournament Backend APIは、バレーボール、卓球、8人制サッカーの3つのスポーツのトーナメント管理を提供するREST APIです。JWT認証を使用した管理者ダッシュボードと、リアルタイムのトーナメント更新機能を含みます。

## ドキュメント形式

### 1. OpenAPI/Swagger ドキュメント

- **swagger.yaml**: OpenAPI 3.0.3形式の完全なAPI仕様
- **swagger.json**: JSON形式のAPI仕様
- **docs.go**: Go言語のSwagger注釈

### 2. インタラクティブドキュメント

サーバー起動時に以下のURLでアクセス可能：

- **Swagger UI**: `http://localhost:8080/swagger/index.html`
- **ドキュメントルート**: `http://localhost:8080/docs`
- **API情報**: `http://localhost:8080/`

## API エンドポイント概要

### 認証 (Authentication)

| メソッド | エンドポイント | 説明 | 認証 |
|---------|---------------|------|------|
| POST | `/api/auth/login` | ログイン | 不要 |
| POST | `/api/auth/refresh` | トークンリフレッシュ | 不要 |
| POST | `/api/auth/logout` | ログアウト | 不要 |
| GET | `/api/auth/profile` | プロフィール取得 | 必要 |
| POST | `/api/auth/validate` | トークン検証 | 不要 |

### トーナメント (Tournaments)

| メソッド | エンドポイント | 説明 | 認証 | 管理者権限 |
|---------|---------------|------|------|----------|
| GET | `/api/tournaments` | 全トーナメント取得 | 必要 | 不要 |
| POST | `/api/tournaments` | トーナメント作成 | 必要 | 必要 |
| GET | `/api/tournaments/{sport}` | スポーツ別トーナメント取得 | 必要 | 不要 |
| PUT | `/api/tournaments/{id}` | トーナメント更新 | 必要 | 必要 |
| DELETE | `/api/tournaments/{id}` | トーナメント削除 | 必要 | 必要 |
| GET | `/api/tournaments/{sport}/bracket` | ブラケット取得 | 必要 | 不要 |
| PUT | `/api/tournaments/{sport}/format` | 形式切り替え | 必要 | 必要 |
| GET | `/api/tournaments/active` | アクティブトーナメント取得 | 必要 | 不要 |
| GET | `/api/tournaments/{sport}/progress` | 進行状況取得 | 必要 | 不要 |
| PUT | `/api/tournaments/{sport}/complete` | トーナメント完了 | 必要 | 必要 |
| PUT | `/api/tournaments/{sport}/activate` | トーナメント有効化 | 必要 | 必要 |

### 試合 (Matches)

| メソッド | エンドポイント | 説明 | 認証 | 管理者権限 |
|---------|---------------|------|------|----------|
| GET | `/api/matches` | 全試合取得 | 必要 | 不要 |
| POST | `/api/matches` | 試合作成 | 必要 | 必要 |
| GET | `/api/matches/{sport}` | スポーツ別試合取得 | 必要 | 不要 |
| GET | `/api/matches/id/{id}` | 特定試合取得 | 必要 | 不要 |
| PUT | `/api/matches/{id}` | 試合更新 | 必要 | 必要 |
| DELETE | `/api/matches/{id}` | 試合削除 | 必要 | 必要 |
| PUT | `/api/matches/{id}/result` | 試合結果提出 | 必要 | 必要 |
| GET | `/api/matches/tournament/{tournament_id}` | トーナメント別試合取得 | 必要 | 不要 |
| GET | `/api/matches/tournament/{tournament_id}/statistics` | 試合統計取得 | 必要 | 不要 |
| GET | `/api/matches/tournament/{tournament_id}/next` | 次の試合取得 | 必要 | 不要 |

### システム (System)

| メソッド | エンドポイント | 説明 | 認証 |
|---------|---------------|------|------|
| GET | `/health` | ヘルスチェック | 不要 |

## 認証

### JWT認証

APIはJWT（JSON Web Token）ベースの認証を使用します。

#### 認証フロー

1. **ログイン**: `/api/auth/login` エンドポイントに認証情報を送信
2. **トークン取得**: 成功時にJWTトークンを受信
3. **認証ヘッダー**: 保護されたエンドポイントには `Authorization: Bearer <token>` ヘッダーを付与

#### 認証情報

- **ユーザー名**: `admin`
- **パスワード**: `password`

#### トークンの有効期限

- **デフォルト**: 24時間
- **リフレッシュ**: `/api/auth/refresh` エンドポイントで更新可能

### レート制限

認証エンドポイント（`/api/auth/login`, `/api/auth/refresh`）には、1分間に10回までのレート制限が適用されます。

## データモデル

### スポーツタイプ

- `volleyball`: バレーボール
- `table_tennis`: 卓球
- `soccer`: サッカー（8人制）

### トーナメント形式

- `standard`: 標準形式
- `rainy`: 雨天時形式（卓球のみ）

### 試合ステータス

- `pending`: 未完了
- `completed`: 完了

### トーナメントステータス

- `active`: アクティブ
- `completed`: 完了

## エラーハンドリング

APIは一貫したエラーレスポンス形式を使用します：

```json
{
  "success": false,
  "message": "エラーメッセージ",
  "error": "エラーコード",
  "code": 400
}
```

### HTTPステータスコード

- **200**: 成功
- **201**: 作成成功
- **400**: 無効なリクエスト
- **401**: 認証が必要
- **403**: 権限不足
- **404**: リソースが見つからない
- **429**: レート制限
- **500**: サーバーエラー

### エラーコード

- `VALIDATION_ERROR`: 入力検証エラー
- `AUTHENTICATION_ERROR`: 認証エラー
- `AUTHORIZATION_ERROR`: 認可エラー
- `NOT_FOUND_ERROR`: リソース未発見
- `RATE_LIMIT_ERROR`: レート制限
- `DATABASE_ERROR`: データベースエラー
- `BUSINESS_LOGIC_ERROR`: ビジネスロジックエラー

## リクエスト/レスポンス例

### ログイン

**リクエスト**:
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

**レスポンス**:
```json
{
  "success": true,
  "message": "ログインに成功しました",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "admin",
  "role": "admin"
}
```

### トーナメント取得

**リクエスト**:
```bash
curl -X GET http://localhost:8080/api/tournaments/volleyball \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**レスポンス**:
```json
{
  "success": true,
  "message": "トーナメント情報を取得しました",
  "data": {
    "id": 1,
    "sport": "volleyball",
    "format": "standard",
    "status": "active",
    "created_at": "2024-01-01T10:00:00Z",
    "updated_at": "2024-01-01T10:00:00Z"
  }
}
```

### 試合結果提出

**リクエスト**:
```bash
curl -X PUT http://localhost:8080/api/matches/1/result \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "score1": 3,
    "score2": 1,
    "winner": "チームA"
  }'
```

**レスポンス**:
```json
{
  "success": true,
  "message": "試合結果を更新しました",
  "data": {
    "id": 1,
    "tournament_id": 1,
    "round": "1st_round",
    "team1": "チームA",
    "team2": "チームB",
    "score1": 3,
    "score2": 1,
    "winner": "チームA",
    "status": "completed",
    "scheduled_at": "2024-01-01T10:00:00Z",
    "completed_at": "2024-01-01T11:00:00Z"
  }
}
```

## 開発者向け情報

### Swagger注釈の更新

新しいエンドポイントを追加する場合：

1. ハンドラーメソッドにSwagger注釈を追加
2. 必要に応じて新しいリクエスト/レスポンス構造体を定義
3. Swaggerドキュメントを再生成

```bash
# Swaggerドキュメント再生成
go run github.com/swaggo/swag/cmd/swag@latest init -g docs/docs.go -o docs
```

### ドキュメント検証

生成されたドキュメントの検証：

```bash
# OpenAPI仕様の検証
npx @apidevtools/swagger-parser validate docs/swagger.yaml

# または、オンラインバリデーター
# https://editor.swagger.io/ にswagger.yamlをアップロード
```

### 自動生成ファイル

以下のファイルは自動生成されるため、直接編集しないでください：

- `docs/docs.go` (Swagger注釈から生成)
- `docs/swagger.json`
- `docs/swagger.yaml` (一部は手動作成、一部は自動生成)

## トラブルシューティング

### よくある問題

1. **Swagger UIが表示されない**
   - サーバーが起動していることを確認
   - `/swagger/index.html` にアクセス
   - ブラウザのキャッシュをクリア

2. **認証エラー**
   - JWTトークンの有効期限を確認
   - `Authorization` ヘッダーの形式を確認 (`Bearer <token>`)

3. **CORS エラー**
   - フロントエンドのオリジンがCORS設定に含まれているか確認
   - プリフライトリクエストが正しく処理されているか確認

### デバッグ

詳細なログを有効にするには：

```bash
export LOG_LEVEL=debug
go run cmd/server/main.go
```

## 関連リンク

- [OpenAPI Specification](https://swagger.io/specification/)
- [Swagger UI](https://swagger.io/tools/swagger-ui/)
- [Go Swagger](https://github.com/swaggo/swag)
- [Gin Web Framework](https://gin-gonic.com/)
- [JWT.io](https://jwt.io/)
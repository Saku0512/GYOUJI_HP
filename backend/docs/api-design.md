# Tournament Backend API Design

## 概要

トーナメント管理システムのバックエンドAPIの設計仕様書です。RESTful APIとして設計され、JSON形式でデータの送受信を行います。

## 基本情報

- **ベースURL**: `/api`
- **データ形式**: JSON
- **認証方式**: JWT Bearer Token
- **文字エンコーディング**: UTF-8

## 認証API

### POST /api/auth/login
管理者ログイン

**リクエスト:**
```json
{
  "username": "admin",
  "password": "password"
}
```

**レスポンス (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "username": "admin",
  "role": "admin",
  "message": "ログインに成功しました"
}
```

### POST /api/auth/refresh
JWTトークンリフレッシュ

**リクエスト:**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
}
```

**レスポンス (200):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "message": "トークンのリフレッシュに成功しました"
}
```

### POST /api/auth/logout
ログアウト

**レスポンス (200):**
```json
{
  "message": "ログアウトしました。クライアント側でトークンを削除してください"
}
```

### GET /api/auth/profile
現在のユーザー情報取得

**認証**: 必須

**レスポンス (200):**
```json
{
  "user_id": 1,
  "username": "admin",
  "role": "admin",
  "message": "ユーザー情報を取得しました"
}
```

## トーナメントAPI

### POST /api/tournaments
トーナメント作成

**認証**: 必須（管理者のみ）

**リクエスト:**
```json
{
  "sport": "volleyball",
  "format": "standard"
}
```

**レスポンス (201):**
```json
{
  "id": 1,
  "sport": "volleyball",
  "format": "standard",
  "status": "active",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z",
  "message": "トーナメントを作成しました"
}
```

### GET /api/tournaments
全トーナメント取得

**レスポンス (200):**
```json
{
  "tournaments": [
    {
      "id": 1,
      "sport": "volleyball",
      "format": "standard",
      "status": "active",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ],
  "count": 1,
  "message": "トーナメント一覧を取得しました"
}
```

### GET /api/tournaments/active
アクティブトーナメント取得

**レスポンス (200):**
```json
{
  "tournaments": [
    {
      "id": 1,
      "sport": "volleyball",
      "format": "standard",
      "status": "active",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ],
  "count": 1,
  "message": "アクティブトーナメント一覧を取得しました"
}
```

### GET /api/tournaments/{sport}
スポーツ別トーナメント取得

**パラメータ:**
- `sport`: スポーツ名 (`volleyball`, `table_tennis`, `soccer`)

**レスポンス (200):**
```json
{
  "id": 1,
  "sport": "volleyball",
  "format": "standard",
  "status": "active",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z",
  "message": "トーナメントを取得しました"
}
```

### GET /api/tournaments/id/{id}
ID別トーナメント取得

**パラメータ:**
- `id`: トーナメントID

**レスポンス (200):**
```json
{
  "id": 1,
  "sport": "volleyball",
  "format": "standard",
  "status": "active",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z",
  "message": "トーナメントを取得しました"
}
```

### PUT /api/tournaments/id/{id}
トーナメント更新

**認証**: 必須（管理者のみ）

**パラメータ:**
- `id`: トーナメントID

**リクエスト:**
```json
{
  "format": "rainy",
  "status": "completed"
}
```

**レスポンス (200):**
```json
{
  "id": 1,
  "sport": "table_tennis",
  "format": "rainy",
  "status": "completed",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T11:00:00Z",
  "message": "トーナメントを更新しました"
}
```

### DELETE /api/tournaments/id/{id}
トーナメント削除

**認証**: 必須（管理者のみ）

**パラメータ:**
- `id`: トーナメントID

**レスポンス (200):**
```json
{
  "message": "トーナメントを削除しました"
}
```

### GET /api/tournaments/{sport}/bracket
トーナメントブラケット取得

**パラメータ:**
- `sport`: スポーツ名 (`volleyball`, `table_tennis`, `soccer`)

**レスポンス (200):**
```json
{
  "tournament_id": 1,
  "sport": "volleyball",
  "format": "standard",
  "rounds": [
    {
      "name": "1st_round",
      "matches": [
        {
          "id": 1,
          "tournament_id": 1,
          "round": "1st_round",
          "team1": "チームA",
          "team2": "チームB",
          "score1": 0,
          "score2": 0,
          "winner": "",
          "status": "pending",
          "scheduled_at": "2024-01-01T11:00:00Z",
          "completed_at": null
        }
      ]
    }
  ],
  "message": "ブラケットを取得しました"
}
```

### PUT /api/tournaments/{sport}/format
トーナメント形式切り替え

**認証**: 必須（管理者のみ）

**パラメータ:**
- `sport`: スポーツ名（現在は `table_tennis` のみサポート）

**リクエスト:**
```json
{
  "format": "rainy"
}
```

**レスポンス (200):**
```json
{
  "message": "トーナメント形式を切り替えました",
  "sport": "table_tennis",
  "format": "rainy"
}
```

### GET /api/tournaments/{sport}/progress
トーナメント進行状況取得

**パラメータ:**
- `sport`: スポーツ名 (`volleyball`, `table_tennis`, `soccer`)

**レスポンス (200):**
```json
{
  "tournament_id": 1,
  "sport": "volleyball",
  "format": "standard",
  "status": "active",
  "total_matches": 8,
  "completed_matches": 4,
  "pending_matches": 4,
  "progress_percent": 50.0,
  "current_round": "quarterfinal",
  "message": "トーナメント進行状況を取得しました"
}
```

### PUT /api/tournaments/{sport}/complete
トーナメント完了

**認証**: 必須（管理者のみ）

**パラメータ:**
- `sport`: スポーツ名 (`volleyball`, `table_tennis`, `soccer`)

**レスポンス (200):**
```json
{
  "message": "トーナメントを完了しました",
  "sport": "volleyball"
}
```

### PUT /api/tournaments/{sport}/activate
トーナメントアクティブ化

**認証**: 必須（管理者のみ）

**パラメータ:**
- `sport`: スポーツ名 (`volleyball`, `table_tennis`, `soccer`)

**レスポンス (200):**
```json
{
  "message": "トーナメントをアクティブ化しました",
  "sport": "volleyball"
}
```

## データモデル

### Tournament（トーナメント）
```json
{
  "id": 1,
  "sport": "volleyball",
  "format": "standard",
  "status": "active",
  "created_at": "2024-01-01T10:00:00Z",
  "updated_at": "2024-01-01T10:00:00Z"
}
```

### Match（試合）
```json
{
  "id": 1,
  "tournament_id": 1,
  "round": "1st_round",
  "team1": "チームA",
  "team2": "チームB",
  "score1": 2,
  "score2": 1,
  "winner": "チームA",
  "status": "completed",
  "scheduled_at": "2024-01-01T11:00:00Z",
  "completed_at": "2024-01-01T12:00:00Z"
}
```

### Bracket（ブラケット）
```json
{
  "tournament_id": 1,
  "sport": "volleyball",
  "format": "standard",
  "rounds": [
    {
      "name": "1st_round",
      "matches": [...]
    }
  ]
}
```

## 定数値

### スポーツタイプ
- `volleyball`: バレーボール
- `table_tennis`: 卓球
- `soccer`: サッカー

### トーナメントフォーマット
- `standard`: 標準フォーマット
- `rainy`: 雨天時フォーマット（卓球のみ）

### トーナメントステータス
- `active`: アクティブ
- `completed`: 完了

### 試合ステータス
- `pending`: 未実施
- `completed`: 完了

### ラウンド名
- `1st_round`: 1回戦
- `quarterfinal`: 準々決勝
- `semifinal`: 準決勝
- `third_place`: 3位決定戦
- `final`: 決勝
- `loser_bracket`: 敗者復活戦（卓球の雨天時のみ）

## エラーレスポンス

### 統一エラー形式
```json
{
  "error": "Bad Request",
  "message": "無効なリクエスト形式です",
  "code": 400
}
```

### HTTPステータスコード
- `200`: 成功
- `201`: 作成成功
- `400`: リクエストエラー
- `401`: 認証エラー
- `404`: リソースが見つからない
- `409`: 競合エラー
- `500`: サーバーエラー

## セキュリティ

### 認証
- JWT Bearer Tokenを使用
- 管理者権限が必要なエンドポイントは認証必須
- トークンの有効期限管理

### 入力検証
- 全てのリクエストパラメータの検証
- SQLインジェクション対策
- XSS対策

### エラーハンドリング
- 詳細なエラー情報の適切な制御
- ログ出力による監査証跡
- 統一されたエラーレスポンス形式

## パフォーマンス考慮事項

### データベース
- 適切なインデックス設計
- トランザクション管理
- 接続プール管理

### キャッシュ戦略
- 頻繁にアクセスされるデータのキャッシュ
- リアルタイム更新との整合性

### レスポンス最適化
- 必要最小限のデータ転送
- ページネーション対応（将来拡張）
- 圧縮対応

## 今後の拡張予定

### 機能拡張
- 試合結果管理API
- リアルタイム通知機能
- 統計・分析API
- ファイルアップロード機能

### 技術的改善
- GraphQL対応
- WebSocket対応
- API バージョニング
- レート制限機能
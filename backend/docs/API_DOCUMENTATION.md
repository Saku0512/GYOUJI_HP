# API Documentation

## 概要

このドキュメントは、トーナメント管理システムのAPI仕様書について説明します。

## ファイル構成

- `openapi.yaml` - メインのOpenAPI 3.0仕様書
- `openapi-matches.yaml` - 試合関連エンドポイントの追加定義
- `openapi-admin-matches.yaml` - 管理者専用試合エンドポイントの追加定義
- `swagger.yaml` - 既存のSwagger仕様書（後方互換性のため維持）

## OpenAPI仕様書の特徴

### 1. 統一されたレスポンス形式

全てのAPIエンドポイントは以下の統一されたレスポンス形式を使用します：

```json
{
  "success": true,
  "data": {...},
  "message": "操作が成功しました",
  "code": 200,
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789"
}
```

### 2. エラーハンドリング

エラーレスポンスも統一された形式を使用します：

```json
{
  "success": false,
  "error": "VALIDATION_ERROR",
  "message": "入力データが無効です",
  "code": 400,
  "timestamp": "2024-01-01T12:00:00Z",
  "request_id": "req_123456789"
}
```

### 3. データ型の統一

- **日時**: ISO 8601形式 (`2024-01-01T12:00:00Z`)
- **列挙型**: 厳密に定義された値セット
- **ID**: 正の整数
- **null値**: 明示的に許可された場合のみ

### 4. 認証

JWT（JSON Web Token）を使用した認証システム：

```
Authorization: Bearer <token>
```

## エンドポイント分類

### 公開エンドポイント（認証不要）
- `/public/tournaments/*` - トーナメント情報の公開取得
- `/public/matches/*` - 試合情報の公開取得
- `/health` - ヘルスチェック

### 認証が必要なエンドポイント
- `/auth/*` - 認証関連
- `/tournaments/*` - トーナメント情報の詳細取得
- `/matches/*` - 試合情報の詳細取得

### 管理者専用エンドポイント
- `/admin/tournaments/*` - トーナメント管理
- `/admin/matches/*` - 試合管理
- `/admin/websocket/*` - WebSocket管理
- `/admin/polling/*` - ポーリング管理

## バージョニング

- **現在のバージョン**: `/api/v1/`
- **旧バージョン**: `/api/` (廃止予定)

旧APIには以下のヘッダーが付与されます：
```
X-API-Deprecated: true
X-API-Deprecation-Message: このAPIは廃止予定です。/api/v1を使用してください
```

## レート制限

認証エンドポイントには以下のレート制限が適用されます：
- ログイン: 10回/分
- トークンリフレッシュ: 10回/分

## ページネーション

リスト取得エンドポイントではページネーションをサポートします：

**リクエストパラメータ:**
- `page`: ページ番号 (デフォルト: 1)
- `page_size`: ページサイズ (デフォルト: 20, 最大: 100)

**レスポンス:**
```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "page": 1,
    "page_size": 20,
    "total_items": 100,
    "total_pages": 5,
    "has_next": true,
    "has_prev": false
  }
}
```

## フィルタリング

多くのエンドポイントでクエリパラメータによるフィルタリングをサポートします：

- `sport`: スポーツ種目でフィルタ
- `status`: ステータスでフィルタ
- `round`: ラウンドでフィルタ
- `tournament_id`: トーナメントIDでフィルタ

## 使用例

### 1. ログイン

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

### 2. トーナメント一覧取得

```bash
curl -X GET http://localhost:8080/api/v1/public/tournaments \
  -H "Content-Type: application/json"
```

### 3. 試合結果提出（管理者のみ）

```bash
curl -X PUT http://localhost:8080/api/v1/admin/matches/1/result \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "score1": 3,
    "score2": 1,
    "winner": "チームA"
  }'
```

## Swagger UI

開発環境では、以下のURLでSwagger UIにアクセスできます：

```
http://localhost:8080/swagger/index.html
```

## 仕様書の更新

API仕様を変更した場合は、以下の手順で仕様書を更新してください：

1. `openapi.yaml`を更新
2. 必要に応じて追加ファイル（`openapi-matches.yaml`等）を更新
3. Swagger UIで仕様書を確認
4. 契約テストを実行して整合性を確認

## 契約テスト

API仕様とバックエンド実装の整合性を確認するため、契約テストを実装しています。

テスト実行：
```bash
cd backend
go test ./integration_test/... -v
```

## 注意事項

1. **後方互換性**: 旧API（`/api/`）は廃止予定ですが、移行期間中は利用可能です
2. **レート制限**: 認証エンドポイントには制限があります
3. **データ形式**: 日時は必ずISO 8601形式で送信してください
4. **エラーハンドリング**: エラーレスポンスの`error`フィールドを適切に処理してください
5. **認証**: 管理者専用エンドポイントには適切な権限が必要です

## サポート

API仕様に関する質問や問題がある場合は、以下にお問い合わせください：

- Email: support@tournament.example.com
- GitHub Issues: [プロジェクトリポジトリ]
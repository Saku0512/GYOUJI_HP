# ログとエラーハンドリング実装

## 概要

タスク8「包括的なエラーハンドリングとログの追加」の実装が完了しました。構造化ログとカスタムエラーハンドリングシステムを導入し、アプリケーション全体の監視性とデバッグ性を向上させました。

## 実装内容

### 8.1 構造化ログの実装

#### 新規作成ファイル
- `internal/logger/logger.go` - 構造化ログのメインインターフェースと実装
- `internal/logger/middleware.go` - リクエストIDとログミドルウェア
- `internal/logger/logger_test.go` - ログ機能のテスト
- `internal/logger/middleware_test.go` - ミドルウェアのテスト

#### 主要機能
1. **構造化ログ**: logrusを使用したJSON/テキスト形式の構造化ログ
2. **ログレベル**: Debug, Info, Warn, Error, Fatalの5段階
3. **環境別フォーマット**: 
   - 開発環境: 読みやすいテキスト形式
   - 本番環境: 解析しやすいJSON形式
4. **リクエストID追跡**: UUIDベースのリクエスト相関ID
5. **コンテキスト対応**: Ginのコンテキストからリクエスト情報を自動取得

#### 使用例
```go
log := logger.GetLogger()
log.Info("ユーザーが作成されました", 
    logger.String("username", "admin"),
    logger.Int("user_id", 123),
)

// リクエストIDを含むログ
log.WithRequestID("req-123").Error("エラーが発生しました", logger.Err(err))
```

### 8.2 包括的なエラーハンドリングの作成

#### 新規作成ファイル
- `internal/errors/errors.go` - カスタムエラータイプとヘルパー関数
- `internal/errors/middleware.go` - エラーハンドリングミドルウェア
- `internal/errors/errors_test.go` - エラー機能のテスト
- `internal/errors/middleware_test.go` - ミドルウェアのテスト

#### エラータイプ
1. **ValidationError** (400) - 入力検証エラー
2. **AuthenticationError** (401) - 認証エラー
3. **AuthorizationError** (403) - 認可エラー
4. **NotFoundError** (404) - リソース未発見エラー
5. **ConflictError** (409) - リソース競合エラー
6. **BusinessLogicError** (422) - ビジネスロジックエラー
7. **DatabaseError** (500) - データベースエラー
8. **InternalError** (500) - 内部サーバーエラー

#### データベースエラーハンドリング
- MySQL固有エラーの自動検出と適切なHTTPステータスコードへの変換
- 重複エントリ、外部キー制約、NULL制約などの自動処理
- 接続エラーやタイムアウトの検出

#### 使用例
```go
// バリデーションエラー
return errors.NewValidationError("無効な入力です").WithField("field", "username")

// データベースエラー
return errors.NewDatabaseError("データ保存に失敗しました", err)

// ビジネスロジックエラー
return errors.NewBusinessLogicError("既に存在するトーナメントです")
```

## 更新されたファイル

### メインサーバー
- `cmd/server/main.go` - 新しいログシステムを使用するよう更新

### ルーター
- `internal/router/router.go` - 新しいミドルウェアを統合

### サービス層（部分的更新）
- `internal/service/tournament.go` - エラーハンドリングとログの一部を更新

## ミドルウェア統合

新しいミドルウェアスタック（順序重要）:
1. **RecoveryMiddleware** - パニック回復
2. **RequestIDMiddleware** - リクエストID生成
3. **LoggingMiddleware** - HTTPリクエスト/レスポンスログ
4. **ErrorHandlerMiddleware** - エラーレスポンス統一

## 環境変数

新しい環境変数:
- `GO_ENV`: 環境設定（development/production）
- `LOG_LEVEL`: ログレベル（debug/info/warn/error）

## テスト

全てのテストが通過:
- ログ機能: 15テストケース
- エラーハンドリング: 20テストケース
- ミドルウェア: 10テストケース

## 利点

1. **監視性向上**: 構造化ログによる効率的な問題追跡
2. **デバッグ効率**: リクエストIDによる分散ログの相関
3. **一貫性**: 統一されたエラーレスポンス形式
4. **保守性**: 型安全なエラーハンドリング
5. **運用性**: 環境別の適切なログフォーマット

## 今後の拡張

- 他のサービス層への完全適用
- メトリクス収集の統合
- 分散トレーシングの追加
- ログ集約システムとの連携
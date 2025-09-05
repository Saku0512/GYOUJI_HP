# 統合テスト

このディレクトリには、トーナメントバックエンドAPIの包括的な統合テストが含まれています。

## 概要

統合テストは、実際のデータベースとHTTPサーバーを使用してAPIエンドポイントの動作を検証します。以下の要件をカバーしています：

### テスト対象の要件

- **要件1.1-1.5**: 認証システム（JWT認証、ログイン、トークン検証）
- **要件2.1-2.4**: トーナメントデータ管理（CRUD操作、ブラケット更新）
- **要件3.1-3.5**: データベース統合（MySQL接続、データ永続化）
- **要件4.1-4.4**: リアルタイムトーナメント更新
- **要件5.1-5.5**: トーナメント構造サポート（複数スポーツ、形式切り替え）
- **要件6.1-6.5**: APIエンドポイント（REST API、認証、エラーハンドリング）
- **要件7.1-7.5**: エラーハンドリングとログ

## テストスイート構成

### 1. 認証統合テスト (`auth_integration_test.go`)

- ログイン機能のテスト
- JWTトークン生成と検証
- トークンリフレッシュ機能
- 認証ミドルウェアの動作
- レート制限機能

### 2. トーナメント統合テスト (`tournament_integration_test.go`)

- トーナメントCRUD操作
- スポーツ別トーナメント取得
- ブラケット生成と更新
- 卓球の形式切り替え（晴天/雨天）
- トーナメント進行ワークフロー

### 3. 試合統合テスト (`match_integration_test.go`)

- 試合CRUD操作
- 試合結果提出
- スポーツ別試合取得
- 試合データ検証
- 試合進行ワークフロー

### 4. ワークフロー統合テスト (`workflow_integration_test.go`)

- エンドツーエンドワークフロー
- 複数スポーツの統合テスト
- エラーハンドリングフロー
- システム全体の動作確認

## セットアップ

### 前提条件

1. **MySQL データベース**
   ```bash
   # Dockerを使用する場合
   docker run --name mysql-test -e MYSQL_ROOT_PASSWORD=test_password -p 3306:3306 -d mysql:8.0
   
   # または既存のMySQLインスタンスを使用
   ```

2. **Go 1.24.2以上**

3. **環境変数**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=3306
   export DB_USER=root
   export DB_PASSWORD=test_password
   export DB_NAME=tournament_test_db
   export JWT_SECRET=test_jwt_secret_key_for_testing
   ```

### テスト実行

#### 自動実行スクリプト使用（推奨）

```bash
# Linux/Mac
./scripts/run_integration_tests.sh

# Windows
scripts\run_integration_tests.bat

# カバレッジレポート付き
./scripts/run_integration_tests.sh --coverage
```

#### 手動実行

```bash
# 全テストスイート実行
go test -v ./integration_test/

# 特定のテストスイート実行
go test -v -run TestAuthIntegrationTestSuite ./integration_test/
go test -v -run TestTournamentIntegrationTestSuite ./integration_test/
go test -v -run TestMatchIntegrationTestSuite ./integration_test/
go test -v -run TestWorkflowIntegrationTestSuite ./integration_test/

# カバレッジ付き実行
go test -v -coverprofile=coverage.out ./integration_test/
go tool cover -html=coverage.out -o coverage.html
```

## テストユーティリティ

### `testutil/database.go`

テストデータベースの管理機能：

- **SetupTestDatabase**: テスト用データベースの初期化
- **TeardownTestDatabase**: テスト後のクリーンアップ
- **SeedTestData**: テストデータの投入
- **CleanupTestData**: テストデータの削除

### `testutil/http.go`

HTTPテスト用のユーティリティ：

- **SetupTestServer**: テスト用HTTPサーバーの初期化
- **MakeRequest**: HTTPリクエストの実行
- **MakeAuthenticatedRequest**: 認証付きリクエストの実行
- **LoginAndGetToken**: ログインしてトークン取得
- **AssertJSONResponse**: JSONレスポンスの検証
- **AssertErrorResponse**: エラーレスポンスの検証

## テストデータ

各テストは以下のテストデータを使用します：

### ユーザー
- **管理者ユーザー**: `admin` / `password`

### トーナメント
- **バレーボール**: 標準形式、アクティブ
- **卓球**: 標準形式（晴天/雨天切り替え可能）、アクティブ
- **サッカー**: 8人制形式、アクティブ

### 試合
- 各スポーツに対して初期試合データ
- 未完了状態（`pending`）で作成

## テスト実行の流れ

1. **セットアップ**
   - テスト用データベース接続
   - マイグレーション実行
   - テストデータ投入

2. **テスト実行**
   - 各テストスイートの実行
   - APIエンドポイントの検証
   - データベース状態の確認

3. **クリーンアップ**
   - テストデータの削除
   - データベース接続の終了

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**
   ```
   Error: データベース接続の初期化に失敗しました
   ```
   - MySQLサーバーが起動していることを確認
   - 接続情報（ホスト、ポート、ユーザー、パスワード）を確認

2. **マイグレーションエラー**
   ```
   Error: マイグレーションの実行に失敗しました
   ```
   - マイグレーションファイルのパスを確認
   - データベース権限を確認

3. **テストタイムアウト**
   ```
   Error: test timed out
   ```
   - データベースの応答時間を確認
   - テスト環境のリソースを確認

### デバッグ

詳細なログを有効にするには：

```bash
# デバッグログ有効
export LOG_LEVEL=debug
go test -v ./integration_test/

# SQLクエリログ有効
export DB_LOG_LEVEL=debug
go test -v ./integration_test/
```

## CI/CD統合

GitHub ActionsやJenkinsなどのCI/CDパイプラインで使用する場合：

```yaml
# .github/workflows/integration-tests.yml の例
name: Integration Tests
on: [push, pull_request]

jobs:
  integration-tests:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.0
        env:
          MYSQL_ROOT_PASSWORD: test_password
        ports:
          - 3306:3306
        options: --health-cmd="mysqladmin ping" --health-interval=10s --health-timeout=5s --health-retries=3

    steps:
    - uses: actions/checkout@v2
    - uses: actions/setup-go@v2
      with:
        go-version: 1.24.2
    
    - name: Run Integration Tests
      run: |
        cd backend
        ./scripts/run_integration_tests.sh
```

## 貢献

新しいテストを追加する場合：

1. 適切なテストスイートファイルに追加
2. テストメソッド名は `Test` で開始
3. 要件番号をコメントで明記
4. 適切なアサーションを使用
5. テスト後のクリーンアップを確実に実行

## 参考資料

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Gin Testing](https://gin-gonic.com/docs/testing/)
- [MySQL Testing Best Practices](https://dev.mysql.com/doc/refman/8.0/en/testing.html)
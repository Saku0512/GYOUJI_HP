# 強化された統合テストガイド

## 概要

このドキュメントは、API統合修正プロジェクトで実装された強化された統合テストスイートについて説明します。これらのテストは、フロントエンドとバックエンド間のAPI統合問題を解決し、一貫性のあるAPI設計を検証するために設計されています。

## テストスイート構成

### 1. 契約テスト (`contract_test.go`)

**目的**: API仕様とバックエンド実装の整合性を検証

**主要テスト項目**:
- 統一されたAPIレスポンス形式の検証
- データ型整合性の確認
- エラーハンドリングの標準化検証
- 日時形式（ISO 8601）の検証
- 列挙型値の検証

**実行例**:
```bash
go test -v -run TestContractTestSuite ./integration_test/
```

### 2. OpenAPI仕様検証テスト (`openapi_validation_test.go`)

**目的**: OpenAPI仕様書と実際のAPI実装の整合性を検証

**主要テスト項目**:
- OpenAPI仕様ファイルの構文検証
- レスポンススキーマの検証
- エンドポイント定義の確認
- セキュリティスキームの検証

**実行例**:
```bash
go test -v -run TestOpenAPIValidationTestSuite ./integration_test/
```

### 3. エンドツーエンドAPIテスト (`e2e_api_test.go`)

**目的**: 実際のユーザーワークフローをシミュレートした包括的テスト

**主要テスト項目**:
- 完全なワークフローテスト（ログイン→データ取得→操作→結果確認）
- 複数スポーツのワークフローテスト
- エラーハンドリングワークフロー
- パフォーマンスと制限のテスト
- 同時アクセステスト

**実行例**:
```bash
go test -v -run TestE2EAPITestSuite ./integration_test/
```

### 4. 認証フローテスト (`auth_flow_test.go`)

**目的**: 認証システムの包括的テスト

**主要テスト項目**:
- 完全なログイン・ログアウトサイクル
- トークンリフレッシュフロー
- 無効な認証情報のテスト
- 無効なトークンのテスト
- トークン検証テスト
- 役割ベースのアクセス制御
- 同時認証テスト

**実行例**:
```bash
go test -v -run TestAuthFlowTestSuite ./integration_test/
```

### 5. エラーケーステスト (`error_cases_test.go`)

**目的**: 様々なエラーシナリオの包括的テスト

**主要テスト項目**:
- 認証エラー（401, 403）
- バリデーションエラー（400）
- リソース未発見エラー（404）
- ビジネスロジックエラー（422）
- 無効なJSONリクエスト
- HTTPメソッドエラー（405）
- レート制限エラー（429）
- 同時エラーハンドリング

**実行例**:
```bash
go test -v -run TestErrorCasesTestSuite ./integration_test/
```

## 実行方法

### 基本実行

```bash
# 全てのテストを実行
./scripts/run_enhanced_tests.sh

# カバレッジ付きで実行
./scripts/run_enhanced_tests.sh --coverage

# 並列実行で高速化
./scripts/run_enhanced_tests.sh --parallel

# 詳細ログ付きで実行
./scripts/run_enhanced_tests.sh --verbose
```

### 特定のテストスイートのみ実行

```bash
# 統合テストのみ
./scripts/run_enhanced_tests.sh --integration-only

# 契約テストのみ
./scripts/run_enhanced_tests.sh --contract-only

# E2Eテストのみ
./scripts/run_enhanced_tests.sh --e2e-only

# 認証テストのみ
./scripts/run_enhanced_tests.sh --auth-only

# エラーケーステストのみ
./scripts/run_enhanced_tests.sh --error-only

# クイックテスト（基本テストのみ）
./scripts/run_enhanced_tests.sh --quick
```

### 契約テスト専用実行

```bash
# 契約テストとOpenAPI検証
./scripts/run_contract_tests.sh

# カバレッジ付き契約テスト
./scripts/run_contract_tests.sh --coverage

# 契約テストのみ
./scripts/run_contract_tests.sh --contract-only

# OpenAPI検証のみ
./scripts/run_contract_tests.sh --openapi-only
```

## 環境設定

### 必要な環境変数

```bash
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=test_password
export DB_NAME=tournament_test_db
export JWT_SECRET=test_jwt_secret_key_for_testing
export LOG_LEVEL=info
```

### データベース設定

```bash
# MySQLサーバーの起動（Dockerを使用する場合）
docker run --name mysql-test \
  -e MYSQL_ROOT_PASSWORD=test_password \
  -e MYSQL_DATABASE=tournament_test_db \
  -p 3306:3306 -d mysql:8.0

# データベースの初期化
mysql -h localhost -P 3306 -u root -ptest_password \
  -e "CREATE DATABASE IF NOT EXISTS tournament_test_db;"
```

## テスト設計原則

### 1. 統一されたレスポンス形式の検証

全てのAPIエンドポイントが以下の統一形式を返すことを検証：

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

### 2. エラーレスポンスの標準化

エラー時も統一された形式を使用：

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

### 3. データ型の一貫性

- **日時**: ISO 8601形式（`2024-01-01T12:00:00Z`）
- **ID**: 正の整数
- **列挙型**: 厳密に定義された値セット
- **null値**: 明示的に許可された場合のみ

### 4. 認証とセキュリティ

- JWT認証の一貫性
- 役割ベースのアクセス制御
- レート制限の適用
- セキュリティヘッダーの設定

## CI/CD統合

### GitHub Actions

`.github/workflows/contract-tests.yml`で自動実行：

```yaml
name: Contract Tests
on: [push, pull_request]
jobs:
  contract-tests:
    runs-on: ubuntu-latest
    services:
      mysql:
        image: mysql:8.0
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
      - name: Run contract tests
        run: ./scripts/run_contract_tests.sh --coverage
```

### 成果物

- カバレッジレポート（HTML）
- テスト結果レポート
- OpenAPIドキュメント

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**
   ```
   Error: データベース接続の初期化に失敗しました
   ```
   - MySQLサーバーが起動していることを確認
   - 環境変数が正しく設定されていることを確認

2. **OpenAPI仕様ファイルが見つからない**
   ```
   Error: OpenAPI仕様ファイルが見つかりません
   ```
   - `backend/docs/openapi.yaml`が存在することを確認
   - ファイルの構文が正しいことを確認

3. **テストタイムアウト**
   ```
   Error: test timed out
   ```
   - データベースの応答時間を確認
   - 並列実行数を調整（`--parallel`オプション）

### デバッグ方法

```bash
# 詳細ログを有効にして実行
export LOG_LEVEL=debug
./scripts/run_enhanced_tests.sh --verbose

# 特定のテストのみ実行
go test -v -run TestContract_AuthLogin_Success ./integration_test/

# カバレッジ詳細を確認
go tool cover -func=coverage.out
```

## カバレッジ目標

- **総合カバレッジ**: 80%以上
- **契約テストカバレッジ**: 90%以上
- **エラーハンドリングカバレッジ**: 85%以上

## 貢献ガイドライン

### 新しいテストの追加

1. 適切なテストスイートファイルに追加
2. テストメソッド名は`Test`で開始
3. 要件番号をコメントで明記
4. 統一されたアサーション形式を使用
5. テスト後のクリーンアップを確実に実行

### テスト命名規則

```go
// 良い例
func (suite *ContractTestSuite) TestContract_AuthLogin_Success() {
    // 要件7.1: API仕様が定義された場合、システムはリクエスト・レスポンスの自動検証を行う
    // テスト実装
}

// 悪い例
func (suite *ContractTestSuite) TestLogin() {
    // 要件番号なし、不明確な名前
}
```

### アサーション形式

```go
// 統一されたレスポンス検証
suite.validateAPIResponse(response, true)

// エラーレスポンス検証
suite.validateErrorResponse(response, http.StatusBadRequest, "VALIDATION_ERROR")

// 日時形式検証
suite.validateDateTimeFormat(dateTime, "created_at")
```

## 参考資料

- [Go Testing Package](https://golang.org/pkg/testing/)
- [Testify Documentation](https://github.com/stretchr/testify)
- [OpenAPI Specification](https://swagger.io/specification/)
- [API Design Guidelines](../docs/API_DOCUMENTATION.md)
- [Contract Testing Best Practices](https://martinfowler.com/articles/practical-test-pyramid.html)

## 更新履歴

- **v1.0.0**: 初期実装（契約テスト、OpenAPI検証）
- **v1.1.0**: E2Eテスト追加
- **v1.2.0**: 認証フローテスト追加
- **v1.3.0**: エラーケーステスト追加
- **v1.4.0**: CI/CD統合、並列実行対応
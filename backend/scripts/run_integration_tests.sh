#!/bin/bash

# 統合テスト実行スクリプト
# このスクリプトは統合テストを実行し、テストデータベースのセットアップとクリーンアップを行います

set -e

# スクリプトのディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== トーナメントバックエンド統合テスト ==="
echo "バックエンドディレクトリ: $BACKEND_DIR"

# バックエンドディレクトリに移動
cd "$BACKEND_DIR"

# 環境変数を設定
export GO_ENV=test
export DB_HOST=localhost
export DB_PORT=3306
export DB_USER=root
export DB_PASSWORD=test_password
export DB_NAME=tournament_test_db
export JWT_SECRET=test_jwt_secret_key_for_testing
export JWT_EXPIRATION_HOURS=24
export JWT_ISSUER=tournament-backend-test
export SERVER_PORT=8081
export SERVER_HOST=localhost

echo "テスト環境変数を設定しました"

# テストデータベースが存在するかチェック
echo "テストデータベースの確認中..."
mysql -h$DB_HOST -P$DB_PORT -u$DB_USER -p$DB_PASSWORD -e "CREATE DATABASE IF NOT EXISTS $DB_NAME;" 2>/dev/null || {
    echo "警告: MySQLデータベースに接続できません。データベースが起動していることを確認してください。"
    echo "以下のコマンドでMySQLを起動できます:"
    echo "  docker run --name mysql-test -e MYSQL_ROOT_PASSWORD=$DB_PASSWORD -p $DB_PORT:3306 -d mysql:8.0"
    echo ""
    echo "または、既存のMySQLインスタンスを使用する場合は、以下の設定を確認してください:"
    echo "  ホスト: $DB_HOST"
    echo "  ポート: $DB_PORT"
    echo "  ユーザー: $DB_USER"
    echo "  パスワード: $DB_PASSWORD"
    echo ""
    read -p "続行しますか？ (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
}

echo "テストデータベースの準備が完了しました"

# Go依存関係の確認
echo "Go依存関係を確認中..."
go mod tidy
go mod download

# テストの実行
echo ""
echo "=== 統合テスト実行開始 ==="
echo ""

# テストの詳細出力を有効にする
export GOMAXPROCS=1

# 各テストスイートを個別に実行
test_suites=(
    "backend/integration_test.TestAuthIntegrationTestSuite"
    "backend/integration_test.TestTournamentIntegrationTestSuite"
    "backend/integration_test.TestMatchIntegrationTestSuite"
    "backend/integration_test.TestWorkflowIntegrationTestSuite"
)

failed_tests=()
passed_tests=()

for suite in "${test_suites[@]}"; do
    echo "--- $suite を実行中 ---"
    if go test -v -run "$suite" ./integration_test/; then
        echo "✅ $suite: 成功"
        passed_tests+=("$suite")
    else
        echo "❌ $suite: 失敗"
        failed_tests+=("$suite")
    fi
    echo ""
done

# 結果の表示
echo "=== テスト結果サマリー ==="
echo "成功したテスト: ${#passed_tests[@]}"
echo "失敗したテスト: ${#failed_tests[@]}"

if [ ${#passed_tests[@]} -gt 0 ]; then
    echo ""
    echo "✅ 成功したテスト:"
    for test in "${passed_tests[@]}"; do
        echo "  - $test"
    done
fi

if [ ${#failed_tests[@]} -gt 0 ]; then
    echo ""
    echo "❌ 失敗したテスト:"
    for test in "${failed_tests[@]}"; do
        echo "  - $test"
    done
    echo ""
    echo "失敗したテストがあります。詳細は上記のログを確認してください。"
    exit 1
fi

echo ""
echo "🎉 すべての統合テストが成功しました！"
echo ""

# カバレッジレポートの生成（オプション）
if [ "$1" = "--coverage" ]; then
    echo "=== カバレッジレポート生成中 ==="
    go test -v -coverprofile=coverage.out ./integration_test/
    go tool cover -html=coverage.out -o coverage.html
    echo "カバレッジレポートが coverage.html に生成されました"
fi

echo "統合テストが完了しました。"
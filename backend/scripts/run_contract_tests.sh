#!/bin/bash

# 契約テスト実行スクリプト
# このスクリプトは、API契約テストとOpenAPI仕様検証を実行します

set -e

# 色付きログ出力用の関数
log_info() {
    echo -e "\033[1;34m[INFO]\033[0m $1"
}

log_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $1"
}

log_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1"
}

log_warning() {
    echo -e "\033[1;33m[WARNING]\033[0m $1"
}

# 使用方法を表示
show_usage() {
    echo "使用方法: $0 [オプション]"
    echo ""
    echo "オプション:"
    echo "  --coverage          カバレッジレポートを生成"
    echo "  --verbose           詳細ログを出力"
    echo "  --contract-only     契約テストのみ実行"
    echo "  --openapi-only      OpenAPI検証テストのみ実行"
    echo "  --help              このヘルプを表示"
    echo ""
    echo "例:"
    echo "  $0                  全ての契約テストを実行"
    echo "  $0 --coverage       カバレッジ付きで実行"
    echo "  $0 --contract-only  契約テストのみ実行"
}

# デフォルト設定
COVERAGE=false
VERBOSE=false
CONTRACT_ONLY=false
OPENAPI_ONLY=false

# コマンドライン引数の解析
while [[ $# -gt 0 ]]; do
    case $1 in
        --coverage)
            COVERAGE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        --contract-only)
            CONTRACT_ONLY=true
            shift
            ;;
        --openapi-only)
            OPENAPI_ONLY=true
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            log_error "不明なオプション: $1"
            show_usage
            exit 1
            ;;
    esac
done

# プロジェクトルートディレクトリに移動
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
cd "$PROJECT_ROOT"

log_info "契約テスト実行を開始します..."
log_info "プロジェクトディレクトリ: $PROJECT_ROOT"

# 環境変数の確認
check_environment() {
    log_info "環境変数を確認しています..."
    
    # 必須環境変数のリスト
    required_vars=(
        "DB_HOST"
        "DB_PORT"
        "DB_USER"
        "DB_PASSWORD"
        "DB_NAME"
        "JWT_SECRET"
    )
    
    missing_vars=()
    
    for var in "${required_vars[@]}"; do
        if [[ -z "${!var}" ]]; then
            missing_vars+=("$var")
        fi
    done
    
    if [[ ${#missing_vars[@]} -gt 0 ]]; then
        log_error "以下の環境変数が設定されていません:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        log_info "環境変数を設定してから再実行してください"
        exit 1
    fi
    
    log_success "環境変数の確認が完了しました"
}

# データベース接続の確認
check_database() {
    log_info "データベース接続を確認しています..."
    
    # MySQLクライアントが利用可能かチェック
    if ! command -v mysql &> /dev/null; then
        log_warning "MySQLクライアントが見つかりません。データベース接続確認をスキップします"
        return 0
    fi
    
    # データベース接続テスト
    if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1;" &> /dev/null; then
        log_success "データベース接続が確認できました"
    else
        log_error "データベースに接続できません"
        log_info "以下を確認してください:"
        echo "  - MySQLサーバーが起動している"
        echo "  - 接続情報が正しい (ホスト: $DB_HOST, ポート: $DB_PORT)"
        echo "  - ユーザー権限が適切に設定されている"
        exit 1
    fi
}

# Go依存関係の確認
check_dependencies() {
    log_info "Go依存関係を確認しています..."
    
    if ! go mod verify &> /dev/null; then
        log_warning "Go依存関係の検証に失敗しました。依存関係を更新します..."
        go mod tidy
        go mod download
    fi
    
    log_success "Go依存関係の確認が完了しました"
}

# OpenAPI仕様ファイルの確認
check_openapi_spec() {
    log_info "OpenAPI仕様ファイルを確認しています..."
    
    openapi_file="docs/openapi.yaml"
    
    if [[ ! -f "$openapi_file" ]]; then
        log_error "OpenAPI仕様ファイルが見つかりません: $openapi_file"
        log_info "OpenAPI仕様ファイルを作成してから再実行してください"
        exit 1
    fi
    
    # YAML構文の確認
    if command -v yamllint &> /dev/null; then
        if yamllint "$openapi_file" &> /dev/null; then
            log_success "OpenAPI仕様ファイルの構文確認が完了しました"
        else
            log_warning "OpenAPI仕様ファイルの構文に問題がある可能性があります"
        fi
    else
        log_info "yamllintが見つかりません。構文確認をスキップします"
    fi
}

# 契約テストの実行
run_contract_tests() {
    log_info "契約テストを実行しています..."
    
    test_args="-v"
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_args="$test_args -test.v"
    fi
    
    if [[ "$COVERAGE" == "true" ]]; then
        test_args="$test_args -coverprofile=contract_coverage.out"
    fi
    
    # 契約テストの実行
    if go test $test_args -run "TestContractTestSuite" ./integration_test/; then
        log_success "契約テストが完了しました"
    else
        log_error "契約テストが失敗しました"
        return 1
    fi
}

# OpenAPI検証テストの実行
run_openapi_tests() {
    log_info "OpenAPI仕様検証テストを実行しています..."
    
    test_args="-v"
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_args="$test_args -test.v"
    fi
    
    if [[ "$COVERAGE" == "true" ]]; then
        test_args="$test_args -coverprofile=openapi_coverage.out"
    fi
    
    # OpenAPI検証テストの実行
    if go test $test_args -run "TestOpenAPIValidationTestSuite" ./integration_test/; then
        log_success "OpenAPI仕様検証テストが完了しました"
    else
        log_error "OpenAPI仕様検証テストが失敗しました"
        return 1
    fi
}

# カバレッジレポートの生成
generate_coverage_report() {
    if [[ "$COVERAGE" != "true" ]]; then
        return 0
    fi
    
    log_info "カバレッジレポートを生成しています..."
    
    # カバレッジファイルの結合
    coverage_files=()
    if [[ -f "contract_coverage.out" ]]; then
        coverage_files+=("contract_coverage.out")
    fi
    if [[ -f "openapi_coverage.out" ]]; then
        coverage_files+=("openapi_coverage.out")
    fi
    
    if [[ ${#coverage_files[@]} -eq 0 ]]; then
        log_warning "カバレッジファイルが見つかりません"
        return 0
    fi
    
    # 結合されたカバレッジファイルの作成
    combined_coverage="combined_contract_coverage.out"
    echo "mode: set" > "$combined_coverage"
    
    for file in "${coverage_files[@]}"; do
        tail -n +2 "$file" >> "$combined_coverage"
    done
    
    # HTMLレポートの生成
    if go tool cover -html="$combined_coverage" -o contract_coverage.html; then
        log_success "カバレッジレポートが生成されました: contract_coverage.html"
    else
        log_error "カバレッジレポートの生成に失敗しました"
    fi
    
    # カバレッジ率の表示
    coverage_percent=$(go tool cover -func="$combined_coverage" | tail -1 | awk '{print $3}')
    log_info "契約テストカバレッジ: $coverage_percent"
    
    # 一時ファイルのクリーンアップ
    rm -f "${coverage_files[@]}" "$combined_coverage"
}

# クリーンアップ
cleanup() {
    log_info "クリーンアップを実行しています..."
    
    # 一時ファイルの削除
    rm -f contract_coverage.out openapi_coverage.out combined_contract_coverage.out
    
    log_success "クリーンアップが完了しました"
}

# メイン実行フロー
main() {
    # 事前チェック
    check_environment
    check_database
    check_dependencies
    
    if [[ "$OPENAPI_ONLY" != "true" ]]; then
        check_openapi_spec
    fi
    
    # テスト実行
    test_failed=false
    
    if [[ "$OPENAPI_ONLY" != "true" ]]; then
        if ! run_contract_tests; then
            test_failed=true
        fi
    fi
    
    if [[ "$CONTRACT_ONLY" != "true" ]]; then
        if ! run_openapi_tests; then
            test_failed=true
        fi
    fi
    
    # カバレッジレポート生成
    generate_coverage_report
    
    # 結果の表示
    if [[ "$test_failed" == "true" ]]; then
        log_error "契約テストの実行中にエラーが発生しました"
        cleanup
        exit 1
    else
        log_success "全ての契約テストが正常に完了しました"
        cleanup
        exit 0
    fi
}

# エラー時のクリーンアップ
trap cleanup EXIT

# メイン実行
main "$@"
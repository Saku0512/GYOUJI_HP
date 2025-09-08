#!/bin/bash

# 強化された統合テスト実行スクリプト
# このスクリプトは、全ての統合テスト、契約テスト、エンドツーエンドテストを実行します

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

log_section() {
    echo -e "\033[1;36m[SECTION]\033[0m $1"
    echo "=================================================="
}

# 使用方法を表示
show_usage() {
    echo "使用方法: $0 [オプション]"
    echo ""
    echo "オプション:"
    echo "  --coverage          カバレッジレポートを生成"
    echo "  --verbose           詳細ログを出力"
    echo "  --parallel          並列テスト実行"
    echo "  --integration-only  統合テストのみ実行"
    echo "  --contract-only     契約テストのみ実行"
    echo "  --e2e-only          E2Eテストのみ実行"
    echo "  --auth-only         認証テストのみ実行"
    echo "  --error-only        エラーケーステストのみ実行"
    echo "  --quick             クイックテスト（基本テストのみ）"
    echo "  --help              このヘルプを表示"
    echo ""
    echo "例:"
    echo "  $0                  全てのテストを実行"
    echo "  $0 --coverage       カバレッジ付きで実行"
    echo "  $0 --parallel       並列実行で高速化"
    echo "  $0 --quick          基本テストのみ実行"
}

# デフォルト設定
COVERAGE=false
VERBOSE=false
PARALLEL=false
INTEGRATION_ONLY=false
CONTRACT_ONLY=false
E2E_ONLY=false
AUTH_ONLY=false
ERROR_ONLY=false
QUICK=false

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
        --parallel)
            PARALLEL=true
            shift
            ;;
        --integration-only)
            INTEGRATION_ONLY=true
            shift
            ;;
        --contract-only)
            CONTRACT_ONLY=true
            shift
            ;;
        --e2e-only)
            E2E_ONLY=true
            shift
            ;;
        --auth-only)
            AUTH_ONLY=true
            shift
            ;;
        --error-only)
            ERROR_ONLY=true
            shift
            ;;
        --quick)
            QUICK=true
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

log_section "強化された統合テスト実行を開始します"
log_info "プロジェクトディレクトリ: $PROJECT_ROOT"

# 環境変数の確認
check_environment() {
    log_info "環境変数を確認しています..."
    
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
        exit 1
    fi
    
    log_success "環境変数の確認が完了しました"
}

# データベース接続の確認
check_database() {
    log_info "データベース接続を確認しています..."
    
    if command -v mysql &> /dev/null; then
        if mysql -h"$DB_HOST" -P"$DB_PORT" -u"$DB_USER" -p"$DB_PASSWORD" -e "SELECT 1;" &> /dev/null; then
            log_success "データベース接続が確認できました"
        else
            log_error "データベースに接続できません"
            exit 1
        fi
    else
        log_warning "MySQLクライアントが見つかりません。データベース接続確認をスキップします"
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

# テスト引数の構築
build_test_args() {
    local test_args="-v"
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_args="$test_args -test.v"
    fi
    
    if [[ "$PARALLEL" == "true" ]]; then
        test_args="$test_args -parallel 4"
    fi
    
    if [[ "$COVERAGE" == "true" ]]; then
        test_args="$test_args -coverprofile=coverage.out"
    fi
    
    echo "$test_args"
}

# 統合テストの実行
run_integration_tests() {
    log_section "統合テスト実行"
    
    local test_args=$(build_test_args)
    local test_pattern=""
    
    if [[ "$QUICK" == "true" ]]; then
        test_pattern="-run TestAuthIntegrationTestSuite|TestTournamentIntegrationTestSuite"
    else
        test_pattern="-run TestAuthIntegrationTestSuite|TestTournamentIntegrationTestSuite|TestMatchIntegrationTestSuite|TestWorkflowIntegrationTestSuite"
    fi
    
    log_info "統合テストを実行しています..."
    
    if go test $test_args $test_pattern ./integration_test/; then
        log_success "統合テストが完了しました"
        return 0
    else
        log_error "統合テストが失敗しました"
        return 1
    fi
}

# 契約テストの実行
run_contract_tests() {
    log_section "契約テスト実行"
    
    local test_args=$(build_test_args)
    
    log_info "契約テストを実行しています..."
    
    if go test $test_args -run "TestContractTestSuite|TestOpenAPIValidationTestSuite" ./integration_test/; then
        log_success "契約テストが完了しました"
        return 0
    else
        log_error "契約テストが失敗しました"
        return 1
    fi
}

# エンドツーエンドテストの実行
run_e2e_tests() {
    log_section "エンドツーエンドテスト実行"
    
    local test_args=$(build_test_args)
    
    log_info "エンドツーエンドテストを実行しています..."
    
    if go test $test_args -run "TestE2EAPITestSuite" ./integration_test/; then
        log_success "エンドツーエンドテストが完了しました"
        return 0
    else
        log_error "エンドツーエンドテストが失敗しました"
        return 1
    fi
}

# 認証フローテストの実行
run_auth_flow_tests() {
    log_section "認証フローテスト実行"
    
    local test_args=$(build_test_args)
    
    log_info "認証フローテストを実行しています..."
    
    if go test $test_args -run "TestAuthFlowTestSuite" ./integration_test/; then
        log_success "認証フローテストが完了しました"
        return 0
    else
        log_error "認証フローテストが失敗しました"
        return 1
    fi
}

# エラーケーステストの実行
run_error_cases_tests() {
    log_section "エラーケーステスト実行"
    
    local test_args=$(build_test_args)
    
    log_info "エラーケーステストを実行しています..."
    
    if go test $test_args -run "TestErrorCasesTestSuite" ./integration_test/; then
        log_success "エラーケーステストが完了しました"
        return 0
    else
        log_error "エラーケーステストが失敗しました"
        return 1
    fi
}

# カバレッジレポートの生成
generate_coverage_report() {
    if [[ "$COVERAGE" != "true" ]]; then
        return 0
    fi
    
    log_section "カバレッジレポート生成"
    
    if [[ -f "coverage.out" ]]; then
        log_info "カバレッジレポートを生成しています..."
        
        # HTMLレポートの生成
        if go tool cover -html=coverage.out -o coverage.html; then
            log_success "カバレッジレポートが生成されました: coverage.html"
        else
            log_error "カバレッジレポートの生成に失敗しました"
        fi
        
        # カバレッジ率の表示
        coverage_percent=$(go tool cover -func=coverage.out | tail -1 | awk '{print $3}')
        log_info "総合カバレッジ: $coverage_percent"
        
        # カバレッジ閾値の確認
        coverage_num=$(echo $coverage_percent | sed 's/%//')
        if (( $(echo "$coverage_num >= 80" | bc -l) )); then
            log_success "カバレッジが目標値（80%）を達成しています"
        else
            log_warning "カバレッジが目標値（80%）を下回っています: $coverage_percent"
        fi
    else
        log_warning "カバレッジファイルが見つかりません"
    fi
}

# テスト結果のサマリー表示
show_test_summary() {
    log_section "テスト結果サマリー"
    
    local total_tests=0
    local passed_tests=0
    local failed_tests=0
    
    # 実行されたテストの集計（簡易版）
    if [[ "$INTEGRATION_ONLY" == "true" ]] || [[ "$QUICK" == "true" ]] || [[ "$AUTH_ONLY" == "false" && "$CONTRACT_ONLY" == "false" && "$E2E_ONLY" == "false" && "$ERROR_ONLY" == "false" ]]; then
        total_tests=$((total_tests + 1))
        if [[ "${test_results[integration]}" == "0" ]]; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
    fi
    
    if [[ "$CONTRACT_ONLY" == "true" ]] || [[ "$AUTH_ONLY" == "false" && "$INTEGRATION_ONLY" == "false" && "$E2E_ONLY" == "false" && "$ERROR_ONLY" == "false" ]]; then
        total_tests=$((total_tests + 1))
        if [[ "${test_results[contract]}" == "0" ]]; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
    fi
    
    if [[ "$E2E_ONLY" == "true" ]] || [[ "$AUTH_ONLY" == "false" && "$INTEGRATION_ONLY" == "false" && "$CONTRACT_ONLY" == "false" && "$ERROR_ONLY" == "false" ]]; then
        total_tests=$((total_tests + 1))
        if [[ "${test_results[e2e]}" == "0" ]]; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
    fi
    
    if [[ "$AUTH_ONLY" == "true" ]] || [[ "$INTEGRATION_ONLY" == "false" && "$CONTRACT_ONLY" == "false" && "$E2E_ONLY" == "false" && "$ERROR_ONLY" == "false" ]]; then
        total_tests=$((total_tests + 1))
        if [[ "${test_results[auth]}" == "0" ]]; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
    fi
    
    if [[ "$ERROR_ONLY" == "true" ]] || [[ "$AUTH_ONLY" == "false" && "$INTEGRATION_ONLY" == "false" && "$CONTRACT_ONLY" == "false" && "$E2E_ONLY" == "false" ]]; then
        total_tests=$((total_tests + 1))
        if [[ "${test_results[error]}" == "0" ]]; then
            passed_tests=$((passed_tests + 1))
        else
            failed_tests=$((failed_tests + 1))
        fi
    fi
    
    echo "テストスイート実行結果:"
    echo "  総テストスイート数: $total_tests"
    echo "  成功: $passed_tests"
    echo "  失敗: $failed_tests"
    echo ""
    
    if [[ "$failed_tests" -eq 0 ]]; then
        log_success "全てのテストが正常に完了しました！"
    else
        log_error "$failed_tests 個のテストスイートが失敗しました"
    fi
}

# クリーンアップ
cleanup() {
    log_info "クリーンアップを実行しています..."
    
    # 一時ファイルの削除
    rm -f coverage.out
    
    log_success "クリーンアップが完了しました"
}

# メイン実行フロー
main() {
    # 事前チェック
    check_environment
    check_database
    check_dependencies
    
    # テスト結果を格納する連想配列
    declare -A test_results
    
    # テスト実行
    if [[ "$INTEGRATION_ONLY" == "true" ]]; then
        run_integration_tests
        test_results[integration]=$?
    elif [[ "$CONTRACT_ONLY" == "true" ]]; then
        run_contract_tests
        test_results[contract]=$?
    elif [[ "$E2E_ONLY" == "true" ]]; then
        run_e2e_tests
        test_results[e2e]=$?
    elif [[ "$AUTH_ONLY" == "true" ]]; then
        run_auth_flow_tests
        test_results[auth]=$?
    elif [[ "$ERROR_ONLY" == "true" ]]; then
        run_error_cases_tests
        test_results[error]=$?
    else
        # 全てのテストを実行
        run_integration_tests
        test_results[integration]=$?
        
        if [[ "$QUICK" != "true" ]]; then
            run_contract_tests
            test_results[contract]=$?
            
            run_e2e_tests
            test_results[e2e]=$?
            
            run_auth_flow_tests
            test_results[auth]=$?
            
            run_error_cases_tests
            test_results[error]=$?
        fi
    fi
    
    # カバレッジレポート生成
    generate_coverage_report
    
    # テスト結果サマリー表示
    show_test_summary
    
    # 終了コードの決定
    local exit_code=0
    for result in "${test_results[@]}"; do
        if [[ "$result" != "0" ]]; then
            exit_code=1
            break
        fi
    done
    
    cleanup
    exit $exit_code
}

# エラー時のクリーンアップ
trap cleanup EXIT

# メイン実行
main "$@"
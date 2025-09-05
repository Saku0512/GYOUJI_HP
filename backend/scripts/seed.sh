#!/bin/bash

# トーナメントデータベースシーディングスクリプト
# 使用方法: ./scripts/seed.sh [オプション]

set -e

# スクリプトのディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# 色付きメッセージ用の関数
print_info() {
    echo -e "\033[1;34m[INFO]\033[0m $1"
}

print_success() {
    echo -e "\033[1;32m[SUCCESS]\033[0m $1"
}

print_error() {
    echo -e "\033[1;31m[ERROR]\033[0m $1"
}

print_warning() {
    echo -e "\033[1;33m[WARNING]\033[0m $1"
}

# ヘルプメッセージ
show_help() {
    echo "トーナメントデータベースシーディングスクリプト"
    echo ""
    echo "使用方法:"
    echo "  ./scripts/seed.sh [オプション]"
    echo ""
    echo "オプション:"
    echo "  --reset                既存データをリセットしてから実行"
    echo "  --sport=SPORT         特定のスポーツのみシーディング"
    echo "                        (volleyball, table_tennis, soccer)"
    echo "  --sql                 SQLファイルを直接実行（Go版の代わりに）"
    echo "  --admin-only          管理者ユーザーのみ作成"
    echo "  --help                このヘルプを表示"
    echo ""
    echo "例:"
    echo "  ./scripts/seed.sh                           # 全トーナメントを初期化"
    echo "  ./scripts/seed.sh --reset                   # リセット後に全初期化"
    echo "  ./scripts/seed.sh --sport=volleyball        # バレーボールのみ初期化"
    echo "  ./scripts/seed.sh --sql                     # SQLファイルで直接実行"
    echo "  ./scripts/seed.sh --admin-only              # 管理者ユーザーのみ作成"
}

# デフォルト値
RESET=false
SPORT=""
USE_SQL=false
ADMIN_ONLY=false

# 引数の解析
while [[ $# -gt 0 ]]; do
    case $1 in
        --reset)
            RESET=true
            shift
            ;;
        --sport=*)
            SPORT="${1#*=}"
            shift
            ;;
        --sql)
            USE_SQL=true
            shift
            ;;
        --admin-only)
            ADMIN_ONLY=true
            shift
            ;;
        --help)
            show_help
            exit 0
            ;;
        *)
            print_error "不明なオプション: $1"
            show_help
            exit 1
            ;;
    esac
done

# プロジェクトルートに移動
cd "$PROJECT_ROOT"

# 環境変数の確認
if [[ ! -f ".env" ]]; then
    print_warning ".envファイルが見つかりません。.env.sampleを参考に作成してください。"
fi

# 管理者ユーザーのみ作成する場合
if [[ "$ADMIN_ONLY" == true ]]; then
    print_info "管理者ユーザーを作成しています..."
    
    # MySQLに接続してユーザーを作成
    if command -v mysql &> /dev/null; then
        mysql -u root -p -e "
        USE tournament_db;
        INSERT INTO users (username, password, role) VALUES 
        ('admin', '\$2a\$10\$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi', 'admin')
        ON DUPLICATE KEY UPDATE 
            password = VALUES(password),
            role = VALUES(role);
        "
        print_success "管理者ユーザーが作成されました（ユーザー名: admin, パスワード: admin123）"
    else
        print_error "MySQLクライアントが見つかりません"
        exit 1
    fi
    exit 0
fi

# SQLファイルを直接実行する場合
if [[ "$USE_SQL" == true ]]; then
    print_info "SQLファイルを使用してシーディングを実行しています..."
    
    if command -v mysql &> /dev/null; then
        # データベースに接続してシーディング実行
        mysql -u root -p tournament_db < migrations/seeds/seed_all.sql
        print_success "SQLシーディングが完了しました"
    else
        print_error "MySQLクライアントが見つかりません"
        exit 1
    fi
    exit 0
fi

# Go版のシーディングツールを使用
print_info "Goシーディングツールを使用してシーディングを実行しています..."

# シーディングコマンドの構築
SEED_CMD="go run cmd/seed/main.go"

if [[ "$RESET" == true ]]; then
    SEED_CMD="$SEED_CMD -reset"
fi

if [[ -n "$SPORT" ]]; then
    SEED_CMD="$SEED_CMD -sport=$SPORT"
fi

# シーディング実行
print_info "実行コマンド: $SEED_CMD"
eval "$SEED_CMD"

print_success "シーディングが正常に完了しました"

# 結果の表示
print_info "データベースの状態を確認しています..."
go run -c "
package main

import (
    \"fmt\"
    \"backend/internal/config\"
    \"backend/internal/database\"
    \"backend/internal/repository\"
)

func main() {
    cfg, _ := config.Load()
    db, _ := database.NewConnection(cfg.Database)
    defer db.Close()
    
    repo := repository.NewRepository(db)
    
    users, _ := repo.User.GetAll()
    tournaments, _ := repo.Tournament.GetAll()
    
    fmt.Printf(\"ユーザー数: %d\\n\", len(users))
    fmt.Printf(\"トーナメント数: %d\\n\", len(tournaments))
    
    for _, tournament := range tournaments {
        matches, _ := repo.Match.GetByTournament(tournament.ID)
        fmt.Printf(\"- %s: %d試合\\n\", tournament.Sport, len(matches))
    }
}
" 2>/dev/null || print_warning "データベース状態の確認に失敗しました"
#!/bin/bash
# トーナメントシステム デプロイメントスクリプト
# 開発環境と本番環境の設定分離

set -e

# 色付きログ出力
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 使用方法を表示
show_usage() {
    echo "使用方法: $0 [ENVIRONMENT] [ACTION]"
    echo ""
    echo "ENVIRONMENT:"
    echo "  dev        開発環境"
    echo "  prod       本番環境"
    echo ""
    echo "ACTION:"
    echo "  up         サービスを起動"
    echo "  down       サービスを停止"
    echo "  restart    サービスを再起動"
    echo "  build      イメージを再ビルド"
    echo "  logs       ログを表示"
    echo "  status     サービス状態を表示"
    echo "  clean      未使用のリソースを削除"
    echo ""
    echo "例:"
    echo "  $0 dev up      # 開発環境を起動"
    echo "  $0 prod build  # 本番環境用イメージをビルド"
    echo "  $0 dev logs    # 開発環境のログを表示"
}

# 引数チェック
if [ $# -lt 2 ]; then
    log_error "引数が不足しています"
    show_usage
    exit 1
fi

ENVIRONMENT=$1
ACTION=$2

# 環境変数ファイルの設定
case $ENVIRONMENT in
    "dev"|"development")
        ENV_FILE=".env.development"
        COMPOSE_FILES="-f docker-compose.yml -f docker-compose.override.yml"
        PROJECT_NAME="tournament-dev"
        ;;
    "prod"|"production")
        ENV_FILE=".env.production"
        COMPOSE_FILES="-f docker-compose.yml -f docker-compose.prod.yml"
        PROJECT_NAME="tournament-prod"
        ;;
    *)
        log_error "無効な環境: $ENVIRONMENT"
        show_usage
        exit 1
        ;;
esac

# 環境変数ファイルの存在確認
if [ ! -f "$ENV_FILE" ]; then
    log_error "環境変数ファイルが見つかりません: $ENV_FILE"
    exit 1
fi

log_info "環境: $ENVIRONMENT"
log_info "環境変数ファイル: $ENV_FILE"
log_info "プロジェクト名: $PROJECT_NAME"

# Docker Composeコマンドの基本形
DOCKER_COMPOSE="docker-compose --env-file $ENV_FILE -p $PROJECT_NAME $COMPOSE_FILES"

# アクション実行
case $ACTION in
    "up")
        log_info "サービスを起動しています..."
        
        # ネットワーク設定
        if [ -f "./docker-network-setup.sh" ]; then
            log_info "Docker ネットワークを設定しています..."
            ./docker-network-setup.sh
        fi
        
        # サービス起動
        $DOCKER_COMPOSE up -d
        
        log_success "サービスが起動しました"
        
        # サービス状態確認
        log_info "サービス状態:"
        $DOCKER_COMPOSE ps
        
        # ヘルスチェック待機
        log_info "ヘルスチェックを待機しています..."
        sleep 30
        
        # 接続テスト
        if [ "$ENVIRONMENT" = "dev" ] || [ "$ENVIRONMENT" = "development" ]; then
            log_info "開発環境の接続テスト:"
            echo "  フロントエンド: http://localhost:5173"
            echo "  バックエンド: http://localhost:8081/api"
        else
            log_info "本番環境が起動しました"
            echo "  URL: https://nitsche-gyouji.com"
        fi
        ;;
        
    "down")
        log_info "サービスを停止しています..."
        $DOCKER_COMPOSE down
        log_success "サービスが停止しました"
        ;;
        
    "restart")
        log_info "サービスを再起動しています..."
        $DOCKER_COMPOSE restart
        log_success "サービスが再起動しました"
        ;;
        
    "build")
        log_info "イメージを再ビルドしています..."
        
        # フロントエンドビルド
        if [ -f "./frontend/docker-build.sh" ]; then
            log_info "フロントエンドイメージをビルドしています..."
            cd frontend
            ./docker-build.sh latest production
            cd ..
        fi
        
        # 全体ビルド
        $DOCKER_COMPOSE build --no-cache
        log_success "イメージのビルドが完了しました"
        ;;
        
    "logs")
        log_info "ログを表示しています..."
        $DOCKER_COMPOSE logs -f --tail=100
        ;;
        
    "status")
        log_info "サービス状態:"
        $DOCKER_COMPOSE ps
        
        log_info "ネットワーク状態:"
        docker network ls --filter "label=project=tournament"
        
        log_info "ボリューム状態:"
        docker volume ls --filter "label=project=tournament"
        ;;
        
    "clean")
        log_warning "未使用のリソースを削除します。続行しますか？ (y/N)"
        read -r response
        if [[ "$response" =~ ^[Yy]$ ]]; then
            log_info "未使用のリソースを削除しています..."
            
            # サービス停止
            $DOCKER_COMPOSE down -v --remove-orphans
            
            # 未使用イメージ削除
            docker image prune -f
            
            # 未使用ボリューム削除
            docker volume prune -f
            
            # 未使用ネットワーク削除
            docker network prune -f
            
            log_success "クリーンアップが完了しました"
        else
            log_info "クリーンアップをキャンセルしました"
        fi
        ;;
        
    *)
        log_error "無効なアクション: $ACTION"
        show_usage
        exit 1
        ;;
esac
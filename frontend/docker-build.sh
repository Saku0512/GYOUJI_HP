#!/bin/bash
# フロントエンド Docker ビルドスクリプト
# セキュリティチェックと最適化を含む

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

# 設定
IMAGE_NAME="tournament-frontend"
TAG=${1:-latest}
BUILD_TARGET=${2:-production}

log_info "フロントエンド Docker イメージをビルドしています..."
log_info "イメージ名: ${IMAGE_NAME}:${TAG}"
log_info "ビルドターゲット: ${BUILD_TARGET}"

# セキュリティチェック: npm audit
log_info "セキュリティ脆弱性をチェックしています..."
if npm audit --audit-level moderate; then
    log_success "セキュリティチェック完了"
else
    log_warning "セキュリティ脆弱性が検出されました。続行しますか？ (y/N)"
    read -r response
    if [[ ! "$response" =~ ^[Yy]$ ]]; then
        log_error "ビルドを中止しました"
        exit 1
    fi
fi

# Dockerイメージビルド
log_info "Dockerイメージをビルドしています..."
if docker build \
    --target ${BUILD_TARGET} \
    --tag ${IMAGE_NAME}:${TAG} \
    --build-arg NODE_ENV=production \
    --build-arg BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ') \
    --build-arg VCS_REF=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown") \
    --no-cache \
    .; then
    log_success "Dockerイメージのビルドが完了しました"
else
    log_error "Dockerイメージのビルドに失敗しました"
    exit 1
fi

# イメージサイズ確認
log_info "ビルドされたイメージのサイズ:"
docker images ${IMAGE_NAME}:${TAG} --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}"

# セキュリティスキャン（Trivyがインストールされている場合）
if command -v trivy &> /dev/null; then
    log_info "セキュリティスキャンを実行しています..."
    if trivy image --severity HIGH,CRITICAL ${IMAGE_NAME}:${TAG}; then
        log_success "セキュリティスキャン完了"
    else
        log_warning "セキュリティスキャンで問題が検出されました"
    fi
else
    log_warning "Trivyがインストールされていません。セキュリティスキャンをスキップします"
fi

# ビルド完了
log_success "フロントエンド Docker イメージのビルドが完了しました"
log_info "イメージを実行するには: docker run -p 80:80 ${IMAGE_NAME}:${TAG}"

# 開発用イメージのビルド（オプション）
if [[ "$BUILD_TARGET" == "production" ]]; then
    log_info "開発用イメージもビルドしますか？ (y/N)"
    read -r response
    if [[ "$response" =~ ^[Yy]$ ]]; then
        log_info "開発用イメージをビルドしています..."
        if docker build \
            --target development \
            --tag ${IMAGE_NAME}:dev \
            --build-arg NODE_ENV=development \
            .; then
            log_success "開発用イメージのビルドが完了しました"
            log_info "開発用イメージを実行するには: docker run -p 5173:5173 -v \$(pwd):/app ${IMAGE_NAME}:dev"
        else
            log_error "開発用イメージのビルドに失敗しました"
        fi
    fi
fi
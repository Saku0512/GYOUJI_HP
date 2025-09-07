#!/bin/bash
# Docker ネットワーク設定スクリプト
# フロントエンド、バックエンド、データベース間のネットワーク設定

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

# ネットワーク設定
NETWORK_WEB="tournament_web"
NETWORK_BACKEND="tournament_backend"
NETWORK_MONITORING="tournament_monitoring"

log_info "Docker ネットワークを設定しています..."

# 既存のネットワークをチェック
check_network() {
    local network_name=$1
    if docker network ls --format "{{.Name}}" | grep -q "^${network_name}$"; then
        return 0
    else
        return 1
    fi
}

# Webネットワーク（フロントエンド、リバースプロキシ用）
if check_network "$NETWORK_WEB"; then
    log_warning "ネットワーク '$NETWORK_WEB' は既に存在します"
else
    log_info "Webネットワーク '$NETWORK_WEB' を作成しています..."
    docker network create \
        --driver bridge \
        --subnet=172.20.0.0/16 \
        --ip-range=172.20.240.0/20 \
        --gateway=172.20.0.1 \
        --opt com.docker.network.bridge.name=br-tournament-web \
        --opt com.docker.network.driver.mtu=1500 \
        --label project=tournament \
        --label environment=production \
        "$NETWORK_WEB"
    log_success "Webネットワークを作成しました"
fi

# バックエンドネットワーク（バックエンド、データベース用）
if check_network "$NETWORK_BACKEND"; then
    log_warning "ネットワーク '$NETWORK_BACKEND' は既に存在します"
else
    log_info "バックエンドネットワーク '$NETWORK_BACKEND' を作成しています..."
    docker network create \
        --driver bridge \
        --subnet=172.21.0.0/16 \
        --ip-range=172.21.240.0/20 \
        --gateway=172.21.0.1 \
        --opt com.docker.network.bridge.name=br-tournament-backend \
        --opt com.docker.network.driver.mtu=1500 \
        --internal \
        --label project=tournament \
        --label environment=production \
        "$NETWORK_BACKEND"
    log_success "バックエンドネットワークを作成しました"
fi

# 監視ネットワーク（将来の拡張用）
if check_network "$NETWORK_MONITORING"; then
    log_warning "ネットワーク '$NETWORK_MONITORING' は既に存在します"
else
    log_info "監視ネットワーク '$NETWORK_MONITORING' を作成しています..."
    docker network create \
        --driver bridge \
        --subnet=172.22.0.0/16 \
        --ip-range=172.22.240.0/20 \
        --gateway=172.22.0.1 \
        --opt com.docker.network.bridge.name=br-tournament-monitoring \
        --opt com.docker.network.driver.mtu=1500 \
        --label project=tournament \
        --label environment=production \
        "$NETWORK_MONITORING"
    log_success "監視ネットワークを作成しました"
fi

# ネットワーク情報を表示
log_info "作成されたネットワーク:"
docker network ls --filter "label=project=tournament" --format "table {{.Name}}\t{{.Driver}}\t{{.Scope}}"

log_success "Docker ネットワークの設定が完了しました"

# ネットワーク詳細情報の表示（オプション）
log_info "詳細なネットワーク情報を表示しますか？ (y/N)"
read -r response
if [[ "$response" =~ ^[Yy]$ ]]; then
    echo
    log_info "=== Webネットワーク詳細 ==="
    docker network inspect "$NETWORK_WEB" --format "{{json .IPAM}}" | jq .
    echo
    log_info "=== バックエンドネットワーク詳細 ==="
    docker network inspect "$NETWORK_BACKEND" --format "{{json .IPAM}}" | jq .
    echo
    log_info "=== 監視ネットワーク詳細 ==="
    docker network inspect "$NETWORK_MONITORING" --format "{{json .IPAM}}" | jq .
fi
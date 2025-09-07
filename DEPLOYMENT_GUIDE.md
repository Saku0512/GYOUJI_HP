# トーナメント管理システム - デプロイメントガイド

## 目次

1. [概要](#概要)
2. [システム要件](#システム要件)
3. [事前準備](#事前準備)
4. [環境変数とシークレット設定](#環境変数とシークレット設定)
5. [本番環境デプロイ手順](#本番環境デプロイ手順)
6. [SSL証明書設定](#ssl証明書設定)
7. [監視とログ設定](#監視とログ設定)
8. [バックアップとリストア](#バックアップとリストア)
9. [セキュリティ設定](#セキュリティ設定)
10. [パフォーマンス最適化](#パフォーマンス最適化)
11. [トラブルシューティング](#トラブルシューティング)
12. [メンテナンス手順](#メンテナンス手順)

## 概要

このガイドでは、トーナメント管理システムを本番環境にデプロイする手順を説明します。システムはDockerコンテナベースで構築されており、以下のコンポーネントで構成されています：

- **Traefik**: リバースプロキシ・ロードバランサー（SSL終端）
- **Frontend**: SvelteKit + Nginx
- **Backend**: Go + Gin フレームワーク
- **Database**: MySQL 8.0

## システム要件

### ハードウェア要件

**最小要件**
- CPU: 2コア
- RAM: 4GB
- ストレージ: 20GB SSD
- ネットワーク: 100Mbps

**推奨要件**
- CPU: 4コア
- RAM: 8GB
- ストレージ: 50GB SSD
- ネットワーク: 1Gbps

### ソフトウェア要件

- **OS**: Ubuntu 20.04 LTS以上 / CentOS 8以上 / Debian 11以上
- **Docker**: 24.0以上
- **Docker Compose**: 2.20以上
- **Git**: 2.25以上

### ネットワーク要件

- **ポート80**: HTTP（HTTPSリダイレクト用）
- **ポート443**: HTTPS
- **ポート22**: SSH（管理用）
- **ドメイン**: SSL証明書取得用の有効なドメイン

## 事前準備

### 1. サーバーセットアップ

```bash
# システムアップデート
sudo apt update && sudo apt upgrade -y

# 必要なパッケージのインストール
sudo apt install -y curl wget git unzip htop

# Dockerのインストール
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# Docker Composeのインストール
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# 非rootユーザーをdockerグループに追加
sudo usermod -aG docker $USER

# 再ログインまたは以下を実行
newgrp docker
```

### 2. ファイアウォール設定

```bash
# UFWの有効化
sudo ufw enable

# 必要なポートを開放
sudo ufw allow 22/tcp    # SSH
sudo ufw allow 80/tcp    # HTTP
sudo ufw allow 443/tcp   # HTTPS

# 設定確認
sudo ufw status
```

### 3. プロジェクトのクローン

```bash
# プロジェクトをクローン
git clone <repository-url> tournament-system
cd tournament-system

# 本番ブランチに切り替え（存在する場合）
git checkout production
```

## 環境変数とシークレット設定

### 1. 環境変数ファイルの作成

```bash
# サンプルファイルをコピー
cp .env.sample .env

# 本番用環境変数ファイルを作成
cp .env.sample .env.production
```

### 2. 環境変数の設定

`.env.production`ファイルを編集：

```bash
# SSL証明書用メールアドレス（Let's Encrypt）
LETSENCRYPT_EMAIL=admin@yourdomain.com

# データベース設定
DB_ROOT_PASSWORD=your_secure_root_password_here_32_chars_min
DB_USER=tournament_user
DB_PASSWORD=your_secure_database_password_here_32_chars_min
DB_NAME=tournament_db

# JWT設定（本番環境では必ず変更）
JWT_SECRET=your-super-secure-jwt-secret-key-minimum-64-characters-for-production-security
JWT_EXPIRATION_HOURS=24
JWT_ISSUER=tournament-backend-production

# ビルド情報（オプション）
BUILD_DATE=$(date -u +'%Y-%m-%dT%H:%M:%SZ')
VCS_REF=$(git rev-parse --short HEAD)
```

### 3. セキュアな秘密鍵の生成

```bash
# JWT秘密鍵の生成（64文字以上推奨）
openssl rand -base64 64

# データベースパスワードの生成
openssl rand -base64 32

# 生成された値を.env.productionに設定
```

### 4. 環境変数ファイルの権限設定

```bash
# 環境変数ファイルのセキュリティ設定
chmod 600 .env.production
chown $USER:$USER .env.production

# 他のユーザーからの読み取りを禁止
sudo chattr +i .env.production  # 変更不可にする（オプション）
```

## 本番環境デプロイ手順

### 1. 本番用設定の確認

```bash
# Docker Composeファイルの構文チェック
docker-compose -f docker-compose.yml -f docker-compose.prod.yml config

# 環境変数の確認
docker-compose -f docker-compose.yml -f docker-compose.prod.yml config | grep -E "(MYSQL_|JWT_|DB_)"
```

### 2. データベースの初期化

```bash
# データベースボリュームの作成
docker volume create mysql_prod_data

# データベースのみ起動（初期化用）
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d mysql

# データベースの起動確認
docker-compose logs mysql

# 初期化完了まで待機（約30秒）
sleep 30
```

### 3. アプリケーションのビルドとデプロイ

```bash
# 本番環境用にビルド・デプロイ
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build

# デプロイ状況の確認
docker-compose ps

# ログの確認
docker-compose logs -f
```

### 4. デプロイ後の動作確認

```bash
# ヘルスチェック
curl -f http://localhost/health
curl -f http://localhost/api/health

# サービス状態確認
docker-compose ps
docker stats --no-stream
```

## SSL証明書設定

### 1. Let's Encrypt証明書の自動取得

システムは自動的にLet's Encrypt証明書を取得しますが、以下の条件が必要です：

```bash
# ドメインがサーバーIPを正しく指していることを確認
nslookup yourdomain.com

# 証明書取得の確認
docker-compose logs traefik | grep -i certificate

# 証明書ファイルの確認
ls -la letsencrypt/
```

### 2. 手動での証明書設定（必要な場合）

```bash
# 既存の証明書がある場合
mkdir -p letsencrypt
cp your-certificate.crt letsencrypt/
cp your-private-key.key letsencrypt/

# 権限設定
chmod 600 letsencrypt/*
```

### 3. 証明書の更新

Let's Encryptの証明書は自動更新されますが、手動更新も可能：

```bash
# Traefikコンテナの再起動（証明書更新）
docker-compose restart traefik

# 証明書の有効期限確認
openssl x509 -in letsencrypt/yourdomain.com.crt -text -noout | grep "Not After"
```

## 監視とログ設定

### 1. ログ設定

#### システムログの設定

```bash
# ログローテーション設定
sudo tee /etc/logrotate.d/tournament-system << EOF
/var/log/tournament-system/*.log {
    daily
    missingok
    rotate 30
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        docker-compose restart traefik frontend backend
    endscript
}
EOF
```

#### Dockerログの設定

```bash
# Docker daemon設定
sudo tee /etc/docker/daemon.json << EOF
{
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "10m",
        "max-file": "3"
    }
}
EOF

# Dockerサービス再起動
sudo systemctl restart docker
```

### 2. 監視スクリプト

#### ヘルスチェックスクリプト

```bash
# ヘルスチェックスクリプトの作成
sudo tee /usr/local/bin/tournament-health-check.sh << 'EOF'
#!/bin/bash

# 設定
DOMAIN="yourdomain.com"
LOG_FILE="/var/log/tournament-health.log"
ALERT_EMAIL="admin@yourdomain.com"

# ログ関数
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" >> $LOG_FILE
}

# ヘルスチェック関数
check_service() {
    local service=$1
    local url=$2
    
    if curl -f -s --max-time 10 "$url" > /dev/null; then
        log "OK: $service is healthy"
        return 0
    else
        log "ERROR: $service is not responding"
        return 1
    fi
}

# メイン処理
log "Starting health check"

# フロントエンドチェック
if ! check_service "Frontend" "https://$DOMAIN/health"; then
    # アラート送信（mailコマンドが設定されている場合）
    echo "Frontend service is down" | mail -s "Tournament System Alert" $ALERT_EMAIL 2>/dev/null
fi

# バックエンドチェック
if ! check_service "Backend" "https://$DOMAIN/api/health"; then
    echo "Backend service is down" | mail -s "Tournament System Alert" $ALERT_EMAIL 2>/dev/null
fi

# データベースチェック
if ! docker-compose exec -T mysql mysqladmin ping -h localhost -u root -p$DB_ROOT_PASSWORD > /dev/null 2>&1; then
    log "ERROR: Database is not responding"
    echo "Database service is down" | mail -s "Tournament System Alert" $ALERT_EMAIL 2>/dev/null
else
    log "OK: Database is healthy"
fi

log "Health check completed"
EOF

# 実行権限付与
sudo chmod +x /usr/local/bin/tournament-health-check.sh
```

#### Cronジョブの設定

```bash
# Cronジョブを追加
crontab -e

# 以下を追加（5分ごとにヘルスチェック）
*/5 * * * * /usr/local/bin/tournament-health-check.sh

# 毎日午前2時にログクリーンアップ
0 2 * * * find /var/log -name "*.log" -mtime +30 -delete
```

### 3. パフォーマンス監視

#### システムリソース監視

```bash
# システム監視スクリプト
sudo tee /usr/local/bin/tournament-monitor.sh << 'EOF'
#!/bin/bash

LOG_FILE="/var/log/tournament-monitor.log"

# システムリソース情報を記録
{
    echo "=== $(date) ==="
    echo "CPU Usage:"
    top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1
    
    echo "Memory Usage:"
    free -h
    
    echo "Disk Usage:"
    df -h
    
    echo "Docker Stats:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
    
    echo "Active Connections:"
    netstat -an | grep :443 | wc -l
    
    echo "========================"
} >> $LOG_FILE
EOF

sudo chmod +x /usr/local/bin/tournament-monitor.sh

# Cronジョブに追加（15分ごと）
echo "*/15 * * * * /usr/local/bin/tournament-monitor.sh" | crontab -
```

## バックアップとリストア

### 1. データベースバックアップ

#### 自動バックアップスクリプト

```bash
# バックアップスクリプトの作成
sudo tee /usr/local/bin/tournament-backup.sh << 'EOF'
#!/bin/bash

# 設定
BACKUP_DIR="/var/backups/tournament"
DB_CONTAINER="tournament-mysql"
RETENTION_DAYS=30

# バックアップディレクトリ作成
mkdir -p $BACKUP_DIR

# 日付付きファイル名
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_FILE="$BACKUP_DIR/tournament_db_$DATE.sql"

# データベースバックアップ
docker-compose exec -T mysql mysqldump \
    -u root -p$DB_ROOT_PASSWORD \
    --single-transaction \
    --routines \
    --triggers \
    $DB_NAME > $BACKUP_FILE

# 圧縮
gzip $BACKUP_FILE

# 古いバックアップを削除
find $BACKUP_DIR -name "*.sql.gz" -mtime +$RETENTION_DAYS -delete

echo "Backup completed: ${BACKUP_FILE}.gz"
EOF

sudo chmod +x /usr/local/bin/tournament-backup.sh
```

#### バックアップのスケジュール設定

```bash
# 毎日午前3時にバックアップ
echo "0 3 * * * /usr/local/bin/tournament-backup.sh" | crontab -
```

### 2. データベースリストア

```bash
# バックアップからのリストア
BACKUP_FILE="/var/backups/tournament/tournament_db_20240101_030000.sql.gz"

# サービス停止
docker-compose stop backend

# データベースリストア
gunzip -c $BACKUP_FILE | docker-compose exec -T mysql mysql -u root -p$DB_ROOT_PASSWORD $DB_NAME

# サービス再開
docker-compose start backend
```

### 3. 完全システムバックアップ

```bash
# システム全体のバックアップスクリプト
sudo tee /usr/local/bin/tournament-full-backup.sh << 'EOF'
#!/bin/bash

BACKUP_DIR="/var/backups/tournament-full"
DATE=$(date +%Y%m%d_%H%M%S)
BACKUP_NAME="tournament_full_$DATE"

mkdir -p $BACKUP_DIR

# アプリケーションファイル
tar -czf "$BACKUP_DIR/${BACKUP_NAME}_app.tar.gz" \
    --exclude='node_modules' \
    --exclude='.git' \
    --exclude='*.log' \
    /path/to/tournament-system/

# Dockerボリューム
docker run --rm -v mysql_prod_data:/data -v $BACKUP_DIR:/backup \
    alpine tar -czf /backup/${BACKUP_NAME}_mysql_data.tar.gz -C /data .

# 設定ファイル
tar -czf "$BACKUP_DIR/${BACKUP_NAME}_config.tar.gz" \
    /etc/nginx/ \
    /etc/ssl/ \
    letsencrypt/

echo "Full backup completed: $BACKUP_NAME"
EOF

sudo chmod +x /usr/local/bin/tournament-full-backup.sh
```

## セキュリティ設定

### 1. ファイアウォール詳細設定

```bash
# 詳細なファイアウォール設定
sudo ufw --force reset
sudo ufw default deny incoming
sudo ufw default allow outgoing

# SSH（ポート変更推奨）
sudo ufw allow 22/tcp

# HTTP/HTTPS
sudo ufw allow 80/tcp
sudo ufw allow 443/tcp

# 特定IPからの管理アクセス（必要に応じて）
# sudo ufw allow from YOUR_ADMIN_IP to any port 22

sudo ufw enable
```

### 2. SSH セキュリティ強化

```bash
# SSH設定の強化
sudo tee -a /etc/ssh/sshd_config << EOF

# セキュリティ強化設定
PermitRootLogin no
PasswordAuthentication no
PubkeyAuthentication yes
AuthorizedKeysFile .ssh/authorized_keys
MaxAuthTries 3
ClientAliveInterval 300
ClientAliveCountMax 2
Protocol 2
EOF

# SSH サービス再起動
sudo systemctl restart sshd
```

### 3. 侵入検知システム

```bash
# Fail2banのインストール
sudo apt install -y fail2ban

# 設定ファイル作成
sudo tee /etc/fail2ban/jail.local << EOF
[DEFAULT]
bantime = 3600
findtime = 600
maxretry = 3

[sshd]
enabled = true
port = ssh
filter = sshd
logpath = /var/log/auth.log

[nginx-http-auth]
enabled = true
filter = nginx-http-auth
port = http,https
logpath = /var/log/nginx/error.log

[nginx-limit-req]
enabled = true
filter = nginx-limit-req
port = http,https
logpath = /var/log/nginx/error.log
maxretry = 10
EOF

# Fail2ban開始
sudo systemctl enable fail2ban
sudo systemctl start fail2ban
```

### 4. セキュリティアップデート

```bash
# 自動セキュリティアップデート設定
sudo apt install -y unattended-upgrades

# 設定
sudo tee /etc/apt/apt.conf.d/50unattended-upgrades << EOF
Unattended-Upgrade::Allowed-Origins {
    "\${distro_id}:\${distro_codename}-security";
};
Unattended-Upgrade::AutoFixInterruptedDpkg "true";
Unattended-Upgrade::MinimalSteps "true";
Unattended-Upgrade::Remove-Unused-Dependencies "true";
Unattended-Upgrade::Automatic-Reboot "false";
EOF

# 有効化
sudo systemctl enable unattended-upgrades
```

## パフォーマンス最適化

### 1. システムレベル最適化

```bash
# カーネルパラメータ最適化
sudo tee -a /etc/sysctl.conf << EOF

# ネットワーク最適化
net.core.rmem_max = 16777216
net.core.wmem_max = 16777216
net.ipv4.tcp_rmem = 4096 87380 16777216
net.ipv4.tcp_wmem = 4096 65536 16777216
net.ipv4.tcp_congestion_control = bbr

# ファイルディスクリプタ制限
fs.file-max = 65536

# メモリ最適化
vm.swappiness = 10
vm.dirty_ratio = 15
vm.dirty_background_ratio = 5
EOF

# 設定適用
sudo sysctl -p
```

### 2. Docker最適化

```bash
# Docker daemon最適化
sudo tee /etc/docker/daemon.json << EOF
{
    "log-driver": "json-file",
    "log-opts": {
        "max-size": "10m",
        "max-file": "3"
    },
    "storage-driver": "overlay2",
    "storage-opts": [
        "overlay2.override_kernel_check=true"
    ],
    "default-ulimits": {
        "nofile": {
            "Name": "nofile",
            "Hard": 64000,
            "Soft": 64000
        }
    }
}
EOF

sudo systemctl restart docker
```

### 3. データベース最適化

```bash
# MySQL設定最適化用ファイル作成
mkdir -p mysql-config

tee mysql-config/my.cnf << EOF
[mysqld]
# 基本設定
innodb_buffer_pool_size = 1G
innodb_log_file_size = 256M
innodb_flush_log_at_trx_commit = 2
innodb_flush_method = O_DIRECT

# 接続設定
max_connections = 200
max_connect_errors = 1000000

# クエリキャッシュ
query_cache_type = 1
query_cache_size = 128M

# ログ設定
slow_query_log = 1
slow_query_log_file = /var/log/mysql/slow.log
long_query_time = 2
EOF

# docker-compose.ymlのmysqlサービスにボリュームマウント追加
# volumes:
#   - ./mysql-config/my.cnf:/etc/mysql/conf.d/custom.cnf
```

## トラブルシューティング

### 1. 一般的な問題と解決方法

#### サービスが起動しない

```bash
# コンテナ状態確認
docker-compose ps

# ログ確認
docker-compose logs [service-name]

# リソース使用量確認
docker stats

# ディスク容量確認
df -h
docker system df
```

#### SSL証明書の問題

```bash
# 証明書状態確認
docker-compose logs traefik | grep -i certificate

# 手動証明書更新
docker-compose restart traefik

# 証明書ファイル確認
ls -la letsencrypt/
```

#### データベース接続エラー

```bash
# データベース接続テスト
docker-compose exec mysql mysql -u root -p$DB_ROOT_PASSWORD -e "SELECT 1;"

# データベースログ確認
docker-compose logs mysql

# 接続設定確認
docker-compose exec backend env | grep DB_
```

#### パフォーマンス問題

```bash
# システムリソース確認
htop
iotop
nethogs

# Dockerリソース確認
docker stats

# ログサイズ確認
docker system df
du -sh /var/lib/docker/
```

### 2. 緊急時対応手順

#### サービス全停止

```bash
# 緊急停止
docker-compose down

# 強制停止
docker-compose kill
```

#### ロールバック手順

```bash
# 前のバージョンに戻す
git checkout [previous-commit]
docker-compose down
docker-compose up -d --build
```

#### データベース復旧

```bash
# 最新バックアップから復旧
LATEST_BACKUP=$(ls -t /var/backups/tournament/*.sql.gz | head -1)
docker-compose stop backend
gunzip -c $LATEST_BACKUP | docker-compose exec -T mysql mysql -u root -p$DB_ROOT_PASSWORD $DB_NAME
docker-compose start backend
```

### 3. ログ分析

#### エラーログ分析

```bash
# エラーパターン検索
docker-compose logs | grep -i error
docker-compose logs | grep -i "500\|502\|503\|504"

# アクセスログ分析
docker-compose exec frontend tail -f /var/log/nginx/access.log

# パフォーマンス分析
docker-compose logs | grep -E "slow|timeout|memory"
```

## メンテナンス手順

### 1. 定期メンテナンス

#### 週次メンテナンス

```bash
#!/bin/bash
# 週次メンテナンススクリプト

echo "Starting weekly maintenance..."

# システムアップデート
sudo apt update && sudo apt upgrade -y

# Dockerイメージクリーンアップ
docker system prune -f

# ログローテーション
sudo logrotate -f /etc/logrotate.conf

# ディスク使用量チェック
df -h

echo "Weekly maintenance completed."
```

#### 月次メンテナンス

```bash
#!/bin/bash
# 月次メンテナンススクリプト

echo "Starting monthly maintenance..."

# 完全バックアップ
/usr/local/bin/tournament-full-backup.sh

# セキュリティスキャン
sudo apt install -y lynis
sudo lynis audit system

# パフォーマンスレポート
docker stats --no-stream > /var/log/docker-stats-$(date +%Y%m).log

echo "Monthly maintenance completed."
```

### 2. アップデート手順

#### アプリケーションアップデート

```bash
# 1. バックアップ作成
/usr/local/bin/tournament-backup.sh

# 2. 新しいバージョンを取得
git fetch origin
git checkout [new-version-tag]

# 3. 設定ファイル確認
diff .env.sample .env.production

# 4. ビルドとデプロイ
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build

# 5. 動作確認
curl -f https://yourdomain.com/health
curl -f https://yourdomain.com/api/health
```

#### データベースマイグレーション

```bash
# マイグレーション実行前のバックアップ
/usr/local/bin/tournament-backup.sh

# マイグレーション実行
docker-compose exec backend ./migrate up

# 結果確認
docker-compose logs backend | grep -i migration
```

### 3. 監視とアラート

#### 監視項目

- **システムリソース**: CPU、メモリ、ディスク使用量
- **ネットワーク**: 応答時間、エラー率
- **アプリケーション**: ヘルスチェック、ログエラー
- **データベース**: 接続数、クエリ実行時間
- **セキュリティ**: 不正アクセス試行、証明書有効期限

#### アラート設定

```bash
# アラート通知スクリプト
sudo tee /usr/local/bin/send-alert.sh << 'EOF'
#!/bin/bash

ALERT_TYPE=$1
MESSAGE=$2
EMAIL="admin@yourdomain.com"

# Slack通知（Webhook URLを設定）
SLACK_WEBHOOK="https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK"

# メール通知
echo "$MESSAGE" | mail -s "Tournament System Alert: $ALERT_TYPE" $EMAIL

# Slack通知
curl -X POST -H 'Content-type: application/json' \
    --data "{\"text\":\"Tournament System Alert: $ALERT_TYPE\\n$MESSAGE\"}" \
    $SLACK_WEBHOOK
EOF

sudo chmod +x /usr/local/bin/send-alert.sh
```

---

## まとめ

このデプロイメントガイドに従って、トーナメント管理システムを安全かつ効率的に本番環境にデプロイできます。

### 重要なポイント

1. **セキュリティ**: 強力なパスワード、SSL証明書、ファイアウォール設定
2. **監視**: 継続的なヘルスチェックとログ監視
3. **バックアップ**: 定期的なデータベースバックアップ
4. **メンテナンス**: 定期的なシステムアップデートとクリーンアップ

### サポート

問題が発生した場合は、以下の順序で対応してください：

1. ログの確認
2. ヘルスチェックの実行
3. このガイドのトラブルシューティングセクションを参照
4. 必要に応じて開発チームに連絡

定期的にこのガイドを更新し、新しい要件や改善点を反映してください。
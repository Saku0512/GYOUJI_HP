# Docker セットアップガイド

## 概要

このドキュメントは、トーナメント管理システムのDocker環境セットアップと運用方法について説明します。

## アーキテクチャ

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Traefik       │    │   Frontend      │    │   Backend       │
│ (Reverse Proxy) │◄──►│  (Svelte+Nginx) │◄──►│   (Go+Gin)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Web Network   │    │   Web Network   │    │ Backend Network │
│  (172.20.0.0/16)│    │  (172.20.0.0/16)│    │ (172.21.0.0/16) │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                                                        │
                                                        ▼
                                               ┌─────────────────┐
                                               │     MySQL       │
                                               │   (Database)    │
                                               └─────────────────┘
```

## 前提条件

- Docker Engine 20.10+
- Docker Compose 2.0+
- Git
- 8GB以上のRAM（推奨）
- 10GB以上の空きディスク容量

## クイックスタート

### 1. リポジトリのクローン

```bash
git clone <repository-url>
cd GYOUJI_HP
```

### 2. 環境変数の設定

#### 開発環境
```bash
cp .env.development .env
```

#### 本番環境
```bash
cp .env.production .env
# .envファイルを編集して本番用の値に変更
```

### 3. 環境の起動

#### 開発環境
```bash
./deploy.sh dev up
```

#### 本番環境
```bash
./deploy.sh prod up
```

## 詳細セットアップ

### 環境変数の設定

#### 必須の環境変数

| 変数名 | 説明 | 開発環境例 | 本番環境 |
|--------|------|------------|----------|
| `DB_ROOT_PASSWORD` | MySQLルートパスワード | `dev_root_password` | **変更必須** |
| `DB_USER` | データベースユーザー | `dev_user` | `tournament_prod_user` |
| `DB_PASSWORD` | データベースパスワード | `dev_password` | **変更必須** |
| `JWT_SECRET` | JWT署名キー | `dev-jwt-secret` | **変更必須（32文字以上）** |
| `LETSENCRYPT_EMAIL` | SSL証明書用メール | `dev@example.com` | **実際のメールアドレス** |

#### セキュリティ重要事項

⚠️ **本番環境では必ず以下を変更してください:**
- `DB_ROOT_PASSWORD`: 強力なパスワード
- `DB_PASSWORD`: 強力なパスワード  
- `JWT_SECRET`: 32文字以上のランダム文字列
- `LETSENCRYPT_EMAIL`: 有効なメールアドレス

### ネットワーク設定

システムは3つのDockerネットワークを使用します:

1. **Web Network** (`172.20.0.0/16`)
   - Traefik、Frontend間の通信
   - 外部からアクセス可能

2. **Backend Network** (`172.21.0.0/16`)
   - Backend、Database間の通信
   - 内部専用（セキュリティ強化）

3. **Monitoring Network** (`172.22.0.0/16`)
   - 将来の監視システム用（予約済み）

### サービス構成

#### Frontend (Svelte + Nginx)
- **ポート**: 80 (内部)
- **ヘルスチェック**: `/health`
- **セキュリティ**: 読み取り専用ファイルシステム、非rootユーザー
- **機能**: SPA対応、APIプロキシ、静的アセット配信

#### Backend (Go + Gin)
- **ポート**: 8080 (内部)
- **データベース**: MySQL 8.0
- **認証**: JWT
- **API**: RESTful API

#### Database (MySQL)
- **ポート**: 3306 (内部のみ)
- **ストレージ**: 永続ボリューム
- **文字セット**: utf8mb4

#### Traefik (Reverse Proxy)
- **ポート**: 80, 443, 8080 (管理画面)
- **SSL**: Let's Encrypt自動取得
- **機能**: 負荷分散、SSL終端、リダイレクト

## 運用コマンド

### デプロイメントスクリプト

```bash
# 基本構文
./deploy.sh [ENVIRONMENT] [ACTION]

# 環境
# dev  - 開発環境
# prod - 本番環境

# アクション
# up      - サービス起動
# down    - サービス停止
# restart - サービス再起動
# build   - イメージ再ビルド
# logs    - ログ表示
# status  - 状態確認
# clean   - リソース削除
```

### 使用例

```bash
# 開発環境を起動
./deploy.sh dev up

# 本番環境のログを確認
./deploy.sh prod logs

# 開発環境のイメージを再ビルド
./deploy.sh dev build

# サービス状態を確認
./deploy.sh prod status

# 未使用リソースを削除
./deploy.sh dev clean
```

### 手動Docker Composeコマンド

```bash
# 開発環境
docker-compose --env-file .env.development -f docker-compose.yml -f docker-compose.override.yml up -d

# 本番環境
docker-compose --env-file .env.production -f docker-compose.yml -f docker-compose.prod.yml up -d
```

## 開発環境

### 特徴
- ホットリロード対応
- デバッグモード有効
- ポート直接公開
- セキュリティ制限緩和
- 詳細ログ出力

### アクセス方法
- **フロントエンド**: http://localhost:5173
- **バックエンドAPI**: http://localhost:8081/api
- **データベース**: localhost:3306
- **Traefik管理画面**: http://localhost:8080

### 開発用コマンド

```bash
# フロントエンド開発サーバー起動
cd frontend
npm run dev

# バックエンド開発サーバー起動
cd backend
go run cmd/server/main.go

# テスト実行
cd frontend && npm test
cd backend && go test ./...
```

## 本番環境

### 特徴
- セキュリティ強化
- パフォーマンス最適化
- 自動SSL証明書
- ログローテーション
- リソース制限

### セキュリティ対策
- 非rootユーザー実行
- 読み取り専用ファイルシステム
- ネットワーク分離
- セキュリティヘッダー
- レート制限

### 監視とログ

```bash
# サービス状態確認
docker-compose ps

# ログ確認
docker-compose logs -f [service_name]

# リソース使用量確認
docker stats

# ヘルスチェック確認
docker-compose exec frontend wget -qO- http://localhost/health
```

## トラブルシューティング

### よくある問題

#### 1. ポート競合エラー
```bash
# 使用中のポートを確認
netstat -tulpn | grep :80
netstat -tulpn | grep :443

# 競合するプロセスを停止
sudo systemctl stop apache2
sudo systemctl stop nginx
```

#### 2. 権限エラー
```bash
# Dockerグループに追加
sudo usermod -aG docker $USER
# ログアウト・ログインが必要
```

#### 3. メモリ不足
```bash
# システムリソース確認
free -h
df -h

# 未使用Dockerリソース削除
docker system prune -a
```

#### 4. SSL証明書エラー
```bash
# Let's Encrypt証明書確認
docker-compose exec traefik cat /letsencrypt/acme.json

# 証明書再取得
docker-compose restart traefik
```

### ログ確認方法

```bash
# 全サービスのログ
./deploy.sh [env] logs

# 特定サービスのログ
docker-compose logs -f frontend
docker-compose logs -f backend
docker-compose logs -f mysql

# エラーログのみ
docker-compose logs --tail=100 | grep -i error
```

### パフォーマンス最適化

#### 1. イメージサイズ削減
```bash
# 未使用イメージ削除
docker image prune -a

# マルチステージビルド確認
docker images | grep tournament
```

#### 2. メモリ使用量最適化
```bash
# コンテナリソース制限確認
docker stats --no-stream

# メモリ制限設定（docker-compose.yml）
deploy:
  resources:
    limits:
      memory: 256M
```

#### 3. ネットワーク最適化
```bash
# ネットワーク遅延確認
docker-compose exec frontend ping backend
docker-compose exec backend ping mysql
```

## バックアップとリストア

### データベースバックアップ

```bash
# バックアップ作成
docker-compose exec mysql mysqldump -u root -p tournament_production > backup_$(date +%Y%m%d_%H%M%S).sql

# バックアップからリストア
docker-compose exec -i mysql mysql -u root -p tournament_production < backup_20240907_120000.sql
```

### ボリュームバックアップ

```bash
# ボリューム一覧確認
docker volume ls

# ボリュームバックアップ
docker run --rm -v tournament_mysql_data:/data -v $(pwd):/backup alpine tar czf /backup/mysql_backup_$(date +%Y%m%d).tar.gz -C /data .
```

## セキュリティベストプラクティス

### 1. 定期的なセキュリティ更新
```bash
# ベースイメージ更新
docker-compose pull
docker-compose up -d --force-recreate

# セキュリティスキャン（Trivyを使用）
trivy image tournament-frontend:latest
trivy image tournament-backend:latest
```

### 2. 秘密情報管理
- 環境変数ファイルを`.gitignore`に追加
- 本番環境では外部シークレット管理システムを使用
- 定期的なパスワード変更

### 3. ネットワークセキュリティ
- 不要なポート公開を避ける
- ファイアウォール設定
- VPN経由でのアクセス制限

## 参考資料

- [Docker公式ドキュメント](https://docs.docker.com/)
- [Docker Compose公式ドキュメント](https://docs.docker.com/compose/)
- [Traefik公式ドキュメント](https://doc.traefik.io/traefik/)
- [SvelteKit公式ドキュメント](https://kit.svelte.dev/)
- [Gin公式ドキュメント](https://gin-gonic.com/)

## サポート

問題が発生した場合は、以下の情報を含めてお問い合わせください:

1. 環境情報（開発/本番）
2. エラーメッセージ
3. 実行したコマンド
4. ログ出力
5. システム情報（OS、Dockerバージョンなど）
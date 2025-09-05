# Docker デプロイメントガイド

## 概要

このプロジェクトは、Docker Composeを使用してフロントエンド（Svelte）、バックエンド（Go）、データベース（MySQL）を統合したトーナメント管理システムです。

## 前提条件

- Docker Engine 20.10+
- Docker Compose 2.0+
- 適切な環境変数設定

## 環境設定

### 1. 環境変数の設定

`.env.sample`をコピーして`.env`ファイルを作成し、適切な値を設定してください：

```bash
cp .env.sample .env
```

重要な環境変数：
- `LETSENCRYPT_EMAIL`: SSL証明書用のメールアドレス
- `DB_ROOT_PASSWORD`: MySQLのrootパスワード
- `DB_USER`, `DB_PASSWORD`: アプリケーション用データベースユーザー
- `JWT_SECRET`: JWT署名用の秘密鍵（32文字以上推奨）

### 2. セキュリティ設定

本番環境では以下を必ず変更してください：
- すべてのパスワードを強力なものに変更
- JWT_SECRETを十分に長い（32文字以上）ランダムな文字列に設定
- データベースのデフォルト認証情報を変更

## デプロイメント

### 開発環境

```bash
# 開発環境での起動
docker-compose up -d

# ログの確認
docker-compose logs -f

# 停止
docker-compose down
```

### 本番環境

```bash
# 本番環境での起動
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d

# ログの確認
docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f

# 停止
docker-compose -f docker-compose.yml -f docker-compose.prod.yml down
```

## サービス構成

### ネットワーク構成

- `web`: フロントエンド、バックエンド、Traefikが接続
- `backend`: バックエンドとMySQLが接続

### ポート構成

#### 開発環境
- フロントエンド: http://localhost (Traefik経由)
- バックエンドAPI: http://localhost:8081
- MySQL: localhost:3306
- Traefik Dashboard: http://localhost:8080

#### 本番環境
- フロントエンド: https://nitsche-gyouji.com
- バックエンドAPI: https://nitsche-gyouji.com/api
- MySQL: 内部ネットワークのみ
- Traefik Dashboard: http://localhost:8080

## データベース管理

### 初期化

データベースは初回起動時に自動的に初期化されます：
- `backend/migrations/`内のSQLファイルが実行されます
- 管理者ユーザーとサンプルデータが作成されます

### バックアップ

```bash
# データベースバックアップ
docker-compose exec mysql mysqldump -u root -p tournament_db > backup.sql

# データベース復元
docker-compose exec -T mysql mysql -u root -p tournament_db < backup.sql
```

### データボリューム

- 開発環境: `mysql_dev_data`
- 本番環境: `mysql_prod_data`

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**
   ```bash
   # MySQLの起動状況確認
   docker-compose ps mysql
   
   # MySQLログ確認
   docker-compose logs mysql
   ```

2. **バックエンドAPI接続エラー**
   ```bash
   # バックエンドログ確認
   docker-compose logs backend
   
   # ヘルスチェック確認
   curl http://localhost:8081/api/health
   ```

3. **SSL証明書エラー**
   ```bash
   # Traefikログ確認
   docker-compose logs traefik
   
   # 証明書ファイル確認
   ls -la letsencrypt/
   ```

### ログ確認

```bash
# 全サービスのログ
docker-compose logs -f

# 特定サービスのログ
docker-compose logs -f backend
docker-compose logs -f mysql
docker-compose logs -f frontend
```

### コンテナ再起動

```bash
# 特定サービスの再起動
docker-compose restart backend

# 全サービスの再起動
docker-compose restart
```

## セキュリティ考慮事項

1. **環境変数管理**
   - `.env`ファイルをバージョン管理に含めない
   - 本番環境では強力なパスワードを使用

2. **ネットワークセキュリティ**
   - 本番環境ではデータベースポートを公開しない
   - Traefikを通じてのみ外部アクセスを許可

3. **SSL/TLS**
   - Let's Encryptによる自動SSL証明書取得
   - HTTPからHTTPSへの自動リダイレクト

4. **コンテナセキュリティ**
   - 非rootユーザーでアプリケーション実行
   - 最小限のベースイメージ使用（scratch）
   - 定期的なセキュリティアップデート

## 監視とメンテナンス

### ヘルスチェック

各サービスにはヘルスチェックが設定されています：
- MySQL: `mysqladmin ping`
- バックエンド: `/api/health`エンドポイント

### ログローテーション

本番環境では自動ログローテーションが設定されています：
- 最大ファイルサイズ: 10MB
- 保持ファイル数: 3個

### アップデート

```bash
# イメージの更新
docker-compose pull

# サービスの再構築と再起動
docker-compose up -d --build
```
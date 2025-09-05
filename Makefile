# トーナメント管理システム Docker操作用Makefile

.PHONY: help dev prod build up down logs clean restart health backup

# デフォルトターゲット
help:
	@echo "利用可能なコマンド:"
	@echo "  dev      - 開発環境でサービスを起動"
	@echo "  prod     - 本番環境でサービスを起動"
	@echo "  build    - 全サービスをビルド"
	@echo "  up       - サービスを起動（デタッチモード）"
	@echo "  down     - サービスを停止・削除"
	@echo "  logs     - 全サービスのログを表示"
	@echo "  clean    - 未使用のDockerリソースを削除"
	@echo "  restart  - サービスを再起動"
	@echo "  health   - サービスの健全性をチェック"
	@echo "  backup   - データベースをバックアップ"

# 開発環境
dev:
	@echo "開発環境を起動中..."
	docker-compose up -d
	@echo "サービスが起動しました。"
	@echo "フロントエンド: http://localhost"
	@echo "バックエンドAPI: http://localhost:8081"
	@echo "Traefik Dashboard: http://localhost:8080"

# 本番環境
prod:
	@echo "本番環境を起動中..."
	docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	@echo "本番環境が起動しました。"

# ビルド
build:
	@echo "全サービスをビルド中..."
	docker-compose build --no-cache

# 起動
up:
	docker-compose up -d

# 停止
down:
	@echo "サービスを停止中..."
	docker-compose down

# ログ表示
logs:
	docker-compose logs -f

# 特定サービスのログ
logs-backend:
	docker-compose logs -f backend

logs-frontend:
	docker-compose logs -f frontend

logs-mysql:
	docker-compose logs -f mysql

logs-traefik:
	docker-compose logs -f traefik

# クリーンアップ
clean:
	@echo "未使用のDockerリソースを削除中..."
	docker system prune -f
	docker volume prune -f

# 再起動
restart:
	@echo "サービスを再起動中..."
	docker-compose restart

# 特定サービスの再起動
restart-backend:
	docker-compose restart backend

restart-frontend:
	docker-compose restart frontend

restart-mysql:
	docker-compose restart mysql

# ヘルスチェック
health:
	@echo "サービスの健全性をチェック中..."
	@docker-compose ps
	@echo "\nバックエンドAPIヘルスチェック:"
	@curl -f http://localhost:8081/api/health 2>/dev/null && echo "✓ バックエンドAPI: 正常" || echo "✗ バックエンドAPI: エラー"
	@echo "\nフロントエンドヘルスチェック:"
	@curl -f http://localhost 2>/dev/null && echo "✓ フロントエンド: 正常" || echo "✗ フロントエンド: エラー"

# データベースバックアップ
backup:
	@echo "データベースをバックアップ中..."
	@mkdir -p backups
	@docker-compose exec -T mysql mysqldump -u root -p$$(grep DB_ROOT_PASSWORD .env | cut -d '=' -f2) tournament_db > backups/backup_$$(date +%Y%m%d_%H%M%S).sql
	@echo "バックアップが完了しました: backups/backup_$$(date +%Y%m%d_%H%M%S).sql"

# データベース復元
restore:
	@echo "復元するバックアップファイルを指定してください:"
	@echo "make restore-file FILE=backups/backup_YYYYMMDD_HHMMSS.sql"

restore-file:
	@if [ -z "$(FILE)" ]; then echo "エラー: FILEパラメータが必要です"; exit 1; fi
	@echo "データベースを復元中: $(FILE)"
	@docker-compose exec -T mysql mysql -u root -p$$(grep DB_ROOT_PASSWORD .env | cut -d '=' -f2) tournament_db < $(FILE)
	@echo "復元が完了しました"

# 開発環境セットアップ
setup-dev:
	@echo "開発環境をセットアップ中..."
	@if [ ! -f .env ]; then cp .env.sample .env; echo ".envファイルを作成しました。適切な値を設定してください。"; fi
	@docker-compose build
	@docker-compose up -d
	@echo "開発環境のセットアップが完了しました"

# 本番環境セットアップ
setup-prod:
	@echo "本番環境をセットアップ中..."
	@if [ ! -f .env ]; then echo "エラー: .envファイルが存在しません。.env.sampleを参考に作成してください。"; exit 1; fi
	@docker-compose -f docker-compose.yml -f docker-compose.prod.yml build
	@docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
	@echo "本番環境のセットアップが完了しました"

# SSL証明書の更新
renew-ssl:
	@echo "SSL証明書を更新中..."
	@docker-compose exec traefik traefik --certificatesresolvers.myresolver.acme.caserver=https://acme-v02.api.letsencrypt.org/directory
	@echo "SSL証明書の更新が完了しました"
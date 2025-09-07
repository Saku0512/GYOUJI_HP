# マルチステージビルドでのフロントエンド用Dockerfile
# セキュリティ対策と本番環境最適化を含む

# ========================================
# Stage 1: 依存関係インストール
# ========================================
FROM node:20-alpine AS deps

# セキュリティ: 非rootユーザーを作成
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

WORKDIR /app

# パッケージファイルのコピー（キャッシュ最適化）
COPY package*.json ./

# セキュリティ: npm auditを実行して脆弱性チェック
RUN npm audit --audit-level moderate

# 本番用依存関係のみインストール
RUN npm ci --only=production && npm cache clean --force

# ========================================
# Stage 2: ビルド
# ========================================
FROM node:20-alpine AS builder

# セキュリティ: 非rootユーザーを作成
RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

WORKDIR /app

# 全ての依存関係をコピー
COPY package*.json ./
RUN npm ci

# ソースコードをコピー
COPY . .

# 環境変数設定（本番ビルド用）
ENV NODE_ENV=production
ENV VITE_API_BASE_URL=/api

# ビルド実行
RUN npm run build

# 不要なファイルを削除してイメージサイズを削減
RUN rm -rf node_modules
RUN rm -rf src
RUN rm -rf static
RUN rm -rf tests

# ========================================
# Stage 3: 本番用Nginxイメージ
# ========================================
FROM nginx:1.25-alpine AS production

# セキュリティ: 最新のセキュリティアップデートを適用
RUN apk update && apk upgrade && apk add --no-cache \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# セキュリティ: 非rootユーザーでNginxを実行
RUN addgroup -g 1001 -S nginx-group
RUN adduser -S -D -H -u 1001 -h /var/cache/nginx -s /sbin/nologin -G nginx-group -g nginx nginx-user

# ビルド済みアプリケーションをコピー
COPY --from=builder --chown=nginx-user:nginx-group /app/build /usr/share/nginx/html

# カスタムNginx設定をコピー
COPY --chown=nginx-user:nginx-group nginx.conf /etc/nginx/nginx.conf

# セキュリティ: 不要なファイルとディレクトリを削除
RUN rm -rf /etc/nginx/conf.d/default.conf
RUN rm -rf /usr/share/nginx/html/index.html.orig 2>/dev/null || true

# セキュリティ: ファイル権限を設定
RUN chown -R nginx-user:nginx-group /usr/share/nginx/html
RUN chown -R nginx-user:nginx-group /var/cache/nginx
RUN chown -R nginx-user:nginx-group /var/log/nginx
RUN chown -R nginx-user:nginx-group /etc/nginx/conf.d
RUN touch /var/run/nginx.pid
RUN chown -R nginx-user:nginx-group /var/run/nginx.pid

# セキュリティ: 読み取り専用ファイルシステム用の一時ディレクトリ
RUN mkdir -p /tmp/nginx
RUN chown -R nginx-user:nginx-group /tmp/nginx

# ヘルスチェック用スクリプト
RUN echo '#!/bin/sh' > /usr/local/bin/healthcheck.sh && \
    echo 'curl -f http://localhost/health || exit 1' >> /usr/local/bin/healthcheck.sh && \
    chmod +x /usr/local/bin/healthcheck.sh

# セキュリティ: 不要なパッケージを削除
RUN apk del --purge

# ポート80を公開
EXPOSE 80

# ヘルスチェック設定
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD /usr/local/bin/healthcheck.sh

# セキュリティ: 非rootユーザーで実行
# USER nginx-user

# Nginxをフォアグラウンドで起動
CMD ["nginx", "-g", "daemon off;"]

# ========================================
# 開発用ステージ（オプション）
# ========================================
FROM node:20-alpine AS development

RUN addgroup --system --gid 1001 nodejs
RUN adduser --system --uid 1001 nextjs

WORKDIR /app

# 開発用依存関係をインストール
COPY package*.json ./
RUN npm ci

# ソースコードをコピー
COPY . .

# 開発サーバー用の環境変数
ENV NODE_ENV=development
ENV VITE_API_BASE_URL=http://localhost:8080/api

# 開発サーバーのポート
EXPOSE 5173

# 非rootユーザーで実行
USER nextjs

# 開発サーバー起動
CMD ["npm", "run", "dev", "--", "--host", "0.0.0.0"]

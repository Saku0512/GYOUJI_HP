# マルチステージビルドでGoバックエンド用Dockerfile
# ステージ1: ビルドステージ
FROM golang:1.24.2-alpine AS builder

# セキュリティ: 非rootユーザーでビルド実行
RUN adduser -D -s /bin/sh -u 1001 appuser

# 必要なパッケージをインストール（セキュリティアップデート含む）
RUN apk update && apk add --no-cache \
    git \
    ca-certificates \
    tzdata \
    && rm -rf /var/cache/apk/*

# 作業ディレクトリを設定
WORKDIR /app

# Go modulesの依存関係をコピーしてダウンロード（キャッシュ最適化）
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# ソースコードをコピー
COPY . .

# バイナリをビルド（最適化フラグ付き）
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o main ./cmd/server

# ステージ2: 実行ステージ
FROM alpine:3.18

# セキュリティアップデートとnetcat-openbsdをインストール
RUN apk update && apk add --no-cache \
    ca-certificates \
    tzdata \
    netcat-openbsd \
    && rm -rf /var/cache/apk/*

# 非rootユーザーを作成
RUN adduser -D -s /bin/sh -u 1001 appuser

# ビルドしたバイナリをコピー
COPY --from=builder /app/main /main

# マイグレーションファイルをコピー
COPY --from=builder /app/migrations /migrations

# 実行権限を設定
RUN chmod +x /main

# 非rootユーザーに切り替え
USER appuser

# ポート8080を公開
EXPOSE 8080

# アプリケーションを実行
ENTRYPOINT ["/main"]
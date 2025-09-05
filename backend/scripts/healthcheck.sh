#!/bin/sh

# ヘルスチェックスクリプト
# Dockerコンテナ内でアプリケーションの健全性をチェック

# アプリケーションのヘルスチェックエンドポイントを呼び出し
curl -f http://localhost:8080/api/health || exit 1

echo "Health check passed"
exit 0
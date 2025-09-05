#!/bin/bash

# APIドキュメント生成スクリプト
# このスクリプトはSwagger/OpenAPIドキュメントを生成し、検証します

set -e

# スクリプトのディレクトリを取得
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$(dirname "$SCRIPT_DIR")"

echo "=== Tournament Backend API ドキュメント生成 ==="
echo "バックエンドディレクトリ: $BACKEND_DIR"

# バックエンドディレクトリに移動
cd "$BACKEND_DIR"

# Go依存関係の確認
echo "Go依存関係を確認中..."
go mod tidy
go mod download

# Swaggerツールの確認
echo "Swaggerツールを確認中..."
if ! command -v swag &> /dev/null; then
    echo "swagコマンドが見つかりません。インストール中..."
    go install github.com/swaggo/swag/cmd/swag@latest
fi

# 既存のドキュメントをバックアップ
if [ -f "docs/swagger.yaml" ]; then
    echo "既存のドキュメントをバックアップ中..."
    cp docs/swagger.yaml docs/swagger.yaml.backup
fi

# Swaggerドキュメントを生成
echo "Swaggerドキュメントを生成中..."
swag init -g docs/docs.go -o docs --parseDependency --parseInternal

# 生成されたファイルを確認
echo "生成されたファイルを確認中..."
if [ ! -f "docs/swagger.json" ]; then
    echo "エラー: swagger.jsonが生成されませんでした"
    exit 1
fi

if [ ! -f "docs/swagger.yaml" ]; then
    echo "エラー: swagger.yamlが生成されませんでした"
    exit 1
fi

echo "✅ Swaggerドキュメントの生成が完了しました"

# ドキュメントの基本検証
echo "ドキュメントの基本検証を実行中..."

# JSONの構文チェック
if command -v jq &> /dev/null; then
    echo "JSON構文をチェック中..."
    if jq empty docs/swagger.json; then
        echo "✅ JSON構文は正常です"
    else
        echo "❌ JSON構文エラーが見つかりました"
        exit 1
    fi
else
    echo "⚠️  jqがインストールされていません。JSON構文チェックをスキップします"
fi

# YAMLの構文チェック
if command -v yq &> /dev/null; then
    echo "YAML構文をチェック中..."
    if yq eval '.' docs/swagger.yaml > /dev/null; then
        echo "✅ YAML構文は正常です"
    else
        echo "❌ YAML構文エラーが見つかりました"
        exit 1
    fi
else
    echo "⚠️  yqがインストールされていません。YAML構文チェックをスキップします"
fi

# OpenAPI仕様の検証（オプション）
if [ "$1" = "--validate" ]; then
    echo "OpenAPI仕様の詳細検証を実行中..."
    
    if command -v swagger-codegen &> /dev/null; then
        echo "swagger-codegenで検証中..."
        swagger-codegen validate -i docs/swagger.yaml
        echo "✅ OpenAPI仕様の検証が完了しました"
    elif command -v npx &> /dev/null; then
        echo "swagger-parserで検証中..."
        npx @apidevtools/swagger-parser validate docs/swagger.yaml
        echo "✅ OpenAPI仕様の検証が完了しました"
    else
        echo "⚠️  OpenAPI検証ツールが見つかりません"
        echo "以下のいずれかをインストールしてください:"
        echo "  - swagger-codegen: https://swagger.io/tools/swagger-codegen/"
        echo "  - Node.js + npm: https://nodejs.org/"
        echo "または、オンラインエディターを使用してください: https://editor.swagger.io/"
    fi
fi

# 統計情報の表示
echo ""
echo "=== ドキュメント統計 ==="

# エンドポイント数をカウント
if command -v jq &> /dev/null; then
    ENDPOINT_COUNT=$(jq '.paths | keys | length' docs/swagger.json)
    echo "エンドポイント数: $ENDPOINT_COUNT"
    
    # HTTPメソッド別の統計
    echo "HTTPメソッド別統計:"
    jq -r '.paths | to_entries[] | .value | keys[]' docs/swagger.json | sort | uniq -c | sort -nr
    
    # タグ別の統計
    echo "タグ別統計:"
    jq -r '.paths | to_entries[] | .value | to_entries[] | .value.tags[]?' docs/swagger.json | sort | uniq -c | sort -nr
fi

# ファイルサイズ
echo "ファイルサイズ:"
ls -lh docs/swagger.json docs/swagger.yaml | awk '{print $9 ": " $5}'

echo ""
echo "=== 生成されたファイル ==="
echo "📄 docs/swagger.json - JSON形式のAPI仕様"
echo "📄 docs/swagger.yaml - YAML形式のAPI仕様"
echo "📄 docs/docs.go - Go言語のSwagger注釈"
echo "📄 docs/README.md - APIドキュメント"

echo ""
echo "=== アクセス方法 ==="
echo "サーバーを起動後、以下のURLでアクセスできます:"
echo "🌐 Swagger UI: http://localhost:8080/swagger/index.html"
echo "🌐 ドキュメント: http://localhost:8080/docs"
echo "🌐 API情報: http://localhost:8080/"

echo ""
echo "=== 次のステップ ==="
echo "1. サーバーを起動: go run cmd/server/main.go"
echo "2. ブラウザでSwagger UIを開く"
echo "3. APIエンドポイントをテスト"

if [ "$1" = "--serve" ]; then
    echo ""
    echo "サーバーを起動しています..."
    go run cmd/server/main.go
fi

echo ""
echo "🎉 APIドキュメントの生成が完了しました！"
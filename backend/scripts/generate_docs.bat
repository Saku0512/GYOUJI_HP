@echo off
setlocal enabledelayedexpansion

REM APIドキュメント生成スクリプト（Windows版）
REM このスクリプトはSwagger/OpenAPIドキュメントを生成し、検証します

echo === Tournament Backend API ドキュメント生成 ===

REM スクリプトのディレクトリを取得
set SCRIPT_DIR=%~dp0
set BACKEND_DIR=%SCRIPT_DIR%..

echo バックエンドディレクトリ: %BACKEND_DIR%

REM バックエンドディレクトリに移動
cd /d "%BACKEND_DIR%"

REM Go依存関係の確認
echo Go依存関係を確認中...
go mod tidy
go mod download

REM Swaggerツールの確認
echo Swaggerツールを確認中...
swag version >nul 2>&1
if errorlevel 1 (
    echo swagコマンドが見つかりません。インストール中...
    go install github.com/swaggo/swag/cmd/swag@latest
)

REM 既存のドキュメントをバックアップ
if exist "docs\swagger.yaml" (
    echo 既存のドキュメントをバックアップ中...
    copy docs\swagger.yaml docs\swagger.yaml.backup >nul
)

REM Swaggerドキュメントを生成
echo Swaggerドキュメントを生成中...
swag init -g docs/docs.go -o docs --parseDependency --parseInternal

REM 生成されたファイルを確認
echo 生成されたファイルを確認中...
if not exist "docs\swagger.json" (
    echo エラー: swagger.jsonが生成されませんでした
    exit /b 1
)

if not exist "docs\swagger.yaml" (
    echo エラー: swagger.yamlが生成されませんでした
    exit /b 1
)

echo ✅ Swaggerドキュメントの生成が完了しました

REM ドキュメントの基本検証
echo ドキュメントの基本検証を実行中...

REM JSONの構文チェック（PowerShellを使用）
echo JSON構文をチェック中...
powershell -Command "try { Get-Content 'docs\swagger.json' | ConvertFrom-Json | Out-Null; Write-Host '✅ JSON構文は正常です' } catch { Write-Host '❌ JSON構文エラーが見つかりました'; exit 1 }"
if errorlevel 1 exit /b 1

REM 統計情報の表示
echo.
echo === ドキュメント統計 ===

REM ファイルサイズ
echo ファイルサイズ:
for %%f in (docs\swagger.json docs\swagger.yaml) do (
    for %%s in (%%f) do echo %%f: %%~zs bytes
)

echo.
echo === 生成されたファイル ===
echo 📄 docs\swagger.json - JSON形式のAPI仕様
echo 📄 docs\swagger.yaml - YAML形式のAPI仕様
echo 📄 docs\docs.go - Go言語のSwagger注釈
echo 📄 docs\README.md - APIドキュメント

echo.
echo === アクセス方法 ===
echo サーバーを起動後、以下のURLでアクセスできます:
echo 🌐 Swagger UI: http://localhost:8080/swagger/index.html
echo 🌐 ドキュメント: http://localhost:8080/docs
echo 🌐 API情報: http://localhost:8080/

echo.
echo === 次のステップ ===
echo 1. サーバーを起動: go run cmd/server/main.go
echo 2. ブラウザでSwagger UIを開く
echo 3. APIエンドポイントをテスト

if "%1"=="--serve" (
    echo.
    echo サーバーを起動しています...
    go run cmd/server/main.go
)

echo.
echo 🎉 APIドキュメントの生成が完了しました！
pause
@echo off
setlocal enabledelayedexpansion

REM 統合テスト実行スクリプト（Windows版）
REM このスクリプトは統合テストを実行し、テストデータベースのセットアップとクリーンアップを行います

echo === トーナメントバックエンド統合テスト ===

REM スクリプトのディレクトリを取得
set SCRIPT_DIR=%~dp0
set BACKEND_DIR=%SCRIPT_DIR%..

echo バックエンドディレクトリ: %BACKEND_DIR%

REM バックエンドディレクトリに移動
cd /d "%BACKEND_DIR%"

REM 環境変数を設定
set GO_ENV=test
set DB_HOST=localhost
set DB_PORT=3306
set DB_USER=root
set DB_PASSWORD=test_password
set DB_NAME=tournament_test_db
set JWT_SECRET=test_jwt_secret_key_for_testing
set JWT_EXPIRATION_HOURS=24
set JWT_ISSUER=tournament-backend-test
set SERVER_PORT=8081
set SERVER_HOST=localhost

echo テスト環境変数を設定しました

REM テストデータベースの確認
echo テストデータベースの確認中...
mysql -h%DB_HOST% -P%DB_PORT% -u%DB_USER% -p%DB_PASSWORD% -e "CREATE DATABASE IF NOT EXISTS %DB_NAME%;" 2>nul
if errorlevel 1 (
    echo 警告: MySQLデータベースに接続できません。データベースが起動していることを確認してください。
    echo 以下のコマンドでMySQLを起動できます:
    echo   docker run --name mysql-test -e MYSQL_ROOT_PASSWORD=%DB_PASSWORD% -p %DB_PORT%:3306 -d mysql:8.0
    echo.
    echo または、既存のMySQLインスタンスを使用する場合は、以下の設定を確認してください:
    echo   ホスト: %DB_HOST%
    echo   ポート: %DB_PORT%
    echo   ユーザー: %DB_USER%
    echo   パスワード: %DB_PASSWORD%
    echo.
    set /p continue="続行しますか？ (y/N): "
    if /i not "!continue!"=="y" exit /b 1
)

echo テストデータベースの準備が完了しました

REM Go依存関係の確認
echo Go依存関係を確認中...
go mod tidy
go mod download

REM テストの実行
echo.
echo === 統合テスト実行開始 ===
echo.

REM テストの詳細出力を有効にする
set GOMAXPROCS=1

REM 各テストスイートを個別に実行
set test_count=0
set failed_count=0
set passed_count=0

echo --- 認証統合テスト実行中 ---
go test -v -run "TestAuthIntegrationTestSuite" ./integration_test/
if errorlevel 1 (
    echo ❌ 認証統合テスト: 失敗
    set /a failed_count+=1
) else (
    echo ✅ 認証統合テスト: 成功
    set /a passed_count+=1
)
echo.

echo --- トーナメント統合テスト実行中 ---
go test -v -run "TestTournamentIntegrationTestSuite" ./integration_test/
if errorlevel 1 (
    echo ❌ トーナメント統合テスト: 失敗
    set /a failed_count+=1
) else (
    echo ✅ トーナメント統合テスト: 成功
    set /a passed_count+=1
)
echo.

echo --- 試合統合テスト実行中 ---
go test -v -run "TestMatchIntegrationTestSuite" ./integration_test/
if errorlevel 1 (
    echo ❌ 試合統合テスト: 失敗
    set /a failed_count+=1
) else (
    echo ✅ 試合統合テスト: 成功
    set /a passed_count+=1
)
echo.

echo --- ワークフロー統合テスト実行中 ---
go test -v -run "TestWorkflowIntegrationTestSuite" ./integration_test/
if errorlevel 1 (
    echo ❌ ワークフロー統合テスト: 失敗
    set /a failed_count+=1
) else (
    echo ✅ ワークフロー統合テスト: 成功
    set /a passed_count+=1
)
echo.

REM 結果の表示
echo === テスト結果サマリー ===
echo 成功したテスト: %passed_count%
echo 失敗したテスト: %failed_count%

if %failed_count% gtr 0 (
    echo.
    echo 失敗したテストがあります。詳細は上記のログを確認してください。
    exit /b 1
)

echo.
echo 🎉 すべての統合テストが成功しました！
echo.

REM カバレッジレポートの生成（オプション）
if "%1"=="--coverage" (
    echo === カバレッジレポート生成中 ===
    go test -v -coverprofile=coverage.out ./integration_test/
    go tool cover -html=coverage.out -o coverage.html
    echo カバレッジレポートが coverage.html に生成されました
)

echo 統合テストが完了しました。
pause
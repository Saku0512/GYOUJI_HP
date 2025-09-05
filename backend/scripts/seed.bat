@echo off
REM トーナメントデータベースシーディングスクリプト（Windows用）
REM 使用方法: scripts\seed.bat [オプション]

setlocal enabledelayedexpansion

REM デフォルト値
set RESET=false
set SPORT=
set USE_SQL=false
set ADMIN_ONLY=false

REM 引数の解析
:parse_args
if "%~1"=="" goto end_parse
if "%~1"=="--reset" (
    set RESET=true
    shift
    goto parse_args
)
if "%~1"=="--sql" (
    set USE_SQL=true
    shift
    goto parse_args
)
if "%~1"=="--admin-only" (
    set ADMIN_ONLY=true
    shift
    goto parse_args
)
if "%~1"=="--help" (
    goto show_help
)
if "%~1:~0,8%"=="--sport=" (
    set SPORT=%~1:~8%
    shift
    goto parse_args
)
echo [ERROR] 不明なオプション: %~1
goto show_help

:end_parse

REM プロジェクトルートに移動
cd /d "%~dp0.."

REM 環境変数の確認
if not exist ".env" (
    echo [WARNING] .envファイルが見つかりません。.env.sampleを参考に作成してください。
)

REM 管理者ユーザーのみ作成する場合
if "%ADMIN_ONLY%"=="true" (
    echo [INFO] 管理者ユーザーを作成しています...
    echo 注意: この機能はLinux/Mac環境でのみ利用可能です。
    echo Windows環境では手動でMySQLに接続してユーザーを作成してください。
    goto end
)

REM SQLファイルを直接実行する場合
if "%USE_SQL%"=="true" (
    echo [INFO] SQLファイルを使用してシーディングを実行しています...
    echo 注意: この機能はLinux/Mac環境でのみ利用可能です。
    echo Windows環境では手動でMySQLに接続してSQLファイルを実行してください。
    goto end
)

REM Go版のシーディングツールを使用
echo [INFO] Goシーディングツールを使用してシーディングを実行しています...

REM シーディングコマンドの構築
set SEED_CMD=go run cmd/seed/main.go

if "%RESET%"=="true" (
    set SEED_CMD=!SEED_CMD! -reset
)

if not "%SPORT%"=="" (
    set SEED_CMD=!SEED_CMD! -sport=!SPORT!
)

REM シーディング実行
echo [INFO] 実行コマンド: !SEED_CMD!
!SEED_CMD!

if errorlevel 1 (
    echo [ERROR] シーディングに失敗しました
    exit /b 1
)

echo [SUCCESS] シーディングが正常に完了しました
goto end

:show_help
echo トーナメントデータベースシーディングスクリプト（Windows用）
echo.
echo 使用方法:
echo   scripts\seed.bat [オプション]
echo.
echo オプション:
echo   --reset                既存データをリセットしてから実行
echo   --sport=SPORT         特定のスポーツのみシーディング
echo                         (volleyball, table_tennis, soccer)
echo   --sql                 SQLファイルを直接実行（Linux/Mac環境のみ）
echo   --admin-only          管理者ユーザーのみ作成（Linux/Mac環境のみ）
echo   --help                このヘルプを表示
echo.
echo 例:
echo   scripts\seed.bat                           # 全トーナメントを初期化
echo   scripts\seed.bat --reset                   # リセット後に全初期化
echo   scripts\seed.bat --sport=volleyball        # バレーボールのみ初期化
echo.
echo 注意: Windows環境では一部の機能が制限されます。
echo       SQLファイルの直接実行や管理者ユーザー作成は手動で行ってください。
goto end

:end
endlocal
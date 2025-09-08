@echo off
REM 契約テスト実行スクリプト (Windows版)
REM このスクリプトは、API契約テストとOpenAPI仕様検証を実行します

setlocal enabledelayedexpansion

REM デフォルト設定
set COVERAGE=false
set VERBOSE=false
set CONTRACT_ONLY=false
set OPENAPI_ONLY=false

REM コマンドライン引数の解析
:parse_args
if "%~1"=="" goto :args_done
if "%~1"=="--coverage" (
    set COVERAGE=true
    shift
    goto :parse_args
)
if "%~1"=="--verbose" (
    set VERBOSE=true
    shift
    goto :parse_args
)
if "%~1"=="--contract-only" (
    set CONTRACT_ONLY=true
    shift
    goto :parse_args
)
if "%~1"=="--openapi-only" (
    set OPENAPI_ONLY=true
    shift
    goto :parse_args
)
if "%~1"=="--help" (
    goto :show_usage
)
echo [ERROR] 不明なオプション: %~1
goto :show_usage

:args_done

REM プロジェクトルートディレクトリに移動
cd /d "%~dp0\.."

echo [INFO] 契約テスト実行を開始します...
echo [INFO] プロジェクトディレクトリ: %CD%

REM 環境変数の確認
call :check_environment
if errorlevel 1 exit /b 1

REM Go依存関係の確認
call :check_dependencies
if errorlevel 1 exit /b 1

REM OpenAPI仕様ファイルの確認
if not "%OPENAPI_ONLY%"=="true" (
    call :check_openapi_spec
    if errorlevel 1 exit /b 1
)

REM テスト実行
set TEST_FAILED=false

if not "%OPENAPI_ONLY%"=="true" (
    call :run_contract_tests
    if errorlevel 1 set TEST_FAILED=true
)

if not "%CONTRACT_ONLY%"=="true" (
    call :run_openapi_tests
    if errorlevel 1 set TEST_FAILED=true
)

REM カバレッジレポート生成
call :generate_coverage_report

REM 結果の表示
if "%TEST_FAILED%"=="true" (
    echo [ERROR] 契約テストの実行中にエラーが発生しました
    call :cleanup
    exit /b 1
) else (
    echo [SUCCESS] 全ての契約テストが正常に完了しました
    call :cleanup
    exit /b 0
)

:show_usage
echo 使用方法: %~nx0 [オプション]
echo.
echo オプション:
echo   --coverage          カバレッジレポートを生成
echo   --verbose           詳細ログを出力
echo   --contract-only     契約テストのみ実行
echo   --openapi-only      OpenAPI検証テストのみ実行
echo   --help              このヘルプを表示
echo.
echo 例:
echo   %~nx0                  全ての契約テストを実行
echo   %~nx0 --coverage       カバレッジ付きで実行
echo   %~nx0 --contract-only  契約テストのみ実行
exit /b 0

:check_environment
echo [INFO] 環境変数を確認しています...

set MISSING_VARS=
if "%DB_HOST%"=="" set MISSING_VARS=!MISSING_VARS! DB_HOST
if "%DB_PORT%"=="" set MISSING_VARS=!MISSING_VARS! DB_PORT
if "%DB_USER%"=="" set MISSING_VARS=!MISSING_VARS! DB_USER
if "%DB_PASSWORD%"=="" set MISSING_VARS=!MISSING_VARS! DB_PASSWORD
if "%DB_NAME%"=="" set MISSING_VARS=!MISSING_VARS! DB_NAME
if "%JWT_SECRET%"=="" set MISSING_VARS=!MISSING_VARS! JWT_SECRET

if not "!MISSING_VARS!"=="" (
    echo [ERROR] 以下の環境変数が設定されていません:
    for %%v in (!MISSING_VARS!) do echo   - %%v
    echo [INFO] 環境変数を設定してから再実行してください
    exit /b 1
)

echo [SUCCESS] 環境変数の確認が完了しました
exit /b 0

:check_dependencies
echo [INFO] Go依存関係を確認しています...

go mod verify >nul 2>&1
if errorlevel 1 (
    echo [WARNING] Go依存関係の検証に失敗しました。依存関係を更新します...
    go mod tidy
    go mod download
)

echo [SUCCESS] Go依存関係の確認が完了しました
exit /b 0

:check_openapi_spec
echo [INFO] OpenAPI仕様ファイルを確認しています...

set OPENAPI_FILE=docs\openapi.yaml

if not exist "%OPENAPI_FILE%" (
    echo [ERROR] OpenAPI仕様ファイルが見つかりません: %OPENAPI_FILE%
    echo [INFO] OpenAPI仕様ファイルを作成してから再実行してください
    exit /b 1
)

echo [SUCCESS] OpenAPI仕様ファイルの確認が完了しました
exit /b 0

:run_contract_tests
echo [INFO] 契約テストを実行しています...

set TEST_ARGS=-v

if "%VERBOSE%"=="true" set TEST_ARGS=!TEST_ARGS! -test.v
if "%COVERAGE%"=="true" set TEST_ARGS=!TEST_ARGS! -coverprofile=contract_coverage.out

go test !TEST_ARGS! -run "TestContractTestSuite" ./integration_test/
if errorlevel 1 (
    echo [ERROR] 契約テストが失敗しました
    exit /b 1
)

echo [SUCCESS] 契約テストが完了しました
exit /b 0

:run_openapi_tests
echo [INFO] OpenAPI仕様検証テストを実行しています...

set TEST_ARGS=-v

if "%VERBOSE%"=="true" set TEST_ARGS=!TEST_ARGS! -test.v
if "%COVERAGE%"=="true" set TEST_ARGS=!TEST_ARGS! -coverprofile=openapi_coverage.out

go test !TEST_ARGS! -run "TestOpenAPIValidationTestSuite" ./integration_test/
if errorlevel 1 (
    echo [ERROR] OpenAPI仕様検証テストが失敗しました
    exit /b 1
)

echo [SUCCESS] OpenAPI仕様検証テストが完了しました
exit /b 0

:generate_coverage_report
if not "%COVERAGE%"=="true" exit /b 0

echo [INFO] カバレッジレポートを生成しています...

set COVERAGE_FILES=
if exist "contract_coverage.out" set COVERAGE_FILES=!COVERAGE_FILES! contract_coverage.out
if exist "openapi_coverage.out" set COVERAGE_FILES=!COVERAGE_FILES! openapi_coverage.out

if "!COVERAGE_FILES!"=="" (
    echo [WARNING] カバレッジファイルが見つかりません
    exit /b 0
)

REM 結合されたカバレッジファイルの作成
set COMBINED_COVERAGE=combined_contract_coverage.out
echo mode: set > "!COMBINED_COVERAGE!"

for %%f in (!COVERAGE_FILES!) do (
    more +1 "%%f" >> "!COMBINED_COVERAGE!"
)

REM HTMLレポートの生成
go tool cover -html="!COMBINED_COVERAGE!" -o contract_coverage.html
if errorlevel 1 (
    echo [ERROR] カバレッジレポートの生成に失敗しました
) else (
    echo [SUCCESS] カバレッジレポートが生成されました: contract_coverage.html
)

REM カバレッジ率の表示
for /f "tokens=3" %%i in ('go tool cover -func="!COMBINED_COVERAGE!" ^| findstr /E "total:"') do (
    echo [INFO] 契約テストカバレッジ: %%i
)

REM 一時ファイルのクリーンアップ
for %%f in (!COVERAGE_FILES!) do del "%%f" >nul 2>&1
del "!COMBINED_COVERAGE!" >nul 2>&1

exit /b 0

:cleanup
echo [INFO] クリーンアップを実行しています...

del contract_coverage.out >nul 2>&1
del openapi_coverage.out >nul 2>&1
del combined_contract_coverage.out >nul 2>&1

echo [SUCCESS] クリーンアップが完了しました
exit /b 0
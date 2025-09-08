# アラートシステム

このディレクトリには、トーナメント管理システムのアラート・監視機能が含まれています。

## 概要

アラートシステムは以下の機能を提供します：

- **異常検知とアラート機能**: システムの異常状態を検知し、適切なアラートを発火
- **システム健全性の監視**: データベース、メモリ、CPU、HTTPエンドポイントなどの監視
- **自動復旧機能**: 一部の問題に対する自動復旧の試行
- **多様な通知方法**: ログ、Webhook、Slack、メールによる通知
- **アラートルール管理**: 動的なアラートルールの作成・更新・削除

## アーキテクチャ

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  HealthMonitor  │    │  AlertManager   │    │   Notifiers     │
│                 │    │                 │    │                 │
│ - Database      │───▶│ - Rule Engine   │───▶│ - Log           │
│ - Memory        │    │ - Alert Store   │    │ - Webhook       │
│ - CPU           │    │ - Evaluation    │    │ - Slack         │
│ - HTTP          │    │ - Recovery      │    │ - Email         │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Metrics       │    │   Alert Store   │    │  HTTP Handler   │
│   Collector     │    │                 │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## ファイル構成

- `alert.go` - アラートの基本構造体とインターフェース定義
- `manager.go` - アラートマネージャーの実装
- `notifiers.go` - 各種通知機能の実装
- `health_monitor.go` - ヘルスモニタリング機能
- `init.go` - システム初期化とグローバル設定
- `alert_test.go` - テストコード
- `config.example.yaml` - 設定例
- `README.md` - このファイル

## 使用方法

### 1. 初期化

```go
import (
    "backend/internal/alert"
    "backend/internal/metrics"
)

// アラートシステムを初期化
config := alert.LoadConfigFromEnv()
alertSystem, err := alert.InitializeGlobalAlertSystem(db, metricsCollector)
if err != nil {
    log.Fatal("Failed to initialize alert system:", err)
}

// システムを開始
ctx := context.Background()
if err := alert.StartGlobalAlertSystem(ctx, alertSystem); err != nil {
    log.Fatal("Failed to start alert system:", err)
}
```

### 2. 手動アラート発火

```go
// アラートマネージャーを取得
manager := alert.GetManager()

// アラートを作成
alert := alert.NewAlert(
    alert.AlertTypeHighCPU,
    alert.SeverityWarning,
    "CPU使用率が高い",
    "CPU使用率が80%を超えています",
    "system_monitor",
)

// アラートを発火
if err := manager.FireAlert(alert); err != nil {
    log.Error("Failed to fire alert:", err)
}
```

### 3. アラートルールの追加

```go
// アラートルールを作成
rule := alert.NewAlertRule(
    "高メモリ使用率",
    alert.AlertTypeHighMemory,
    alert.SeverityWarning,
)
rule.Threshold = 80.0
rule.Operator = alert.OperatorGreaterThan
rule.Duration = 5 * time.Minute

// ルールを追加
if err := manager.AddRule(rule); err != nil {
    log.Error("Failed to add alert rule:", err)
}
```

### 4. HTTP API使用例

```bash
# アラート一覧取得
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/v1/alerts

# アクティブアラート取得
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/v1/alerts/active

# ヘルス状態取得
curl -H "Authorization: Bearer <token>" \
     http://localhost:8080/api/v1/health/status

# アラートをサイレンス（管理者のみ）
curl -X POST \
     -H "Authorization: Bearer <admin-token>" \
     -H "Content-Type: application/json" \
     -d '{"duration": "1h", "reason": "メンテナンス中"}' \
     http://localhost:8080/api/v1/admin/alerts/{alert-id}/silence
```

## 環境変数設定

アラートシステムは以下の環境変数で設定できます：

### 基本設定
```bash
# アラート評価間隔
ALERT_EVALUATION_INTERVAL=30s

# ヘルスチェック間隔
ALERT_HEALTH_CHECK_INTERVAL=30s

# 機能有効化フラグ
ALERT_ENABLE_HEALTH_MONITOR=true
ALERT_ENABLE_NOTIFICATIONS=true
ALERT_ENABLE_AUTO_RECOVERY=true

# ストレージ設定
ALERT_USE_MEMORY_STORE=true
ALERT_DATABASE_URL=postgres://user:pass@localhost/alerts
```

### 通知設定
```bash
# Webhook通知
ALERT_WEBHOOK_URL=https://your-webhook-endpoint.com/alerts
ALERT_WEBHOOK_TIMEOUT=10s

# Slack通知
ALERT_SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/SLACK/WEBHOOK
ALERT_SLACK_CHANNEL=#alerts
ALERT_SLACK_USERNAME=Alert Bot

# メール通知
ALERT_SMTP_HOST=smtp.gmail.com
ALERT_SMTP_PORT=587
ALERT_SMTP_USERNAME=your-email@gmail.com
ALERT_SMTP_PASSWORD=your-app-password
ALERT_EMAIL_FROM=alerts@yourcompany.com
ALERT_EMAIL_TO=admin@yourcompany.com,devops@yourcompany.com
```

## アラートタイプ

システムでサポートされているアラートタイプ：

### システム関連
- `high_cpu` - 高CPU使用率
- `high_memory` - 高メモリ使用量
- `high_disk_usage` - 高ディスク使用量
- `high_error_rate` - 高エラー率
- `slow_response` - レスポンス時間遅延
- `high_latency` - 高レイテンシ

### データベース関連
- `db_connection_error` - データベース接続エラー
- `db_slow_query` - 遅いクエリ
- `db_high_connections` - 高接続数

### アプリケーション関連
- `auth_failure` - 認証失敗
- `api_rate_limit` - APIレート制限
- `websocket_error` - WebSocketエラー
- `business_logic_error` - ビジネスロジックエラー

### インフラ関連
- `service_down` - サービス停止
- `health_check_failed` - ヘルスチェック失敗
- `external_api_error` - 外部API エラー

## 重要度レベル

- `info` - 情報レベル
- `warning` - 警告レベル
- `error` - エラーレベル
- `critical` - クリティカルレベル

## 通知機能

### 1. ログ通知
- 常に有効
- 構造化ログとして出力
- 重要度に応じたログレベル

### 2. Webhook通知
- カスタムエンドポイントへのHTTP POST
- JSON形式でアラート情報を送信
- タイムアウトとリトライ機能

### 3. Slack通知
- Slack Incoming Webhookを使用
- 重要度に応じた色分け
- リッチフォーマットでの表示

### 4. メール通知
- SMTP経由でのメール送信
- 複数の宛先に対応
- HTML形式での送信（将来実装予定）

## ヘルスモニタリング

システムは以下の項目を定期的に監視します：

### 1. データベース
- 接続テスト
- 簡単なクエリ実行
- 接続プール統計

### 2. メモリ使用量
- ヒープメモリ使用量
- システムメモリ使用量
- GC統計

### 3. CPU使用量
- Goroutine数による簡易判定
- 実際のCPU使用率（将来実装予定）

### 4. HTTPエンドポイント
- 内部ヘルスチェックエンドポイント
- レスポンス時間測定
- ステータスコード確認

## 自動復旧機能

一部の問題に対して自動復旧を試行します：

### データベース復旧
- 接続プール統計の確認
- 接続テストの再実行
- 接続プールのリセット（将来実装予定）

### 設定可能な復旧アクション
- 有効/無効の切り替え
- タイムアウト設定
- 最大リトライ回数
- リトライ間隔

## テスト

```bash
# 単体テストの実行
go test ./internal/alert/...

# ベンチマークテストの実行
go test -bench=. ./internal/alert/...

# カバレッジ付きテスト
go test -cover ./internal/alert/...
```

## 監視とメトリクス

アラートシステム自体も監視されます：

- `alerts_total` - 発火したアラート総数
- `health_checks_total` - ヘルスチェック実行回数
- `health_check_duration_seconds` - ヘルスチェック実行時間
- `health_status` - ヘルス状態（0=異常, 1=劣化, 2=正常）

## トラブルシューティング

### よくある問題

1. **アラートが発火しない**
   - アラートルールが有効になっているか確認
   - 閾値設定が適切か確認
   - ログでエラーメッセージを確認

2. **通知が届かない**
   - 通知機能が有効になっているか確認
   - 環境変数の設定を確認
   - ネットワーク接続を確認

3. **ヘルスチェックが失敗する**
   - データベース接続を確認
   - システムリソースを確認
   - ログでエラー詳細を確認

### ログ確認

```bash
# アラート関連のログを確認
grep "alert" /var/log/tournament-api.log

# ヘルスチェック関連のログを確認
grep "health" /var/log/tournament-api.log
```

## 今後の改善予定

- [ ] データベースストアの実装
- [ ] より詳細なCPU監視
- [ ] ディスク使用量の実際の監視
- [ ] メール通知のHTML形式対応
- [ ] アラートのグループ化機能
- [ ] アラート履歴の長期保存
- [ ] ダッシュボード機能
- [ ] アラートのエスカレーション機能

## 貢献

アラートシステムの改善に貢献する場合は、以下のガイドラインに従ってください：

1. 新機能は適切なテストを含める
2. ログ出力は日本語で統一
3. エラーハンドリングを適切に実装
4. ドキュメントを更新する
5. パフォーマンスへの影響を考慮する
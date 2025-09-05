# データベースシーディング

このディレクトリには、トーナメント管理システムのデータベースシーディング用のSQLファイルが含まれています。

## ファイル構成

### SQLファイル

- `001_seed_admin_user.sql` - 管理者ユーザーの作成
- `002_seed_tournaments.sql` - 基本トーナメントデータの作成
- `003_seed_volleyball_matches.sql` - バレーボール試合データ
- `004_seed_table_tennis_matches.sql` - 卓球試合データ
- `005_seed_soccer_matches.sql` - サッカー試合データ
- `seed_all.sql` - 全てのシーディングを実行するマスターファイル

### 実行ツール

- `../../cmd/seed/main.go` - Goベースのシーディングツール
- `../../scripts/seed.sh` - Linux/Mac用シェルスクリプト
- `../../scripts/seed.bat` - Windows用バッチスクリプト

## 使用方法

### 1. Goツールを使用（推奨）

```bash
# プロジェクトルートで実行
cd backend

# 全トーナメントを初期化
go run cmd/seed/main.go

# 既存データをリセットしてから初期化
go run cmd/seed/main.go -reset

# 特定のスポーツのみ初期化
go run cmd/seed/main.go -sport=volleyball
go run cmd/seed/main.go -sport=table_tennis
go run cmd/seed/main.go -sport=soccer

# ヘルプを表示
go run cmd/seed/main.go -help
```

### 2. シェルスクリプトを使用

```bash
# Linux/Mac環境
./scripts/seed.sh
./scripts/seed.sh --reset
./scripts/seed.sh --sport=volleyball
./scripts/seed.sh --admin-only

# Windows環境
scripts\seed.bat
scripts\seed.bat --reset
scripts\seed.bat --sport=volleyball
```

### 3. SQLファイルを直接実行

```bash
# MySQLクライアントを使用
mysql -u root -p tournament_db < migrations/seeds/seed_all.sql

# または個別に実行
mysql -u root -p tournament_db < migrations/seeds/001_seed_admin_user.sql
mysql -u root -p tournament_db < migrations/seeds/002_seed_tournaments.sql
# ... 他のファイルも同様に
```

## 初期データ

### 管理者ユーザー

- **ユーザー名**: `admin`
- **パスワード**: `admin123`
- **役割**: `admin`

⚠️ **セキュリティ注意**: 本番環境では必ず強力なパスワードに変更してください。

### トーナメントデータ

システムは以下の3つのスポーツのトーナメントを自動作成します：

1. **バレーボール** (`volleyball`)
   - 第1体育館で開催
   - 8チームによるトーナメント
   - 1回戦 → 準々決勝 → 準決勝 → 3位決定戦・決勝

2. **卓球** (`table_tennis`)
   - 第2体育館で開催
   - 8チームによるトーナメント
   - 晴天時と雨天時で異なるフォーマット
   - 雨天時は敗者復活戦あり

3. **サッカー** (`soccer`)
   - グラウンドで開催
   - 8チームによる8人制サッカー
   - 1回戦 → 準々決勝 → 準決勝 → 3位決定戦・決勝

### 試合データ

各スポーツの試合データは、READMEファイルに記載されている実際のスケジュールに基づいて作成されます：

- **チーム名**: 実際のクラス名（1-1, 1-2, IE4, IS5など）
- **試合時間**: READMEに記載された時間スロット
- **ステータス**: 全て `pending`（未実施）
- **プレースホルダー**: 準々決勝以降は `TBD`（To Be Determined）

## 開発用ユーティリティ

### データリセット

```bash
# 全データをリセット
go run cmd/seed/main.go -reset

# 特定のスポーツのみリセット
go run cmd/seed/main.go -reset -sport=volleyball
```

### 部分的なシーディング

```bash
# 管理者ユーザーのみ作成
./scripts/seed.sh --admin-only

# 特定のスポーツのみ
go run cmd/seed/main.go -sport=table_tennis
```

## トラブルシューティング

### よくある問題

1. **データベース接続エラー**
   - `.env`ファイルの設定を確認
   - MySQLサーバーが起動していることを確認

2. **外部キー制約エラー**
   - 既存データがある場合は `-reset` オプションを使用
   - マイグレーションが正しく実行されていることを確認

3. **権限エラー**
   - データベースユーザーに適切な権限があることを確認
   - `CREATE`, `INSERT`, `DELETE` 権限が必要

### ログの確認

シーディング実行時のログを確認して問題を特定してください：

```bash
# 詳細ログ付きで実行
go run cmd/seed/main.go -reset 2>&1 | tee seed.log
```

## カスタマイズ

### 新しいチームの追加

1. 対応するSQLファイル（`003_seed_volleyball_matches.sql`など）を編集
2. 新しいチーム名を追加
3. 必要に応じて試合数を調整

### 新しいスポーツの追加

1. `internal/models/constants.go`に新しいスポーツを追加
2. 新しいシーディングSQLファイルを作成
3. `seed_all.sql`に新しいファイルを追加
4. Goシーディングサービスに対応ロジックを追加

### 時間スケジュールの変更

各スポーツのSQLファイルで `scheduled_at` の値を調整してください。時間は `DATE_ADD(CURDATE(), INTERVAL X HOUR)` 形式で指定されています。
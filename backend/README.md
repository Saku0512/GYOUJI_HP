ディレクトリ構成

```bash
backend/
├─ cmd/
│   └─ server/
│       └─ main.go        ← エントリポイント
├─ internal/
│   ├─ handler/
│   │   └─ user_handler.go
│   ├─ service/
│   │   └─ user_service.go
│   └─ repository/
│       └─ user_repository.go
├─ go.mod
└─ go.sum
```

- ログイン機能(JWT認証)->単一のユーザーのみを承認し、管理者でアカウントを共有する
- ログイン成功後ダッシュボードに移動
- ダッシュボードで競技の結果を入力
- MySQLでデータを管理(ginからのみ接続可能)
- ダッシュボードの入力要項
    - 競技名(select optionで選択)
    - 結果(どちらが勝ったか、点数はそれぞれ何点か)
- 別途各競技のトーナメントを作成し、それにデータを反映させ、リアルタイムでユーザーは確認できる
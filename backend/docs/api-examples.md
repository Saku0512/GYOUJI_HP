# API 使用例

このドキュメントでは、Tournament Backend APIの実際の使用例を示します。

## 前提条件

- サーバーが `http://localhost:8080` で起動していること
- 管理者認証情報: `admin` / `password`

## 認証フロー

### 1. ログイン

```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "password"
  }'
```

**レスポンス例**:
```json
{
  "success": true,
  "message": "ログインに成功しました",
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoxLCJ1c2VybmFtZSI6ImFkbWluIiwicm9sZSI6ImFkbWluIiwiZXhwIjoxNzA0MTUzNjAwfQ.example",
  "username": "admin",
  "role": "admin"
}
```

### 2. トークンリフレッシュ

```bash
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

## トーナメント管理

### 1. 全トーナメント取得

```bash
curl -X GET http://localhost:8080/api/tournaments \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**レスポンス例**:
```json
{
  "success": true,
  "message": "トーナメント一覧を取得しました",
  "data": [
    {
      "id": 1,
      "sport": "volleyball",
      "format": "standard",
      "status": "active",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    },
    {
      "id": 2,
      "sport": "table_tennis",
      "format": "standard",
      "status": "active",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    },
    {
      "id": 3,
      "sport": "soccer",
      "format": "standard",
      "status": "active",
      "created_at": "2024-01-01T10:00:00Z",
      "updated_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

### 2. スポーツ別トーナメント取得

```bash
# バレーボール
curl -X GET http://localhost:8080/api/tournaments/volleyball \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 卓球
curl -X GET http://localhost:8080/api/tournaments/table_tennis \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# サッカー
curl -X GET http://localhost:8080/api/tournaments/soccer \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 3. トーナメントブラケット取得

```bash
curl -X GET http://localhost:8080/api/tournaments/volleyball/bracket \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

**レスポンス例**:
```json
{
  "success": true,
  "message": "ブラケット情報を取得しました",
  "data": {
    "tournament_id": 1,
    "sport": "volleyball",
    "format": "standard",
    "rounds": [
      {
        "name": "1st_round",
        "matches": [
          {
            "id": 1,
            "tournament_id": 1,
            "round": "1st_round",
            "team1": "チームA",
            "team2": "チームB",
            "score1": null,
            "score2": null,
            "winner": null,
            "status": "pending",
            "scheduled_at": "2024-01-01T10:00:00Z",
            "completed_at": null
          }
        ]
      }
    ]
  }
}
```

### 4. 新規トーナメント作成（管理者のみ）

```bash
curl -X POST http://localhost:8080/api/tournaments \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "sport": "volleyball",
    "format": "standard",
    "status": "active"
  }'
```

### 5. 卓球の形式切り替え（雨天時対応）

```bash
# 雨天時形式に切り替え
curl -X PUT http://localhost:8080/api/tournaments/table_tennis/format \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "format": "rainy"
  }'

# 標準形式に戻す
curl -X PUT http://localhost:8080/api/tournaments/table_tennis/format \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "format": "standard"
  }'
```

## 試合管理

### 1. 全試合取得

```bash
# 全試合
curl -X GET http://localhost:8080/api/matches \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 未完了の試合のみ
curl -X GET "http://localhost:8080/api/matches?status=pending" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 完了した試合のみ
curl -X GET "http://localhost:8080/api/matches?status=completed" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# ページネーション
curl -X GET "http://localhost:8080/api/matches?limit=10&offset=0" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 2. スポーツ別試合取得

```bash
# バレーボールの試合
curl -X GET http://localhost:8080/api/matches/volleyball \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

# 卓球の未完了試合
curl -X GET "http://localhost:8080/api/matches/table_tennis?status=pending" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 3. 特定試合取得

```bash
curl -X GET http://localhost:8080/api/matches/match/1 \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 4. 新規試合作成（管理者のみ）

```bash
curl -X POST http://localhost:8080/api/matches \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "tournament_id": 1,
    "round": "quarterfinal",
    "team1": "チームA",
    "team2": "チームB",
    "scheduled_at": "2024-01-01T15:00:00Z"
  }'
```

### 5. 試合結果提出（管理者のみ）

```bash
# バレーボール試合結果
curl -X PUT http://localhost:8080/api/matches/1/result \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "score1": 3,
    "score2": 1,
    "winner": "チームA"
  }'

# 卓球試合結果
curl -X PUT http://localhost:8080/api/matches/2/result \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "score1": 11,
    "score2": 9,
    "winner": "選手A"
  }'

# サッカー試合結果
curl -X PUT http://localhost:8080/api/matches/3/result \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -H "Content-Type: application/json" \
  -d '{
    "score1": 2,
    "score2": 1,
    "winner": "チームC"
  }'
```

**レスポンス例**:
```json
{
  "success": true,
  "message": "試合結果を更新しました",
  "data": {
    "id": 1,
    "tournament_id": 1,
    "round": "1st_round",
    "team1": "チームA",
    "team2": "チームB",
    "score1": 3,
    "score2": 1,
    "winner": "チームA",
    "status": "completed",
    "scheduled_at": "2024-01-01T10:00:00Z",
    "completed_at": "2024-01-01T11:00:00Z"
  }
}
```

## 完全なワークフロー例

### シナリオ: バレーボールトーナメントの管理

```bash
# 1. ログイン
TOKEN=$(curl -s -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "password"}' | \
  jq -r '.token')

# 2. バレーボールトーナメントの状況確認
curl -X GET http://localhost:8080/api/tournaments/volleyball \
  -H "Authorization: Bearer $TOKEN"

# 3. バレーボールの試合一覧取得
curl -X GET http://localhost:8080/api/matches/volleyball \
  -H "Authorization: Bearer $TOKEN"

# 4. 未完了の試合を確認
curl -X GET "http://localhost:8080/api/matches/volleyball?status=pending" \
  -H "Authorization: Bearer $TOKEN"

# 5. 最初の試合結果を提出
curl -X PUT http://localhost:8080/api/matches/1/result \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "score1": 3,
    "score2": 1,
    "winner": "チームA"
  }'

# 6. 更新されたブラケットを確認
curl -X GET http://localhost:8080/api/tournaments/volleyball/bracket \
  -H "Authorization: Bearer $TOKEN"

# 7. トーナメント進行状況を確認
curl -X GET http://localhost:8080/api/tournaments/volleyball/progress \
  -H "Authorization: Bearer $TOKEN"
```

## エラーハンドリング例

### 1. 認証エラー

```bash
# 無効なトークン
curl -X GET http://localhost:8080/api/tournaments \
  -H "Authorization: Bearer invalid_token"
```

**レスポンス**:
```json
{
  "success": false,
  "message": "無効または期限切れのトークンです",
  "error": "AUTHENTICATION_ERROR",
  "code": 401
}
```

### 2. 権限エラー

```bash
# 認証なしで管理者専用エンドポイントにアクセス
curl -X POST http://localhost:8080/api/tournaments \
  -H "Content-Type: application/json" \
  -d '{"sport": "volleyball"}'
```

**レスポンス**:
```json
{
  "success": false,
  "message": "認証が必要です",
  "error": "AUTHORIZATION_ERROR",
  "code": 401
}
```

### 3. バリデーションエラー

```bash
# 無効なデータで試合結果提出
curl -X PUT http://localhost:8080/api/matches/1/result \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "score1": -1,
    "score2": 2,
    "winner": "存在しないチーム"
  }'
```

**レスポンス**:
```json
{
  "success": false,
  "message": "無効な試合結果です",
  "error": "VALIDATION_ERROR",
  "code": 400
}
```

### 4. リソース未発見エラー

```bash
# 存在しない試合にアクセス
curl -X GET http://localhost:8080/api/matches/match/99999 \
  -H "Authorization: Bearer $TOKEN"
```

**レスポンス**:
```json
{
  "success": false,
  "message": "試合が見つかりません",
  "error": "NOT_FOUND_ERROR",
  "code": 404
}
```

## JavaScript/Fetch API例

### ログインとトークン管理

```javascript
class TournamentAPI {
  constructor(baseURL = 'http://localhost:8080') {
    this.baseURL = baseURL;
    this.token = localStorage.getItem('tournament_token');
  }

  async login(username, password) {
    const response = await fetch(`${this.baseURL}/api/auth/login`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });

    const data = await response.json();
    
    if (data.success) {
      this.token = data.token;
      localStorage.setItem('tournament_token', this.token);
    }
    
    return data;
  }

  async getTournaments() {
    const response = await fetch(`${this.baseURL}/api/tournaments`, {
      headers: {
        'Authorization': `Bearer ${this.token}`,
      },
    });

    return await response.json();
  }

  async submitMatchResult(matchId, score1, score2, winner) {
    const response = await fetch(`${this.baseURL}/api/matches/${matchId}/result`, {
      method: 'PUT',
      headers: {
        'Authorization': `Bearer ${this.token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ score1, score2, winner }),
    });

    return await response.json();
  }
}

// 使用例
const api = new TournamentAPI();

// ログイン
await api.login('admin', 'password');

// トーナメント一覧取得
const tournaments = await api.getTournaments();
console.log(tournaments);

// 試合結果提出
const result = await api.submitMatchResult(1, 3, 1, 'チームA');
console.log(result);
```

## Python/requests例

```python
import requests
import json

class TournamentAPI:
    def __init__(self, base_url='http://localhost:8080'):
        self.base_url = base_url
        self.token = None
    
    def login(self, username, password):
        response = requests.post(
            f'{self.base_url}/api/auth/login',
            json={'username': username, 'password': password}
        )
        
        data = response.json()
        if data.get('success'):
            self.token = data['token']
        
        return data
    
    def get_headers(self):
        return {'Authorization': f'Bearer {self.token}'}
    
    def get_tournaments(self):
        response = requests.get(
            f'{self.base_url}/api/tournaments',
            headers=self.get_headers()
        )
        return response.json()
    
    def submit_match_result(self, match_id, score1, score2, winner):
        response = requests.put(
            f'{self.base_url}/api/matches/{match_id}/result',
            headers=self.get_headers(),
            json={'score1': score1, 'score2': score2, 'winner': winner}
        )
        return response.json()

# 使用例
api = TournamentAPI()

# ログイン
login_result = api.login('admin', 'password')
print(f"ログイン結果: {login_result}")

# トーナメント一覧取得
tournaments = api.get_tournaments()
print(f"トーナメント: {tournaments}")

# 試合結果提出
result = api.submit_match_result(1, 3, 1, 'チームA')
print(f"試合結果: {result}")
```

## テスト用データセット

### 初期データ

システムには以下の初期データが含まれています：

```json
{
  "users": [
    {
      "id": 1,
      "username": "admin",
      "password": "password",
      "role": "admin"
    }
  ],
  "tournaments": [
    {
      "id": 1,
      "sport": "volleyball",
      "format": "standard",
      "status": "active"
    },
    {
      "id": 2,
      "sport": "table_tennis",
      "format": "standard",
      "status": "active"
    },
    {
      "id": 3,
      "sport": "soccer",
      "format": "standard",
      "status": "active"
    }
  ],
  "matches": [
    {
      "id": 1,
      "tournament_id": 1,
      "round": "1st_round",
      "team1": "チームA",
      "team2": "チームB",
      "status": "pending"
    },
    {
      "id": 2,
      "tournament_id": 2,
      "round": "1st_round",
      "team1": "選手A",
      "team2": "選手B",
      "status": "pending"
    },
    {
      "id": 3,
      "tournament_id": 3,
      "round": "1st_round",
      "team1": "チームC",
      "team2": "チームD",
      "status": "pending"
    }
  ]
}
```

## トラブルシューティング

### よくある問題と解決方法

1. **CORS エラー**
   ```bash
   # プリフライトリクエストの確認
   curl -X OPTIONS http://localhost:8080/api/tournaments \
     -H "Origin: http://localhost:3000" \
     -H "Access-Control-Request-Method: GET" \
     -H "Access-Control-Request-Headers: Authorization"
   ```

2. **トークン期限切れ**
   ```bash
   # トークンの有効性確認
   curl -X POST http://localhost:8080/api/auth/validate \
     -H "Content-Type: application/json" \
     -d '{"token": "your_token_here"}'
   ```

3. **レート制限**
   ```bash
   # レート制限の確認（1分間に10回まで）
   for i in {1..15}; do
     curl -X POST http://localhost:8080/api/auth/login \
       -H "Content-Type: application/json" \
       -d '{"username": "admin", "password": "wrong"}' \
       -w "Request $i: %{http_code}\n" \
       -o /dev/null -s
   done
   ```

このドキュメントを参考に、Tournament Backend APIを効果的に活用してください。
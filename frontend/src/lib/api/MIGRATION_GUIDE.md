# API統合移行ガイド

## 概要

このドキュメントは、既存のAPIクライアントから統一APIクライアントへの移行方法を説明します。

## 統一APIクライアントの利点

1. **一貫したレスポンス形式**: 全てのAPIが統一されたレスポンス形式を返します
2. **型安全性**: TypeScriptの型定義により、コンパイル時にエラーを検出できます
3. **エラーハンドリングの標準化**: 統一されたエラーコードとメッセージ
4. **エンドポイントの統一**: `/api/v1`プレフィックスによるバージョン管理

## 移行方法

### 1. 基本的な使用方法

#### 旧方式
```javascript
import { authAPI, tournamentAPI, matchAPI } from '$lib/api';

// 認証
const loginResult = await authAPI.login('username', 'password');

// トーナメント取得
const tournaments = await tournamentAPI.getTournaments();

// 試合取得
const matches = await matchAPI.getMatches('volleyball');
```

#### 新方式（推奨）
```javascript
import { unifiedAPI } from '$lib/api';

// 認証
const loginResult = await unifiedAPI.auth.login({
  username: 'username',
  password: 'password'
});

// トーナメント取得
const tournaments = await unifiedAPI.tournaments.getAll();

// 試合取得
const matches = await unifiedAPI.matches.getBySport('volleyball');
```

### 2. 型安全性の活用

```typescript
import type { APIResponse, Tournament, Match } from '$lib/api/types';
import { unifiedAPI } from '$lib/api';

// 型安全なAPI呼び出し
const response: APIResponse<Tournament[]> = await unifiedAPI.tournaments.getAll();

if (response.success) {
  // response.dataはTournament[]型として扱われる
  const tournaments: Tournament[] = response.data;
}
```

### 3. エラーハンドリング

```javascript
import { unifiedAPI, ErrorCode } from '$lib/api';

try {
  const response = await unifiedAPI.auth.login({ username, password });
  
  if (!response.success) {
    switch (response.error) {
      case ErrorCode.AUTH_INVALID_CREDENTIALS:
        console.error('認証情報が正しくありません');
        break;
      case ErrorCode.SYSTEM_NETWORK_ERROR:
        console.error('ネットワークエラーが発生しました');
        break;
      default:
        console.error('予期しないエラー:', response.message);
    }
  }
} catch (error) {
  console.error('API呼び出しエラー:', error);
}
```

## 後方互換性

既存のコードは引き続き動作しますが、内部的には統一APIクライアントを使用するように更新されています。

```javascript
// これらは引き続き動作します
import { authAPI, tournamentAPI, matchAPI } from '$lib/api';

const result = await authAPI.login('username', 'password');
const tournaments = await tournamentAPI.getTournaments();
const matches = await matchAPI.getMatches('volleyball');
```

## 段階的移行計画

### Phase 1: 新しいコードでは統一APIクライアントを使用
- 新機能の実装では`unifiedAPI`を使用
- 既存コードはそのまま維持

### Phase 2: 既存コードの段階的移行
- 重要度の高い機能から順次移行
- テストを実行して動作確認

### Phase 3: 旧APIクライアントの廃止
- 全ての移行が完了後、旧APIクライアントを削除

## 主要な変更点

### エンドポイントパス
- 旧: `/auth/login`, `/tournaments`, `/matches`
- 新: `/api/v1/auth/login`, `/api/v1/tournaments`, `/api/v1/matches`

### レスポンス形式
```javascript
// 旧形式（一部で不統一）
{
  success: true,
  data: {...},
  message: "成功"
}

// 新形式（統一）
{
  success: true,
  data: {...},
  message: "成功",
  code: 200,
  timestamp: "2024-01-01T12:00:00Z",
  request_id: "req_123456"
}
```

### エラーレスポンス
```javascript
// 旧形式
{
  success: false,
  error: "エラーメッセージ",
  message: "詳細メッセージ"
}

// 新形式
{
  success: false,
  error: "AUTH_INVALID_CREDENTIALS",
  message: "認証情報が正しくありません",
  code: 401,
  timestamp: "2024-01-01T12:00:00Z",
  request_id: "req_123456"
}
```

## トラブルシューティング

### よくある問題

1. **型エラーが発生する**
   - TypeScriptの型定義を確認してください
   - `npm run type-check`でコンパイルエラーを確認

2. **エンドポイントが見つからない**
   - バックエンドが新しいエンドポイント構造に対応しているか確認
   - `/api/v1`プレフィックスが正しく設定されているか確認

3. **認証エラーが発生する**
   - JWTトークンの形式が統一されているか確認
   - トークンの有効期限を確認

## サポート

移行に関する質問や問題がある場合は、開発チームまでお問い合わせください。
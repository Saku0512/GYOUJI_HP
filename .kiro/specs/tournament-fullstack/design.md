# 設計書

## 概要

トーナメントフルスタックシステムは、既存のGo言語バックエンドAPIと統合されたSvelteフロントエンドアプリケーションです。システムは管理者向けの試合結果入力機能と一般ユーザー向けのトーナメント閲覧機能を提供し、リアルタイム更新とレスポンシブデザインを特徴とします。

## アーキテクチャ

### システム全体アーキテクチャ

```
┌─────────────────┐    HTTP/REST API    ┌─────────────────┐
│   Svelte        │ ←─────────────────→ │   Go Backend    │
│   Frontend      │                     │   (Gin + MySQL) │
│                 │                     │                 │
│ ┌─────────────┐ │                     │ ┌─────────────┐ │
│ │   Pages     │ │                     │ │  Handlers   │ │
│ │             │ │                     │ │             │ │
│ ├─────────────┤ │                     │ ├─────────────┤ │
│ │ Components  │ │                     │ │  Services   │ │
│ │             │ │                     │ │             │ │
│ ├─────────────┤ │                     │ ├─────────────┤ │
│ │   Stores    │ │                     │ │ Repository  │ │
│ │             │ │                     │ │             │ │
│ ├─────────────┤ │                     │ └─────────────┘ │
│ │   Utils     │ │                     │                 │
│ └─────────────┘ │                     │                 │
└─────────────────┘                     └─────────────────┘
```

### フロントエンドアーキテクチャ

```
src/
├── routes/                 # SvelteKitページルーティング
│   ├── +layout.svelte     # 共通レイアウト
│   ├── +page.svelte       # ホームページ（トーナメント表示）
│   ├── login/             # ログインページ
│   └── admin/             # 管理者ダッシュボード
├── lib/
│   ├── components/        # 再利用可能コンポーネント
│   │   ├── TournamentBracket.svelte
│   │   ├── MatchCard.svelte
│   │   ├── AdminMatchForm.svelte
│   │   └── LoadingSpinner.svelte
│   ├── stores/           # Svelteストア（状態管理）
│   │   ├── auth.js       # 認証状態
│   │   ├── tournament.js # トーナメントデータ
│   │   └── ui.js         # UI状態
│   ├── api/              # APIクライアント
│   │   ├── client.js     # HTTP クライアント設定
│   │   ├── auth.js       # 認証API
│   │   ├── tournament.js # トーナメントAPI
│   │   └── matches.js    # 試合API
│   └── utils/            # ユーティリティ関数
│       ├── validation.js # フォーム検証
│       ├── formatting.js # データフォーマット
│       └── storage.js    # ローカルストレージ
└── app.html              # HTMLテンプレート
```

## コンポーネントとインターフェース

### 1. ページコンポーネント

**HomePage (+page.svelte)**
- トーナメントブラケット表示のメインページ
- スポーツタブ切り替え機能
- リアルタイム更新対応

**LoginPage (login/+page.svelte)**
- 管理者ログインフォーム
- 認証エラーハンドリング
- ログイン成功時のリダイレクト

**AdminDashboard (admin/+page.svelte)**
- 管理者専用ダッシュボード
- 試合結果入力機能
- トーナメント形式切り替え

### 2. 再利用可能コンポーネント

**TournamentBracket.svelte**
```javascript
// Props
export let sport = 'volleyball';
export let matches = [];
export let isAdmin = false;

// 機能
- ブラケット形式でのトーナメント表示
- 試合結果の視覚的表現
- レスポンシブレイアウト
- 管理者向け編集リンク
```

**MatchCard.svelte**
```javascript
// Props
export let match = {};
export let editable = false;

// 機能
- 個別試合情報の表示
- スコア表示/入力
- 試合ステータス表示
- 編集モード切り替え
```

**AdminMatchForm.svelte**
```javascript
// Props
export let match = {};
export let onSubmit = () => {};

// 機能
- 試合結果入力フォーム
- リアルタイム検証
- 送信処理とエラーハンドリング
```

### 3. ストア（状態管理）

**authStore (lib/stores/auth.js)**
```javascript
import { writable } from 'svelte/store';

export const authStore = writable({
  isAuthenticated: false,
  token: null,
  user: null,
  loading: false
});

// メソッド
- login(credentials)
- logout()
- checkAuthStatus()
- refreshToken()
```

**tournamentStore (lib/stores/tournament.js)**
```javascript
import { writable } from 'svelte/store';

export const tournamentStore = writable({
  tournaments: {},
  currentSport: 'volleyball',
  loading: false,
  error: null
});

// メソッド
- fetchTournaments()
- updateMatch(matchId, result)
- switchSport(sport)
- refreshData()
```

**uiStore (lib/stores/ui.js)**
```javascript
import { writable } from 'svelte/store';

export const uiStore = writable({
  notifications: [],
  loading: false,
  theme: 'light'
});

// メソッド
- showNotification(message, type)
- setLoading(state)
- clearNotifications()
```

### 4. APIクライアント

**APIClient (lib/api/client.js)**
```javascript
class APIClient {
  constructor(baseURL) {
    this.baseURL = baseURL;
    this.token = null;
  }

  // メソッド
  setToken(token)
  get(endpoint, options)
  post(endpoint, data, options)
  put(endpoint, data, options)
  delete(endpoint, options)
  handleResponse(response)
  handleError(error)
}
```

**AuthAPI (lib/api/auth.js)**
```javascript
// 認証関連API呼び出し
- login(username, password)
- logout()
- refreshToken()
- validateToken()
```

**TournamentAPI (lib/api/tournament.js)**
```javascript
// トーナメント関連API呼び出し
- getTournaments()
- getTournament(sport)
- getTournamentBracket(sport)
- updateTournamentFormat(sport, format)
```

**MatchAPI (lib/api/matches.js)**
```javascript
// 試合関連API呼び出し
- getMatches(sport)
- getMatch(id)
- updateMatch(id, result)
- createMatch(matchData)
```

## データモデル

### フロントエンド データ型

**Tournament**
```typescript
interface Tournament {
  id: number;
  sport: 'volleyball' | 'table_tennis' | 'soccer';
  format: string;
  status: 'active' | 'completed';
  created_at: string;
  updated_at: string;
}
```

**Match**
```typescript
interface Match {
  id: number;
  tournament_id: number;
  round: string;
  team1: string;
  team2: string;
  score1?: number;
  score2?: number;
  winner?: string;
  status: 'pending' | 'completed';
  scheduled_at: string;
  completed_at?: string;
}
```

**Bracket**
```typescript
interface Bracket {
  tournament_id: number;
  sport: string;
  format: string;
  rounds: Round[];
}

interface Round {
  name: string;
  matches: Match[];
}
```

**AuthState**
```typescript
interface AuthState {
  isAuthenticated: boolean;
  token: string | null;
  user: User | null;
  loading: boolean;
}
```

### API レスポンス形式

**成功レスポンス**
```json
{
  "success": true,
  "data": { ... },
  "message": "操作が成功しました"
}
```

**エラーレスポンス**
```json
{
  "success": false,
  "error": "エラーコード",
  "message": "エラーメッセージ",
  "details": { ... }
}
```

## ユーザーインターフェース設計

### 1. レイアウト構造

**共通レイアウト (+layout.svelte)**
- ヘッダー（ナビゲーション、ログイン状態）
- メインコンテンツエリア
- フッター
- 通知システム
- ローディングオーバーレイ

### 2. ページレイアウト

**ホームページ**
```
┌─────────────────────────────────────┐
│ Header (Logo, Login/Admin Button)   │
├─────────────────────────────────────┤
│ Sport Tabs [Volleyball|Table Tennis|Soccer] │
├─────────────────────────────────────┤
│                                     │
│        Tournament Bracket           │
│                                     │
│  ┌─────┐    ┌─────┐    ┌─────┐     │
│  │Match│ -> │Match│ -> │Final│     │
│  │ 1-2 │    │ 5-6 │    │     │     │
│  └─────┘    └─────┘    └─────┘     │
│                                     │
└─────────────────────────────────────┘
```

**管理ダッシュボード**
```
┌─────────────────────────────────────┐
│ Admin Header (Logout Button)        │
├─────────────────────────────────────┤
│ Sport Selection & Format Toggle     │
├─────────────────────────────────────┤
│ Pending Matches List                │
│ ┌─────────────────────────────────┐ │
│ │ Match: Team A vs Team B         │ │
│ │ [Edit] [Complete]               │ │
│ └─────────────────────────────────┘ │
│ ┌─────────────────────────────────┐ │
│ │ Match: Team C vs Team D         │ │
│ │ [Edit] [Complete]               │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

### 3. レスポンシブデザイン

**ブレークポイント**
- Mobile: < 768px
- Tablet: 768px - 1024px
- Desktop: > 1024px

**モバイル最適化**
- タッチフレンドリーなボタンサイズ
- スワイプ可能なタブ
- 縦向きブラケット表示
- 簡略化されたナビゲーション

## 状態管理とデータフロー

### 1. 状態管理パターン

```
User Action -> Store Update -> Component Re-render -> API Call -> Store Update
```

### 2. データフロー例

**試合結果更新フロー**
1. 管理者が試合結果を入力
2. AdminMatchForm が onSubmit イベントを発火
3. matchAPI.updateMatch() が呼び出される
4. 成功時、tournamentStore が更新される
5. 全ての関連コンポーネントが再レンダリング
6. 成功通知が表示される

### 3. リアルタイム更新

**ポーリング戦略**
```javascript
// 30秒ごとにトーナメントデータを更新
setInterval(() => {
  if (!document.hidden) {
    tournamentStore.refreshData();
  }
}, 30000);
```

**WebSocket統合（将来の拡張）**
- リアルタイム試合結果更新
- 管理者アクション通知
- 接続状態管理

## セキュリティ考慮事項

### 1. 認証とセッション管理

**JWT トークン管理**
- ローカルストレージでのトークン保存
- 自動トークンリフレッシュ
- トークン期限切れ時の自動ログアウト
- XSS攻撃対策

**ルートガード**
```javascript
// 管理者専用ページの保護
export async function load({ url, fetch }) {
  const token = getStoredToken();
  if (!token || !isValidToken(token)) {
    throw redirect(302, '/login');
  }
}
```

### 2. データ検証

**クライアントサイド検証**
- フォーム入力の即座検証
- 型安全性の確保
- サニタイゼーション

**API通信セキュリティ**
- HTTPS通信の強制
- CORS設定の適切な管理
- CSRFトークンの実装

## パフォーマンス最適化

### 1. バンドル最適化

**Vite設定**
```javascript
// vite.config.js
export default {
  build: {
    rollupOptions: {
      output: {
        manualChunks: {
          vendor: ['svelte'],
          api: ['./src/lib/api']
        }
      }
    }
  }
};
```

### 2. データキャッシュ戦略

**ブラウザキャッシュ**
- 静的アセットの長期キャッシュ
- APIレスポンスの適切なキャッシュヘッダー
- Service Worker による オフライン対応

**メモリキャッシュ**
```javascript
// トーナメントデータのメモリキャッシュ
const cache = new Map();
const CACHE_DURATION = 5 * 60 * 1000; // 5分

function getCachedData(key) {
  const cached = cache.get(key);
  if (cached && Date.now() - cached.timestamp < CACHE_DURATION) {
    return cached.data;
  }
  return null;
}
```

### 3. レンダリング最適化

**遅延読み込み**
- ルートベースのコード分割
- 画像の遅延読み込み
- 大きなコンポーネントの動的インポート

**仮想化**
- 大量の試合データの効率的な表示
- スクロール最適化

## テスト戦略

### 1. 単体テスト

**コンポーネントテスト**
```javascript
// TournamentBracket.test.js
import { render, screen } from '@testing-library/svelte';
import TournamentBracket from '../TournamentBracket.svelte';

test('displays tournament matches correctly', () => {
  const matches = [/* test data */];
  render(TournamentBracket, { props: { matches } });
  
  expect(screen.getByText('Team A vs Team B')).toBeInTheDocument();
});
```

**ストアテスト**
```javascript
// auth.test.js
import { get } from 'svelte/store';
import { authStore, login } from '../stores/auth.js';

test('login updates auth state', async () => {
  await login('admin', 'password');
  const state = get(authStore);
  
  expect(state.isAuthenticated).toBe(true);
  expect(state.token).toBeTruthy();
});
```

### 2. 統合テスト

**E2Eテスト (Playwright)**
```javascript
// tournament.spec.js
import { test, expect } from '@playwright/test';

test('admin can update match results', async ({ page }) => {
  await page.goto('/login');
  await page.fill('[data-testid=username]', 'admin');
  await page.fill('[data-testid=password]', 'password');
  await page.click('[data-testid=login-button]');
  
  await page.goto('/admin');
  await page.click('[data-testid=edit-match-1]');
  await page.fill('[data-testid=score1]', '3');
  await page.fill('[data-testid=score2]', '1');
  await page.click('[data-testid=submit-result]');
  
  await expect(page.locator('[data-testid=success-message]')).toBeVisible();
});
```

### 3. APIモック

**MSW (Mock Service Worker)**
```javascript
// mocks/handlers.js
import { rest } from 'msw';

export const handlers = [
  rest.post('/api/auth/login', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        data: { token: 'mock-jwt-token' }
      })
    );
  }),
  
  rest.get('/api/tournaments/:sport', (req, res, ctx) => {
    return res(
      ctx.json({
        success: true,
        data: { /* mock tournament data */ }
      })
    );
  })
];
```

## デプロイメントと運用

### 1. ビルドプロセス

**本番ビルド**
```bash
npm run build
```

**Docker統合**
```dockerfile
# Dockerfile
FROM node:18-alpine as builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM nginx:alpine
COPY --from=builder /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/nginx.conf
```

### 2. 環境設定

**環境変数**
```javascript
// .env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_APP_TITLE=Tournament Management System
VITE_ENABLE_POLLING=true
VITE_POLLING_INTERVAL=30000
```

### 3. 監視とログ

**エラー追跡**
- フロントエンドエラーの自動収集
- ユーザーアクションの追跡
- パフォーマンスメトリクスの監視

**ログ出力**
```javascript
// utils/logger.js
export const logger = {
  info: (message, data) => console.log(`[INFO] ${message}`, data),
  error: (message, error) => console.error(`[ERROR] ${message}`, error),
  warn: (message, data) => console.warn(`[WARN] ${message}`, data)
};
```
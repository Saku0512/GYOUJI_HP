# トーナメント管理システム - 開発者ガイド

## 目次

1. [プロジェクト概要](#プロジェクト概要)
2. [アーキテクチャ](#アーキテクチャ)
3. [開発環境のセットアップ](#開発環境のセットアップ)
4. [プロジェクト構造](#プロジェクト構造)
5. [コンポーネントガイド](#コンポーネントガイド)
6. [API使用方法](#api使用方法)
7. [状態管理](#状態管理)
8. [テスト](#テスト)
9. [コーディング規約](#コーディング規約)
10. [ベストプラクティス](#ベストプラクティス)
11. [トラブルシューティング](#トラブルシューティング)

## プロジェクト概要

トーナメント管理システムは、バレーボール、卓球、8人制サッカーのトーナメント管理を行うフルスタックWebアプリケーションです。

### 主要機能

- **一般ユーザー機能**
  - トーナメントブラケットの閲覧
  - リアルタイム試合結果更新
  - レスポンシブデザイン対応

- **管理者機能**
  - 管理者認証・ログイン
  - 試合結果の入力・更新
  - トーナメント形式の切り替え（卓球の晴天時/雨天時）

### 技術スタック

**フロントエンド**
- **フレームワーク**: SvelteKit 2.14.0
- **言語**: JavaScript (TypeScript設定済み)
- **ビルドツール**: Vite 7.1.2
- **テスト**: Vitest + Playwright
- **スタイリング**: CSS + レスポンシブデザイン
- **状態管理**: Svelte Stores

**バックエンド**
- **言語**: Go 1.24.2
- **フレームワーク**: Gin
- **データベース**: MySQL
- **認証**: JWT + bcrypt

## アーキテクチャ

### システム全体アーキテクチャ

```
┌─────────────────┐    HTTP/REST API    ┌─────────────────┐
│   SvelteKit     │ ←─────────────────→ │   Go Backend    │
│   Frontend      │                     │   (Gin + MySQL) │
│                 │                     │                 │
│ ┌─────────────┐ │                     │ ┌─────────────┐ │
│ │   Routes    │ │                     │ │  Handlers   │ │
│ ├─────────────┤ │                     │ ├─────────────┤ │
│ │ Components  │ │                     │ │  Services   │ │
│ ├─────────────┤ │                     │ ├─────────────┤ │
│ │   Stores    │ │                     │ │ Repository  │ │
│ ├─────────────┤ │                     │ └─────────────┘ │
│ │   Utils     │ │                     │                 │
│ └─────────────┘ │                     │                 │
└─────────────────┘                     └─────────────────┘
```

### フロントエンドアーキテクチャ

- **Clean Architecture**: レイヤー分離による保守性の向上
- **Component-Based**: 再利用可能なコンポーネント設計
- **Reactive State Management**: Svelte Storesによるリアクティブな状態管理
- **API Layer**: 統一されたAPI通信レイヤー

## 開発環境のセットアップ

### 前提条件

- Node.js 18以上
- npm または yarn
- Go 1.24.2以上（バックエンド開発時）
- MySQL 8.0以上（バックエンド開発時）
- Docker & Docker Compose（推奨）

### フロントエンド開発環境

1. **依存関係のインストール**
```bash
cd frontend
npm install
```

2. **環境変数の設定**
```bash
cp .env.example .env
```

`.env`ファイルを編集：
```env
VITE_API_BASE_URL=http://localhost:8080/api
VITE_APP_TITLE=Tournament Management System
VITE_ENABLE_POLLING=true
VITE_POLLING_INTERVAL=30000
```

3. **開発サーバーの起動**
```bash
npm run dev
```

4. **テストの実行**
```bash
# 単体テスト
npm run test

# E2Eテスト
npm run test:e2e

# テスト監視モード
npm run test:watch
```

### Docker環境での開発

1. **全体システムの起動**
```bash
# プロジェクトルートで実行
docker-compose up -d
```

2. **フロントエンドのみの開発**
```bash
cd frontend
npm run dev
```

## プロジェクト構造

```
frontend/
├── src/
│   ├── routes/                 # SvelteKitページルーティング
│   │   ├── +layout.svelte     # 共通レイアウト
│   │   ├── +page.svelte       # ホームページ
│   │   ├── login/             # ログインページ
│   │   ├── admin/             # 管理者ダッシュボード
│   │   └── offline/           # オフラインページ
│   ├── lib/
│   │   ├── components/        # 再利用可能コンポーネント
│   │   ├── stores/           # 状態管理ストア
│   │   ├── api/              # APIクライアント
│   │   ├── utils/            # ユーティリティ関数
│   │   └── styles/           # 共通スタイル
│   ├── assets/               # 静的アセット
│   ├── app.html              # HTMLテンプレート
│   └── app.css               # グローバルスタイル
├── static/                   # 静的ファイル
├── tests/                    # テストファイル
├── playwright.config.js      # E2Eテスト設定
├── vite.config.js           # Vite設定
├── svelte.config.js         # Svelte設定
└── package.json             # 依存関係とスクリプト
```

### 主要ディレクトリの説明

#### `src/routes/`
SvelteKitのファイルベースルーティング。各ディレクトリが URLパスに対応。

#### `src/lib/components/`
再利用可能なSvelteコンポーネント。機能別に分類：
- **UI Components**: Button, Input, Modal など
- **Business Components**: TournamentBracket, MatchCard など
- **Layout Components**: ResponsiveLayout, Navigation など

#### `src/lib/stores/`
Svelte Storesによる状態管理：
- `auth.js`: 認証状態
- `tournament.js`: トーナメントデータ
- `ui.js`: UI状態（通知、ローディングなど）

#### `src/lib/api/`
バックエンドAPI通信レイヤー：
- `client.js`: HTTP クライアント基底クラス
- `auth.js`: 認証API
- `tournament.js`: トーナメントAPI
- `matches.js`: 試合API

#### `src/lib/utils/`
共通ユーティリティ関数：
- `validation.js`: フォーム検証
- `formatting.js`: データフォーマット
- `storage.js`: ローカルストレージ管理
- `cache.js`: キャッシュ管理

## コンポーネントガイド

### 基本的なコンポーネント使用方法

#### TournamentBracket
トーナメントブラケットを表示するメインコンポーネント。

```svelte
<script>
  import TournamentBracket from '$lib/components/TournamentBracket.svelte';
  
  let matches = [
    {
      id: 1,
      team1: 'チームA',
      team2: 'チームB',
      score1: 3,
      score2: 1,
      winner: 'チームA'
    }
  ];
</script>

<TournamentBracket 
  sport="volleyball" 
  {matches} 
  isAdmin={false} 
/>
```

**Props:**
- `sport`: スポーツ種目 ('volleyball' | 'table_tennis' | 'soccer')
- `matches`: 試合データ配列
- `isAdmin`: 管理者モード (boolean)

#### MatchCard
個別試合情報を表示するコンポーネント。

```svelte
<script>
  import MatchCard from '$lib/components/MatchCard.svelte';
  
  let match = {
    id: 1,
    team1: 'チームA',
    team2: 'チームB',
    score1: 3,
    score2: 1,
    status: 'completed'
  };
</script>

<MatchCard {match} editable={true} />
```

**Props:**
- `match`: 試合データオブジェクト
- `editable`: 編集可能モード (boolean)

#### AdminMatchForm
管理者用試合結果入力フォーム。

```svelte
<script>
  import AdminMatchForm from '$lib/components/AdminMatchForm.svelte';
  
  function handleSubmit(result) {
    console.log('試合結果:', result);
  }
</script>

<AdminMatchForm 
  match={selectedMatch} 
  onSubmit={handleSubmit} 
/>
```

### レスポンシブコンポーネント

#### ResponsiveLayout
画面サイズに応じてレイアウトを調整するコンポーネント。

```svelte
<script>
  import ResponsiveLayout from '$lib/components/ResponsiveLayout.svelte';
</script>

<ResponsiveLayout>
  <div slot="mobile">モバイル用コンテンツ</div>
  <div slot="tablet">タブレット用コンテンツ</div>
  <div slot="desktop">デスクトップ用コンテンツ</div>
</ResponsiveLayout>
```

### UI コンポーネント

#### Button
統一されたボタンコンポーネント。

```svelte
<script>
  import Button from '$lib/components/Button.svelte';
</script>

<Button 
  variant="primary" 
  size="medium" 
  disabled={false}
  on:click={handleClick}
>
  ボタンテキスト
</Button>
```

**Props:**
- `variant`: 'primary' | 'secondary' | 'danger'
- `size`: 'small' | 'medium' | 'large'
- `disabled`: boolean

## API使用方法

### APIクライアントの基本使用方法

#### 認証API

```javascript
import { authAPI } from '$lib/api/auth.js';

// ログイン
try {
  const result = await authAPI.login('username', 'password');
  console.log('ログイン成功:', result);
} catch (error) {
  console.error('ログイン失敗:', error);
}

// ログアウト
await authAPI.logout();

// トークン検証
const isValid = await authAPI.validateToken();
```

#### トーナメントAPI

```javascript
import { tournamentAPI } from '$lib/api/tournament.js';

// 全トーナメント取得
const tournaments = await tournamentAPI.getTournaments();

// 特定スポーツのトーナメント取得
const volleyball = await tournamentAPI.getTournament('volleyball');

// ブラケット取得
const bracket = await tournamentAPI.getTournamentBracket('volleyball');

// 形式切り替え（卓球のみ）
await tournamentAPI.updateTournamentFormat('table_tennis', 'rainy');
```

#### 試合API

```javascript
import { matchAPI } from '$lib/api/matches.js';

// 試合一覧取得
const matches = await matchAPI.getMatches('volleyball');

// 試合詳細取得
const match = await matchAPI.getMatch(1);

// 試合結果更新
await matchAPI.updateMatch(1, {
  score1: 3,
  score2: 1,
  winner: 'チームA'
});
```

### エラーハンドリング

APIクライアントは統一されたエラーハンドリングを提供します：

```javascript
try {
  const result = await tournamentAPI.getTournaments();
} catch (error) {
  if (error.status === 401) {
    // 認証エラー
    console.log('認証が必要です');
  } else if (error.status === 500) {
    // サーバーエラー
    console.log('サーバーエラーが発生しました');
  } else {
    // その他のエラー
    console.log('エラー:', error.message);
  }
}
```

## 状態管理

### Svelte Stores の使用方法

#### 認証ストア (authStore)

```javascript
import { authStore } from '$lib/stores/auth.js';

// ストアの購読
$: isAuthenticated = $authStore.isAuthenticated;
$: user = $authStore.user;
$: loading = $authStore.loading;

// アクションの実行
import { login, logout, checkAuthStatus } from '$lib/stores/auth.js';

// ログイン
await login('username', 'password');

// ログアウト
await logout();

// 認証状態確認
await checkAuthStatus();
```

#### トーナメントストア (tournamentStore)

```javascript
import { tournamentStore } from '$lib/stores/tournament.js';
import { 
  fetchTournaments, 
  updateMatch, 
  switchSport 
} from '$lib/stores/tournament.js';

// ストアの購読
$: tournaments = $tournamentStore.tournaments;
$: currentSport = $tournamentStore.currentSport;
$: loading = $tournamentStore.loading;

// データ取得
await fetchTournaments();

// スポーツ切り替え
await switchSport('table_tennis');

// 試合結果更新
await updateMatch(1, { score1: 3, score2: 1 });
```

#### UIストア (uiStore)

```javascript
import { uiStore } from '$lib/stores/ui.js';
import { 
  showNotification, 
  setLoading, 
  clearNotifications 
} from '$lib/stores/ui.js';

// 通知表示
showNotification('操作が完了しました', 'success');
showNotification('エラーが発生しました', 'error');

// ローディング状態設定
setLoading(true);
// ... 非同期処理
setLoading(false);

// 通知クリア
clearNotifications();
```

### リアクティブな状態更新

Svelteの`$:`構文を使用してリアクティブな更新を実装：

```svelte
<script>
  import { tournamentStore } from '$lib/stores/tournament.js';
  
  // リアクティブな値
  $: tournaments = $tournamentStore.tournaments;
  $: currentSport = $tournamentStore.currentSport;
  
  // リアクティブな計算
  $: currentTournament = tournaments[currentSport];
  $: hasMatches = currentTournament?.matches?.length > 0;
  
  // リアクティブな副作用
  $: if (currentSport) {
    console.log('スポーツが変更されました:', currentSport);
  }
</script>

{#if hasMatches}
  <TournamentBracket matches={currentTournament.matches} />
{:else}
  <p>試合データがありません</p>
{/if}
```

## テスト

### 単体テスト (Vitest)

#### コンポーネントテスト

```javascript
// TournamentBracket.test.js
import { render, screen } from '@testing-library/svelte';
import { describe, it, expect } from 'vitest';
import TournamentBracket from '../TournamentBracket.svelte';

describe('TournamentBracket', () => {
  it('試合データを正しく表示する', () => {
    const matches = [
      {
        id: 1,
        team1: 'チームA',
        team2: 'チームB',
        score1: 3,
        score2: 1
      }
    ];
    
    render(TournamentBracket, { 
      props: { matches, sport: 'volleyball' } 
    });
    
    expect(screen.getByText('チームA')).toBeInTheDocument();
    expect(screen.getByText('チームB')).toBeInTheDocument();
  });
});
```

#### ストアテスト

```javascript
// auth.test.js
import { get } from 'svelte/store';
import { describe, it, expect, beforeEach } from 'vitest';
import { authStore, login } from '../auth.js';

describe('authStore', () => {
  beforeEach(() => {
    // ストアをリセット
    authStore.set({
      isAuthenticated: false,
      token: null,
      user: null,
      loading: false
    });
  });

  it('ログイン時に認証状態を更新する', async () => {
    await login('admin', 'password');
    const state = get(authStore);
    
    expect(state.isAuthenticated).toBe(true);
    expect(state.token).toBeTruthy();
  });
});
```

#### APIテスト

```javascript
// tournament.test.js
import { describe, it, expect, vi } from 'vitest';
import { tournamentAPI } from '../tournament.js';

// APIクライアントをモック
vi.mock('../client.js', () => ({
  apiClient: {
    get: vi.fn()
  }
}));

describe('tournamentAPI', () => {
  it('トーナメント一覧を取得する', async () => {
    const mockData = { tournaments: [] };
    apiClient.get.mockResolvedValue({ data: mockData });
    
    const result = await tournamentAPI.getTournaments();
    
    expect(result).toEqual(mockData);
    expect(apiClient.get).toHaveBeenCalledWith('/tournaments');
  });
});
```

### E2Eテスト (Playwright)

#### 基本的なE2Eテスト

```javascript
// tournament.spec.js
import { test, expect } from '@playwright/test';

test.describe('トーナメント表示', () => {
  test('ホームページでトーナメントが表示される', async ({ page }) => {
    await page.goto('/');
    
    // スポーツタブが表示される
    await expect(page.locator('[data-testid=sport-tabs]')).toBeVisible();
    
    // バレーボールタブをクリック
    await page.click('[data-testid=volleyball-tab]');
    
    // ブラケットが表示される
    await expect(page.locator('[data-testid=tournament-bracket]')).toBeVisible();
  });
});

test.describe('管理者機能', () => {
  test('管理者が試合結果を入力できる', async ({ page }) => {
    // ログイン
    await page.goto('/login');
    await page.fill('[data-testid=username]', 'admin');
    await page.fill('[data-testid=password]', 'password');
    await page.click('[data-testid=login-button]');
    
    // 管理ダッシュボードに移動
    await page.goto('/admin');
    
    // 試合を選択
    await page.click('[data-testid=edit-match-1]');
    
    // 結果を入力
    await page.fill('[data-testid=score1]', '3');
    await page.fill('[data-testid=score2]', '1');
    await page.click('[data-testid=submit-result]');
    
    // 成功メッセージを確認
    await expect(page.locator('[data-testid=success-message]')).toBeVisible();
  });
});
```

### テスト実行コマンド

```bash
# 単体テスト実行
npm run test

# 単体テスト監視モード
npm run test:watch

# E2Eテスト実行
npm run test:e2e

# E2EテストUI表示
npm run test:e2e:ui

# カバレッジ付きテスト
npm run test -- --coverage
```

## コーディング規約

### JavaScript/Svelte規約

#### 命名規則

```javascript
// 変数・関数: camelCase
const userName = 'admin';
const fetchTournaments = async () => {};

// 定数: UPPER_SNAKE_CASE
const API_BASE_URL = 'http://localhost:8080/api';
const MAX_RETRY_COUNT = 3;

// コンポーネント: PascalCase
import TournamentBracket from './TournamentBracket.svelte';

// ファイル名: kebab-case
// tournament-bracket.svelte
// api-client.js
```

#### コメント規約

```javascript
/**
 * トーナメントデータを取得する
 * @param {string} sport - スポーツ種目
 * @returns {Promise<Object>} トーナメントデータ
 */
async function fetchTournament(sport) {
  // APIからデータを取得
  const response = await apiClient.get(`/tournaments/${sport}`);
  
  // データを正規化
  return normalizeData(response.data);
}
```

#### エラーハンドリング

```javascript
// 適切なエラーハンドリング
try {
  const result = await apiCall();
  return result;
} catch (error) {
  // ログ出力
  console.error('API呼び出しエラー:', error);
  
  // ユーザーフレンドリーなエラーメッセージ
  throw new Error('データの取得に失敗しました');
}

// エラー境界の使用
<ErrorBoundary>
  <TournamentBracket {matches} />
</ErrorBoundary>
```

### Svelteコンポーネント規約

#### コンポーネント構造

```svelte
<script>
  // 1. インポート
  import { onMount } from 'svelte';
  import Button from './Button.svelte';
  
  // 2. Props
  export let matches = [];
  export let sport = 'volleyball';
  export let isAdmin = false;
  
  // 3. 内部変数
  let loading = false;
  let error = null;
  
  // 4. リアクティブ宣言
  $: hasMatches = matches.length > 0;
  $: currentMatches = matches.filter(m => m.sport === sport);
  
  // 5. 関数
  function handleMatchClick(match) {
    console.log('試合クリック:', match);
  }
  
  // 6. ライフサイクル
  onMount(() => {
    console.log('コンポーネントマウント');
  });
</script>

<!-- 7. HTML -->
<div class="tournament-bracket">
  {#if loading}
    <LoadingSpinner />
  {:else if error}
    <ErrorMessage {error} />
  {:else if hasMatches}
    {#each currentMatches as match (match.id)}
      <MatchCard {match} on:click={() => handleMatchClick(match)} />
    {/each}
  {:else}
    <p>試合データがありません</p>
  {/if}
</div>

<!-- 8. CSS -->
<style>
  .tournament-bracket {
    display: grid;
    gap: 1rem;
    padding: 1rem;
  }
  
  @media (max-width: 768px) {
    .tournament-bracket {
      grid-template-columns: 1fr;
    }
  }
</style>
```

### CSS規約

#### クラス命名 (BEM)

```css
/* Block */
.tournament-bracket { }

/* Element */
.tournament-bracket__match { }
.tournament-bracket__score { }

/* Modifier */
.tournament-bracket--mobile { }
.tournament-bracket__match--completed { }
```

#### レスポンシブデザイン

```css
/* モバイルファースト */
.component {
  /* モバイル用スタイル */
}

@media (min-width: 768px) {
  .component {
    /* タブレット用スタイル */
  }
}

@media (min-width: 1024px) {
  .component {
    /* デスクトップ用スタイル */
  }
}
```

## ベストプラクティス

### パフォーマンス最適化

#### 1. コンポーネントの遅延読み込み

```javascript
// 動的インポート
const LazyComponent = lazy(() => import('./HeavyComponent.svelte'));

// 条件付き読み込み
{#if showHeavyComponent}
  {#await import('./HeavyComponent.svelte') then { default: HeavyComponent }}
    <HeavyComponent />
  {/await}
{/if}
```

#### 2. データキャッシュ

```javascript
// メモリキャッシュ
const cache = new Map();

async function getCachedData(key) {
  if (cache.has(key)) {
    return cache.get(key);
  }
  
  const data = await fetchData(key);
  cache.set(key, data);
  return data;
}
```

#### 3. 画像最適化

```svelte
<script>
  import LazyImage from '$lib/components/LazyImage.svelte';
</script>

<LazyImage 
  src="/images/tournament.jpg" 
  alt="トーナメント画像"
  loading="lazy"
/>
```

### セキュリティ

#### 1. XSS対策

```svelte
<!-- 安全な文字列表示 -->
<p>{userInput}</p>

<!-- HTMLを表示する場合は検証済みデータのみ -->
<div>{@html sanitizedHtml}</div>
```

#### 2. 認証トークン管理

```javascript
// トークンの安全な保存
import { browser } from '$app/environment';

function storeToken(token) {
  if (browser) {
    localStorage.setItem('auth_token', token);
  }
}

function getToken() {
  if (browser) {
    return localStorage.getItem('auth_token');
  }
  return null;
}
```

### アクセシビリティ

#### 1. セマンティックHTML

```svelte
<nav aria-label="メインナビゲーション">
  <ul>
    <li><a href="/" aria-current="page">ホーム</a></li>
    <li><a href="/admin">管理画面</a></li>
  </ul>
</nav>

<main>
  <h1>トーナメント結果</h1>
  <section aria-labelledby="volleyball-heading">
    <h2 id="volleyball-heading">バレーボール</h2>
    <!-- コンテンツ -->
  </section>
</main>
```

#### 2. キーボードナビゲーション

```svelte
<button 
  on:click={handleClick}
  on:keydown={(e) => e.key === 'Enter' && handleClick()}
  tabindex="0"
>
  ボタン
</button>
```

### エラーハンドリング

#### 1. エラー境界

```svelte
<!-- ErrorBoundary.svelte -->
<script>
  export let fallback = null;
  let error = null;
  
  function handleError(event) {
    error = event.detail;
  }
</script>

<svelte:window on:error={handleError} />

{#if error}
  {#if fallback}
    <svelte:component this={fallback} {error} />
  {:else}
    <div class="error-boundary">
      <h2>エラーが発生しました</h2>
      <p>{error.message}</p>
    </div>
  {/if}
{:else}
  <slot />
{/if}
```

#### 2. 統一されたエラー処理

```javascript
// error-handler.js
export class AppError extends Error {
  constructor(message, code, status) {
    super(message);
    this.code = code;
    this.status = status;
  }
}

export function handleApiError(error) {
  if (error.status === 401) {
    // 認証エラー
    goto('/login');
  } else if (error.status >= 500) {
    // サーバーエラー
    showNotification('サーバーエラーが発生しました', 'error');
  } else {
    // その他のエラー
    showNotification(error.message, 'error');
  }
}
```

## トラブルシューティング

### よくある問題と解決方法

#### 1. ビルドエラー

**問題**: `npm run build` でエラーが発生する

**解決方法**:
```bash
# node_modulesを削除して再インストール
rm -rf node_modules package-lock.json
npm install

# キャッシュをクリア
npm run dev -- --force
```

#### 2. API接続エラー

**問題**: フロントエンドからバックエンドAPIに接続できない

**解決方法**:
1. 環境変数を確認
```bash
# .envファイルの確認
cat .env
```

2. バックエンドサーバーの起動確認
```bash
# バックエンドの健康チェック
curl http://localhost:8080/health
```

3. CORS設定の確認

#### 3. 認証エラー

**問題**: ログイン後に認証が維持されない

**解決方法**:
1. ローカルストレージの確認
```javascript
// ブラウザコンソールで実行
console.log(localStorage.getItem('auth_token'));
```

2. トークンの有効期限確認
3. APIレスポンスの確認

#### 4. パフォーマンス問題

**問題**: ページの読み込みが遅い

**解決方法**:
1. バンドルサイズの分析
```bash
npm run analyze:bundle
```

2. ネットワークタブでリクエストを確認
3. 不要なコンポーネントの遅延読み込み

#### 5. テストエラー

**問題**: テストが失敗する

**解決方法**:
1. テスト環境の確認
```bash
# テスト用データベースの確認
npm run test:setup
```

2. モックの設定確認
3. 非同期処理の適切な待機

### デバッグ方法

#### 1. ブラウザ開発者ツール

- **Console**: エラーログとデバッグ出力
- **Network**: API リクエスト/レスポンス
- **Application**: ローカルストレージ、セッション
- **Performance**: パフォーマンス分析

#### 2. Svelte DevTools

```bash
# Svelte DevToolsの使用
# ブラウザ拡張機能をインストール
# コンポーネントの状態とプロパティを確認
```

#### 3. ログ出力

```javascript
// 開発環境でのみログ出力
import { dev } from '$app/environment';

if (dev) {
  console.log('デバッグ情報:', data);
}
```

### サポートとリソース

#### 公式ドキュメント
- [SvelteKit Documentation](https://kit.svelte.dev/docs)
- [Svelte Documentation](https://svelte.dev/docs)
- [Vite Documentation](https://vitejs.dev/guide/)

#### コミュニティ
- [Svelte Discord](https://svelte.dev/chat)
- [SvelteKit GitHub](https://github.com/sveltejs/kit)

#### 内部リソース
- バックエンドAPI仕様: `backend/docs/swagger.yaml`
- 統合テスト: `backend/integration_test/`
- パフォーマンステスト: `frontend/scripts/performance-test.js`

---

このガイドは継続的に更新されます。質問や改善提案がある場合は、開発チームまでお知らせください。
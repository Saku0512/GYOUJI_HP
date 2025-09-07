# E2Eテスト

このディレクトリには、Playwrightを使用したエンドツーエンド（E2E）テストが含まれています。

## 概要

E2Eテストは、実際のブラウザ環境でアプリケーション全体の動作を検証します。以下の主要な機能をテストしています：

- 認証フロー（ログイン・ログアウト）
- トーナメント表示機能
- 管理者ダッシュボード機能
- ユーザーフロー統合テスト
- クロスブラウザ互換性

## テストファイル構成

```
tests/e2e/
├── auth.spec.js              # 認証機能のテスト
├── tournament-display.spec.js # トーナメント表示のテスト
├── admin-dashboard.spec.js    # 管理者ダッシュボードのテスト
├── user-flows.spec.js         # ユーザーフロー統合テスト
├── cross-browser.spec.js      # クロスブラウザ互換性テスト
├── utils/
│   └── test-helpers.js        # テスト用ヘルパー関数
├── fixtures/
│   └── test-data.js          # テストデータフィクスチャ
└── README.md                 # このファイル
```

## セットアップ

### 前提条件

- Node.js 18以上
- npm または yarn

### インストール

```bash
# 依存関係をインストール
npm install

# Playwrightブラウザをインストール
npx playwright install
```

## テストの実行

### 基本的な実行方法

```bash
# 全てのE2Eテストを実行
npm run test:e2e

# ヘッドレスモードで実行（デフォルト）
npx playwright test

# ヘッド付きモードで実行（ブラウザが表示される）
npm run test:e2e:headed

# デバッグモードで実行
npm run test:e2e:debug

# UIモードで実行（インタラクティブ）
npm run test:e2e:ui
```

### 特定のテストファイルを実行

```bash
# 認証テストのみ実行
npx playwright test auth.spec.js

# 管理者ダッシュボードテストのみ実行
npx playwright test admin-dashboard.spec.js
```

### 特定のブラウザで実行

```bash
# Chromiumのみ
npx playwright test --project=chromium

# Firefoxのみ
npx playwright test --project=firefox

# WebKitのみ
npx playwright test --project=webkit

# モバイルChromeのみ
npx playwright test --project="Mobile Chrome"
```

### 並列実行の制御

```bash
# シーケンシャル実行（並列度1）
npx playwright test --workers=1

# 並列度を指定
npx playwright test --workers=4
```

## テスト結果の確認

### HTMLレポート

```bash
# テスト実行後にHTMLレポートを表示
npm run test:e2e:report

# または
npx playwright show-report
```

### トレース表示

テストが失敗した場合、トレースファイルが生成されます：

```bash
# トレースビューアーで確認
npx playwright show-trace test-results/[test-name]/trace.zip
```

## テストの書き方

### 基本的なテスト構造

```javascript
import { test, expect } from '@playwright/test';
import { loginAsAdmin, mockApiResponse } from './utils/test-helpers.js';

test.describe('機能名', () => {
  test.beforeEach(async ({ page }) => {
    // 各テスト前の共通セットアップ
    await mockApiResponse(page, '/api/endpoint', mockData);
    await page.goto('/');
  });

  test('テストケース名', async ({ page }) => {
    // テストの実装
    await expect(page.locator('[data-testid="element"]')).toBeVisible();
  });
});
```

### ヘルパー関数の使用

```javascript
import { 
  loginAsAdmin, 
  selectSportTab, 
  updateMatchResult,
  expectNotification 
} from './utils/test-helpers.js';

test('管理者が試合結果を更新できる', async ({ page }) => {
  await loginAsAdmin(page);
  await updateMatchResult(page, 1, 3, 1);
  await expectNotification(page, '試合結果を更新しました', 'success');
});
```

### APIモック

```javascript
import { mockApiResponse, mockApiError } from './utils/test-helpers.js';

// 成功レスポンスをモック
await mockApiResponse(page, '/api/tournaments/volleyball', {
  success: true,
  data: { /* データ */ }
});

// エラーレスポンスをモック
await mockApiError(page, '/api/auth/login', 401, '認証に失敗しました');
```

## テストデータ

テストデータは `fixtures/test-data.js` で管理されています：

```javascript
import { mockTournaments, mockUsers, mockApiResponses } from './fixtures/test-data.js';

// 事前定義されたテストデータを使用
await mockApiResponse(page, '/api/tournaments/volleyball', {
  success: true,
  data: mockTournaments.volleyball
});
```

## デバッグ

### デバッグモード

```bash
# デバッグモードで実行（ブラウザが一時停止）
npx playwright test --debug

# 特定のテストをデバッグ
npx playwright test auth.spec.js --debug
```

### ブラウザ開発者ツール

```javascript
test('デバッグ用テスト', async ({ page }) => {
  await page.goto('/');
  
  // ブラウザ開発者ツールを開く
  await page.pause();
  
  // テストの続き...
});
```

### スクリーンショット

```javascript
test('スクリーンショット付きテスト', async ({ page }) => {
  await page.goto('/');
  
  // スクリーンショットを撮影
  await page.screenshot({ path: 'screenshot.png' });
  
  // 要素のスクリーンショット
  await page.locator('[data-testid="element"]').screenshot({ 
    path: 'element.png' 
  });
});
```

## CI/CD統合

GitHub Actionsでの自動実行設定は `.github/workflows/e2e-tests.yml` に定義されています。

### ローカルでCI環境をシミュレート

```bash
# CI環境変数を設定して実行
CI=true npx playwright test

# リトライ回数を指定
npx playwright test --retries=2
```

## トラブルシューティング

### よくある問題

1. **ブラウザが起動しない**
   ```bash
   # ブラウザを再インストール
   npx playwright install --force
   ```

2. **テストがタイムアウトする**
   ```javascript
   // タイムアウト時間を延長
   test.setTimeout(60000); // 60秒
   ```

3. **要素が見つからない**
   ```javascript
   // 要素の出現を待機
   await page.waitForSelector('[data-testid="element"]');
   
   // より長い待機時間を設定
   await expect(page.locator('[data-testid="element"]')).toBeVisible({ 
     timeout: 10000 
   });
   ```

4. **APIモックが効かない**
   ```javascript
   // ルートを先に設定
   await page.route('**/api/**', route => {
     // モック処理
   });
   
   // その後にページに移動
   await page.goto('/');
   ```

### ログの確認

```bash
# 詳細ログを有効にして実行
DEBUG=pw:api npx playwright test

# ブラウザコンソールログを表示
npx playwright test --reporter=line
```

## ベストプラクティス

1. **data-testid属性を使用**
   - CSSセレクターではなく、`data-testid`属性を使用してテストの安定性を向上

2. **適切な待機**
   - `page.waitForSelector()`や`expect().toBeVisible()`を使用して要素の出現を待機

3. **テストの独立性**
   - 各テストは独立して実行できるように設計

4. **APIモック**
   - 外部APIに依存せず、モックを使用してテストの安定性を確保

5. **適切なテストデータ**
   - 実際のデータに近いテストデータを使用

6. **エラーハンドリング**
   - 正常系だけでなく、エラー系のテストも実装

## 参考資料

- [Playwright公式ドキュメント](https://playwright.dev/)
- [Playwright Test API](https://playwright.dev/docs/api/class-test)
- [SvelteKit Testing](https://kit.svelte.dev/docs/testing)
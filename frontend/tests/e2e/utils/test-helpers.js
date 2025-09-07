import { expect } from '@playwright/test';

/**
 * テスト用のヘルパー関数
 */

/**
 * 管理者としてログインする
 * @param {import('@playwright/test').Page} page
 * @param {string} username - ユーザー名（デフォルト: 'admin'）
 * @param {string} password - パスワード（デフォルト: 'password'）
 */
export async function loginAsAdmin(page, username = 'admin', password = 'password') {
  await page.goto('/login');
  
  // ログインフォームの要素を待機
  await page.waitForSelector('[data-testid="username-input"]');
  await page.waitForSelector('[data-testid="password-input"]');
  
  // 認証情報を入力
  await page.fill('[data-testid="username-input"]', username);
  await page.fill('[data-testid="password-input"]', password);
  
  // ログインボタンをクリック
  await page.click('[data-testid="login-button"]');
  
  // 管理ダッシュボードへのリダイレクトを待機
  await page.waitForURL('/admin');
  
  // 認証状態の確認
  await expect(page.locator('[data-testid="admin-header"]')).toBeVisible();
}

/**
 * ログアウトする
 * @param {import('@playwright/test').Page} page
 */
export async function logout(page) {
  await page.click('[data-testid="logout-button"]');
  await page.waitForURL('/login');
  await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
}

/**
 * 特定のスポーツタブを選択する
 * @param {import('@playwright/test').Page} page
 * @param {string} sport - スポーツ名 ('volleyball', 'table_tennis', 'soccer')
 */
export async function selectSportTab(page, sport) {
  const sportTabSelector = `[data-testid="sport-tab-${sport}"]`;
  await page.click(sportTabSelector);
  
  // タブが選択されたことを確認
  await expect(page.locator(sportTabSelector)).toHaveClass(/active|selected/);
  
  // トーナメントブラケットが表示されるまで待機
  await page.waitForSelector('[data-testid="tournament-bracket"]');
}

/**
 * 試合結果を入力する
 * @param {import('@playwright/test').Page} page
 * @param {number} matchId - 試合ID
 * @param {number} score1 - チーム1のスコア
 * @param {number} score2 - チーム2のスコア
 */
export async function updateMatchResult(page, matchId, score1, score2) {
  // 試合編集ボタンをクリック
  await page.click(`[data-testid="edit-match-${matchId}"]`);
  
  // 試合結果入力フォームが表示されるまで待機
  await page.waitForSelector('[data-testid="match-form"]');
  
  // スコアを入力
  await page.fill('[data-testid="score1-input"]', score1.toString());
  await page.fill('[data-testid="score2-input"]', score2.toString());
  
  // 結果を送信
  await page.click('[data-testid="submit-result-button"]');
  
  // 成功メッセージを待機
  await expect(page.locator('[data-testid="success-message"]')).toBeVisible();
}

/**
 * 通知メッセージを確認する
 * @param {import('@playwright/test').Page} page
 * @param {string} message - 期待するメッセージ
 * @param {string} type - 通知タイプ ('success', 'error', 'warning')
 */
export async function expectNotification(page, message, type = 'success') {
  const notificationSelector = `[data-testid="notification-${type}"]`;
  await expect(page.locator(notificationSelector)).toBeVisible();
  await expect(page.locator(notificationSelector)).toContainText(message);
}

/**
 * ローディング状態を待機する
 * @param {import('@playwright/test').Page} page
 */
export async function waitForLoading(page) {
  // ローディングスピナーが表示されるまで待機
  await page.waitForSelector('[data-testid="loading-spinner"]', { state: 'visible' });
  
  // ローディングスピナーが非表示になるまで待機
  await page.waitForSelector('[data-testid="loading-spinner"]', { state: 'hidden' });
}

/**
 * トーナメントブラケットが正しく表示されているかを確認する
 * @param {import('@playwright/test').Page} page
 * @param {string} sport - スポーツ名
 */
export async function expectTournamentBracket(page, sport) {
  // ブラケットコンテナが表示されていることを確認
  await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();
  
  // スポーツ固有のブラケットが表示されていることを確認
  await expect(page.locator(`[data-testid="bracket-${sport}"]`)).toBeVisible();
  
  // 少なくとも1つの試合カードが表示されていることを確認
  await expect(page.locator('[data-testid^="match-card-"]').first()).toBeVisible();
}

/**
 * レスポンシブデザインのテスト用にビューポートを設定する
 * @param {import('@playwright/test').Page} page
 * @param {string} device - デバイスタイプ ('mobile', 'tablet', 'desktop')
 */
export async function setViewport(page, device) {
  const viewports = {
    mobile: { width: 375, height: 667 },
    tablet: { width: 768, height: 1024 },
    desktop: { width: 1920, height: 1080 }
  };
  
  await page.setViewportSize(viewports[device]);
}

/**
 * APIレスポンスをモックする
 * @param {import('@playwright/test').Page} page
 * @param {string} endpoint - APIエンドポイント
 * @param {Object} response - モックレスポンス
 */
export async function mockApiResponse(page, endpoint, response) {
  await page.route(`**/api${endpoint}`, route => {
    route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify(response)
    });
  });
}

/**
 * エラーレスポンスをモックする
 * @param {import('@playwright/test').Page} page
 * @param {string} endpoint - APIエンドポイント
 * @param {number} status - HTTPステータスコード
 * @param {string} message - エラーメッセージ
 */
export async function mockApiError(page, endpoint, status = 500, message = 'Internal Server Error') {
  await page.route(`**/api${endpoint}`, route => {
    route.fulfill({
      status,
      contentType: 'application/json',
      body: JSON.stringify({
        success: false,
        error: 'API_ERROR',
        message
      })
    });
  });
}
import { test, expect } from '@playwright/test';
import { loginAsAdmin, logout, mockApiResponse, mockApiError } from './utils/test-helpers.js';

test.describe('認証フロー', () => {
  test.beforeEach(async ({ page }) => {
    // 各テスト前にホームページに移動
    await page.goto('/');
  });

  test('管理者ログインが成功する', async ({ page }) => {
    // 正常なログインAPIレスポンスをモック
    await mockApiResponse(page, '/auth/login', {
      success: true,
      data: {
        token: 'mock-jwt-token',
        user: {
          id: 1,
          username: 'admin',
          role: 'admin'
        }
      }
    });

    // ログインページに移動
    await page.goto('/login');

    // ログインフォームが表示されることを確認
    await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
    await expect(page.locator('[data-testid="username-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="password-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="login-button"]')).toBeVisible();

    // 認証情報を入力
    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'password');

    // ログインボタンをクリック
    await page.click('[data-testid="login-button"]');

    // 管理ダッシュボードにリダイレクトされることを確認
    await page.waitForURL('/admin');
    await expect(page.locator('[data-testid="admin-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="admin-dashboard"]')).toBeVisible();
  });

  test('無効な認証情報でログインが失敗する', async ({ page }) => {
    // エラーレスポンスをモック
    await mockApiError(page, '/auth/login', 401, '認証情報が正しくありません');

    await page.goto('/login');

    // 無効な認証情報を入力
    await page.fill('[data-testid="username-input"]', 'invalid');
    await page.fill('[data-testid="password-input"]', 'invalid');
    await page.click('[data-testid="login-button"]');

    // エラーメッセージが表示されることを確認
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('認証情報が正しくありません');

    // ログインページに留まることを確認
    await expect(page).toHaveURL('/login');
  });

  test('フォーム検証が正しく動作する', async ({ page }) => {
    await page.goto('/login');

    // 空のフォームで送信を試行
    await page.click('[data-testid="login-button"]');

    // バリデーションエラーが表示されることを確認
    await expect(page.locator('[data-testid="username-error"]')).toBeVisible();
    await expect(page.locator('[data-testid="password-error"]')).toBeVisible();

    // ユーザー名のみ入力
    await page.fill('[data-testid="username-input"]', 'admin');
    await page.click('[data-testid="login-button"]');

    // パスワードのバリデーションエラーのみ表示されることを確認
    await expect(page.locator('[data-testid="username-error"]')).not.toBeVisible();
    await expect(page.locator('[data-testid="password-error"]')).toBeVisible();
  });

  test('ログアウトが正常に動作する', async ({ page }) => {
    // ログインAPIレスポンスをモック
    await mockApiResponse(page, '/auth/login', {
      success: true,
      data: {
        token: 'mock-jwt-token',
        user: { id: 1, username: 'admin', role: 'admin' }
      }
    });

    // ログアウトAPIレスポンスをモック
    await mockApiResponse(page, '/auth/logout', {
      success: true,
      message: 'ログアウトしました'
    });

    // 管理者としてログイン
    await loginAsAdmin(page);

    // ログアウトボタンが表示されることを確認
    await expect(page.locator('[data-testid="logout-button"]')).toBeVisible();

    // ログアウト
    await logout(page);

    // ログインページにリダイレクトされることを確認
    await expect(page).toHaveURL('/login');
    await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
  });

  test('認証が必要なページへの未認証アクセスがリダイレクトされる', async ({ page }) => {
    // 未認証で管理ページにアクセス
    await page.goto('/admin');

    // ログインページにリダイレクトされることを確認
    await page.waitForURL('/login');
    await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
  });

  test('JWTトークンの期限切れ時に自動ログアウトされる', async ({ page }) => {
    // 初回ログインは成功
    await mockApiResponse(page, '/auth/login', {
      success: true,
      data: {
        token: 'expired-token',
        user: { id: 1, username: 'admin', role: 'admin' }
      }
    });

    // トークン検証で期限切れエラーをモック
    await mockApiError(page, '/auth/validate', 401, 'トークンが期限切れです');

    await loginAsAdmin(page);

    // 管理ページで何らかのアクションを実行（APIコールが発生）
    await page.reload();

    // 自動的にログインページにリダイレクトされることを確認
    await page.waitForURL('/login');
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('セッションが期限切れです');
  });

  test('ローディング状態が正しく表示される', async ({ page }) => {
    // 遅延のあるAPIレスポンスをモック
    await page.route('**/api/auth/login', async route => {
      // 2秒の遅延を追加
      await new Promise(resolve => setTimeout(resolve, 2000));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            token: 'mock-jwt-token',
            user: { id: 1, username: 'admin', role: 'admin' }
          }
        })
      });
    });

    await page.goto('/login');

    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'password');
    await page.click('[data-testid="login-button"]');

    // ローディング状態が表示されることを確認
    await expect(page.locator('[data-testid="loading-spinner"]')).toBeVisible();
    await expect(page.locator('[data-testid="login-button"]')).toBeDisabled();

    // ローディング完了後、リダイレクトされることを確認
    await page.waitForURL('/admin');
    await expect(page.locator('[data-testid="loading-spinner"]')).not.toBeVisible();
  });

  test('ネットワークエラー時の処理', async ({ page }) => {
    // ネットワークエラーをモック
    await page.route('**/api/auth/login', route => {
      route.abort('failed');
    });

    await page.goto('/login');

    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'password');
    await page.click('[data-testid="login-button"]');

    // ネットワークエラーメッセージが表示されることを確認
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('ネットワークエラー');

    // 再試行ボタンが表示されることを確認
    await expect(page.locator('[data-testid="retry-button"]')).toBeVisible();
  });
});
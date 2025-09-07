import { test, expect } from '@playwright/test';
import { 
  loginAsAdmin, 
  logout, 
  updateMatchResult, 
  expectNotification, 
  mockApiResponse, 
  mockApiError 
} from './utils/test-helpers.js';

test.describe('管理者ダッシュボード', () => {
  test.beforeEach(async ({ page }) => {
    // 認証APIをモック
    await mockApiResponse(page, '/auth/login', {
      success: true,
      data: {
        token: 'mock-jwt-token',
        user: { id: 1, username: 'admin', role: 'admin' }
      }
    });

    // 管理者用トーナメントデータをモック
    await mockApiResponse(page, '/admin/tournaments/volleyball', {
      success: true,
      data: {
        tournament: {
          id: 1,
          sport: 'volleyball',
          format: 'single_elimination',
          status: 'active'
        },
        pendingMatches: [
          {
            id: 1,
            team1: 'チームA',
            team2: 'チームB',
            round: '準々決勝',
            scheduled_at: '2024-01-15T10:00:00Z',
            status: 'pending'
          },
          {
            id: 2,
            team1: 'チームC',
            team2: 'チームD',
            round: '準々決勝',
            scheduled_at: '2024-01-15T11:00:00Z',
            status: 'pending'
          }
        ]
      }
    });

    await mockApiResponse(page, '/admin/tournaments/table_tennis', {
      success: true,
      data: {
        tournament: {
          id: 2,
          sport: 'table_tennis',
          format: 'sunny_weather',
          status: 'active'
        },
        pendingMatches: [
          {
            id: 3,
            team1: 'チームE',
            team2: 'チームF',
            round: '1回戦',
            scheduled_at: '2024-01-15T14:00:00Z',
            status: 'pending'
          }
        ]
      }
    });

    // 管理者としてログイン
    await loginAsAdmin(page);
  });

  test('管理者ダッシュボードが正しく表示される', async ({ page }) => {
    // 管理者ヘッダーが表示されることを確認
    await expect(page.locator('[data-testid="admin-header"]')).toBeVisible();
    await expect(page.locator('[data-testid="admin-title"]')).toContainText('管理者ダッシュボード');

    // ログアウトボタンが表示されることを確認
    await expect(page.locator('[data-testid="logout-button"]')).toBeVisible();

    // スポーツ選択セクションが表示されることを確認
    await expect(page.locator('[data-testid="sport-selector"]')).toBeVisible();
    await expect(page.locator('[data-testid="sport-option-volleyball"]')).toBeVisible();
    await expect(page.locator('[data-testid="sport-option-table_tennis"]')).toBeVisible();
    await expect(page.locator('[data-testid="sport-option-soccer"]')).toBeVisible();

    // 未完了試合一覧が表示されることを確認
    await expect(page.locator('[data-testid="pending-matches"]')).toBeVisible();
    await expect(page.locator('[data-testid="match-item-1"]')).toBeVisible();
    await expect(page.locator('[data-testid="match-item-2"]')).toBeVisible();
  });

  test('スポーツ選択が正常に動作する', async ({ page }) => {
    // 初期状態でバレーボールが選択されていることを確認
    await expect(page.locator('[data-testid="sport-option-volleyball"]')).toHaveClass(/selected|active/);
    await expect(page.locator('[data-testid="match-item-1"]')).toContainText('チームA vs チームB');

    // 卓球を選択
    await page.click('[data-testid="sport-option-table_tennis"]');

    // 卓球の試合が表示されることを確認
    await expect(page.locator('[data-testid="sport-option-table_tennis"]')).toHaveClass(/selected|active/);
    await expect(page.locator('[data-testid="match-item-3"]')).toContainText('チームE vs チームF');

    // バレーボールの試合が非表示になることを確認
    await expect(page.locator('[data-testid="match-item-1"]')).not.toBeVisible();
  });

  test('試合結果入力が正常に動作する', async ({ page }) => {
    // 試合結果更新APIをモック
    await mockApiResponse(page, '/admin/matches/1', {
      success: true,
      message: '試合結果を更新しました'
    });

    // 試合編集ボタンをクリック
    await page.click('[data-testid="edit-match-1"]');

    // 試合結果入力フォームが表示されることを確認
    await expect(page.locator('[data-testid="match-form"]')).toBeVisible();
    await expect(page.locator('[data-testid="match-form-title"]')).toContainText('チームA vs チームB');

    // スコア入力フィールドが表示されることを確認
    await expect(page.locator('[data-testid="score1-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="score2-input"]')).toBeVisible();
    await expect(page.locator('[data-testid="submit-result-button"]')).toBeVisible();
    await expect(page.locator('[data-testid="cancel-button"]')).toBeVisible();

    // スコアを入力
    await page.fill('[data-testid="score1-input"]', '3');
    await page.fill('[data-testid="score2-input"]', '1');

    // 結果を送信
    await page.click('[data-testid="submit-result-button"]');

    // 成功メッセージが表示されることを確認
    await expectNotification(page, '試合結果を更新しました', 'success');

    // フォームが閉じられることを確認
    await expect(page.locator('[data-testid="match-form"]')).not.toBeVisible();
  });

  test('試合結果入力のバリデーションが動作する', async ({ page }) => {
    await page.click('[data-testid="edit-match-1"]');

    // 空のフォームで送信を試行
    await page.click('[data-testid="submit-result-button"]');

    // バリデーションエラーが表示されることを確認
    await expect(page.locator('[data-testid="score1-error"]')).toBeVisible();
    await expect(page.locator('[data-testid="score2-error"]')).toBeVisible();

    // 無効な値を入力
    await page.fill('[data-testid="score1-input"]', '-1');
    await page.fill('[data-testid="score2-input"]', 'abc');
    await page.click('[data-testid="submit-result-button"]');

    // バリデーションエラーが表示されることを確認
    await expect(page.locator('[data-testid="score1-error"]')).toContainText('0以上の数値を入力してください');
    await expect(page.locator('[data-testid="score2-error"]')).toContainText('数値を入力してください');

    // 正しい値を入力
    await page.fill('[data-testid="score1-input"]', '3');
    await page.fill('[data-testid="score2-input"]', '1');

    // エラーが消えることを確認
    await expect(page.locator('[data-testid="score1-error"]')).not.toBeVisible();
    await expect(page.locator('[data-testid="score2-error"]')).not.toBeVisible();
  });

  test('試合結果入力のキャンセルが動作する', async ({ page }) => {
    await page.click('[data-testid="edit-match-1"]');

    // スコアを入力
    await page.fill('[data-testid="score1-input"]', '2');
    await page.fill('[data-testid="score2-input"]', '3');

    // キャンセルボタンをクリック
    await page.click('[data-testid="cancel-button"]');

    // フォームが閉じられることを確認
    await expect(page.locator('[data-testid="match-form"]')).not.toBeVisible();

    // 再度フォームを開いて、値がリセットされていることを確認
    await page.click('[data-testid="edit-match-1"]');
    await expect(page.locator('[data-testid="score1-input"]')).toHaveValue('');
    await expect(page.locator('[data-testid="score2-input"]')).toHaveValue('');
  });

  test('トーナメント形式切り替えが動作する（卓球）', async ({ page }) => {
    // 形式切り替えAPIをモック
    await mockApiResponse(page, '/admin/tournaments/2/format', {
      success: true,
      message: 'トーナメント形式を変更しました'
    });

    // 卓球を選択
    await page.click('[data-testid="sport-option-table_tennis"]');

    // 形式切り替えセクションが表示されることを確認
    await expect(page.locator('[data-testid="format-toggle"]')).toBeVisible();
    await expect(page.locator('[data-testid="format-sunny"]')).toBeVisible();
    await expect(page.locator('[data-testid="format-rainy"]')).toBeVisible();

    // 現在の形式が表示されることを確認
    await expect(page.locator('[data-testid="current-format"]')).toContainText('晴天時形式');

    // 雨天時形式に切り替え
    await page.click('[data-testid="format-rainy"]');

    // 確認ダイアログが表示されることを確認
    await expect(page.locator('[data-testid="format-change-dialog"]')).toBeVisible();
    await expect(page.locator('[data-testid="dialog-message"]')).toContainText('形式を変更しますか？');

    // 確認ボタンをクリック
    await page.click('[data-testid="confirm-format-change"]');

    // 成功メッセージが表示されることを確認
    await expectNotification(page, 'トーナメント形式を変更しました', 'success');

    // 現在の形式が更新されることを確認
    await expect(page.locator('[data-testid="current-format"]')).toContainText('雨天時形式');
  });

  test('形式切り替えのキャンセルが動作する', async ({ page }) => {
    await page.click('[data-testid="sport-option-table_tennis"]');
    await page.click('[data-testid="format-rainy"]');

    // 確認ダイアログでキャンセル
    await page.click('[data-testid="cancel-format-change"]');

    // ダイアログが閉じられることを確認
    await expect(page.locator('[data-testid="format-change-dialog"]')).not.toBeVisible();

    // 形式が変更されていないことを確認
    await expect(page.locator('[data-testid="current-format"]')).toContainText('晴天時形式');
  });

  test('エラーハンドリングが正常に動作する', async ({ page }) => {
    // 試合結果更新でエラーをモック
    await mockApiError(page, '/admin/matches/1', 500, '試合結果の更新に失敗しました');

    await page.click('[data-testid="edit-match-1"]');
    await page.fill('[data-testid="score1-input"]', '3');
    await page.fill('[data-testid="score2-input"]', '1');
    await page.click('[data-testid="submit-result-button"]');

    // エラーメッセージが表示されることを確認
    await expectNotification(page, '試合結果の更新に失敗しました', 'error');

    // フォームが開いたままであることを確認
    await expect(page.locator('[data-testid="match-form"]')).toBeVisible();
  });

  test('ローディング状態が正しく表示される', async ({ page }) => {
    // 遅延のあるAPIレスポンスをモック
    await page.route('**/api/admin/matches/1', async route => {
      await new Promise(resolve => setTimeout(resolve, 2000));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          message: '試合結果を更新しました'
        })
      });
    });

    await page.click('[data-testid="edit-match-1"]');
    await page.fill('[data-testid="score1-input"]', '3');
    await page.fill('[data-testid="score2-input"]', '1');
    await page.click('[data-testid="submit-result-button"]');

    // ローディング状態が表示されることを確認
    await expect(page.locator('[data-testid="loading-spinner"]')).toBeVisible();
    await expect(page.locator('[data-testid="submit-result-button"]')).toBeDisabled();

    // ローディング完了後、成功メッセージが表示されることを確認
    await expectNotification(page, '試合結果を更新しました', 'success');
    await expect(page.locator('[data-testid="loading-spinner"]')).not.toBeVisible();
  });

  test('未認証ユーザーがアクセスできない', async ({ page }) => {
    // ログアウト
    await logout(page);

    // 管理ページに直接アクセスを試行
    await page.goto('/admin');

    // ログインページにリダイレクトされることを確認
    await page.waitForURL('/login');
    await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
  });

  test('権限のないユーザーがアクセスできない', async ({ page }) => {
    // 一般ユーザーの認証レスポンスをモック
    await mockApiResponse(page, '/auth/login', {
      success: true,
      data: {
        token: 'user-jwt-token',
        user: { id: 2, username: 'user', role: 'user' }
      }
    });

    // 管理ページアクセス時に権限エラーをモック
    await mockApiError(page, '/admin/tournaments/volleyball', 403, 'アクセス権限がありません');

    await logout(page);
    await page.goto('/login');
    await page.fill('[data-testid="username-input"]', 'user');
    await page.fill('[data-testid="password-input"]', 'password');
    await page.click('[data-testid="login-button"]');

    // 管理ページにアクセスを試行
    await page.goto('/admin');

    // 権限エラーが表示されることを確認
    await expectNotification(page, 'アクセス権限がありません', 'error');
  });

  test('リアルタイム更新が動作する', async ({ page }) => {
    // 初期状態を確認
    await expect(page.locator('[data-testid="match-item-1"]')).toContainText('未実施');

    // 更新されたデータをモック
    await mockApiResponse(page, '/admin/tournaments/volleyball', {
      success: true,
      data: {
        tournament: {
          id: 1,
          sport: 'volleyball',
          format: 'single_elimination',
          status: 'active'
        },
        pendingMatches: [
          {
            id: 2,
            team1: 'チームC',
            team2: 'チームD',
            round: '準々決勝',
            scheduled_at: '2024-01-15T11:00:00Z',
            status: 'pending'
          }
        ]
      }
    });

    // リフレッシュボタンをクリック
    await page.click('[data-testid="refresh-button"]');

    // 完了した試合が一覧から消えることを確認
    await expect(page.locator('[data-testid="match-item-1"]')).not.toBeVisible();
    await expect(page.locator('[data-testid="match-item-2"]')).toBeVisible();
  });

  test('キーボードナビゲーションが動作する', async ({ page }) => {
    // Tabキーでナビゲーション
    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="sport-option-volleyball"]')).toBeFocused();

    // 矢印キーでスポーツ選択
    await page.keyboard.press('ArrowRight');
    await expect(page.locator('[data-testid="sport-option-table_tennis"]')).toBeFocused();

    // Enterキーで選択
    await page.keyboard.press('Enter');
    await expect(page.locator('[data-testid="sport-option-table_tennis"]')).toHaveClass(/selected|active/);

    // Tabキーで試合一覧に移動
    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="match-item-3"]')).toBeFocused();

    // Enterキーで試合編集
    await page.keyboard.press('Enter');
    await expect(page.locator('[data-testid="match-form"]')).toBeVisible();
  });
});
import { test, expect } from '@playwright/test';
import { 
  loginAsAdmin, 
  logout, 
  selectSportTab, 
  updateMatchResult, 
  expectNotification, 
  mockApiResponse 
} from './utils/test-helpers.js';

test.describe('ユーザーフロー統合テスト', () => {
  test.beforeEach(async ({ page }) => {
    // 基本的なAPIレスポンスをモック
    await mockApiResponse(page, '/auth/login', {
      success: true,
      data: {
        token: 'mock-jwt-token',
        user: { id: 1, username: 'admin', role: 'admin' }
      }
    });

    await mockApiResponse(page, '/tournaments/volleyball', {
      success: true,
      data: {
        tournament: {
          id: 1,
          sport: 'volleyball',
          format: 'single_elimination',
          status: 'active'
        },
        bracket: {
          rounds: [
            {
              name: '準々決勝',
              matches: [
                {
                  id: 1,
                  team1: 'チームA',
                  team2: 'チームB',
                  score1: null,
                  score2: null,
                  winner: null,
                  status: 'pending'
                }
              ]
            }
          ]
        }
      }
    });

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
          }
        ]
      }
    });
  });

  test('完全な管理者ワークフロー: ログイン → 試合結果入力 → ログアウト', async ({ page }) => {
    // 試合結果更新APIをモック
    await mockApiResponse(page, '/admin/matches/1', {
      success: true,
      message: '試合結果を更新しました'
    });

    // 更新後のトーナメントデータをモック
    await mockApiResponse(page, '/tournaments/volleyball', {
      success: true,
      data: {
        tournament: {
          id: 1,
          sport: 'volleyball',
          format: 'single_elimination',
          status: 'active'
        },
        bracket: {
          rounds: [
            {
              name: '準々決勝',
              matches: [
                {
                  id: 1,
                  team1: 'チームA',
                  team2: 'チームB',
                  score1: 3,
                  score2: 1,
                  winner: 'チームA',
                  status: 'completed'
                }
              ]
            }
          ]
        }
      }
    });

    // ステップ1: ホームページから開始
    await page.goto('/');
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();
    await expect(page.locator('[data-testid="match-card-1"]')).toContainText('未実施');

    // ステップ2: 管理者ログイン
    await page.click('[data-testid="admin-login-link"]');
    await expect(page).toHaveURL('/login');
    
    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'password');
    await page.click('[data-testid="login-button"]');
    
    await page.waitForURL('/admin');
    await expect(page.locator('[data-testid="admin-dashboard"]')).toBeVisible();

    // ステップ3: 試合結果を入力
    await expect(page.locator('[data-testid="match-item-1"]')).toContainText('チームA vs チームB');
    await page.click('[data-testid="edit-match-1"]');
    
    await expect(page.locator('[data-testid="match-form"]')).toBeVisible();
    await page.fill('[data-testid="score1-input"]', '3');
    await page.fill('[data-testid="score2-input"]', '1');
    await page.click('[data-testid="submit-result-button"]');
    
    await expectNotification(page, '試合結果を更新しました', 'success');

    // ステップ4: 結果をホームページで確認
    await page.click('[data-testid="view-tournament-link"]');
    await page.waitForURL('/');
    
    await expect(page.locator('[data-testid="match-card-1"]')).toContainText('チームA');
    await expect(page.locator('[data-testid="match-card-1"]')).toContainText('3 - 1');

    // ステップ5: 管理ページに戻ってログアウト
    await page.click('[data-testid="admin-link"]');
    await page.waitForURL('/admin');
    
    await page.click('[data-testid="logout-button"]');
    await page.waitForURL('/login');
    
    await expect(page.locator('[data-testid="login-form"]')).toBeVisible();
  });

  test('一般ユーザーのトーナメント閲覧フロー', async ({ page }) => {
    // 複数スポーツのデータをモック
    await mockApiResponse(page, '/tournaments/table_tennis', {
      success: true,
      data: {
        tournament: {
          id: 2,
          sport: 'table_tennis',
          format: 'sunny_weather',
          status: 'active'
        },
        bracket: {
          rounds: [
            {
              name: '1回戦',
              matches: [
                {
                  id: 2,
                  team1: 'チームC',
                  team2: 'チームD',
                  score1: 2,
                  score2: 0,
                  winner: 'チームC',
                  status: 'completed'
                }
              ]
            }
          ]
        }
      }
    });

    await mockApiResponse(page, '/tournaments/soccer', {
      success: true,
      data: {
        tournament: {
          id: 3,
          sport: 'soccer',
          format: 'single_elimination',
          status: 'active'
        },
        bracket: {
          rounds: [
            {
              name: '1回戦',
              matches: [
                {
                  id: 3,
                  team1: 'チームE',
                  team2: 'チームF',
                  score1: null,
                  score2: null,
                  winner: null,
                  status: 'pending'
                }
              ]
            }
          ]
        }
      }
    });

    // ステップ1: ホームページにアクセス
    await page.goto('/');
    await expect(page.locator('[data-testid="sport-tabs"]')).toBeVisible();

    // ステップ2: バレーボールトーナメントを確認
    await expect(page.locator('[data-testid="sport-tab-volleyball"]')).toHaveClass(/active|selected/);
    await expect(page.locator('[data-testid="match-card-1"]')).toBeVisible();

    // ステップ3: 卓球タブに切り替え
    await selectSportTab(page, 'table_tennis');
    await expect(page.locator('[data-testid="match-card-2"]')).toBeVisible();
    await expect(page.locator('[data-testid="match-card-2"]')).toContainText('チームC');
    await expect(page.locator('[data-testid="match-card-2"]')).toContainText('2 - 0');

    // ステップ4: サッカータブに切り替え
    await selectSportTab(page, 'soccer');
    await expect(page.locator('[data-testid="match-card-3"]')).toBeVisible();
    await expect(page.locator('[data-testid="match-card-3"]')).toContainText('未実施');

    // ステップ5: バレーボールタブに戻る
    await selectSportTab(page, 'volleyball');
    await expect(page.locator('[data-testid="match-card-1"]')).toBeVisible();
  });

  test('リアルタイム更新フロー: 管理者の更新が一般ユーザーに反映される', async ({ page, context }) => {
    // 2つのページを作成（管理者用と一般ユーザー用）
    const adminPage = page;
    const userPage = await context.newPage();

    // 試合結果更新APIをモック
    await mockApiResponse(adminPage, '/admin/matches/1', {
      success: true,
      message: '試合結果を更新しました'
    });

    // 更新後のデータをモック
    const updatedTournamentData = {
      success: true,
      data: {
        tournament: {
          id: 1,
          sport: 'volleyball',
          format: 'single_elimination',
          status: 'active'
        },
        bracket: {
          rounds: [
            {
              name: '準々決勝',
              matches: [
                {
                  id: 1,
                  team1: 'チームA',
                  team2: 'チームB',
                  score1: 3,
                  score2: 1,
                  winner: 'チームA',
                  status: 'completed'
                }
              ]
            }
          ]
        }
      }
    };

    // ステップ1: 一般ユーザーがトーナメントを閲覧
    await userPage.goto('/');
    await expect(userPage.locator('[data-testid="match-card-1"]')).toContainText('未実施');

    // ステップ2: 管理者がログインして試合結果を入力
    await loginAsAdmin(adminPage);
    await adminPage.click('[data-testid="edit-match-1"]');
    await adminPage.fill('[data-testid="score1-input"]', '3');
    await adminPage.fill('[data-testid="score2-input"]', '1');

    // 両方のページで更新後のデータをモック
    await mockApiResponse(adminPage, '/tournaments/volleyball', updatedTournamentData);
    await mockApiResponse(userPage, '/tournaments/volleyball', updatedTournamentData);

    await adminPage.click('[data-testid="submit-result-button"]');
    await expectNotification(adminPage, '試合結果を更新しました', 'success');

    // ステップ3: 一般ユーザーページで更新を確認（ポーリングまたは手動リフレッシュ）
    await userPage.click('[data-testid="refresh-button"]');
    await expect(userPage.locator('[data-testid="match-card-1"]')).toContainText('3 - 1');
    await expect(userPage.locator('[data-testid="match-card-1"]')).toContainText('チームA');
  });

  test('エラー回復フロー: ネットワークエラーからの復旧', async ({ page }) => {
    // ステップ1: 正常にページを読み込み
    await page.goto('/');
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();

    // ステップ2: ネットワークエラーをシミュレート
    await page.route('**/api/tournaments/table_tennis', route => {
      route.abort('failed');
    });

    await page.click('[data-testid="sport-tab-table_tennis"]');
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('ネットワークエラー');

    // ステップ3: ネットワーク復旧をシミュレート
    await mockApiResponse(page, '/tournaments/table_tennis', {
      success: true,
      data: {
        tournament: {
          id: 2,
          sport: 'table_tennis',
          format: 'sunny_weather',
          status: 'active'
        },
        bracket: {
          rounds: [
            {
              name: '1回戦',
              matches: [
                {
                  id: 2,
                  team1: 'チームC',
                  team2: 'チームD',
                  score1: null,
                  score2: null,
                  winner: null,
                  status: 'pending'
                }
              ]
            }
          ]
        }
      }
    });

    // ステップ4: 再試行ボタンをクリック
    await page.click('[data-testid="retry-button"]');
    await expect(page.locator('[data-testid="error-message"]')).not.toBeVisible();
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();
  });

  test('モバイルデバイスでの完全なユーザーフロー', async ({ page }) => {
    // モバイルビューポートに設定
    await page.setViewportSize({ width: 375, height: 667 });

    // ステップ1: モバイルでホームページにアクセス
    await page.goto('/');
    await expect(page.locator('[data-testid="mobile-nav"]')).toBeVisible();

    // ステップ2: ハンバーガーメニューを開く
    await page.click('[data-testid="mobile-menu-button"]');
    await expect(page.locator('[data-testid="mobile-menu"]')).toBeVisible();

    // ステップ3: スポーツタブをスワイプで切り替え
    const tabContainer = page.locator('[data-testid="sport-tabs"]');
    await tabContainer.hover();
    await page.mouse.down();
    await page.mouse.move(100, 0);
    await page.mouse.up();

    // ステップ4: モバイル用ブラケット表示を確認
    await expect(page.locator('[data-testid="tournament-bracket"]')).toHaveClass(/mobile-bracket/);

    // ステップ5: 管理者ログイン（モバイル）
    await page.click('[data-testid="admin-login-link"]');
    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'password');
    await page.click('[data-testid="login-button"]');

    // ステップ6: モバイル管理ダッシュボードを確認
    await expect(page.locator('[data-testid="mobile-admin-dashboard"]')).toBeVisible();
  });

  test('アクセシビリティ対応フロー: キーボードナビゲーション', async ({ page }) => {
    await page.goto('/');

    // ステップ1: キーボードでスポーツタブをナビゲート
    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="sport-tab-volleyball"]')).toBeFocused();

    await page.keyboard.press('ArrowRight');
    await expect(page.locator('[data-testid="sport-tab-table_tennis"]')).toBeFocused();

    await page.keyboard.press('Enter');
    await expect(page.locator('[data-testid="sport-tab-table_tennis"]')).toHaveClass(/active|selected/);

    // ステップ2: キーボードでブラケット内をナビゲート
    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeFocused();

    // ステップ3: キーボードで管理者ログイン
    await page.keyboard.press('Tab');
    await page.keyboard.press('Tab'); // 管理者ログインリンクまで移動
    await page.keyboard.press('Enter');

    await expect(page).toHaveURL('/login');
    
    // フォームをキーボードで操作
    await page.keyboard.press('Tab');
    await page.keyboard.type('admin');
    await page.keyboard.press('Tab');
    await page.keyboard.type('password');
    await page.keyboard.press('Tab');
    await page.keyboard.press('Enter');

    await page.waitForURL('/admin');
    await expect(page.locator('[data-testid="admin-dashboard"]')).toBeVisible();
  });

  test('セッション管理フロー: トークン期限切れとリフレッシュ', async ({ page }) => {
    // ステップ1: 正常にログイン
    await loginAsAdmin(page);
    await expect(page.locator('[data-testid="admin-dashboard"]')).toBeVisible();

    // ステップ2: トークン期限切れをシミュレート
    await page.route('**/api/admin/tournaments/volleyball', route => {
      route.fulfill({
        status: 401,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'TOKEN_EXPIRED',
          message: 'トークンが期限切れです'
        })
      });
    });

    // ステップ3: API呼び出しでトークン期限切れエラーが発生
    await page.click('[data-testid="refresh-button"]');
    
    // ステップ4: 自動的にログインページにリダイレクト
    await page.waitForURL('/login');
    await expect(page.locator('[data-testid="error-message"]')).toContainText('セッションが期限切れです');

    // ステップ5: 再ログイン
    await page.fill('[data-testid="username-input"]', 'admin');
    await page.fill('[data-testid="password-input"]', 'password');
    await page.click('[data-testid="login-button"]');

    await page.waitForURL('/admin');
    await expect(page.locator('[data-testid="admin-dashboard"]')).toBeVisible();
  });
});
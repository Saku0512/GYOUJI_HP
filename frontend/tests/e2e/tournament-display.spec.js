import { test, expect } from '@playwright/test';
import { selectSportTab, expectTournamentBracket, setViewport, mockApiResponse } from './utils/test-helpers.js';

test.describe('トーナメント表示', () => {
  test.beforeEach(async ({ page }) => {
    // トーナメントデータをモック
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
                },
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
            },
            {
              name: '準決勝',
              matches: [
                {
                  id: 3,
                  team1: 'チームA',
                  team2: 'TBD',
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
                  id: 4,
                  team1: 'チームE',
                  team2: 'チームF',
                  score1: 2,
                  score2: 0,
                  winner: 'チームE',
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
                  id: 5,
                  team1: 'チームG',
                  team2: 'チームH',
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

    await page.goto('/');
  });

  test('ホームページが正しく表示される', async ({ page }) => {
    // ページタイトルを確認
    await expect(page).toHaveTitle(/トーナメント管理システム/);

    // メインコンテンツが表示されることを確認
    await expect(page.locator('[data-testid="main-content"]')).toBeVisible();

    // スポーツタブが表示されることを確認
    await expect(page.locator('[data-testid="sport-tabs"]')).toBeVisible();
    await expect(page.locator('[data-testid="sport-tab-volleyball"]')).toBeVisible();
    await expect(page.locator('[data-testid="sport-tab-table_tennis"]')).toBeVisible();
    await expect(page.locator('[data-testid="sport-tab-soccer"]')).toBeVisible();

    // デフォルトでバレーボールタブが選択されていることを確認
    await expect(page.locator('[data-testid="sport-tab-volleyball"]')).toHaveClass(/active|selected/);
  });

  test('スポーツタブの切り替えが正常に動作する', async ({ page }) => {
    // 初期状態でバレーボールブラケットが表示されることを確認
    await expectTournamentBracket(page, 'volleyball');

    // 卓球タブに切り替え
    await selectSportTab(page, 'table_tennis');
    await expectTournamentBracket(page, 'table_tennis');

    // サッカータブに切り替え
    await selectSportTab(page, 'soccer');
    await expectTournamentBracket(page, 'soccer');

    // バレーボールタブに戻る
    await selectSportTab(page, 'volleyball');
    await expectTournamentBracket(page, 'volleyball');
  });

  test('トーナメントブラケットが正しく表示される', async ({ page }) => {
    // ブラケットコンテナが表示されることを確認
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();

    // 試合カードが表示されることを確認
    await expect(page.locator('[data-testid="match-card-1"]')).toBeVisible();
    await expect(page.locator('[data-testid="match-card-2"]')).toBeVisible();

    // 完了した試合のスコアが表示されることを確認
    const completedMatch = page.locator('[data-testid="match-card-1"]');
    await expect(completedMatch.locator('[data-testid="team1-name"]')).toContainText('チームA');
    await expect(completedMatch.locator('[data-testid="team2-name"]')).toContainText('チームB');
    await expect(completedMatch.locator('[data-testid="score1"]')).toContainText('3');
    await expect(completedMatch.locator('[data-testid="score2"]')).toContainText('1');
    await expect(completedMatch.locator('[data-testid="winner"]')).toContainText('チームA');

    // 未実施の試合が正しく表示されることを確認
    const pendingMatch = page.locator('[data-testid="match-card-2"]');
    await expect(pendingMatch.locator('[data-testid="team1-name"]')).toContainText('チームC');
    await expect(pendingMatch.locator('[data-testid="team2-name"]')).toContainText('チームD');
    await expect(pendingMatch.locator('[data-testid="status"]')).toContainText('未実施');
  });

  test('リアルタイム更新が動作する', async ({ page }) => {
    // 初期状態を確認
    await expect(page.locator('[data-testid="match-card-2"]')).toContainText('未実施');

    // 更新されたデータをモック
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
                },
                {
                  id: 2,
                  team1: 'チームC',
                  team2: 'チームD',
                  score1: 2,
                  score2: 3,
                  winner: 'チームD',
                  status: 'completed'
                }
              ]
            }
          ]
        }
      }
    });

    // 手動でリフレッシュボタンをクリック（ポーリングのシミュレーション）
    await page.click('[data-testid="refresh-button"]');

    // 更新されたデータが表示されることを確認
    const updatedMatch = page.locator('[data-testid="match-card-2"]');
    await expect(updatedMatch.locator('[data-testid="score1"]')).toContainText('2');
    await expect(updatedMatch.locator('[data-testid="score2"]')).toContainText('3');
    await expect(updatedMatch.locator('[data-testid="winner"]')).toContainText('チームD');
  });

  test('エラー状態が正しく処理される', async ({ page }) => {
    // エラーレスポンスをモック
    await page.route('**/api/tournaments/volleyball', route => {
      route.fulfill({
        status: 500,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'SERVER_ERROR',
          message: 'サーバーエラーが発生しました'
        })
      });
    });

    await page.reload();

    // エラーメッセージが表示されることを確認
    await expect(page.locator('[data-testid="error-message"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-message"]')).toContainText('サーバーエラーが発生しました');

    // 再試行ボタンが表示されることを確認
    await expect(page.locator('[data-testid="retry-button"]')).toBeVisible();
  });

  test('ローディング状態が正しく表示される', async ({ page }) => {
    // 遅延のあるAPIレスポンスをモック
    await page.route('**/api/tournaments/table_tennis', async route => {
      await new Promise(resolve => setTimeout(resolve, 1000));
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          success: true,
          data: {
            tournament: { id: 2, sport: 'table_tennis', format: 'sunny_weather', status: 'active' },
            bracket: { rounds: [] }
          }
        })
      });
    });

    // 卓球タブをクリック
    await page.click('[data-testid="sport-tab-table_tennis"]');

    // ローディングスピナーが表示されることを確認
    await expect(page.locator('[data-testid="loading-spinner"]')).toBeVisible();

    // ローディング完了後、ブラケットが表示されることを確認
    await expect(page.locator('[data-testid="loading-spinner"]')).not.toBeVisible();
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();
  });

  test('モバイルデバイスでレスポンシブ表示される', async ({ page }) => {
    // モバイルビューポートに設定
    await setViewport(page, 'mobile');

    // モバイル用のナビゲーションが表示されることを確認
    await expect(page.locator('[data-testid="mobile-nav"]')).toBeVisible();

    // スポーツタブがモバイル用レイアウトで表示されることを確認
    await expect(page.locator('[data-testid="sport-tabs"]')).toHaveClass(/mobile-layout/);

    // ブラケットがモバイル用レイアウトで表示されることを確認
    await expect(page.locator('[data-testid="tournament-bracket"]')).toHaveClass(/mobile-bracket/);

    // スワイプ可能なタブ操作をテスト
    const tabContainer = page.locator('[data-testid="sport-tabs"]');
    await tabContainer.hover();
    await page.mouse.down();
    await page.mouse.move(100, 0);
    await page.mouse.up();

    // タブが切り替わることを確認
    await expect(page.locator('[data-testid="sport-tab-table_tennis"]')).toHaveClass(/active|selected/);
  });

  test('タブレットデバイスでレスポンシブ表示される', async ({ page }) => {
    // タブレットビューポートに設定
    await setViewport(page, 'tablet');

    // タブレット用のレイアウトが適用されることを確認
    await expect(page.locator('[data-testid="main-content"]')).toHaveClass(/tablet-layout/);

    // ブラケットがタブレット用レイアウトで表示されることを確認
    await expect(page.locator('[data-testid="tournament-bracket"]')).toHaveClass(/tablet-bracket/);

    // タッチフレンドリーなインターフェースが表示されることを確認
    const matchCards = page.locator('[data-testid^="match-card-"]');
    await expect(matchCards.first()).toHaveClass(/touch-friendly/);
  });

  test('キーボードナビゲーションが動作する', async ({ page }) => {
    // フォーカスがスポーツタブに移動することを確認
    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="sport-tab-volleyball"]')).toBeFocused();

    // 矢印キーでタブを切り替え
    await page.keyboard.press('ArrowRight');
    await expect(page.locator('[data-testid="sport-tab-table_tennis"]')).toBeFocused();

    // Enterキーでタブを選択
    await page.keyboard.press('Enter');
    await expect(page.locator('[data-testid="sport-tab-table_tennis"]')).toHaveClass(/active|selected/);

    // Tabキーでブラケット内の要素に移動
    await page.keyboard.press('Tab');
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeFocused();
  });

  test('アクセシビリティ要件が満たされている', async ({ page }) => {
    // ARIA属性が正しく設定されていることを確認
    await expect(page.locator('[data-testid="sport-tabs"]')).toHaveAttribute('role', 'tablist');
    await expect(page.locator('[data-testid="sport-tab-volleyball"]')).toHaveAttribute('role', 'tab');
    await expect(page.locator('[data-testid="tournament-bracket"]')).toHaveAttribute('role', 'tabpanel');

    // スクリーンリーダー用のテキストが提供されていることを確認
    await expect(page.locator('[data-testid="sr-only-tournament-info"]')).toBeInViewport();

    // 色だけでなくテキストでも状態が示されていることを確認
    const completedMatch = page.locator('[data-testid="match-card-1"]');
    await expect(completedMatch.locator('[data-testid="status-text"]')).toContainText('完了');

    const pendingMatch = page.locator('[data-testid="match-card-2"]');
    await expect(pendingMatch.locator('[data-testid="status-text"]')).toContainText('未実施');
  });
});
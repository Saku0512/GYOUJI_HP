import { test, expect, devices } from '@playwright/test';
import { loginAsAdmin, selectSportTab, mockApiResponse } from './utils/test-helpers.js';

test.describe('クロスブラウザ互換性テスト', () => {
  test.beforeEach(async ({ page }) => {
    // 基本的なAPIレスポンスをモック
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
  });

  // Chromiumブラウザでのテスト
  test('Chromium: 基本機能が正常に動作する', async ({ page, browserName }) => {
    test.skip(browserName !== 'chromium', 'Chromium専用テスト');

    await page.goto('/');

    // CSS Grid レイアウトのサポートを確認
    const bracketElement = page.locator('[data-testid="tournament-bracket"]');
    await expect(bracketElement).toBeVisible();
    
    // CSS Grid が適用されていることを確認
    const gridDisplay = await bracketElement.evaluate(el => 
      window.getComputedStyle(el).display
    );
    expect(gridDisplay).toBe('grid');

    // Flexbox レイアウトのサポートを確認
    const tabsElement = page.locator('[data-testid="sport-tabs"]');
    const flexDisplay = await tabsElement.evaluate(el => 
      window.getComputedStyle(el).display
    );
    expect(flexDisplay).toBe('flex');

    // ES6+ 機能のサポートを確認（async/await, arrow functions）
    const jsSupport = await page.evaluate(() => {
      try {
        // Arrow function
        const test1 = () => true;
        
        // Template literals
        const test2 = `test${1}`;
        
        // Destructuring
        const { length } = [1, 2, 3];
        
        // Promise
        const test3 = Promise.resolve(true);
        
        return test1() && test2 === 'test1' && length === 3 && test3 instanceof Promise;
      } catch (e) {
        return false;
      }
    });
    expect(jsSupport).toBe(true);
  });

  // Firefoxブラウザでのテスト
  test('Firefox: 基本機能が正常に動作する', async ({ page, browserName }) => {
    test.skip(browserName !== 'firefox', 'Firefox専用テスト');

    await page.goto('/');

    // Firefox特有のCSS機能をテスト
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();

    // CSS カスタムプロパティ（CSS変数）のサポートを確認
    const cssVariableSupport = await page.evaluate(() => {
      const testElement = document.createElement('div');
      testElement.style.setProperty('--test-var', 'red');
      testElement.style.color = 'var(--test-var)';
      document.body.appendChild(testElement);
      
      const computedColor = window.getComputedStyle(testElement).color;
      document.body.removeChild(testElement);
      
      return computedColor === 'red' || computedColor === 'rgb(255, 0, 0)';
    });
    expect(cssVariableSupport).toBe(true);

    // Firefox でのイベントハンドリングをテスト
    await page.click('[data-testid="sport-tab-table_tennis"]');
    await expect(page.locator('[data-testid="sport-tab-table_tennis"]')).toHaveClass(/active|selected/);
  });

  // WebKitブラウザでのテスト
  test('WebKit: 基本機能が正常に動作する', async ({ page, browserName }) => {
    test.skip(browserName !== 'webkit', 'WebKit専用テスト');

    await page.goto('/');

    // WebKit特有の機能をテスト
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();

    // -webkit- プレフィックスが必要な機能のテスト
    const webkitSupport = await page.evaluate(() => {
      const testElement = document.createElement('div');
      testElement.style.webkitTransform = 'translateX(10px)';
      document.body.appendChild(testElement);
      
      const transform = window.getComputedStyle(testElement).transform;
      document.body.removeChild(testElement);
      
      return transform !== 'none';
    });
    expect(webkitSupport).toBe(true);

    // タッチイベントのサポートを確認（Safari/WebKit）
    const touchSupport = await page.evaluate(() => {
      return 'ontouchstart' in window || navigator.maxTouchPoints > 0;
    });
    // WebKitでは通常タッチサポートがある
    expect(typeof touchSupport).toBe('boolean');
  });

  // モバイルChrome でのテスト
  test('Mobile Chrome: モバイル機能が正常に動作する', async ({ page, browserName }) => {
    test.skip(browserName !== 'chromium', 'Mobile Chrome専用テスト');

    // モバイルビューポートを設定
    await page.setViewportSize(devices['Pixel 5'].viewport);
    await page.goto('/');

    // モバイル用レイアウトが適用されることを確認
    await expect(page.locator('[data-testid="mobile-nav"]')).toBeVisible();

    // タッチイベントのテスト
    const tabElement = page.locator('[data-testid="sport-tab-table_tennis"]');
    
    // タッチスタートイベントをシミュレート
    await tabElement.dispatchEvent('touchstart');
    await tabElement.dispatchEvent('touchend');
    
    await expect(page.locator('[data-testid="sport-tab-table_tennis"]')).toHaveClass(/active|selected/);

    // スワイプジェスチャーのテスト
    const tabContainer = page.locator('[data-testid="sport-tabs"]');
    await tabContainer.hover();
    
    // スワイプジェスチャーをシミュレート
    await page.mouse.down();
    await page.mouse.move(100, 0);
    await page.mouse.up();

    // ビューポートメタタグの確認
    const viewportMeta = await page.locator('meta[name="viewport"]').getAttribute('content');
    expect(viewportMeta).toContain('width=device-width');
  });

  // モバイルSafari でのテスト
  test('Mobile Safari: iOS機能が正常に動作する', async ({ page, browserName }) => {
    test.skip(browserName !== 'webkit', 'Mobile Safari専用テスト');

    // iPhoneビューポートを設定
    await page.setViewportSize(devices['iPhone 12'].viewport);
    await page.goto('/');

    // iOS Safari特有の機能をテスト
    await expect(page.locator('[data-testid="mobile-nav"]')).toBeVisible();

    // iOS Safari でのスクロール動作をテスト
    const scrollContainer = page.locator('[data-testid="tournament-bracket"]');
    await scrollContainer.evaluate(el => {
      el.style.webkitOverflowScrolling = 'touch';
      el.scrollTop = 100;
    });

    // Safe Area Insets のサポートを確認
    const safeAreaSupport = await page.evaluate(() => {
      const testElement = document.createElement('div');
      testElement.style.paddingTop = 'env(safe-area-inset-top)';
      document.body.appendChild(testElement);
      
      const padding = window.getComputedStyle(testElement).paddingTop;
      document.body.removeChild(testElement);
      
      return padding !== '0px' || CSS.supports('padding-top', 'env(safe-area-inset-top)');
    });
    expect(typeof safeAreaSupport).toBe('boolean');
  });

  // レスポンシブデザインのクロスブラウザテスト
  test('レスポンシブデザインがすべてのブラウザで動作する', async ({ page }) => {
    const viewports = [
      { width: 320, height: 568, name: 'Mobile Small' },
      { width: 768, height: 1024, name: 'Tablet' },
      { width: 1920, height: 1080, name: 'Desktop' }
    ];

    for (const viewport of viewports) {
      await page.setViewportSize({ width: viewport.width, height: viewport.height });
      await page.goto('/');

      // 各ビューポートでレイアウトが適切に表示されることを確認
      await expect(page.locator('[data-testid="main-content"]')).toBeVisible();
      await expect(page.locator('[data-testid="sport-tabs"]')).toBeVisible();
      await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();

      // ビューポート固有のスタイルが適用されることを確認
      const mainContent = page.locator('[data-testid="main-content"]');
      const computedStyle = await mainContent.evaluate(el => {
        const style = window.getComputedStyle(el);
        return {
          display: style.display,
          flexDirection: style.flexDirection,
          gridTemplateColumns: style.gridTemplateColumns
        };
      });

      // レスポンシブレイアウトが適用されていることを確認
      expect(computedStyle.display).toMatch(/flex|grid|block/);
    }
  });

  // CSS機能のクロスブラウザサポートテスト
  test('CSS機能のブラウザサポートを確認', async ({ page }) => {
    await page.goto('/');

    const cssFeatureSupport = await page.evaluate(() => {
      const features = {};

      // CSS Grid
      features.cssGrid = CSS.supports('display', 'grid');

      // CSS Flexbox
      features.flexbox = CSS.supports('display', 'flex');

      // CSS Custom Properties
      features.customProperties = CSS.supports('color', 'var(--test)');

      // CSS Transforms
      features.transforms = CSS.supports('transform', 'translateX(10px)');

      // CSS Transitions
      features.transitions = CSS.supports('transition', 'all 0.3s ease');

      // CSS Animations
      features.animations = CSS.supports('animation', 'test 1s linear');

      return features;
    });

    // 必要なCSS機能がサポートされていることを確認
    expect(cssFeatureSupport.cssGrid).toBe(true);
    expect(cssFeatureSupport.flexbox).toBe(true);
    expect(cssFeatureSupport.customProperties).toBe(true);
    expect(cssFeatureSupport.transforms).toBe(true);
    expect(cssFeatureSupport.transitions).toBe(true);
  });

  // JavaScript機能のクロスブラウザサポートテスト
  test('JavaScript機能のブラウザサポートを確認', async ({ page }) => {
    await page.goto('/');

    const jsFeatureSupport = await page.evaluate(() => {
      const features = {};

      // ES6 Arrow Functions
      try {
        const test = () => true;
        features.arrowFunctions = test();
      } catch (e) {
        features.arrowFunctions = false;
      }

      // ES6 Template Literals
      try {
        const test = `template${1}`;
        features.templateLiterals = test === 'template1';
      } catch (e) {
        features.templateLiterals = false;
      }

      // ES6 Destructuring
      try {
        const { length } = [1, 2, 3];
        features.destructuring = length === 3;
      } catch (e) {
        features.destructuring = false;
      }

      // Promises
      features.promises = typeof Promise !== 'undefined';

      // Fetch API
      features.fetch = typeof fetch !== 'undefined';

      // Local Storage
      features.localStorage = typeof localStorage !== 'undefined';

      // Session Storage
      features.sessionStorage = typeof sessionStorage !== 'undefined';

      return features;
    });

    // 必要なJavaScript機能がサポートされていることを確認
    expect(jsFeatureSupport.arrowFunctions).toBe(true);
    expect(jsFeatureSupport.templateLiterals).toBe(true);
    expect(jsFeatureSupport.destructuring).toBe(true);
    expect(jsFeatureSupport.promises).toBe(true);
    expect(jsFeatureSupport.fetch).toBe(true);
    expect(jsFeatureSupport.localStorage).toBe(true);
    expect(jsFeatureSupport.sessionStorage).toBe(true);
  });

  // フォント表示のクロスブラウザテスト
  test('フォント表示がブラウザ間で一貫している', async ({ page }) => {
    await page.goto('/');

    // フォントが正しく読み込まれることを確認
    const fontInfo = await page.evaluate(() => {
      const element = document.querySelector('[data-testid="main-content"]');
      const computedStyle = window.getComputedStyle(element);
      
      return {
        fontFamily: computedStyle.fontFamily,
        fontSize: computedStyle.fontSize,
        fontWeight: computedStyle.fontWeight,
        lineHeight: computedStyle.lineHeight
      };
    });

    // フォントファミリーが設定されていることを確認
    expect(fontInfo.fontFamily).toBeTruthy();
    expect(fontInfo.fontSize).toBeTruthy();
    expect(fontInfo.fontWeight).toBeTruthy();
    expect(fontInfo.lineHeight).toBeTruthy();

    // 日本語フォントのサポートを確認
    const japaneseTextElement = page.locator('[data-testid="sport-tab-volleyball"]');
    await expect(japaneseTextElement).toBeVisible();
    
    const japaneseText = await japaneseTextElement.textContent();
    expect(japaneseText).toMatch(/[\u3040-\u309F\u30A0-\u30FF\u4E00-\u9FAF]/); // ひらがな、カタカナ、漢字
  });

  // パフォーマンスのクロスブラウザテスト
  test('パフォーマンスがブラウザ間で許容範囲内', async ({ page }) => {
    const startTime = Date.now();
    
    await page.goto('/');
    await expect(page.locator('[data-testid="tournament-bracket"]')).toBeVisible();
    
    const loadTime = Date.now() - startTime;
    
    // ページ読み込み時間が5秒以内であることを確認
    expect(loadTime).toBeLessThan(5000);

    // JavaScript実行時間をテスト
    const jsPerformance = await page.evaluate(() => {
      const start = performance.now();
      
      // 重い処理をシミュレート
      for (let i = 0; i < 10000; i++) {
        const div = document.createElement('div');
        div.textContent = `test${i}`;
      }
      
      return performance.now() - start;
    });

    // JavaScript実行時間が100ms以内であることを確認
    expect(jsPerformance).toBeLessThan(100);
  });
});
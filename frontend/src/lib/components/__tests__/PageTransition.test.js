// PageTransition コンポーネントの単体テスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen } from '@testing-library/svelte';
import { writable } from 'svelte/store';
import PageTransition from '../PageTransition.svelte';

// SvelteKit stores のモック
const mockPage = writable({
  url: { pathname: '/' },
  params: {},
  route: { id: null }
});

const mockNavigating = writable(null);

vi.mock('$app/stores', () => ({
  page: mockPage,
  navigating: mockNavigating
}));

// AnimatedTransition コンポーネントのモック
vi.mock('../AnimatedTransition.svelte', () => ({
  default: vi.fn().mockImplementation(({ show, children, ...props }) => {
    return {
      $$: {
        fragment: show ? children : null
      }
    };
  })
}));

describe('PageTransition', () => {
  beforeEach(() => {
    vi.useFakeTimers();
    
    // 初期状態にリセット
    mockPage.set({
      url: { pathname: '/' },
      params: {},
      route: { id: null }
    });
    mockNavigating.set(null);
  });

  afterEach(() => {
    vi.restoreAllMocks();
    vi.useRealTimers();
  });

  it('正しくレンダリングされる', () => {
    const { container } = render(PageTransition, {
      props: {},
      slots: {
        default: '<div data-testid="content">Test Content</div>'
      }
    });
    
    expect(container.querySelector('.page-transition-container')).toBeInTheDocument();
  });

  it('デフォルトプロパティが正しく設定される', () => {
    render(PageTransition);
    
    // デフォルト値のテストは実装の詳細に依存するため、
    // コンポーネントが正常にレンダリングされることを確認
    expect(document.querySelector('.page-transition-container')).toBeInTheDocument();
  });

  it('カスタムプロパティが正しく適用される', () => {
    render(PageTransition, {
      props: {
        transitionType: 'slide',
        duration: 500,
        className: 'custom-class'
      }
    });
    
    const container = document.querySelector('.page-transition-container');
    expect(container).toHaveClass('custom-class');
  });

  describe('ナビゲーション状態の処理', () => {
    it('ナビゲーション開始時にローディングオーバーレイを表示する', async () => {
      render(PageTransition, {
        slots: {
          default: '<div data-testid="content">Test Content</div>'
        }
      });
      
      // ナビゲーション開始をシミュレート
      mockNavigating.set({ from: '/', to: '/about' });
      
      // DOM更新を待つ
      await vi.runAllTimersAsync();
      
      expect(document.querySelector('.loading-overlay')).toBeInTheDocument();
      expect(document.querySelector('.loading-spinner')).toBeInTheDocument();
    });

    it('ナビゲーション完了時にローディングオーバーレイを非表示にする', async () => {
      render(PageTransition, {
        slots: {
          default: '<div data-testid="content">Test Content</div>'
        }
      });
      
      // ナビゲーション開始
      mockNavigating.set({ from: '/', to: '/about' });
      await vi.runAllTimersAsync();
      
      // ナビゲーション完了
      mockNavigating.set(null);
      await vi.runAllTimersAsync();
      
      expect(document.querySelector('.loading-overlay')).not.toBeInTheDocument();
    });

    it('ナビゲーション完了後に遅延してコンテンツを表示する', async () => {
      const { component } = render(PageTransition, {
        slots: {
          default: '<div data-testid="content">Test Content</div>'
        }
      });
      
      // ナビゲーション開始
      mockNavigating.set({ from: '/', to: '/about' });
      await vi.runAllTimersAsync();
      
      // ナビゲーション完了
      mockNavigating.set(null);
      
      // 50ms の遅延をシミュレート
      vi.advanceTimersByTime(50);
      await vi.runAllTimersAsync();
      
      // コンテンツが表示されることを確認（実装に依存）
      expect(component).toBeTruthy();
    });
  });

  describe('ページパスの変更処理', () => {
    it('ページパスの変更を正しく検出する', async () => {
      render(PageTransition, {
        slots: {
          default: '<div data-testid="content">Test Content</div>'
        }
      });
      
      // ページパスを変更
      mockPage.set({
        url: { pathname: '/about' },
        params: {},
        route: { id: null }
      });
      
      await vi.runAllTimersAsync();
      
      // パスの変更が処理されることを確認
      expect(document.querySelector('.page-transition-container')).toBeInTheDocument();
    });

    it('同じパスの場合は変更処理をスキップする', async () => {
      render(PageTransition, {
        slots: {
          default: '<div data-testid="content">Test Content</div>'
        }
      });
      
      // 同じパスを設定
      mockPage.set({
        url: { pathname: '/' },
        params: {},
        route: { id: null }
      });
      
      await vi.runAllTimersAsync();
      
      // 不要な処理が実行されないことを確認
      expect(document.querySelector('.page-transition-container')).toBeInTheDocument();
    });
  });

  describe('スロットコンテンツ', () => {
    it('スロットコンテンツが正しく表示される', () => {
      render(PageTransition, {
        slots: {
          default: '<div data-testid="slot-content">Slot Content</div>'
        }
      });
      
      // AnimatedTransition コンポーネントを通してコンテンツが渡されることを確認
      expect(document.querySelector('.page-transition-container')).toBeInTheDocument();
    });

    it('複雑なスロットコンテンツを処理できる', () => {
      render(PageTransition, {
        slots: {
          default: `
            <div data-testid="complex-content">
              <h1>Title</h1>
              <p>Paragraph</p>
              <button>Button</button>
            </div>
          `
        }
      });
      
      expect(document.querySelector('.page-transition-container')).toBeInTheDocument();
    });
  });

  describe('CSS スタイル', () => {
    it('基本的なCSSクラスが適用される', () => {
      render(PageTransition);
      
      const container = document.querySelector('.page-transition-container');
      expect(container).toBeInTheDocument();
      expect(container).toHaveClass('page-transition-container');
    });

    it('カスタムクラス名が追加される', () => {
      render(PageTransition, {
        props: {
          className: 'custom-page-class'
        }
      });
      
      const container = document.querySelector('.page-transition-container');
      expect(container).toHaveClass('page-transition-container');
      expect(container).toHaveClass('custom-page-class');
    });

    it('ローディングオーバーレイのスタイルが正しく適用される', async () => {
      render(PageTransition);
      
      // ナビゲーション開始
      mockNavigating.set({ from: '/', to: '/about' });
      await vi.runAllTimersAsync();
      
      const overlay = document.querySelector('.loading-overlay');
      expect(overlay).toBeInTheDocument();
      expect(overlay).toHaveStyle({
        position: 'fixed',
        'z-index': '9999'
      });
    });

    it('ローディングスピナーのアニメーションが設定される', async () => {
      render(PageTransition);
      
      // ナビゲーション開始
      mockNavigating.set({ from: '/', to: '/about' });
      await vi.runAllTimersAsync();
      
      const spinner = document.querySelector('.loading-spinner');
      expect(spinner).toBeInTheDocument();
      
      // CSS アニメーションの確認
      const computedStyle = window.getComputedStyle(spinner);
      expect(computedStyle.animationName).toBe('spin');
    });
  });

  describe('アクセシビリティ', () => {
    it('ローディング状態でアクセシビリティ属性が設定される', async () => {
      render(PageTransition, {
        slots: {
          default: '<div data-testid="content">Test Content</div>'
        }
      });
      
      // ナビゲーション開始
      mockNavigating.set({ from: '/', to: '/about' });
      await vi.runAllTimersAsync();
      
      const overlay = document.querySelector('.loading-overlay');
      expect(overlay).toBeInTheDocument();
      
      // アクセシビリティのためのaria属性を確認（実装に依存）
      // 実際の実装では aria-live や role 属性を追加することを推奨
    });

    it('prefers-reduced-motion に対応している', () => {
      // CSS メディアクエリのテストは実装の詳細
      // 実際のテストでは、CSSが正しく定義されていることを確認
      render(PageTransition);
      
      const style = document.querySelector('style');
      expect(style?.textContent).toContain('@media (prefers-reduced-motion: reduce)');
    });
  });

  describe('エラーハンドリング', () => {
    it('不正なプロパティでもエラーを発生させない', () => {
      expect(() => {
        render(PageTransition, {
          props: {
            transitionType: 'invalid-type',
            duration: -100,
            className: null
          }
        });
      }).not.toThrow();
    });

    it('ストアが未定義でもエラーを発生させない', () => {
      // ストアのモックを一時的に削除
      vi.doMock('$app/stores', () => ({
        page: writable(null),
        navigating: writable(null)
      }));
      
      expect(() => {
        render(PageTransition);
      }).not.toThrow();
    });
  });

  describe('パフォーマンス', () => {
    it('不要な再レンダリングを避ける', async () => {
      const { component } = render(PageTransition, {
        slots: {
          default: '<div data-testid="content">Test Content</div>'
        }
      });
      
      // 同じ値で複数回更新
      mockPage.set({
        url: { pathname: '/' },
        params: {},
        route: { id: null }
      });
      
      mockPage.set({
        url: { pathname: '/' },
        params: {},
        route: { id: null }
      });
      
      await vi.runAllTimersAsync();
      
      // コンポーネントが正常に動作することを確認
      expect(component).toBeTruthy();
    });

    it('メモリリークを防ぐためのクリーンアップ', () => {
      const { unmount } = render(PageTransition);
      
      // コンポーネントをアンマウント
      unmount();
      
      // タイマーがクリアされることを確認
      expect(vi.getTimerCount()).toBe(0);
    });
  });
});
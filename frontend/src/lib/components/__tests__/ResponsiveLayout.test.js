import { render, screen } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import ResponsiveLayout from '../ResponsiveLayout.svelte';

// モックウィンドウオブジェクト
const mockWindow = {
  innerWidth: 1024,
  addEventListener: vi.fn(),
  removeEventListener: vi.fn()
};

// ResizeObserverのモック
class MockResizeObserver {
  constructor(callback) {
    this.callback = callback;
  }
  observe() {}
  disconnect() {}
}

describe('ResponsiveLayout', () => {
  beforeEach(() => {
    // グローバルオブジェクトのモック
    global.window = mockWindow;
    global.ResizeObserver = MockResizeObserver;
    global.document = {
      documentElement: {}
    };
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('デフォルトプロパティでレンダリングされる', () => {
    render(ResponsiveLayout, {
      props: {},
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    const layout = screen.getByText('Test Content').parentElement;
    expect(layout).toHaveClass('responsive-layout');
    expect(layout).toHaveClass('container');
    expect(layout).toHaveClass('with-padding');
  });

  it('fluidプロパティが適用される', () => {
    render(ResponsiveLayout, {
      props: { fluid: true },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    const layout = screen.getByText('Test Content').parentElement;
    expect(layout).toHaveClass('container-fluid');
    expect(layout).not.toHaveClass('container');
  });

  it('paddingプロパティが無効化される', () => {
    render(ResponsiveLayout, {
      props: { padding: false },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    const layout = screen.getByText('Test Content').parentElement;
    expect(layout).not.toHaveClass('with-padding');
  });

  it('カスタムクラス名が適用される', () => {
    render(ResponsiveLayout, {
      props: { className: 'custom-class' },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    const layout = screen.getByText('Test Content').parentElement;
    expect(layout).toHaveClass('custom-class');
  });

  it('画面サイズに応じたクラスが適用される', () => {
    // モバイルサイズ
    mockWindow.innerWidth = 500;
    
    render(ResponsiveLayout, {
      props: {},
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    const layout = screen.getByText('Test Content').parentElement;
    expect(layout).toHaveAttribute('data-screen-size', 'mobile');
  });

  it('カスタムブレークポイントが機能する', () => {
    const customBreakpoints = {
      mobile: 600,
      tablet: 900,
      desktop: 1200
    };

    mockWindow.innerWidth = 700;

    render(ResponsiveLayout, {
      props: { breakpoints: customBreakpoints },
      $$slots: {
        default: [{ component: 'div', props: {}, children: 'Test Content' }]
      }
    });

    const layout = screen.getByText('Test Content').parentElement;
    expect(layout).toHaveAttribute('data-screen-size', 'tablet');
  });
});
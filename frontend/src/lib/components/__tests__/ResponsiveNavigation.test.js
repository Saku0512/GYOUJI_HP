import { render, screen, fireEvent } from '@testing-library/svelte';
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import ResponsiveNavigation from '../ResponsiveNavigation.svelte';

// モックウィンドウオブジェクト
const mockWindow = {
  innerWidth: 1024,
  addEventListener: vi.fn(),
  removeEventListener: vi.fn()
};

describe('ResponsiveNavigation', () => {
  const mockItems = [
    { key: 'home', label: 'ホーム', href: '/', icon: '🏠' },
    { key: 'about', label: 'About', href: '/about', icon: 'ℹ️' },
    { key: 'logout', label: 'ログアウト', onClick: vi.fn(), icon: '🚪' }
  ];

  beforeEach(() => {
    global.window = mockWindow;
    global.document = {
      addEventListener: vi.fn(),
      removeEventListener: vi.fn()
    };
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  it('ブランドとナビゲーションアイテムがレンダリングされる', () => {
    render(ResponsiveNavigation, {
      props: {
        brand: 'Test Brand',
        brandHref: '/',
        items: mockItems,
        activeItem: 'home'
      }
    });

    expect(screen.getByText('Test Brand')).toBeInTheDocument();
    expect(screen.getByText('ホーム')).toBeInTheDocument();
    expect(screen.getByText('About')).toBeInTheDocument();
    expect(screen.getByText('ログアウト')).toBeInTheDocument();
  });

  it('アクティブアイテムが正しくハイライトされる', () => {
    render(ResponsiveNavigation, {
      props: {
        brand: 'Test Brand',
        items: mockItems,
        activeItem: 'home'
      }
    });

    const homeLink = screen.getByText('ホーム').closest('a');
    expect(homeLink).toHaveClass('active');
  });

  it('デスクトップ表示でモバイルメニューボタンが非表示', () => {
    mockWindow.innerWidth = 1024;
    
    render(ResponsiveNavigation, {
      props: {
        brand: 'Test Brand',
        items: mockItems
      }
    });

    const nav = screen.getByRole('navigation');
    expect(nav).toHaveClass('nav-desktop');
  });

  it('モバイル表示でハンバーガーメニューが表示される', () => {
    mockWindow.innerWidth = 500;
    
    render(ResponsiveNavigation, {
      props: {
        brand: 'Test Brand',
        items: mockItems,
        mobileBreakpoint: 768
      }
    });

    const nav = screen.getByRole('navigation');
    expect(nav).toHaveClass('nav-mobile');
    
    const toggleButton = screen.getByLabelText('メニューを開く');
    expect(toggleButton).toBeInTheDocument();
  });

  it('モバイルメニューの開閉が機能する', async () => {
    mockWindow.innerWidth = 500;
    
    const { component } = render(ResponsiveNavigation, {
      props: {
        brand: 'Test Brand',
        items: mockItems,
        mobileBreakpoint: 768
      }
    });

    const toggleButton = screen.getByLabelText('メニューを開く');
    
    // メニューを開く
    await fireEvent.click(toggleButton);
    
    const menu = screen.getByRole('list').closest('.nav-menu');
    expect(menu).toHaveClass('show');
    
    // メニューを閉じる
    await fireEvent.click(toggleButton);
    expect(menu).not.toHaveClass('show');
  });

  it('ナビゲーションアイテムクリック時にイベントが発火される', async () => {
    const mockOnItemClick = vi.fn();
    
    render(ResponsiveNavigation, {
      props: {
        brand: 'Test Brand',
        items: mockItems
      }
    });

    // コンポーネントのイベントリスナーを設定
    const nav = screen.getByRole('navigation');
    nav.addEventListener('itemClick', mockOnItemClick);

    const logoutButton = screen.getByText('ログアウト');
    await fireEvent.click(logoutButton);

    expect(mockItems[2].onClick).toHaveBeenCalled();
  });

  it('アイコンが正しく表示される', () => {
    render(ResponsiveNavigation, {
      props: {
        brand: 'Test Brand',
        items: mockItems
      }
    });

    expect(screen.getByText('🏠')).toBeInTheDocument();
    expect(screen.getByText('ℹ️')).toBeInTheDocument();
    expect(screen.getByText('🚪')).toBeInTheDocument();
  });

  it('カスタムクラス名が適用される', () => {
    render(ResponsiveNavigation, {
      props: {
        brand: 'Test Brand',
        items: mockItems,
        className: 'custom-nav'
      }
    });

    const nav = screen.getByRole('navigation');
    expect(nav).toHaveClass('custom-nav');
  });

  it('ブランドリンクが正しく設定される', () => {
    render(ResponsiveNavigation, {
      props: {
        brand: 'Test Brand',
        brandHref: '/custom',
        items: mockItems
      }
    });

    const brandLink = screen.getByText('Test Brand').closest('a');
    expect(brandLink).toHaveAttribute('href', '/custom');
  });
});
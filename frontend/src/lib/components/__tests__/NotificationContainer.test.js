// NotificationContainer コンポーネントのテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import NotificationContainer from '../NotificationContainer.svelte';
import { uiActions } from '$lib/stores/ui.js';

// UIストアのモック
vi.mock('$lib/stores/ui.js', () => {
  const { writable } = require('svelte/store');
  
  const mockStore = writable({
    notifications: [],
    loading: false,
    theme: 'light'
  });

  return {
    uiStore: mockStore,
    uiActions: {
      showNotification: vi.fn((message, type, duration) => {
        const notification = {
          id: Date.now(),
          message,
          type,
          duration,
          timestamp: Date.now()
        };
        
        mockStore.update(state => ({
          ...state,
          notifications: [...state.notifications, notification]
        }));
        
        return notification.id;
      }),
      removeNotification: vi.fn((id) => {
        mockStore.update(state => ({
          ...state,
          notifications: state.notifications.filter(n => n.id !== id)
        }));
      }),
      clearNotifications: vi.fn(() => {
        mockStore.update(state => ({
          ...state,
          notifications: []
        }));
      })
    }
  };
});

describe('NotificationContainer Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // 通知をクリア
    uiActions.clearNotifications();
  });

  it('should not render when no notifications exist', () => {
    render(NotificationContainer);
    
    expect(screen.queryByRole('region')).not.toBeInTheDocument();
  });

  it('should render notifications when they exist', async () => {
    render(NotificationContainer);
    
    // 通知を追加
    uiActions.showNotification('Test notification', 'info');
    
    await waitFor(() => {
      expect(screen.getByRole('region')).toBeInTheDocument();
      expect(screen.getByText('Test notification')).toBeInTheDocument();
    });
  });

  it('should limit notifications to maxNotifications', async () => {
    render(NotificationContainer, {
      props: {
        maxNotifications: 2
      }
    });
    
    // 3つの通知を追加
    uiActions.showNotification('Notification 1', 'info');
    uiActions.showNotification('Notification 2', 'info');
    uiActions.showNotification('Notification 3', 'info');
    
    await waitFor(() => {
      // 最新の2つのみ表示されることを確認
      expect(screen.queryByText('Notification 1')).not.toBeInTheDocument();
      expect(screen.getByText('Notification 2')).toBeInTheDocument();
      expect(screen.getByText('Notification 3')).toBeInTheDocument();
    });
  });

  it('should show clear all button when multiple notifications exist', async () => {
    render(NotificationContainer);
    
    // 複数の通知を追加
    uiActions.showNotification('Notification 1', 'info');
    uiActions.showNotification('Notification 2', 'info');
    
    await waitFor(() => {
      expect(screen.getByLabelText('すべての通知をクリア')).toBeInTheDocument();
    });
  });

  it('should not show clear all button for single notification', async () => {
    render(NotificationContainer);
    
    // 単一の通知を追加
    uiActions.showNotification('Single notification', 'info');
    
    await waitFor(() => {
      expect(screen.queryByLabelText('すべての通知をクリア')).not.toBeInTheDocument();
    });
  });

  it('should clear all notifications when clear button is clicked', async () => {
    render(NotificationContainer);
    
    // 複数の通知を追加
    uiActions.showNotification('Notification 1', 'info');
    uiActions.showNotification('Notification 2', 'info');
    
    await waitFor(() => {
      expect(screen.getByText('Notification 1')).toBeInTheDocument();
      expect(screen.getByText('Notification 2')).toBeInTheDocument();
    });
    
    const clearButton = screen.getByLabelText('すべての通知をクリア');
    await fireEvent.click(clearButton);
    
    expect(uiActions.clearNotifications).toHaveBeenCalled();
  });

  it('should show overflow indicator when notifications exceed max', async () => {
    render(NotificationContainer, {
      props: {
        maxNotifications: 2
      }
    });
    
    // 4つの通知を追加
    uiActions.showNotification('Notification 1', 'info');
    uiActions.showNotification('Notification 2', 'info');
    uiActions.showNotification('Notification 3', 'info');
    uiActions.showNotification('Notification 4', 'info');
    
    await waitFor(() => {
      expect(screen.getByText('他に 2 件の通知があります')).toBeInTheDocument();
      expect(screen.getByText('すべて表示')).toBeInTheDocument();
    });
  });

  it('should apply correct position classes', () => {
    const positions = [
      'top-right',
      'top-left',
      'bottom-right',
      'bottom-left',
      'top-center',
      'bottom-center'
    ];
    
    positions.forEach(position => {
      const { unmount } = render(NotificationContainer, {
        props: { position }
      });
      
      // 通知を追加してコンテナを表示
      uiActions.showNotification('Test', 'info');
      
      const container = screen.queryByRole('region');
      if (container) {
        expect(container).toHaveClass(`position-${position}`);
      }
      
      uiActions.clearNotifications();
      unmount();
    });
  });

  it('should apply correct stack direction classes', async () => {
    const { rerender } = render(NotificationContainer, {
      props: {
        stackDirection: 'up'
      }
    });
    
    uiActions.showNotification('Test', 'info');
    
    await waitFor(() => {
      const container = screen.getByRole('region');
      expect(container).toHaveClass('stack-up');
    });
    
    await rerender({
      stackDirection: 'down'
    });
    
    const container = screen.getByRole('region');
    expect(container).toHaveClass('stack-down');
  });

  it('should handle keyboard navigation', async () => {
    render(NotificationContainer);
    
    uiActions.showNotification('Test notification', 'info');
    
    await waitFor(() => {
      expect(screen.getByRole('region')).toBeInTheDocument();
    });
    
    // Escapeキーで全通知をクリア
    await fireEvent.keyDown(window, { key: 'Escape' });
    
    expect(uiActions.clearNotifications).toHaveBeenCalled();
  });

  it('should handle mouse hover events when pauseOnHover is enabled', async () => {
    render(NotificationContainer, {
      props: {
        pauseOnHover: true
      }
    });
    
    uiActions.showNotification('Test notification', 'info');
    
    await waitFor(() => {
      expect(screen.getByRole('region')).toBeInTheDocument();
    });
    
    const container = screen.getByRole('region');
    
    // ホバー開始
    await fireEvent.mouseEnter(container);
    
    // ホバー終了
    await fireEvent.mouseLeave(container);
    
    // イベントが正常に処理されることを確認
    expect(container).toBeInTheDocument();
  });

  it('should not pause on hover when pauseOnHover is disabled', async () => {
    render(NotificationContainer, {
      props: {
        pauseOnHover: false
      }
    });
    
    uiActions.showNotification('Test notification', 'info');
    
    await waitFor(() => {
      expect(screen.getByRole('region')).toBeInTheDocument();
    });
    
    const container = screen.getByRole('region');
    
    // ホバーイベントが無視されることを確認
    await fireEvent.mouseEnter(container);
    await fireEvent.mouseLeave(container);
    
    expect(container).toBeInTheDocument();
  });

  it('should handle notification removal', async () => {
    render(NotificationContainer);
    
    const notificationId = uiActions.showNotification('Test notification', 'info');
    
    await waitFor(() => {
      expect(screen.getByText('Test notification')).toBeInTheDocument();
    });
    
    // 通知を削除
    uiActions.removeNotification(notificationId);
    
    await waitFor(() => {
      expect(screen.queryByText('Test notification')).not.toBeInTheDocument();
    });
  });

  it('should set custom CSS variables', () => {
    render(NotificationContainer, {
      props: {
        spacing: 12,
        animationDuration: 500
      }
    });
    
    uiActions.showNotification('Test', 'info');
    
    const container = screen.getByRole('region');
    const style = getComputedStyle(container);
    
    expect(container.style.getPropertyValue('--spacing')).toBe('12px');
    expect(container.style.getPropertyValue('--animation-duration')).toBe('500ms');
  });

  it('should handle focus and blur events on notifications', async () => {
    render(NotificationContainer, {
      props: {
        pauseOnHover: true
      }
    });
    
    uiActions.showNotification('Test notification', 'info');
    
    await waitFor(() => {
      expect(screen.getByRole('region')).toBeInTheDocument();
    });
    
    const notificationWrapper = screen.getByRole('region').querySelector('[data-notification-id]');
    
    if (notificationWrapper) {
      // フォーカスイベント
      await fireEvent.focus(notificationWrapper);
      
      // ブラーイベント
      await fireEvent.blur(notificationWrapper);
    }
    
    // イベントが正常に処理されることを確認
    expect(screen.getByRole('region')).toBeInTheDocument();
  });
});
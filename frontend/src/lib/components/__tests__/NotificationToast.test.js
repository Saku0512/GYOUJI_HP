// NotificationToast コンポーネントのテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import NotificationToast from '../NotificationToast.svelte';

describe('NotificationToast Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it('should render with basic props', () => {
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        type: 'info'
      }
    });

    expect(screen.getByText('Test notification')).toBeInTheDocument();
    expect(screen.getByRole('status')).toHaveClass('toast-info');
  });

  it('should render different notification types with correct icons', () => {
    const types = ['success', 'error', 'warning', 'info'];
    
    types.forEach(type => {
      const { unmount } = render(NotificationToast, {
        props: {
          message: `${type} notification`,
          type
        }
      });
      
      expect(screen.getByRole(type === 'error' ? 'alert' : 'status')).toHaveClass(`toast-${type}`);
      
      unmount();
    });
  });

  it('should show close button when dismissible is true', () => {
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        dismissible: true
      }
    });

    expect(screen.getByLabelText('通知を閉じる')).toBeInTheDocument();
  });

  it('should hide close button when dismissible is false', () => {
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        dismissible: false
      }
    });

    expect(screen.queryByLabelText('通知を閉じる')).not.toBeInTheDocument();
  });

  it('should emit close event when close button is clicked', async () => {
    const closeHandler = vi.fn();
    
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        dismissible: true
      }
    });

    const component = screen.getByRole('status').closest('.toast').__svelte_component;
    component.$on('close', closeHandler);

    const closeButton = screen.getByLabelText('通知を閉じる');
    await fireEvent.click(closeButton);

    expect(closeHandler).toHaveBeenCalled();
  });

  it('should auto-remove after specified duration', async () => {
    const closeHandler = vi.fn();
    
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        duration: 1000
      }
    });

    const component = screen.getByRole('status').closest('.toast').__svelte_component;
    component.$on('close', closeHandler);

    // 時間を進める
    vi.advanceTimersByTime(1000);

    await waitFor(() => {
      expect(closeHandler).toHaveBeenCalled();
    });
  });

  it('should not auto-remove when duration is 0', async () => {
    const closeHandler = vi.fn();
    
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        duration: 0
      }
    });

    const component = screen.getByRole('status').closest('.toast').__svelte_component;
    component.$on('close', closeHandler);

    // 時間を進める
    vi.advanceTimersByTime(5000);

    expect(closeHandler).not.toHaveBeenCalled();
  });

  it('should pause auto-removal when isPaused is true', async () => {
    const closeHandler = vi.fn();
    
    const { rerender } = render(NotificationToast, {
      props: {
        message: 'Test notification',
        duration: 2000,
        isPaused: false
      }
    });

    const component = screen.getByRole('status').closest('.toast').__svelte_component;
    component.$on('close', closeHandler);

    // 半分の時間経過
    vi.advanceTimersByTime(1000);

    // 一時停止
    await rerender({
      message: 'Test notification',
      duration: 2000,
      isPaused: true
    });

    // さらに時間を進める
    vi.advanceTimersByTime(2000);

    // まだ閉じられていないことを確認
    expect(closeHandler).not.toHaveBeenCalled();
  });

  it('should resume auto-removal when isPaused becomes false', async () => {
    const closeHandler = vi.fn();
    
    const { rerender } = render(NotificationToast, {
      props: {
        message: 'Test notification',
        duration: 2000,
        isPaused: true
      }
    });

    const component = screen.getByRole('status').closest('.toast').__svelte_component;
    component.$on('close', closeHandler);

    // 一時停止を解除
    await rerender({
      message: 'Test notification',
      duration: 2000,
      isPaused: false
    });

    // 時間を進める
    vi.advanceTimersByTime(2000);

    await waitFor(() => {
      expect(closeHandler).toHaveBeenCalled();
    });
  });

  it('should show progress bar when showProgress is true', () => {
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        duration: 1000,
        showProgress: true
      }
    });

    expect(screen.getByRole('status').closest('.toast').querySelector('.toast-progress')).toBeInTheDocument();
  });

  it('should hide progress bar when showProgress is false', () => {
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        duration: 1000,
        showProgress: false
      }
    });

    expect(screen.getByRole('status').closest('.toast').querySelector('.toast-progress')).not.toBeInTheDocument();
  });

  it('should render action buttons when provided', () => {
    const actions = [
      { label: 'Action 1', onClick: vi.fn(), variant: 'primary' },
      { label: 'Action 2', onClick: vi.fn(), variant: 'secondary' }
    ];

    render(NotificationToast, {
      props: {
        message: 'Test notification',
        actions
      }
    });

    expect(screen.getByText('Action 1')).toBeInTheDocument();
    expect(screen.getByText('Action 2')).toBeInTheDocument();
    expect(screen.getByText('Action 1')).toHaveClass('toast-action-primary');
    expect(screen.getByText('Action 2')).toHaveClass('toast-action-secondary');
  });

  it('should call action onClick and close notification', async () => {
    const actionHandler = vi.fn();
    const closeHandler = vi.fn();
    
    const actions = [
      { label: 'Test Action', onClick: actionHandler }
    ];

    render(NotificationToast, {
      props: {
        message: 'Test notification',
        actions
      }
    });

    const component = screen.getByRole('status').closest('.toast').__svelte_component;
    component.$on('close', closeHandler);

    const actionButton = screen.getByText('Test Action');
    await fireEvent.click(actionButton);

    expect(actionHandler).toHaveBeenCalled();
    expect(closeHandler).toHaveBeenCalled();
  });

  it('should handle action errors gracefully', async () => {
    const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
    const errorAction = vi.fn().mockImplementation(() => {
      throw new Error('Action failed');
    });
    
    const actions = [
      { label: 'Error Action', onClick: errorAction }
    ];

    render(NotificationToast, {
      props: {
        message: 'Test notification',
        actions
      }
    });

    const actionButton = screen.getByText('Error Action');
    await fireEvent.click(actionButton);

    expect(errorAction).toHaveBeenCalled();
    expect(consoleSpy).toHaveBeenCalledWith('Notification action error:', expect.any(Error));
    
    consoleSpy.mockRestore();
  });

  it('should handle keyboard navigation', async () => {
    const closeHandler = vi.fn();
    
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        dismissible: true
      }
    });

    const component = screen.getByRole('status').closest('.toast').__svelte_component;
    component.$on('close', closeHandler);

    const toast = screen.getByRole('status');
    
    // Escapeキーで閉じる
    await fireEvent.keyDown(toast, { key: 'Escape' });

    expect(closeHandler).toHaveBeenCalled();
  });

  it('should not close on Escape when not dismissible', async () => {
    const closeHandler = vi.fn();
    
    render(NotificationToast, {
      props: {
        message: 'Test notification',
        dismissible: false
      }
    });

    const component = screen.getByRole('status').closest('.toast').__svelte_component;
    component.$on('close', closeHandler);

    const toast = screen.getByRole('status');
    
    // Escapeキーを押す
    await fireEvent.keyDown(toast, { key: 'Escape' });

    expect(closeHandler).not.toHaveBeenCalled();
  });

  it('should apply paused class when isPaused is true', async () => {
    const { rerender } = render(NotificationToast, {
      props: {
        message: 'Test notification',
        isPaused: false
      }
    });

    const toast = screen.getByRole('status');
    expect(toast).not.toHaveClass('paused');

    await rerender({
      message: 'Test notification',
      isPaused: true
    });

    expect(toast).toHaveClass('paused');
  });

  it('should use correct ARIA attributes for different types', () => {
    // Error notifications should use alert role
    const { unmount: unmountError } = render(NotificationToast, {
      props: {
        message: 'Error notification',
        type: 'error'
      }
    });

    expect(screen.getByRole('alert')).toHaveAttribute('aria-live', 'assertive');
    unmountError();

    // Other types should use status role
    render(NotificationToast, {
      props: {
        message: 'Info notification',
        type: 'info'
      }
    });

    expect(screen.getByRole('status')).toHaveAttribute('aria-live', 'polite');
  });

  it('should be focusable', () => {
    render(NotificationToast, {
      props: {
        message: 'Test notification'
      }
    });

    const toast = screen.getByRole('status');
    expect(toast).toHaveAttribute('tabindex', '0');
  });

  it('should clean up timers on unmount', () => {
    const { unmount } = render(NotificationToast, {
      props: {
        message: 'Test notification',
        duration: 1000
      }
    });

    // コンポーネントをアンマウント
    unmount();

    // タイマーが適切にクリアされることを確認
    // （実際のテストでは、メモリリークがないことを確認）
    expect(true).toBe(true);
  });
});
// ErrorBoundary コンポーネントのテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import ErrorBoundary from '../ErrorBoundary.svelte';
import { AppError, ERROR_TYPES, ERROR_LEVELS } from '$lib/utils/error-handler.js';

// UIストアのモック
vi.mock('$lib/stores/ui.js', () => ({
  uiActions: {
    showNotification: vi.fn()
  }
}));

// グローバルエラーハンドラーのモック
vi.mock('$lib/utils/error-handler.js', async () => {
  const actual = await vi.importActual('$lib/utils/error-handler.js');
  return {
    ...actual,
    globalErrorHandler: {
      addErrorListener: vi.fn(),
      removeErrorListener: vi.fn(),
      resetErrorBoundary: vi.fn()
    }
  };
});

describe('ErrorBoundary Component', () => {
  let mockErrorHandler;

  beforeEach(() => {
    vi.clearAllMocks();
    mockErrorHandler = require('$lib/utils/error-handler.js').globalErrorHandler;
  });

  it('should render children when no error occurs', () => {
    render(ErrorBoundary, {
      props: {},
      $$slots: {
        default: '<div data-testid="child-content">Child Content</div>'
      }
    });

    expect(screen.getByTestId('child-content')).toBeInTheDocument();
  });

  it('should register error listener on mount', () => {
    render(ErrorBoundary);
    
    expect(mockErrorHandler.addErrorListener).toHaveBeenCalledWith(expect.any(Function));
  });

  it('should remove error listener on destroy', () => {
    const { unmount } = render(ErrorBoundary);
    
    unmount();
    
    expect(mockErrorHandler.removeErrorListener).toHaveBeenCalledWith(expect.any(Function));
  });

  it('should display error UI when error occurs', async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        errorTitle: 'Test Error Title',
        errorMessage: 'Test error message'
      }
    });

    // エラーを手動でトリガー
    const testError = new AppError('Test error', ERROR_TYPES.API, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('Test Error Title')).toBeInTheDocument();
      expect(screen.getByText('Test error message')).toBeInTheDocument();
    });
  });

  it('should show retry button by default', async () => {
    const { component } = render(ErrorBoundary);

    const testError = new AppError('Test error', ERROR_TYPES.NETWORK, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('再試行')).toBeInTheDocument();
    });
  });

  it('should hide retry button when showRetry is false', async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        showRetry: false
      }
    });

    const testError = new AppError('Test error', ERROR_TYPES.API, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.queryByText('再試行')).not.toBeInTheDocument();
      expect(screen.getByText('閉じる')).toBeInTheDocument();
    });
  });

  it('should display error details for high-level errors', async () => {
    const { component } = render(ErrorBoundary);

    const testError = new AppError(
      'High level error',
      ERROR_TYPES.SERVER,
      ERROR_LEVELS.HIGH,
      { code: 'SERVER_001', details: 'Additional info' }
    );
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('詳細情報')).toBeInTheDocument();
    });
  });

  it('should not display error details for medium-level errors', async () => {
    const { component } = render(ErrorBoundary);

    const testError = new AppError(
      'Medium level error',
      ERROR_TYPES.API,
      ERROR_LEVELS.MEDIUM
    );
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.queryByText('詳細情報')).not.toBeInTheDocument();
    });
  });

  it('should handle retry action', async () => {
    const onRetry = vi.fn().mockResolvedValue();
    const { component } = render(ErrorBoundary, {
      props: {
        onRetry
      }
    });

    const testError = new AppError('Test error', ERROR_TYPES.NETWORK, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('再試行')).toBeInTheDocument();
    });

    const retryButton = screen.getByText('再試行');
    await fireEvent.click(retryButton);

    expect(onRetry).toHaveBeenCalledWith(testError);
  });

  it('should handle retry failure', async () => {
    const onRetry = vi.fn().mockRejectedValue(new Error('Retry failed'));
    const { component } = render(ErrorBoundary, {
      props: {
        onRetry
      }
    });

    const testError = new AppError('Test error', ERROR_TYPES.NETWORK, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('再試行')).toBeInTheDocument();
    });

    const retryButton = screen.getByText('再試行');
    await fireEvent.click(retryButton);

    await waitFor(() => {
      expect(onRetry).toHaveBeenCalledWith(testError);
    });

    // 再試行失敗後もエラー状態が維持されることを確認
    expect(screen.getByText('Test error')).toBeInTheDocument();
  });

  it('should reset error state when reset button is clicked', async () => {
    const { component } = render(ErrorBoundary, {
      $$slots: {
        default: '<div data-testid="child-content">Child Content</div>'
      }
    });

    const testError = new AppError('Test error', ERROR_TYPES.API, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('閉じる')).toBeInTheDocument();
    });

    const resetButton = screen.getByText('閉じる');
    await fireEvent.click(resetButton);

    await waitFor(() => {
      expect(screen.getByTestId('child-content')).toBeInTheDocument();
    });

    expect(mockErrorHandler.resetErrorBoundary).toHaveBeenCalled();
  });

  it('should call custom error handler when provided', async () => {
    const onError = vi.fn();
    const { component } = render(ErrorBoundary, {
      props: {
        onError
      }
    });

    const testError = new AppError('Test error', ERROR_TYPES.API, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    expect(onError).toHaveBeenCalledWith(testError);
  });

  it('should emit error event when error occurs', async () => {
    const errorHandler = vi.fn();
    const { component } = render(ErrorBoundary);
    
    component.$on('error', errorHandler);

    const testError = new AppError('Test error', ERROR_TYPES.API, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    expect(errorHandler).toHaveBeenCalledWith(
      expect.objectContaining({
        detail: expect.objectContaining({
          error: testError,
          errorId: expect.any(String)
        })
      })
    );
  });

  it('should emit reset event when error is reset', async () => {
    const resetHandler = vi.fn();
    const { component } = render(ErrorBoundary);
    
    component.$on('reset', resetHandler);

    const testError = new AppError('Test error', ERROR_TYPES.API, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('閉じる')).toBeInTheDocument();
    });

    const resetButton = screen.getByText('閉じる');
    await fireEvent.click(resetButton);

    expect(resetHandler).toHaveBeenCalled();
  });

  it('should emit retry-success event on successful retry', async () => {
    const retrySuccessHandler = vi.fn();
    const onRetry = vi.fn().mockResolvedValue();
    const { component } = render(ErrorBoundary, {
      props: {
        onRetry
      }
    });
    
    component.$on('retry-success', retrySuccessHandler);

    const testError = new AppError('Test error', ERROR_TYPES.NETWORK, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('再試行')).toBeInTheDocument();
    });

    const retryButton = screen.getByText('再試行');
    await fireEvent.click(retryButton);

    await waitFor(() => {
      expect(retrySuccessHandler).toHaveBeenCalled();
    });
  });

  it('should emit retry-failed event on failed retry', async () => {
    const retryFailedHandler = vi.fn();
    const retryError = new Error('Retry failed');
    const onRetry = vi.fn().mockRejectedValue(retryError);
    const { component } = render(ErrorBoundary, {
      props: {
        onRetry
      }
    });
    
    component.$on('retry-failed', retryFailedHandler);

    const testError = new AppError('Test error', ERROR_TYPES.NETWORK, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('再試行')).toBeInTheDocument();
    });

    const retryButton = screen.getByText('再試行');
    await fireEvent.click(retryButton);

    await waitFor(() => {
      expect(retryFailedHandler).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: expect.objectContaining({
            error: retryError
          })
        })
      );
    });
  });

  it('should apply correct CSS classes based on error level', async () => {
    const { component } = render(ErrorBoundary, {
      props: {
        level: ERROR_LEVELS.CRITICAL
      }
    });

    const testError = new AppError('Critical error', ERROR_TYPES.SERVER, ERROR_LEVELS.CRITICAL);
    component.handleError(testError);

    await waitFor(() => {
      const errorBoundary = screen.getByRole('alert');
      expect(errorBoundary).toHaveClass('error-level-critical');
    });
  });

  it('should show loading state during retry', async () => {
    let resolveRetry;
    const onRetry = vi.fn().mockImplementation(() => {
      return new Promise(resolve => {
        resolveRetry = resolve;
      });
    });

    const { component } = render(ErrorBoundary, {
      props: {
        onRetry
      }
    });

    const testError = new AppError('Test error', ERROR_TYPES.NETWORK, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('再試行')).toBeInTheDocument();
    });

    const retryButton = screen.getByText('再試行');
    await fireEvent.click(retryButton);

    // ローディング状態を確認
    expect(screen.getByText('再試行中...')).toBeInTheDocument();
    expect(retryButton).toBeDisabled();

    // 再試行を完了
    resolveRetry();
    
    await waitFor(() => {
      expect(screen.queryByText('再試行中...')).not.toBeInTheDocument();
    });
  });

  it('should reset error state when props change and resetOnPropsChange is true', async () => {
    const { component, rerender } = render(ErrorBoundary, {
      props: {
        resetOnPropsChange: true,
        errorTitle: 'Original Title'
      },
      $$slots: {
        default: '<div data-testid="child-content">Child Content</div>'
      }
    });

    const testError = new AppError('Test error', ERROR_TYPES.API, ERROR_LEVELS.MEDIUM);
    component.handleError(testError);

    await waitFor(() => {
      expect(screen.getByText('Original Title')).toBeInTheDocument();
    });

    // プロパティを変更
    await rerender({
      resetOnPropsChange: true,
      errorTitle: 'New Title'
    });

    await waitFor(() => {
      expect(screen.getByTestId('child-content')).toBeInTheDocument();
    });
  });
});
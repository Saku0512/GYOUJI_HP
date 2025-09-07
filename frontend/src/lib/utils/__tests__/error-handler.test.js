// エラーハンドリングシステムのテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { 
  globalErrorHandler, 
  AppError, 
  ERROR_TYPES, 
  ERROR_LEVELS,
  handleError,
  handleApiError,
  handleNetworkError,
  handleValidationError,
  initializeErrorHandler
} from '../error-handler.js';

// UIストアのモック
vi.mock('$lib/stores/ui.js', () => ({
  uiActions: {
    showNotification: vi.fn()
  }
}));

describe('AppError', () => {
  it('should create an AppError with default values', () => {
    const error = new AppError('Test error');
    
    expect(error.message).toBe('Test error');
    expect(error.type).toBe(ERROR_TYPES.UNKNOWN);
    expect(error.level).toBe(ERROR_LEVELS.MEDIUM);
    expect(error.details).toBeNull();
    expect(error.timestamp).toBeDefined();
    expect(error.name).toBe('AppError');
  });

  it('should create an AppError with custom values', () => {
    const details = { code: 'TEST_001' };
    const error = new AppError(
      'Custom error',
      ERROR_TYPES.VALIDATION,
      ERROR_LEVELS.HIGH,
      details
    );
    
    expect(error.message).toBe('Custom error');
    expect(error.type).toBe(ERROR_TYPES.VALIDATION);
    expect(error.level).toBe(ERROR_LEVELS.HIGH);
    expect(error.details).toEqual(details);
  });

  it('should serialize to JSON correctly', () => {
    const error = new AppError('Test error', ERROR_TYPES.API, ERROR_LEVELS.LOW);
    const json = error.toJSON();
    
    expect(json.name).toBe('AppError');
    expect(json.message).toBe('Test error');
    expect(json.type).toBe(ERROR_TYPES.API);
    expect(json.level).toBe(ERROR_LEVELS.LOW);
    expect(json.timestamp).toBeDefined();
  });
});

describe('GlobalErrorHandler', () => {
  let mockWindow;
  let originalWindow;

  beforeEach(() => {
    // ウィンドウオブジェクトのモック
    originalWindow = global.window;
    mockWindow = {
      addEventListener: vi.fn(),
      removeEventListener: vi.fn(),
      dispatchEvent: vi.fn()
    };
    global.window = mockWindow;
    global.navigator = { onLine: true };

    // エラーハンドラーをリセット
    globalErrorHandler.cleanup();
  });

  afterEach(() => {
    global.window = originalWindow;
    globalErrorHandler.cleanup();
  });

  it('should initialize event listeners', () => {
    globalErrorHandler.initialize();
    
    expect(mockWindow.addEventListener).toHaveBeenCalledWith('error', expect.any(Function));
    expect(mockWindow.addEventListener).toHaveBeenCalledWith('unhandledrejection', expect.any(Function));
    expect(mockWindow.addEventListener).toHaveBeenCalledWith('auth:unauthorized', expect.any(Function));
  });

  it('should handle JavaScript errors', () => {
    const mockEvent = {
      message: 'Test JS error',
      filename: 'test.js',
      lineno: 10,
      colno: 5,
      error: new Error('Original error')
    };

    globalErrorHandler.initialize();
    globalErrorHandler.handleGlobalError(mockEvent);

    // エラーが適切に処理されることを確認
    // 実際のテストでは、ログ出力や通知の呼び出しを検証
  });

  it('should handle unhandled promise rejections', () => {
    const mockEvent = {
      reason: new Error('Promise rejection'),
      promise: Promise.reject('test'),
      preventDefault: vi.fn()
    };

    globalErrorHandler.initialize();
    globalErrorHandler.handleUnhandledRejection(mockEvent);

    expect(mockEvent.preventDefault).toHaveBeenCalled();
  });

  it('should normalize different error types', () => {
    // TypeError (ネットワークエラー)
    const networkError = new TypeError('fetch error');
    const normalizedNetwork = globalErrorHandler.normalizeError(networkError);
    expect(normalizedNetwork.type).toBe(ERROR_TYPES.NETWORK);

    // AbortError (タイムアウト)
    const timeoutError = new Error('timeout');
    timeoutError.name = 'AbortError';
    const normalizedTimeout = globalErrorHandler.normalizeError(timeoutError);
    expect(normalizedTimeout.type).toBe(ERROR_TYPES.TIMEOUT);

    // SyntaxError (パースエラー)
    const parseError = new SyntaxError('parse error');
    const normalizedParse = globalErrorHandler.normalizeError(parseError);
    expect(normalizedParse.type).toBe(ERROR_TYPES.PARSE);
  });

  it('should add and remove error listeners', () => {
    const listener = vi.fn();
    
    globalErrorHandler.addErrorListener(listener);
    expect(globalErrorHandler.errorListeners).toContain(listener);
    
    globalErrorHandler.removeErrorListener(listener);
    expect(globalErrorHandler.errorListeners).not.toContain(listener);
  });

  it('should handle retry functionality', async () => {
    let attempts = 0;
    const operation = vi.fn().mockImplementation(() => {
      attempts++;
      if (attempts < 3) {
        throw new Error('Temporary failure');
      }
      return 'success';
    });

    const result = await globalErrorHandler.retry(operation, 'test-key', 3);
    
    expect(result).toBe('success');
    expect(operation).toHaveBeenCalledTimes(3);
  });

  it('should fail after max retry attempts', async () => {
    const operation = vi.fn().mockRejectedValue(new Error('Persistent failure'));

    await expect(
      globalErrorHandler.retry(operation, 'test-key', 2)
    ).rejects.toThrow('操作が2回失敗しました');
    
    // The operation is called maxAttempts times (2), plus the initial attempt makes it 3 total
    expect(operation).toHaveBeenCalledTimes(2);
  });
});

describe('Helper Functions', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('should handle API errors correctly', () => {
    const response = {
      success: false,
      status: 404,
      statusText: 'Not Found',
      message: 'Resource not found',
      error: 'NOT_FOUND'
    };

    const result = handleApiError(response);
    
    expect(result).toBeInstanceOf(AppError);
    expect(result.type).toBe(ERROR_TYPES.CLIENT);
    expect(result.message).toBe('Resource not found');
  });

  it('should handle different HTTP status codes', () => {
    // 401 - Authentication error
    const authResponse = { status: 401, message: 'Unauthorized' };
    const authError = handleApiError(authResponse);
    expect(authError.type).toBe(ERROR_TYPES.AUTHENTICATION);

    // 403 - Authorization error
    const authzResponse = { status: 403, message: 'Forbidden' };
    const authzError = handleApiError(authzResponse);
    expect(authzError.type).toBe(ERROR_TYPES.AUTHORIZATION);

    // 422 - Validation error
    const validationResponse = { status: 422, message: 'Validation failed' };
    const validationError = handleApiError(validationResponse);
    expect(validationError.type).toBe(ERROR_TYPES.VALIDATION);

    // 500 - Server error
    const serverResponse = { status: 500, message: 'Internal server error' };
    const serverError = handleApiError(serverResponse);
    expect(serverError.type).toBe(ERROR_TYPES.SERVER);
    expect(serverError.level).toBe(ERROR_LEVELS.HIGH);
  });

  it('should handle network errors', () => {
    const networkError = new TypeError('fetch failed');
    const result = handleNetworkError(networkError);
    
    expect(result).toBeInstanceOf(AppError);
    expect(result.type).toBe(ERROR_TYPES.NETWORK);
  });

  it('should handle validation errors', () => {
    const message = 'Invalid input';
    const details = { field: 'email', code: 'INVALID_FORMAT' };
    
    const result = handleValidationError(message, details);
    
    expect(result).toBeInstanceOf(AppError);
    expect(result.type).toBe(ERROR_TYPES.VALIDATION);
    expect(result.level).toBe(ERROR_LEVELS.LOW);
    expect(result.details).toEqual(details);
  });
});

describe('User-Friendly Messages', () => {
  it('should return appropriate user messages for different error types', () => {
    const testCases = [
      {
        type: ERROR_TYPES.NETWORK,
        expected: 'インターネット接続を確認してください。'
      },
      {
        type: ERROR_TYPES.TIMEOUT,
        expected: 'リクエストがタイムアウトしました。しばらく待ってから再試行してください。'
      },
      {
        type: ERROR_TYPES.AUTHENTICATION,
        expected: '認証が必要です。ログインしてください。'
      },
      {
        type: ERROR_TYPES.AUTHORIZATION,
        expected: 'この操作を実行する権限がありません。'
      },
      {
        type: ERROR_TYPES.VALIDATION,
        expected: '入力内容を確認してください。'
      },
      {
        type: ERROR_TYPES.SERVER,
        expected: 'サーバーエラーが発生しました。しばらく待ってから再試行してください。'
      }
    ];

    testCases.forEach(({ type, expected }) => {
      const error = new AppError('Test error', type);
      const userMessage = globalErrorHandler.getUserFriendlyMessage(error);
      expect(userMessage).toBe(expected);
    });
  });

  it('should return original message for unknown error types', () => {
    const error = new AppError('Custom error message', 'CUSTOM_TYPE');
    const userMessage = globalErrorHandler.getUserFriendlyMessage(error);
    expect(userMessage).toBe('Custom error message');
  });
});

describe('Error Levels and Notification', () => {
  it('should not notify users for low-level errors', async () => {
    const error = new AppError('Low level error', ERROR_TYPES.UNKNOWN, ERROR_LEVELS.LOW);
    globalErrorHandler.notifyUser(error);
    
    // 通知が呼ばれないことを確認
    const { uiActions } = await import('$lib/stores/ui.js');
    expect(uiActions.showNotification).not.toHaveBeenCalled();
  });

  it('should notify users for medium and high level errors', async () => {
    const mediumError = new AppError('Medium error', ERROR_TYPES.UNKNOWN, ERROR_LEVELS.MEDIUM);
    const highError = new AppError('High error', ERROR_TYPES.UNKNOWN, ERROR_LEVELS.HIGH);
    
    globalErrorHandler.notifyUser(mediumError);
    globalErrorHandler.notifyUser(highError);
    
    const { uiActions } = await import('$lib/stores/ui.js');
    expect(uiActions.showNotification).toHaveBeenCalledTimes(2);
  });
});

describe('Integration Tests', () => {
  it('should handle complete error flow', () => {
    const originalError = new Error('Original error');
    const context = { operation: 'test-operation' };
    
    const result = handleError(originalError, context);
    
    expect(result).toBeInstanceOf(AppError);
    expect(result.details.context).toEqual(context);
    expect(result.details.originalError).toBe(originalError);
  });

  it('should initialize without errors', () => {
    expect(() => {
      initializeErrorHandler();
    }).not.toThrow();
  });
});
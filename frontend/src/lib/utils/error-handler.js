// グローバルエラーハンドラーとエラー境界システム
import { uiActions } from '$lib/stores/ui.js';

// エラータイプの定義
export const ERROR_TYPES = {
  NETWORK: 'NETWORK_ERROR',
  API: 'API_ERROR',
  VALIDATION: 'VALIDATION_ERROR',
  AUTHENTICATION: 'AUTHENTICATION_ERROR',
  AUTHORIZATION: 'AUTHORIZATION_ERROR',
  TIMEOUT: 'TIMEOUT_ERROR',
  PARSE: 'PARSE_ERROR',
  UNKNOWN: 'UNKNOWN_ERROR',
  CLIENT: 'CLIENT_ERROR',
  SERVER: 'SERVER_ERROR'
};

// エラーレベルの定義
export const ERROR_LEVELS = {
  LOW: 'low',
  MEDIUM: 'medium',
  HIGH: 'high',
  CRITICAL: 'critical'
};

// エラー情報を標準化するクラス
export class AppError extends Error {
  constructor(message, type = ERROR_TYPES.UNKNOWN, level = ERROR_LEVELS.MEDIUM, details = null) {
    super(message);
    this.name = 'AppError';
    this.type = type;
    this.level = level;
    this.details = details;
    this.timestamp = new Date().toISOString();
    this.userAgent = typeof navigator !== 'undefined' ? navigator.userAgent : 'Unknown';
  }

  // エラーをJSON形式で出力
  toJSON() {
    return {
      name: this.name,
      message: this.message,
      type: this.type,
      level: this.level,
      details: this.details,
      timestamp: this.timestamp,
      userAgent: this.userAgent,
      stack: this.stack
    };
  }
}

// グローバルエラーハンドラークラス
class GlobalErrorHandler {
  constructor() {
    this.errorListeners = [];
    this.retryAttempts = new Map();
    this.maxRetryAttempts = 3;
    this.retryDelay = 1000; // 1秒
    this.isInitialized = false;
  }

  // 初期化
  initialize() {
    if (this.isInitialized || typeof window === 'undefined') {
      return;
    }

    // 未処理のJavaScriptエラーをキャッチ
    window.addEventListener('error', this.handleGlobalError.bind(this));
    
    // 未処理のPromise拒否をキャッチ
    window.addEventListener('unhandledrejection', this.handleUnhandledRejection.bind(this));
    
    // 認証エラーイベントをリッスン
    window.addEventListener('auth:unauthorized', this.handleAuthError.bind(this));
    
    this.isInitialized = true;
    console.log('[ErrorHandler] Global error handler initialized');
  }

  // エラーリスナーを追加
  addErrorListener(listener) {
    this.errorListeners.push(listener);
  }

  // エラーリスナーを削除
  removeErrorListener(listener) {
    const index = this.errorListeners.indexOf(listener);
    if (index > -1) {
      this.errorListeners.splice(index, 1);
    }
  }

  // グローバルJavaScriptエラーハンドラー
  handleGlobalError(event) {
    const error = new AppError(
      event.message || 'JavaScript実行エラーが発生しました',
      ERROR_TYPES.CLIENT,
      ERROR_LEVELS.HIGH,
      {
        filename: event.filename,
        lineno: event.lineno,
        colno: event.colno,
        error: event.error
      }
    );

    this.handleError(error);
  }

  // 未処理のPromise拒否ハンドラー
  handleUnhandledRejection(event) {
    const error = new AppError(
      event.reason?.message || 'Promise処理でエラーが発生しました',
      ERROR_TYPES.CLIENT,
      ERROR_LEVELS.HIGH,
      {
        reason: event.reason,
        promise: event.promise
      }
    );

    this.handleError(error);
    event.preventDefault(); // デフォルトのエラー表示を防ぐ
  }

  // 認証エラーハンドラー
  handleAuthError(event) {
    const error = new AppError(
      '認証が無効になりました。再度ログインしてください。',
      ERROR_TYPES.AUTHENTICATION,
      ERROR_LEVELS.MEDIUM
    );

    this.handleError(error);
  }

  // メインエラーハンドラー
  handleError(error, context = {}) {
    // AppErrorでない場合は変換
    if (!(error instanceof AppError)) {
      error = this.normalizeError(error);
    }

    // コンテキスト情報を追加
    if (Object.keys(context).length > 0) {
      error.details = { ...error.details, context };
    }

    // エラーをログ出力
    this.logError(error);

    // エラーリスナーに通知
    this.notifyErrorListeners(error);

    // ユーザーに通知
    this.notifyUser(error);

    // 重要なエラーの場合は追加処理
    if (error.level === ERROR_LEVELS.CRITICAL) {
      this.handleCriticalError(error);
    }

    return error;
  }

  // エラーの正規化
  normalizeError(error) {
    if (error instanceof AppError) {
      return error;
    }

    let type = ERROR_TYPES.UNKNOWN;
    let level = ERROR_LEVELS.MEDIUM;
    let message = error.message || 'エラーが発生しました';

    // エラータイプの判定
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      type = ERROR_TYPES.NETWORK;
      message = 'ネットワークエラーが発生しました';
    } else if (error.name === 'AbortError') {
      type = ERROR_TYPES.TIMEOUT;
      message = 'リクエストがタイムアウトしました';
    } else if (error.name === 'SyntaxError') {
      type = ERROR_TYPES.PARSE;
      message = 'データの解析に失敗しました';
    }

    return new AppError(message, type, level, {
      originalError: error,
      stack: error.stack
    });
  }

  // エラーログ出力
  logError(error) {
    const logData = {
      timestamp: error.timestamp,
      type: error.type,
      level: error.level,
      message: error.message,
      details: error.details,
      stack: error.stack
    };

    // レベルに応じてログ出力
    switch (error.level) {
      case ERROR_LEVELS.LOW:
        console.info('[ErrorHandler]', logData);
        break;
      case ERROR_LEVELS.MEDIUM:
        console.warn('[ErrorHandler]', logData);
        break;
      case ERROR_LEVELS.HIGH:
      case ERROR_LEVELS.CRITICAL:
        console.error('[ErrorHandler]', logData);
        break;
      default:
        console.log('[ErrorHandler]', logData);
    }
  }

  // エラーリスナーに通知
  notifyErrorListeners(error) {
    this.errorListeners.forEach(listener => {
      try {
        listener(error);
      } catch (listenerError) {
        console.error('[ErrorHandler] Error in error listener:', listenerError);
      }
    });
  }

  // ユーザーに通知
  notifyUser(error) {
    // 低レベルエラーは通知しない
    if (error.level === ERROR_LEVELS.LOW) {
      return;
    }

    // 通知タイプの決定
    let notificationType = 'error';
    if (error.level === ERROR_LEVELS.MEDIUM) {
      notificationType = 'warning';
    }

    // ユーザーフレンドリーなメッセージに変換
    const userMessage = this.getUserFriendlyMessage(error);

    // 通知を表示
    uiActions.showNotification(userMessage, notificationType, 8000);
  }

  // ユーザーフレンドリーなメッセージに変換
  getUserFriendlyMessage(error) {
    switch (error.type) {
      case ERROR_TYPES.NETWORK:
        return 'インターネット接続を確認してください。';
      case ERROR_TYPES.TIMEOUT:
        return 'リクエストがタイムアウトしました。しばらく待ってから再試行してください。';
      case ERROR_TYPES.AUTHENTICATION:
        return '認証が必要です。ログインしてください。';
      case ERROR_TYPES.AUTHORIZATION:
        return 'この操作を実行する権限がありません。';
      case ERROR_TYPES.VALIDATION:
        return '入力内容を確認してください。';
      case ERROR_TYPES.SERVER:
        return 'サーバーエラーが発生しました。しばらく待ってから再試行してください。';
      default:
        return error.message || 'エラーが発生しました。';
    }
  }

  // 重要なエラーの処理
  handleCriticalError(error) {
    // 重要なエラーの場合は追加のログ送信やアラートなどを実装
    console.error('[ErrorHandler] CRITICAL ERROR:', error.toJSON());
    
    // 必要に応じてエラー報告サービスに送信
    // this.reportError(error);
  }

  // 再試行機能
  async retry(operation, key, maxAttempts = this.maxRetryAttempts) {
    const attempts = this.retryAttempts.get(key) || 0;
    
    try {
      const result = await operation();
      // 成功した場合は再試行カウンターをリセット
      this.retryAttempts.delete(key);
      return result;
    } catch (error) {
      if (attempts < maxAttempts) {
        this.retryAttempts.set(key, attempts + 1);
        
        // 指数バックオフで待機
        const delay = this.retryDelay * Math.pow(2, attempts);
        await new Promise(resolve => setTimeout(resolve, delay));
        
        console.log(`[ErrorHandler] Retrying operation "${key}" (attempt ${attempts + 1}/${maxAttempts})`);
        return this.retry(operation, key, maxAttempts);
      } else {
        // 最大試行回数に達した場合
        this.retryAttempts.delete(key);
        throw new AppError(
          `操作が${maxAttempts}回失敗しました: ${error.message}`,
          ERROR_TYPES.UNKNOWN,
          ERROR_LEVELS.HIGH,
          { originalError: error, attempts: maxAttempts }
        );
      }
    }
  }

  // エラー境界のリセット
  resetErrorBoundary(key) {
    this.retryAttempts.delete(key);
  }

  // クリーンアップ
  cleanup() {
    if (typeof window !== 'undefined') {
      window.removeEventListener('error', this.handleGlobalError.bind(this));
      window.removeEventListener('unhandledrejection', this.handleUnhandledRejection.bind(this));
      window.removeEventListener('auth:unauthorized', this.handleAuthError.bind(this));
    }
    
    this.errorListeners = [];
    this.retryAttempts.clear();
    this.isInitialized = false;
  }
}

// グローバルインスタンス
export const globalErrorHandler = new GlobalErrorHandler();

// 便利な関数をエクスポート
export const handleError = (error, context) => globalErrorHandler.handleError(error, context);
export const retry = (operation, key, maxAttempts) => globalErrorHandler.retry(operation, key, maxAttempts);
export const addErrorListener = (listener) => globalErrorHandler.addErrorListener(listener);
export const removeErrorListener = (listener) => globalErrorHandler.removeErrorListener(listener);

// 初期化関数
export const initializeErrorHandler = () => {
  globalErrorHandler.initialize();
};

// APIエラー専用のヘルパー関数
export const handleApiError = (response, context = {}) => {
  let type = ERROR_TYPES.API;
  let level = ERROR_LEVELS.MEDIUM;

  // HTTPステータスコードに基づいてエラータイプを決定
  if (response.status >= 400 && response.status < 500) {
    type = ERROR_TYPES.CLIENT;
    if (response.status === 401) {
      type = ERROR_TYPES.AUTHENTICATION;
    } else if (response.status === 403) {
      type = ERROR_TYPES.AUTHORIZATION;
    } else if (response.status === 422) {
      type = ERROR_TYPES.VALIDATION;
    }
  } else if (response.status >= 500) {
    type = ERROR_TYPES.SERVER;
    level = ERROR_LEVELS.HIGH;
  }

  const error = new AppError(
    response.message || `HTTP ${response.status}: ${response.statusText}`,
    type,
    level,
    {
      status: response.status,
      statusText: response.statusText,
      error: response.error,
      details: response.details,
      ...context
    }
  );

  return globalErrorHandler.handleError(error, context);
};

// ネットワークエラー専用のヘルパー関数
export const handleNetworkError = (error, context = {}) => {
  const networkError = new AppError(
    'ネットワーク接続に問題があります',
    ERROR_TYPES.NETWORK,
    ERROR_LEVELS.MEDIUM,
    {
      originalError: error,
      ...context
    }
  );

  return globalErrorHandler.handleError(networkError, context);
};

// バリデーションエラー専用のヘルパー関数
export const handleValidationError = (message, details = {}) => {
  const validationError = new AppError(
    message,
    ERROR_TYPES.VALIDATION,
    ERROR_LEVELS.LOW,
    details
  );

  return globalErrorHandler.handleError(validationError);
};
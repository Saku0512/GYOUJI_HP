// 統一エラーレスポンスハンドラー
// フィールド別エラー情報の詳細化と多言語対応

import { ErrorCode } from '../api/types.js';

/**
 * エラーメッセージの多言語対応
 */
const ERROR_MESSAGES = {
  ja: {
    // 認証関連エラー
    [ErrorCode.AUTH_INVALID_CREDENTIALS]: 'ユーザー名またはパスワードが正しくありません',
    [ErrorCode.AUTH_TOKEN_EXPIRED]: 'セッションが期限切れです。再度ログインしてください',
    [ErrorCode.AUTH_TOKEN_INVALID]: '認証トークンが無効です',
    [ErrorCode.AUTH_UNAUTHORIZED]: '認証が必要です',
    [ErrorCode.AUTH_FORBIDDEN]: 'この操作を実行する権限がありません',

    // バリデーション関連エラー
    [ErrorCode.VALIDATION_REQUIRED_FIELD]: '{field}は必須です',
    [ErrorCode.VALIDATION_INVALID_FORMAT]: '{field}の形式が正しくありません',
    [ErrorCode.VALIDATION_OUT_OF_RANGE]: '{field}の値が範囲外です',
    [ErrorCode.VALIDATION_DUPLICATE_VALUE]: '{field}は既に使用されています',

    // リソース関連エラー
    [ErrorCode.RESOURCE_NOT_FOUND]: '指定されたリソースが見つかりません',
    [ErrorCode.RESOURCE_ALREADY_EXISTS]: 'リソースは既に存在します',
    [ErrorCode.RESOURCE_CONFLICT]: 'リソースの競合が発生しました',

    // ビジネスロジック関連エラー
    [ErrorCode.BUSINESS_TOURNAMENT_COMPLETED]: 'トーナメントは既に完了しています',
    [ErrorCode.BUSINESS_MATCH_ALREADY_COMPLETED]: '試合は既に完了しています',
    [ErrorCode.BUSINESS_INVALID_MATCH_RESULT]: '無効な試合結果です',

    // システム関連エラー
    [ErrorCode.SYSTEM_DATABASE_ERROR]: 'データベースエラーが発生しました',
    [ErrorCode.SYSTEM_NETWORK_ERROR]: 'ネットワークエラーが発生しました',
    [ErrorCode.SYSTEM_TIMEOUT]: 'タイムアウトが発生しました',
    [ErrorCode.SYSTEM_UNKNOWN_ERROR]: '予期しないエラーが発生しました',

    // フィールド名の日本語対応
    fields: {
      username: 'ユーザー名',
      password: 'パスワード',
      email: 'メールアドレス',
      score1: 'チーム1のスコア',
      score2: 'チーム2のスコア',
      winner: '勝者',
      sport: 'スポーツ',
      format: 'フォーマット',
      status: 'ステータス',
      team1: 'チーム1',
      team2: 'チーム2',
      round: 'ラウンド',
      scheduledAt: '予定日時',
      completedAt: '完了日時'
    },

    // 一般的なメッセージ
    general: {
      validationFailed: '入力内容に問題があります',
      serverError: 'サーバーエラーが発生しました',
      networkError: 'ネットワークに接続できません',
      unknownError: '不明なエラーが発生しました',
      tryAgain: 'もう一度お試しください',
      contactSupport: 'サポートにお問い合わせください'
    }
  },

  en: {
    // Authentication errors
    [ErrorCode.AUTH_INVALID_CREDENTIALS]: 'Invalid username or password',
    [ErrorCode.AUTH_TOKEN_EXPIRED]: 'Session expired. Please login again',
    [ErrorCode.AUTH_TOKEN_INVALID]: 'Invalid authentication token',
    [ErrorCode.AUTH_UNAUTHORIZED]: 'Authentication required',
    [ErrorCode.AUTH_FORBIDDEN]: 'You do not have permission to perform this action',

    // Validation errors
    [ErrorCode.VALIDATION_REQUIRED_FIELD]: '{field} is required',
    [ErrorCode.VALIDATION_INVALID_FORMAT]: '{field} format is invalid',
    [ErrorCode.VALIDATION_OUT_OF_RANGE]: '{field} value is out of range',
    [ErrorCode.VALIDATION_DUPLICATE_VALUE]: '{field} is already in use',

    // Resource errors
    [ErrorCode.RESOURCE_NOT_FOUND]: 'Resource not found',
    [ErrorCode.RESOURCE_ALREADY_EXISTS]: 'Resource already exists',
    [ErrorCode.RESOURCE_CONFLICT]: 'Resource conflict occurred',

    // Business logic errors
    [ErrorCode.BUSINESS_TOURNAMENT_COMPLETED]: 'Tournament is already completed',
    [ErrorCode.BUSINESS_MATCH_ALREADY_COMPLETED]: 'Match is already completed',
    [ErrorCode.BUSINESS_INVALID_MATCH_RESULT]: 'Invalid match result',

    // System errors
    [ErrorCode.SYSTEM_DATABASE_ERROR]: 'Database error occurred',
    [ErrorCode.SYSTEM_NETWORK_ERROR]: 'Network error occurred',
    [ErrorCode.SYSTEM_TIMEOUT]: 'Timeout occurred',
    [ErrorCode.SYSTEM_UNKNOWN_ERROR]: 'Unexpected error occurred',

    // Field names in English
    fields: {
      username: 'Username',
      password: 'Password',
      email: 'Email',
      score1: 'Team 1 Score',
      score2: 'Team 2 Score',
      winner: 'Winner',
      sport: 'Sport',
      format: 'Format',
      status: 'Status',
      team1: 'Team 1',
      team2: 'Team 2',
      round: 'Round',
      scheduledAt: 'Scheduled At',
      completedAt: 'Completed At'
    },

    // General messages
    general: {
      validationFailed: 'Input validation failed',
      serverError: 'Server error occurred',
      networkError: 'Cannot connect to network',
      unknownError: 'Unknown error occurred',
      tryAgain: 'Please try again',
      contactSupport: 'Please contact support'
    }
  }
};

/**
 * 詳細エラー情報クラス
 */
export class DetailedErrorInfo {
  constructor(options = {}) {
    this.field = options.field || null;
    this.code = options.code || null;
    this.message = options.message || '';
    this.value = options.value !== undefined ? options.value : null;
    this.rule = options.rule || null;
    this.severity = options.severity || 'error'; // 'error', 'warning', 'info'
    this.suggestions = options.suggestions || [];
    this.context = options.context || {};
  }

  /**
   * エラー情報をJSON形式で取得
   */
  toJSON() {
    return {
      field: this.field,
      code: this.code,
      message: this.message,
      value: this.value,
      rule: this.rule,
      severity: this.severity,
      suggestions: this.suggestions,
      context: this.context
    };
  }

  /**
   * ユーザーフレンドリーなメッセージを取得
   */
  getUserMessage(language = 'ja') {
    return this.message || ERROR_MESSAGES[language]?.[this.code] || 
           ERROR_MESSAGES[language]?.general?.unknownError || 
           'An error occurred';
  }
}

/**
 * 統一エラーレスポンスハンドラー
 */
export class UnifiedErrorResponseHandler {
  constructor(options = {}) {
    this.defaultLanguage = options.defaultLanguage || 'ja';
    this.enableLogging = options.enableLogging !== false;
    this.logLevel = options.logLevel || 'error';
  }

  /**
   * APIレスポンスエラーを処理
   */
  handleAPIError(response, context = {}) {
    const errorInfo = this.parseAPIError(response);
    
    if (this.enableLogging) {
      this.logError(errorInfo, context);
    }

    return this.formatErrorResponse(errorInfo, context);
  }

  /**
   * バリデーションエラーを処理
   */
  handleValidationError(validationResult, context = {}) {
    const errorInfo = this.parseValidationError(validationResult);
    
    if (this.enableLogging) {
      this.logError(errorInfo, context);
    }

    return this.formatErrorResponse(errorInfo, context);
  }

  /**
   * ネットワークエラーを処理
   */
  handleNetworkError(error, context = {}) {
    const errorInfo = this.parseNetworkError(error);
    
    if (this.enableLogging) {
      this.logError(errorInfo, context);
    }

    return this.formatErrorResponse(errorInfo, context);
  }

  /**
   * APIエラーレスポンスを解析
   */
  parseAPIError(response) {
    const errors = [];

    if (response.error) {
      // 単一エラーの場合
      errors.push(new DetailedErrorInfo({
        code: response.error,
        message: response.message,
        severity: this.getErrorSeverity(response.code),
        context: {
          httpStatus: response.code,
          timestamp: response.timestamp,
          requestId: response.request_id
        }
      }));
    } else if (response.errors && typeof response.errors === 'object') {
      // フィールド別エラーの場合
      Object.entries(response.errors).forEach(([field, message]) => {
        errors.push(new DetailedErrorInfo({
          field,
          message: Array.isArray(message) ? message[0] : message,
          code: this.inferErrorCode(message, field),
          severity: 'error',
          context: {
            httpStatus: response.code,
            timestamp: response.timestamp,
            requestId: response.request_id
          }
        }));
      });
    } else {
      // 不明なエラー形式
      errors.push(new DetailedErrorInfo({
        code: ErrorCode.SYSTEM_UNKNOWN_ERROR,
        message: response.message || 'Unknown error occurred',
        severity: 'error',
        context: {
          httpStatus: response.code,
          timestamp: response.timestamp,
          requestId: response.request_id
        }
      }));
    }

    return errors;
  }

  /**
   * バリデーションエラーを解析
   */
  parseValidationError(validationResult) {
    const errors = [];

    if (validationResult.errors) {
      Object.entries(validationResult.errors).forEach(([field, message]) => {
        errors.push(new DetailedErrorInfo({
          field,
          message,
          code: ErrorCode.VALIDATION_INVALID_FORMAT,
          severity: 'error',
          suggestions: this.generateSuggestions(field, message),
          context: {
            source: 'client_validation'
          }
        }));
      });
    }

    return errors;
  }

  /**
   * ネットワークエラーを解析
   */
  parseNetworkError(error) {
    const errors = [];

    let code = ErrorCode.SYSTEM_NETWORK_ERROR;
    let message = 'Network error occurred';
    let suggestions = ['インターネット接続を確認してください', 'しばらく時間をおいて再試行してください'];

    if (error.name === 'AbortError') {
      code = ErrorCode.SYSTEM_TIMEOUT;
      message = 'Request timeout';
      suggestions = ['処理に時間がかかっています', 'しばらく待ってから再試行してください'];
    } else if (error.message && error.message.includes('fetch')) {
      message = 'Failed to connect to server';
      suggestions = ['サーバーが利用できない可能性があります', 'しばらく時間をおいて再試行してください'];
    }

    errors.push(new DetailedErrorInfo({
      code,
      message,
      severity: 'error',
      suggestions,
      context: {
        errorName: error.name,
        errorMessage: error.message,
        source: 'network'
      }
    }));

    return errors;
  }

  /**
   * エラーレスポンスをフォーマット
   */
  formatErrorResponse(errors, context = {}) {
    const language = context.language || this.defaultLanguage;
    
    return {
      hasErrors: errors.length > 0,
      errorCount: errors.length,
      errors: errors.map(error => ({
        ...error.toJSON(),
        localizedMessage: this.localizeMessage(error, language),
        userMessage: error.getUserMessage(language)
      })),
      summary: this.generateErrorSummary(errors, language),
      suggestions: this.consolidateSuggestions(errors),
      context: {
        language,
        timestamp: new Date().toISOString(),
        ...context
      }
    };
  }

  /**
   * エラーメッセージをローカライズ
   */
  localizeMessage(error, language) {
    const messages = ERROR_MESSAGES[language];
    if (!messages) return error.message;

    let message = messages[error.code] || error.message;

    // フィールド名の置換
    if (error.field && messages.fields[error.field]) {
      message = message.replace('{field}', messages.fields[error.field]);
    }

    return message;
  }

  /**
   * エラーサマリーを生成
   */
  generateErrorSummary(errors, language) {
    const messages = ERROR_MESSAGES[language];
    if (!messages) return 'Errors occurred';

    if (errors.length === 0) {
      return '';
    }

    if (errors.length === 1) {
      return errors[0].getUserMessage(language);
    }

    const fieldErrors = errors.filter(e => e.field);
    const generalErrors = errors.filter(e => !e.field);

    if (fieldErrors.length > 0 && generalErrors.length === 0) {
      return messages.general.validationFailed;
    }

    if (generalErrors.length > 0) {
      return generalErrors[0].getUserMessage(language);
    }

    return `${errors.length}個のエラーが発生しました`;
  }

  /**
   * 提案を統合
   */
  consolidateSuggestions(errors) {
    const allSuggestions = errors.flatMap(error => error.suggestions);
    return [...new Set(allSuggestions)]; // 重複を除去
  }

  /**
   * エラーコードを推測
   */
  inferErrorCode(message, field) {
    if (typeof message !== 'string') return ErrorCode.VALIDATION_INVALID_FORMAT;

    const lowerMessage = message.toLowerCase();

    if (lowerMessage.includes('必須') || lowerMessage.includes('required')) {
      return ErrorCode.VALIDATION_REQUIRED_FIELD;
    }
    if (lowerMessage.includes('形式') || lowerMessage.includes('format')) {
      return ErrorCode.VALIDATION_INVALID_FORMAT;
    }
    if (lowerMessage.includes('範囲') || lowerMessage.includes('range') || 
        lowerMessage.includes('以上') || lowerMessage.includes('以下')) {
      return ErrorCode.VALIDATION_OUT_OF_RANGE;
    }
    if (lowerMessage.includes('重複') || lowerMessage.includes('duplicate') || 
        lowerMessage.includes('既に') || lowerMessage.includes('already')) {
      return ErrorCode.VALIDATION_DUPLICATE_VALUE;
    }

    return ErrorCode.VALIDATION_INVALID_FORMAT;
  }

  /**
   * エラーの重要度を取得
   */
  getErrorSeverity(httpStatus) {
    if (httpStatus >= 500) return 'error';
    if (httpStatus >= 400) return 'warning';
    return 'info';
  }

  /**
   * 提案を生成
   */
  generateSuggestions(field, message) {
    const suggestions = [];

    if (typeof message !== 'string') return suggestions;

    const lowerMessage = message.toLowerCase();

    if (lowerMessage.includes('必須') || lowerMessage.includes('required')) {
      suggestions.push(`${field}を入力してください`);
    }
    if (lowerMessage.includes('文字数') || lowerMessage.includes('length')) {
      suggestions.push('適切な文字数で入力してください');
    }
    if (lowerMessage.includes('形式') || lowerMessage.includes('format')) {
      suggestions.push('正しい形式で入力してください');
    }
    if (lowerMessage.includes('範囲') || lowerMessage.includes('range')) {
      suggestions.push('有効な範囲内の値を入力してください');
    }

    // フィールド固有の提案
    switch (field) {
      case 'username':
        suggestions.push('英数字のみ使用してください');
        break;
      case 'password':
        suggestions.push('8文字以上で入力してください');
        break;
      case 'email':
        suggestions.push('有効なメールアドレスを入力してください');
        break;
      case 'score1':
      case 'score2':
        suggestions.push('0以上の数値を入力してください');
        break;
    }

    return suggestions;
  }

  /**
   * エラーをログに記録
   */
  logError(errors, context) {
    if (!this.enableLogging) return;

    const logData = {
      timestamp: new Date().toISOString(),
      errors: errors.map(e => e.toJSON()),
      context,
      level: this.logLevel
    };

    console.error('Unified Error Handler:', logData);

    // 外部ログサービスへの送信（オプション）
    if (typeof window !== 'undefined' && window.errorLogger) {
      window.errorLogger.log(logData);
    }
  }
}

/**
 * デフォルトのエラーハンドラーインスタンス
 */
export const defaultErrorHandler = new UnifiedErrorResponseHandler();

/**
 * 便利な関数群
 */

/**
 * APIエラーを処理
 */
export function handleAPIError(response, context = {}) {
  return defaultErrorHandler.handleAPIError(response, context);
}

/**
 * バリデーションエラーを処理
 */
export function handleValidationError(validationResult, context = {}) {
  return defaultErrorHandler.handleValidationError(validationResult, context);
}

/**
 * ネットワークエラーを処理
 */
export function handleNetworkError(error, context = {}) {
  return defaultErrorHandler.handleNetworkError(error, context);
}

/**
 * エラーメッセージをローカライズ
 */
export function localizeErrorMessage(code, field = null, language = 'ja') {
  const messages = ERROR_MESSAGES[language];
  if (!messages) return 'Error occurred';

  let message = messages[code] || messages.general.unknownError;

  if (field && messages.fields[field]) {
    message = message.replace('{field}', messages.fields[field]);
  }

  return message;
}

/**
 * フィールド名をローカライズ
 */
export function localizeFieldName(field, language = 'ja') {
  const messages = ERROR_MESSAGES[language];
  return messages?.fields?.[field] || field;
}

/**
 * エラーコードから重要度を取得
 */
export function getErrorSeverityFromCode(code) {
  if (code.startsWith('SYSTEM_')) return 'error';
  if (code.startsWith('AUTH_')) return 'warning';
  if (code.startsWith('VALIDATION_')) return 'info';
  return 'error';
}
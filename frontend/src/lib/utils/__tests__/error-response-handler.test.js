// 統一エラーレスポンスハンドラーのテスト

import { describe, it, expect, beforeEach } from 'vitest';
import {
  DetailedErrorInfo,
  UnifiedErrorResponseHandler,
  handleAPIError,
  handleValidationError,
  handleNetworkError,
  localizeErrorMessage,
  localizeFieldName,
  getErrorSeverityFromCode
} from '../error-response-handler.js';
import { ErrorCode } from '../../api/types.js';

describe('DetailedErrorInfo', () => {
  it('should create error info with all properties', () => {
    const error = new DetailedErrorInfo({
      field: 'username',
      code: ErrorCode.VALIDATION_REQUIRED_FIELD,
      message: 'ユーザー名は必須です',
      value: '',
      rule: 'required',
      severity: 'error',
      suggestions: ['ユーザー名を入力してください'],
      context: { source: 'client' }
    });

    expect(error.field).toBe('username');
    expect(error.code).toBe(ErrorCode.VALIDATION_REQUIRED_FIELD);
    expect(error.message).toBe('ユーザー名は必須です');
    expect(error.value).toBe('');
    expect(error.rule).toBe('required');
    expect(error.severity).toBe('error');
    expect(error.suggestions).toEqual(['ユーザー名を入力してください']);
    expect(error.context).toEqual({ source: 'client' });
  });

  it('should convert to JSON', () => {
    const error = new DetailedErrorInfo({
      field: 'username',
      code: ErrorCode.VALIDATION_REQUIRED_FIELD,
      message: 'ユーザー名は必須です'
    });

    const json = error.toJSON();
    expect(json).toEqual({
      field: 'username',
      code: ErrorCode.VALIDATION_REQUIRED_FIELD,
      message: 'ユーザー名は必須です',
      value: null,
      rule: null,
      severity: 'error',
      suggestions: [],
      context: {}
    });
  });

  it('should get user message in different languages', () => {
    const error = new DetailedErrorInfo({
      code: ErrorCode.VALIDATION_REQUIRED_FIELD,
      message: 'Custom message'
    });

    expect(error.getUserMessage('ja')).toBe('Custom message');
    expect(error.getUserMessage('en')).toBe('Custom message');
  });
});

describe('UnifiedErrorResponseHandler', () => {
  let handler;

  beforeEach(() => {
    handler = new UnifiedErrorResponseHandler({
      defaultLanguage: 'ja',
      enableLogging: false
    });
  });

  describe('parseAPIError', () => {
    it('should parse single error response', () => {
      const response = {
        error: ErrorCode.AUTH_INVALID_CREDENTIALS,
        message: 'Invalid credentials',
        code: 401,
        timestamp: '2024-01-01T00:00:00Z',
        request_id: 'req_123'
      };

      const errors = handler.parseAPIError(response);
      
      expect(errors).toHaveLength(1);
      expect(errors[0].code).toBe(ErrorCode.AUTH_INVALID_CREDENTIALS);
      expect(errors[0].message).toBe('Invalid credentials');
      expect(errors[0].context.httpStatus).toBe(401);
    });

    it('should parse field errors response', () => {
      const response = {
        errors: {
          username: 'ユーザー名は必須です',
          password: 'パスワードは8文字以上必要です'
        },
        code: 400,
        timestamp: '2024-01-01T00:00:00Z',
        request_id: 'req_123'
      };

      const errors = handler.parseAPIError(response);
      
      expect(errors).toHaveLength(2);
      expect(errors[0].field).toBe('username');
      expect(errors[0].message).toBe('ユーザー名は必須です');
      expect(errors[1].field).toBe('password');
      expect(errors[1].message).toBe('パスワードは8文字以上必要です');
    });

    it('should handle array error messages', () => {
      const response = {
        errors: {
          username: ['ユーザー名は必須です', '英数字のみ使用可能です']
        },
        code: 400
      };

      const errors = handler.parseAPIError(response);
      
      expect(errors).toHaveLength(1);
      expect(errors[0].message).toBe('ユーザー名は必須です'); // 最初のエラーのみ
    });

    it('should handle unknown error format', () => {
      const response = {
        message: 'Something went wrong',
        code: 500
      };

      const errors = handler.parseAPIError(response);
      
      expect(errors).toHaveLength(1);
      expect(errors[0].code).toBe(ErrorCode.SYSTEM_UNKNOWN_ERROR);
      expect(errors[0].message).toBe('Something went wrong');
    });
  });

  describe('parseValidationError', () => {
    it('should parse validation result errors', () => {
      const validationResult = {
        isValid: false,
        errors: {
          username: 'ユーザー名は必須です',
          password: 'パスワードが短すぎます'
        }
      };

      const errors = handler.parseValidationError(validationResult);
      
      expect(errors).toHaveLength(2);
      expect(errors[0].field).toBe('username');
      expect(errors[0].message).toBe('ユーザー名は必須です');
      expect(errors[0].code).toBe(ErrorCode.VALIDATION_INVALID_FORMAT);
      expect(errors[0].context.source).toBe('client_validation');
    });
  });

  describe('parseNetworkError', () => {
    it('should parse timeout error', () => {
      const error = new Error('Request timeout');
      error.name = 'AbortError';

      const errors = handler.parseNetworkError(error);
      
      expect(errors).toHaveLength(1);
      expect(errors[0].code).toBe(ErrorCode.SYSTEM_TIMEOUT);
      expect(errors[0].message).toBe('Request timeout');
      expect(errors[0].suggestions).toContain('処理に時間がかかっています');
    });

    it('should parse fetch error', () => {
      const error = new Error('fetch failed');
      error.name = 'TypeError';

      const errors = handler.parseNetworkError(error);
      
      expect(errors).toHaveLength(1);
      expect(errors[0].code).toBe(ErrorCode.SYSTEM_NETWORK_ERROR);
      expect(errors[0].message).toBe('Failed to connect to server');
    });

    it('should parse generic network error', () => {
      const error = new Error('Network error');

      const errors = handler.parseNetworkError(error);
      
      expect(errors).toHaveLength(1);
      expect(errors[0].code).toBe(ErrorCode.SYSTEM_NETWORK_ERROR);
      expect(errors[0].message).toBe('Network error occurred');
    });
  });

  describe('formatErrorResponse', () => {
    it('should format error response with localization', () => {
      const errors = [
        new DetailedErrorInfo({
          field: 'username',
          code: ErrorCode.VALIDATION_REQUIRED_FIELD,
          message: 'Username is required'
        })
      ];

      const response = handler.formatErrorResponse(errors, { language: 'ja' });
      
      expect(response.hasErrors).toBe(true);
      expect(response.errorCount).toBe(1);
      expect(response.errors).toHaveLength(1);
      expect(response.errors[0].localizedMessage).toBe('ユーザー名は必須です');
      expect(response.context.language).toBe('ja');
    });

    it('should generate error summary', () => {
      const errors = [
        new DetailedErrorInfo({
          field: 'username',
          code: ErrorCode.VALIDATION_REQUIRED_FIELD,
          message: 'Username is required'
        }),
        new DetailedErrorInfo({
          field: 'password',
          code: ErrorCode.VALIDATION_OUT_OF_RANGE,
          message: 'Password too short'
        })
      ];

      const response = handler.formatErrorResponse(errors, { language: 'ja' });
      
      expect(response.summary).toBe('入力内容に問題があります');
    });
  });

  describe('inferErrorCode', () => {
    it('should infer required field error', () => {
      expect(handler.inferErrorCode('必須です', 'username')).toBe(ErrorCode.VALIDATION_REQUIRED_FIELD);
      expect(handler.inferErrorCode('is required', 'username')).toBe(ErrorCode.VALIDATION_REQUIRED_FIELD);
    });

    it('should infer format error', () => {
      expect(handler.inferErrorCode('形式が正しくありません', 'email')).toBe(ErrorCode.VALIDATION_INVALID_FORMAT);
      expect(handler.inferErrorCode('invalid format', 'email')).toBe(ErrorCode.VALIDATION_INVALID_FORMAT);
    });

    it('should infer range error', () => {
      expect(handler.inferErrorCode('範囲外です', 'score')).toBe(ErrorCode.VALIDATION_OUT_OF_RANGE);
      expect(handler.inferErrorCode('3文字以上', 'username')).toBe(ErrorCode.VALIDATION_OUT_OF_RANGE);
    });

    it('should infer duplicate error', () => {
      expect(handler.inferErrorCode('既に使用されています', 'username')).toBe(ErrorCode.VALIDATION_DUPLICATE_VALUE);
      expect(handler.inferErrorCode('already exists', 'username')).toBe(ErrorCode.VALIDATION_DUPLICATE_VALUE);
    });

    it('should default to invalid format', () => {
      expect(handler.inferErrorCode('unknown error', 'field')).toBe(ErrorCode.VALIDATION_INVALID_FORMAT);
    });
  });

  describe('generateSuggestions', () => {
    it('should generate field-specific suggestions', () => {
      const suggestions = handler.generateSuggestions('username', 'ユーザー名は必須です');
      
      expect(suggestions).toContain('usernameを入力してください');
      expect(suggestions).toContain('英数字のみ使用してください');
    });

    it('should generate password suggestions', () => {
      const suggestions = handler.generateSuggestions('password', 'パスワードが短すぎます');
      
      expect(suggestions).toContain('8文字以上で入力してください');
    });

    it('should generate email suggestions', () => {
      const suggestions = handler.generateSuggestions('email', '形式が正しくありません');
      
      expect(suggestions).toContain('有効なメールアドレスを入力してください');
    });
  });
});

describe('Convenience Functions', () => {
  describe('handleAPIError', () => {
    it('should handle API error response', () => {
      const response = {
        error: ErrorCode.AUTH_INVALID_CREDENTIALS,
        message: 'Invalid credentials',
        code: 401
      };

      const result = handleAPIError(response);
      
      expect(result.hasErrors).toBe(true);
      expect(result.errors).toHaveLength(1);
      expect(result.errors[0].code).toBe(ErrorCode.AUTH_INVALID_CREDENTIALS);
    });
  });

  describe('handleValidationError', () => {
    it('should handle validation error', () => {
      const validationResult = {
        isValid: false,
        errors: {
          username: 'ユーザー名は必須です'
        }
      };

      const result = handleValidationError(validationResult);
      
      expect(result.hasErrors).toBe(true);
      expect(result.errors).toHaveLength(1);
      expect(result.errors[0].field).toBe('username');
    });
  });

  describe('handleNetworkError', () => {
    it('should handle network error', () => {
      const error = new Error('Network failed');

      const result = handleNetworkError(error);
      
      expect(result.hasErrors).toBe(true);
      expect(result.errors).toHaveLength(1);
      expect(result.errors[0].code).toBe(ErrorCode.SYSTEM_NETWORK_ERROR);
    });
  });

  describe('localizeErrorMessage', () => {
    it('should localize error message in Japanese', () => {
      const message = localizeErrorMessage(ErrorCode.VALIDATION_REQUIRED_FIELD, 'username', 'ja');
      expect(message).toBe('ユーザー名は必須です');
    });

    it('should localize error message in English', () => {
      const message = localizeErrorMessage(ErrorCode.VALIDATION_REQUIRED_FIELD, 'username', 'en');
      expect(message).toBe('Username is required');
    });

    it('should handle unknown language', () => {
      const message = localizeErrorMessage(ErrorCode.VALIDATION_REQUIRED_FIELD, 'username', 'fr');
      expect(message).toBe('Error occurred');
    });

    it('should handle unknown error code', () => {
      const message = localizeErrorMessage('UNKNOWN_CODE', 'username', 'ja');
      expect(message).toBe('不明なエラーが発生しました');
    });
  });

  describe('localizeFieldName', () => {
    it('should localize field name in Japanese', () => {
      expect(localizeFieldName('username', 'ja')).toBe('ユーザー名');
      expect(localizeFieldName('password', 'ja')).toBe('パスワード');
      expect(localizeFieldName('email', 'ja')).toBe('メールアドレス');
    });

    it('should localize field name in English', () => {
      expect(localizeFieldName('username', 'en')).toBe('Username');
      expect(localizeFieldName('password', 'en')).toBe('Password');
      expect(localizeFieldName('email', 'en')).toBe('Email');
    });

    it('should return original field name for unknown fields', () => {
      expect(localizeFieldName('unknownField', 'ja')).toBe('unknownField');
    });

    it('should handle unknown language', () => {
      expect(localizeFieldName('username', 'fr')).toBe('username');
    });
  });

  describe('getErrorSeverityFromCode', () => {
    it('should return error for system codes', () => {
      expect(getErrorSeverityFromCode(ErrorCode.SYSTEM_DATABASE_ERROR)).toBe('error');
      expect(getErrorSeverityFromCode(ErrorCode.SYSTEM_NETWORK_ERROR)).toBe('error');
    });

    it('should return warning for auth codes', () => {
      expect(getErrorSeverityFromCode(ErrorCode.AUTH_UNAUTHORIZED)).toBe('warning');
      expect(getErrorSeverityFromCode(ErrorCode.AUTH_FORBIDDEN)).toBe('warning');
    });

    it('should return info for validation codes', () => {
      expect(getErrorSeverityFromCode(ErrorCode.VALIDATION_REQUIRED_FIELD)).toBe('info');
      expect(getErrorSeverityFromCode(ErrorCode.VALIDATION_INVALID_FORMAT)).toBe('info');
    });

    it('should default to error for unknown codes', () => {
      expect(getErrorSeverityFromCode('UNKNOWN_CODE')).toBe('error');
    });
  });
});
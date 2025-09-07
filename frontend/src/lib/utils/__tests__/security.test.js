// セキュリティユーティリティのテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  generateSecureRandomString,
  csrfTokenManager,
  addSecurityHeaders,
  enforceInputLimits,
  defaultRateLimiter,
  secureStorage,
  securityLogger
} from '../security.js';

// モック設定
const mockLocalStorage = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn()
};

const mockSessionStorage = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn()
};

const mockCrypto = {
  getRandomValues: vi.fn()
};

describe('セキュリティユーティリティ', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    
    // ブラウザ環境をシミュレート
    Object.defineProperty(window, 'localStorage', {
      value: mockLocalStorage,
      writable: true
    });
    
    Object.defineProperty(window, 'sessionStorage', {
      value: mockSessionStorage,
      writable: true
    });
    
    Object.defineProperty(window, 'crypto', {
      value: mockCrypto,
      writable: true
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('generateSecureRandomString', () => {
    it('指定した長さの文字列を生成する', () => {
      // crypto.getRandomValuesをモック
      mockCrypto.getRandomValues.mockImplementation((array) => {
        for (let i = 0; i < array.length; i++) {
          array[i] = i % 62; // 0-61の範囲で循環
        }
        return array;
      });

      const result = generateSecureRandomString(32);
      expect(result).toHaveLength(32);
      expect(typeof result).toBe('string');
    });

    it('デフォルトで32文字の文字列を生成する', () => {
      mockCrypto.getRandomValues.mockImplementation((array) => {
        for (let i = 0; i < array.length; i++) {
          array[i] = i % 62;
        }
        return array;
      });

      const result = generateSecureRandomString();
      expect(result).toHaveLength(32);
    });

    it('crypto.getRandomValuesが利用できない場合はフォールバックを使用', () => {
      // cryptoを無効にする
      Object.defineProperty(window, 'crypto', {
        value: undefined,
        writable: true
      });

      const result = generateSecureRandomString(16);
      expect(result).toHaveLength(16);
      expect(typeof result).toBe('string');
    });
  });

  describe('CSRFTokenManager', () => {
    beforeEach(() => {
      // セッションストレージをクリア
      mockSessionStorage.getItem.mockReturnValue(null);
    });

    it('新しいトークンを生成する', () => {
      const token = csrfTokenManager.generateToken();
      
      expect(typeof token).toBe('string');
      expect(token.length).toBeGreaterThanOrEqual(32);
      expect(mockSessionStorage.setItem).toHaveBeenCalledWith('csrf_token', token);
      expect(mockSessionStorage.setItem).toHaveBeenCalledWith('csrf_token_expiry', expect.any(String));
    });

    it('既存の有効なトークンを取得する', () => {
      const existingToken = 'existing_token_12345678901234567890';
      const futureExpiry = (Date.now() + 30 * 60 * 1000).toString(); // 30分後
      
      mockSessionStorage.getItem.mockImplementation((key) => {
        if (key === 'csrf_token') return existingToken;
        if (key === 'csrf_token_expiry') return futureExpiry;
        return null;
      });

      const token = csrfTokenManager.getToken();
      expect(token).toBe(existingToken);
    });

    it('期限切れのトークンは新しいものを生成する', () => {
      const expiredToken = 'expired_token_12345678901234567890';
      const pastExpiry = (Date.now() - 60 * 1000).toString(); // 1分前
      
      mockSessionStorage.getItem.mockImplementation((key) => {
        if (key === 'csrf_token') return expiredToken;
        if (key === 'csrf_token_expiry') return pastExpiry;
        return null;
      });

      const token = csrfTokenManager.getToken();
      expect(token).not.toBe(expiredToken);
      expect(mockSessionStorage.removeItem).toHaveBeenCalledWith('csrf_token');
      expect(mockSessionStorage.removeItem).toHaveBeenCalledWith('csrf_token_expiry');
    });

    it('トークンの検証を行う', () => {
      const validToken = 'valid_token_123456789012345678901';
      const futureExpiry = (Date.now() + 30 * 60 * 1000).toString();
      
      mockSessionStorage.getItem.mockImplementation((key) => {
        if (key === 'csrf_token') return validToken;
        if (key === 'csrf_token_expiry') return futureExpiry;
        return null;
      });

      expect(csrfTokenManager.validateToken(validToken)).toBe(true);
      expect(csrfTokenManager.validateToken('invalid_token')).toBe(false);
    });

    it('トークンをクリアする', () => {
      csrfTokenManager.clearToken();
      
      expect(mockSessionStorage.removeItem).toHaveBeenCalledWith('csrf_token');
      expect(mockSessionStorage.removeItem).toHaveBeenCalledWith('csrf_token_expiry');
    });
  });

  describe('addSecurityHeaders', () => {
    it('セキュリティヘッダーを追加する', () => {
      const existingHeaders = { 'Content-Type': 'application/json' };
      
      // CSRFトークンをモック
      mockSessionStorage.getItem.mockImplementation((key) => {
        if (key === 'csrf_token') return 'mock_csrf_token';
        if (key === 'csrf_token_expiry') return (Date.now() + 30 * 60 * 1000).toString();
        return null;
      });

      const result = addSecurityHeaders(existingHeaders);
      
      expect(result).toHaveProperty('Content-Type', 'application/json');
      expect(result).toHaveProperty('X-Content-Type-Options', 'nosniff');
      expect(result).toHaveProperty('X-Frame-Options', 'DENY');
      expect(result).toHaveProperty('X-XSS-Protection', '1; mode=block');
      expect(result).toHaveProperty('X-CSRF-Token', 'mock_csrf_token');
    });

    it('空のヘッダーでも動作する', () => {
      mockSessionStorage.getItem.mockImplementation((key) => {
        if (key === 'csrf_token') return 'mock_csrf_token';
        if (key === 'csrf_token_expiry') return (Date.now() + 30 * 60 * 1000).toString();
        return null;
      });

      const result = addSecurityHeaders();
      
      expect(result).toHaveProperty('X-Content-Type-Options', 'nosniff');
      expect(result).toHaveProperty('X-CSRF-Token', 'mock_csrf_token');
    });
  });

  describe('enforceInputLimits', () => {
    it('制限内の文字列はそのまま返す', () => {
      const input = 'Hello World';
      const result = enforceInputLimits(input, 20);
      expect(result).toBe(input);
    });

    it('制限を超える文字列は切り詰める', () => {
      const input = 'This is a very long string that exceeds the limit';
      const result = enforceInputLimits(input, 10);
      expect(result).toBe('This is a ');
      expect(result).toHaveLength(10);
    });

    it('文字列以外の値はそのまま返す', () => {
      expect(enforceInputLimits(123, 10)).toBe(123);
      expect(enforceInputLimits(null, 10)).toBe(null);
      expect(enforceInputLimits(undefined, 10)).toBe(undefined);
    });

    it('デフォルトの制限は1000文字', () => {
      const longInput = 'a'.repeat(1500);
      const result = enforceInputLimits(longInput);
      expect(result).toHaveLength(1000);
    });
  });

  describe('RateLimiter', () => {
    let rateLimiter;

    beforeEach(() => {
      // テスト用に短い時間窓を使用
      rateLimiter = new (class extends defaultRateLimiter.constructor {
        constructor() {
          super(3, 1000); // 1秒間に3リクエスト
        }
      })();
    });

    it('制限内のリクエストは許可する', () => {
      expect(rateLimiter.isAllowed('user1')).toBe(true);
      expect(rateLimiter.isAllowed('user1')).toBe(true);
      expect(rateLimiter.isAllowed('user1')).toBe(true);
    });

    it('制限を超えるリクエストは拒否する', () => {
      // 制限まで使い切る
      rateLimiter.isAllowed('user1');
      rateLimiter.isAllowed('user1');
      rateLimiter.isAllowed('user1');
      
      // 4回目は拒否される
      expect(rateLimiter.isAllowed('user1')).toBe(false);
    });

    it('異なる識別子は独立してカウントする', () => {
      rateLimiter.isAllowed('user1');
      rateLimiter.isAllowed('user1');
      rateLimiter.isAllowed('user1');
      
      // user1は制限に達したがuser2は許可される
      expect(rateLimiter.isAllowed('user1')).toBe(false);
      expect(rateLimiter.isAllowed('user2')).toBe(true);
    });

    it('残りリクエスト数を正しく計算する', () => {
      expect(rateLimiter.getRemainingRequests('user1')).toBe(3);
      
      rateLimiter.isAllowed('user1');
      expect(rateLimiter.getRemainingRequests('user1')).toBe(2);
      
      rateLimiter.isAllowed('user1');
      expect(rateLimiter.getRemainingRequests('user1')).toBe(1);
    });
  });

  describe('SecureStorage', () => {
    let storage;

    beforeEach(() => {
      storage = new (secureStorage.constructor)('test_');
      mockLocalStorage.getItem.mockReturnValue(null);
    });

    it('データを暗号化して保存する', () => {
      const testData = { username: 'admin', role: 'administrator' };
      
      const result = storage.setItem('user', testData);
      expect(result).toBe(true);
      expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
        'test_user',
        expect.stringContaining('"value"')
      );
    });

    it('保存されたデータを復号化して取得する', () => {
      const testData = { username: 'admin', role: 'administrator' };
      const storedData = {
        value: testData,
        timestamp: Date.now(),
        checksum: storage.generateChecksum(JSON.stringify(testData))
      };
      
      mockLocalStorage.getItem.mockReturnValue(JSON.stringify(storedData));
      
      const result = storage.getItem('user');
      expect(result).toEqual(testData);
    });

    it('チェックサムが一致しない場合はnullを返す', () => {
      const testData = { username: 'admin', role: 'administrator' };
      const corruptedData = {
        value: testData,
        timestamp: Date.now(),
        checksum: 'invalid_checksum'
      };
      
      mockLocalStorage.getItem.mockReturnValue(JSON.stringify(corruptedData));
      
      const result = storage.getItem('user');
      expect(result).toBe(null);
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('test_user');
    });

    it('存在しないキーの場合はnullを返す', () => {
      mockLocalStorage.getItem.mockReturnValue(null);
      
      const result = storage.getItem('nonexistent');
      expect(result).toBe(null);
    });

    it('データを削除する', () => {
      const result = storage.removeItem('user');
      expect(result).toBe(true);
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('test_user');
    });
  });

  describe('SecurityLogger', () => {
    let logger;

    beforeEach(() => {
      logger = new (securityLogger.constructor)();
      // コンソールをモック
      vi.spyOn(console, 'warn').mockImplementation(() => {});
      vi.spyOn(console, 'error').mockImplementation(() => {});
    });

    afterEach(() => {
      console.warn.mockRestore();
      console.error.mockRestore();
    });

    it('セキュリティイベントをログに記録する', () => {
      const eventType = 'TEST_EVENT';
      const details = { userId: 123, action: 'test' };
      
      logger.logEvent(eventType, details);
      
      const events = logger.getEvents();
      expect(events).toHaveLength(1);
      expect(events[0].type).toBe(eventType);
      expect(events[0].details).toEqual(details);
      expect(events[0].timestamp).toBeDefined();
    });

    it('重要なイベントを識別する', () => {
      expect(logger.isCriticalEvent('XSS_ATTEMPT')).toBe(true);
      expect(logger.isCriticalEvent('SQL_INJECTION_ATTEMPT')).toBe(true);
      expect(logger.isCriticalEvent('CSRF_TOKEN_MISMATCH')).toBe(true);
      expect(logger.isCriticalEvent('NORMAL_EVENT')).toBe(false);
    });

    it('重要なイベントは即座に報告する', () => {
      logger.logEvent('XSS_ATTEMPT', { payload: '<script>alert(1)</script>' });
      
      expect(console.error).toHaveBeenCalledWith(
        'CRITICAL SECURITY EVENT:',
        expect.objectContaining({
          type: 'XSS_ATTEMPT',
          details: { payload: '<script>alert(1)</script>' }
        })
      );
    });

    it('特定のタイプのイベントをフィルタリングして取得する', () => {
      logger.logEvent('LOGIN_ATTEMPT', { username: 'user1' });
      logger.logEvent('XSS_ATTEMPT', { payload: 'malicious' });
      logger.logEvent('LOGIN_ATTEMPT', { username: 'user2' });
      
      const loginEvents = logger.getEvents('LOGIN_ATTEMPT');
      expect(loginEvents).toHaveLength(2);
      expect(loginEvents.every(event => event.type === 'LOGIN_ATTEMPT')).toBe(true);
    });

    it('イベント履歴をクリアする', () => {
      logger.logEvent('TEST_EVENT', {});
      expect(logger.getEvents()).toHaveLength(1);
      
      logger.clearEvents();
      expect(logger.getEvents()).toHaveLength(0);
    });

    it('最大イベント数を超えると古いイベントを削除する', () => {
      // maxEventsを小さく設定したロガーを作成
      const smallLogger = new (securityLogger.constructor)();
      smallLogger.maxEvents = 3;
      
      smallLogger.logEvent('EVENT_1', {});
      smallLogger.logEvent('EVENT_2', {});
      smallLogger.logEvent('EVENT_3', {});
      smallLogger.logEvent('EVENT_4', {});
      
      const events = smallLogger.getEvents();
      expect(events).toHaveLength(3);
      expect(events[0].type).toBe('EVENT_2'); // 最初のイベントが削除される
      expect(events[2].type).toBe('EVENT_4'); // 最新のイベントが保持される
    });
  });
});
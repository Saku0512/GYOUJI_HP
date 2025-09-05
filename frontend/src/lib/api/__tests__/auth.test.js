// 認証APIクライアントの単体テスト
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { AuthAPI } from '../auth.js';

// モックAPIクライアント
const mockApiClient = {
  setToken: vi.fn(),
  post: vi.fn(),
  get: vi.fn()
};

// ローカルストレージのモック
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn()
};

// グローバルオブジェクトのモック
Object.defineProperty(window, 'localStorage', {
  value: localStorageMock
});

describe('AuthAPI', () => {
  let authAPI;

  beforeEach(() => {
    // モックをリセット
    vi.clearAllMocks();
    
    // AuthAPIインスタンスを作成
    authAPI = new AuthAPI(mockApiClient);
  });

  describe('login', () => {
    it('成功時にトークンとユーザー情報を保存する', async () => {
      const mockResponse = {
        success: true,
        data: {
          token: 'mock-jwt-token',
          refresh_token: 'mock-refresh-token',
          user: { id: 1, username: 'admin' }
        }
      };

      mockApiClient.post.mockResolvedValue(mockResponse);

      const result = await authAPI.login('admin', 'password');

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/login', {
        username: 'admin',
        password: 'password'
      });

      expect(localStorageMock.setItem).toHaveBeenCalledWith('auth_token', 'mock-jwt-token');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('refresh_token', 'mock-refresh-token');
      expect(localStorageMock.setItem).toHaveBeenCalledWith('auth_user', JSON.stringify({ id: 1, username: 'admin' }));
      
      expect(mockApiClient.setToken).toHaveBeenCalledWith('mock-jwt-token');
      
      expect(result.success).toBe(true);
      expect(result.data.token).toBe('mock-jwt-token');
    });

    it('失敗時にエラーを返す', async () => {
      const mockResponse = {
        success: false,
        error: 'INVALID_CREDENTIALS',
        message: '認証情報が正しくありません'
      };

      mockApiClient.post.mockResolvedValue(mockResponse);

      const result = await authAPI.login('admin', 'wrong-password');

      expect(result.success).toBe(false);
      expect(result.error).toBe('INVALID_CREDENTIALS');
    });
  });

  describe('logout', () => {
    it('ローカルの認証情報をクリアする', async () => {
      localStorageMock.getItem.mockReturnValue('mock-token');
      mockApiClient.post.mockResolvedValue({ success: true });

      const result = await authAPI.logout();

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/logout');
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('auth_token');
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('refresh_token');
      expect(localStorageMock.removeItem).toHaveBeenCalledWith('auth_user');
      expect(mockApiClient.setToken).toHaveBeenCalledWith(null);
      
      expect(result.success).toBe(true);
    });
  });

  describe('refreshToken', () => {
    it('リフレッシュトークンを使用して新しいトークンを取得する', async () => {
      localStorageMock.getItem.mockReturnValue('mock-refresh-token');
      
      const mockResponse = {
        success: true,
        data: {
          token: 'new-jwt-token',
          refresh_token: 'new-refresh-token',
          user: { id: 1, username: 'admin' }
        }
      };

      mockApiClient.post.mockResolvedValue(mockResponse);

      const result = await authAPI.refreshToken();

      expect(mockApiClient.post).toHaveBeenCalledWith('/auth/refresh', {
        refresh_token: 'mock-refresh-token'
      });

      expect(localStorageMock.setItem).toHaveBeenCalledWith('auth_token', 'new-jwt-token');
      expect(result.success).toBe(true);
    });

    it('リフレッシュトークンがない場合はエラーを返す', async () => {
      localStorageMock.getItem.mockReturnValue(null);

      const result = await authAPI.refreshToken();

      expect(result.success).toBe(false);
      expect(result.error).toBe('NO_REFRESH_TOKEN');
    });
  });

  describe('validateToken', () => {
    it('有効なトークンの場合は成功を返す', async () => {
      // 有効なJWTトークンのモック（期限が未来）
      const futureTimestamp = Math.floor(Date.now() / 1000) + 3600; // 1時間後
      const mockToken = `header.${btoa(JSON.stringify({ exp: futureTimestamp }))}.signature`;
      
      localStorageMock.getItem.mockReturnValue(mockToken);
      mockApiClient.get.mockResolvedValue({ success: true, data: { valid: true } });

      const result = await authAPI.validateToken();

      expect(mockApiClient.get).toHaveBeenCalledWith('/auth/validate');
      expect(result.success).toBe(true);
    });

    it('トークンがない場合はエラーを返す', async () => {
      localStorageMock.getItem.mockReturnValue(null);

      const result = await authAPI.validateToken();

      expect(result.success).toBe(false);
      expect(result.error).toBe('NO_TOKEN');
    });
  });

  describe('isTokenExpired', () => {
    it('期限切れトークンの場合はtrueを返す', () => {
      const pastTimestamp = Math.floor(Date.now() / 1000) - 3600; // 1時間前
      const expiredToken = `header.${btoa(JSON.stringify({ exp: pastTimestamp }))}.signature`;

      const result = authAPI.isTokenExpired(expiredToken);

      expect(result).toBe(true);
    });

    it('有効なトークンの場合はfalseを返す', () => {
      const futureTimestamp = Math.floor(Date.now() / 1000) + 3600; // 1時間後
      const validToken = `header.${btoa(JSON.stringify({ exp: futureTimestamp }))}.signature`;

      const result = authAPI.isTokenExpired(validToken);

      expect(result).toBe(false);
    });

    it('無効なトークンの場合はtrueを返す', () => {
      const result = authAPI.isTokenExpired('invalid-token');

      expect(result).toBe(true);
    });
  });

  describe('getAuthState', () => {
    it('認証状態を正しく返す', () => {
      const futureTimestamp = Math.floor(Date.now() / 1000) + 3600;
      const validToken = `header.${btoa(JSON.stringify({ exp: futureTimestamp }))}.signature`;
      const mockUser = { id: 1, username: 'admin' };

      localStorageMock.getItem
        .mockReturnValueOnce(validToken) // getStoredToken
        .mockReturnValueOnce(JSON.stringify(mockUser)); // getStoredUser

      const result = authAPI.getAuthState();

      expect(result.isAuthenticated).toBe(true);
      expect(result.token).toBe(validToken);
      expect(result.user).toEqual(mockUser);
    });
  });
});
// 認証ストアの単体テスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import { authStore } from '../auth.js';
import * as authAPI from '../../api/auth.js';
import * as storage from '../../utils/storage.js';

// モック設定
vi.mock('../../api/auth.js');
vi.mock('../../utils/storage.js');

describe('認証ストア', () => {
  beforeEach(() => {
    // 各テスト前にモックをリセット
    vi.clearAllMocks();
    
    // ローカルストレージのモック
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      },
      writable: true,
    });

    // DOM イベントのモック
    Object.defineProperty(document, 'addEventListener', {
      value: vi.fn(),
      writable: true,
    });

    Object.defineProperty(window, 'addEventListener', {
      value: vi.fn(),
      writable: true,
    });
  });

  afterEach(() => {
    // ストア状態をリセット
    authStore.clearAuth();
  });

  describe('初期状態', () => {
    it('初期状態が正しく設定されている', () => {
      const state = get(authStore);
      
      expect(state.isAuthenticated).toBe(false);
      expect(state.token).toBeNull();
      expect(state.user).toBeNull();
      expect(state.loading).toBe(false);
    });
  });

  describe('ローディング状態管理', () => {
    it('ローディング状態を正しく設定できる', () => {
      authStore.setLoading(true);
      
      const state = get(authStore);
      expect(state.loading).toBe(true);
    });

    it('ローディング状態を解除できる', () => {
      authStore.setLoading(true);
      authStore.setLoading(false);
      
      const state = get(authStore);
      expect(state.loading).toBe(false);
    });
  });

  describe('認証状態設定', () => {
    it('認証状態を正しく設定できる', () => {
      const authData = {
        token: 'test-token',
        user: { id: 1, username: 'testuser' }
      };

      authStore.setAuthState(authData);
      
      const state = get(authStore);
      expect(state.isAuthenticated).toBe(true);
      expect(state.token).toBe('test-token');
      expect(state.user).toEqual({ id: 1, username: 'testuser' });
      expect(state.loading).toBe(false);
    });

    it('トークンがない場合は未認証状態になる', () => {
      const authData = {
        token: null,
        user: null
      };

      authStore.setAuthState(authData);
      
      const state = get(authStore);
      expect(state.isAuthenticated).toBe(false);
    });
  });

  describe('認証状態クリア', () => {
    it('認証状態を正しくクリアできる', () => {
      // まず認証状態を設定
      authStore.setAuthState({
        token: 'test-token',
        user: { id: 1, username: 'testuser' }
      });

      // クリア実行
      authStore.clearAuth();
      
      const state = get(authStore);
      expect(state.isAuthenticated).toBe(false);
      expect(state.token).toBeNull();
      expect(state.user).toBeNull();
      expect(state.loading).toBe(false);
    });
  });

  describe('ログイン処理', () => {
    it('ログインが成功した場合、認証状態が更新される', async () => {
      const mockResponse = {
        success: true,
        data: {
          token: 'new-token',
          user: { id: 1, username: 'testuser' }
        }
      };

      // authAPI.loginのモック
      authAPI.authAPI.login = vi.fn().mockResolvedValue(mockResponse);
      
      // ストレージ関数のモック
      storage.saveAuthToken = vi.fn();
      storage.saveUserData = vi.fn();

      const credentials = { username: 'testuser', password: 'password' };
      const result = await authStore.login(credentials);

      expect(result.success).toBe(true);
      expect(authAPI.authAPI.login).toHaveBeenCalledWith('testuser', 'password');
      expect(storage.saveAuthToken).toHaveBeenCalledWith('new-token');
      expect(storage.saveUserData).toHaveBeenCalledWith({ id: 1, username: 'testuser' });

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(true);
      expect(state.token).toBe('new-token');
      expect(state.user).toEqual({ id: 1, username: 'testuser' });
      expect(state.loading).toBe(false);
    });

    it('ログインが失敗した場合、エラーが返される', async () => {
      const mockResponse = {
        success: false,
        error: 'INVALID_CREDENTIALS',
        message: '認証情報が正しくありません'
      };

      authAPI.authAPI.login = vi.fn().mockResolvedValue(mockResponse);

      const credentials = { username: 'testuser', password: 'wrongpassword' };
      const result = await authStore.login(credentials);

      expect(result.success).toBe(false);
      expect(result.error).toBe('INVALID_CREDENTIALS');

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(false);
      expect(state.loading).toBe(false);
    });

    it('ログイン処理中にエラーが発生した場合、適切に処理される', async () => {
      authAPI.authAPI.login = vi.fn().mockRejectedValue(new Error('Network error'));

      const credentials = { username: 'testuser', password: 'password' };
      const result = await authStore.login(credentials);

      expect(result.success).toBe(false);
      expect(result.error).toBe('LOGIN_STORE_ERROR');

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(false);
      expect(state.loading).toBe(false);
    });
  });

  describe('ログアウト処理', () => {
    it('ログアウトが成功した場合、認証状態がクリアされる', async () => {
      // まず認証状態を設定
      authStore.setAuthState({
        token: 'test-token',
        user: { id: 1, username: 'testuser' }
      });

      const mockResponse = {
        success: true,
        message: 'ログアウトしました'
      };

      authAPI.authAPI.logout = vi.fn().mockResolvedValue(mockResponse);
      storage.clearAuthData = vi.fn();

      const result = await authStore.logout();

      expect(result.success).toBe(true);
      expect(authAPI.authAPI.logout).toHaveBeenCalled();
      expect(storage.clearAuthData).toHaveBeenCalled();

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(false);
      expect(state.token).toBeNull();
      expect(state.user).toBeNull();
    });

    it('ログアウト処理中にエラーが発生してもローカル状態はクリアされる', async () => {
      // まず認証状態を設定
      authStore.setAuthState({
        token: 'test-token',
        user: { id: 1, username: 'testuser' }
      });

      authAPI.authAPI.logout = vi.fn().mockRejectedValue(new Error('Network error'));
      storage.clearAuthData = vi.fn();

      const result = await authStore.logout();

      expect(result.success).toBe(true);
      expect(result.warning).toBeDefined();
      expect(storage.clearAuthData).toHaveBeenCalled();

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(false);
      expect(state.token).toBeNull();
      expect(state.user).toBeNull();
    });
  });

  describe('認証状態チェック', () => {
    it('有効なトークンがある場合、認証状態が復元される', async () => {
      const mockToken = 'valid-token';
      const mockUser = { id: 1, username: 'testuser' };

      storage.getAuthToken = vi.fn().mockReturnValue(mockToken);
      storage.getUserData = vi.fn().mockReturnValue(mockUser);

      const mockResponse = {
        success: true,
        data: { user: mockUser }
      };

      authAPI.authAPI.validateToken = vi.fn().mockResolvedValue(mockResponse);

      const result = await authStore.checkAuthStatus();

      expect(result.success).toBe(true);
      expect(storage.getAuthToken).toHaveBeenCalled();
      expect(authAPI.authAPI.validateToken).toHaveBeenCalled();

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(true);
      expect(state.token).toBe(mockToken);
      expect(state.user).toEqual(mockUser);
    });

    it('トークンがない場合、エラーが返される', async () => {
      storage.getAuthToken = vi.fn().mockReturnValue(null);

      const result = await authStore.checkAuthStatus();

      expect(result.success).toBe(false);
      expect(result.error).toBe('NO_TOKEN');

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(false);
    });

    it('トークン検証が失敗した場合、認証状態がクリアされる', async () => {
      storage.getAuthToken = vi.fn().mockReturnValue('invalid-token');
      storage.getUserData = vi.fn().mockReturnValue({ id: 1, username: 'testuser' });

      const mockResponse = {
        success: false,
        error: 'INVALID_TOKEN',
        message: 'トークンが無効です'
      };

      authAPI.authAPI.validateToken = vi.fn().mockResolvedValue(mockResponse);
      storage.clearAuthData = vi.fn();

      const result = await authStore.checkAuthStatus();

      expect(result.success).toBe(false);
      expect(storage.clearAuthData).toHaveBeenCalled();

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(false);
    });
  });

  describe('トークンリフレッシュ', () => {
    it('トークンリフレッシュが成功した場合、新しいトークンで状態が更新される', async () => {
      const mockResponse = {
        success: true,
        data: {
          token: 'new-refreshed-token',
          user: { id: 1, username: 'testuser' }
        }
      };

      authAPI.authAPI.refreshToken = vi.fn().mockResolvedValue(mockResponse);
      storage.saveAuthToken = vi.fn();
      storage.saveUserData = vi.fn();

      const result = await authStore.refreshToken();

      expect(result.success).toBe(true);
      expect(authAPI.authAPI.refreshToken).toHaveBeenCalled();
      expect(storage.saveAuthToken).toHaveBeenCalledWith('new-refreshed-token');

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(true);
      expect(state.token).toBe('new-refreshed-token');
    });

    it('トークンリフレッシュが失敗した場合、認証状態がクリアされる', async () => {
      const mockResponse = {
        success: false,
        error: 'REFRESH_FAILED',
        message: 'リフレッシュに失敗しました'
      };

      authAPI.authAPI.refreshToken = vi.fn().mockResolvedValue(mockResponse);
      storage.clearAuthData = vi.fn();

      const result = await authStore.refreshToken();

      expect(result.success).toBe(false);
      expect(storage.clearAuthData).toHaveBeenCalled();

      const state = get(authStore);
      expect(state.isAuthenticated).toBe(false);
    });

    it('既にローディング中の場合、処理をスキップする', async () => {
      // ローディング状態を設定
      authStore.setLoading(true);

      const result = await authStore.refreshToken();

      expect(result.success).toBe(false);
      expect(result.error).toBe('ALREADY_LOADING');
    });
  });

  describe('初期化処理', () => {
    it('ストレージに認証情報がある場合、認証状態チェックが実行される', async () => {
      storage.getAuthToken = vi.fn().mockReturnValue('stored-token');
      storage.getUserData = vi.fn().mockReturnValue({ id: 1, username: 'testuser' });

      // checkAuthStatusをスパイ
      const checkAuthStatusSpy = vi.spyOn(authStore, 'checkAuthStatus').mockResolvedValue({
        success: true
      });

      await authStore.initialize();

      expect(storage.getAuthToken).toHaveBeenCalled();
      expect(storage.getUserData).toHaveBeenCalled();
      expect(checkAuthStatusSpy).toHaveBeenCalled();
    });

    it('ストレージに認証情報がない場合、認証状態チェックは実行されない', async () => {
      storage.getAuthToken = vi.fn().mockReturnValue(null);
      storage.getUserData = vi.fn().mockReturnValue(null);

      const checkAuthStatusSpy = vi.spyOn(authStore, 'checkAuthStatus');

      await authStore.initialize();

      expect(checkAuthStatusSpy).not.toHaveBeenCalled();
    });
  });
});
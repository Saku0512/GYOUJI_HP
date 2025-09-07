// 認証ガードのテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { 
  isTokenExpired, 
  getTokenTimeRemaining,
  requireAuth,
  requireAdmin,
  redirectIfAuthenticated,
  handleAuthError
} from '../auth-guard.js';

// モック設定
vi.mock('@sveltejs/kit', () => ({
  redirect: vi.fn((status, location) => {
    throw new Error(`Redirect: ${status} ${location}`);
  })
}));

vi.mock('../storage.js', () => ({
  getAuthToken: vi.fn(),
  getUserData: vi.fn()
}));

vi.mock('../../stores/auth.js', () => ({
  authStore: {
    refreshToken: vi.fn(),
    checkAuthStatus: vi.fn(),
    logout: vi.fn()
  }
}));

vi.mock('svelte/store', () => ({
  get: vi.fn()
}));

import { redirect } from '@sveltejs/kit';
import { getAuthToken, getUserData } from '../storage.js';
import { authStore } from '../../stores/auth.js';
import { get } from 'svelte/store';

describe('認証ガードユーティリティ', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // ブラウザ環境をシミュレート
    Object.defineProperty(window, 'window', {
      value: {},
      writable: true
    });
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('isTokenExpired', () => {
    it('トークンがnullの場合はtrueを返す', () => {
      expect(isTokenExpired(null)).toBe(true);
      expect(isTokenExpired(undefined)).toBe(true);
      expect(isTokenExpired('')).toBe(true);
    });

    it('有効なトークンの場合はfalseを返す', () => {
      // 1時間後に期限切れのトークンを作成
      const futureTime = Math.floor(Date.now() / 1000) + 3600;
      const payload = { exp: futureTime };
      const token = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      expect(isTokenExpired(token)).toBe(false);
    });

    it('期限切れのトークンの場合はtrueを返す', () => {
      // 1時間前に期限切れのトークンを作成
      const pastTime = Math.floor(Date.now() / 1000) - 3600;
      const payload = { exp: pastTime };
      const token = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      expect(isTokenExpired(token)).toBe(true);
    });

    it('5分以内に期限切れのトークンの場合はtrueを返す（バッファ時間）', () => {
      // 3分後に期限切れのトークンを作成
      const nearFutureTime = Math.floor(Date.now() / 1000) + 180;
      const payload = { exp: nearFutureTime };
      const token = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      expect(isTokenExpired(token)).toBe(true);
    });

    it('不正なトークンの場合はtrueを返す', () => {
      expect(isTokenExpired('invalid.token')).toBe(true);
      expect(isTokenExpired('invalid')).toBe(true);
    });
  });

  describe('getTokenTimeRemaining', () => {
    it('トークンがnullの場合は0を返す', () => {
      expect(getTokenTimeRemaining(null)).toBe(0);
      expect(getTokenTimeRemaining(undefined)).toBe(0);
      expect(getTokenTimeRemaining('')).toBe(0);
    });

    it('有効なトークンの残り時間を正しく計算する', () => {
      // 1時間後に期限切れのトークンを作成
      const futureTime = Math.floor(Date.now() / 1000) + 3600;
      const payload = { exp: futureTime };
      const token = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      const remaining = getTokenTimeRemaining(token);
      expect(remaining).toBeGreaterThan(3590); // 約1時間（少し余裕を持たせる）
      expect(remaining).toBeLessThanOrEqual(3600);
    });

    it('期限切れのトークンの場合は0を返す', () => {
      // 1時間前に期限切れのトークンを作成
      const pastTime = Math.floor(Date.now() / 1000) - 3600;
      const payload = { exp: pastTime };
      const token = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      expect(getTokenTimeRemaining(token)).toBe(0);
    });

    it('不正なトークンの場合は0を返す', () => {
      expect(getTokenTimeRemaining('invalid.token')).toBe(0);
      expect(getTokenTimeRemaining('invalid')).toBe(0);
    });
  });

  describe('requireAuth', () => {
    const mockUrl = {
      pathname: '/admin',
      search: '?test=1'
    };

    it('トークンが存在しない場合はログインページにリダイレクト', async () => {
      getAuthToken.mockReturnValue(null);

      await expect(requireAuth(mockUrl)).rejects.toThrow('Redirect: 302 /login?redirect=%2Fadmin%3Ftest%3D1');
      expect(redirect).toHaveBeenCalledWith(302, '/login?redirect=%2Fadmin%3Ftest%3D1');
    });

    it('期限切れトークンの場合はリフレッシュを試行', async () => {
      // 期限切れトークンを設定
      const pastTime = Math.floor(Date.now() / 1000) - 3600;
      const payload = { exp: pastTime };
      const expiredToken = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      getAuthToken.mockReturnValue(expiredToken);
      authStore.refreshToken.mockResolvedValue({ success: true });
      authStore.checkAuthStatus.mockResolvedValue({ success: true });

      const result = await requireAuth(mockUrl);
      expect(result).toEqual({ authenticated: true });
      expect(authStore.refreshToken).toHaveBeenCalled();
    });

    it('リフレッシュに失敗した場合はログインページにリダイレクト', async () => {
      // 期限切れトークンを設定
      const pastTime = Math.floor(Date.now() / 1000) - 3600;
      const payload = { exp: pastTime };
      const expiredToken = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      getAuthToken.mockReturnValue(expiredToken);
      authStore.refreshToken.mockResolvedValue({ success: false });

      await expect(requireAuth(mockUrl)).rejects.toThrow('Redirect: 302 /login?expired=true&redirect=%2Fadmin%3Ftest%3D1');
    });

    it('有効なトークンの場合は認証成功を返す', async () => {
      // 有効なトークンを設定
      const futureTime = Math.floor(Date.now() / 1000) + 3600;
      const payload = { exp: futureTime };
      const validToken = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      getAuthToken.mockReturnValue(validToken);
      authStore.checkAuthStatus.mockResolvedValue({ success: true });

      const result = await requireAuth(mockUrl);
      expect(result).toEqual({ authenticated: true });
    });

    it('サーバーサイド検証に失敗した場合はリダイレクト', async () => {
      // 有効なトークンを設定
      const futureTime = Math.floor(Date.now() / 1000) + 3600;
      const payload = { exp: futureTime };
      const validToken = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      getAuthToken.mockReturnValue(validToken);
      authStore.checkAuthStatus.mockResolvedValue({ success: false });

      await expect(requireAuth(mockUrl)).rejects.toThrow('Redirect: 302 /login?error=true&redirect=%2Fadmin%3Ftest%3D1');
    });
  });

  describe('requireAdmin', () => {
    const mockUrl = {
      pathname: '/admin',
      search: ''
    };

    it('認証されていない場合は認証チェックが実行される', async () => {
      getAuthToken.mockReturnValue(null);

      await expect(requireAdmin(mockUrl)).rejects.toThrow('Redirect: 302 /login?redirect=%2Fadmin');
    });

    it('管理者権限がない場合は未認可エラー', async () => {
      // 有効なトークンを設定
      const futureTime = Math.floor(Date.now() / 1000) + 3600;
      const payload = { exp: futureTime };
      const validToken = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      getAuthToken.mockReturnValue(validToken);
      authStore.checkAuthStatus.mockResolvedValue({ success: true });
      
      // 一般ユーザーを設定
      get.mockReturnValue({
        user: { role: 'user' }
      });

      await expect(requireAdmin(mockUrl)).rejects.toThrow('Redirect: 302 /login?unauthorized=true');
    });

    it('管理者権限がある場合は成功', async () => {
      // 有効なトークンを設定
      const futureTime = Math.floor(Date.now() / 1000) + 3600;
      const payload = { exp: futureTime };
      const validToken = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      getAuthToken.mockReturnValue(validToken);
      authStore.checkAuthStatus.mockResolvedValue({ success: true });
      
      // 管理者ユーザーを設定
      get.mockReturnValue({
        user: { role: 'admin' }
      });

      const result = await requireAdmin(mockUrl);
      expect(result).toEqual({ authenticated: true, isAdmin: true });
    });
  });

  describe('redirectIfAuthenticated', () => {
    const mockUrl = {
      pathname: '/login',
      search: '',
      searchParams: {
        get: vi.fn()
      }
    };

    it('認証されている場合はリダイレクト', () => {
      // 有効なトークンを設定
      const futureTime = Math.floor(Date.now() / 1000) + 3600;
      const payload = { exp: futureTime };
      const validToken = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      getAuthToken.mockReturnValue(validToken);
      mockUrl.searchParams.get.mockReturnValue(null);

      expect(() => redirectIfAuthenticated(mockUrl)).toThrow('Redirect: 302 /admin');
    });

    it('リダイレクト先パラメータがある場合はそちらを優先', () => {
      // 有効なトークンを設定
      const futureTime = Math.floor(Date.now() / 1000) + 3600;
      const payload = { exp: futureTime };
      const validToken = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      getAuthToken.mockReturnValue(validToken);
      mockUrl.searchParams.get.mockReturnValue('/dashboard');

      expect(() => redirectIfAuthenticated(mockUrl)).toThrow('Redirect: 302 /dashboard');
    });

    it('認証されていない場合はリダイレクトしない', () => {
      getAuthToken.mockReturnValue(null);
      mockUrl.searchParams.get.mockReturnValue(null);

      const result = redirectIfAuthenticated(mockUrl);
      expect(result).toEqual({ shouldRedirect: false });
    });

    it('期限切れトークンの場合はリダイレクトしない', () => {
      // 期限切れトークンを設定
      const pastTime = Math.floor(Date.now() / 1000) - 3600;
      const payload = { exp: pastTime };
      const expiredToken = `header.${btoa(JSON.stringify(payload))}.signature`;
      
      getAuthToken.mockReturnValue(expiredToken);
      mockUrl.searchParams.get.mockReturnValue(null);

      const result = redirectIfAuthenticated(mockUrl);
      expect(result).toEqual({ shouldRedirect: false });
    });
  });

  describe('handleAuthError', () => {
    const currentUrl = '/admin/dashboard';

    it('TOKEN_EXPIREDエラーの場合', () => {
      const error = { error: 'TOKEN_EXPIRED' };
      
      expect(() => handleAuthError(error, currentUrl)).toThrow('Redirect: 302 /login?expired=true&redirect=%2Fadmin%2Fdashboard');
    });

    it('NO_TOKENエラーの場合', () => {
      const error = { error: 'NO_TOKEN' };
      
      expect(() => handleAuthError(error, currentUrl)).toThrow('Redirect: 302 /login?expired=true&redirect=%2Fadmin%2Fdashboard');
    });

    it('INVALID_TOKENエラーの場合', () => {
      const error = { error: 'INVALID_TOKEN' };
      
      expect(() => handleAuthError(error, currentUrl)).toThrow('Redirect: 302 /login?invalid=true&redirect=%2Fadmin%2Fdashboard');
    });

    it('UNAUTHORIZEDエラーの場合', () => {
      const error = { error: 'UNAUTHORIZED' };
      
      expect(() => handleAuthError(error, currentUrl)).toThrow('Redirect: 302 /login?unauthorized=true');
    });

    it('NETWORK_ERRORエラーの場合', () => {
      const error = { error: 'NETWORK_ERROR' };
      
      expect(() => handleAuthError(error, currentUrl)).toThrow('Redirect: 302 /login?network_error=true&redirect=%2Fadmin%2Fdashboard');
    });

    it('未知のエラーの場合', () => {
      const error = { error: 'UNKNOWN_ERROR' };
      
      expect(() => handleAuthError(error, currentUrl)).toThrow('Redirect: 302 /login?error=true&redirect=%2Fadmin%2Fdashboard');
    });
  });
});
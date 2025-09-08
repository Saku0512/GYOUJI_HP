// 統一APIクライアントのテスト
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { UnifiedAPI, UnifiedAPIClient } from '../unified-client.js';

// fetchのモック
global.fetch = vi.fn();

describe('UnifiedAPIClient', () => {
  let client;

  beforeEach(() => {
    client = new UnifiedAPIClient('/api');
    vi.clearAllMocks();
  });

  describe('基本的なリクエスト機能', () => {
    it('GETリクエストが正しく送信される', async () => {
      const mockResponse = {
        success: true,
        data: { test: 'data' },
        message: 'Success',
        code: 200,
        timestamp: '2024-01-01T00:00:00Z'
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: () => Promise.resolve(mockResponse)
      });

      const result = await client.get('/test');

      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/test',
        expect.objectContaining({
          method: 'GET',
          headers: expect.objectContaining({
            'Content-Type': 'application/json',
            'Accept': 'application/json'
          })
        })
      );

      expect(result).toEqual(mockResponse);
    });

    it('POSTリクエストでデータが正しく送信される', async () => {
      const testData = { username: 'test', password: 'password' };
      const mockResponse = {
        success: true,
        data: { token: 'test-token' },
        message: 'Login successful',
        code: 200,
        timestamp: '2024-01-01T00:00:00Z'
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: () => Promise.resolve(mockResponse)
      });

      const result = await client.post('/auth/login', testData);

      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/auth/login',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify(testData),
          headers: expect.objectContaining({
            'Content-Type': 'application/json'
          })
        })
      );

      expect(result).toEqual(mockResponse);
    });

    it('認証トークンがヘッダーに含まれる', async () => {
      client.setToken('test-token');

      fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: () => Promise.resolve({ success: true })
      });

      await client.get('/protected');

      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/protected',
        expect.objectContaining({
          headers: expect.objectContaining({
            'Authorization': 'Bearer test-token'
          })
        })
      );
    });
  });

  describe('エラーハンドリング', () => {
    it('ネットワークエラーが適切に処理される', async () => {
      fetch.mockRejectedValueOnce(new TypeError('Failed to fetch'));

      const result = await client.get('/test');

      expect(result.success).toBe(false);
      expect(result.error).toBe('SYSTEM_NETWORK_ERROR');
      expect(result.message).toBe('ネットワークエラーが発生しました');
    });

    it('HTTPエラーステータスが適切に処理される', async () => {
      const errorResponse = {
        success: false,
        error: 'AUTH_UNAUTHORIZED',
        message: '認証が必要です',
        code: 401
      };

      fetch.mockResolvedValueOnce({
        ok: false,
        status: 401,
        headers: new Map([['content-type', 'application/json']]),
        json: () => Promise.resolve(errorResponse)
      });

      const result = await client.get('/protected');

      expect(result).toEqual(expect.objectContaining({
        success: false,
        error: 'AUTH_UNAUTHORIZED',
        code: 401
      }));
    });
  });
});

describe('UnifiedAPI', () => {
  let api;

  beforeEach(() => {
    api = new UnifiedAPI('/api');
    vi.clearAllMocks();
  });

  describe('認証API', () => {
    it('ログインが正しく動作する', async () => {
      const mockResponse = {
        success: true,
        data: {
          token: 'test-token',
          user: { id: 1, username: 'test' }
        },
        message: 'Login successful',
        code: 200
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: () => Promise.resolve(mockResponse)
      });

      const result = await api.auth.login({
        username: 'test',
        password: 'password'
      });

      expect(result).toEqual(mockResponse);
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/auth/login',
        expect.objectContaining({
          method: 'POST',
          body: JSON.stringify({
            username: 'test',
            password: 'password'
          })
        })
      );
    });
  });

  describe('トーナメントAPI', () => {
    it('トーナメント一覧取得が正しく動作する', async () => {
      const mockResponse = {
        success: true,
        data: [
          { id: 1, sport: 'volleyball', format: 'single_elimination' }
        ],
        message: 'Success',
        code: 200
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: () => Promise.resolve(mockResponse)
      });

      const result = await api.tournaments.getAll();

      expect(result).toEqual(mockResponse);
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/tournaments',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });
  });

  describe('試合API', () => {
    it('スポーツ別試合取得が正しく動作する', async () => {
      const mockResponse = {
        success: true,
        data: [
          { id: 1, team1: 'Team A', team2: 'Team B', status: 'pending' }
        ],
        message: 'Success',
        code: 200
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: () => Promise.resolve(mockResponse)
      });

      const result = await api.matches.getBySport('volleyball');

      expect(result).toEqual(mockResponse);
      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/matches/volleyball',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });

    it('フィルター付き試合取得が正しく動作する', async () => {
      const mockResponse = {
        success: true,
        data: [],
        message: 'Success',
        code: 200
      };

      fetch.mockResolvedValueOnce({
        ok: true,
        status: 200,
        headers: new Map([['content-type', 'application/json']]),
        json: () => Promise.resolve(mockResponse)
      });

      const filters = { status: 'completed', limit: 10 };
      await api.matches.getBySport('volleyball', filters);

      expect(fetch).toHaveBeenCalledWith(
        '/api/v1/matches/volleyball?status=completed&limit=10',
        expect.objectContaining({
          method: 'GET'
        })
      );
    });
  });
});
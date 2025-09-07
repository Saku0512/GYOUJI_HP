// APIクライアントの単体テスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { APIClient } from '../client.js';

describe('APIClient', () => {
  let apiClient;
  let mockFetch;

  beforeEach(() => {
    mockFetch = vi.fn();
    global.fetch = mockFetch;
    apiClient = new APIClient('http://localhost:8080/api');
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  describe('constructor', () => {
    it('正しいベースURLで初期化される', () => {
      expect(apiClient.baseURL).toBe('http://localhost:8080/api');
      expect(apiClient.token).toBeNull();
    });
  });

  describe('setToken', () => {
    it('認証トークンを設定できる', () => {
      const token = 'test-token';
      apiClient.setToken(token);
      expect(apiClient.token).toBe(token);
    });

    it('nullトークンを設定できる', () => {
      apiClient.setToken('test-token');
      apiClient.setToken(null);
      expect(apiClient.token).toBeNull();
    });
  });

  describe('get', () => {
    it('GETリクエストを正しく送信する', async () => {
      const mockResponse = { success: true, data: { id: 1 } };
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse)
      });

      const result = await apiClient.get('/test');

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/test', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      expect(result).toEqual(mockResponse);
    });

    it('認証トークンがある場合はAuthorizationヘッダーを追加する', async () => {
      apiClient.setToken('test-token');
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ success: true })
      });

      await apiClient.get('/test');

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/test', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': 'Bearer test-token'
        }
      });
    });

    it('カスタムヘッダーを追加できる', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ success: true })
      });

      await apiClient.get('/test', {
        headers: { 'X-Custom': 'value' }
      });

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/test', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'X-Custom': 'value'
        }
      });
    });
  });

  describe('post', () => {
    it('POSTリクエストを正しく送信する', async () => {
      const postData = { name: 'test' };
      const mockResponse = { success: true, data: { id: 1 } };
      mockFetch.mockResolvedValue({
        ok: true,
        status: 201,
        json: () => Promise.resolve(mockResponse)
      });

      const result = await apiClient.post('/test', postData);

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/test', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(postData)
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe('put', () => {
    it('PUTリクエストを正しく送信する', async () => {
      const putData = { id: 1, name: 'updated' };
      const mockResponse = { success: true, data: putData };
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockResponse)
      });

      const result = await apiClient.put('/test/1', putData);

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/test/1', {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(putData)
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe('delete', () => {
    it('DELETEリクエストを正しく送信する', async () => {
      const mockResponse = { success: true };
      mockFetch.mockResolvedValue({
        ok: true,
        status: 204,
        json: () => Promise.resolve(mockResponse)
      });

      const result = await apiClient.delete('/test/1');

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/test/1', {
        method: 'DELETE',
        headers: {
          'Content-Type': 'application/json'
        }
      });
      expect(result).toEqual(mockResponse);
    });
  });

  describe('handleResponse', () => {
    it('成功レスポンスを正しく処理する', async () => {
      const mockData = { success: true, data: { id: 1 } };
      const mockResponse = {
        ok: true,
        status: 200,
        json: () => Promise.resolve(mockData)
      };

      const result = await apiClient.handleResponse(mockResponse);
      expect(result).toEqual(mockData);
    });

    it('401エラーでトークンをクリアする', async () => {
      apiClient.setToken('test-token');
      const mockResponse = {
        ok: false,
        status: 401,
        json: () => Promise.resolve({ error: 'Unauthorized' })
      };

      await expect(apiClient.handleResponse(mockResponse)).rejects.toThrow();
      expect(apiClient.token).toBeNull();
    });

    it('HTTPエラーで例外を投げる', async () => {
      const mockResponse = {
        ok: false,
        status: 500,
        json: () => Promise.resolve({ error: 'Internal Server Error' })
      };

      await expect(apiClient.handleResponse(mockResponse)).rejects.toThrow('HTTP error! status: 500');
    });
  });

  describe('handleError', () => {
    it('ネットワークエラーを正しく処理する', () => {
      const networkError = new TypeError('Failed to fetch');
      
      expect(() => apiClient.handleError(networkError)).toThrow('ネットワークエラーが発生しました');
    });

    it('一般的なエラーを正しく処理する', () => {
      const genericError = new Error('Something went wrong');
      
      expect(() => apiClient.handleError(genericError)).toThrow('APIエラーが発生しました: Something went wrong');
    });
  });

  describe('エラーハンドリング統合テスト', () => {
    it('ネットワークエラー時に適切なエラーメッセージを返す', async () => {
      mockFetch.mockRejectedValue(new TypeError('Failed to fetch'));

      await expect(apiClient.get('/test')).rejects.toThrow('ネットワークエラーが発生しました');
    });

    it('JSONパースエラー時に適切なエラーメッセージを返す', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.reject(new SyntaxError('Unexpected token'))
      });

      await expect(apiClient.get('/test')).rejects.toThrow('APIエラーが発生しました');
    });
  });

  describe('リクエストオプション', () => {
    it('カスタムオプションを正しく適用する', async () => {
      mockFetch.mockResolvedValue({
        ok: true,
        status: 200,
        json: () => Promise.resolve({ success: true })
      });

      const customOptions = {
        timeout: 5000,
        headers: { 'X-Custom': 'value' }
      };

      await apiClient.get('/test', customOptions);

      expect(mockFetch).toHaveBeenCalledWith('http://localhost:8080/api/test', {
        method: 'GET',
        headers: {
          'Content-Type': 'application/json',
          'X-Custom': 'value'
        },
        timeout: 5000
      });
    });
  });
});
// 統一APIクライアント
// 統一されたリクエスト・レスポンス処理とエラーハンドリング

import { ErrorCode } from './types.js';
import { handleAPIError, handleNetworkError } from '../utils/error-response-handler.js';

/**
 * 統一APIクライアント
 * 全てのAPI呼び出しを統一された形式で処理
 */
export class UnifiedAPIClient {
  constructor(baseURL = import.meta.env.VITE_API_BASE_URL || '/api') {
    this.baseURL = baseURL;
    this.version = 'v1';
    this.token = null;
    this.defaultTimeout = 30000;
  }

  // トークン管理
  setToken(token) {
    this.token = token;
  }

  getToken() {
    return this.token;
  }

  // 共通ヘッダー生成
  getHeaders(customHeaders = {}) {
    const headers = {
      'Content-Type': 'application/json',
      'Accept': 'application/json',
      ...customHeaders
    };

    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    return headers;
  }

  // リクエストID生成
  generateRequestId() {
    return `req_${Date.now()}_${Math.random().toString(36).substring(2, 11)}`;
  }

  // AbortSignal作成
  createAbortSignal(timeout) {
    if (typeof AbortController === 'undefined') {
      return undefined;
    }

    const controller = new AbortController();
    setTimeout(() => controller.abort(), timeout);
    return controller.signal;
  }

  // 基本リクエストメソッド
  async request(method, endpoint, data, options = {}) {
    const url = `${this.baseURL}/v1${endpoint}`;
    const requestId = this.generateRequestId();
    
    const config = {
      method,
      headers: this.getHeaders(options.headers),
      signal: options.signal || this.createAbortSignal(options.timeout || this.defaultTimeout),
    };

    if (data && ['POST', 'PUT', 'PATCH'].includes(method)) {
      config.body = JSON.stringify(data);
    }

    try {
      const response = await fetch(url, config);
      return await this.handleResponse(response, requestId);
    } catch (error) {
      return this.handleError(error, requestId);
    }
  }

  // ページネーション付きGETリクエスト
  async requestPaginated(endpoint, params = {}, options = {}) {
    const searchParams = new URLSearchParams();
    
    // ページネーションパラメータを追加
    if (params.page) {
      searchParams.set('page', params.page.toString());
    }
    if (params.page_size) {
      searchParams.set('page_size', params.page_size.toString());
    }
    
    // その他のパラメータを追加
    Object.entries(params).forEach(([key, value]) => {
      if (key !== 'page' && key !== 'page_size' && value !== undefined && value !== null) {
        searchParams.set(key, value.toString());
      }
    });

    const queryString = searchParams.toString();
    const fullEndpoint = queryString ? `${endpoint}?${queryString}` : endpoint;
    
    return this.request('GET', fullEndpoint, null, options);
  }

  // レスポンス処理
  async handleResponse(response, requestId) {
    const contentType = response.headers.get('content-type');
    
    try {
      let responseData;
      
      if (contentType?.includes('application/json')) {
        responseData = await response.json();
      } else {
        responseData = await response.text();
      }

      // 統一されたレスポンス形式の場合
      if (responseData && typeof responseData === 'object' && 'success' in responseData) {
        return responseData;
      }

      // 旧形式のレスポンスを新形式に変換
      if (response.ok) {
        return {
          success: true,
          data: responseData,
          message: 'リクエストが成功しました',
          code: response.status,
          timestamp: new Date().toISOString(),
          request_id: requestId
        };
      } else {
        // エラーレスポンスを統一エラーハンドラーで処理
        const errorResponse = handleAPIError({
          error: this.mapStatusToErrorCode(response.status),
          message: responseData.message || `HTTP ${response.status}: ${response.statusText}`,
          code: response.status,
          timestamp: new Date().toISOString(),
          request_id: requestId,
          errors: responseData.errors // フィールド別エラーがある場合
        }, {
          httpStatus: response.status,
          endpoint: response.url
        });

        return {
          success: false,
          error: errorResponse.errors[0]?.code || this.mapStatusToErrorCode(response.status),
          message: errorResponse.errors[0]?.userMessage || responseData.message || `HTTP ${response.status}: ${response.statusText}`,
          code: response.status,
          timestamp: new Date().toISOString(),
          request_id: requestId,
          errorDetails: errorResponse
        };
      }
    } catch (parseError) {
      return {
        success: false,
        error: ErrorCode.SYSTEM_UNKNOWN_ERROR,
        message: 'レスポンスの解析に失敗しました',
        code: response.status,
        timestamp: new Date().toISOString(),
        request_id: requestId
      };
    }
  }

  // エラー処理
  handleError(error, requestId) {
    console.error('API Client Error:', error);
    
    // 統一エラーハンドラーを使用
    const errorResponse = handleNetworkError(error, {
      requestId,
      timestamp: new Date().toISOString()
    });

    // 統一形式のレスポンスに変換
    const firstError = errorResponse.errors[0];
    
    return {
      success: false,
      error: firstError?.code || ErrorCode.SYSTEM_UNKNOWN_ERROR,
      message: firstError?.userMessage || '予期しないエラーが発生しました',
      code: 0,
      timestamp: new Date().toISOString(),
      request_id: requestId,
      errorDetails: errorResponse
    };
  }

  // HTTPステータスコードをエラーコードにマッピング
  mapStatusToErrorCode(status) {
    switch (status) {
      case 400:
        return ErrorCode.VALIDATION_INVALID_FORMAT;
      case 401:
        return ErrorCode.AUTH_UNAUTHORIZED;
      case 403:
        return ErrorCode.AUTH_FORBIDDEN;
      case 404:
        return ErrorCode.RESOURCE_NOT_FOUND;
      case 409:
        return ErrorCode.RESOURCE_CONFLICT;
      case 500:
        return ErrorCode.SYSTEM_DATABASE_ERROR;
      default:
        return ErrorCode.SYSTEM_UNKNOWN_ERROR;
    }
  }

  // HTTP メソッド別のヘルパーメソッド
  async get(endpoint, options) {
    return this.request('GET', endpoint, undefined, options);
  }

  async post(endpoint, data, options) {
    return this.request('POST', endpoint, data, options);
  }

  async put(endpoint, data, options) {
    return this.request('PUT', endpoint, data, options);
  }

  async patch(endpoint, data, options) {
    return this.request('PATCH', endpoint, data, options);
  }

  async delete(endpoint, options) {
    return this.request('DELETE', endpoint, undefined, options);
  }
}

/**
 * 認証API
 */
export class AuthAPI {
  constructor(client) {
    this.client = client;
  }

  async login(credentials) {
    return this.client.post('/auth/login', credentials);
  }

  async logout() {
    return this.client.post('/auth/logout');
  }

  async refresh(refreshToken) {
    return this.client.post('/auth/refresh', { refresh_token: refreshToken });
  }

  async validate() {
    return this.client.get('/auth/validate');
  }
}

/**
 * トーナメントAPI
 */
export class TournamentAPI {
  constructor(client) {
    this.client = client;
  }

  async getAll() {
    return this.client.get('/tournaments');
  }

  // ページネーション付きトーナメント一覧取得
  async getAllPaginated(params = {}) {
    return this.client.requestPaginated('/tournaments', params);
  }

  // フィルター付きページネーション
  async getByFilterPaginated(filters = {}, paginationParams = {}) {
    const params = { ...filters, ...paginationParams };
    return this.client.requestPaginated('/tournaments', params);
  }

  async getBySport(sport) {
    return this.client.get(`/tournaments/${sport}`);
  }

  async getBracket(sport) {
    return this.client.get(`/tournaments/${sport}/bracket`);
  }

  async updateFormat(sport, format) {
    return this.client.put(`/tournaments/${sport}/format`, { format });
  }

  async create(data) {
    return this.client.post('/tournaments', data);
  }

  async update(id, data) {
    return this.client.put(`/tournaments/${id}`, data);
  }

  async delete(id) {
    return this.client.delete(`/tournaments/${id}`);
  }
}

/**
 * 試合API
 */
export class MatchAPI {
  constructor(client) {
    this.client = client;
  }

  async getAll(filters) {
    const queryParams = new URLSearchParams();
    
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams.append(key, value.toString());
        }
      });
    }

    const queryString = queryParams.toString();
    const endpoint = `/matches${queryString ? `?${queryString}` : ''}`;
    
    return this.client.get(endpoint);
  }

  // ページネーション付き試合一覧取得
  async getAllPaginated(params = {}) {
    return this.client.requestPaginated('/matches', params);
  }

  // スポーツ別ページネーション付き試合取得
  async getBySportPaginated(sport, params = {}) {
    return this.client.requestPaginated(`/matches/${sport}`, params);
  }

  // フィルター付きページネーション
  async getByFilterPaginated(filters = {}, paginationParams = {}) {
    const params = { ...filters, ...paginationParams };
    return this.client.requestPaginated('/matches', params);
  }

  async getBySport(sport, filters) {
    const queryParams = new URLSearchParams();
    
    if (filters) {
      Object.entries(filters).forEach(([key, value]) => {
        if (value !== undefined) {
          queryParams.append(key, value.toString());
        }
      });
    }

    const queryString = queryParams.toString();
    const endpoint = `/matches/${sport}${queryString ? `?${queryString}` : ''}`;
    
    return this.client.get(endpoint);
  }

  async getById(id) {
    return this.client.get(`/matches/${id}`);
  }

  async create(data) {
    return this.client.post('/matches', data);
  }

  async update(id, data) {
    return this.client.put(`/matches/${id}`, data);
  }

  async updateResult(id, result) {
    return this.client.put(`/matches/${id}/result`, result);
  }

  async delete(id) {
    return this.client.delete(`/matches/${id}`);
  }
}

/**
 * 統一APIクライアントのメインクラス
 * 全てのAPIエンドポイントへのアクセスを提供
 */
export class UnifiedAPI {
  constructor(baseURL) {
    this.client = new UnifiedAPIClient(baseURL);
    
    this.auth = new AuthAPI(this.client);
    this.tournaments = new TournamentAPI(this.client);
    this.matches = new MatchAPI(this.client);
  }

  // トークン管理の委譲
  setToken(token) {
    this.client.setToken(token);
  }

  getToken() {
    return this.client.getToken();
  }

  // 直接的なリクエストアクセス（必要に応じて）
  async request(method, endpoint, data, options) {
    return this.client.request(method, endpoint, data, options);
  }
}

// デフォルトインスタンス
export const unifiedAPI = new UnifiedAPI();

// 後方互換性のためのエクスポート
export default UnifiedAPI;
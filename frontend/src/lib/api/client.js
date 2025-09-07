// HTTPクライアント設定
import { handleApiError, handleNetworkError, retry, AppError, ERROR_TYPES, ERROR_LEVELS } from '$lib/utils/error-handler.js';

class APIClient {
  constructor(baseURL = import.meta.env.VITE_API_BASE_URL || '/api') {
    this.baseURL = baseURL;
    this.token = null;
    this.requestInterceptors = [];
    this.responseInterceptors = [];
    this.defaultTimeout = 30000; // 30秒
    this.retryConfig = {
      maxAttempts: 3,
      retryableErrors: [ERROR_TYPES.NETWORK, ERROR_TYPES.TIMEOUT],
      retryDelay: 1000
    };
  }

  // トークン設定
  setToken(token) {
    this.token = token;
  }

  // トークン取得
  getToken() {
    return this.token;
  }

  // リクエストインターセプターの追加
  addRequestInterceptor(interceptor) {
    this.requestInterceptors.push(interceptor);
  }

  // レスポンスインターセプターの追加
  addResponseInterceptor(interceptor) {
    this.responseInterceptors.push(interceptor);
  }

  // 共通リクエストヘッダーの取得
  getHeaders(customHeaders = {}) {
    const headers = {
      'Content-Type': 'application/json',
      ...customHeaders
    };

    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    return headers;
  }

  // リクエスト前処理（インターセプター適用）
  async applyRequestInterceptors(config) {
    let processedConfig = { ...config };
    
    for (const interceptor of this.requestInterceptors) {
      try {
        processedConfig = await interceptor(processedConfig);
      } catch (error) {
        console.error('Request interceptor error:', error);
      }
    }
    
    return processedConfig;
  }

  // レスポンス後処理（インターセプター適用）
  async applyResponseInterceptors(response) {
    let processedResponse = response;
    
    for (const interceptor of this.responseInterceptors) {
      try {
        processedResponse = await interceptor(processedResponse);
      } catch (error) {
        console.error('Response interceptor error:', error);
      }
    }
    
    return processedResponse;
  }

  // 基本リクエスト処理
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const requestId = `${options.method || 'GET'}-${endpoint}-${Date.now()}`;
    
    // デフォルト設定
    const config = {
      method: 'GET',
      headers: this.getHeaders(options.headers),
      signal: this.createAbortSignal(options.timeout || this.defaultTimeout),
      ...options
    };

    // リクエストインターセプター適用
    const processedConfig = await this.applyRequestInterceptors(config);

    // 再試行可能なリクエストの場合は retry を使用
    const shouldRetry = this.shouldRetryRequest(options);
    
    if (shouldRetry) {
      return retry(
        () => this.executeRequest(url, processedConfig),
        requestId,
        this.retryConfig.maxAttempts
      );
    } else {
      return this.executeRequest(url, processedConfig);
    }
  }

  // リクエスト実行
  async executeRequest(url, config) {
    try {
      const response = await fetch(url, config);
      
      // レスポンスインターセプター適用
      const processedResponse = await this.applyResponseInterceptors(response);
      
      return await this.handleResponse(processedResponse);
    } catch (error) {
      return this.handleError(error);
    }
  }

  // AbortSignal の作成（タイムアウト対応）
  createAbortSignal(timeout) {
    if (typeof AbortController === 'undefined') {
      return undefined;
    }

    const controller = new AbortController();
    
    if (timeout > 0) {
      setTimeout(() => {
        controller.abort();
      }, timeout);
    }
    
    return controller.signal;
  }

  // 再試行判定
  shouldRetryRequest(options) {
    // 明示的に再試行を無効にしている場合
    if (options.retry === false) {
      return false;
    }

    // GET リクエストは再試行可能
    const method = (options.method || 'GET').toUpperCase();
    if (method === 'GET' || method === 'HEAD') {
      return true;
    }

    // その他のメソッドは明示的に再試行を有効にした場合のみ
    return options.retry === true;
  }

  // GET リクエスト
  async get(endpoint, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'GET'
    });
  }

  // POST リクエスト
  async post(endpoint, data, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'POST',
      body: data ? JSON.stringify(data) : undefined
    });
  }

  // PUT リクエスト
  async put(endpoint, data, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'PUT',
      body: data ? JSON.stringify(data) : undefined
    });
  }

  // DELETE リクエスト
  async delete(endpoint, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'DELETE'
    });
  }

  // PATCH リクエスト
  async patch(endpoint, data, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: 'PATCH',
      body: data ? JSON.stringify(data) : undefined
    });
  }

  // レスポンス処理
  async handleResponse(response) {
    const contentType = response.headers.get('content-type');
    
    try {
      let data;
      
      if (contentType && contentType.includes('application/json')) {
        data = await response.json();
      } else {
        data = await response.text();
      }

      if (response.ok) {
        return {
          success: true,
          data: data,
          status: response.status,
          statusText: response.statusText
        };
      } else {
        // HTTPエラーステータスの場合 - 新しいエラーハンドラーを使用
        const errorResponse = {
          success: false,
          error: data.error || 'HTTP_ERROR',
          message: data.message || `HTTP ${response.status}: ${response.statusText}`,
          status: response.status,
          statusText: response.statusText,
          details: data.details || null
        };

        // エラーハンドラーに処理を委譲
        handleApiError(errorResponse);
        
        return errorResponse;
      }
    } catch (parseError) {
      // JSONパースエラーなど
      const errorResponse = {
        success: false,
        error: 'PARSE_ERROR',
        message: 'レスポンスの解析に失敗しました',
        status: response.status,
        statusText: response.statusText,
        details: parseError.message
      };

      // パースエラーとして処理
      const appError = new AppError(
        errorResponse.message,
        ERROR_TYPES.PARSE,
        ERROR_LEVELS.MEDIUM,
        errorResponse.details
      );
      
      handleApiError(errorResponse);
      
      return errorResponse;
    }
  }

  // エラー処理
  handleError(error) {
    console.error('API Client Error:', error);
    
    let errorResponse;
    
    // ネットワークエラーやその他のfetchエラー
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      errorResponse = {
        success: false,
        error: 'NETWORK_ERROR',
        message: 'ネットワークエラーが発生しました。接続を確認してください。',
        details: error.message
      };
      
      // ネットワークエラーハンドラーを使用
      handleNetworkError(error);
    }
    // タイムアウトエラー
    else if (error.name === 'AbortError') {
      errorResponse = {
        success: false,
        error: 'TIMEOUT_ERROR',
        message: 'リクエストがタイムアウトしました。',
        details: error.message
      };
      
      // タイムアウトエラーとして処理
      const timeoutError = new AppError(
        errorResponse.message,
        ERROR_TYPES.TIMEOUT,
        ERROR_LEVELS.MEDIUM,
        { originalError: error }
      );
      
      handleApiError(errorResponse);
    }
    // その他のエラー
    else {
      errorResponse = {
        success: false,
        error: 'UNKNOWN_ERROR',
        message: '予期しないエラーが発生しました。',
        details: error.message
      };
      
      // 不明なエラーとして処理
      const unknownError = new AppError(
        errorResponse.message,
        ERROR_TYPES.UNKNOWN,
        ERROR_LEVELS.HIGH,
        { originalError: error }
      );
      
      handleApiError(errorResponse);
    }
    
    return errorResponse;
  }

  // ヘルスチェック
  async healthCheck() {
    try {
      const response = await this.get('/health');
      return response.success;
    } catch (error) {
      return false;
    }
  }
}

// デフォルトのAPIクライアントインスタンス
export const apiClient = new APIClient();

// 共通のリクエストインターセプター設定
apiClient.addRequestInterceptor(async (config) => {
  // リクエストログ出力（開発環境のみ）
  if (import.meta.env.DEV) {
    console.log(`[API Request] ${config.method} ${config.url || 'Unknown URL'}`, config);
  }
  return config;
});

// 共通のレスポンスインターセプター設定
apiClient.addResponseInterceptor(async (response) => {
  // レスポンスログ出力（開発環境のみ）
  if (import.meta.env.DEV) {
    console.log(`[API Response] ${response.status} ${response.statusText}`, response);
  }
  
  // 401エラー（認証エラー）の場合の共通処理
  if (response.status === 401) {
    // トークンをクリア
    apiClient.setToken(null);
    
    // ローカルストレージからトークンを削除
    if (typeof window !== 'undefined') {
      localStorage.removeItem('auth_token');
    }
    
    // カスタムイベントを発火して認証状態の変更を通知
    if (typeof window !== 'undefined') {
      window.dispatchEvent(new CustomEvent('auth:unauthorized'));
    }
  }
  
  return response;
});

export default APIClient;

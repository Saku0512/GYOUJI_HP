// 統一APIクライアント
// 統一されたリクエスト・レスポンス処理とエラーハンドリング

import type {
  APIResponse,
  PaginatedAPIResponse,
  HTTPMethod,
  RequestOptions,
  SportType,
  Tournament,
  TournamentBracket,
  CreateTournamentRequest,
  UpdateTournamentRequest,
  Match,
  CreateMatchRequest,
  UpdateMatchRequest,
  MatchResult,
  MatchFilters,
  LoginRequest,
  AuthResponse,
  User
} from './types.js';

import { ErrorCode } from './types.js';

/**
 * 統一APIクライアント
 * 全てのAPI呼び出しを統一された形式で処理
 */
export class UnifiedAPIClient {
  private baseURL: string;
  private version: string = 'v1';
  private token: string | null = null;
  private defaultTimeout: number = 30000;

  constructor(baseURL: string = import.meta.env.VITE_API_BASE_URL || '/api') {
    this.baseURL = baseURL;
  }

  // トークン管理
  setToken(token: string | null): void {
    this.token = token;
  }

  getToken(): string | null {
    return this.token;
  }

  // 共通ヘッダー生成
  private getHeaders(customHeaders: Record<string, string> = {}): Record<string, string> {
    const headers: Record<string, string> = {
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
  private generateRequestId(): string {
    return `req_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  // AbortSignal作成
  private createAbortSignal(timeout: number): AbortSignal | undefined {
    if (typeof AbortController === 'undefined') {
      return undefined;
    }

    const controller = new AbortController();
    setTimeout(() => controller.abort(), timeout);
    return controller.signal;
  }

  // 基本リクエストメソッド
  async request<T = any>(
    method: HTTPMethod,
    endpoint: string,
    data?: any,
    options: RequestOptions = {}
  ): Promise<APIResponse<T>> {
    const url = `${this.baseURL}/v1${endpoint}`;
    const requestId = this.generateRequestId();
    
    const config: RequestInit = {
      method,
      headers: this.getHeaders(options.headers),
      signal: options.signal || this.createAbortSignal(options.timeout || this.defaultTimeout),
    };

    if (data && ['POST', 'PUT', 'PATCH'].includes(method)) {
      config.body = JSON.stringify(data);
    }

    try {
      const response = await fetch(url, config);
      return await this.handleResponse<T>(response, requestId);
    } catch (error) {
      return this.handleError(error, requestId);
    }
  }

  // レスポンス処理
  private async handleResponse<T>(response: Response, requestId: string): Promise<APIResponse<T>> {
    const contentType = response.headers.get('content-type');
    
    try {
      let responseData: any;
      
      if (contentType?.includes('application/json')) {
        responseData = await response.json();
      } else {
        responseData = await response.text();
      }

      // 統一されたレスポンス形式の場合
      if (responseData && typeof responseData === 'object' && 'success' in responseData) {
        return responseData as APIResponse<T>;
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
        return {
          success: false,
          error: this.mapStatusToErrorCode(response.status),
          message: responseData.message || `HTTP ${response.status}: ${response.statusText}`,
          code: response.status,
          timestamp: new Date().toISOString(),
          request_id: requestId
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
  private handleError(error: any, requestId: string): APIResponse<any> {
    console.error('API Client Error:', error);
    
    let errorCode: string;
    let message: string;

    if (error.name === 'AbortError') {
      errorCode = ErrorCode.SYSTEM_TIMEOUT;
      message = 'リクエストがタイムアウトしました';
    } else if (error.name === 'TypeError' && error.message.includes('fetch')) {
      errorCode = ErrorCode.SYSTEM_NETWORK_ERROR;
      message = 'ネットワークエラーが発生しました';
    } else {
      errorCode = ErrorCode.SYSTEM_UNKNOWN_ERROR;
      message = '予期しないエラーが発生しました';
    }

    return {
      success: false,
      error: errorCode,
      message,
      code: 0,
      timestamp: new Date().toISOString(),
      request_id: requestId
    };
  }

  // HTTPステータスコードをエラーコードにマッピング
  private mapStatusToErrorCode(status: number): string {
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
  async get<T = any>(endpoint: string, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('GET', endpoint, undefined, options);
  }

  async post<T = any>(endpoint: string, data?: any, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('POST', endpoint, data, options);
  }

  async put<T = any>(endpoint: string, data?: any, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('PUT', endpoint, data, options);
  }

  async patch<T = any>(endpoint: string, data?: any, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('PATCH', endpoint, data, options);
  }

  async delete<T = any>(endpoint: string, options?: RequestOptions): Promise<APIResponse<T>> {
    return this.request<T>('DELETE', endpoint, undefined, options);
  }
}

/**
 * 認証API
 */
export class AuthAPI {
  constructor(private client: UnifiedAPIClient) {}

  async login(credentials: LoginRequest): Promise<APIResponse<AuthResponse>> {
    return this.client.post<AuthResponse>('/auth/login', credentials);
  }

  async logout(): Promise<APIResponse<void>> {
    return this.client.post<void>('/auth/logout');
  }

  async refresh(refreshToken: string): Promise<APIResponse<AuthResponse>> {
    return this.client.post<AuthResponse>('/auth/refresh', { refresh_token: refreshToken });
  }

  async validate(): Promise<APIResponse<User>> {
    return this.client.get<User>('/auth/validate');
  }
}

/**
 * トーナメントAPI
 */
export class TournamentAPI {
  constructor(private client: UnifiedAPIClient) {}

  async getAll(): Promise<APIResponse<Tournament[]>> {
    return this.client.get<Tournament[]>('/tournaments');
  }

  async getBySport(sport: SportType): Promise<APIResponse<Tournament>> {
    return this.client.get<Tournament>(`/tournaments/${sport}`);
  }

  async getBracket(sport: SportType): Promise<APIResponse<TournamentBracket>> {
    return this.client.get<TournamentBracket>(`/tournaments/${sport}/bracket`);
  }

  async updateFormat(sport: SportType, format: string): Promise<APIResponse<Tournament>> {
    return this.client.put<Tournament>(`/tournaments/${sport}/format`, { format });
  }

  async create(data: CreateTournamentRequest): Promise<APIResponse<Tournament>> {
    return this.client.post<Tournament>('/tournaments', data);
  }

  async update(id: number, data: UpdateTournamentRequest): Promise<APIResponse<Tournament>> {
    return this.client.put<Tournament>(`/tournaments/${id}`, data);
  }

  async delete(id: number): Promise<APIResponse<void>> {
    return this.client.delete<void>(`/tournaments/${id}`);
  }
}

/**
 * 試合API
 */
export class MatchAPI {
  constructor(private client: UnifiedAPIClient) {}

  async getAll(filters?: MatchFilters): Promise<APIResponse<Match[]>> {
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
    
    return this.client.get<Match[]>(endpoint);
  }

  async getBySport(sport: SportType, filters?: MatchFilters): Promise<APIResponse<Match[]>> {
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
    
    return this.client.get<Match[]>(endpoint);
  }

  async getById(id: number): Promise<APIResponse<Match>> {
    return this.client.get<Match>(`/matches/${id}`);
  }

  async create(data: CreateMatchRequest): Promise<APIResponse<Match>> {
    return this.client.post<Match>('/matches', data);
  }

  async update(id: number, data: UpdateMatchRequest): Promise<APIResponse<Match>> {
    return this.client.put<Match>(`/matches/${id}`, data);
  }

  async updateResult(id: number, result: MatchResult): Promise<APIResponse<Match>> {
    return this.client.put<Match>(`/matches/${id}/result`, result);
  }

  async delete(id: number): Promise<APIResponse<void>> {
    return this.client.delete<void>(`/matches/${id}`);
  }
}

/**
 * 統一APIクライアントのメインクラス
 * 全てのAPIエンドポイントへのアクセスを提供
 */
export class UnifiedAPI {
  private client: UnifiedAPIClient;
  
  public readonly auth: AuthAPI;
  public readonly tournaments: TournamentAPI;
  public readonly matches: MatchAPI;

  constructor(baseURL?: string) {
    this.client = new UnifiedAPIClient(baseURL);
    
    this.auth = new AuthAPI(this.client);
    this.tournaments = new TournamentAPI(this.client);
    this.matches = new MatchAPI(this.client);
  }

  // トークン管理の委譲
  setToken(token: string | null): void {
    this.client.setToken(token);
  }

  getToken(): string | null {
    return this.client.getToken();
  }

  // 直接的なリクエストアクセス（必要に応じて）
  async request<T = any>(
    method: HTTPMethod,
    endpoint: string,
    data?: any,
    options?: RequestOptions
  ): Promise<APIResponse<T>> {
    return this.client.request<T>(method, endpoint, data, options);
  }
}

// デフォルトインスタンス
export const unifiedAPI = new UnifiedAPI();

// 後方互換性のためのエクスポート
export default UnifiedAPI;
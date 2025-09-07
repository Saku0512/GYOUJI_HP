class APIClient {
  constructor(baseURL = "http://localhost:8080/api") {
    this.baseURL = baseURL;
    this.token = null;
    this.requestInterceptors = [];
    this.responseInterceptors = [];
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
      "Content-Type": "application/json",
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
        console.error("Request interceptor error:", error);
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
        console.error("Response interceptor error:", error);
      }
    }
    return processedResponse;
  }
  // 基本リクエスト処理
  async request(endpoint, options = {}) {
    const url = `${this.baseURL}${endpoint}`;
    const config = {
      method: "GET",
      headers: this.getHeaders(options.headers),
      ...options
    };
    const processedConfig = await this.applyRequestInterceptors(config);
    try {
      const response = await fetch(url, processedConfig);
      const processedResponse = await this.applyResponseInterceptors(response);
      return await this.handleResponse(processedResponse);
    } catch (error) {
      return this.handleError(error);
    }
  }
  // GET リクエスト
  async get(endpoint, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: "GET"
    });
  }
  // POST リクエスト
  async post(endpoint, data, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: "POST",
      body: data ? JSON.stringify(data) : void 0
    });
  }
  // PUT リクエスト
  async put(endpoint, data, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: "PUT",
      body: data ? JSON.stringify(data) : void 0
    });
  }
  // DELETE リクエスト
  async delete(endpoint, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: "DELETE"
    });
  }
  // PATCH リクエスト
  async patch(endpoint, data, options = {}) {
    return this.request(endpoint, {
      ...options,
      method: "PATCH",
      body: data ? JSON.stringify(data) : void 0
    });
  }
  // レスポンス処理
  async handleResponse(response) {
    const contentType = response.headers.get("content-type");
    try {
      let data;
      if (contentType && contentType.includes("application/json")) {
        data = await response.json();
      } else {
        data = await response.text();
      }
      if (response.ok) {
        return {
          success: true,
          data,
          status: response.status,
          statusText: response.statusText
        };
      } else {
        return {
          success: false,
          error: data.error || "HTTP_ERROR",
          message: data.message || `HTTP ${response.status}: ${response.statusText}`,
          status: response.status,
          statusText: response.statusText,
          details: data.details || null
        };
      }
    } catch (parseError) {
      return {
        success: false,
        error: "PARSE_ERROR",
        message: "レスポンスの解析に失敗しました",
        status: response.status,
        statusText: response.statusText,
        details: parseError.message
      };
    }
  }
  // エラー処理
  handleError(error) {
    console.error("API Client Error:", error);
    if (error.name === "TypeError" && error.message.includes("fetch")) {
      return {
        success: false,
        error: "NETWORK_ERROR",
        message: "ネットワークエラーが発生しました。接続を確認してください。",
        details: error.message
      };
    }
    if (error.name === "AbortError") {
      return {
        success: false,
        error: "TIMEOUT_ERROR",
        message: "リクエストがタイムアウトしました。",
        details: error.message
      };
    }
    return {
      success: false,
      error: "UNKNOWN_ERROR",
      message: "予期しないエラーが発生しました。",
      details: error.message
    };
  }
  // ヘルスチェック
  async healthCheck() {
    try {
      const response = await this.get("/health");
      return response.success;
    } catch (error) {
      return false;
    }
  }
}
const apiClient = new APIClient();
apiClient.addRequestInterceptor(async (config) => {
  return config;
});
apiClient.addResponseInterceptor(async (response) => {
  if (response.status === 401) {
    apiClient.setToken(null);
    if (typeof window !== "undefined") {
      localStorage.removeItem("auth_token");
    }
    if (typeof window !== "undefined") {
      window.dispatchEvent(new CustomEvent("auth:unauthorized"));
    }
  }
  return response;
});
export {
  apiClient as a
};

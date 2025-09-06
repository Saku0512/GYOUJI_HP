import { w as writable } from "./index.js";
import { E as fallback, F as attr_class, U as attr_style, I as bind_props, J as stringify } from "./index2.js";
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
const initialUIState = {
  notifications: [],
  loading: false,
  theme: "light"
};
const uiStore = writable(initialUIState);
let notificationIdCounter = 0;
const uiActions = {
  /**
   * 通知を表示する
   * @param {string} message - 通知メッセージ
   * @param {string} type - 通知タイプ ('success', 'error', 'warning', 'info')
   * @param {number} duration - 自動消去までの時間（ミリ秒）、0の場合は自動消去しない
   */
  showNotification: (message, type = "info", duration = 5e3) => {
    const notification = {
      id: ++notificationIdCounter,
      message,
      type,
      timestamp: Date.now()
    };
    uiStore.update((state) => ({
      ...state,
      notifications: [...state.notifications, notification]
    }));
    if (duration > 0) {
      setTimeout(() => {
        uiActions.removeNotification(notification.id);
      }, duration);
    }
    return notification.id;
  },
  /**
   * 特定の通知を削除する
   * @param {number} id - 削除する通知のID
   */
  removeNotification: (id) => {
    uiStore.update((state) => ({
      ...state,
      notifications: state.notifications.filter((notification) => notification.id !== id)
    }));
  },
  /**
   * ローディング状態を設定する
   * @param {boolean} state - ローディング状態
   */
  setLoading: (state) => {
    uiStore.update((currentState) => ({
      ...currentState,
      loading: Boolean(state)
    }));
  },
  /**
   * 全ての通知をクリアする
   */
  clearNotifications: () => {
    uiStore.update((state) => ({
      ...state,
      notifications: []
    }));
  },
  /**
   * テーマを設定する
   * @param {string} theme - テーマ ('light', 'dark')
   */
  setTheme: (theme) => {
    if (theme !== "light" && theme !== "dark") {
      console.warn('Invalid theme. Only "light" and "dark" are supported.');
      return;
    }
    uiStore.update((state) => ({
      ...state,
      theme
    }));
    if (typeof localStorage !== "undefined") {
      localStorage.setItem("ui-theme", theme);
    }
  },
  /**
   * ローカルストレージからテーマを読み込む
   */
  loadTheme: () => {
    if (typeof localStorage !== "undefined") {
      const savedTheme = localStorage.getItem("ui-theme");
      if (savedTheme && (savedTheme === "light" || savedTheme === "dark")) {
        uiActions.setTheme(savedTheme);
      }
    }
  },
  /**
   * UIストアを初期状態にリセットする
   */
  reset: () => {
    uiStore.set(initialUIState);
  }
};
function LoadingSpinner($$payload, $$props) {
  let size = fallback(
    $$props["size"],
    "medium"
    // 'small', 'medium', 'large'
  );
  let color = fallback($$props["color"], "#007bff");
  $$payload.out.push(`<div class="spinner-container svelte-1yjjzjh"><div${attr_class(`spinner ${stringify(size)}`, "svelte-1yjjzjh")}${attr_style(`border-top-color: ${stringify(color)}`)}></div></div>`);
  bind_props($$props, { size, color });
}
export {
  LoadingSpinner as L,
  apiClient as a,
  uiActions as b,
  uiStore as u
};

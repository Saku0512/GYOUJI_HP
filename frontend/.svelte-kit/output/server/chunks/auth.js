import { w as writable, g as get } from "./index.js";
import { a as apiClient } from "./client.js";
import { g as getAuthToken, a as getUserData, s as saveAuthToken, b as saveUserData, c as clearAuthData } from "./storage.js";
class AuthAPI {
  constructor(client = apiClient) {
    this.client = client;
    this.tokenKey = "auth_token";
    this.refreshTokenKey = "refresh_token";
    this.userKey = "auth_user";
  }
  /**
   * ローカルストレージからトークンを取得
   */
  getStoredToken() {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(this.tokenKey);
  }
  /**
   * ローカルストレージからリフレッシュトークンを取得
   */
  getStoredRefreshToken() {
    if (typeof window === "undefined") return null;
    return localStorage.getItem(this.refreshTokenKey);
  }
  /**
   * ローカルストレージからユーザー情報を取得
   */
  getStoredUser() {
    if (typeof window === "undefined") return null;
    const userStr = localStorage.getItem(this.userKey);
    try {
      return userStr ? JSON.parse(userStr) : null;
    } catch (error) {
      console.error("Failed to parse stored user data:", error);
      return null;
    }
  }
  /**
   * トークンをローカルストレージに保存
   */
  storeToken(token, refreshToken = null, user = null) {
    if (typeof window === "undefined") return;
    localStorage.setItem(this.tokenKey, token);
    if (refreshToken) {
      localStorage.setItem(this.refreshTokenKey, refreshToken);
    }
    if (user) {
      localStorage.setItem(this.userKey, JSON.stringify(user));
    }
    this.client.setToken(token);
  }
  /**
   * 認証情報をローカルストレージから削除
   */
  clearStoredAuth() {
    if (typeof window === "undefined") return;
    localStorage.removeItem(this.tokenKey);
    localStorage.removeItem(this.refreshTokenKey);
    localStorage.removeItem(this.userKey);
    this.client.setToken(null);
  }
  /**
   * JWTトークンの有効期限をチェック
   */
  isTokenExpired(token) {
    if (!token) return true;
    try {
      const payload = JSON.parse(atob(token.split(".")[1]));
      const currentTime = Math.floor(Date.now() / 1e3);
      return payload.exp < currentTime;
    } catch (error) {
      console.error("Failed to parse JWT token:", error);
      return true;
    }
  }
  /**
   * 初期化時にストレージからトークンを復元
   */
  initializeAuth() {
    const token = this.getStoredToken();
    if (token && !this.isTokenExpired(token)) {
      this.client.setToken(token);
      return true;
    } else if (token) {
      this.clearStoredAuth();
    }
    return false;
  }
  /**
   * ログイン
   */
  async login(username, password) {
    try {
      const response = await this.client.post("/auth/login", {
        username,
        password
      });
      if (response.success && response.data) {
        const { token, refresh_token, user } = response.data;
        this.storeToken(token, refresh_token, user);
        return {
          success: true,
          data: {
            token,
            refresh_token,
            user
          },
          message: "ログインに成功しました"
        };
      }
      return response;
    } catch (error) {
      console.error("Login error:", error);
      return {
        success: false,
        error: "LOGIN_ERROR",
        message: "ログインに失敗しました",
        details: error.message
      };
    }
  }
  /**
   * ログアウト
   */
  async logout() {
    try {
      const token = this.getStoredToken();
      if (token) {
        await this.client.post("/auth/logout");
      }
      this.clearStoredAuth();
      if (typeof window !== "undefined") {
        window.dispatchEvent(new CustomEvent("auth:logout"));
      }
      return {
        success: true,
        message: "ログアウトしました"
      };
    } catch (error) {
      console.error("Logout error:", error);
      this.clearStoredAuth();
      return {
        success: true,
        message: "ログアウトしました",
        warning: "サーバーとの通信でエラーが発生しましたが、ローカルの認証情報はクリアされました"
      };
    }
  }
  /**
   * トークンリフレッシュ
   */
  async refreshToken() {
    try {
      const refreshToken = this.getStoredRefreshToken();
      if (!refreshToken) {
        return {
          success: false,
          error: "NO_REFRESH_TOKEN",
          message: "リフレッシュトークンが見つかりません"
        };
      }
      const response = await this.client.post("/auth/refresh", {
        refresh_token: refreshToken
      });
      if (response.success && response.data) {
        const { token, refresh_token: newRefreshToken, user } = response.data;
        this.storeToken(token, newRefreshToken || refreshToken, user);
        return {
          success: true,
          data: {
            token,
            refresh_token: newRefreshToken || refreshToken,
            user
          },
          message: "トークンを更新しました"
        };
      }
      this.clearStoredAuth();
      return response;
    } catch (error) {
      console.error("Token refresh error:", error);
      this.clearStoredAuth();
      return {
        success: false,
        error: "REFRESH_ERROR",
        message: "トークンの更新に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * トークン検証
   */
  async validateToken() {
    try {
      const token = this.getStoredToken();
      if (!token) {
        return {
          success: false,
          error: "NO_TOKEN",
          message: "トークンが見つかりません"
        };
      }
      if (this.isTokenExpired(token)) {
        const refreshResult = await this.refreshToken();
        if (refreshResult.success) {
          return {
            success: true,
            data: refreshResult.data,
            message: "トークンを更新して検証しました"
          };
        } else {
          return {
            success: false,
            error: "TOKEN_EXPIRED",
            message: "トークンの期限が切れています"
          };
        }
      }
      const response = await this.client.get("/auth/validate");
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: "トークンは有効です"
        };
      }
      this.clearStoredAuth();
      return response;
    } catch (error) {
      console.error("Token validation error:", error);
      return {
        success: false,
        error: "VALIDATION_ERROR",
        message: "トークンの検証に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 現在の認証状態を取得
   */
  getAuthState() {
    const token = this.getStoredToken();
    const user = this.getStoredUser();
    return {
      isAuthenticated: !!(token && !this.isTokenExpired(token)),
      token,
      user
    };
  }
  /**
   * 自動トークンリフレッシュの設定
   */
  setupAutoRefresh(intervalMinutes = 50) {
    if (typeof window === "undefined") return;
    const interval = intervalMinutes * 60 * 1e3;
    setInterval(async () => {
      const token = this.getStoredToken();
      if (token && !this.isTokenExpired(token)) {
        try {
          const payload = JSON.parse(atob(token.split(".")[1]));
          const currentTime = Math.floor(Date.now() / 1e3);
          const timeUntilExpiry = payload.exp - currentTime;
          if (timeUntilExpiry < 600) {
            await this.refreshToken();
          }
        } catch (error) {
          console.error("Auto refresh error:", error);
        }
      }
    }, interval);
  }
}
const authAPI = new AuthAPI();
if (typeof window !== "undefined") {
  authAPI.initializeAuth();
  authAPI.setupAutoRefresh();
}
const initialAuthState = {
  isAuthenticated: false,
  token: null,
  user: null,
  loading: false
};
function createAuthStore() {
  const { subscribe, set, update } = writable(initialAuthState);
  return {
    subscribe,
    // ローディング状態の設定
    setLoading: (loading) => {
      update((state) => ({ ...state, loading }));
    },
    // 認証状態の設定
    setAuthState: (authData) => {
      update((state) => ({
        ...state,
        isAuthenticated: !!authData.token,
        token: authData.token,
        user: authData.user,
        loading: false
      }));
    },
    // 認証状態のクリア
    clearAuth: () => {
      set({ ...initialAuthState });
    },
    // ログイン処理
    login: async (credentials) => {
      update((state) => ({ ...state, loading: true }));
      try {
        const result = await authAPI.login(credentials.username, credentials.password);
        if (result.success) {
          const authData = {
            token: result.data.token,
            user: result.data.user
          };
          saveAuthToken(result.data.token);
          saveUserData(result.data.user);
          update((state) => ({
            ...state,
            isAuthenticated: true,
            token: result.data.token,
            user: result.data.user,
            loading: false
          }));
          return result;
        } else {
          update((state) => ({ ...state, loading: false }));
          return result;
        }
      } catch (error) {
        console.error("Login error in store:", error);
        update((state) => ({ ...state, loading: false }));
        return {
          success: false,
          error: "LOGIN_STORE_ERROR",
          message: "ログイン処理でエラーが発生しました",
          details: error.message
        };
      }
    },
    // ログアウト処理
    logout: async () => {
      update((state) => ({ ...state, loading: true }));
      try {
        const result = await authAPI.logout();
        clearAuthData();
        set({ ...initialAuthState });
        return result;
      } catch (error) {
        console.error("Logout error in store:", error);
        clearAuthData();
        set({ ...initialAuthState });
        return {
          success: true,
          message: "ログアウトしました",
          warning: "サーバーとの通信でエラーが発生しましたが、ローカルの認証情報はクリアされました"
        };
      }
    },
    // 認証状態チェック
    checkAuthStatus: async () => {
      update((state) => ({ ...state, loading: true }));
      try {
        const storedToken = getAuthToken();
        const storedUser = getUserData();
        if (!storedToken) {
          update((state) => ({ ...state, loading: false }));
          return {
            success: false,
            error: "NO_TOKEN",
            message: "トークンが見つかりません"
          };
        }
        const result = await authAPI.validateToken();
        if (result.success) {
          update((state) => ({
            ...state,
            isAuthenticated: true,
            token: storedToken,
            user: result.data?.user || storedUser,
            loading: false
          }));
          return result;
        } else {
          clearAuthData();
          set({ ...initialAuthState });
          return result;
        }
      } catch (error) {
        console.error("Auth status check error in store:", error);
        clearAuthData();
        set({ ...initialAuthState });
        return {
          success: false,
          error: "AUTH_CHECK_ERROR",
          message: "認証状態の確認でエラーが発生しました",
          details: error.message
        };
      }
    },
    // トークンリフレッシュ
    refreshToken: async () => {
      const currentState = get({ subscribe });
      if (currentState.loading) {
        return {
          success: false,
          error: "ALREADY_LOADING",
          message: "既に処理中です"
        };
      }
      update((state) => ({ ...state, loading: true }));
      try {
        const result = await authAPI.refreshToken();
        if (result.success) {
          const authData = {
            token: result.data.token,
            user: result.data.user
          };
          saveAuthToken(result.data.token);
          saveUserData(result.data.user);
          update((state) => ({
            ...state,
            isAuthenticated: true,
            token: result.data.token,
            user: result.data.user,
            loading: false
          }));
          return result;
        } else {
          clearAuthData();
          set({ ...initialAuthState });
          return result;
        }
      } catch (error) {
        console.error("Token refresh error in store:", error);
        clearAuthData();
        set({ ...initialAuthState });
        return {
          success: false,
          error: "REFRESH_STORE_ERROR",
          message: "トークンの更新でエラーが発生しました",
          details: error.message
        };
      }
    },
    // 初期化処理（アプリ起動時に呼び出し）
    initialize: async function() {
      const storedToken = getAuthToken();
      const storedUser = getUserData();
      if (storedToken && storedUser) {
        await this.checkAuthStatus();
      }
    }
  };
}
const authStore = createAuthStore();
if (typeof window !== "undefined") {
  authStore.initialize();
  window.addEventListener("auth:logout", () => {
    authStore.clearAuth();
  });
  document.addEventListener("visibilitychange", () => {
    if (!document.hidden) {
      const storedToken = getAuthToken();
      if (storedToken) {
        authStore.checkAuthStatus();
      }
    }
  });
}
export {
  authStore as a
};

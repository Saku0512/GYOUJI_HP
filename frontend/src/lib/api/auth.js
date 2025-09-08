// 認証関連API呼び出し - 統一APIクライアントに移行
import { unifiedAPI } from './unified-client.js';

/**
 * 認証APIクライアント（後方互換性維持）
 * 統一APIクライアントを使用するように更新
 */
export class AuthAPI {
  constructor(client = unifiedAPI) {
    this.client = client;
    this.tokenKey = 'auth_token';
    this.refreshTokenKey = 'refresh_token';
    this.userKey = 'auth_user';
  }

  /**
   * ローカルストレージからトークンを取得
   */
  getStoredToken() {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(this.tokenKey);
  }

  /**
   * ローカルストレージからリフレッシュトークンを取得
   */
  getStoredRefreshToken() {
    if (typeof window === 'undefined') return null;
    return localStorage.getItem(this.refreshTokenKey);
  }

  /**
   * ローカルストレージからユーザー情報を取得
   */
  getStoredUser() {
    if (typeof window === 'undefined') return null;
    const userStr = localStorage.getItem(this.userKey);
    try {
      return userStr ? JSON.parse(userStr) : null;
    } catch (error) {
      console.error('Failed to parse stored user data:', error);
      return null;
    }
  }

  /**
   * トークンをローカルストレージに保存
   */
  storeToken(token, refreshToken = null, user = null) {
    if (typeof window === 'undefined') return;
    
    localStorage.setItem(this.tokenKey, token);
    if (refreshToken) {
      localStorage.setItem(this.refreshTokenKey, refreshToken);
    }
    if (user) {
      localStorage.setItem(this.userKey, JSON.stringify(user));
    }
    
    // 統一APIクライアントにトークンを設定
    this.client.setToken(token);
  }

  /**
   * 認証情報をローカルストレージから削除
   */
  clearStoredAuth() {
    if (typeof window === 'undefined') return;
    
    localStorage.removeItem(this.tokenKey);
    localStorage.removeItem(this.refreshTokenKey);
    localStorage.removeItem(this.userKey);
    
    // 統一APIクライアントからトークンを削除
    this.client.setToken(null);
  }

  /**
   * JWTトークンの有効期限をチェック
   */
  isTokenExpired(token) {
    if (!token) return true;
    
    try {
      const payload = JSON.parse(atob(token.split('.')[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      return payload.exp < currentTime;
    } catch (error) {
      console.error('Failed to parse JWT token:', error);
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
      // 期限切れトークンをクリア
      this.clearStoredAuth();
    }
    return false;
  }

  /**
   * ログイン
   */
  async login(username, password) {
    try {
      const response = await this.client.auth.login({
        username,
        password
      });

      if (response.success && response.data) {
        const { token, refresh_token, user } = response.data;
        
        // トークンとユーザー情報を保存
        this.storeToken(token, refresh_token, user);
        
        return {
          success: true,
          data: {
            token,
            refresh_token,
            user
          },
          message: 'ログインに成功しました'
        };
      }

      return response;
    } catch (error) {
      console.error('Login error:', error);
      return {
        success: false,
        error: 'LOGIN_ERROR',
        message: 'ログインに失敗しました',
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
        // サーバーサイドでのログアウト処理
        await this.client.auth.logout();
      }
      
      // ローカルの認証情報をクリア
      this.clearStoredAuth();
      
      // カスタムイベントを発火
      if (typeof window !== 'undefined') {
        window.dispatchEvent(new CustomEvent('auth:logout'));
      }
      
      return {
        success: true,
        message: 'ログアウトしました'
      };
    } catch (error) {
      console.error('Logout error:', error);
      
      // エラーが発生してもローカルの認証情報はクリア
      this.clearStoredAuth();
      
      return {
        success: true,
        message: 'ログアウトしました',
        warning: 'サーバーとの通信でエラーが発生しましたが、ローカルの認証情報はクリアされました'
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
          error: 'NO_REFRESH_TOKEN',
          message: 'リフレッシュトークンが見つかりません'
        };
      }

      const response = await this.client.auth.refresh(refreshToken);

      if (response.success && response.data) {
        const { token, refresh_token: newRefreshToken, user } = response.data;
        
        // 新しいトークンを保存
        this.storeToken(token, newRefreshToken || refreshToken, user);
        
        return {
          success: true,
          data: {
            token,
            refresh_token: newRefreshToken || refreshToken,
            user
          },
          message: 'トークンを更新しました'
        };
      }

      // リフレッシュに失敗した場合は認証情報をクリア
      this.clearStoredAuth();
      return response;
    } catch (error) {
      console.error('Token refresh error:', error);
      
      // エラーが発生した場合も認証情報をクリア
      this.clearStoredAuth();
      
      return {
        success: false,
        error: 'REFRESH_ERROR',
        message: 'トークンの更新に失敗しました',
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
          error: 'NO_TOKEN',
          message: 'トークンが見つかりません'
        };
      }

      // ローカルでの期限チェック
      if (this.isTokenExpired(token)) {
        // 期限切れの場合、リフレッシュを試行
        const refreshResult = await this.refreshToken();
        if (refreshResult.success) {
          return {
            success: true,
            data: refreshResult.data,
            message: 'トークンを更新して検証しました'
          };
        } else {
          return {
            success: false,
            error: 'TOKEN_EXPIRED',
            message: 'トークンの期限が切れています'
          };
        }
      }

      // サーバーサイドでの検証
      const response = await this.client.auth.validate();

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: 'トークンは有効です'
        };
      }

      // 検証に失敗した場合は認証情報をクリア
      this.clearStoredAuth();
      return response;
    } catch (error) {
      console.error('Token validation error:', error);
      
      return {
        success: false,
        error: 'VALIDATION_ERROR',
        message: 'トークンの検証に失敗しました',
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
    if (typeof window === 'undefined') return;
    
    const interval = intervalMinutes * 60 * 1000; // ミリ秒に変換
    
    setInterval(async () => {
      const token = this.getStoredToken();
      if (token && !this.isTokenExpired(token)) {
        // トークンの残り時間をチェック（10分以内なら更新）
        try {
          const payload = JSON.parse(atob(token.split('.')[1]));
          const currentTime = Math.floor(Date.now() / 1000);
          const timeUntilExpiry = payload.exp - currentTime;
          
          if (timeUntilExpiry < 600) { // 10分 = 600秒
            await this.refreshToken();
          }
        } catch (error) {
          console.error('Auto refresh error:', error);
        }
      }
    }, interval);
  }
}

// デフォルトのAuthAPIインスタンス
export const authAPI = new AuthAPI();

// 初期化
if (typeof window !== 'undefined') {
  authAPI.initializeAuth();
  authAPI.setupAutoRefresh();
}

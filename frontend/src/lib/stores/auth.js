// 認証状態管理ストア
import { writable, get } from 'svelte/store';
import { authAPI } from '../api/auth.js';
import { 
  saveAuthToken, 
  getAuthToken, 
  saveUserData, 
  getUserData, 
  clearAuthData 
} from '../utils/storage.js';

// 認証状態の初期値
const initialAuthState = {
  isAuthenticated: false,
  token: null,
  user: null,
  loading: false
};

// 認証ストアの作成
function createAuthStore() {
  const { subscribe, set, update } = writable(initialAuthState);

  return {
    subscribe,
    
    // ローディング状態の設定
    setLoading: (loading) => {
      update(state => ({ ...state, loading }));
    },

    // 認証状態の設定
    setAuthState: (authData) => {
      update(state => ({
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
      update(state => ({ ...state, loading: true }));

      try {
        const result = await authAPI.login(credentials.username, credentials.password);
        
        if (result.success) {
          const authData = {
            token: result.data.token,
            user: result.data.user
          };

          // ローカルストレージに保存
          saveAuthToken(result.data.token);
          saveUserData(result.data.user);

          // ストア状態を更新
          update(state => ({
            ...state,
            isAuthenticated: true,
            token: result.data.token,
            user: result.data.user,
            loading: false
          }));

          return result;
        } else {
          update(state => ({ ...state, loading: false }));
          return result;
        }
      } catch (error) {
        console.error('Login error in store:', error);
        update(state => ({ ...state, loading: false }));
        return {
          success: false,
          error: 'LOGIN_STORE_ERROR',
          message: 'ログイン処理でエラーが発生しました',
          details: error.message
        };
      }
    },

    // ログアウト処理
    logout: async () => {
      update(state => ({ ...state, loading: true }));

      try {
        const result = await authAPI.logout();
        
        // ローカルストレージをクリア
        clearAuthData();
        
        // ストア状態をクリア
        set({ ...initialAuthState });

        return result;
      } catch (error) {
        console.error('Logout error in store:', error);
        
        // エラーが発生してもローカル状態はクリア
        clearAuthData();
        set({ ...initialAuthState });

        return {
          success: true,
          message: 'ログアウトしました',
          warning: 'サーバーとの通信でエラーが発生しましたが、ローカルの認証情報はクリアされました'
        };
      }
    },

    // 認証状態チェック
    checkAuthStatus: async () => {
      update(state => ({ ...state, loading: true }));

      try {
        // まずローカルストレージから認証情報を取得
        const storedToken = getAuthToken();
        const storedUser = getUserData();

        if (!storedToken) {
          update(state => ({ ...state, loading: false }));
          return {
            success: false,
            error: 'NO_TOKEN',
            message: 'トークンが見つかりません'
          };
        }

        // トークンの有効性をサーバーで検証
        const result = await authAPI.validateToken();

        if (result.success) {
          // 認証状態を更新
          update(state => ({
            ...state,
            isAuthenticated: true,
            token: storedToken,
            user: result.data?.user || storedUser,
            loading: false
          }));

          return result;
        } else {
          // 検証に失敗した場合は認証情報をクリア
          clearAuthData();
          set({ ...initialAuthState });
          return result;
        }
      } catch (error) {
        console.error('Auth status check error in store:', error);
        
        // エラーが発生した場合も認証情報をクリア
        clearAuthData();
        set({ ...initialAuthState });

        return {
          success: false,
          error: 'AUTH_CHECK_ERROR',
          message: '認証状態の確認でエラーが発生しました',
          details: error.message
        };
      }
    },

    // トークンリフレッシュ
    refreshToken: async () => {
      const currentState = get({ subscribe });
      
      // 既にローディング中の場合は処理をスキップ
      if (currentState.loading) {
        return {
          success: false,
          error: 'ALREADY_LOADING',
          message: '既に処理中です'
        };
      }

      update(state => ({ ...state, loading: true }));

      try {
        const result = await authAPI.refreshToken();

        if (result.success) {
          // 新しいトークンとユーザー情報でストアを更新
          const authData = {
            token: result.data.token,
            user: result.data.user
          };

          // ローカルストレージを更新
          saveAuthToken(result.data.token);
          saveUserData(result.data.user);

          // ストア状態を更新
          update(state => ({
            ...state,
            isAuthenticated: true,
            token: result.data.token,
            user: result.data.user,
            loading: false
          }));

          return result;
        } else {
          // リフレッシュに失敗した場合は認証情報をクリア
          clearAuthData();
          set({ ...initialAuthState });
          return result;
        }
      } catch (error) {
        console.error('Token refresh error in store:', error);
        
        // エラーが発生した場合も認証情報をクリア
        clearAuthData();
        set({ ...initialAuthState });

        return {
          success: false,
          error: 'REFRESH_STORE_ERROR',
          message: 'トークンの更新でエラーが発生しました',
          details: error.message
        };
      }
    },

    // 初期化処理（アプリ起動時に呼び出し）
    initialize: async function() {
      const storedToken = getAuthToken();
      const storedUser = getUserData();

      if (storedToken && storedUser) {
        // ローカルストレージに認証情報がある場合、検証を実行
        await this.checkAuthStatus();
      }
    }
  };
}

// 認証ストアのインスタンス作成
export const authStore = createAuthStore();

// ブラウザ環境でのみ初期化を実行
if (typeof window !== 'undefined') {
  // ページロード時に認証状態を復元
  authStore.initialize();

  // 認証関連のイベントリスナーを設定
  window.addEventListener('auth:logout', () => {
    authStore.clearAuth();
  });

  // ページの可視性が変わった時に認証状態をチェック
  document.addEventListener('visibilitychange', () => {
    if (!document.hidden) {
      const storedToken = getAuthToken();
      if (storedToken) {
        authStore.checkAuthStatus();
      }
    }
  });
}

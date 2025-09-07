// 認証ガードユーティリティ
import { redirect } from '@sveltejs/kit';
import { get } from 'svelte/store';
import { authStore } from '../stores/auth.js';
import { getAuthToken } from './storage.js';

/**
 * JWTトークンの有効期限をチェック
 */
export function isTokenExpired(token) {
  if (!token) return true;
  
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    const currentTime = Math.floor(Date.now() / 1000);
    
    // 期限切れまで5分以内の場合も期限切れとして扱う（リフレッシュのため）
    const bufferTime = 5 * 60; // 5分
    return payload.exp < (currentTime + bufferTime);
  } catch (error) {
    console.error('Failed to parse JWT token:', error);
    return true;
  }
}

/**
 * トークンの残り時間を秒で取得
 */
export function getTokenTimeRemaining(token) {
  if (!token) return 0;
  
  try {
    const payload = JSON.parse(atob(token.split('.')[1]));
    const currentTime = Math.floor(Date.now() / 1000);
    return Math.max(0, payload.exp - currentTime);
  } catch (error) {
    console.error('Failed to parse JWT token:', error);
    return 0;
  }
}

/**
 * 認証が必要なページの保護
 * SvelteKitのload関数で使用
 */
export async function requireAuth(url, options = {}) {
  const {
    redirectTo = '/login',
    allowExpired = false,
    checkServer = true
  } = options;

  // ブラウザ環境でのみ実行
  if (typeof window === 'undefined') {
    return { authenticated: false };
  }

  const token = getAuthToken();
  
  // トークンが存在しない場合
  if (!token) {
    const redirectUrl = `${redirectTo}?redirect=${encodeURIComponent(url.pathname + url.search)}`;
    throw redirect(302, redirectUrl);
  }

  // トークンの期限チェック
  if (!allowExpired && isTokenExpired(token)) {
    // 期限切れの場合、リフレッシュを試行
    try {
      const refreshResult = await authStore.refreshToken();
      if (!refreshResult.success) {
        const redirectUrl = `${redirectTo}?expired=true&redirect=${encodeURIComponent(url.pathname + url.search)}`;
        throw redirect(302, redirectUrl);
      }
    } catch (error) {
      console.error('Token refresh failed during auth guard:', error);
      const redirectUrl = `${redirectTo}?expired=true&redirect=${encodeURIComponent(url.pathname + url.search)}`;
      throw redirect(302, redirectUrl);
    }
  }

  // サーバーサイドでの検証（オプション）
  if (checkServer) {
    try {
      const validationResult = await authStore.checkAuthStatus();
      if (!validationResult.success) {
        const redirectUrl = `${redirectTo}?invalid=true&redirect=${encodeURIComponent(url.pathname + url.search)}`;
        throw redirect(302, redirectUrl);
      }
    } catch (error) {
      console.error('Server-side auth validation failed:', error);
      const redirectUrl = `${redirectTo}?error=true&redirect=${encodeURIComponent(url.pathname + url.search)}`;
      throw redirect(302, redirectUrl);
    }
  }

  return { authenticated: true };
}

/**
 * 管理者権限が必要なページの保護
 */
export async function requireAdmin(url, options = {}) {
  // まず基本的な認証チェック
  const authResult = await requireAuth(url, options);
  
  if (!authResult.authenticated) {
    return authResult;
  }

  // 管理者権限のチェック
  const authState = get(authStore);
  const user = authState.user;

  if (!user || user.role !== 'admin') {
    throw redirect(302, '/login?unauthorized=true');
  }

  return { authenticated: true, isAdmin: true };
}

/**
 * 既にログイン済みの場合のリダイレクト処理
 * ログインページなどで使用
 */
export function redirectIfAuthenticated(url, options = {}) {
  const {
    redirectTo = '/admin',
    checkExpired = true
  } = options;

  // ブラウザ環境でのみ実行
  if (typeof window === 'undefined') {
    return { shouldRedirect: false };
  }

  const token = getAuthToken();
  
  if (token && (!checkExpired || !isTokenExpired(token))) {
    // リダイレクト先のパラメータがある場合はそちらを優先
    const redirectTarget = url.searchParams.get('redirect') || redirectTo;
    throw redirect(302, redirectTarget);
  }

  return { shouldRedirect: false };
}

/**
 * 認証状態の監視とトークン期限切れ時の自動ログアウト
 */
export function setupAuthMonitoring() {
  if (typeof window === 'undefined') return;

  let tokenCheckInterval;
  let visibilityChangeHandler;

  // 定期的なトークンチェック（1分ごと）
  const startTokenCheck = () => {
    tokenCheckInterval = setInterval(async () => {
      const token = getAuthToken();
      if (token && isTokenExpired(token)) {
        console.log('Token expired, attempting refresh...');
        
        try {
          const refreshResult = await authStore.refreshToken();
          if (!refreshResult.success) {
            console.log('Token refresh failed, logging out...');
            await authStore.logout();
            
            // 現在のページが認証が必要なページの場合はログインページにリダイレクト
            if (window.location.pathname.startsWith('/admin')) {
              window.location.href = '/login?expired=true';
            }
          }
        } catch (error) {
          console.error('Auto token refresh error:', error);
          await authStore.logout();
          
          if (window.location.pathname.startsWith('/admin')) {
            window.location.href = '/login?error=true';
          }
        }
      }
    }, 60000); // 1分
  };

  // ページの可視性が変わった時の処理
  visibilityChangeHandler = async () => {
    if (!document.hidden) {
      const token = getAuthToken();
      if (token) {
        // ページが再び表示された時に認証状態をチェック
        try {
          const authResult = await authStore.checkAuthStatus();
          if (!authResult.success && window.location.pathname.startsWith('/admin')) {
            window.location.href = '/login?session_expired=true';
          }
        } catch (error) {
          console.error('Auth status check on visibility change failed:', error);
        }
      }
    }
  };

  // イベントリスナーの設定
  document.addEventListener('visibilitychange', visibilityChangeHandler);
  
  // 認証状態の監視開始
  startTokenCheck();

  // クリーンアップ関数を返す
  return () => {
    if (tokenCheckInterval) {
      clearInterval(tokenCheckInterval);
    }
    if (visibilityChangeHandler) {
      document.removeEventListener('visibilitychange', visibilityChangeHandler);
    }
  };
}

/**
 * ログアウト時のクリーンアップ処理
 */
export async function performLogout(redirectTo = '/login') {
  try {
    // ストアのログアウト処理を実行
    await authStore.logout();
    
    // ページをリダイレクト
    if (typeof window !== 'undefined') {
      window.location.href = redirectTo;
    }
  } catch (error) {
    console.error('Logout error:', error);
    
    // エラーが発生してもリダイレクトは実行
    if (typeof window !== 'undefined') {
      window.location.href = redirectTo + '?logout_error=true';
    }
  }
}

/**
 * 認証エラーハンドリング
 */
export function handleAuthError(error, currentUrl) {
  console.error('Authentication error:', error);
  
  // エラーの種類に応じて適切な処理を実行
  switch (error.error) {
    case 'TOKEN_EXPIRED':
    case 'NO_TOKEN':
      return redirect(302, `/login?expired=true&redirect=${encodeURIComponent(currentUrl)}`);
    
    case 'INVALID_TOKEN':
      return redirect(302, `/login?invalid=true&redirect=${encodeURIComponent(currentUrl)}`);
    
    case 'UNAUTHORIZED':
      return redirect(302, `/login?unauthorized=true`);
    
    case 'NETWORK_ERROR':
      return redirect(302, `/login?network_error=true&redirect=${encodeURIComponent(currentUrl)}`);
    
    default:
      return redirect(302, `/login?error=true&redirect=${encodeURIComponent(currentUrl)}`);
  }
}
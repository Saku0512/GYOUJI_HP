// @ts-nocheck
// ログインページのロード関数
import { redirectIfAuthenticated } from '../../lib/utils/auth-guard.js';

/** @param {Parameters<import('./$types').PageLoad>[0]} event */
export async function load({ url }) {
  // 既にログイン済みの場合はリダイレクト
  redirectIfAuthenticated(url, {
    redirectTo: '/admin',
    checkExpired: true
  });

  // URLパラメータからエラー情報を取得
  const errorType = url.searchParams.get('expired') ? 'expired' :
                   url.searchParams.get('invalid') ? 'invalid' :
                   url.searchParams.get('unauthorized') ? 'unauthorized' :
                   url.searchParams.get('network_error') ? 'network_error' :
                   url.searchParams.get('logout_error') ? 'logout_error' :
                   url.searchParams.get('session_expired') ? 'session_expired' :
                   url.searchParams.get('error') ? 'error' : null;

  const redirectTarget = url.searchParams.get('redirect');

  return {
    errorType,
    redirectTarget
  };
}
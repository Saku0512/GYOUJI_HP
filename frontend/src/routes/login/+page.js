// ログインページのロード関数
import { redirect } from '@sveltejs/kit';
import { getAuthToken } from '$lib/utils/storage.js';

/** @type {import('./$types').PageLoad} */
export async function load({ url }) {
  // ブラウザ環境でのみ実行
  if (typeof window !== 'undefined') {
    const token = getAuthToken();
    
    // 既にログイン済みの場合は管理者ダッシュボードにリダイレクト
    if (token) {
      // リダイレクト先のパラメータがある場合はそちらを優先
      const redirectTo = url.searchParams.get('redirect') || '/admin';
      throw redirect(302, redirectTo);
    }
  }

  return {};
}
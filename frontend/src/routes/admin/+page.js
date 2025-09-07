// 管理者ページの認証ガード
import { redirect } from '@sveltejs/kit';
import { getAuthToken } from '../../lib/utils/storage.js';

/**
 * 管理者ページのロード処理
 * 認証チェックとアクセス制御を行う
 */
export async function load({ url, fetch }) {
  // ブラウザ環境でのみ認証チェックを実行
  if (typeof window !== 'undefined') {
    const token = getAuthToken();
    
    // トークンが存在しない場合はログインページにリダイレクト
    if (!token) {
      throw redirect(302, '/login?redirect=' + encodeURIComponent(url.pathname));
    }
    
    // トークンの有効性をチェック（簡易版）
    try {
      // JWTトークンの期限切れチェック
      const payload = JSON.parse(atob(token.split('.')[1]));
      const currentTime = Math.floor(Date.now() / 1000);
      
      if (payload.exp && payload.exp < currentTime) {
        // トークンが期限切れの場合
        throw redirect(302, '/login?expired=true&redirect=' + encodeURIComponent(url.pathname));
      }
    } catch (error) {
      // トークンの解析に失敗した場合
      console.error('Token validation error:', error);
      throw redirect(302, '/login?invalid=true&redirect=' + encodeURIComponent(url.pathname));
    }
  }
  
  return {
    // ページに必要な初期データがあればここで取得
    title: '管理者ダッシュボード'
  };
}
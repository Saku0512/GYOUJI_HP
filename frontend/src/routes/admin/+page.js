// 管理者ページの認証ガード
import { requireAdmin } from '../../lib/utils/auth-guard.js';

/**
 * 管理者ページのロード処理
 * 認証チェックとアクセス制御を行う
 */
export async function load({ url, fetch }) {
  // 管理者権限の認証チェック
  await requireAdmin(url, {
    redirectTo: '/login',
    checkServer: true
  });
  
  return {
    // ページに必要な初期データがあればここで取得
    title: '管理者ダッシュボード'
  };
}
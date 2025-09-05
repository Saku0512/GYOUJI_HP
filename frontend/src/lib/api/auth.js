// 認証関連API呼び出し
import { apiClient } from './client.js';

export const authAPI = {
  // ログイン
  login: async (username, password) => {
    // 実装は後のタスクで行う
    console.log('Login API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  },

  // ログアウト
  logout: async () => {
    // 実装は後のタスクで行う
    console.log('Logout API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  },

  // トークンリフレッシュ
  refreshToken: async () => {
    // 実装は後のタスクで行う
    console.log('Refresh token API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  },

  // トークン検証
  validateToken: async () => {
    // 実装は後のタスクで行う
    console.log('Validate token API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  }
};

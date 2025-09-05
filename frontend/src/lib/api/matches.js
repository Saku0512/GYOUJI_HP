// 試合関連API呼び出し
import { apiClient } from './client.js';

export const matchAPI = {
  // 試合一覧取得
  getMatches: async (sport) => {
    // 実装は後のタスクで行う
    console.log('Get matches API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  },

  // 特定試合取得
  getMatch: async (id) => {
    // 実装は後のタスクで行う
    console.log('Get match API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  },

  // 試合結果更新
  updateMatch: async (id, result) => {
    // 実装は後のタスクで行う
    console.log('Update match API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  },

  // 試合作成
  createMatch: async (matchData) => {
    // 実装は後のタスクで行う
    console.log('Create match API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  }
};

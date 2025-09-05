// トーナメント関連API呼び出し
import { apiClient } from './client.js';

export const tournamentAPI = {
  // トーナメント一覧取得
  getTournaments: async () => {
    // 実装は後のタスクで行う
    console.log('Get tournaments API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  },

  // 特定スポーツのトーナメント取得
  getTournament: async (sport) => {
    // 実装は後のタスクで行う
    console.log('Get tournament API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  },

  // トーナメントブラケット取得
  getTournamentBracket: async (sport) => {
    // 実装は後のタスクで行う
    console.log('Get tournament bracket API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  },

  // トーナメント形式更新
  updateTournamentFormat: async (sport, format) => {
    // 実装は後のタスクで行う
    console.log('Update tournament format API will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  }
};

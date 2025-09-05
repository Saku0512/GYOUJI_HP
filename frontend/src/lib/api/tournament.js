// トーナメント関連API呼び出し
import { apiClient } from './client.js';

/**
 * トーナメントAPIクライアント
 * トーナメント情報の取得と管理を行う
 */
export class TournamentAPI {
  constructor(client = apiClient) {
    this.client = client;
    this.supportedSports = ['volleyball', 'table_tennis', 'soccer'];
  }

  /**
   * スポーツ名の検証
   */
  validateSport(sport) {
    if (!sport) {
      throw new Error('スポーツ名が指定されていません');
    }
    
    if (!this.supportedSports.includes(sport)) {
      throw new Error(`サポートされていないスポーツです: ${sport}`);
    }
    
    return true;
  }

  /**
   * トーナメント一覧取得
   * 全スポーツのトーナメント情報を取得
   */
  async getTournaments() {
    try {
      const response = await this.client.get('/tournaments');

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: 'トーナメント一覧を取得しました'
        };
      }

      return response;
    } catch (error) {
      console.error('Get tournaments error:', error);
      return {
        success: false,
        error: 'GET_TOURNAMENTS_ERROR',
        message: 'トーナメント一覧の取得に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 特定スポーツのトーナメント取得
   * 指定されたスポーツのトーナメント情報を取得
   */
  async getTournament(sport) {
    try {
      this.validateSport(sport);

      const response = await this.client.get(`/tournaments/${sport}`);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメント情報を取得しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Get tournament error:', error);
      return {
        success: false,
        error: 'GET_TOURNAMENT_ERROR',
        message: 'トーナメント情報の取得に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * トーナメントブラケット取得
   * 指定されたスポーツのブラケット情報（試合組み合わせ）を取得
   */
  async getTournamentBracket(sport) {
    try {
      this.validateSport(sport);

      const response = await this.client.get(`/tournaments/${sport}/bracket`);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のブラケット情報を取得しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Get tournament bracket error:', error);
      return {
        success: false,
        error: 'GET_BRACKET_ERROR',
        message: 'ブラケット情報の取得に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * トーナメント形式更新
   * 卓球の晴天時/雨天時形式切り替えなど
   */
  async updateTournamentFormat(sport, format) {
    try {
      this.validateSport(sport);

      if (!format) {
        throw new Error('形式が指定されていません');
      }

      const response = await this.client.put(`/tournaments/${sport}/format`, {
        format: format
      });

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメント形式を${format}に更新しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Update tournament format error:', error);
      return {
        success: false,
        error: 'UPDATE_FORMAT_ERROR',
        message: 'トーナメント形式の更新に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * トーナメント作成
   * 新しいトーナメントを作成（管理者用）
   */
  async createTournament(tournamentData) {
    try {
      const { sport, format, teams } = tournamentData;

      this.validateSport(sport);

      if (!format) {
        throw new Error('トーナメント形式が指定されていません');
      }

      if (!teams || !Array.isArray(teams) || teams.length === 0) {
        throw new Error('参加チーム情報が正しくありません');
      }

      const response = await this.client.post('/tournaments', {
        sport,
        format,
        teams
      });

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメントを作成しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Create tournament error:', error);
      return {
        success: false,
        error: 'CREATE_TOURNAMENT_ERROR',
        message: 'トーナメントの作成に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * トーナメント削除
   * 指定されたトーナメントを削除（管理者用）
   */
  async deleteTournament(sport) {
    try {
      this.validateSport(sport);

      const response = await this.client.delete(`/tournaments/${sport}`);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメントを削除しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Delete tournament error:', error);
      return {
        success: false,
        error: 'DELETE_TOURNAMENT_ERROR',
        message: 'トーナメントの削除に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * トーナメント状態更新
   * トーナメントの状態（開始、終了など）を更新
   */
  async updateTournamentStatus(sport, status) {
    try {
      this.validateSport(sport);

      const validStatuses = ['pending', 'active', 'completed', 'cancelled'];
      if (!validStatuses.includes(status)) {
        throw new Error(`無効なステータスです: ${status}`);
      }

      const response = await this.client.patch(`/tournaments/${sport}/status`, {
        status
      });

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメント状態を${status}に更新しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Update tournament status error:', error);
      return {
        success: false,
        error: 'UPDATE_STATUS_ERROR',
        message: 'トーナメント状態の更新に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * トーナメント統計情報取得
   * 試合数、完了率などの統計情報を取得
   */
  async getTournamentStats(sport) {
    try {
      this.validateSport(sport);

      const response = await this.client.get(`/tournaments/${sport}/stats`);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の統計情報を取得しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Get tournament stats error:', error);
      return {
        success: false,
        error: 'GET_STATS_ERROR',
        message: '統計情報の取得に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 利用可能な形式一覧取得
   * 指定されたスポーツで利用可能なトーナメント形式を取得
   */
  async getAvailableFormats(sport) {
    try {
      this.validateSport(sport);

      const response = await this.client.get(`/tournaments/${sport}/formats`);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の利用可能な形式一覧を取得しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Get available formats error:', error);
      return {
        success: false,
        error: 'GET_FORMATS_ERROR',
        message: '形式一覧の取得に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * サポートされているスポーツ一覧を取得
   */
  getSupportedSports() {
    return {
      success: true,
      data: this.supportedSports,
      message: 'サポートされているスポーツ一覧'
    };
  }
}

// デフォルトのTournamentAPIインスタンス
export const tournamentAPI = new TournamentAPI();

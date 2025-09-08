// 試合関連API呼び出し - 統一APIクライアントに移行
import { unifiedAPI } from './unified-client.js';

/**
 * 試合APIクライアント（後方互換性維持）
 * 統一APIクライアントを使用するように更新
 */
export class MatchAPI {
  constructor(client = unifiedAPI) {
    this.client = client;
    this.supportedSports = ['volleyball', 'table_tennis', 'soccer'];
    this.validStatuses = ['pending', 'in_progress', 'completed', 'cancelled'];
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
   * 試合IDの検証
   */
  validateMatchId(matchId) {
    if (!matchId) {
      throw new Error('試合IDが指定されていません');
    }
    
    if (typeof matchId !== 'number' && typeof matchId !== 'string') {
      throw new Error('試合IDの形式が正しくありません');
    }
    
    return true;
  }

  /**
   * 試合結果データの検証
   */
  validateMatchResult(result) {
    if (!result || typeof result !== 'object') {
      throw new Error('試合結果データが正しくありません');
    }

    const { score1, score2, winner } = result;

    // スコアの検証
    if (score1 !== undefined && (typeof score1 !== 'number' || score1 < 0)) {
      throw new Error('チーム1のスコアが正しくありません');
    }

    if (score2 !== undefined && (typeof score2 !== 'number' || score2 < 0)) {
      throw new Error('チーム2のスコアが正しくありません');
    }

    // 勝者の検証（スコアが入力されている場合）
    if (score1 !== undefined && score2 !== undefined && winner !== undefined) {
      if (typeof winner !== 'string' || winner.trim() === '') {
        throw new Error('勝者の情報が正しくありません');
      }
    }

    return true;
  }

  /**
   * 試合一覧取得
   * 指定されたスポーツの全試合を取得
   */
  async getMatches(sport, options = {}) {
    try {
      this.validateSport(sport);

      // クエリパラメータの構築
      const queryParams = new URLSearchParams();
      
      if (options.status) {
        queryParams.append('status', options.status);
      }
      
      if (options.round) {
        queryParams.append('round', options.round);
      }
      
      if (options.limit) {
        queryParams.append('limit', options.limit.toString());
      }
      
      if (options.offset) {
        queryParams.append('offset', options.offset.toString());
      }

      const queryString = queryParams.toString();
      const response = await this.client.matches.getBySport(sport, options);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の試合一覧を取得しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Get matches error:', error);
      return {
        success: false,
        error: 'GET_MATCHES_ERROR',
        message: '試合一覧の取得に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 特定試合の詳細取得
   * 指定されたIDの試合詳細情報を取得
   */
  async getMatch(matchId) {
    try {
      this.validateMatchId(matchId);

      const response = await this.client.matches.getById(matchId);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: '試合詳細を取得しました'
        };
      }

      return response;
    } catch (error) {
      console.error('Get match error:', error);
      return {
        success: false,
        error: 'GET_MATCH_ERROR',
        message: '試合詳細の取得に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 試合結果更新
   * 指定された試合の結果を更新
   */
  async updateMatch(matchId, result) {
    try {
      this.validateMatchId(matchId);
      this.validateMatchResult(result);

      const response = await this.client.matches.updateResult(matchId, result);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: '試合結果を更新しました'
        };
      }

      return response;
    } catch (error) {
      console.error('Update match error:', error);
      return {
        success: false,
        error: 'UPDATE_MATCH_ERROR',
        message: '試合結果の更新に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 新規試合作成
   * 新しい試合を作成（管理者用）
   */
  async createMatch(matchData) {
    try {
      if (!matchData || typeof matchData !== 'object') {
        throw new Error('試合データが正しくありません');
      }

      const { sport, tournament_id, round, team1, team2, scheduled_at } = matchData;

      // 必須フィールドの検証
      if (!sport) {
        throw new Error('スポーツが指定されていません');
      }
      this.validateSport(sport);

      if (!tournament_id) {
        throw new Error('トーナメントIDが指定されていません');
      }

      if (!round) {
        throw new Error('ラウンドが指定されていません');
      }

      if (!team1 || !team2) {
        throw new Error('対戦チームが正しく指定されていません');
      }

      if (team1 === team2) {
        throw new Error('同じチーム同士の試合は作成できません');
      }

      const response = await this.client.matches.create(matchData);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: '新しい試合を作成しました'
        };
      }

      return response;
    } catch (error) {
      console.error('Create match error:', error);
      return {
        success: false,
        error: 'CREATE_MATCH_ERROR',
        message: '試合の作成に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 試合削除
   * 指定された試合を削除（管理者用）
   */
  async deleteMatch(matchId) {
    try {
      this.validateMatchId(matchId);

      const response = await this.client.matches.delete(matchId);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: '試合を削除しました'
        };
      }

      return response;
    } catch (error) {
      console.error('Delete match error:', error);
      return {
        success: false,
        error: 'DELETE_MATCH_ERROR',
        message: '試合の削除に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 試合状態更新
   * 試合の状態（開始、終了など）を更新
   */
  async updateMatchStatus(matchId, status) {
    try {
      this.validateMatchId(matchId);

      if (!this.validStatuses.includes(status)) {
        throw new Error(`無効なステータスです: ${status}`);
      }

      // 統一APIクライアントで直接リクエストを使用
      const response = await this.client.request('PATCH', `/matches/${matchId}/status`, {
        status
      });

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `試合状態を${status}に更新しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Update match status error:', error);
      return {
        success: false,
        error: 'UPDATE_MATCH_STATUS_ERROR',
        message: '試合状態の更新に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 未完了試合一覧取得
   * 管理者ダッシュボード用の未完了試合一覧を取得
   */
  async getPendingMatches(sport) {
    try {
      this.validateSport(sport);

      const response = await this.client.request('GET', `/matches/${sport}/pending`);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の未完了試合一覧を取得しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Get pending matches error:', error);
      return {
        success: false,
        error: 'GET_PENDING_MATCHES_ERROR',
        message: '未完了試合一覧の取得に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 試合結果の一括更新
   * 複数の試合結果を一度に更新
   */
  async updateMultipleMatches(updates) {
    try {
      if (!Array.isArray(updates) || updates.length === 0) {
        throw new Error('更新データが正しくありません');
      }

      // 各更新データの検証
      for (const update of updates) {
        if (!update.matchId) {
          throw new Error('試合IDが指定されていない更新データがあります');
        }
        this.validateMatchId(update.matchId);
        this.validateMatchResult(update.result);
      }

      const response = await this.client.request('PUT', '/matches/batch', {
        updates
      });

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${updates.length}件の試合結果を更新しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Update multiple matches error:', error);
      return {
        success: false,
        error: 'UPDATE_MULTIPLE_MATCHES_ERROR',
        message: '試合結果の一括更新に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 試合統計情報取得
   * 指定された試合の統計情報を取得
   */
  async getMatchStats(matchId) {
    try {
      this.validateMatchId(matchId);

      const response = await this.client.request('GET', `/matches/${matchId}/stats`);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: '試合統計情報を取得しました'
        };
      }

      return response;
    } catch (error) {
      console.error('Get match stats error:', error);
      return {
        success: false,
        error: 'GET_MATCH_STATS_ERROR',
        message: '試合統計情報の取得に失敗しました',
        details: error.message
      };
    }
  }

  /**
   * 次の試合取得
   * 指定されたスポーツの次に予定されている試合を取得
   */
  async getNextMatch(sport) {
    try {
      this.validateSport(sport);

      const response = await this.client.request('GET', `/matches/${sport}/next`);

      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の次の試合情報を取得しました`
        };
      }

      return response;
    } catch (error) {
      console.error('Get next match error:', error);
      return {
        success: false,
        error: 'GET_NEXT_MATCH_ERROR',
        message: '次の試合情報の取得に失敗しました',
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

  /**
   * 有効なステータス一覧を取得
   */
  getValidStatuses() {
    return {
      success: true,
      data: this.validStatuses,
      message: '有効なステータス一覧'
    };
  }
}

// デフォルトのMatchAPIインスタンス
export const matchAPI = new MatchAPI();
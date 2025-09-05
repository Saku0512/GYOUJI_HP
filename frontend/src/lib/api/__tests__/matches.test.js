// 試合APIクライアントの単体テスト
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { MatchAPI } from '../matches.js';

// モックAPIクライアント
const mockApiClient = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  patch: vi.fn(),
  delete: vi.fn()
};

describe('MatchAPI', () => {
  let matchAPI;

  beforeEach(() => {
    // モックをリセット
    vi.clearAllMocks();
    
    // MatchAPIインスタンスを作成
    matchAPI = new MatchAPI(mockApiClient);
  });

  describe('validateSport', () => {
    it('有効なスポーツ名の場合はtrueを返す', () => {
      expect(() => matchAPI.validateSport('volleyball')).not.toThrow();
      expect(() => matchAPI.validateSport('table_tennis')).not.toThrow();
      expect(() => matchAPI.validateSport('soccer')).not.toThrow();
    });

    it('無効なスポーツ名の場合はエラーを投げる', () => {
      expect(() => matchAPI.validateSport('invalid_sport')).toThrow('サポートされていないスポーツです: invalid_sport');
      expect(() => matchAPI.validateSport('')).toThrow('スポーツ名が指定されていません');
      expect(() => matchAPI.validateSport(null)).toThrow('スポーツ名が指定されていません');
    });
  });

  describe('validateMatchId', () => {
    it('有効なIDの場合はtrueを返す', () => {
      expect(() => matchAPI.validateMatchId(1)).not.toThrow();
      expect(() => matchAPI.validateMatchId('123')).not.toThrow();
    });

    it('無効なIDの場合はエラーを投げる', () => {
      expect(() => matchAPI.validateMatchId(null)).toThrow('試合IDが指定されていません');
      expect(() => matchAPI.validateMatchId('')).toThrow('試合IDが指定されていません');
      expect(() => matchAPI.validateMatchId({})).toThrow('試合IDの形式が正しくありません');
    });
  });

  describe('validateMatchResult', () => {
    it('有効な試合結果の場合はtrueを返す', () => {
      expect(() => matchAPI.validateMatchResult({ score1: 3, score2: 1, winner: 'team1' })).not.toThrow();
      expect(() => matchAPI.validateMatchResult({ score1: 0, score2: 2, winner: 'team2' })).not.toThrow();
      expect(() => matchAPI.validateMatchResult({ winner: 'team1' })).not.toThrow();
      expect(() => matchAPI.validateMatchResult({})).not.toThrow();
    });

    it('無効な試合結果の場合はエラーを投げる', () => {
      expect(() => matchAPI.validateMatchResult(null)).toThrow('試合結果データが正しくありません');
      expect(() => matchAPI.validateMatchResult({ score1: -1, score2: 1 })).toThrow('チーム1のスコアが正しくありません');
      expect(() => matchAPI.validateMatchResult({ score1: 3, score2: -1 })).toThrow('チーム2のスコアが正しくありません');
      expect(() => matchAPI.validateMatchResult({ score1: 3, score2: 1, winner: '' })).toThrow('勝者の情報が正しくありません');
    });
  });

  describe('getMatches', () => {
    it('成功時に試合一覧を返す', async () => {
      const mockResponse = {
        success: true,
        data: [
          { id: 1, team1: 'Team A', team2: 'Team B', status: 'pending' },
          { id: 2, team1: 'Team C', team2: 'Team D', status: 'completed' }
        ]
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await matchAPI.getMatches('volleyball');

      expect(mockApiClient.get).toHaveBeenCalledWith('/matches/volleyball');
      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse.data);
      expect(result.message).toBe('volleyballの試合一覧を取得しました');
    });

    it('オプション付きで試合一覧を取得する', async () => {
      const mockResponse = {
        success: true,
        data: [{ id: 1, team1: 'Team A', team2: 'Team B', status: 'pending' }]
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const options = { status: 'pending', round: '準決勝', limit: 10, offset: 5 };
      const result = await matchAPI.getMatches('volleyball', options);

      expect(mockApiClient.get).toHaveBeenCalledWith('/matches/volleyball?status=pending&round=%E6%BA%96%E6%B1%BA%E5%8B%9D&limit=10&offset=5');
      expect(result.success).toBe(true);
    });

    it('無効なスポーツ名の場合はエラーを返す', async () => {
      const result = await matchAPI.getMatches('invalid_sport');

      expect(result.success).toBe(false);
      expect(result.error).toBe('GET_MATCHES_ERROR');
      expect(result.message).toBe('試合一覧の取得に失敗しました');
    });
  });

  describe('getMatch', () => {
    it('成功時に試合情報を返す', async () => {
      const mockResponse = {
        success: true,
        data: { id: 1, team1: 'Team A', team2: 'Team B', score1: 3, score2: 1 }
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await matchAPI.getMatch(1);

      expect(mockApiClient.get).toHaveBeenCalledWith('/matches/1');
      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse.data);
      expect(result.message).toBe('試合詳細を取得しました');
    });

    it('無効なIDの場合はエラーを返す', async () => {
      const result = await matchAPI.getMatch(null);

      expect(result.success).toBe(false);
      expect(result.error).toBe('GET_MATCH_ERROR');
      expect(result.message).toBe('試合詳細の取得に失敗しました');
    });
  });

  describe('updateMatch', () => {
    it('成功時に試合結果を更新する', async () => {
      const matchResult = { score1: 3, score2: 1, winner: 'team1' };
      const mockResponse = {
        success: true,
        data: { id: 1, ...matchResult, status: 'completed' }
      };

      mockApiClient.put.mockResolvedValue(mockResponse);

      const result = await matchAPI.updateMatch(1, matchResult);

      expect(mockApiClient.put).toHaveBeenCalledWith('/matches/1', matchResult);
      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse.data);
      expect(result.message).toBe('試合結果を更新しました');
    });

    it('無効な試合結果の場合はエラーを返す', async () => {
      const result = await matchAPI.updateMatch(1, { score1: -1, score2: 1 });

      expect(result.success).toBe(false);
      expect(result.error).toBe('UPDATE_MATCH_ERROR');
      expect(result.message).toBe('試合結果の更新に失敗しました');
    });
  });

  describe('createMatch', () => {
    it('成功時に試合を作成する', async () => {
      const matchData = {
        sport: 'volleyball',
        tournament_id: 1,
        round: '準決勝',
        team1: 'Team A',
        team2: 'Team B',
        scheduled_at: '2024-01-01T10:00:00Z'
      };

      const mockResponse = {
        success: true,
        data: { id: 1, ...matchData, status: 'pending' }
      };

      mockApiClient.post.mockResolvedValue(mockResponse);

      const result = await matchAPI.createMatch(matchData);

      expect(mockApiClient.post).toHaveBeenCalledWith('/matches', matchData);
      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse.data);
      expect(result.message).toBe('新しい試合を作成しました');
    });

    it('必須フィールドが不足している場合はエラーを返す', async () => {
      const matchData = {
        sport: 'volleyball',
        tournament_id: 1,
        round: '準決勝'
        // team1, team2が不足
      };

      const result = await matchAPI.createMatch(matchData);

      expect(result.success).toBe(false);
      expect(result.error).toBe('CREATE_MATCH_ERROR');
      expect(result.message).toBe('試合の作成に失敗しました');
    });

    it('同じチーム同士の試合の場合はエラーを返す', async () => {
      const matchData = {
        sport: 'volleyball',
        tournament_id: 1,
        round: '準決勝',
        team1: 'Team A',
        team2: 'Team A'
      };

      const result = await matchAPI.createMatch(matchData);

      expect(result.success).toBe(false);
      expect(result.error).toBe('CREATE_MATCH_ERROR');
      expect(result.message).toBe('試合の作成に失敗しました');
    });
  });

  describe('deleteMatch', () => {
    it('成功時に試合を削除する', async () => {
      const mockResponse = {
        success: true,
        data: { id: 1, deleted: true }
      };

      mockApiClient.delete.mockResolvedValue(mockResponse);

      const result = await matchAPI.deleteMatch(1);

      expect(mockApiClient.delete).toHaveBeenCalledWith('/matches/1');
      expect(result.success).toBe(true);
      expect(result.message).toBe('試合を削除しました');
    });

    it('無効なIDの場合はエラーを返す', async () => {
      const result = await matchAPI.deleteMatch(null);

      expect(result.success).toBe(false);
      expect(result.error).toBe('DELETE_MATCH_ERROR');
    });
  });

  describe('updateMatchStatus', () => {
    it('成功時に試合状態を更新する', async () => {
      const mockResponse = {
        success: true,
        data: { id: 1, status: 'completed' }
      };

      mockApiClient.patch.mockResolvedValue(mockResponse);

      const result = await matchAPI.updateMatchStatus(1, 'completed');

      expect(mockApiClient.patch).toHaveBeenCalledWith('/matches/1/status', {
        status: 'completed'
      });
      expect(result.success).toBe(true);
      expect(result.message).toBe('試合状態をcompletedに更新しました');
    });

    it('無効なステータスの場合はエラーを返す', async () => {
      const result = await matchAPI.updateMatchStatus(1, 'invalid_status');

      expect(result.success).toBe(false);
      expect(result.error).toBe('UPDATE_MATCH_STATUS_ERROR');
      expect(result.message).toBe('試合状態の更新に失敗しました');
    });
  });

  describe('getPendingMatches', () => {
    it('未完了試合一覧を取得する', async () => {
      const mockResponse = {
        success: true,
        data: [{ id: 1, team1: 'Team A', team2: 'Team B', status: 'pending' }]
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await matchAPI.getPendingMatches('volleyball');

      expect(mockApiClient.get).toHaveBeenCalledWith('/matches/volleyball/pending');
      expect(result.success).toBe(true);
      expect(result.message).toBe('volleyballの未完了試合一覧を取得しました');
    });

    it('無効なスポーツ名の場合はエラーを返す', async () => {
      const result = await matchAPI.getPendingMatches('invalid_sport');

      expect(result.success).toBe(false);
      expect(result.error).toBe('GET_PENDING_MATCHES_ERROR');
    });
  });

  describe('updateMultipleMatches', () => {
    it('成功時に複数の試合結果を更新する', async () => {
      const updates = [
        { matchId: 1, result: { score1: 3, score2: 1, winner: 'team1' } },
        { matchId: 2, result: { score1: 0, score2: 2, winner: 'team2' } }
      ];

      const mockResponse = {
        success: true,
        data: { updated_count: 2 }
      };

      mockApiClient.put.mockResolvedValue(mockResponse);

      const result = await matchAPI.updateMultipleMatches(updates);

      expect(mockApiClient.put).toHaveBeenCalledWith('/matches/batch', { updates });
      expect(result.success).toBe(true);
      expect(result.message).toBe('2件の試合結果を更新しました');
    });

    it('空の更新データの場合はエラーを返す', async () => {
      const result = await matchAPI.updateMultipleMatches([]);

      expect(result.success).toBe(false);
      expect(result.error).toBe('UPDATE_MULTIPLE_MATCHES_ERROR');
      expect(result.message).toBe('試合結果の一括更新に失敗しました');
    });

    it('無効な更新データの場合はエラーを返す', async () => {
      const updates = [
        { matchId: 1, result: { score1: -1, score2: 1 } } // 無効なスコア
      ];

      const result = await matchAPI.updateMultipleMatches(updates);

      expect(result.success).toBe(false);
      expect(result.error).toBe('UPDATE_MULTIPLE_MATCHES_ERROR');
    });
  });

  describe('getMatchStats', () => {
    it('成功時に試合統計情報を取得する', async () => {
      const mockResponse = {
        success: true,
        data: { total_points: 10, duration: 120 }
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await matchAPI.getMatchStats(1);

      expect(mockApiClient.get).toHaveBeenCalledWith('/matches/1/stats');
      expect(result.success).toBe(true);
      expect(result.message).toBe('試合統計情報を取得しました');
    });

    it('無効なIDの場合はエラーを返す', async () => {
      const result = await matchAPI.getMatchStats(null);

      expect(result.success).toBe(false);
      expect(result.error).toBe('GET_MATCH_STATS_ERROR');
    });
  });

  describe('getNextMatch', () => {
    it('成功時に次の試合情報を取得する', async () => {
      const mockResponse = {
        success: true,
        data: { id: 1, team1: 'Team A', team2: 'Team B', scheduled_at: '2024-01-01T10:00:00Z' }
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await matchAPI.getNextMatch('volleyball');

      expect(mockApiClient.get).toHaveBeenCalledWith('/matches/volleyball/next');
      expect(result.success).toBe(true);
      expect(result.message).toBe('volleyballの次の試合情報を取得しました');
    });

    it('無効なスポーツ名の場合はエラーを返す', async () => {
      const result = await matchAPI.getNextMatch('invalid_sport');

      expect(result.success).toBe(false);
      expect(result.error).toBe('GET_NEXT_MATCH_ERROR');
    });
  });

  describe('getSupportedSports', () => {
    it('サポートされているスポーツ一覧を返す', () => {
      const result = matchAPI.getSupportedSports();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(['volleyball', 'table_tennis', 'soccer']);
      expect(result.message).toBe('サポートされているスポーツ一覧');
    });
  });

  describe('getValidStatuses', () => {
    it('有効なステータス一覧を返す', () => {
      const result = matchAPI.getValidStatuses();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(['pending', 'in_progress', 'completed', 'cancelled']);
      expect(result.message).toBe('有効なステータス一覧');
    });
  });

  describe('エラーハンドリング', () => {
    it('APIクライアントエラーを適切に処理する', async () => {
      const error = new Error('Network error');
      mockApiClient.get.mockRejectedValue(error);

      const result = await matchAPI.getMatches('volleyball');

      expect(result.success).toBe(false);
      expect(result.error).toBe('GET_MATCHES_ERROR');
      expect(result.details).toBe('Network error');
    });

    it('APIレスポンスエラーを適切に処理する', async () => {
      const mockResponse = {
        success: false,
        error: 'NOT_FOUND',
        message: '試合が見つかりません'
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await matchAPI.getMatch(999);

      expect(result.success).toBe(false);
      expect(result.error).toBe('NOT_FOUND');
      expect(result.message).toBe('試合が見つかりません');
    });
  });
});
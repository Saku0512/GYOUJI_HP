// トーナメントAPIクライアントの単体テスト
import { describe, it, expect, beforeEach, vi } from 'vitest';
import { TournamentAPI } from '../tournament.js';

// モックAPIクライアント
const mockApiClient = {
  get: vi.fn(),
  post: vi.fn(),
  put: vi.fn(),
  patch: vi.fn(),
  delete: vi.fn()
};

describe('TournamentAPI', () => {
  let tournamentAPI;

  beforeEach(() => {
    // モックをリセット
    vi.clearAllMocks();
    
    // TournamentAPIインスタンスを作成
    tournamentAPI = new TournamentAPI(mockApiClient);
  });

  describe('validateSport', () => {
    it('有効なスポーツ名の場合はtrueを返す', () => {
      expect(() => tournamentAPI.validateSport('volleyball')).not.toThrow();
      expect(() => tournamentAPI.validateSport('table_tennis')).not.toThrow();
      expect(() => tournamentAPI.validateSport('soccer')).not.toThrow();
    });

    it('無効なスポーツ名の場合はエラーを投げる', () => {
      expect(() => tournamentAPI.validateSport('invalid_sport')).toThrow();
      expect(() => tournamentAPI.validateSport('')).toThrow();
      expect(() => tournamentAPI.validateSport(null)).toThrow();
    });
  });

  describe('getTournaments', () => {
    it('成功時にトーナメント一覧を返す', async () => {
      const mockResponse = {
        success: true,
        data: [
          { id: 1, sport: 'volleyball', status: 'active' },
          { id: 2, sport: 'table_tennis', status: 'pending' }
        ]
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await tournamentAPI.getTournaments();

      expect(mockApiClient.get).toHaveBeenCalledWith('/tournaments');
      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse.data);
    });

    it('失敗時にエラーを返す', async () => {
      const mockResponse = {
        success: false,
        error: 'SERVER_ERROR',
        message: 'サーバーエラーが発生しました'
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await tournamentAPI.getTournaments();

      expect(result.success).toBe(false);
      expect(result.error).toBe('SERVER_ERROR');
    });
  });

  describe('getTournament', () => {
    it('成功時に指定スポーツのトーナメント情報を返す', async () => {
      const mockResponse = {
        success: true,
        data: { id: 1, sport: 'volleyball', status: 'active' }
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await tournamentAPI.getTournament('volleyball');

      expect(mockApiClient.get).toHaveBeenCalledWith('/tournaments/volleyball');
      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse.data);
    });

    it('無効なスポーツ名の場合はエラーを返す', async () => {
      const result = await tournamentAPI.getTournament('invalid_sport');

      expect(result.success).toBe(false);
      expect(result.error).toBe('GET_TOURNAMENT_ERROR');
    });
  });

  describe('getTournamentBracket', () => {
    it('成功時にブラケット情報を返す', async () => {
      const mockResponse = {
        success: true,
        data: {
          tournament_id: 1,
          sport: 'volleyball',
          rounds: [
            { name: '準決勝', matches: [] },
            { name: '決勝', matches: [] }
          ]
        }
      };

      mockApiClient.get.mockResolvedValue(mockResponse);

      const result = await tournamentAPI.getTournamentBracket('volleyball');

      expect(mockApiClient.get).toHaveBeenCalledWith('/tournaments/volleyball/bracket');
      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse.data);
    });
  });

  describe('updateTournamentFormat', () => {
    it('成功時にトーナメント形式を更新する', async () => {
      const mockResponse = {
        success: true,
        data: { id: 1, sport: 'table_tennis', format: 'rainy_day' }
      };

      mockApiClient.put.mockResolvedValue(mockResponse);

      const result = await tournamentAPI.updateTournamentFormat('table_tennis', 'rainy_day');

      expect(mockApiClient.put).toHaveBeenCalledWith('/tournaments/table_tennis/format', {
        format: 'rainy_day'
      });
      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse.data);
    });

    it('形式が指定されていない場合はエラーを返す', async () => {
      const result = await tournamentAPI.updateTournamentFormat('table_tennis', '');

      expect(result.success).toBe(false);
      expect(result.error).toBe('UPDATE_FORMAT_ERROR');
    });
  });

  describe('createTournament', () => {
    it('成功時にトーナメントを作成する', async () => {
      const tournamentData = {
        sport: 'volleyball',
        format: 'single_elimination',
        teams: ['Team A', 'Team B', 'Team C', 'Team D']
      };

      const mockResponse = {
        success: true,
        data: { id: 1, ...tournamentData }
      };

      mockApiClient.post.mockResolvedValue(mockResponse);

      const result = await tournamentAPI.createTournament(tournamentData);

      expect(mockApiClient.post).toHaveBeenCalledWith('/tournaments', tournamentData);
      expect(result.success).toBe(true);
      expect(result.data).toEqual(mockResponse.data);
    });

    it('チーム情報が不正な場合はエラーを返す', async () => {
      const tournamentData = {
        sport: 'volleyball',
        format: 'single_elimination',
        teams: []
      };

      const result = await tournamentAPI.createTournament(tournamentData);

      expect(result.success).toBe(false);
      expect(result.error).toBe('CREATE_TOURNAMENT_ERROR');
    });
  });

  describe('updateTournamentStatus', () => {
    it('成功時にトーナメント状態を更新する', async () => {
      const mockResponse = {
        success: true,
        data: { id: 1, sport: 'volleyball', status: 'completed' }
      };

      mockApiClient.patch.mockResolvedValue(mockResponse);

      const result = await tournamentAPI.updateTournamentStatus('volleyball', 'completed');

      expect(mockApiClient.patch).toHaveBeenCalledWith('/tournaments/volleyball/status', {
        status: 'completed'
      });
      expect(result.success).toBe(true);
    });

    it('無効なステータスの場合はエラーを返す', async () => {
      const result = await tournamentAPI.updateTournamentStatus('volleyball', 'invalid_status');

      expect(result.success).toBe(false);
      expect(result.error).toBe('UPDATE_STATUS_ERROR');
    });
  });

  describe('getSupportedSports', () => {
    it('サポートされているスポーツ一覧を返す', () => {
      const result = tournamentAPI.getSupportedSports();

      expect(result.success).toBe(true);
      expect(result.data).toEqual(['volleyball', 'table_tennis', 'soccer']);
    });
  });
});
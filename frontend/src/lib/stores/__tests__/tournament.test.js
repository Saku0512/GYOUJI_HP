// トーナメントストアの単体テスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import { tournamentStore } from '../tournament.js';
import * as tournamentAPI from '../../api/tournament.js';
import * as matchAPI from '../../api/matches.js';

// モック設定
vi.mock('../../api/tournament.js');
vi.mock('../../api/matches.js');

describe('トーナメントストア', () => {
  beforeEach(() => {
    // 各テスト前にモックをリセット
    vi.clearAllMocks();
    
    // タイマーのモック
    vi.useFakeTimers();
    
    // DOM イベントのモック
    Object.defineProperty(document, 'addEventListener', {
      value: vi.fn(),
      writable: true,
    });

    Object.defineProperty(document, 'hidden', {
      value: false,
      writable: true,
    });

    // ストアをクリーンアップ
    tournamentStore.cleanup();
  });

  afterEach(() => {
    // タイマーをリセット
    vi.useRealTimers();
    
    // ストアをクリーンアップ
    tournamentStore.cleanup();
  });

  describe('初期状態', () => {
    it('初期状態が正しく設定されている', () => {
      const state = get(tournamentStore);
      
      expect(state.tournaments).toEqual({});
      expect(state.currentSport).toBe('volleyball');
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.lastUpdated).toBeNull();
      expect(state.cache).toEqual({});
      expect(state.pollingInterval).toBeNull();
    });
  });

  describe('ローディング状態管理', () => {
    it('ローディング状態を正しく設定できる', () => {
      tournamentStore.setLoading(true);
      
      const state = get(tournamentStore);
      expect(state.loading).toBe(true);
    });

    it('ローディング状態を解除できる', () => {
      tournamentStore.setLoading(true);
      tournamentStore.setLoading(false);
      
      const state = get(tournamentStore);
      expect(state.loading).toBe(false);
    });
  });

  describe('エラー状態管理', () => {
    it('エラー状態を正しく設定できる', () => {
      const errorMessage = 'テストエラー';
      tournamentStore.setError(errorMessage);
      
      const state = get(tournamentStore);
      expect(state.error).toBe(errorMessage);
      expect(state.loading).toBe(false);
    });

    it('エラー状態をクリアできる', () => {
      tournamentStore.setError('テストエラー');
      tournamentStore.clearError();
      
      const state = get(tournamentStore);
      expect(state.error).toBeNull();
    });
  });

  describe('トーナメントデータ取得', () => {
    it('特定スポーツのトーナメントデータを取得できる', async () => {
      const mockTournamentData = {
        id: 1,
        sport: 'volleyball',
        matches: []
      };

      const mockResponse = {
        success: true,
        data: mockTournamentData,
        message: 'volleyballのトーナメント情報を取得しました'
      };

      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue(mockResponse);

      const result = await tournamentStore.fetchTournaments('volleyball');

      expect(result.success).toBe(true);
      expect(tournamentAPI.tournamentAPI.getTournament).toHaveBeenCalledWith('volleyball');

      const state = get(tournamentStore);
      expect(state.tournaments.volleyball).toEqual(mockTournamentData);
      expect(state.loading).toBe(false);
      expect(state.lastUpdated).toBeTruthy();
    });

    it('全スポーツのトーナメントデータを取得できる', async () => {
      const mockTournamentsData = {
        volleyball: { id: 1, sport: 'volleyball' },
        table_tennis: { id: 2, sport: 'table_tennis' },
        soccer: { id: 3, sport: 'soccer' }
      };

      const mockResponse = {
        success: true,
        data: mockTournamentsData,
        message: 'トーナメント一覧を取得しました'
      };

      tournamentAPI.tournamentAPI.getTournaments = vi.fn().mockResolvedValue(mockResponse);

      const result = await tournamentStore.fetchTournaments();

      expect(result.success).toBe(true);
      expect(tournamentAPI.tournamentAPI.getTournaments).toHaveBeenCalled();

      const state = get(tournamentStore);
      expect(state.tournaments).toEqual(mockTournamentsData);
      expect(state.loading).toBe(false);
    });

    it('キャッシュされたデータを使用できる', async () => {
      const mockTournamentData = {
        id: 1,
        sport: 'volleyball',
        matches: []
      };

      // 最初のリクエストでデータを取得
      const mockResponse = {
        success: true,
        data: mockTournamentData,
        message: 'volleyballのトーナメント情報を取得しました'
      };

      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue(mockResponse);

      await tournamentStore.fetchTournaments('volleyball');

      // 2回目のリクエストはキャッシュを使用
      const result = await tournamentStore.fetchTournaments('volleyball');

      expect(result.success).toBe(true);
      expect(result.message).toContain('キャッシュ');
      expect(tournamentAPI.tournamentAPI.getTournament).toHaveBeenCalledTimes(1);
    });

    it('APIエラー時に適切にエラーを処理する', async () => {
      const mockResponse = {
        success: false,
        error: 'API_ERROR',
        message: 'APIエラーが発生しました'
      };

      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue(mockResponse);

      const result = await tournamentStore.fetchTournaments('volleyball');

      expect(result.success).toBe(false);

      const state = get(tournamentStore);
      expect(state.error).toBe('APIエラーが発生しました');
      expect(state.loading).toBe(false);
    });

    it('無効なスポーツ名でエラーが発生する', async () => {
      const result = await tournamentStore.fetchTournaments('invalid_sport');

      expect(result.success).toBe(false);
      expect(result.error).toBe('FETCH_TOURNAMENTS_ERROR');

      const state = get(tournamentStore);
      expect(state.error).toContain('サポートされていないスポーツです');
    });
  });

  describe('試合結果更新', () => {
    it('試合結果を正常に更新できる', async () => {
      const matchId = 1;
      const matchResult = {
        score1: 3,
        score2: 1,
        winner: 'Team A'
      };

      const mockResponse = {
        success: true,
        data: { id: matchId, ...matchResult },
        message: '試合結果を更新しました'
      };

      matchAPI.matchAPI.updateMatch = vi.fn().mockResolvedValue(mockResponse);
      
      // tournamentAPIもモック
      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue({
        success: true,
        data: { id: 1, sport: 'volleyball' },
        message: 'データを取得しました'
      });

      const result = await tournamentStore.updateMatch(matchId, matchResult);

      expect(result.success).toBe(true);
      expect(matchAPI.matchAPI.updateMatch).toHaveBeenCalledWith(matchId, matchResult);
    });

    it('試合ID未指定でエラーが発生する', async () => {
      const result = await tournamentStore.updateMatch(null, { score1: 3, score2: 1 });

      expect(result.success).toBe(false);
      expect(result.error).toBe('UPDATE_MATCH_ERROR');

      const state = get(tournamentStore);
      expect(state.error).toContain('試合IDが指定されていません');
    });

    it('試合結果データが無効でエラーが発生する', async () => {
      const result = await tournamentStore.updateMatch(1, null);

      expect(result.success).toBe(false);
      expect(result.error).toBe('UPDATE_MATCH_ERROR');

      const state = get(tournamentStore);
      expect(state.error).toContain('試合結果データが正しくありません');
    });

    it('API更新失敗時に適切にエラーを処理する', async () => {
      const mockResponse = {
        success: false,
        error: 'UPDATE_FAILED',
        message: '更新に失敗しました'
      };

      matchAPI.matchAPI.updateMatch = vi.fn().mockResolvedValue(mockResponse);

      const result = await tournamentStore.updateMatch(1, { score1: 3, score2: 1 });

      expect(result.success).toBe(false);

      const state = get(tournamentStore);
      expect(state.error).toBe('更新に失敗しました');
    });
  });

  describe('スポーツ切り替え', () => {
    it('有効なスポーツに切り替えできる', () => {
      // tournamentAPIをモック
      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue({
        success: true,
        data: { id: 1, sport: 'table_tennis' },
        message: 'データを取得しました'
      });

      const result = tournamentStore.switchSport('table_tennis');

      expect(result.success).toBe(true);

      const state = get(tournamentStore);
      expect(state.currentSport).toBe('table_tennis');
      expect(state.error).toBeNull();
    });

    it('無効なスポーツ名でエラーが発生する', () => {
      const result = tournamentStore.switchSport('invalid_sport');

      expect(result.success).toBe(false);
      expect(result.error).toBe('SWITCH_SPORT_ERROR');

      const state = get(tournamentStore);
      expect(state.error).toContain('サポートされていないスポーツです');
    });

    it('スポーツ名未指定でエラーが発生する', () => {
      const result = tournamentStore.switchSport(null);

      expect(result.success).toBe(false);
      expect(result.error).toBe('SWITCH_SPORT_ERROR');

      const state = get(tournamentStore);
      expect(state.error).toContain('スポーツ名が指定されていません');
    });
  });

  describe('データリフレッシュ', () => {
    it('現在のスポーツのデータをリフレッシュできる', async () => {
      // tournamentAPIをモック
      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue({
        success: true,
        data: { id: 1, sport: 'volleyball' },
        message: 'データを取得しました'
      });

      const result = await tournamentStore.refreshData();

      expect(result.success).toBe(true);
    });

    it('指定されたスポーツのデータをリフレッシュできる', async () => {
      // tournamentAPIをモック
      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue({
        success: true,
        data: { id: 1, sport: 'soccer' },
        message: 'データを取得しました'
      });

      const result = await tournamentStore.refreshData('soccer');

      expect(result.success).toBe(true);
    });

    it('リフレッシュ中にエラーが発生した場合の処理', async () => {
      // tournamentAPIをモック（エラーレスポンスを返す）
      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue({
        success: false,
        error: 'NETWORK_ERROR',
        message: 'Network error'
      });

      const result = await tournamentStore.refreshData();

      expect(result.success).toBe(true); // refreshDataは常にsuccessを返す
      
      const state = get(tournamentStore);
      expect(state.error).toContain('Network error');
    });
  });

  describe('ポーリング機能', () => {
    it('ポーリングを開始できる', () => {
      tournamentStore.startPolling();

      const state = get(tournamentStore);
      expect(state.pollingInterval).toBeTruthy();
    });

    it('ポーリングを停止できる', () => {
      tournamentStore.startPolling();
      tournamentStore.stopPolling();

      const state = get(tournamentStore);
      expect(state.pollingInterval).toBeNull();
    });

    it('既にポーリングが開始されている場合は重複開始しない', () => {
      tournamentStore.startPolling();
      const state1 = get(tournamentStore);
      const intervalId1 = state1.pollingInterval;

      tournamentStore.startPolling();
      const state2 = get(tournamentStore);
      const intervalId2 = state2.pollingInterval;

      expect(intervalId1).toBe(intervalId2);
    });

    it('ポーリング間隔でデータ更新が実行される', async () => {
      // tournamentAPIをモック
      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue({
        success: true,
        data: { id: 1, sport: 'volleyball' },
        message: 'データを取得しました'
      });

      tournamentStore.startPolling();

      // 30秒進める
      vi.advanceTimersByTime(30000);

      // ポーリングが実行されたかチェック
      expect(tournamentAPI.tournamentAPI.getTournament).toHaveBeenCalled();
    });

    it('ページが非表示の場合はポーリングをスキップする', async () => {
      // documentを非表示に設定
      Object.defineProperty(document, 'hidden', {
        value: true,
        writable: true,
      });

      // tournamentAPIをモック
      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue({
        success: true,
        data: { id: 1, sport: 'volleyball' },
        message: 'データを取得しました'
      });

      tournamentStore.startPolling();

      // 30秒進める
      vi.advanceTimersByTime(30000);

      expect(tournamentAPI.tournamentAPI.getTournament).not.toHaveBeenCalled();
    });

    it('ローディング中の場合はポーリングをスキップする', async () => {
      // ローディング状態を設定
      tournamentStore.setLoading(true);

      // tournamentAPIをモック
      tournamentAPI.tournamentAPI.getTournament = vi.fn().mockResolvedValue({
        success: true,
        data: { id: 1, sport: 'volleyball' },
        message: 'データを取得しました'
      });

      tournamentStore.startPolling();

      // 30秒進める
      vi.advanceTimersByTime(30000);

      expect(tournamentAPI.tournamentAPI.getTournament).not.toHaveBeenCalled();
    });
  });

  describe('初期化処理', () => {
    it('初期化が正常に実行される', async () => {
      // tournamentAPIをモック
      tournamentAPI.tournamentAPI.getTournaments = vi.fn().mockResolvedValue({
        success: true,
        data: { volleyball: { id: 1, sport: 'volleyball' } },
        message: 'データを取得しました'
      });

      const result = await tournamentStore.initialize();

      expect(result.success).toBe(true);
      expect(tournamentAPI.tournamentAPI.getTournaments).toHaveBeenCalled();
    });

    it('初期化中にエラーが発生した場合の処理', async () => {
      // tournamentAPIをモック（エラーレスポンスを返す）
      tournamentAPI.tournamentAPI.getTournaments = vi.fn().mockResolvedValue({
        success: false,
        error: 'INIT_ERROR',
        message: 'Init error'
      });

      const result = await tournamentStore.initialize();

      expect(result.success).toBe(true); // initializeは常にsuccessを返す
      
      const state = get(tournamentStore);
      expect(state.error).toContain('Init error');
    });
  });

  describe('クリーンアップ処理', () => {
    it('クリーンアップが正常に実行される', () => {
      // ポーリングを開始してからクリーンアップ
      tournamentStore.startPolling();
      
      const result = tournamentStore.cleanup();

      expect(result.success).toBe(true);

      const state = get(tournamentStore);
      expect(state.tournaments).toEqual({});
      expect(state.currentSport).toBe('volleyball');
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.pollingInterval).toBeNull();
    });
  });

  describe('ユーティリティ関数', () => {
    it('現在のトーナメントデータを取得できる', () => {
      // テストデータを設定
      const mockData = { id: 1, sport: 'volleyball' };
      tournamentStore.fetchTournaments = vi.fn().mockResolvedValue({ success: true });
      
      // 直接状態を更新（テスト用）
      const state = get(tournamentStore);
      state.tournaments.volleyball = mockData;

      const currentTournament = tournamentStore.getCurrentTournament();
      expect(currentTournament).toEqual(mockData);
    });

    it('特定スポーツのトーナメントデータを取得できる', () => {
      // テストデータを設定
      const mockData = { id: 2, sport: 'table_tennis' };
      
      // 直接状態を更新（テスト用）
      const state = get(tournamentStore);
      state.tournaments.table_tennis = mockData;

      const tournament = tournamentStore.getTournamentBySport('table_tennis');
      expect(tournament).toEqual(mockData);
    });

    it('存在しないスポーツのデータ取得でnullが返される', () => {
      // 初期状態では何もデータがないのでnullが返される
      const tournament = tournamentStore.getTournamentBySport('soccer');
      expect(tournament).toBeNull();
    });

    it('無効なスポーツ名でエラーが発生する', () => {
      expect(() => {
        tournamentStore.getTournamentBySport('invalid_sport');
      }).toThrow('サポートされていないスポーツです');
    });

    it('サポートされているスポーツ一覧を取得できる', () => {
      const sports = tournamentStore.getSupportedSports();
      expect(sports).toEqual(['volleyball', 'table_tennis', 'soccer']);
    });
  });

  describe('追加機能テスト', () => {
    it('サポートされているスポーツ一覧を取得できる', () => {
      const sports = tournamentStore.getSupportedSports();
      expect(sports).toEqual(['volleyball', 'table_tennis', 'soccer']);
      expect(sports).toHaveLength(3);
    });

    it('ストアのクリーンアップが正常に動作する', () => {
      // 何らかの状態を設定
      tournamentStore.setLoading(true);
      tournamentStore.setError('テストエラー');
      
      // クリーンアップ実行
      const result = tournamentStore.cleanup();
      
      expect(result.success).toBe(true);
      expect(result.message).toContain('クリーンアップ');
      
      // 基本的な状態がリセットされていることを確認
      const state = get(tournamentStore);
      expect(state.loading).toBe(false);
      expect(state.error).toBeNull();
      expect(state.currentSport).toBe('volleyball');
      expect(state.pollingInterval).toBeNull();
    });

    it('ポーリング制御が正常に動作する', () => {
      // ポーリング開始
      tournamentStore.startPolling();
      
      let state = get(tournamentStore);
      expect(state.pollingInterval).toBeTruthy();
      
      // ポーリング停止
      tournamentStore.stopPolling();
      
      state = get(tournamentStore);
      expect(state.pollingInterval).toBeNull();
    });
  });
});
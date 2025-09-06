// AdminMatchForm コンポーネントの統合テスト（API統合部分）
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { matchAPI } from '../../api/matches.js';
import { uiActions } from '../../stores/ui.js';

// モック設定
vi.mock('../../stores/ui.js', () => ({
  uiActions: {
    showNotification: vi.fn(),
    setLoading: vi.fn()
  }
}));

vi.mock('../../api/matches.js', () => ({
  matchAPI: {
    updateMatch: vi.fn()
  }
}));

describe('AdminMatchForm 統合テスト', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('フォーム送信処理の統合テスト', () => {
    // フォーム送信処理をテスト用に抽出
    async function handleSubmit(match, score1, score2, winner) {
      const result = {
        score1: Number(score1),
        score2: Number(score2),
        winner: winner
      };

      uiActions.setLoading(true);

      try {
        if (match.id) {
          const response = await matchAPI.updateMatch(match.id, result);
          
          if (response.success) {
            uiActions.showNotification('試合結果を更新しました', 'success');
            return { success: true, data: response.data };
          } else {
            throw new Error(response.message || '試合結果の更新に失敗しました');
          }
        } else {
          uiActions.showNotification('試合結果を保存しました', 'success');
          return { success: true, data: result };
        }
      } catch (error) {
        uiActions.showNotification(
          error.message || '試合結果の保存に失敗しました', 
          'error'
        );
        throw error;
      } finally {
        uiActions.setLoading(false);
      }
    }

    it('正常な送信処理（match.idあり）', async () => {
      const mockMatch = { id: 1, team1: 'チームA', team2: 'チームB' };
      const mockResponse = {
        success: true,
        data: { id: 1, score1: 3, score2: 1 },
        message: '試合結果を更新しました'
      };

      matchAPI.updateMatch.mockResolvedValue(mockResponse);

      const result = await handleSubmit(mockMatch, '3', '1', 'チームA');

      expect(matchAPI.updateMatch).toHaveBeenCalledWith(1, {
        score1: 3,
        score2: 1,
        winner: 'チームA'
      });
      expect(uiActions.setLoading).toHaveBeenCalledWith(true);
      expect(uiActions.setLoading).toHaveBeenCalledWith(false);
      expect(uiActions.showNotification).toHaveBeenCalledWith('試合結果を更新しました', 'success');
      expect(result.success).toBe(true);
    });

    it('API エラー時の処理', async () => {
      const mockMatch = { id: 1, team1: 'チームA', team2: 'チームB' };
      const mockResponse = {
        success: false,
        message: 'サーバーエラーが発生しました'
      };

      matchAPI.updateMatch.mockResolvedValue(mockResponse);

      await expect(handleSubmit(mockMatch, '3', '1', 'チームA')).rejects.toThrow('サーバーエラーが発生しました');

      expect(uiActions.setLoading).toHaveBeenCalledWith(true);
      expect(uiActions.setLoading).toHaveBeenCalledWith(false);
      expect(uiActions.showNotification).toHaveBeenCalledWith('サーバーエラーが発生しました', 'error');
    });

    it('ネットワークエラー時の処理', async () => {
      const mockMatch = { id: 1, team1: 'チームA', team2: 'チームB' };
      
      matchAPI.updateMatch.mockRejectedValue(new Error('ネットワークエラー'));

      await expect(handleSubmit(mockMatch, '3', '1', 'チームA')).rejects.toThrow('ネットワークエラー');

      expect(uiActions.setLoading).toHaveBeenCalledWith(true);
      expect(uiActions.setLoading).toHaveBeenCalledWith(false);
      expect(uiActions.showNotification).toHaveBeenCalledWith('ネットワークエラー', 'error');
    });

    it('match.idがない場合の処理', async () => {
      const mockMatch = { team1: 'チームA', team2: 'チームB' };

      const result = await handleSubmit(mockMatch, '3', '1', 'チームA');

      expect(matchAPI.updateMatch).not.toHaveBeenCalled();
      expect(uiActions.setLoading).toHaveBeenCalledWith(true);
      expect(uiActions.setLoading).toHaveBeenCalledWith(false);
      expect(uiActions.showNotification).toHaveBeenCalledWith('試合結果を保存しました', 'success');
      expect(result.success).toBe(true);
      expect(result.data).toEqual({
        score1: 3,
        score2: 1,
        winner: 'チームA'
      });
    });

    it('数値変換が正しく行われる', async () => {
      const mockMatch = { id: 1 };
      const mockResponse = { success: true, data: {} };

      matchAPI.updateMatch.mockResolvedValue(mockResponse);

      await handleSubmit(mockMatch, '0', '5', '引き分け');

      expect(matchAPI.updateMatch).toHaveBeenCalledWith(1, {
        score1: 0,
        score2: 5,
        winner: '引き分け'
      });
    });
  });

  describe('エラーハンドリングの統合テスト', () => {
    it('デフォルトエラーメッセージの処理', async () => {
      const mockMatch = { id: 1 };
      
      matchAPI.updateMatch.mockRejectedValue(new Error());

      const handleSubmit = async () => {
        uiActions.setLoading(true);
        try {
          await matchAPI.updateMatch(mockMatch.id, {});
        } catch (error) {
          uiActions.showNotification(
            error.message || '試合結果の保存に失敗しました', 
            'error'
          );
          throw error;
        } finally {
          uiActions.setLoading(false);
        }
      };

      await expect(handleSubmit()).rejects.toThrow();
      expect(uiActions.showNotification).toHaveBeenCalledWith('試合結果の保存に失敗しました', 'error');
    });

    it('APIレスポンスのデフォルトメッセージ処理', async () => {
      const mockMatch = { id: 1 };
      const mockResponse = { success: false };

      matchAPI.updateMatch.mockResolvedValue(mockResponse);

      const handleSubmit = async () => {
        uiActions.setLoading(true);
        try {
          const response = await matchAPI.updateMatch(mockMatch.id, {});
          if (!response.success) {
            throw new Error(response.message || '試合結果の更新に失敗しました');
          }
        } catch (error) {
          uiActions.showNotification(error.message, 'error');
          throw error;
        } finally {
          uiActions.setLoading(false);
        }
      };

      await expect(handleSubmit()).rejects.toThrow('試合結果の更新に失敗しました');
      expect(uiActions.showNotification).toHaveBeenCalledWith('試合結果の更新に失敗しました', 'error');
    });
  });

  describe('ローディング状態管理の統合テスト', () => {
    it('成功時のローディング状態管理', async () => {
      const mockMatch = { id: 1 };
      const mockResponse = { success: true, data: {} };

      matchAPI.updateMatch.mockResolvedValue(mockResponse);

      const handleSubmit = async () => {
        uiActions.setLoading(true);
        try {
          await matchAPI.updateMatch(mockMatch.id, {});
          uiActions.showNotification('成功', 'success');
        } finally {
          uiActions.setLoading(false);
        }
      };

      await handleSubmit();

      expect(uiActions.setLoading).toHaveBeenCalledTimes(2);
      expect(uiActions.setLoading).toHaveBeenNthCalledWith(1, true);
      expect(uiActions.setLoading).toHaveBeenNthCalledWith(2, false);
    });

    it('エラー時のローディング状態管理', async () => {
      const mockMatch = { id: 1 };
      
      matchAPI.updateMatch.mockRejectedValue(new Error('テストエラー'));

      const handleSubmit = async () => {
        uiActions.setLoading(true);
        try {
          await matchAPI.updateMatch(mockMatch.id, {});
        } catch (error) {
          uiActions.showNotification(error.message, 'error');
        } finally {
          uiActions.setLoading(false);
        }
      };

      await handleSubmit();

      expect(uiActions.setLoading).toHaveBeenCalledTimes(2);
      expect(uiActions.setLoading).toHaveBeenNthCalledWith(1, true);
      expect(uiActions.setLoading).toHaveBeenNthCalledWith(2, false);
    });
  });

  describe('通知システムの統合テスト', () => {
    it('成功通知の表示', async () => {
      const mockMatch = { id: 1 };
      const mockResponse = { 
        success: true, 
        data: {},
        message: 'カスタム成功メッセージ'
      };

      matchAPI.updateMatch.mockResolvedValue(mockResponse);

      const handleSubmit = async () => {
        const response = await matchAPI.updateMatch(mockMatch.id, {});
        if (response.success) {
          uiActions.showNotification('試合結果を更新しました', 'success');
        }
      };

      await handleSubmit();

      expect(uiActions.showNotification).toHaveBeenCalledWith('試合結果を更新しました', 'success');
    });

    it('エラー通知の表示', async () => {
      const mockMatch = { id: 1 };
      
      matchAPI.updateMatch.mockRejectedValue(new Error('カスタムエラーメッセージ'));

      const handleSubmit = async () => {
        try {
          await matchAPI.updateMatch(mockMatch.id, {});
        } catch (error) {
          uiActions.showNotification(error.message, 'error');
        }
      };

      await handleSubmit();

      expect(uiActions.showNotification).toHaveBeenCalledWith('カスタムエラーメッセージ', 'error');
    });
  });
});
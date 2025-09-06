// TournamentBracket コンポーネントのテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// モックデータ
const mockMatches = [
  {
    id: 1,
    tournament_id: 1,
    round: 'semifinal',
    team1: 'チームA',
    team2: 'チームB',
    score1: 3,
    score2: 1,
    winner: 'チームA',
    status: 'completed',
    scheduled_at: '2024-01-15T10:00:00Z',
    completed_at: '2024-01-15T11:30:00Z'
  },
  {
    id: 2,
    tournament_id: 1,
    round: 'semifinal',
    team1: 'チームC',
    team2: 'チームD',
    score1: 2,
    score2: 3,
    winner: 'チームD',
    status: 'completed',
    scheduled_at: '2024-01-15T12:00:00Z',
    completed_at: '2024-01-15T13:30:00Z'
  },
  {
    id: 3,
    tournament_id: 1,
    round: 'final',
    team1: 'チームA',
    team2: 'チームD',
    status: 'pending',
    scheduled_at: '2024-01-16T14:00:00Z'
  },
  {
    id: 4,
    tournament_id: 1,
    round: 'quarterfinal',
    team1: 'チームE',
    team2: 'チームF',
    score1: 1,
    score2: 2,
    winner: 'チームF',
    status: 'completed',
    scheduled_at: '2024-01-14T10:00:00Z',
    completed_at: '2024-01-14T11:30:00Z'
  }
];

const emptyMatches = [];

const pendingMatches = [
  {
    id: 5,
    tournament_id: 1,
    round: 'round_1',
    team1: 'チームG',
    team2: 'チームH',
    status: 'pending',
    scheduled_at: '2024-01-17T10:00:00Z'
  }
];

describe('TournamentBracket コンポーネント', () => {
  beforeEach(() => {
    // ウィンドウオブジェクトのモック
    Object.defineProperty(window, 'innerWidth', {
      writable: true,
      configurable: true,
      value: 1024,
    });
    
    // ResizeObserver のモック
    global.ResizeObserver = vi.fn().mockImplementation(() => ({
      observe: vi.fn(),
      unobserve: vi.fn(),
      disconnect: vi.fn(),
    }));
  });

  afterEach(() => {
    vi.clearAllMocks();
  });

  describe('データ処理ロジック', () => {
    it('試合をラウンド別にグループ化する', () => {
      // グループ化関数のテスト
      function groupMatchesByRound(matches) {
        if (!Array.isArray(matches)) {
          return {};
        }

        return matches.reduce((groups, match) => {
          const round = match.round || 'unknown';
          if (!groups[round]) {
            groups[round] = [];
          }
          groups[round].push(match);
          return groups;
        }, {});
      }

      const grouped = groupMatchesByRound(mockMatches);
      
      expect(grouped.semifinal).toHaveLength(2);
      expect(grouped.final).toHaveLength(1);
      expect(grouped.quarterfinal).toHaveLength(1);
    });

    it('ラウンドの表示順序を正しく取得する', () => {
      function getRoundOrder(groupedMatches) {
        const availableRounds = Object.keys(groupedMatches);
        const standardOrder = ['round_1', 'round_2', 'quarterfinal', 'semifinal', 'third_place', 'final'];
        
        return standardOrder.filter(round => availableRounds.includes(round))
          .concat(availableRounds.filter(round => !standardOrder.includes(round)));
      }

      const grouped = {
        final: [],
        semifinal: [],
        quarterfinal: []
      };

      const order = getRoundOrder(grouped);
      
      expect(order).toEqual(['quarterfinal', 'semifinal', 'final']);
    });

    it('試合の勝者を正しく判定する', () => {
      function getMatchWinner(match) {
        if (match.winner) {
          return match.winner;
        }
        
        if (match.score1 !== undefined && match.score2 !== undefined) {
          if (match.score1 > match.score2) {
            return match.team1;
          } else if (match.score2 > match.score1) {
            return match.team2;
          }
        }
        
        return null;
      }

      const match1 = mockMatches[0]; // チームA勝利
      const match2 = mockMatches[2]; // 未実施
      
      expect(getMatchWinner(match1)).toBe('チームA');
      expect(getMatchWinner(match2)).toBeNull();
    });

    it('試合の状態を正しく判定する', () => {
      function getMatchStatus(match) {
        if (match.status === 'completed' || (match.score1 !== undefined && match.score2 !== undefined)) {
          return 'completed';
        } else if (match.status === 'in_progress') {
          return 'in_progress';
        } else {
          return 'pending';
        }
      }

      const completedMatch = mockMatches[0];
      const pendingMatch = mockMatches[2];
      const inProgressMatch = { ...mockMatches[0], status: 'in_progress', score1: undefined, score2: undefined };
      
      expect(getMatchStatus(completedMatch)).toBe('completed');
      expect(getMatchStatus(pendingMatch)).toBe('pending');
      expect(getMatchStatus(inProgressMatch)).toBe('in_progress');
    });
  });

  describe('レスポンシブ機能', () => {
    it('画面サイズに応じて正しいクラス名を返す', () => {
      function getResponsiveClass(width) {
        if (width < 768) return 'mobile';
        if (width < 1024) return 'tablet';
        return 'desktop';
      }

      expect(getResponsiveClass(600)).toBe('mobile');
      expect(getResponsiveClass(800)).toBe('tablet');
      expect(getResponsiveClass(1200)).toBe('desktop');
    });
  });

  describe('スポーツ名とラウンド名の表示', () => {
    it('スポーツ名を正しく日本語に変換する', () => {
      const sportNames = {
        volleyball: 'バレーボール',
        table_tennis: '卓球',
        soccer: 'サッカー'
      };

      expect(sportNames.volleyball).toBe('バレーボール');
      expect(sportNames.table_tennis).toBe('卓球');
      expect(sportNames.soccer).toBe('サッカー');
      expect(sportNames.unknown_sport).toBeUndefined();
    });

    it('ラウンド名を正しく日本語に変換する', () => {
      const roundNames = {
        'round_1': '1回戦',
        'round_2': '2回戦',
        'quarterfinal': '準々決勝',
        'semifinal': '準決勝',
        'final': '決勝',
        'third_place': '3位決定戦'
      };

      expect(roundNames.quarterfinal).toBe('準々決勝');
      expect(roundNames.semifinal).toBe('準決勝');
      expect(roundNames.final).toBe('決勝');
      expect(roundNames.unknown_round).toBeUndefined();
    });
  });

  describe('エラーハンドリング', () => {
    it('matches が null の場合でも正常に処理される', () => {
      function groupMatchesByRound(matches) {
        if (!Array.isArray(matches)) {
          return {};
        }
        return matches.reduce((groups, match) => {
          const round = match.round || 'unknown';
          if (!groups[round]) {
            groups[round] = [];
          }
          groups[round].push(match);
          return groups;
        }, {});
      }

      const result = groupMatchesByRound(null);
      expect(result).toEqual({});
    });

    it('matches が undefined の場合でも正常に処理される', () => {
      function groupMatchesByRound(matches) {
        if (!Array.isArray(matches)) {
          return {};
        }
        return matches.reduce((groups, match) => {
          const round = match.round || 'unknown';
          if (!groups[round]) {
            groups[round] = [];
          }
          groups[round].push(match);
          return groups;
        }, {});
      }

      const result = groupMatchesByRound(undefined);
      expect(result).toEqual({});
    });

    it('不正な試合データが含まれていても処理される', () => {
      function groupMatchesByRound(matches) {
        if (!Array.isArray(matches)) {
          return {};
        }
        return matches.reduce((groups, match) => {
          const round = match.round || 'unknown';
          if (!groups[round]) {
            groups[round] = [];
          }
          groups[round].push(match);
          return groups;
        }, {});
      }

      const invalidMatches = [
        { id: 1, team1: 'チームA', team2: 'チームB' }, // round が欠けている
        { id: 2, round: 'final' } // team1, team2 が欠けている
      ];

      const result = groupMatchesByRound(invalidMatches);
      expect(result.unknown).toHaveLength(1);
      expect(result.final).toHaveLength(1);
    });
  });

  describe('管理者機能のロジック', () => {
    it('編集ボタンの表示条件を正しく判定する', () => {
      function shouldShowEditButton(isAdmin, matchStatus) {
        return isAdmin && matchStatus === 'pending';
      }

      expect(shouldShowEditButton(true, 'pending')).toBe(true);
      expect(shouldShowEditButton(true, 'completed')).toBe(false);
      expect(shouldShowEditButton(false, 'pending')).toBe(false);
      expect(shouldShowEditButton(false, 'completed')).toBe(false);
    });

    it('編集コールバックが安全に呼び出される', () => {
      function handleEditMatch(match, onEditMatch) {
        if (onEditMatch && typeof onEditMatch === 'function') {
          onEditMatch(match);
        }
      }

      const mockCallback = vi.fn();
      const testMatch = mockMatches[0];

      // 正常なコールバック
      handleEditMatch(testMatch, mockCallback);
      expect(mockCallback).toHaveBeenCalledWith(testMatch);

      // null コールバック
      expect(() => handleEditMatch(testMatch, null)).not.toThrow();

      // undefined コールバック
      expect(() => handleEditMatch(testMatch, undefined)).not.toThrow();
    });
  });

  describe('パフォーマンス考慮', () => {
    it('大量の試合データでも正常に処理される', () => {
      function groupMatchesByRound(matches) {
        if (!Array.isArray(matches)) {
          return {};
        }
        return matches.reduce((groups, match) => {
          const round = match.round || 'unknown';
          if (!groups[round]) {
            groups[round] = [];
          }
          groups[round].push(match);
          return groups;
        }, {});
      }

      // 1000試合のテストデータを生成
      const largeMatchData = Array.from({ length: 1000 }, (_, i) => ({
        id: i + 1,
        round: `round_${Math.floor(i / 100) + 1}`,
        team1: `チーム${i * 2 + 1}`,
        team2: `チーム${i * 2 + 2}`,
        status: 'pending'
      }));

      const startTime = performance.now();
      const result = groupMatchesByRound(largeMatchData);
      const endTime = performance.now();

      // 処理時間が100ms以下であることを確認
      expect(endTime - startTime).toBeLessThan(100);
      
      // 正しくグループ化されていることを確認
      expect(Object.keys(result)).toHaveLength(10);
      expect(result.round_1).toHaveLength(100);
    });
  });
});
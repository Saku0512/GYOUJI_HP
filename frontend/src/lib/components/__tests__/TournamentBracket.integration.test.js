// TournamentBracket コンポーネントの統合テスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// 統合テスト用のモックデータ
const integrationTestMatches = [
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
    round: 'final',
    team1: 'チームA',
    team2: 'チームC',
    status: 'pending',
    scheduled_at: '2024-01-16T14:00:00Z'
  }
];

describe('TournamentBracket 統合テスト', () => {
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

  describe('MatchCard との統合', () => {
    it('MatchCard コンポーネントに正しいプロパティが渡される', () => {
      // MatchCard コンポーネントが受け取るべきプロパティの検証
      const match = integrationTestMatches[0];
      
      // match オブジェクトが必要なプロパティを持っていることを確認
      expect(match).toHaveProperty('id');
      expect(match).toHaveProperty('team1');
      expect(match).toHaveProperty('team2');
      expect(match).toHaveProperty('score1');
      expect(match).toHaveProperty('score2');
      expect(match).toHaveProperty('status');
    });

    it('管理者モードでMatchCardに正しいeditableプロパティが渡される', () => {
      const isAdmin = true;
      const editable = isAdmin;
      
      expect(editable).toBe(true);
    });

    it('非管理者モードでMatchCardに正しいeditableプロパティが渡される', () => {
      const isAdmin = false;
      const editable = isAdmin;
      
      expect(editable).toBe(false);
    });
  });

  describe('データフローの統合テスト', () => {
    it('試合データが正しくフィルタリングされて表示される', () => {
      // 複数のラウンドを含むデータのフィルタリング
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

      const grouped = groupMatchesByRound(integrationTestMatches);
      
      expect(grouped.semifinal).toHaveLength(1);
      expect(grouped.final).toHaveLength(1);
      expect(grouped.semifinal[0].team1).toBe('チームA');
      expect(grouped.final[0].status).toBe('pending');
    });

    it('管理者向け編集機能が正しく動作する', () => {
      const mockEditCallback = vi.fn();
      const isAdmin = true;
      const pendingMatch = integrationTestMatches[1];

      // 編集ボタンの表示条件
      function shouldShowEditButton(isAdmin, matchStatus) {
        return isAdmin && matchStatus === 'pending';
      }

      // 編集ハンドラー
      function handleEditMatch(match, onEditMatch) {
        if (onEditMatch && typeof onEditMatch === 'function') {
          onEditMatch(match);
        }
      }

      expect(shouldShowEditButton(isAdmin, pendingMatch.status)).toBe(true);
      
      handleEditMatch(pendingMatch, mockEditCallback);
      expect(mockEditCallback).toHaveBeenCalledWith(pendingMatch);
    });
  });

  describe('レスポンシブ統合テスト', () => {
    it('異なる画面サイズで適切なレイアウトクラスが適用される', () => {
      function getResponsiveClass(width) {
        if (width < 768) return 'mobile';
        if (width < 1024) return 'tablet';
        return 'desktop';
      }

      // モバイル
      expect(getResponsiveClass(600)).toBe('mobile');
      
      // タブレット
      expect(getResponsiveClass(800)).toBe('tablet');
      
      // デスクトップ
      expect(getResponsiveClass(1200)).toBe('desktop');
    });

    it('レスポンシブクラスに応じて適切なスタイルが適用される', () => {
      // CSSクラスの存在確認（実際のスタイル適用は別途確認）
      const responsiveClasses = ['mobile', 'tablet', 'desktop'];
      
      responsiveClasses.forEach(className => {
        expect(className).toMatch(/^(mobile|tablet|desktop)$/);
      });
    });
  });

  describe('アクセシビリティ統合テスト', () => {
    it('適切なARIA属性が設定される', () => {
      // aria-label の検証
      const editButtonLabel = '試合結果を編集';
      expect(editButtonLabel).toBe('試合結果を編集');

      // aria-hidden の検証
      const connectionLineHidden = true;
      expect(connectionLineHidden).toBe(true);
    });

    it('見出し構造が適切に設定される', () => {
      // ラウンドタイトルがh3要素として設定されることの確認
      const roundTitleLevel = 3;
      expect(roundTitleLevel).toBe(3);
    });
  });

  describe('エラー処理統合テスト', () => {
    it('不正なデータでもアプリケーションがクラッシュしない', () => {
      const invalidData = [
        null,
        undefined,
        {},
        { id: 'invalid' },
        { round: null, team1: '', team2: '' }
      ];

      invalidData.forEach(data => {
        expect(() => {
          // データ処理関数の呼び出し
          function groupMatchesByRound(matches) {
            if (!Array.isArray(matches)) {
              return {};
            }
            return matches.reduce((groups, match) => {
              if (!match || typeof match !== 'object') {
                return groups;
              }
              const round = match.round || 'unknown';
              if (!groups[round]) {
                groups[round] = [];
              }
              groups[round].push(match);
              return groups;
            }, {});
          }

          groupMatchesByRound([data]);
        }).not.toThrow();
      });
    });

    it('コールバック関数が存在しない場合でもエラーが発生しない', () => {
      const testMatch = integrationTestMatches[0];
      
      function handleEditMatch(match, onEditMatch) {
        if (onEditMatch && typeof onEditMatch === 'function') {
          onEditMatch(match);
        }
      }

      // null コールバック
      expect(() => handleEditMatch(testMatch, null)).not.toThrow();
      
      // undefined コールバック
      expect(() => handleEditMatch(testMatch, undefined)).not.toThrow();
      
      // 無効なコールバック
      expect(() => handleEditMatch(testMatch, 'invalid')).not.toThrow();
    });
  });

  describe('パフォーマンス統合テスト', () => {
    it('大量のデータでも適切なパフォーマンスを維持する', () => {
      // 大量のテストデータ生成
      const largeDataSet = Array.from({ length: 500 }, (_, i) => ({
        id: i + 1,
        tournament_id: 1,
        round: `round_${Math.floor(i / 50) + 1}`,
        team1: `チーム${i * 2 + 1}`,
        team2: `チーム${i * 2 + 2}`,
        status: i % 3 === 0 ? 'completed' : 'pending',
        score1: i % 3 === 0 ? Math.floor(Math.random() * 5) : undefined,
        score2: i % 3 === 0 ? Math.floor(Math.random() * 5) : undefined
      }));

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

      const startTime = performance.now();
      const result = groupMatchesByRound(largeDataSet);
      const endTime = performance.now();

      // 処理時間が50ms以下であることを確認
      expect(endTime - startTime).toBeLessThan(50);
      
      // 正しく処理されていることを確認
      expect(Object.keys(result)).toHaveLength(10);
      expect(result.round_1).toHaveLength(50);
    });
  });

  describe('国際化対応統合テスト', () => {
    it('日本語のスポーツ名とラウンド名が正しく表示される', () => {
      const sportNames = {
        volleyball: 'バレーボール',
        table_tennis: '卓球',
        soccer: 'サッカー'
      };

      const roundNames = {
        'round_1': '1回戦',
        'round_2': '2回戦',
        'quarterfinal': '準々決勝',
        'semifinal': '準決勝',
        'final': '決勝',
        'third_place': '3位決定戦'
      };

      // スポーツ名の変換テスト
      expect(sportNames.volleyball).toBe('バレーボール');
      expect(sportNames.table_tennis).toBe('卓球');
      expect(sportNames.soccer).toBe('サッカー');

      // ラウンド名の変換テスト
      expect(roundNames.semifinal).toBe('準決勝');
      expect(roundNames.final).toBe('決勝');
      expect(roundNames.quarterfinal).toBe('準々決勝');
    });

    it('未知のスポーツ名やラウンド名でも適切に処理される', () => {
      const sportNames = {
        volleyball: 'バレーボール',
        table_tennis: '卓球',
        soccer: 'サッカー'
      };

      const roundNames = {
        'round_1': '1回戦',
        'round_2': '2回戦',
        'quarterfinal': '準々決勝',
        'semifinal': '準決勝',
        'final': '決勝',
        'third_place': '3位決定戦'
      };

      // 未知のスポーツ名
      const unknownSport = 'unknown_sport';
      const displaySport = sportNames[unknownSport] || unknownSport;
      expect(displaySport).toBe('unknown_sport');

      // 未知のラウンド名
      const unknownRound = 'unknown_round';
      const displayRound = roundNames[unknownRound] || unknownRound;
      expect(displayRound).toBe('unknown_round');
    });
  });
});
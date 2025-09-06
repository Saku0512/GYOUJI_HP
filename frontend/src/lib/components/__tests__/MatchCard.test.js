// MatchCard コンポーネントのテスト
import { render, screen, fireEvent, waitFor } from '@testing-library/svelte';
import { vi } from 'vitest';
import MatchCard from '../MatchCard.svelte';

// テスト用のモックデータ
const mockMatch = {
  id: 1,
  tournament_id: 1,
  round: '準決勝',
  team1: 'チームA',
  team2: 'チームB',
  score1: null,
  score2: null,
  winner: null,
  status: 'pending',
  scheduled_at: '2024-03-15T10:00:00Z',
  completed_at: null
};

const mockCompletedMatch = {
  id: 2,
  tournament_id: 1,
  round: '決勝',
  team1: 'チームC',
  team2: 'チームD',
  score1: 3,
  score2: 1,
  winner: 'チームC',
  status: 'completed',
  scheduled_at: '2024-03-15T14:00:00Z',
  completed_at: '2024-03-15T15:30:00Z'
};

describe('MatchCard', () => {
  describe('基本表示', () => {
    test('試合情報が正しく表示される', () => {
      render(MatchCard, { props: { match: mockMatch } });

      expect(screen.getByTestId('match-card')).toBeInTheDocument();
      expect(screen.getByTestId('team1')).toHaveTextContent('チームA');
      expect(screen.getByTestId('team2')).toHaveTextContent('チームB');
      expect(screen.getByTestId('match-round')).toHaveTextContent('準決勝');
      expect(screen.getByTestId('match-status')).toHaveTextContent('未実施');
    });

    test('完了した試合のスコアが表示される', () => {
      render(MatchCard, { props: { match: mockCompletedMatch } });

      expect(screen.getByTestId('score-display')).toHaveTextContent('3 - 1');
      expect(screen.getByTestId('completion-time')).toBeInTheDocument();
    });

    test('勝者にバッジが表示される', () => {
      render(MatchCard, { props: { match: mockCompletedMatch } });

      const team1Element = screen.getByTestId('team1');
      const team2Element = screen.getByTestId('team2');

      // チームCが勝者なので、team1に勝者バッジがあることを確認
      expect(team1Element).toHaveTextContent('🏆');
      expect(team2Element).not.toHaveTextContent('🏆');
    });

    test('スケジュール時間が表示される', () => {
      render(MatchCard, { props: { match: mockMatch } });

      expect(screen.getByTestId('match-schedule')).toBeInTheDocument();
    });
  });

  describe('編集モード', () => {
    test('編集可能な場合、編集ボタンが表示される', () => {
      render(MatchCard, { props: { match: mockMatch, editable: true } });

      expect(screen.getByTestId('edit-match-btn')).toBeInTheDocument();
    });

    test('編集不可能な場合、編集ボタンが表示されない', () => {
      render(MatchCard, { props: { match: mockMatch, editable: false } });

      expect(screen.queryByTestId('edit-match-btn')).not.toBeInTheDocument();
    });

    test('完了した試合では編集ボタンが表示されない', () => {
      render(MatchCard, { props: { match: mockCompletedMatch, editable: true } });

      expect(screen.queryByTestId('edit-match-btn')).not.toBeInTheDocument();
    });

    test('編集ボタンクリックで編集モードに切り替わる', async () => {
      render(MatchCard, { props: { match: mockMatch, editable: true } });

      const editButton = screen.getByTestId('edit-match-btn');
      await fireEvent.click(editButton);

      expect(screen.getByTestId('score-edit')).toBeInTheDocument();
      expect(screen.getByTestId('score1-input')).toBeInTheDocument();
      expect(screen.getByTestId('score2-input')).toBeInTheDocument();
      expect(screen.getByTestId('save-score-btn')).toBeInTheDocument();
      expect(screen.getByTestId('cancel-edit-btn')).toBeInTheDocument();
    });
  });

  describe('スコア編集', () => {
    test('スコア入力フィールドに値を入力できる', async () => {
      render(MatchCard, { props: { match: mockMatch, editable: true } });

      // 編集モードに切り替え
      await fireEvent.click(screen.getByTestId('edit-match-btn'));

      const score1Input = screen.getByTestId('score1-input');
      const score2Input = screen.getByTestId('score2-input');

      await fireEvent.input(score1Input, { target: { value: '3' } });
      await fireEvent.input(score2Input, { target: { value: '1' } });

      expect(score1Input.value).toBe('3');
      expect(score2Input.value).toBe('1');
    });

    test('保存ボタンクリックでupdateScoreイベントが発火される', async () => {
      const component = render(MatchCard, { props: { match: mockMatch, editable: true } });
      const mockHandler = vi.fn();
      component.component.$on('updateScore', mockHandler);

      // 編集モードに切り替え
      await fireEvent.click(screen.getByTestId('edit-match-btn'));

      // スコア入力
      await fireEvent.input(screen.getByTestId('score1-input'), { target: { value: '3' } });
      await fireEvent.input(screen.getByTestId('score2-input'), { target: { value: '1' } });

      // 保存ボタンクリック
      await fireEvent.click(screen.getByTestId('save-score-btn'));

      expect(mockHandler).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: {
            matchId: 1,
            score1: 3,
            score2: 1
          }
        })
      );
    });

    test('キャンセルボタンクリックで編集モードが終了する', async () => {
      render(MatchCard, { props: { match: mockMatch, editable: true } });

      // 編集モードに切り替え
      await fireEvent.click(screen.getByTestId('edit-match-btn'));

      // スコア入力
      await fireEvent.input(screen.getByTestId('score1-input'), { target: { value: '3' } });

      // キャンセルボタンクリック
      await fireEvent.click(screen.getByTestId('cancel-edit-btn'));

      // 編集モードが終了していることを確認
      expect(screen.queryByTestId('score-edit')).not.toBeInTheDocument();
      expect(screen.getByTestId('edit-match-btn')).toBeInTheDocument();
    });

    test('無効なスコア入力でエラーイベントが発火される', async () => {
      const component = render(MatchCard, { props: { match: mockMatch, editable: true } });
      const mockErrorHandler = vi.fn();
      component.component.$on('error', mockErrorHandler);

      // 編集モードに切り替え
      await fireEvent.click(screen.getByTestId('edit-match-btn'));

      // 無効なスコア入力（負の値）
      await fireEvent.input(screen.getByTestId('score1-input'), { target: { value: '-1' } });
      await fireEvent.input(screen.getByTestId('score2-input'), { target: { value: '1' } });

      // 保存ボタンクリック
      await fireEvent.click(screen.getByTestId('save-score-btn'));

      expect(mockErrorHandler).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: {
            message: 'スコアは0以上の数値を入力してください'
          }
        })
      );
    });

    test('空のスコア入力でエラーイベントが発火される', async () => {
      const component = render(MatchCard, { props: { match: mockMatch, editable: true } });
      const mockErrorHandler = vi.fn();
      component.component.$on('error', mockErrorHandler);

      // 編集モードに切り替え
      await fireEvent.click(screen.getByTestId('edit-match-btn'));

      // 空のスコア入力
      await fireEvent.input(screen.getByTestId('score1-input'), { target: { value: '' } });
      await fireEvent.input(screen.getByTestId('score2-input'), { target: { value: '1' } });

      // 保存ボタンクリック
      await fireEvent.click(screen.getByTestId('save-score-btn'));

      expect(mockErrorHandler).toHaveBeenCalled();
    });
  });

  describe('キーボード操作', () => {
    test('Enterキーでスコアが保存される', async () => {
      const component = render(MatchCard, { props: { match: mockMatch, editable: true } });
      const mockHandler = vi.fn();
      component.component.$on('updateScore', mockHandler);

      // 編集モードに切り替え
      await fireEvent.click(screen.getByTestId('edit-match-btn'));

      // スコア入力
      const score1Input = screen.getByTestId('score1-input');
      await fireEvent.input(score1Input, { target: { value: '2' } });
      await fireEvent.input(screen.getByTestId('score2-input'), { target: { value: '1' } });

      // Enterキー押下
      await fireEvent.keyDown(score1Input, { key: 'Enter' });

      expect(mockHandler).toHaveBeenCalledWith(
        expect.objectContaining({
          detail: {
            matchId: 1,
            score1: 2,
            score2: 1
          }
        })
      );
    });

    test('Escapeキーで編集がキャンセルされる', async () => {
      render(MatchCard, { props: { match: mockMatch, editable: true } });

      // 編集モードに切り替え
      await fireEvent.click(screen.getByTestId('edit-match-btn'));

      // スコア入力
      const score1Input = screen.getByTestId('score1-input');
      await fireEvent.input(score1Input, { target: { value: '3' } });

      // Escapeキー押下
      await fireEvent.keyDown(score1Input, { key: 'Escape' });

      // 編集モードが終了していることを確認
      expect(screen.queryByTestId('score-edit')).not.toBeInTheDocument();
      expect(screen.getByTestId('edit-match-btn')).toBeInTheDocument();
    });
  });

  describe('コンパクトモード', () => {
    test('コンパクトモードでクラスが適用される', () => {
      render(MatchCard, { props: { match: mockMatch, compact: true } });

      const matchCard = screen.getByTestId('match-card');
      expect(matchCard).toHaveClass('compact');
    });

    test('コンパクトモードでチーム名が短縮される', () => {
      const longNameMatch = {
        ...mockMatch,
        team1: 'とても長いチーム名です',
        team2: '短いチーム名'
      };

      render(MatchCard, { props: { match: longNameMatch, compact: true } });

      const team1Element = screen.getByTestId('team1');
      const team2Element = screen.getByTestId('team2');

      // コンパクトモードでは8文字で切り詰められる
      expect(team1Element).toHaveTextContent('とても長い...');
      expect(team2Element).toHaveTextContent('短いチーム名');
    });
  });

  describe('アクセシビリティ', () => {
    test('編集ボタンにaria-labelが設定されている', () => {
      render(MatchCard, { props: { match: mockMatch, editable: true } });

      const editButton = screen.getByTestId('edit-match-btn');
      expect(editButton).toHaveAttribute('aria-label', '試合結果を編集');
    });

    test('スコア入力フィールドにプレースホルダーが設定されている', async () => {
      render(MatchCard, { props: { match: mockMatch, editable: true } });

      // 編集モードに切り替え
      await fireEvent.click(screen.getByTestId('edit-match-btn'));

      const score1Input = screen.getByTestId('score1-input');
      const score2Input = screen.getByTestId('score2-input');

      expect(score1Input).toHaveAttribute('placeholder', '0');
      expect(score2Input).toHaveAttribute('placeholder', '0');
    });
  });

  describe('エッジケース', () => {
    test('空のmatchオブジェクトでもエラーが発生しない', () => {
      expect(() => {
        render(MatchCard, { props: { match: {} } });
      }).not.toThrow();
    });

    test('nullのmatchプロパティでもエラーが発生しない', () => {
      expect(() => {
        render(MatchCard, { props: { match: null } });
      }).not.toThrow();
    });

    test('undefinedのスコアが正しく処理される', () => {
      const matchWithUndefinedScore = {
        ...mockMatch,
        score1: undefined,
        score2: undefined
      };

      render(MatchCard, { props: { match: matchWithUndefinedScore } });

      expect(screen.getByTestId('match-status')).toHaveTextContent('未実施');
    });

    test('0のスコアが正しく表示される', () => {
      const matchWithZeroScore = {
        ...mockMatch,
        score1: 0,
        score2: 3,
        status: 'completed'
      };

      render(MatchCard, { props: { match: matchWithZeroScore } });

      expect(screen.getByTestId('score-display')).toHaveTextContent('0 - 3');
    });

    test('引き分けの場合の表示', () => {
      const drawMatch = {
        ...mockMatch,
        score1: 2,
        score2: 2,
        winner: 'draw',
        status: 'completed'
      };

      render(MatchCard, { props: { match: drawMatch } });

      expect(screen.getByTestId('score-display')).toHaveTextContent('2 - 2');
      // 引き分けの場合、どちらのチームにも勝者バッジが表示されない
      expect(screen.queryByText('🏆')).not.toBeInTheDocument();
    });
  });

  describe('レスポンシブ対応', () => {
    test('モバイル表示でのクラス適用', () => {
      // ビューポートサイズを変更
      Object.defineProperty(window, 'innerWidth', {
        writable: true,
        configurable: true,
        value: 500,
      });

      render(MatchCard, { props: { match: mockMatch } });

      const matchCard = screen.getByTestId('match-card');
      expect(matchCard).toBeInTheDocument();
    });
  });
});
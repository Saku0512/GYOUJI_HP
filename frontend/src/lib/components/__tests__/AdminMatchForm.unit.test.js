// AdminMatchForm コンポーネントのユニットテスト（ロジック部分のみ）
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { validateMatchResult } from '../../utils/validation.js';

describe('AdminMatchForm ロジックテスト', () => {
  describe('勝者決定ロジック', () => {
    // 勝者決定関数をテスト用に抽出
    function determineWinner(s1, s2, team1 = 'Team A', team2 = 'Team B') {
      const num1 = Number(s1);
      const num2 = Number(s2);
      
      if (isNaN(num1) || isNaN(num2) || s1 === '' || s2 === '') {
        return '';
      }
      
      if (num1 > num2) {
        return team1;
      } else if (num2 > num1) {
        return team2;
      } else {
        return '引き分け';
      }
    }

    it('チーム1が勝利した場合', () => {
      const result = determineWinner('3', '1', 'チームA', 'チームB');
      expect(result).toBe('チームA');
    });

    it('チーム2が勝利した場合', () => {
      const result = determineWinner('1', '3', 'チームA', 'チームB');
      expect(result).toBe('チームB');
    });

    it('引き分けの場合', () => {
      const result = determineWinner('2', '2', 'チームA', 'チームB');
      expect(result).toBe('引き分け');
    });

    it('空の値の場合は空文字を返す', () => {
      expect(determineWinner('', '2')).toBe('');
      expect(determineWinner('2', '')).toBe('');
      expect(determineWinner('', '')).toBe('');
    });

    it('無効な値の場合は空文字を返す', () => {
      expect(determineWinner('abc', '2')).toBe('');
      expect(determineWinner('2', 'xyz')).toBe('');
      expect(determineWinner('abc', 'xyz')).toBe('');
    });

    it('0スコアでも正しく動作する', () => {
      expect(determineWinner('0', '1', 'A', 'B')).toBe('B');
      expect(determineWinner('1', '0', 'A', 'B')).toBe('A');
      expect(determineWinner('0', '0', 'A', 'B')).toBe('引き分け');
    });
  });

  describe('フォーム検証ロジック', () => {
    it('有効なスコアの場合', () => {
      const result = validateMatchResult('3', '1');
      expect(result.isValid).toBe(true);
      expect(result.errors).toEqual({});
    });

    it('負の数値の場合はエラー', () => {
      const result = validateMatchResult('-1', '2');
      expect(result.isValid).toBe(false);
      expect(result.errors.score1).toBe('チーム1のスコアは0以上である必要があります');
    });

    it('空の値の場合はエラー', () => {
      const result = validateMatchResult('', '2');
      expect(result.isValid).toBe(false);
      expect(result.errors.score1).toBe('チーム1のスコアは必須です');
    });

    it('文字列の場合はエラー', () => {
      const result = validateMatchResult('abc', '2');
      expect(result.isValid).toBe(false);
      expect(result.errors.score1).toBe('チーム1のスコアは数値である必要があります');
    });

    it('両方のスコアが無効な場合', () => {
      const result = validateMatchResult('', 'xyz');
      expect(result.isValid).toBe(false);
      expect(result.errors.score1).toBe('チーム1のスコアは必須です');
      expect(result.errors.score2).toBe('チーム2のスコアは数値である必要があります');
    });

    it('0は有効なスコア', () => {
      const result = validateMatchResult('0', '1');
      expect(result.isValid).toBe(true);
      expect(result.errors).toEqual({});
    });
  });

  describe('フォーム状態管理ロジック', () => {
    // フォームの有効性チェック関数をテスト用に抽出
    function isFormValid(errors, score1, score2) {
      return Object.keys(errors).length === 0 && score1 !== '' && score2 !== '';
    }

    it('有効なスコアが入力されている場合はtrue', () => {
      const errors = {};
      const result = isFormValid(errors, '3', '1');
      expect(result).toBe(true);
    });

    it('エラーがある場合はfalse', () => {
      const errors = { score1: 'エラー' };
      const result = isFormValid(errors, '3', '1');
      expect(result).toBe(false);
    });

    it('スコアが空の場合はfalse', () => {
      const errors = {};
      expect(isFormValid(errors, '', '1')).toBe(false);
      expect(isFormValid(errors, '3', '')).toBe(false);
      expect(isFormValid(errors, '', '')).toBe(false);
    });
  });

  describe('タッチ状態管理ロジック', () => {
    // タッチ状態管理のテスト
    function handleFieldTouch(touched, field) {
      return { ...touched, [field]: true };
    }

    it('フィールドタッチでタッチ状態が更新される', () => {
      const initialTouched = { score1: false, score2: false };
      
      const result1 = handleFieldTouch(initialTouched, 'score1');
      expect(result1).toEqual({ score1: true, score2: false });

      const result2 = handleFieldTouch(result1, 'score2');
      expect(result2).toEqual({ score1: true, score2: true });
    });
  });

  describe('結果データ構築ロジック', () => {
    // 結果データ構築関数をテスト用に抽出
    function buildResultData(score1, score2, winner) {
      return {
        score1: Number(score1),
        score2: Number(score2),
        winner: winner
      };
    }

    it('正しい結果データが構築される', () => {
      const result = buildResultData('3', '1', 'チームA');
      expect(result).toEqual({
        score1: 3,
        score2: 1,
        winner: 'チームA'
      });
    });

    it('数値変換が正しく行われる', () => {
      const result = buildResultData('0', '5', '引き分け');
      expect(result).toEqual({
        score1: 0,
        score2: 5,
        winner: '引き分け'
      });
    });
  });

  describe('フォームリセットロジック', () => {
    // フォームリセット関数をテスト用に抽出
    function resetForm(match) {
      return {
        score1: match.score1 || '',
        score2: match.score2 || '',
        errors: {},
        touched: { score1: false, score2: false }
      };
    }

    it('既存のスコアがある場合はそれを使用', () => {
      const match = { score1: 2, score2: 1 };
      const result = resetForm(match);
      
      expect(result.score1).toBe(2);
      expect(result.score2).toBe(1);
      expect(result.errors).toEqual({});
      expect(result.touched).toEqual({ score1: false, score2: false });
    });

    it('既存のスコアがない場合は空文字', () => {
      const match = {};
      const result = resetForm(match);
      
      expect(result.score1).toBe('');
      expect(result.score2).toBe('');
      expect(result.errors).toEqual({});
      expect(result.touched).toEqual({ score1: false, score2: false });
    });
  });
});
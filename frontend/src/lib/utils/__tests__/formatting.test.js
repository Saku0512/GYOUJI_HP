// フォーマットユーティリティの単体テスト
import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  formatDate,
  formatSportName,
  formatMatchStatus,
  formatScore,
  formatTournamentFormat,
  formatTeamName
} from '../formatting.js';

describe('formatting utilities', () => {
  beforeEach(() => {
    // 日付のテストのためにタイムゾーンを固定
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2024-01-15T10:30:00Z'));
  });

  describe('formatDate', () => {
    it('有効な日付文字列を正しくフォーマットする', () => {
      const dateString = '2024-01-15T10:30:00Z';
      const result = formatDate(dateString);
      
      // 日本語ロケールでの日付フォーマットをテスト
      expect(result).toMatch(/2024年1月15日/);
    });

    it('null値に対して空文字を返す', () => {
      expect(formatDate(null)).toBe('');
    });

    it('undefined値に対して空文字を返す', () => {
      expect(formatDate(undefined)).toBe('');
    });

    it('空文字に対して空文字を返す', () => {
      expect(formatDate('')).toBe('');
    });

    it('無効な日付文字列に対してInvalid Dateを含む文字列を返す', () => {
      const result = formatDate('invalid-date');
      expect(result).toMatch(/Invalid Date/);
    });
  });

  describe('formatSportName', () => {
    it('volleyballを正しく変換する', () => {
      expect(formatSportName('volleyball')).toBe('バレーボール');
    });

    it('table_tennisを正しく変換する', () => {
      expect(formatSportName('table_tennis')).toBe('卓球');
    });

    it('soccerを正しく変換する', () => {
      expect(formatSportName('soccer')).toBe('サッカー');
    });

    it('未知のスポーツ名はそのまま返す', () => {
      expect(formatSportName('basketball')).toBe('basketball');
    });

    it('null値はそのまま返す', () => {
      expect(formatSportName(null)).toBe(null);
    });

    it('undefined値はそのまま返す', () => {
      expect(formatSportName(undefined)).toBe(undefined);
    });
  });

  describe('formatMatchStatus', () => {
    it('pendingを正しく変換する', () => {
      expect(formatMatchStatus('pending')).toBe('未実施');
    });

    it('in_progressを正しく変換する', () => {
      expect(formatMatchStatus('in_progress')).toBe('進行中');
    });

    it('completedを正しく変換する', () => {
      expect(formatMatchStatus('completed')).toBe('完了');
    });

    it('cancelledを正しく変換する', () => {
      expect(formatMatchStatus('cancelled')).toBe('キャンセル');
    });

    it('未知のステータスはそのまま返す', () => {
      expect(formatMatchStatus('unknown')).toBe('unknown');
    });
  });

  describe('formatScore', () => {
    it('有効なスコアを正しくフォーマットする', () => {
      expect(formatScore(3, 1)).toBe('3 - 1');
      expect(formatScore(0, 0)).toBe('0 - 0');
      expect(formatScore(10, 5)).toBe('10 - 5');
    });

    it('score1がnullの場合は未実施を返す', () => {
      expect(formatScore(null, 1)).toBe('未実施');
    });

    it('score2がnullの場合は未実施を返す', () => {
      expect(formatScore(3, null)).toBe('未実施');
    });

    it('両方のスコアがnullの場合は未実施を返す', () => {
      expect(formatScore(null, null)).toBe('未実施');
    });

    it('score1がundefinedの場合は未実施を返す', () => {
      expect(formatScore(undefined, 1)).toBe('未実施');
    });

    it('score2がundefinedの場合は未実施を返す', () => {
      expect(formatScore(3, undefined)).toBe('未実施');
    });

    it('両方のスコアがundefinedの場合は未実施を返す', () => {
      expect(formatScore(undefined, undefined)).toBe('未実施');
    });
  });

  describe('formatTournamentFormat', () => {
    it('sunnyを正しく変換する', () => {
      expect(formatTournamentFormat('sunny')).toBe('晴天時形式');
    });

    it('rainyを正しく変換する', () => {
      expect(formatTournamentFormat('rainy')).toBe('雨天時形式');
    });

    it('single_eliminationを正しく変換する', () => {
      expect(formatTournamentFormat('single_elimination')).toBe('シングルエリミネーション');
    });

    it('double_eliminationを正しく変換する', () => {
      expect(formatTournamentFormat('double_elimination')).toBe('ダブルエリミネーション');
    });

    it('未知のフォーマットはそのまま返す', () => {
      expect(formatTournamentFormat('round_robin')).toBe('round_robin');
    });
  });

  describe('formatTeamName', () => {
    it('短いチーム名はそのまま返す', () => {
      expect(formatTeamName('チームA')).toBe('チームA');
      expect(formatTeamName('Team')).toBe('Team');
    });

    it('長いチーム名を正しく短縮する（デフォルト10文字）', () => {
      const longName = 'とても長いチーム名です';
      const result = formatTeamName(longName);
      expect(result).toBe('とても長いチー...');
      expect(result.length).toBe(10);
    });

    it('カスタム最大長で正しく短縮する', () => {
      const longName = 'Long Team Name';
      const result = formatTeamName(longName, 8);
      expect(result).toBe('Long ...');
      expect(result.length).toBe(8);
    });

    it('最大長と同じ長さの場合は短縮しない', () => {
      const name = '1234567890';
      expect(formatTeamName(name, 10)).toBe('1234567890');
    });

    it('null値に対して空文字を返す', () => {
      expect(formatTeamName(null)).toBe('');
    });

    it('undefined値に対して空文字を返す', () => {
      expect(formatTeamName(undefined)).toBe('');
    });

    it('空文字に対して空文字を返す', () => {
      expect(formatTeamName('')).toBe('');
    });

    it('最大長が3以下の場合でも正しく動作する', () => {
      expect(formatTeamName('Long Name', 3)).toBe('...');
    });
  });

  describe('エッジケース', () => {
    it('formatDateで異なるタイムゾーンの日付を処理する', () => {
      const utcDate = '2024-01-15T00:00:00Z';
      const jstDate = '2024-01-15T09:00:00+09:00';
      
      const utcResult = formatDate(utcDate);
      const jstResult = formatDate(jstDate);
      
      // 両方とも有効な日付文字列を返すことを確認
      expect(utcResult).toMatch(/2024年/);
      expect(jstResult).toMatch(/2024年/);
    });

    it('formatScoreで文字列の数値を処理する', () => {
      expect(formatScore('3', '1')).toBe('3 - 1');
      expect(formatScore('0', '0')).toBe('0 - 0');
    });

    it('formatTeamNameで特殊文字を含む名前を処理する', () => {
      const specialName = 'チーム★☆♪♫♪♫♪♫';
      const result = formatTeamName(specialName, 8);
      expect(result).toBe('チーム★☆♪...');
    });
  });
});
// 検証ユーティリティのテスト
import { describe, it, expect, vi, beforeEach } from 'vitest';
import {
  sanitizeInput,
  escapeHtml,
  validateSqlSafety,
  validateXssSafety,
  validateLength,
  validateRange,
  validatePattern,
  validateLoginCredentials,
  validateScore,
  validateMatchResult,
  validateTeamName,
  validateRequired,
  validateForm,
  debounce,
  validateCsrfToken
} from '../validation.js';

describe('検証ユーティリティ', () => {
  describe('sanitizeInput', () => {
    it('HTMLタグを除去する', () => {
      const input = '<script>alert("xss")</script>Hello<b>World</b>';
      const result = sanitizeInput(input);
      expect(result).toBe('alert(&quot;xss&quot;)HelloWorld');
    });

    it('JavaScriptイベントハンドラーを除去する', () => {
      const input = '<div onclick="alert(1)">Click me</div>';
      const result = sanitizeInput(input);
      expect(result).toBe('Click me');
    });

    it('javascript:プロトコルを除去する', () => {
      const input = 'javascript:alert("xss")';
      const result = sanitizeInput(input);
      expect(result).toBe('alert(&quot;xss&quot;)');
    });

    it('危険な文字をエスケープする', () => {
      const input = '<>&"\'';
      const result = sanitizeInput(input);
      expect(result).toBe('&amp;&quot;&#x27;');
    });

    it('文字列以外の値はそのまま返す', () => {
      expect(sanitizeInput(123)).toBe(123);
      expect(sanitizeInput(null)).toBe(null);
      expect(sanitizeInput(undefined)).toBe(undefined);
    });
  });

  describe('escapeHtml', () => {
    it('HTMLエンティティを正しくエスケープする', () => {
      const input = '<script>alert("test")</script>';
      const result = escapeHtml(input);
      expect(result).toBe('&lt;script&gt;alert(&quot;test&quot;)&lt;&#x2F;script&gt;');
    });

    it('文字列以外の値はそのまま返す', () => {
      expect(escapeHtml(123)).toBe(123);
      expect(escapeHtml(null)).toBe(null);
    });
  });

  describe('validateSqlSafety', () => {
    it('危険なSQLキーワードを検出する', () => {
      const dangerousInputs = [
        'SELECT * FROM users',
        'DROP TABLE users',
        'INSERT INTO users',
        'UPDATE users SET',
        'DELETE FROM users',
        "'; DROP TABLE users; --"
      ];

      dangerousInputs.forEach(input => {
        const result = validateSqlSafety(input);
        expect(result.isValid).toBe(false);
        expect(result.error).toBe('不正な文字が含まれています');
      });
    });

    it('安全な文字列は通す', () => {
      const safeInputs = [
        'Hello World',
        'ユーザー名',
        '123456',
        'test@example.com'
      ];

      safeInputs.forEach(input => {
        const result = validateSqlSafety(input);
        expect(result.isValid).toBe(true);
        expect(result.error).toBe(null);
      });
    });

    it('文字列以外の値は有効とする', () => {
      const result = validateSqlSafety(123);
      expect(result.isValid).toBe(true);
    });
  });

  describe('validateXssSafety', () => {
    it('危険なXSSパターンを検出する', () => {
      const dangerousInputs = [
        '<script>alert("xss")</script>',
        '<iframe src="javascript:alert(1)"></iframe>',
        '<div onclick="alert(1)">Click</div>',
        'javascript:alert("xss")',
        'vbscript:msgbox("xss")',
        'data:text/html,<script>alert(1)</script>'
      ];

      dangerousInputs.forEach(input => {
        const result = validateXssSafety(input);
        expect(result.isValid).toBe(false);
        expect(result.error).toBe('不正なスクリプトが含まれています');
      });
    });

    it('安全な文字列は通す', () => {
      const safeInputs = [
        'Hello World',
        '<p>Safe HTML</p>',
        'https://example.com',
        'user@example.com'
      ];

      safeInputs.forEach(input => {
        const result = validateXssSafety(input);
        expect(result.isValid).toBe(true);
        expect(result.error).toBe(null);
      });
    });
  });

  describe('validateLength', () => {
    it('最小長度をチェックする', () => {
      const result = validateLength('ab', 3, 10, 'テスト');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('テストは3文字以上である必要があります');
    });

    it('最大長度をチェックする', () => {
      const result = validateLength('abcdefghijk', 1, 10, 'テスト');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('テストは10文字以下である必要があります');
    });

    it('有効な長さの場合は成功する', () => {
      const result = validateLength('abcde', 3, 10, 'テスト');
      expect(result.isValid).toBe(true);
      expect(result.error).toBe(null);
    });

    it('文字列以外の値はエラーとする', () => {
      const result = validateLength(123, 1, 10, 'テスト');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('テストは文字列である必要があります');
    });
  });

  describe('validateRange', () => {
    it('最小値をチェックする', () => {
      const result = validateRange(5, 10, 20, 'スコア');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('スコアは10以上である必要があります');
    });

    it('最大値をチェックする', () => {
      const result = validateRange(25, 10, 20, 'スコア');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('スコアは20以下である必要があります');
    });

    it('有効な範囲の場合は成功する', () => {
      const result = validateRange(15, 10, 20, 'スコア');
      expect(result.isValid).toBe(true);
      expect(result.error).toBe(null);
    });

    it('数値以外の値はエラーとする', () => {
      const result = validateRange('abc', 10, 20, 'スコア');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('スコアは数値である必要があります');
    });
  });

  describe('validateLoginCredentials', () => {
    it('有効な認証情報の場合は成功する', () => {
      const result = validateLoginCredentials('admin', 'password123');
      expect(result.isValid).toBe(true);
      expect(result.errors).toEqual({});
      expect(result.sanitizedData.username).toBe('admin');
      expect(result.sanitizedData.password).toBe('password123');
    });

    it('空のユーザー名はエラーとする', () => {
      const result = validateLoginCredentials('', 'password123');
      expect(result.isValid).toBe(false);
      expect(result.errors.username).toBe('ユーザー名は必須です');
    });

    it('空のパスワードはエラーとする', () => {
      const result = validateLoginCredentials('admin', '');
      expect(result.isValid).toBe(false);
      expect(result.errors.password).toBe('パスワードは必須です');
    });

    it('危険な文字を含むユーザー名はエラーとする', () => {
      const result = validateLoginCredentials('<script>alert(1)</script>', 'password');
      expect(result.isValid).toBe(false);
      expect(result.errors.username).toBe('不正な文字が含まれています');
    });

    it('長すぎるユーザー名はエラーとする', () => {
      const longUsername = 'a'.repeat(51);
      const result = validateLoginCredentials(longUsername, 'password');
      expect(result.isValid).toBe(false);
      expect(result.errors.username).toBe('ユーザー名は50文字以下である必要があります');
    });
  });

  describe('validateScore', () => {
    it('有効なスコアの場合は成功する', () => {
      const result = validateScore('10');
      expect(result.isValid).toBe(true);
      expect(result.error).toBe(null);
      expect(result.sanitizedValue).toBe(10);
    });

    it('空のスコアはエラーとする', () => {
      const result = validateScore('');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('スコアは必須です');
    });

    it('負の数はエラーとする', () => {
      const result = validateScore('-5');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('スコアは0以上である必要があります');
    });

    it('大きすぎる数はエラーとする', () => {
      const result = validateScore('1000');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('スコアは999以下である必要があります');
    });

    it('危険な文字を含む場合はエラーとする', () => {
      const result = validateScore('<script>alert(1)</script>');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('不正な文字が含まれています');
    });
  });

  describe('validateMatchResult', () => {
    it('有効な試合結果の場合は成功する', () => {
      const result = validateMatchResult('3', '1', 'Team A', 'Team B');
      expect(result.isValid).toBe(true);
      expect(result.errors).toEqual({});
      expect(result.sanitizedData.score1).toBe(3);
      expect(result.sanitizedData.score2).toBe(1);
    });

    it('同点の場合はエラーとする', () => {
      const result = validateMatchResult('2', '2');
      expect(result.isValid).toBe(false);
      expect(result.errors.general).toBe('引き分けは許可されていません');
    });

    it('無効なスコアの場合はエラーとする', () => {
      const result = validateMatchResult('abc', '1');
      expect(result.isValid).toBe(false);
      expect(result.errors.score1).toBeDefined();
    });

    it('チーム名の検証も行う', () => {
      const result = validateMatchResult('3', '1', '<script>alert(1)</script>', 'Team B');
      expect(result.isValid).toBe(false);
      expect(result.errors.team1).toBe('不正な文字が含まれています');
    });
  });

  describe('validateTeamName', () => {
    it('有効なチーム名の場合は成功する', () => {
      const result = validateTeamName('Team Alpha');
      expect(result.isValid).toBe(true);
      expect(result.error).toBe(null);
      expect(result.sanitizedValue).toBe('Team Alpha');
    });

    it('空のチーム名はエラーとする', () => {
      const result = validateTeamName('');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('チーム名は必須です');
    });

    it('長すぎるチーム名はエラーとする', () => {
      const longName = 'a'.repeat(51);
      const result = validateTeamName(longName);
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('チーム名は50文字以下である必要があります');
    });

    it('危険な文字を含む場合はエラーとする', () => {
      const result = validateTeamName('<script>alert(1)</script>');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('不正な文字が含まれています');
    });
  });

  describe('validateForm', () => {
    it('フォーム全体の検証を行う', () => {
      const formData = {
        username: 'admin',
        password: 'password123',
        score: '10'
      };

      const validationRules = {
        username: [
          { validator: validateRequired, params: 'ユーザー名' }
        ],
        password: [
          { validator: validateRequired, params: 'パスワード' }
        ],
        score: [
          { validator: validateScore, params: 'スコア' }
        ]
      };

      const result = validateForm(formData, validationRules);
      expect(result.isValid).toBe(true);
      expect(result.errors).toEqual({});
    });

    it('複数のエラーを検出する', () => {
      const formData = {
        username: '',
        password: '',
        score: 'invalid'
      };

      const validationRules = {
        username: [
          { validator: validateRequired, params: 'ユーザー名' }
        ],
        password: [
          { validator: validateRequired, params: 'パスワード' }
        ],
        score: [
          { validator: validateScore, params: 'スコア' }
        ]
      };

      const result = validateForm(formData, validationRules);
      expect(result.isValid).toBe(false);
      expect(Object.keys(result.errors)).toHaveLength(3);
    });
  });

  describe('debounce', () => {
    it('指定した時間後に関数を実行する', async () => {
      const mockFn = vi.fn();
      const debouncedFn = debounce(mockFn, 100);

      debouncedFn('test');
      expect(mockFn).not.toHaveBeenCalled();

      await new Promise(resolve => setTimeout(resolve, 150));
      expect(mockFn).toHaveBeenCalledWith('test');
    });

    it('連続呼び出しでは最後の呼び出しのみ実行する', async () => {
      const mockFn = vi.fn();
      const debouncedFn = debounce(mockFn, 100);

      debouncedFn('first');
      debouncedFn('second');
      debouncedFn('third');

      await new Promise(resolve => setTimeout(resolve, 150));
      expect(mockFn).toHaveBeenCalledTimes(1);
      expect(mockFn).toHaveBeenCalledWith('third');
    });
  });

  describe('validateCsrfToken', () => {
    it('有効なCSRFトークンの場合は成功する', () => {
      const validToken = 'abcdef1234567890abcdef1234567890';
      const result = validateCsrfToken(validToken);
      expect(result.isValid).toBe(true);
      expect(result.error).toBe(null);
    });

    it('空のトークンはエラーとする', () => {
      const result = validateCsrfToken('');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('CSRFトークンが無効です');
    });

    it('短すぎるトークンはエラーとする', () => {
      const result = validateCsrfToken('short');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('CSRFトークンの形式が正しくありません');
    });

    it('無効な文字を含むトークンはエラーとする', () => {
      const result = validateCsrfToken('invalid-token-with-special-chars!@#');
      expect(result.isValid).toBe(false);
      expect(result.error).toBe('CSRFトークンの形式が正しくありません');
    });
  });
});
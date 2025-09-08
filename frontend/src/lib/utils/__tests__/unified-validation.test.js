// 統一バリデーションシステムのテスト

import { describe, it, expect, beforeEach } from 'vitest';
import {
  ValidationError,
  ValidationResult,
  RequiredRule,
  MinLengthRule,
  MaxLengthRule,
  MinValueRule,
  MaxValueRule,
  AlphanumericRule,
  EnumRule,
  PatternRule,
  SportTypeRule,
  TournamentStatusRule,
  MatchStatusRule,
  PasswordRule,
  FieldValidator,
  FormValidator,
  createLoginValidator,
  createMatchResultValidator,
  createTournamentValidator
} from '../unified-validation.js';
import { ErrorCode } from '../../api/types.js';

describe('ValidationError', () => {
  it('should create validation error with all properties', () => {
    const error = new ValidationError(
      'username',
      'ユーザー名は必須です',
      '',
      ErrorCode.VALIDATION_REQUIRED_FIELD,
      'required'
    );

    expect(error.field).toBe('username');
    expect(error.message).toBe('ユーザー名は必須です');
    expect(error.value).toBe('');
    expect(error.code).toBe(ErrorCode.VALIDATION_REQUIRED_FIELD);
    expect(error.rule).toBe('required');
  });
});

describe('ValidationResult', () => {
  let result;

  beforeEach(() => {
    result = new ValidationResult();
  });

  it('should initialize with valid state', () => {
    expect(result.isValid).toBe(true);
    expect(result.errors).toEqual({});
    expect(result.sanitizedData).toEqual({});
  });

  it('should detect errors correctly', () => {
    expect(result.hasErrors()).toBe(false);
    
    result.addError('username', 'ユーザー名は必須です');
    expect(result.hasErrors()).toBe(true);
    expect(result.isValid).toBe(false);
  });

  it('should get field error', () => {
    result.addError('username', 'ユーザー名は必須です');
    expect(result.getFieldError('username')).toBe('ユーザー名は必須です');
    expect(result.getFieldError('password')).toBeNull();
  });

  it('should set sanitized value', () => {
    result.setSanitizedValue('username', 'admin');
    expect(result.sanitizedData.username).toBe('admin');
  });
});

describe('RequiredRule', () => {
  let rule;

  beforeEach(() => {
    rule = new RequiredRule();
  });

  it('should pass for valid values', () => {
    expect(rule.validate('admin', 'ユーザー名')).toBeNull();
    expect(rule.validate('password123', 'パスワード')).toBeNull();
    expect(rule.validate(123, '数値')).toBeNull();
  });

  it('should fail for empty values', () => {
    const error = rule.validate('', 'ユーザー名');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('ユーザー名は必須です');
    expect(error.code).toBe(ErrorCode.VALIDATION_REQUIRED_FIELD);
  });

  it('should fail for null/undefined values', () => {
    expect(rule.validate(null, 'ユーザー名')).toBeInstanceOf(ValidationError);
    expect(rule.validate(undefined, 'ユーザー名')).toBeInstanceOf(ValidationError);
  });

  it('should fail for whitespace-only strings', () => {
    const error = rule.validate('   ', 'ユーザー名');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('ユーザー名は必須です');
  });
});

describe('MinLengthRule', () => {
  let rule;

  beforeEach(() => {
    rule = new MinLengthRule(3);
  });

  it('should pass for valid length', () => {
    expect(rule.validate('abc', 'テキスト')).toBeNull();
    expect(rule.validate('abcd', 'テキスト')).toBeNull();
  });

  it('should fail for short strings', () => {
    const error = rule.validate('ab', 'テキスト');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('テキストは3文字以上である必要があります');
    expect(error.code).toBe(ErrorCode.VALIDATION_OUT_OF_RANGE);
  });

  it('should ignore non-string values', () => {
    expect(rule.validate(123, 'テキスト')).toBeNull();
    expect(rule.validate(null, 'テキスト')).toBeNull();
  });
});

describe('MaxLengthRule', () => {
  let rule;

  beforeEach(() => {
    rule = new MaxLengthRule(5);
  });

  it('should pass for valid length', () => {
    expect(rule.validate('abc', 'テキスト')).toBeNull();
    expect(rule.validate('abcde', 'テキスト')).toBeNull();
  });

  it('should fail for long strings', () => {
    const error = rule.validate('abcdef', 'テキスト');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('テキストは5文字以下である必要があります');
    expect(error.code).toBe(ErrorCode.VALIDATION_OUT_OF_RANGE);
  });
});

describe('MinValueRule', () => {
  let rule;

  beforeEach(() => {
    rule = new MinValueRule(0);
  });

  it('should pass for valid values', () => {
    expect(rule.validate(0, 'スコア')).toBeNull();
    expect(rule.validate(10, 'スコア')).toBeNull();
    expect(rule.validate('5', 'スコア')).toBeNull();
  });

  it('should fail for small values', () => {
    const error = rule.validate(-1, 'スコア');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('スコアは0以上である必要があります');
    expect(error.code).toBe(ErrorCode.VALIDATION_OUT_OF_RANGE);
  });

  it('should handle string numbers', () => {
    const error = rule.validate('-1', 'スコア');
    expect(error).toBeInstanceOf(ValidationError);
  });
});

describe('MaxValueRule', () => {
  let rule;

  beforeEach(() => {
    rule = new MaxValueRule(100);
  });

  it('should pass for valid values', () => {
    expect(rule.validate(50, 'スコア')).toBeNull();
    expect(rule.validate(100, 'スコア')).toBeNull();
    expect(rule.validate('99', 'スコア')).toBeNull();
  });

  it('should fail for large values', () => {
    const error = rule.validate(101, 'スコア');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('スコアは100以下である必要があります');
    expect(error.code).toBe(ErrorCode.VALIDATION_OUT_OF_RANGE);
  });
});

describe('AlphanumericRule', () => {
  let rule;

  beforeEach(() => {
    rule = new AlphanumericRule();
  });

  it('should pass for alphanumeric strings', () => {
    expect(rule.validate('admin123', 'ユーザー名')).toBeNull();
    expect(rule.validate('ABC', 'ユーザー名')).toBeNull();
    expect(rule.validate('123', 'ユーザー名')).toBeNull();
  });

  it('should fail for non-alphanumeric strings', () => {
    const error = rule.validate('admin@123', 'ユーザー名');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('ユーザー名は英数字のみ使用可能です');
    expect(error.code).toBe(ErrorCode.VALIDATION_INVALID_FORMAT);
  });

  it('should pass for empty strings', () => {
    expect(rule.validate('', 'ユーザー名')).toBeNull();
  });
});

describe('EnumRule', () => {
  let rule;

  beforeEach(() => {
    rule = new EnumRule(['volleyball', 'table_tennis', 'soccer']);
  });

  it('should pass for valid values', () => {
    expect(rule.validate('volleyball', 'スポーツ')).toBeNull();
    expect(rule.validate('table_tennis', 'スポーツ')).toBeNull();
    expect(rule.validate('soccer', 'スポーツ')).toBeNull();
  });

  it('should fail for invalid values', () => {
    const error = rule.validate('basketball', 'スポーツ');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toContain('スポーツは無効な値です');
    expect(error.message).toContain('volleyball, table_tennis, soccer');
    expect(error.code).toBe(ErrorCode.VALIDATION_INVALID_FORMAT);
  });

  it('should pass for empty strings', () => {
    expect(rule.validate('', 'スポーツ')).toBeNull();
  });
});

describe('PatternRule', () => {
  let rule;

  beforeEach(() => {
    rule = new PatternRule(/^[A-Z][a-z]+$/, 'カスタムメッセージ');
  });

  it('should pass for matching patterns', () => {
    expect(rule.validate('Admin', 'テキスト')).toBeNull();
    expect(rule.validate('Test', 'テキスト')).toBeNull();
  });

  it('should fail for non-matching patterns', () => {
    const error = rule.validate('admin', 'テキスト');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('カスタムメッセージ');
    expect(error.code).toBe(ErrorCode.VALIDATION_INVALID_FORMAT);
  });

  it('should use default message when not provided', () => {
    const defaultRule = new PatternRule(/^[A-Z]+$/);
    const error = defaultRule.validate('abc', 'テキスト');
    expect(error.message).toBe('テキストの形式が正しくありません');
  });
});

describe('SportTypeRule', () => {
  let rule;

  beforeEach(() => {
    rule = new SportTypeRule();
  });

  it('should pass for valid sports', () => {
    expect(rule.validate('volleyball', 'スポーツ')).toBeNull();
    expect(rule.validate('table_tennis', 'スポーツ')).toBeNull();
    expect(rule.validate('soccer', 'スポーツ')).toBeNull();
  });

  it('should fail for invalid sports', () => {
    const error = rule.validate('basketball', 'スポーツ');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toContain('スポーツは無効なスポーツです');
  });

  it('should pass for empty strings', () => {
    expect(rule.validate('', 'スポーツ')).toBeNull();
  });
});

describe('PasswordRule', () => {
  let rule;

  beforeEach(() => {
    rule = new PasswordRule({
      minLength: 8,
      maxLength: 20,
      requireAlpha: true,
      requireNumeric: true
    });
  });

  it('should pass for valid passwords', () => {
    expect(rule.validate('password123', 'パスワード')).toBeNull();
    expect(rule.validate('MyPassword1', 'パスワード')).toBeNull();
  });

  it('should fail for short passwords', () => {
    const error = rule.validate('pass1', 'パスワード');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('パスワードは8文字以上である必要があります');
  });

  it('should fail for long passwords', () => {
    const error = rule.validate('a'.repeat(21) + '1', 'パスワード');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('パスワードは20文字以下である必要があります');
  });

  it('should fail when missing required alpha', () => {
    const error = rule.validate('12345678', 'パスワード');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('パスワードは英字を含む必要があります');
  });

  it('should fail when missing required numeric', () => {
    const error = rule.validate('password', 'パスワード');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.message).toBe('パスワードは数字を含む必要があります');
  });
});

describe('FieldValidator', () => {
  let validator;

  beforeEach(() => {
    validator = new FieldValidator('ユーザー名', [
      new RequiredRule(),
      new MinLengthRule(3),
      new MaxLengthRule(20),
      new AlphanumericRule()
    ]);
  });

  it('should pass for valid values', () => {
    expect(validator.validate('admin')).toBeNull();
    expect(validator.validate('user123')).toBeNull();
  });

  it('should fail on first rule violation', () => {
    const error = validator.validate('');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.rule).toBe('required');
  });

  it('should fail on subsequent rule violations', () => {
    const error = validator.validate('ab');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.rule).toBe('min_length');
  });

  it('should add rules dynamically', () => {
    validator.addRule(new PatternRule(/^admin/));
    const error = validator.validate('user');
    expect(error).toBeInstanceOf(ValidationError);
    expect(error.rule).toBe('pattern');
  });
});

describe('FormValidator', () => {
  let validator;

  beforeEach(() => {
    validator = new FormValidator();
    validator.addField('username', [
      new RequiredRule(),
      new MinLengthRule(3),
      new AlphanumericRule()
    ]);
    validator.addField('password', [
      new RequiredRule(),
      new MinLengthRule(8)
    ]);
  });

  it('should validate individual fields', () => {
    const result = validator.validateField('username', 'admin');
    expect(result.error).toBeNull();
    expect(result.sanitizedValue).toBe('admin');
  });

  it('should return null for unknown fields', () => {
    const result = validator.validateField('unknown', 'value');
    expect(result).toBeNull();
  });

  it('should validate entire form', () => {
    const formData = {
      username: 'admin',
      password: 'password123'
    };

    const result = validator.validateForm(formData);
    expect(result.isValid).toBe(true);
    expect(result.hasErrors()).toBe(false);
  });

  it('should collect all field errors', () => {
    const formData = {
      username: 'ab',
      password: 'short'
    };

    const result = validator.validateForm(formData);
    expect(result.isValid).toBe(false);
    expect(result.hasErrors()).toBe(true);
    expect(result.errors.username).toContain('3文字以上');
    expect(result.errors.password).toContain('8文字以上');
  });

  it('should apply sanitizers', () => {
    validator.addSanitizer('username', (value) => value.trim().toLowerCase());

    const result = validator.validateField('username', '  ADMIN  ');
    expect(result.sanitizedValue).toBe('admin');
  });
});

describe('Predefined Validators', () => {
  describe('createLoginValidator', () => {
    let validator;

    beforeEach(() => {
      validator = createLoginValidator();
    });

    it('should validate valid login data', () => {
      const formData = {
        username: 'admin',
        password: 'password123'
      };

      const result = validator.validateForm(formData);
      expect(result.isValid).toBe(true);
    });

    it('should reject invalid login data', () => {
      const formData = {
        username: '',
        password: ''
      };

      const result = validator.validateForm(formData);
      expect(result.isValid).toBe(false);
      expect(result.errors.username).toContain('必須');
      expect(result.errors.password).toContain('必須');
    });

    it('should sanitize username', () => {
      const result = validator.validateField('username', '  admin  ');
      expect(result.sanitizedValue).toBe('admin');
    });
  });

  describe('createMatchResultValidator', () => {
    let validator;

    beforeEach(() => {
      validator = createMatchResultValidator();
    });

    it('should validate valid match result', () => {
      const formData = {
        score1: 3,
        score2: 1,
        winner: 'Team A'
      };

      const result = validator.validateForm(formData);
      expect(result.isValid).toBe(true);
    });

    it('should reject invalid scores', () => {
      const formData = {
        score1: -1,
        score2: 1000,
        winner: 'Team A'
      };

      const result = validator.validateForm(formData);
      expect(result.isValid).toBe(false);
      expect(result.errors.score1).toContain('0以上');
      expect(result.errors.score2).toContain('999以下');
    });
  });

  describe('createTournamentValidator', () => {
    let validator;

    beforeEach(() => {
      validator = createTournamentValidator();
    });

    it('should validate valid tournament data', () => {
      const formData = {
        sport: 'volleyball',
        format: 'single_elimination'
      };

      const result = validator.validateForm(formData);
      expect(result.isValid).toBe(true);
    });

    it('should reject invalid sport', () => {
      const formData = {
        sport: 'basketball',
        format: 'single_elimination'
      };

      const result = validator.validateForm(formData);
      expect(result.isValid).toBe(false);
      expect(result.errors.sport).toContain('無効なスポーツ');
    });
  });
});
// 統一バリデーションシステム
// バックエンドと一致するバリデーションルールの実装

import { ErrorCode } from '../api/types.js';

/**
 * バリデーションエラー構造体
 */
export class ValidationError {
  constructor(field, message, value, code, rule) {
    this.field = field;
    this.message = message;
    this.value = value;
    this.code = code;
    this.rule = rule;
  }
}

/**
 * バリデーション結果
 */
export class ValidationResult {
  constructor(isValid = true, errors = {}, sanitizedData = {}) {
    this.isValid = isValid;
    this.errors = errors;
    this.sanitizedData = sanitizedData;
  }

  hasErrors() {
    return Object.keys(this.errors).length > 0;
  }

  getFieldError(fieldName) {
    return this.errors[fieldName] || null;
  }

  addError(fieldName, error) {
    this.errors[fieldName] = error;
    this.isValid = false;
  }

  setSanitizedValue(fieldName, value) {
    this.sanitizedData[fieldName] = value;
  }
}

/**
 * 基底バリデーションルールクラス
 */
export class ValidationRule {
  constructor(ruleName) {
    this.ruleName = ruleName;
  }

  validate(value, fieldName) {
    throw new Error('validate method must be implemented');
  }

  getRuleName() {
    return this.ruleName;
  }
}

/**
 * 必須フィールドルール
 */
export class RequiredRule extends ValidationRule {
  constructor() {
    super('required');
  }

  validate(value, fieldName) {
    if (value === null || value === undefined) {
      return new ValidationError(
        fieldName,
        `${fieldName}は必須です`,
        value,
        ErrorCode.VALIDATION_REQUIRED_FIELD,
        this.ruleName
      );
    }

    if (typeof value === 'string' && value.trim() === '') {
      return new ValidationError(
        fieldName,
        `${fieldName}は必須です`,
        value,
        ErrorCode.VALIDATION_REQUIRED_FIELD,
        this.ruleName
      );
    }

    return null;
  }
}

/**
 * 最小文字数ルール
 */
export class MinLengthRule extends ValidationRule {
  constructor(minLength) {
    super('min_length');
    this.minLength = minLength;
  }

  validate(value, fieldName) {
    if (typeof value === 'string' && value.length < this.minLength) {
      return new ValidationError(
        fieldName,
        `${fieldName}は${this.minLength}文字以上である必要があります`,
        value,
        ErrorCode.VALIDATION_OUT_OF_RANGE,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * 最大文字数ルール
 */
export class MaxLengthRule extends ValidationRule {
  constructor(maxLength) {
    super('max_length');
    this.maxLength = maxLength;
  }

  validate(value, fieldName) {
    if (typeof value === 'string' && value.length > this.maxLength) {
      return new ValidationError(
        fieldName,
        `${fieldName}は${this.maxLength}文字以下である必要があります`,
        value,
        ErrorCode.VALIDATION_OUT_OF_RANGE,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * 最小値ルール
 */
export class MinValueRule extends ValidationRule {
  constructor(minValue) {
    super('min_value');
    this.minValue = minValue;
  }

  validate(value, fieldName) {
    const numValue = Number(value);
    if (!isNaN(numValue) && numValue < this.minValue) {
      return new ValidationError(
        fieldName,
        `${fieldName}は${this.minValue}以上である必要があります`,
        value,
        ErrorCode.VALIDATION_OUT_OF_RANGE,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * 最大値ルール
 */
export class MaxValueRule extends ValidationRule {
  constructor(maxValue) {
    super('max_value');
    this.maxValue = maxValue;
  }

  validate(value, fieldName) {
    const numValue = Number(value);
    if (!isNaN(numValue) && numValue > this.maxValue) {
      return new ValidationError(
        fieldName,
        `${fieldName}は${this.maxValue}以下である必要があります`,
        value,
        ErrorCode.VALIDATION_OUT_OF_RANGE,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * 英数字ルール
 */
export class AlphanumericRule extends ValidationRule {
  constructor() {
    super('alphanumeric');
    this.pattern = /^[a-zA-Z0-9]+$/;
  }

  validate(value, fieldName) {
    if (typeof value === 'string' && value !== '' && !this.pattern.test(value)) {
      return new ValidationError(
        fieldName,
        `${fieldName}は英数字のみ使用可能です`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * 列挙型ルール
 */
export class EnumRule extends ValidationRule {
  constructor(validValues) {
    super('enum');
    this.validValues = validValues;
  }

  validate(value, fieldName) {
    if (value !== '' && !this.validValues.includes(value)) {
      return new ValidationError(
        fieldName,
        `${fieldName}は無効な値です。有効な値: ${this.validValues.join(', ')}`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * パターンルール
 */
export class PatternRule extends ValidationRule {
  constructor(pattern, message = null) {
    super('pattern');
    this.pattern = pattern instanceof RegExp ? pattern : new RegExp(pattern);
    this.customMessage = message;
  }

  validate(value, fieldName) {
    if (typeof value === 'string' && value !== '' && !this.pattern.test(value)) {
      return new ValidationError(
        fieldName,
        this.customMessage || `${fieldName}の形式が正しくありません`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * スポーツタイプルール
 */
export class SportTypeRule extends ValidationRule {
  constructor() {
    super('sport_type');
    this.validSports = ['volleyball', 'table_tennis', 'soccer'];
  }

  validate(value, fieldName) {
    if (value !== '' && !this.validSports.includes(value)) {
      return new ValidationError(
        fieldName,
        `${fieldName}は無効なスポーツです。有効な値: ${this.validSports.join(', ')}`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * トーナメントステータスルール
 */
export class TournamentStatusRule extends ValidationRule {
  constructor() {
    super('tournament_status');
    this.validStatuses = ['registration', 'active', 'completed', 'cancelled'];
  }

  validate(value, fieldName) {
    if (value !== '' && !this.validStatuses.includes(value)) {
      return new ValidationError(
        fieldName,
        `${fieldName}は無効なトーナメントステータスです。有効な値: ${this.validStatuses.join(', ')}`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * 試合ステータスルール
 */
export class MatchStatusRule extends ValidationRule {
  constructor() {
    super('match_status');
    this.validStatuses = ['pending', 'in_progress', 'completed', 'cancelled'];
  }

  validate(value, fieldName) {
    if (value !== '' && !this.validStatuses.includes(value)) {
      return new ValidationError(
        fieldName,
        `${fieldName}は無効な試合ステータスです。有効な値: ${this.validStatuses.join(', ')}`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        this.ruleName
      );
    }
    return null;
  }
}

/**
 * パスワードルール
 */
export class PasswordRule extends ValidationRule {
  constructor(options = {}) {
    super('password');
    this.minLength = options.minLength || 8;
    this.maxLength = options.maxLength || 100;
    this.requireAlpha = options.requireAlpha !== false;
    this.requireNumeric = options.requireNumeric !== false;
    this.requireSpecial = options.requireSpecial || false;
    this.requireMixedCase = options.requireMixedCase || false;
  }

  validate(value, fieldName) {
    if (typeof value !== 'string') {
      return null;
    }

    if (value.length < this.minLength) {
      return new ValidationError(
        fieldName,
        `${fieldName}は${this.minLength}文字以上である必要があります`,
        value,
        ErrorCode.VALIDATION_OUT_OF_RANGE,
        'password_min_length'
      );
    }

    if (value.length > this.maxLength) {
      return new ValidationError(
        fieldName,
        `${fieldName}は${this.maxLength}文字以下である必要があります`,
        value,
        ErrorCode.VALIDATION_OUT_OF_RANGE,
        'password_max_length'
      );
    }

    if (this.requireAlpha && !/[a-zA-Z]/.test(value)) {
      return new ValidationError(
        fieldName,
        `${fieldName}は英字を含む必要があります`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        'password_alpha'
      );
    }

    if (this.requireNumeric && !/[0-9]/.test(value)) {
      return new ValidationError(
        fieldName,
        `${fieldName}は数字を含む必要があります`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        'password_numeric'
      );
    }

    if (this.requireSpecial && !/[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]/.test(value)) {
      return new ValidationError(
        fieldName,
        `${fieldName}は特殊文字を含む必要があります`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        'password_special'
      );
    }

    if (this.requireMixedCase && (!/[A-Z]/.test(value) || !/[a-z]/.test(value))) {
      return new ValidationError(
        fieldName,
        `${fieldName}は大文字と小文字を含む必要があります`,
        value,
        ErrorCode.VALIDATION_INVALID_FORMAT,
        'password_mixed_case'
      );
    }

    return null;
  }
}

/**
 * バリデーター - 複数のルールを組み合わせてフィールドを検証
 */
export class FieldValidator {
  constructor(fieldName, rules = []) {
    this.fieldName = fieldName;
    this.rules = rules;
  }

  addRule(rule) {
    this.rules.push(rule);
    return this;
  }

  validate(value) {
    for (const rule of this.rules) {
      const error = rule.validate(value, this.fieldName);
      if (error) {
        return error;
      }
    }
    return null;
  }
}

/**
 * フォームバリデーター - フォーム全体の検証を管理
 */
export class FormValidator {
  constructor() {
    this.fieldValidators = new Map();
    this.sanitizers = new Map();
  }

  addField(fieldName, rules = []) {
    this.fieldValidators.set(fieldName, new FieldValidator(fieldName, rules));
    return this;
  }

  addSanitizer(fieldName, sanitizerFn) {
    this.sanitizers.set(fieldName, sanitizerFn);
    return this;
  }

  validateField(fieldName, value) {
    const validator = this.fieldValidators.get(fieldName);
    if (!validator) {
      return null;
    }

    // サニタイゼーション
    let sanitizedValue = value;
    const sanitizer = this.sanitizers.get(fieldName);
    if (sanitizer) {
      sanitizedValue = sanitizer(value);
    }

    // バリデーション
    const error = validator.validate(sanitizedValue);
    return {
      error,
      sanitizedValue
    };
  }

  validateForm(formData) {
    const result = new ValidationResult();

    for (const [fieldName, value] of Object.entries(formData)) {
      const fieldResult = this.validateField(fieldName, value);
      if (fieldResult) {
        if (fieldResult.error) {
          result.addError(fieldName, fieldResult.error.message);
        }
        result.setSanitizedValue(fieldName, fieldResult.sanitizedValue);
      }
    }

    return result;
  }
}

// ファクトリー関数

export function createRequiredRule() {
  return new RequiredRule();
}

export function createMinLengthRule(minLength) {
  return new MinLengthRule(minLength);
}

export function createMaxLengthRule(maxLength) {
  return new MaxLengthRule(maxLength);
}

export function createMinValueRule(minValue) {
  return new MinValueRule(minValue);
}

export function createMaxValueRule(maxValue) {
  return new MaxValueRule(maxValue);
}

export function createAlphanumericRule() {
  return new AlphanumericRule();
}

export function createEnumRule(validValues) {
  return new EnumRule(validValues);
}

export function createPatternRule(pattern, message) {
  return new PatternRule(pattern, message);
}

export function createSportTypeRule() {
  return new SportTypeRule();
}

export function createTournamentStatusRule() {
  return new TournamentStatusRule();
}

export function createMatchStatusRule() {
  return new MatchStatusRule();
}

export function createPasswordRule(options) {
  return new PasswordRule(options);
}

// 事前定義されたバリデーター

/**
 * ログイン認証情報バリデーター
 */
export function createLoginValidator() {
  const validator = new FormValidator();
  
  validator.addField('username', [
    createRequiredRule(),
    createMinLengthRule(1),
    createMaxLengthRule(50),
    createAlphanumericRule()
  ]);

  validator.addField('password', [
    createRequiredRule(),
    createMinLengthRule(1),
    createMaxLengthRule(100)
  ]);

  // サニタイゼーション
  validator.addSanitizer('username', (value) => {
    return typeof value === 'string' ? value.trim() : value;
  });

  return validator;
}

/**
 * 試合結果バリデーター
 */
export function createMatchResultValidator() {
  const validator = new FormValidator();
  
  validator.addField('score1', [
    createRequiredRule(),
    createMinValueRule(0),
    createMaxValueRule(999)
  ]);

  validator.addField('score2', [
    createRequiredRule(),
    createMinValueRule(0),
    createMaxValueRule(999)
  ]);

  validator.addField('winner', [
    createRequiredRule(),
    createMinLengthRule(1),
    createMaxLengthRule(100)
  ]);

  return validator;
}

/**
 * トーナメント作成バリデーター
 */
export function createTournamentValidator() {
  const validator = new FormValidator();
  
  validator.addField('sport', [
    createRequiredRule(),
    createSportTypeRule()
  ]);

  validator.addField('format', [
    createRequiredRule(),
    createMinLengthRule(1),
    createMaxLengthRule(50)
  ]);

  return validator;
}

/**
 * チーム名バリデーター
 */
export function createTeamNameValidator() {
  const validator = new FormValidator();
  
  validator.addField('teamName', [
    createRequiredRule(),
    createMinLengthRule(1),
    createMaxLengthRule(50)
  ]);

  // サニタイゼーション
  validator.addSanitizer('teamName', (value) => {
    return typeof value === 'string' ? value.trim() : value;
  });

  return validator;
}
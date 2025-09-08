// 統一バリデーションシステムの使用例

import {
  createLoginValidator,
  createMatchResultValidator,
  createTournamentValidator
} from './unified-validation.js';
import { createRealtimeValidator } from './realtime-validation.js';

/**
 * 基本的なフォームバリデーションの例
 */
export function basicValidationExample() {
  // 1. バリデーターを作成
  const validator = createLoginValidator();
  
  // 2. フォームデータを検証
  const formData = {
    username: 'admin',
    password: 'password123'
  };
  
  const result = validator.validateForm(formData);
  
  if (result.isValid) {
    console.log('バリデーション成功:', result.sanitizedData);
  } else {
    console.log('バリデーションエラー:', result.errors);
  }
  
  return result;
}

/**
 * リアルタイムバリデーションの例
 */
export function realtimeValidationExample() {
  // 1. フォームバリデーターを作成
  const formValidator = createLoginValidator();
  
  // 2. リアルタイムバリデーターを作成
  const realtimeValidator = createRealtimeValidator(formValidator, {
    debounceMs: 300,
    validateOnChange: true,
    validateOnBlur: true
  });
  
  // 3. フィールド値を設定（リアクティブに検証される）
  realtimeValidator.setFieldValue('username', 'admin');
  realtimeValidator.setFieldValue('password', 'password123');
  
  // 4. フィールドをタッチ済みにマーク
  realtimeValidator.touchField('username');
  realtimeValidator.touchField('password');
  
  // 5. フォーム全体を検証
  const result = realtimeValidator.validateAll();
  
  return {
    realtimeValidator,
    result
  };
}

/**
 * 試合結果バリデーションの例
 */
export function matchResultValidationExample() {
  const validator = createMatchResultValidator();
  
  const matchData = {
    score1: 3,
    score2: 1,
    winner: 'Team A'
  };
  
  const result = validator.validateForm(matchData);
  
  if (result.isValid) {
    console.log('試合結果が有効:', result.sanitizedData);
  } else {
    console.log('試合結果エラー:', result.errors);
  }
  
  return result;
}

/**
 * トーナメントバリデーションの例
 */
export function tournamentValidationExample() {
  const validator = createTournamentValidator();
  
  const tournamentData = {
    sport: 'volleyball',
    format: 'single_elimination'
  };
  
  const result = validator.validateForm(tournamentData);
  
  if (result.isValid) {
    console.log('トーナメントデータが有効:', result.sanitizedData);
  } else {
    console.log('トーナメントデータエラー:', result.errors);
  }
  
  return result;
}

/**
 * カスタムバリデーターの例
 */
export function customValidatorExample() {
  const { FormValidator, createRequiredRule, createMinLengthRule, createPatternRule } = 
    await import('./unified-validation.js');
  
  // カスタムバリデーターを作成
  const validator = new FormValidator();
  
  // カスタムフィールドを追加
  validator.addField('email', [
    createRequiredRule(),
    createPatternRule(/^[^\s@]+@[^\s@]+\.[^\s@]+$/, 'メールアドレスの形式が正しくありません')
  ]);
  
  validator.addField('phone', [
    createRequiredRule(),
    createPatternRule(/^\d{3}-\d{4}-\d{4}$/, '電話番号は000-0000-0000の形式で入力してください')
  ]);
  
  // サニタイザーを追加
  validator.addSanitizer('email', (value) => value.trim().toLowerCase());
  validator.addSanitizer('phone', (value) => value.replace(/[^\d-]/g, ''));
  
  const formData = {
    email: '  USER@EXAMPLE.COM  ',
    phone: '090-1234-5678'
  };
  
  const result = validator.validateForm(formData);
  
  return result;
}

/**
 * 条件付きバリデーションの例
 */
export function conditionalValidationExample() {
  const { FormValidator, createRequiredRule, createMinLengthRule } = 
    await import('./unified-validation.js');
  
  const validator = new FormValidator();
  
  // 基本フィールド
  validator.addField('userType', [
    createRequiredRule()
  ]);
  
  validator.addField('adminCode', [
    // 管理者の場合のみ必須
    {
      validate: (value, fieldName, formData) => {
        if (formData && formData.userType === 'admin') {
          return createRequiredRule().validate(value, fieldName);
        }
        return null;
      }
    }
  ]);
  
  const formData1 = {
    userType: 'user',
    adminCode: '' // 一般ユーザーなので空でもOK
  };
  
  const formData2 = {
    userType: 'admin',
    adminCode: '' // 管理者なので必須
  };
  
  const result1 = validator.validateForm(formData1);
  const result2 = validator.validateForm(formData2);
  
  return { result1, result2 };
}

/**
 * エラーメッセージのカスタマイズ例
 */
export function customErrorMessagesExample() {
  const { 
    FormValidator, 
    RequiredRule, 
    MinLengthRule, 
    PatternRule 
  } = await import('./unified-validation.js');
  
  // カスタムエラーメッセージ付きルール
  class CustomRequiredRule extends RequiredRule {
    validate(value, fieldName) {
      const result = super.validate(value, fieldName);
      if (result) {
        result.message = `${fieldName}を入力してください`;
      }
      return result;
    }
  }
  
  class CustomMinLengthRule extends MinLengthRule {
    validate(value, fieldName) {
      const result = super.validate(value, fieldName);
      if (result) {
        result.message = `${fieldName}は最低${this.minLength}文字必要です`;
      }
      return result;
    }
  }
  
  const validator = new FormValidator();
  
  validator.addField('username', [
    new CustomRequiredRule(),
    new CustomMinLengthRule(3)
  ]);
  
  const result = validator.validateForm({ username: 'ab' });
  
  return result;
}

/**
 * 非同期バリデーションの例（将来の拡張用）
 */
export async function asyncValidationExample() {
  // 非同期バリデーション（例：ユーザー名の重複チェック）
  const checkUsernameAvailability = async (username) => {
    // 実際のAPIコールをシミュレート
    await new Promise(resolve => setTimeout(resolve, 500));
    
    const unavailableUsernames = ['admin', 'root', 'test'];
    if (unavailableUsernames.includes(username.toLowerCase())) {
      return {
        isValid: false,
        error: 'このユーザー名は既に使用されています'
      };
    }
    
    return { isValid: true, error: null };
  };
  
  // 使用例
  const result = await checkUsernameAvailability('admin');
  
  return result;
}

// 使用例の実行
if (typeof window !== 'undefined') {
  // ブラウザ環境でのみ実行
  console.log('=== 基本バリデーション例 ===');
  basicValidationExample();
  
  console.log('=== リアルタイムバリデーション例 ===');
  realtimeValidationExample();
  
  console.log('=== 試合結果バリデーション例 ===');
  matchResultValidationExample();
  
  console.log('=== トーナメントバリデーション例 ===');
  tournamentValidationExample();
}
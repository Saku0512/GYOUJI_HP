// リアルタイムバリデーション機能
// Svelteコンポーネント用のリアクティブバリデーション

import { writable, derived } from 'svelte/store';
import { debounce } from './validation.js';

/**
 * リアルタイムバリデーション状態管理
 */
export class RealtimeValidator {
  constructor(formValidator, options = {}) {
    this.formValidator = formValidator;
    this.debounceMs = options.debounceMs || 300;
    this.validateOnChange = options.validateOnChange !== false;
    this.validateOnBlur = options.validateOnBlur !== false;
    
    // ストア
    this.formData = writable({});
    this.errors = writable({});
    this.touched = writable({});
    this.isValidating = writable(false);
    this.sanitizedData = writable({});
    
    // 派生ストア
    this.isValid = derived(
      [this.errors],
      ([$errors]) => Object.keys($errors).length === 0
    );
    
    this.hasErrors = derived(
      [this.errors],
      ([$errors]) => Object.keys($errors).length > 0
    );
    
    // デバウンス付きバリデーション
    this.debouncedValidate = debounce(this._performValidation.bind(this), this.debounceMs);
  }

  /**
   * フィールド値を設定
   */
  setFieldValue(fieldName, value) {
    this.formData.update(data => ({
      ...data,
      [fieldName]: value
    }));

    if (this.validateOnChange) {
      this._scheduleValidation(fieldName);
    }
  }

  /**
   * フィールドをタッチ済みにマーク
   */
  touchField(fieldName) {
    this.touched.update(touched => ({
      ...touched,
      [fieldName]: true
    }));

    if (this.validateOnBlur) {
      this._scheduleValidation(fieldName);
    }
  }

  /**
   * フォーム全体をタッチ済みにマーク
   */
  touchAllFields() {
    this.formData.subscribe(data => {
      const touchedFields = {};
      Object.keys(data).forEach(field => {
        touchedFields[field] = true;
      });
      this.touched.set(touchedFields);
    })();
  }

  /**
   * 特定フィールドのバリデーションをスケジュール
   */
  _scheduleValidation(fieldName) {
    this.isValidating.set(true);
    this.debouncedValidate(fieldName);
  }

  /**
   * バリデーション実行
   */
  _performValidation(fieldName = null) {
    let currentFormData;
    let currentTouched;
    
    this.formData.subscribe(data => currentFormData = data)();
    this.touched.subscribe(touched => currentTouched = touched)();

    if (fieldName) {
      // 単一フィールドのバリデーション
      if (currentTouched[fieldName]) {
        const fieldResult = this.formValidator.validateField(fieldName, currentFormData[fieldName]);
        
        if (fieldResult) {
          this.errors.update(errors => ({
            ...errors,
            [fieldName]: fieldResult.error ? fieldResult.error.message : null
          }));
          
          this.sanitizedData.update(data => ({
            ...data,
            [fieldName]: fieldResult.sanitizedValue
          }));
        }
      }
    } else {
      // フォーム全体のバリデーション
      const result = this.formValidator.validateForm(currentFormData);
      
      // タッチされたフィールドのエラーのみ表示
      const filteredErrors = {};
      Object.keys(result.errors).forEach(field => {
        if (currentTouched[field]) {
          filteredErrors[field] = result.errors[field];
        }
      });
      
      this.errors.set(filteredErrors);
      this.sanitizedData.set(result.sanitizedData);
    }

    this.isValidating.set(false);
  }

  /**
   * フォーム全体のバリデーション（送信時用）
   */
  validateAll() {
    this.touchAllFields();
    
    let currentFormData;
    this.formData.subscribe(data => currentFormData = data)();
    
    const result = this.formValidator.validateForm(currentFormData);
    
    this.errors.set(result.errors);
    this.sanitizedData.set(result.sanitizedData);
    
    return result;
  }

  /**
   * 特定フィールドのエラーをクリア
   */
  clearFieldError(fieldName) {
    this.errors.update(errors => {
      const newErrors = { ...errors };
      delete newErrors[fieldName];
      return newErrors;
    });
  }

  /**
   * 全エラーをクリア
   */
  clearAllErrors() {
    this.errors.set({});
  }

  /**
   * フォームをリセット
   */
  reset() {
    this.formData.set({});
    this.errors.set({});
    this.touched.set({});
    this.sanitizedData.set({});
    this.isValidating.set(false);
  }

  /**
   * フィールドの現在の状態を取得
   */
  getFieldState(fieldName) {
    let currentFormData, currentErrors, currentTouched;
    
    this.formData.subscribe(data => currentFormData = data)();
    this.errors.subscribe(errors => currentErrors = errors)();
    this.touched.subscribe(touched => currentTouched = touched)();
    
    return {
      value: currentFormData[fieldName] || '',
      error: currentErrors[fieldName] || null,
      touched: currentTouched[fieldName] || false,
      hasError: !!(currentErrors[fieldName] && currentTouched[fieldName])
    };
  }

  /**
   * フォームの現在の状態を取得
   */
  getFormState() {
    let currentFormData, currentErrors, currentTouched, currentIsValid, currentSanitizedData;
    
    this.formData.subscribe(data => currentFormData = data)();
    this.errors.subscribe(errors => currentErrors = errors)();
    this.touched.subscribe(touched => currentTouched = touched)();
    this.isValid.subscribe(valid => currentIsValid = valid)();
    this.sanitizedData.subscribe(data => currentSanitizedData = data)();
    
    return {
      formData: currentFormData,
      errors: currentErrors,
      touched: currentTouched,
      isValid: currentIsValid,
      sanitizedData: currentSanitizedData
    };
  }
}

/**
 * Svelte用のリアルタイムバリデーションフック
 */
export function createRealtimeValidator(formValidator, options = {}) {
  return new RealtimeValidator(formValidator, options);
}

/**
 * フィールド用のバリデーションヘルパー
 */
export function createFieldHelper(realtimeValidator, fieldName) {
  return {
    // フィールド値の設定
    setValue: (value) => realtimeValidator.setFieldValue(fieldName, value),
    
    // フィールドをタッチ済みにマーク
    touch: () => realtimeValidator.touchField(fieldName),
    
    // フィー���ド状態の取得
    getState: () => realtimeValidator.getFieldState(fieldName),
    
    // エラーのクリア
    clearError: () => realtimeValidator.clearFieldError(fieldName),
    
    // 入力イベントハンドラー
    handleInput: (event) => {
      const value = event.target.value;
      realtimeValidator.setFieldValue(fieldName, value);
    },
    
    // ブラーイベントハンドラー
    handleBlur: () => {
      realtimeValidator.touchField(fieldName);
    },
    
    // 変更イベントハンドラー
    handleChange: (event) => {
      const value = event.target.value;
      realtimeValidator.setFieldValue(fieldName, value);
      realtimeValidator.touchField(fieldName);
    }
  };
}

/**
 * 条件付きバリデーション
 */
export class ConditionalValidator {
  constructor(condition, validator) {
    this.condition = condition;
    this.validator = validator;
  }

  validate(value, fieldName, formData) {
    if (this.condition(formData)) {
      return this.validator.validate(value, fieldName);
    }
    return null;
  }
}

/**
 * 非同期バリデーション
 */
export class AsyncValidator {
  constructor(asyncValidationFn, debounceMs = 500) {
    this.asyncValidationFn = asyncValidationFn;
    this.debounceMs = debounceMs;
    this.pendingValidations = new Map();
    this.debouncedValidate = debounce(this._performAsyncValidation.bind(this), debounceMs);
  }

  async validate(value, fieldName) {
    // 既存の検証をキャンセル
    if (this.pendingValidations.has(fieldName)) {
      this.pendingValidations.get(fieldName).cancel();
    }

    return new Promise((resolve, reject) => {
      const validation = {
        resolve,
        reject,
        cancel: () => {
          this.pendingValidations.delete(fieldName);
          resolve(null); // キャンセル時はエラーなしとする
        }
      };

      this.pendingValidations.set(fieldName, validation);
      this.debouncedValidate(value, fieldName);
    });
  }

  async _performAsyncValidation(value, fieldName) {
    const validation = this.pendingValidations.get(fieldName);
    if (!validation) return;

    try {
      const result = await this.asyncValidationFn(value, fieldName);
      validation.resolve(result);
    } catch (error) {
      validation.reject(error);
    } finally {
      this.pendingValidations.delete(fieldName);
    }
  }
}

/**
 * バリデーション結果の表示用ヘルパー
 */
export function getValidationClass(hasError, touched) {
  if (!touched) return '';
  return hasError ? 'error' : 'success';
}

export function getValidationVariant(hasError, touched) {
  if (!touched) return 'default';
  return hasError ? 'error' : 'success';
}

/**
 * エラーメッセージの表示判定
 */
export function shouldShowError(error, touched) {
  return !!(error && touched);
}

/**
 * フォーム送信用ヘルパー
 */
export function createSubmitHandler(realtimeValidator, onSubmit) {
  return async (event) => {
    if (event) {
      event.preventDefault();
    }

    const result = realtimeValidator.validateAll();
    
    if (result.isValid) {
      try {
        await onSubmit(result.sanitizedData);
      } catch (error) {
        console.error('Submit error:', error);
        throw error;
      }
    } else {
      // バリデーションエラーがある場合の処理
      console.warn('Form validation failed:', result.errors);
    }
    
    return result;
  };
}
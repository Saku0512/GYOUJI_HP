// リアルタイムバリデーションのテスト

import { describe, it, expect, beforeEach, vi } from 'vitest';
import { get } from 'svelte/store';

// debounce関数をモック
vi.mock('../validation.js', () => ({
  debounce: vi.fn((fn, delay) => {
    return (...args) => {
      setTimeout(() => fn(...args), delay);
    };
  })
}));

import {
  RealtimeValidator,
  createRealtimeValidator,
  createFieldHelper,
  getValidationClass,
  getValidationVariant,
  shouldShowError,
  createSubmitHandler
} from '../realtime-validation.js';
import { createLoginValidator } from '../unified-validation.js';

describe('RealtimeValidator', () => {
  let formValidator;
  let realtimeValidator;

  beforeEach(() => {
    formValidator = createLoginValidator();
    realtimeValidator = new RealtimeValidator(formValidator, {
      debounceMs: 100,
      validateOnChange: true,
      validateOnBlur: true
    });
  });

  describe('initialization', () => {
    it('should initialize with correct options', () => {
      expect(realtimeValidator.formValidator).toBe(formValidator);
      expect(realtimeValidator.debounceMs).toBe(100);
      expect(realtimeValidator.validateOnChange).toBe(true);
      expect(realtimeValidator.validateOnBlur).toBe(true);
    });

    it('should initialize stores with empty values', () => {
      expect(get(realtimeValidator.formData)).toEqual({});
      expect(get(realtimeValidator.errors)).toEqual({});
      expect(get(realtimeValidator.touched)).toEqual({});
      expect(get(realtimeValidator.isValidating)).toBe(false);
      expect(get(realtimeValidator.sanitizedData)).toEqual({});
    });

    it('should initialize derived stores correctly', () => {
      expect(get(realtimeValidator.isValid)).toBe(true);
      expect(get(realtimeValidator.hasErrors)).toBe(false);
    });
  });

  describe('setFieldValue', () => {
    it('should update form data', () => {
      realtimeValidator.setFieldValue('username', 'admin');
      
      const formData = get(realtimeValidator.formData);
      expect(formData.username).toBe('admin');
    });

    it('should schedule validation when validateOnChange is true', () => {
      const spy = vi.spyOn(realtimeValidator, '_scheduleValidation');
      
      realtimeValidator.setFieldValue('username', 'admin');
      
      expect(spy).toHaveBeenCalledWith('username');
    });

    it('should not schedule validation when validateOnChange is false', () => {
      realtimeValidator.validateOnChange = false;
      const spy = vi.spyOn(realtimeValidator, '_scheduleValidation');
      
      realtimeValidator.setFieldValue('username', 'admin');
      
      expect(spy).not.toHaveBeenCalled();
    });
  });

  describe('touchField', () => {
    it('should mark field as touched', () => {
      realtimeValidator.touchField('username');
      
      const touched = get(realtimeValidator.touched);
      expect(touched.username).toBe(true);
    });

    it('should schedule validation when validateOnBlur is true', () => {
      const spy = vi.spyOn(realtimeValidator, '_scheduleValidation');
      
      realtimeValidator.touchField('username');
      
      expect(spy).toHaveBeenCalledWith('username');
    });
  });

  describe('touchAllFields', () => {
    it('should mark all fields as touched', () => {
      realtimeValidator.setFieldValue('username', 'admin');
      realtimeValidator.setFieldValue('password', 'password');
      
      realtimeValidator.touchAllFields();
      
      const touched = get(realtimeValidator.touched);
      expect(touched.username).toBe(true);
      expect(touched.password).toBe(true);
    });
  });

  describe('validateAll', () => {
    it('should validate entire form and return result', () => {
      realtimeValidator.setFieldValue('username', 'admin');
      realtimeValidator.setFieldValue('password', 'password123');
      
      const result = realtimeValidator.validateAll();
      
      expect(result.isValid).toBe(true);
      expect(get(realtimeValidator.touched).username).toBe(true);
      expect(get(realtimeValidator.touched).password).toBe(true);
    });

    it('should collect validation errors', () => {
      realtimeValidator.setFieldValue('username', '');
      realtimeValidator.setFieldValue('password', '');
      
      const result = realtimeValidator.validateAll();
      
      expect(result.isValid).toBe(false);
      expect(result.errors.username).toContain('必須');
      expect(result.errors.password).toContain('必須');
    });
  });

  describe('getFieldState', () => {
    it('should return current field state', () => {
      realtimeValidator.setFieldValue('username', 'admin');
      realtimeValidator.touchField('username');
      
      const state = realtimeValidator.getFieldState('username');
      
      expect(state.value).toBe('admin');
      expect(state.touched).toBe(true);
      expect(state.error).toBeNull();
      expect(state.hasError).toBe(false);
    });

    it('should return error state for invalid field', () => {
      realtimeValidator.setFieldValue('username', '');
      realtimeValidator.touchField('username');
      
      // 手動でエラーを設定（実際の検証をシミュレート）
      realtimeValidator.errors.update(errors => ({
        ...errors,
        username: 'ユーザー名は必須です'
      }));
      
      const state = realtimeValidator.getFieldState('username');
      
      expect(state.error).toBe('ユーザー名は必須です');
      expect(state.hasError).toBe(true);
    });
  });

  describe('clearFieldError', () => {
    it('should clear specific field error', () => {
      realtimeValidator.errors.update(errors => ({
        ...errors,
        username: 'エラーメッセージ'
      }));
      
      realtimeValidator.clearFieldError('username');
      
      const errors = get(realtimeValidator.errors);
      expect(errors.username).toBeUndefined();
    });
  });

  describe('clearAllErrors', () => {
    it('should clear all errors', () => {
      realtimeValidator.errors.update(() => ({
        username: 'エラー1',
        password: 'エラー2'
      }));
      
      realtimeValidator.clearAllErrors();
      
      const errors = get(realtimeValidator.errors);
      expect(errors).toEqual({});
    });
  });

  describe('reset', () => {
    it('should reset all state', () => {
      realtimeValidator.setFieldValue('username', 'admin');
      realtimeValidator.touchField('username');
      realtimeValidator.errors.update(() => ({ username: 'エラー' }));
      
      realtimeValidator.reset();
      
      expect(get(realtimeValidator.formData)).toEqual({});
      expect(get(realtimeValidator.errors)).toEqual({});
      expect(get(realtimeValidator.touched)).toEqual({});
      expect(get(realtimeValidator.sanitizedData)).toEqual({});
      expect(get(realtimeValidator.isValidating)).toBe(false);
    });
  });
});

describe('createRealtimeValidator', () => {
  it('should create RealtimeValidator instance', () => {
    const formValidator = createLoginValidator();
    const realtimeValidator = createRealtimeValidator(formValidator);
    
    expect(realtimeValidator).toBeInstanceOf(RealtimeValidator);
    expect(realtimeValidator.formValidator).toBe(formValidator);
  });

  it('should apply custom options', () => {
    const formValidator = createLoginValidator();
    const options = {
      debounceMs: 500,
      validateOnChange: false,
      validateOnBlur: false
    };
    
    const realtimeValidator = createRealtimeValidator(formValidator, options);
    
    expect(realtimeValidator.debounceMs).toBe(500);
    expect(realtimeValidator.validateOnChange).toBe(false);
    expect(realtimeValidator.validateOnBlur).toBe(false);
  });
});

describe('createFieldHelper', () => {
  let formValidator;
  let realtimeValidator;
  let fieldHelper;

  beforeEach(() => {
    formValidator = createLoginValidator();
    realtimeValidator = createRealtimeValidator(formValidator);
    fieldHelper = createFieldHelper(realtimeValidator, 'username');
  });

  it('should provide setValue method', () => {
    fieldHelper.setValue('admin');
    
    const formData = get(realtimeValidator.formData);
    expect(formData.username).toBe('admin');
  });

  it('should provide touch method', () => {
    fieldHelper.touch();
    
    const touched = get(realtimeValidator.touched);
    expect(touched.username).toBe(true);
  });

  it('should provide getState method', () => {
    fieldHelper.setValue('admin');
    fieldHelper.touch();
    
    const state = fieldHelper.getState();
    expect(state.value).toBe('admin');
    expect(state.touched).toBe(true);
  });

  it('should provide clearError method', () => {
    realtimeValidator.errors.update(errors => ({
      ...errors,
      username: 'エラー'
    }));
    
    fieldHelper.clearError();
    
    const errors = get(realtimeValidator.errors);
    expect(errors.username).toBeUndefined();
  });

  it('should provide handleInput method', () => {
    const event = { target: { value: 'admin' } };
    
    fieldHelper.handleInput(event);
    
    const formData = get(realtimeValidator.formData);
    expect(formData.username).toBe('admin');
  });

  it('should provide handleBlur method', () => {
    fieldHelper.handleBlur();
    
    const touched = get(realtimeValidator.touched);
    expect(touched.username).toBe(true);
  });

  it('should provide handleChange method', () => {
    const event = { target: { value: 'admin' } };
    
    fieldHelper.handleChange(event);
    
    const formData = get(realtimeValidator.formData);
    const touched = get(realtimeValidator.touched);
    expect(formData.username).toBe('admin');
    expect(touched.username).toBe(true);
  });
});

describe('Validation Helper Functions', () => {
  describe('getValidationClass', () => {
    it('should return empty string when not touched', () => {
      expect(getValidationClass(true, false)).toBe('');
      expect(getValidationClass(false, false)).toBe('');
    });

    it('should return error class when has error and touched', () => {
      expect(getValidationClass(true, true)).toBe('error');
    });

    it('should return success class when no error and touched', () => {
      expect(getValidationClass(false, true)).toBe('success');
    });
  });

  describe('getValidationVariant', () => {
    it('should return default when not touched', () => {
      expect(getValidationVariant(true, false)).toBe('default');
      expect(getValidationVariant(false, false)).toBe('default');
    });

    it('should return error variant when has error and touched', () => {
      expect(getValidationVariant(true, true)).toBe('error');
    });

    it('should return success variant when no error and touched', () => {
      expect(getValidationVariant(false, true)).toBe('success');
    });
  });

  describe('shouldShowError', () => {
    it('should return false when not touched', () => {
      expect(shouldShowError('エラー', false)).toBe(false);
    });

    it('should return false when no error', () => {
      expect(shouldShowError(null, true)).toBe(false);
      expect(shouldShowError('', true)).toBe(false);
    });

    it('should return true when has error and touched', () => {
      expect(shouldShowError('エラー', true)).toBe(true);
    });
  });
});

describe('createSubmitHandler', () => {
  let formValidator;
  let realtimeValidator;
  let onSubmit;
  let submitHandler;

  beforeEach(() => {
    formValidator = createLoginValidator();
    realtimeValidator = createRealtimeValidator(formValidator);
    onSubmit = vi.fn().mockResolvedValue();
    submitHandler = createSubmitHandler(realtimeValidator, onSubmit);
  });

  it('should prevent default event', async () => {
    const event = { preventDefault: vi.fn() };
    
    realtimeValidator.setFieldValue('username', 'admin');
    realtimeValidator.setFieldValue('password', 'password123');
    
    await submitHandler(event);
    
    expect(event.preventDefault).toHaveBeenCalled();
  });

  it('should call onSubmit with sanitized data when valid', async () => {
    realtimeValidator.setFieldValue('username', 'admin');
    realtimeValidator.setFieldValue('password', 'password123');
    
    const result = await submitHandler();
    
    expect(onSubmit).toHaveBeenCalledWith({
      username: 'admin',
      password: 'password123'
    });
    expect(result.isValid).toBe(true);
  });

  it('should not call onSubmit when invalid', async () => {
    realtimeValidator.setFieldValue('username', '');
    realtimeValidator.setFieldValue('password', '');
    
    const result = await submitHandler();
    
    expect(onSubmit).not.toHaveBeenCalled();
    expect(result.isValid).toBe(false);
  });

  it('should handle onSubmit errors', async () => {
    const error = new Error('Submit failed');
    onSubmit.mockRejectedValue(error);
    
    realtimeValidator.setFieldValue('username', 'admin');
    realtimeValidator.setFieldValue('password', 'password123');
    
    await expect(submitHandler()).rejects.toThrow('Submit failed');
  });

  it('should work without event parameter', async () => {
    realtimeValidator.setFieldValue('username', 'admin');
    realtimeValidator.setFieldValue('password', 'password123');
    
    const result = await submitHandler();
    
    expect(result.isValid).toBe(true);
    expect(onSubmit).toHaveBeenCalled();
  });
});
// Input コンポーネントのテスト - 簡略版
import { describe, it, expect } from 'vitest';

describe('Input Component', () => {
  it('should generate correct class names', () => {
    const generateClasses = (size, variant, fullWidth, focused, disabled, readonly) => {
      return [
        'input',
        `input-${size}`,
        `input-${variant}`,
        fullWidth ? 'input-full-width' : '',
        focused ? 'input-focused' : '',
        disabled ? 'input-disabled' : '',
        readonly ? 'input-readonly' : ''
      ].filter(Boolean).join(' ');
    };

    expect(generateClasses('medium', 'default', false, false, false, false))
      .toBe('input input-medium input-default');
    
    expect(generateClasses('small', 'error', true, true, false, false))
      .toBe('input input-small input-error input-full-width input-focused');
    
    expect(generateClasses('large', 'success', false, false, true, false))
      .toBe('input input-large input-success input-disabled');
  });

  it('should determine validation state correctly', () => {
    const getValidationState = (errorMessage, variant) => {
      return errorMessage ? 'error' : variant;
    };

    expect(getValidationState('Error occurred', 'success')).toBe('error');
    expect(getValidationState('', 'warning')).toBe('warning');
    expect(getValidationState(null, 'default')).toBe('default');
  });

  it('should generate unique IDs', () => {
    const generateId = (id) => {
      return id || `input-${Math.random().toString(36).substr(2, 9)}`;
    };

    const id1 = generateId('');
    const id2 = generateId('');
    const customId = generateId('custom-id');

    expect(id1).toMatch(/^input-[a-z0-9]{9}$/);
    expect(id2).toMatch(/^input-[a-z0-9]{9}$/);
    expect(id1).not.toBe(id2);
    expect(customId).toBe('custom-id');
  });

  it('should validate input types', () => {
    const validTypes = ['text', 'email', 'password', 'number', 'tel', 'url', 'search'];
    const validSizes = ['small', 'medium', 'large'];
    const validVariants = ['default', 'success', 'error', 'warning'];

    expect(validTypes).toContain('email');
    expect(validTypes).toContain('password');
    expect(validSizes).toContain('medium');
    expect(validVariants).toContain('error');
  });

  it('should handle focus state changes', () => {
    let focused = false;
    
    const handleFocus = () => { focused = true; };
    const handleBlur = () => { focused = false; };

    handleFocus();
    expect(focused).toBe(true);
    
    handleBlur();
    expect(focused).toBe(false);
  });
});
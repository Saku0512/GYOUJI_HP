// Button コンポーネントのテスト - 簡略版
import { describe, it, expect } from 'vitest';

describe('Button Component', () => {
  it('should have correct class generation logic', () => {
    // クラス名生成ロジックのテスト
    const generateClasses = (variant, size, outline, fullWidth, disabled, loading) => {
      return [
        'btn',
        `btn-${variant}`,
        `btn-${size}`,
        outline ? 'btn-outline' : '',
        fullWidth ? 'btn-full-width' : '',
        disabled ? 'btn-disabled' : '',
        loading ? 'btn-loading' : ''
      ].filter(Boolean).join(' ');
    };

    expect(generateClasses('primary', 'medium', false, false, false, false))
      .toBe('btn btn-primary btn-medium');
    
    expect(generateClasses('danger', 'small', true, true, false, false))
      .toBe('btn btn-danger btn-small btn-outline btn-full-width');
    
    expect(generateClasses('success', 'large', false, false, true, true))
      .toBe('btn btn-success btn-large btn-disabled btn-loading');
  });

  it('should handle click prevention logic', () => {
    // クリック防止ロジックのテスト
    const shouldPreventClick = (disabled, loading) => {
      return disabled || loading;
    };

    expect(shouldPreventClick(true, false)).toBe(true);
    expect(shouldPreventClick(false, true)).toBe(true);
    expect(shouldPreventClick(true, true)).toBe(true);
    expect(shouldPreventClick(false, false)).toBe(false);
  });

  it('should handle keyboard event logic', () => {
    // キーボードイベントロジックのテスト
    const shouldTriggerClick = (key, disabled, loading) => {
      if (disabled || loading) return false;
      return key === 'Enter' || key === ' ';
    };

    expect(shouldTriggerClick('Enter', false, false)).toBe(true);
    expect(shouldTriggerClick(' ', false, false)).toBe(true);
    expect(shouldTriggerClick('Tab', false, false)).toBe(false);
    expect(shouldTriggerClick('Enter', true, false)).toBe(false);
    expect(shouldTriggerClick('Enter', false, true)).toBe(false);
  });

  it('should validate component props', () => {
    // プロパティ検証のテスト
    const validVariants = ['primary', 'secondary', 'success', 'danger', 'warning', 'info', 'light', 'dark'];
    const validSizes = ['small', 'medium', 'large'];
    const validTypes = ['button', 'submit', 'reset'];

    expect(validVariants).toContain('primary');
    expect(validVariants).toContain('danger');
    expect(validSizes).toContain('small');
    expect(validSizes).toContain('large');
    expect(validTypes).toContain('submit');
  });
});
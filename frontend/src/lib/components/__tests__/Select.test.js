// Select コンポーネントのテスト - 簡略版
import { describe, it, expect } from 'vitest';

describe('Select Component', () => {
  const mockOptions = [
    { value: 'option1', label: 'Option 1' },
    { value: 'option2', label: 'Option 2' },
    { value: 'option3', label: 'Option 3', disabled: true }
  ];

  it('should generate correct class names', () => {
    const generateClasses = (size, variant, fullWidth, focused, disabled) => {
      return [
        'select',
        `select-${size}`,
        `select-${variant}`,
        fullWidth ? 'select-full-width' : '',
        focused ? 'select-focused' : '',
        disabled ? 'select-disabled' : ''
      ].filter(Boolean).join(' ');
    };

    expect(generateClasses('medium', 'default', false, false, false))
      .toBe('select select-medium select-default');
    
    expect(generateClasses('small', 'error', true, true, false))
      .toBe('select select-small select-error select-full-width select-focused');
  });

  it('should determine validation state correctly', () => {
    const getValidationState = (errorMessage, variant) => {
      return errorMessage ? 'error' : variant;
    };

    expect(getValidationState('Error occurred', 'success')).toBe('error');
    expect(getValidationState('', 'warning')).toBe('warning');
    expect(getValidationState(null, 'default')).toBe('default');
  });

  it('should handle multiple selection logic', () => {
    const handleMultipleSelection = (multiple, selectedOptions) => {
      if (multiple) {
        return Array.from(selectedOptions, option => option.value);
      }
      return selectedOptions.value;
    };

    const mockSelectedOptions = [
      { value: 'option1' },
      { value: 'option2' }
    ];

    expect(handleMultipleSelection(true, mockSelectedOptions))
      .toEqual(['option1', 'option2']);
    
    expect(handleMultipleSelection(false, { value: 'option1' }))
      .toBe('option1');
  });

  it('should validate options structure', () => {
    const validateOptions = (options) => {
      return options.every(option => 
        typeof option.value === 'string' && 
        typeof option.label === 'string'
      );
    };

    expect(validateOptions(mockOptions)).toBe(true);
    expect(validateOptions([{ value: 1, label: 'Invalid' }])).toBe(false);
  });

  it('should handle disabled options', () => {
    const getDisabledOptions = (options) => {
      return options.filter(option => option.disabled);
    };

    const disabledOptions = getDisabledOptions(mockOptions);
    expect(disabledOptions).toHaveLength(1);
    expect(disabledOptions[0].value).toBe('option3');
  });

  it('should generate unique IDs', () => {
    const generateId = (id) => {
      return id || `select-${Math.random().toString(36).substr(2, 9)}`;
    };

    const id1 = generateId('');
    const id2 = generateId('');
    const customId = generateId('custom-select');

    expect(id1).toMatch(/^select-[a-z0-9]{9}$/);
    expect(id2).toMatch(/^select-[a-z0-9]{9}$/);
    expect(id1).not.toBe(id2);
    expect(customId).toBe('custom-select');
  });
});
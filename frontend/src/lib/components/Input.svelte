<script>
  // Input コンポーネント - 基本入力フィールド
  import { createEventDispatcher } from 'svelte';
  
  export let type = 'text'; // 'text', 'email', 'password', 'number', 'tel', 'url', 'search'
  export let value = '';
  export let placeholder = '';
  export let disabled = false;
  export let readonly = false;
  export let required = false;
  export let size = 'medium'; // 'small', 'medium', 'large'
  export let variant = 'default'; // 'default', 'success', 'error', 'warning'
  export let label = '';
  export let helperText = '';
  export let errorMessage = '';
  export let id = '';
  export let name = '';
  export let autocomplete = '';
  export let min = null;
  export let max = null;
  export let step = null;
  export let maxlength = null;
  export let pattern = null;
  export let fullWidth = false;
  
  const dispatch = createEventDispatcher();
  
  let inputElement;
  let focused = false;
  
  // 一意のIDを生成
  $: inputId = id || `input-${Math.random().toString(36).substr(2, 9)}`;
  
  // バリデーション状態を決定
  $: validationState = errorMessage ? 'error' : variant;
  
  // クラス名を構築
  $: classes = [
    'input',
    `input-${size}`,
    `input-${validationState}`,
    fullWidth ? 'input-full-width' : '',
    focused ? 'input-focused' : '',
    disabled ? 'input-disabled' : '',
    readonly ? 'input-readonly' : ''
  ].filter(Boolean).join(' ');
  
  function handleInput(event) {
    value = event.target.value;
    dispatch('input', { value, event });
  }
  
  function handleChange(event) {
    dispatch('change', { value, event });
  }
  
  function handleFocus(event) {
    focused = true;
    dispatch('focus', { value, event });
  }
  
  function handleBlur(event) {
    focused = false;
    dispatch('blur', { value, event });
  }
  
  function handleKeydown(event) {
    dispatch('keydown', { value, event });
  }
  
  // 外部からフォーカスを設定するためのメソッド
  export function focus() {
    if (inputElement) {
      inputElement.focus();
    }
  }
  
  export function blur() {
    if (inputElement) {
      inputElement.blur();
    }
  }
</script>

<div class="input-container" class:input-container-full-width={fullWidth}>
  {#if label}
    <label for={inputId} class="input-label" class:input-label-required={required}>
      {label}
      {#if required}
        <span class="input-required-mark" aria-label="必須">*</span>
      {/if}
    </label>
  {/if}
  
  <div class="input-wrapper">
    <input
      bind:this={inputElement}
      {type}
      id={inputId}
      {name}
      {placeholder}
      {disabled}
      {readonly}
      {required}
      {autocomplete}
      {min}
      {max}
      {step}
      {maxlength}
      {pattern}
      bind:value
      class={classes}
      on:input={handleInput}
      on:change={handleChange}
      on:focus={handleFocus}
      on:blur={handleBlur}
      on:keydown={handleKeydown}
      aria-describedby={helperText || errorMessage ? `${inputId}-help` : undefined}
      aria-invalid={errorMessage ? 'true' : 'false'}
    />
    
    <slot name="suffix" />
  </div>
  
  {#if helperText && !errorMessage}
    <div id="{inputId}-help" class="input-helper-text">
      {helperText}
    </div>
  {/if}
  
  {#if errorMessage}
    <div id="{inputId}-help" class="input-error-message" role="alert">
      {errorMessage}
    </div>
  {/if}
</div>

<style>
  .input-container {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  
  .input-container-full-width {
    width: 100%;
  }
  
  .input-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
    margin-bottom: 0.25rem;
  }
  
  .input-required-mark {
    color: #dc3545;
    margin-left: 0.125rem;
  }
  
  .input-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }
  
  .input {
    font-family: inherit;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    background-color: white;
    color: #111827;
    transition: all 0.2s ease-in-out;
    width: 100%;
  }
  
  .input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
  }
  
  /* サイズ */
  .input-small {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
  }
  
  .input-medium {
    padding: 0.5rem 0.75rem;
    font-size: 1rem;
    line-height: 1.5rem;
  }
  
  .input-large {
    padding: 0.75rem 1rem;
    font-size: 1.125rem;
    line-height: 1.75rem;
  }
  
  /* バリアント */
  .input-default {
    border-color: #d1d5db;
  }
  
  .input-success {
    border-color: #10b981;
  }
  
  .input-success:focus {
    border-color: #10b981;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
  }
  
  .input-error {
    border-color: #ef4444;
  }
  
  .input-error:focus {
    border-color: #ef4444;
    box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
  }
  
  .input-warning {
    border-color: #f59e0b;
  }
  
  .input-warning:focus {
    border-color: #f59e0b;
    box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.1);
  }
  
  /* 状態 */
  .input-disabled {
    background-color: #f9fafb;
    color: #6b7280;
    cursor: not-allowed;
  }
  
  .input-readonly {
    background-color: #f9fafb;
    cursor: default;
  }
  
  .input-full-width {
    width: 100%;
  }
  
  /* ヘルパーテキスト */
  .input-helper-text {
    font-size: 0.75rem;
    color: #6b7280;
    margin-top: 0.25rem;
  }
  
  .input-error-message {
    font-size: 0.75rem;
    color: #ef4444;
    margin-top: 0.25rem;
  }
  
  /* プレースホルダー */
  .input::placeholder {
    color: #9ca3af;
  }
  
  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .input-small {
      padding: 0.5rem 0.75rem;
      font-size: 1rem; /* モバイルでのズーム防止 */
    }
    
    .input-medium {
      padding: 0.625rem 0.75rem;
      font-size: 1rem;
    }
    
    .input-large {
      padding: 0.75rem 1rem;
      font-size: 1rem;
    }
  }
</style>
<script>
  // Select コンポーネント - セレクトボックス
  import { createEventDispatcher } from 'svelte';
  
  export let value = '';
  export let options = []; // { value, label, disabled? } の配列
  export let placeholder = '選択してください';
  export let disabled = false;
  export let required = false;
  export let size = 'medium'; // 'small', 'medium', 'large'
  export let variant = 'default'; // 'default', 'success', 'error', 'warning'
  export let label = '';
  export let helperText = '';
  export let errorMessage = '';
  export let id = '';
  export let name = '';
  export let multiple = false;
  export let fullWidth = false;
  
  const dispatch = createEventDispatcher();
  
  let selectElement;
  let focused = false;
  
  // 一意のIDを生成
  $: selectId = id || `select-${Math.random().toString(36).substr(2, 9)}`;
  
  // バリデーション状態を決定
  $: validationState = errorMessage ? 'error' : variant;
  
  // クラス名を構築
  $: classes = [
    'select',
    `select-${size}`,
    `select-${validationState}`,
    fullWidth ? 'select-full-width' : '',
    focused ? 'select-focused' : '',
    disabled ? 'select-disabled' : ''
  ].filter(Boolean).join(' ');
  
  function handleChange(event) {
    if (multiple) {
      value = Array.from(event.target.selectedOptions, option => option.value);
    } else {
      value = event.target.value;
    }
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
    if (selectElement) {
      selectElement.focus();
    }
  }
  
  export function blur() {
    if (selectElement) {
      selectElement.blur();
    }
  }
</script>

<div class="select-container" class:select-container-full-width={fullWidth}>
  {#if label}
    <label for={selectId} class="select-label" class:select-label-required={required}>
      {label}
      {#if required}
        <span class="select-required-mark" aria-label="必須">*</span>
      {/if}
    </label>
  {/if}
  
  <div class="select-wrapper">
    <select
      bind:this={selectElement}
      {id}={selectId}
      {name}
      {disabled}
      {required}
      {multiple}
      bind:value
      class={classes}
      on:change={handleChange}
      on:focus={handleFocus}
      on:blur={handleBlur}
      on:keydown={handleKeydown}
      aria-describedby={helperText || errorMessage ? `${selectId}-help` : undefined}
      aria-invalid={errorMessage ? 'true' : 'false'}
    >
      {#if !multiple && placeholder}
        <option value="" disabled selected={!value}>
          {placeholder}
        </option>
      {/if}
      
      {#each options as option}
        <option
          value={option.value}
          disabled={option.disabled || false}
        >
          {option.label}
        </option>
      {/each}
    </select>
    
    {#if !multiple}
      <div class="select-arrow" aria-hidden="true">
        <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
          <path fill-rule="evenodd" d="M5.293 7.293a1 1 0 011.414 0L10 10.586l3.293-3.293a1 1 0 111.414 1.414l-4 4a1 1 0 01-1.414 0l-4-4a1 1 0 010-1.414z" clip-rule="evenodd" />
        </svg>
      </div>
    {/if}
  </div>
  
  {#if helperText && !errorMessage}
    <div id="{selectId}-help" class="select-helper-text">
      {helperText}
    </div>
  {/if}
  
  {#if errorMessage}
    <div id="{selectId}-help" class="select-error-message" role="alert">
      {errorMessage}
    </div>
  {/if}
</div>

<style>
  .select-container {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  
  .select-container-full-width {
    width: 100%;
  }
  
  .select-label {
    font-size: 0.875rem;
    font-weight: 500;
    color: #374151;
    margin-bottom: 0.25rem;
  }
  
  .select-required-mark {
    color: #dc3545;
    margin-left: 0.125rem;
  }
  
  .select-wrapper {
    position: relative;
    display: flex;
    align-items: center;
  }
  
  .select {
    font-family: inherit;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    background-color: white;
    color: #111827;
    transition: all 0.2s ease-in-out;
    width: 100%;
    cursor: pointer;
    appearance: none;
    -webkit-appearance: none;
    -moz-appearance: none;
  }
  
  .select:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
  }
  
  /* サイズ */
  .select-small {
    padding: 0.375rem 2rem 0.375rem 0.75rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
  }
  
  .select-medium {
    padding: 0.5rem 2.5rem 0.5rem 0.75rem;
    font-size: 1rem;
    line-height: 1.5rem;
  }
  
  .select-large {
    padding: 0.75rem 3rem 0.75rem 1rem;
    font-size: 1.125rem;
    line-height: 1.75rem;
  }
  
  /* マルチプル選択の場合はパディングを調整 */
  .select[multiple] {
    padding-right: 0.75rem;
    min-height: 6rem;
  }
  
  /* バリアント */
  .select-default {
    border-color: #d1d5db;
  }
  
  .select-success {
    border-color: #10b981;
  }
  
  .select-success:focus {
    border-color: #10b981;
    box-shadow: 0 0 0 3px rgba(16, 185, 129, 0.1);
  }
  
  .select-error {
    border-color: #ef4444;
  }
  
  .select-error:focus {
    border-color: #ef4444;
    box-shadow: 0 0 0 3px rgba(239, 68, 68, 0.1);
  }
  
  .select-warning {
    border-color: #f59e0b;
  }
  
  .select-warning:focus {
    border-color: #f59e0b;
    box-shadow: 0 0 0 3px rgba(245, 158, 11, 0.1);
  }
  
  /* 状態 */
  .select-disabled {
    background-color: #f9fafb;
    color: #6b7280;
    cursor: not-allowed;
  }
  
  .select-full-width {
    width: 100%;
  }
  
  /* 矢印アイコン */
  .select-arrow {
    position: absolute;
    right: 0.75rem;
    top: 50%;
    transform: translateY(-50%);
    pointer-events: none;
    color: #6b7280;
  }
  
  .select-small + .select-arrow {
    right: 0.5rem;
  }
  
  .select-large + .select-arrow {
    right: 1rem;
  }
  
  /* ヘルパーテキスト */
  .select-helper-text {
    font-size: 0.75rem;
    color: #6b7280;
    margin-top: 0.25rem;
  }
  
  .select-error-message {
    font-size: 0.75rem;
    color: #ef4444;
    margin-top: 0.25rem;
  }
  
  /* オプションのスタイル */
  .select option {
    padding: 0.5rem;
  }
  
  .select option:disabled {
    color: #9ca3af;
  }
  
  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .select-small {
      padding: 0.5rem 2rem 0.5rem 0.75rem;
      font-size: 1rem; /* モバイルでのズーム防止 */
    }
    
    .select-medium {
      padding: 0.625rem 2.5rem 0.625rem 0.75rem;
      font-size: 1rem;
    }
    
    .select-large {
      padding: 0.75rem 3rem 0.75rem 1rem;
      font-size: 1rem;
    }
  }
</style>
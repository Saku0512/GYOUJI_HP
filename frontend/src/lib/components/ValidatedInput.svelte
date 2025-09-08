<script>
  // ValidatedInput コンポーネント - 統一バリデーション機能付き入力フィールド
  import { createEventDispatcher, onMount } from 'svelte';
  import Input from './Input.svelte';
  import ValidationMessage from './ValidationMessage.svelte';
  import { createFieldHelper } from '../utils/realtime-validation.js';
  
  // Props
  export let realtimeValidator = null;  // RealtimeValidator インスタンス
  export let fieldName = '';            // フィールド名
  export let type = 'text';
  export let value = '';
  export let placeholder = '';
  export let disabled = false;
  export let readonly = false;
  export let required = false;
  export let size = 'medium';
  export let label = '';
  export let helperText = '';
  export let id = '';
  export let name = '';
  export let autocomplete = '';
  export let min = null;
  export let max = null;
  export let step = null;
  export let maxlength = null;
  export let pattern = null;
  export let fullWidth = false;
  export let validateOnChange = true;   // 変更時にバリデーション
  export let validateOnBlur = true;     // ブラー時にバリデーション
  export let showValidationIcon = true; // バリデーション状態アイコン表示
  
  const dispatch = createEventDispatcher();
  
  let inputComponent;
  let fieldHelper;
  let fieldState = {
    value: '',
    error: null,
    touched: false,
    hasError: false
  };
  
  // フィールドヘルパーの初期化
  $: if (realtimeValidator && fieldName) {
    fieldHelper = createFieldHelper(realtimeValidator, fieldName);
    updateFieldState();
  }
  
  // 外部からの値の変更を監視
  $: if (fieldHelper && value !== fieldState.value) {
    fieldHelper.setValue(value);
    updateFieldState();
  }
  
  // フィールド状態の更新
  function updateFieldState() {
    if (fieldHelper) {
      fieldState = fieldHelper.getState();
    }
  }
  
  // バリデーション状態の監視（リアクティブ）
  $: if (realtimeValidator) {
    // ストアの変更を監視
    const unsubscribeFormData = realtimeValidator.formData.subscribe(() => updateFieldState());
    const unsubscribeErrors = realtimeValidator.errors.subscribe(() => updateFieldState());
    const unsubscribeTouched = realtimeValidator.touched.subscribe(() => updateFieldState());
  }
  
  // バリデーション状態に基づくバリアント
  $: validationVariant = getValidationVariant(fieldState.hasError, fieldState.touched);
  
  // 表示するエラーメッセージ
  $: displayError = fieldState.hasError ? fieldState.error : null;
  
  function getValidationVariant(hasError, touched) {
    if (!touched) return 'default';
    return hasError ? 'error' : 'success';
  }
  
  function handleInput(event) {
    const newValue = event.detail.value;
    
    if (fieldHelper && validateOnChange) {
      fieldHelper.setValue(newValue);
      updateFieldState();
    }
    
    // 外部への通知
    dispatch('input', {
      value: newValue,
      fieldName,
      fieldState: fieldState,
      event: event.detail.event
    });
  }
  
  function handleChange(event) {
    const newValue = event.detail.value;
    
    if (fieldHelper) {
      fieldHelper.setValue(newValue);
      if (validateOnChange) {
        fieldHelper.touch();
      }
      updateFieldState();
    }
    
    dispatch('change', {
      value: newValue,
      fieldName,
      fieldState: fieldState,
      event: event.detail.event
    });
  }
  
  function handleBlur(event) {
    if (fieldHelper && validateOnBlur) {
      fieldHelper.touch();
      updateFieldState();
    }
    
    dispatch('blur', {
      value: fieldState.value,
      fieldName,
      fieldState: fieldState,
      event: event.detail.event
    });
  }
  
  function handleFocus(event) {
    dispatch('focus', {
      value: fieldState.value,
      fieldName,
      fieldState: fieldState,
      event: event.detail.event
    });
  }
  
  function handleKeydown(event) {
    dispatch('keydown', {
      value: fieldState.value,
      fieldName,
      fieldState: fieldState,
      event: event.detail.event
    });
  }
  
  // 外部からフォーカスを設定するためのメソッド
  export function focus() {
    if (inputComponent) {
      inputComponent.focus();
    }
  }
  
  export function blur() {
    if (inputComponent) {
      inputComponent.blur();
    }
  }
  
  // フィールド状態の取得
  export function getFieldState() {
    return fieldState;
  }
  
  // エラーのクリア
  export function clearError() {
    if (fieldHelper) {
      fieldHelper.clearError();
      updateFieldState();
    }
  }
  
  onMount(() => {
    // 初期値の設定
    if (fieldHelper && value) {
      fieldHelper.setValue(value);
      updateFieldState();
    }
  });
</script>

<div class="validated-input-container" class:validated-input-full-width={fullWidth}>
  <Input
    bind:this={inputComponent}
    {type}
    value={fieldState.value || value}
    {placeholder}
    {disabled}
    {readonly}
    {required}
    {size}
    variant={validationVariant}
    {label}
    helperText={!displayError ? helperText : ''}
    errorMessage=""
    {id}
    {name}
    {autocomplete}
    {min}
    {max}
    {step}
    {maxlength}
    {pattern}
    {fullWidth}
    on:input={handleInput}
    on:change={handleChange}
    on:blur={handleBlur}
    on:focus={handleFocus}
    on:keydown={handleKeydown}
  >
    <svelte:fragment slot="suffix">
      {#if showValidationIcon && fieldState.touched}
        <div class="validation-icon">
          {#if fieldState.hasError}
            <span class="validation-icon-error" aria-label="エラー">⚠️</span>
          {:else}
            <span class="validation-icon-success" aria-label="正常">✅</span>
          {/if}
        </div>
      {/if}
      <slot name="suffix" />
    </svelte:fragment>
  </Input>
  
  <ValidationMessage
    error={displayError}
    touched={fieldState.touched}
    type="error"
    size={size === 'large' ? 'medium' : 'small'}
    {fullWidth}
  />
</div>

<style>
  .validated-input-container {
    display: flex;
    flex-direction: column;
  }
  
  .validated-input-full-width {
    width: 100%;
  }
  
  .validation-icon {
    display: flex;
    align-items: center;
    padding-right: 0.5rem;
    pointer-events: none;
  }
  
  .validation-icon-error {
    color: #ef4444;
    font-size: 1rem;
  }
  
  .validation-icon-success {
    color: #10b981;
    font-size: 1rem;
  }
  
  /* アニメーション */
  .validation-icon {
    transition: opacity 0.2s ease-in-out;
  }
  
  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .validation-icon {
      padding-right: 0.375rem;
    }
    
    .validation-icon-error,
    .validation-icon-success {
      font-size: 0.875rem;
    }
  }
  
  /* 縮小モーション設定対応 */
  @media (prefers-reduced-motion: reduce) {
    .validation-icon {
      transition: none;
    }
  }
</style>
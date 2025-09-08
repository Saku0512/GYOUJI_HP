<script>
  // ValidationMessage コンポーネント - 統一されたバリデーションメッセージ表示
  import { createEventDispatcher } from 'svelte';
  
  export let error = null;           // エラーメッセージ
  export let touched = false;        // フィールドがタッチされたか
  export let type = 'error';         // 'error', 'warning', 'info', 'success'
  export let show = true;            // 表示するかどうか
  export let icon = true;            // アイコンを表示するか
  export let dismissible = false;    // 閉じるボタンを表示するか
  export let animate = true;         // アニメーションを有効にするか
  export let size = 'medium';        // 'small', 'medium', 'large'
  export let fullWidth = false;      // 全幅表示
  
  const dispatch = createEventDispatcher();
  
  // 表示判定
  $: shouldShow = show && error && touched;
  
  // アイコンの決定
  $: iconClass = getIconClass(type);
  
  // クラス名の構築
  $: classes = [
    'validation-message',
    `validation-message-${type}`,
    `validation-message-${size}`,
    fullWidth ? 'validation-message-full-width' : '',
    animate ? 'validation-message-animate' : '',
    dismissible ? 'validation-message-dismissible' : ''
  ].filter(Boolean).join(' ');
  
  function getIconClass(messageType) {
    const iconMap = {
      error: '⚠️',
      warning: '⚠️',
      info: 'ℹ️',
      success: '✅'
    };
    return iconMap[messageType] || iconMap.error;
  }
  
  function handleDismiss() {
    dispatch('dismiss');
  }
</script>

{#if shouldShow}
  <div 
    class={classes}
    role="alert"
    aria-live="polite"
    transition:slide={{ duration: animate ? 200 : 0 }}
  >
    {#if icon}
      <span class="validation-message-icon" aria-hidden="true">
        {iconClass}
      </span>
    {/if}
    
    <span class="validation-message-text">
      {error}
    </span>
    
    {#if dismissible}
      <button 
        class="validation-message-dismiss"
        on:click={handleDismiss}
        aria-label="エラーメッセージを閉じる"
        type="button"
      >
        ×
      </button>
    {/if}
  </div>
{/if}

<script context="module">
  import { slide } from 'svelte/transition';
</script>

<style>
  .validation-message {
    display: flex;
    align-items: flex-start;
    gap: 0.5rem;
    padding: 0.5rem 0.75rem;
    border-radius: 0.375rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
    border: 1px solid;
    margin-top: 0.25rem;
  }
  
  .validation-message-animate {
    transition: all 0.2s ease-in-out;
  }
  
  .validation-message-full-width {
    width: 100%;
  }
  
  /* サイズ */
  .validation-message-small {
    padding: 0.25rem 0.5rem;
    font-size: 0.75rem;
    line-height: 1rem;
  }
  
  .validation-message-medium {
    padding: 0.5rem 0.75rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
  }
  
  .validation-message-large {
    padding: 0.75rem 1rem;
    font-size: 1rem;
    line-height: 1.5rem;
  }
  
  /* タイプ別スタイル */
  .validation-message-error {
    background-color: #fef2f2;
    border-color: #fecaca;
    color: #dc2626;
  }
  
  .validation-message-warning {
    background-color: #fffbeb;
    border-color: #fed7aa;
    color: #d97706;
  }
  
  .validation-message-info {
    background-color: #eff6ff;
    border-color: #bfdbfe;
    color: #2563eb;
  }
  
  .validation-message-success {
    background-color: #f0fdf4;
    border-color: #bbf7d0;
    color: #16a34a;
  }
  
  .validation-message-icon {
    flex-shrink: 0;
    font-size: 1em;
    line-height: 1;
  }
  
  .validation-message-text {
    flex: 1;
    word-break: break-word;
  }
  
  .validation-message-dismiss {
    flex-shrink: 0;
    background: none;
    border: none;
    color: inherit;
    cursor: pointer;
    font-size: 1.25em;
    line-height: 1;
    padding: 0;
    margin-left: 0.5rem;
    opacity: 0.7;
    transition: opacity 0.2s ease-in-out;
  }
  
  .validation-message-dismiss:hover {
    opacity: 1;
  }
  
  .validation-message-dismiss:focus {
    outline: 2px solid currentColor;
    outline-offset: 2px;
    opacity: 1;
  }
  
  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .validation-message-error {
      background-color: #450a0a;
      border-color: #7f1d1d;
      color: #fca5a5;
    }
    
    .validation-message-warning {
      background-color: #451a03;
      border-color: #92400e;
      color: #fbbf24;
    }
    
    .validation-message-info {
      background-color: #1e3a8a;
      border-color: #3730a3;
      color: #93c5fd;
    }
    
    .validation-message-success {
      background-color: #14532d;
      border-color: #166534;
      color: #86efac;
    }
  }
  
  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .validation-message {
      border-width: 2px;
    }
    
    .validation-message-error {
      background-color: #ffffff;
      border-color: #dc2626;
      color: #dc2626;
    }
    
    .validation-message-warning {
      background-color: #ffffff;
      border-color: #d97706;
      color: #d97706;
    }
    
    .validation-message-info {
      background-color: #ffffff;
      border-color: #2563eb;
      color: #2563eb;
    }
    
    .validation-message-success {
      background-color: #ffffff;
      border-color: #16a34a;
      color: #16a34a;
    }
  }
  
  /* 縮小モーション設定対応 */
  @media (prefers-reduced-motion: reduce) {
    .validation-message-animate {
      transition: none;
    }
    
    .validation-message-dismiss {
      transition: none;
    }
  }
  
  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .validation-message-small {
      padding: 0.375rem 0.5rem;
      font-size: 0.875rem; /* モバイルでの可読性向上 */
    }
    
    .validation-message-medium {
      padding: 0.5rem 0.75rem;
      font-size: 0.875rem;
    }
    
    .validation-message-large {
      padding: 0.625rem 0.875rem;
      font-size: 0.875rem;
    }
  }
</style>
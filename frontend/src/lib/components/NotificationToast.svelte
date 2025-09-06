<script>
  // NotificationToast コンポーネント - 通知表示
  import { createEventDispatcher } from 'svelte';
  
  export let message = '';
  export let type = 'info'; // 'success', 'error', 'warning', 'info'
  export let duration = 5000; // 自動消去時間（ミリ秒）
  export let dismissible = true; // 手動で閉じることができるか
  
  const dispatch = createEventDispatcher();
  
  let visible = true;
  let timeoutId;
  
  // 自動消去タイマーを設定
  if (duration > 0) {
    timeoutId = setTimeout(() => {
      close();
    }, duration);
  }
  
  function close() {
    visible = false;
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
    dispatch('close');
  }
  
  // アクセシビリティ: 要素の参照
  let toastElement;
</script>

{#if visible}
  <div
    bind:this={toastElement}
    class="toast toast-{type}"
    role="alert"
    aria-live="polite"
    aria-atomic="true"
  >
    <div class="toast-content">
      <div class="toast-icon">
        {#if type === 'success'}
          <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clip-rule="evenodd" />
          </svg>
        {:else if type === 'error'}
          <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        {:else if type === 'warning'}
          <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
          </svg>
        {:else}
          <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
          </svg>
        {/if}
      </div>
      <div class="toast-message">
        {message}
      </div>
      {#if dismissible}
        <button
          class="toast-close"
          on:click={close}
          aria-label="通知を閉じる"
          type="button"
        >
          <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
        </button>
      {/if}
    </div>
  </div>
{/if}

<style>
  .toast {
    position: fixed;
    top: 1rem;
    right: 1rem;
    min-width: 300px;
    max-width: 500px;
    padding: 1rem;
    border-radius: 0.5rem;
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05);
    z-index: 1000;
    animation: slideIn 0.3s ease-out;
  }
  
  .toast:focus {
    outline: 2px solid #007bff;
    outline-offset: 2px;
  }
  
  .toast-success {
    background-color: #d1fae5;
    border: 1px solid #a7f3d0;
    color: #065f46;
  }
  
  .toast-error {
    background-color: #fee2e2;
    border: 1px solid #fca5a5;
    color: #991b1b;
  }
  
  .toast-warning {
    background-color: #fef3c7;
    border: 1px solid #fcd34d;
    color: #92400e;
  }
  
  .toast-info {
    background-color: #dbeafe;
    border: 1px solid #93c5fd;
    color: #1e40af;
  }
  
  .toast-content {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
  }
  
  .toast-icon {
    flex-shrink: 0;
    margin-top: 0.125rem;
  }
  
  .toast-message {
    flex: 1;
    font-size: 0.875rem;
    line-height: 1.25rem;
  }
  
  .toast-close {
    flex-shrink: 0;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.25rem;
    border-radius: 0.25rem;
    opacity: 0.7;
    transition: opacity 0.2s;
  }
  
  .toast-close:hover {
    opacity: 1;
  }
  
  .toast-close:focus {
    outline: 2px solid currentColor;
    outline-offset: 2px;
  }
  
  @keyframes slideIn {
    from {
      transform: translateX(100%);
      opacity: 0;
    }
    to {
      transform: translateX(0);
      opacity: 1;
    }
  }
  
  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .toast {
      left: 1rem;
      right: 1rem;
      min-width: auto;
      max-width: none;
    }
  }
</style>
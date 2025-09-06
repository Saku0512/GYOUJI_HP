<script>
  // Button コンポーネント - 基本ボタン
  export let variant = 'primary'; // 'primary', 'secondary', 'success', 'danger', 'warning', 'info', 'light', 'dark'
  export let size = 'medium'; // 'small', 'medium', 'large'
  export let disabled = false;
  export let loading = false;
  export let type = 'button'; // 'button', 'submit', 'reset'
  export let href = null; // リンクとして使用する場合
  export let target = null; // リンクのターゲット
  export let fullWidth = false; // 全幅表示
  export let outline = false; // アウトラインスタイル

  // クリックイベントハンドラー
  function handleClick(event) {
    if (disabled || loading) {
      event.preventDefault();
      return;
    }
  }

  // キーボードイベントハンドラー
  function handleKeydown(event) {
    if (disabled || loading) {
      return;
    }

    // Enter または Space キーでクリック
    if (event.key === 'Enter' || event.key === ' ') {
      event.preventDefault();
      event.currentTarget.click();
    }
  }

  $: classes = [
    'btn',
    `btn-${variant}`,
    `btn-${size}`,
    outline ? 'btn-outline' : '',
    fullWidth ? 'btn-full-width' : '',
    disabled ? 'btn-disabled' : '',
    loading ? 'btn-loading' : ''
  ]
    .filter(Boolean)
    .join(' ');
</script>

{#if href}
  <a
    {href}
    {target}
    class={classes}
    class:disabled
    on:click={handleClick}
    on:keydown={handleKeydown}
    role="button"
    tabindex={disabled ? -1 : 0}
    aria-disabled={disabled}
  >
    {#if loading}
      <span class="btn-spinner" aria-hidden="true"></span>
    {/if}
    <slot />
  </a>
{:else}
  <button {type} {disabled} class={classes} on:click={handleClick} aria-disabled={disabled}>
    {#if loading}
      <span class="btn-spinner" aria-hidden="true"></span>
    {/if}
    <slot />
  </button>
{/if}

<style>
  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    font-family: inherit;
    font-weight: 500;
    text-align: center;
    text-decoration: none;
    border: 1px solid transparent;
    border-radius: 0.375rem;
    cursor: pointer;
    transition: all 0.2s ease-in-out;
    position: relative;
    overflow: hidden;
  }

  .btn:focus {
    outline: 2px solid #007bff;
    outline-offset: 2px;
  }

  /* サイズ */
  .btn-small {
    padding: 0.375rem 0.75rem;
    font-size: 0.875rem;
    line-height: 1.25rem;
  }

  .btn-medium {
    padding: 0.5rem 1rem;
    font-size: 1rem;
    line-height: 1.5rem;
  }

  .btn-large {
    padding: 0.75rem 1.5rem;
    font-size: 1.125rem;
    line-height: 1.75rem;
  }

  /* バリアント - Primary */
  .btn-primary {
    background-color: #007bff;
    border-color: #007bff;
    color: white;
  }

  .btn-primary:hover:not(.btn-disabled) {
    background-color: #0056b3;
    border-color: #0056b3;
  }

  .btn-primary.btn-outline {
    background-color: transparent;
    color: #007bff;
  }

  .btn-primary.btn-outline:hover:not(.btn-disabled) {
    background-color: #007bff;
    color: white;
  }

  /* バリアント - Secondary */
  .btn-secondary {
    background-color: #6c757d;
    border-color: #6c757d;
    color: white;
  }

  .btn-secondary:hover:not(.btn-disabled) {
    background-color: #545b62;
    border-color: #545b62;
  }

  .btn-secondary.btn-outline {
    background-color: transparent;
    color: #6c757d;
  }

  .btn-secondary.btn-outline:hover:not(.btn-disabled) {
    background-color: #6c757d;
    color: white;
  }

  /* バリアント - Success */
  .btn-success {
    background-color: #28a745;
    border-color: #28a745;
    color: white;
  }

  .btn-success:hover:not(.btn-disabled) {
    background-color: #1e7e34;
    border-color: #1e7e34;
  }

  .btn-success.btn-outline {
    background-color: transparent;
    color: #28a745;
  }

  .btn-success.btn-outline:hover:not(.btn-disabled) {
    background-color: #28a745;
    color: white;
  }

  /* バリアント - Danger */
  .btn-danger {
    background-color: #dc3545;
    border-color: #dc3545;
    color: white;
  }

  .btn-danger:hover:not(.btn-disabled) {
    background-color: #c82333;
    border-color: #c82333;
  }

  .btn-danger.btn-outline {
    background-color: transparent;
    color: #dc3545;
  }

  .btn-danger.btn-outline:hover:not(.btn-disabled) {
    background-color: #dc3545;
    color: white;
  }

  /* バリアント - Warning */
  .btn-warning {
    background-color: #ffc107;
    border-color: #ffc107;
    color: #212529;
  }

  .btn-warning:hover:not(.btn-disabled) {
    background-color: #e0a800;
    border-color: #e0a800;
  }

  .btn-warning.btn-outline {
    background-color: transparent;
    color: #ffc107;
  }

  .btn-warning.btn-outline:hover:not(.btn-disabled) {
    background-color: #ffc107;
    color: #212529;
  }

  /* バリアント - Info */
  .btn-info {
    background-color: #17a2b8;
    border-color: #17a2b8;
    color: white;
  }

  .btn-info:hover:not(.btn-disabled) {
    background-color: #117a8b;
    border-color: #117a8b;
  }

  .btn-info.btn-outline {
    background-color: transparent;
    color: #17a2b8;
  }

  .btn-info.btn-outline:hover:not(.btn-disabled) {
    background-color: #17a2b8;
    color: white;
  }

  /* バリアント - Light */
  .btn-light {
    background-color: #f8f9fa;
    border-color: #f8f9fa;
    color: #212529;
  }

  .btn-light:hover:not(.btn-disabled) {
    background-color: #e2e6ea;
    border-color: #e2e6ea;
  }

  .btn-light.btn-outline {
    background-color: transparent;
    color: #f8f9fa;
    border-color: #f8f9fa;
  }

  .btn-light.btn-outline:hover:not(.btn-disabled) {
    background-color: #f8f9fa;
    color: #212529;
  }

  /* バリアント - Dark */
  .btn-dark {
    background-color: #343a40;
    border-color: #343a40;
    color: white;
  }

  .btn-dark:hover:not(.btn-disabled) {
    background-color: #23272b;
    border-color: #23272b;
  }

  .btn-dark.btn-outline {
    background-color: transparent;
    color: #343a40;
  }

  .btn-dark.btn-outline:hover:not(.btn-disabled) {
    background-color: #343a40;
    color: white;
  }

  /* 状態 */
  .btn-disabled {
    opacity: 0.6;
    cursor: not-allowed;
    pointer-events: none;
  }

  .btn-loading {
    cursor: wait;
  }

  .btn-full-width {
    width: 100%;
  }

  /* ローディングスピナー */
  .btn-spinner {
    width: 1rem;
    height: 1rem;
    border: 2px solid transparent;
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .btn-small {
      padding: 0.5rem 0.75rem;
      font-size: 0.875rem;
    }

    .btn-medium {
      padding: 0.625rem 1rem;
      font-size: 1rem;
    }

    .btn-large {
      padding: 0.75rem 1.25rem;
      font-size: 1rem;
    }
  }
</style>

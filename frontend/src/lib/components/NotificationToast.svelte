<script>
  // NotificationToast コンポーネント - 通知表示
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  
  export let message = '';
  export let type = 'info'; // 'success', 'error', 'warning', 'info'
  export let duration = 5000; // 自動消去時間（ミリ秒）
  export let dismissible = true; // 手動で閉じることができるか
  export let isPaused = false; // 一時停止状態
  export let showProgress = true; // プログレスバーを表示するか
  export let actions = []; // アクションボタンの配列 [{ label, onClick, variant }]
  
  const dispatch = createEventDispatcher();
  
  let visible = true;
  let timeoutId;
  let startTime;
  let remainingTime = duration;
  let progressElement;
  let toastElement;
  
  // プログレスバーの状態
  let progressWidth = 100;
  let progressAnimationId;
  
  onMount(() => {
    if (duration > 0) {
      startTime = Date.now();
      startAutoRemoval();
    }
  });
  
  onDestroy(() => {
    clearAutoRemoval();
  });
  
  // 一時停止状態の変更を監視
  $: {
    if (duration > 0) {
      if (isPaused) {
        pauseAutoRemoval();
      } else {
        resumeAutoRemoval();
      }
    }
  }
  
  // 自動消去の開始
  function startAutoRemoval() {
    if (duration <= 0) return;
    
    clearAutoRemoval();
    startTime = Date.now();
    remainingTime = duration;
    
    timeoutId = setTimeout(() => {
      close();
    }, duration);
    
    if (showProgress) {
      startProgressAnimation();
    }
  }
  
  // 自動消去の一時停止
  function pauseAutoRemoval() {
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
      
      // 残り時間を計算
      const elapsed = Date.now() - startTime;
      remainingTime = Math.max(0, duration - elapsed);
    }
    
    if (progressAnimationId) {
      cancelAnimationFrame(progressAnimationId);
      progressAnimationId = null;
    }
  }
  
  // 自動消去の再開
  function resumeAutoRemoval() {
    if (remainingTime > 0 && !timeoutId) {
      startTime = Date.now();
      
      timeoutId = setTimeout(() => {
        close();
      }, remainingTime);
      
      if (showProgress) {
        startProgressAnimation();
      }
    }
  }
  
  // 自動消去のクリア
  function clearAutoRemoval() {
    if (timeoutId) {
      clearTimeout(timeoutId);
      timeoutId = null;
    }
    
    if (progressAnimationId) {
      cancelAnimationFrame(progressAnimationId);
      progressAnimationId = null;
    }
  }
  
  // プログレスアニメーションの開始
  function startProgressAnimation() {
    if (!showProgress || duration <= 0) return;
    
    const animate = () => {
      if (!visible || isPaused) return;
      
      const elapsed = Date.now() - startTime;
      const progress = Math.max(0, 1 - elapsed / remainingTime);
      progressWidth = progress * 100;
      
      if (progress > 0) {
        progressAnimationId = requestAnimationFrame(animate);
      }
    };
    
    progressAnimationId = requestAnimationFrame(animate);
  }
  
  // 通知を閉じる
  function close() {
    visible = false;
    clearAutoRemoval();
    dispatch('close');
  }
  
  // アクションボタンのクリック処理
  function handleActionClick(action) {
    try {
      action.onClick();
      // アクションが成功した場合は通知を閉じる
      close();
    } catch (error) {
      console.error('Notification action error:', error);
    }
  }
  
  // キーボードナビゲーション
  function handleKeyDown(event) {
    if (event.key === 'Escape' && dismissible) {
      close();
    }
  }
  
  // アクセシビリティ: 通知タイプに基づくARIAロール
  $: ariaRole = type === 'error' ? 'alert' : 'status';
  
  // アクセシビリティ: 通知の重要度
  $: ariaLive = type === 'error' ? 'assertive' : 'polite';
  
  // プログレスバーの色
  $: progressColor = {
    success: '#10b981',
    error: '#ef4444',
    warning: '#f59e0b',
    info: '#3b82f6'
  }[type] || '#6b7280';
</script>

{#if visible}
  <div
    bind:this={toastElement}
    class="toast toast-{type}"
    class:paused={isPaused}
    role={ariaRole}
    aria-live={ariaLive}
    aria-atomic="true"
    tabindex="0"
    on:keydown={handleKeyDown}
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
      
      <!-- アクションボタン -->
      {#if actions && actions.length > 0}
        <div class="toast-actions">
          {#each actions as action}
            <button
              class="toast-action toast-action-{action.variant || 'primary'}"
              on:click={() => handleActionClick(action)}
              type="button"
            >
              {action.label}
            </button>
          {/each}
        </div>
      {/if}
      
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
    
    <!-- プログレスバー -->
    {#if showProgress && duration > 0}
      <div class="toast-progress" bind:this={progressElement}>
        <div 
          class="toast-progress-bar"
          style="width: {progressWidth}%; background-color: {progressColor};"
        ></div>
      </div>
    {/if}
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

  /* アクションボタン */
  .toast-actions {
    display: flex;
    gap: 0.5rem;
    margin-top: 0.75rem;
    flex-wrap: wrap;
  }

  .toast-action {
    padding: 0.25rem 0.75rem;
    border: 1px solid transparent;
    border-radius: 4px;
    font-size: 0.75rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .toast-action-primary {
    background-color: #3b82f6;
    color: white;
  }

  .toast-action-primary:hover {
    background-color: #2563eb;
  }

  .toast-action-secondary {
    background-color: transparent;
    color: currentColor;
    border-color: currentColor;
  }

  .toast-action-secondary:hover {
    background-color: rgba(0, 0, 0, 0.1);
  }

  .toast-action:focus {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  /* プログレスバー */
  .toast-progress {
    position: absolute;
    bottom: 0;
    left: 0;
    right: 0;
    height: 3px;
    background-color: rgba(0, 0, 0, 0.1);
    border-radius: 0 0 0.5rem 0.5rem;
    overflow: hidden;
  }

  .toast-progress-bar {
    height: 100%;
    transition: width 0.1s linear;
    border-radius: 0 0 0.5rem 0.5rem;
  }

  /* 一時停止状態 */
  .toast.paused {
    box-shadow: 0 10px 15px -3px rgba(0, 0, 0, 0.1), 0 4px 6px -2px rgba(0, 0, 0, 0.05), 0 0 0 2px #3b82f6;
  }

  .toast.paused .toast-progress-bar {
    animation-play-state: paused;
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
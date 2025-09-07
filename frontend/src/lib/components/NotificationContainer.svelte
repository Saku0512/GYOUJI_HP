<script>
  // NotificationContainer コンポーネント - 通知システムのコンテナ
  import { onMount, onDestroy } from 'svelte';
  import { uiStore, uiActions } from '$lib/stores/ui.js';
  import NotificationToast from './NotificationToast.svelte';
  import { fly } from 'svelte/transition';
  import { flip } from 'svelte/animate';

  // Props
  export let position = 'top-right'; // 'top-right', 'top-left', 'bottom-right', 'bottom-left', 'top-center', 'bottom-center'
  export let maxNotifications = 5; // 最大表示通知数
  export let defaultDuration = 5000; // デフォルト表示時間
  export let pauseOnHover = true; // ホバー時に自動消去を一時停止
  export let stackDirection = 'down'; // 'up' or 'down' - 新しい通知の積み重ね方向
  export let animationDuration = 300; // アニメーション時間
  export let spacing = 8; // 通知間のスペース（px）

  // 内部状態
  let notifications = [];
  let containerElement;
  let isHovered = false;
  let pausedNotifications = new Set();

  // ストアの購読
  const unsubscribe = uiStore.subscribe(state => {
    notifications = state.notifications.slice(-maxNotifications);
  });

  // クリーンアップ
  onDestroy(() => {
    unsubscribe();
  });

  // 通知の削除処理
  function handleNotificationClose(notificationId) {
    uiActions.removeNotification(notificationId);
    pausedNotifications.delete(notificationId);
  }

  // 通知のクリア処理
  function handleClearAll() {
    uiActions.clearNotifications();
    pausedNotifications.clear();
  }

  // ホバー開始処理
  function handleMouseEnter() {
    if (!pauseOnHover) return;
    
    isHovered = true;
    // 現在表示中の通知を一時停止リストに追加
    notifications.forEach(notification => {
      pausedNotifications.add(notification.id);
    });
  }

  // ホバー終了処理
  function handleMouseLeave() {
    if (!pauseOnHover) return;
    
    isHovered = false;
    // 一時停止を解除し、残り時間で自動消去を再開
    pausedNotifications.forEach(notificationId => {
      const notification = notifications.find(n => n.id === notificationId);
      if (notification) {
        const elapsed = Date.now() - notification.timestamp;
        const remaining = defaultDuration - elapsed;
        
        if (remaining > 0) {
          setTimeout(() => {
            if (!pausedNotifications.has(notificationId)) {
              uiActions.removeNotification(notificationId);
            }
          }, remaining);
        }
      }
    });
    pausedNotifications.clear();
  }

  // 通知の自動消去設定
  function setupAutoRemoval(notification) {
    if (notification.duration === 0) return; // 手動消去のみ
    
    const duration = notification.duration || defaultDuration;
    
    setTimeout(() => {
      if (!pausedNotifications.has(notification.id)) {
        uiActions.removeNotification(notification.id);
      }
    }, duration);
  }

  // 新しい通知の監視
  $: {
    notifications.forEach(notification => {
      if (!notification._autoRemovalSet) {
        notification._autoRemovalSet = true;
        setupAutoRemoval(notification);
      }
    });
  }

  // ポジションに基づくCSSクラス
  $: positionClass = `position-${position}`;
  $: stackClass = `stack-${stackDirection}`;

  // キーボードナビゲーション
  function handleKeyDown(event) {
    if (event.key === 'Escape') {
      handleClearAll();
    }
  }

  // アクセシビリティ: フォーカス管理
  function handleNotificationFocus(event) {
    // 通知にフォーカスが当たった場合の処理
    const notificationId = parseInt(event.target.dataset.notificationId);
    if (notificationId && pauseOnHover) {
      pausedNotifications.add(notificationId);
    }
  }

  function handleNotificationBlur(event) {
    // 通知からフォーカスが外れた場合の処理
    const notificationId = parseInt(event.target.dataset.notificationId);
    if (notificationId) {
      pausedNotifications.delete(notificationId);
    }
  }
</script>

<svelte:window on:keydown={handleKeyDown} />

{#if notifications.length > 0}
  <div
    bind:this={containerElement}
    class="notification-container {positionClass} {stackClass}"
    style="--spacing: {spacing}px; --animation-duration: {animationDuration}ms;"
    on:mouseenter={handleMouseEnter}
    on:mouseleave={handleMouseLeave}
    role="region"
    aria-label="通知"
    aria-live="polite"
  >
    <!-- クリアボタン（複数通知がある場合） -->
    {#if notifications.length > 1}
      <div class="notification-header">
        <button
          class="clear-all-button"
          on:click={handleClearAll}
          aria-label="すべての通知をクリア"
          type="button"
        >
          <svg width="16" height="16" viewBox="0 0 20 20" fill="currentColor">
            <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
          </svg>
          すべてクリア
        </button>
      </div>
    {/if}

    <!-- 通知リスト -->
    <div class="notifications-list">
      {#each notifications as notification (notification.id)}
        <div
          class="notification-wrapper"
          data-notification-id={notification.id}
          on:focus={handleNotificationFocus}
          on:blur={handleNotificationBlur}
          in:fly={{
            x: position.includes('right') ? 300 : position.includes('left') ? -300 : 0,
            y: position.includes('top') ? -50 : position.includes('bottom') ? 50 : 0,
            duration: animationDuration
          }}
          out:fly={{
            x: position.includes('right') ? 300 : position.includes('left') ? -300 : 0,
            duration: animationDuration
          }}
          animate:flip={{ duration: animationDuration }}
        >
          <NotificationToast
            message={notification.message}
            type={notification.type}
            duration={0}
            dismissible={true}
            isPaused={pausedNotifications.has(notification.id)}
            on:close={() => handleNotificationClose(notification.id)}
          />
        </div>
      {/each}
    </div>

    <!-- 通知数インジケーター（最大数を超えた場合） -->
    {#if $uiStore.notifications.length > maxNotifications}
      <div class="notification-overflow">
        <span class="overflow-text">
          他に {$uiStore.notifications.length - maxNotifications} 件の通知があります
        </span>
        <button
          class="show-all-button"
          on:click={() => maxNotifications = $uiStore.notifications.length}
          type="button"
        >
          すべて表示
        </button>
      </div>
    {/if}
  </div>
{/if}

<style>
  .notification-container {
    position: fixed;
    z-index: 1000;
    display: flex;
    flex-direction: column;
    gap: var(--spacing);
    max-width: 400px;
    min-width: 300px;
  }

  /* ポジション設定 */
  .position-top-right {
    top: 1rem;
    right: 1rem;
  }

  .position-top-left {
    top: 1rem;
    left: 1rem;
  }

  .position-bottom-right {
    bottom: 1rem;
    right: 1rem;
  }

  .position-bottom-left {
    bottom: 1rem;
    left: 1rem;
  }

  .position-top-center {
    top: 1rem;
    left: 50%;
    transform: translateX(-50%);
  }

  .position-bottom-center {
    bottom: 1rem;
    left: 50%;
    transform: translateX(-50%);
  }

  /* スタック方向 */
  .stack-up {
    flex-direction: column-reverse;
  }

  .stack-down {
    flex-direction: column;
  }

  /* ヘッダー */
  .notification-header {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 0.5rem;
  }

  .clear-all-button {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.25rem 0.5rem;
    background-color: rgba(0, 0, 0, 0.1);
    border: 1px solid rgba(0, 0, 0, 0.2);
    border-radius: 4px;
    color: #6b7280;
    font-size: 0.75rem;
    cursor: pointer;
    transition: all 0.2s ease;
  }

  .clear-all-button:hover {
    background-color: rgba(0, 0, 0, 0.15);
    color: #374151;
  }

  .clear-all-button:focus {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  /* 通知リスト */
  .notifications-list {
    display: flex;
    flex-direction: inherit;
    gap: var(--spacing);
  }

  .notification-wrapper {
    width: 100%;
  }

  /* オーバーフロー表示 */
  .notification-overflow {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0.5rem;
    background-color: #f3f4f6;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    font-size: 0.75rem;
    color: #6b7280;
  }

  .overflow-text {
    flex: 1;
  }

  .show-all-button {
    padding: 0.25rem 0.5rem;
    background-color: #3b82f6;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 0.75rem;
    cursor: pointer;
    transition: background-color 0.2s ease;
  }

  .show-all-button:hover {
    background-color: #2563eb;
  }

  .show-all-button:focus {
    outline: 2px solid #1d4ed8;
    outline-offset: 2px;
  }

  /* レスポンシブデザイン */
  @media (max-width: 768px) {
    .notification-container {
      left: 1rem !important;
      right: 1rem !important;
      max-width: none;
      min-width: auto;
      transform: none !important;
    }

    .position-top-center,
    .position-bottom-center {
      left: 1rem;
      right: 1rem;
      transform: none;
    }
  }

  @media (max-width: 480px) {
    .notification-container {
      left: 0.5rem !important;
      right: 0.5rem !important;
    }

    .clear-all-button {
      font-size: 0.625rem;
      padding: 0.125rem 0.25rem;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .clear-all-button {
      background-color: rgba(255, 255, 255, 0.1);
      border-color: rgba(255, 255, 255, 0.2);
      color: #d1d5db;
    }

    .clear-all-button:hover {
      background-color: rgba(255, 255, 255, 0.15);
      color: #f9fafb;
    }

    .notification-overflow {
      background-color: #374151;
      border-color: #4b5563;
      color: #d1d5db;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .notification-wrapper {
      transition: none;
    }
  }

  /* ハイコントラストモード */
  @media (prefers-contrast: high) {
    .clear-all-button {
      border-width: 2px;
    }

    .notification-overflow {
      border-width: 2px;
    }
  }
</style>
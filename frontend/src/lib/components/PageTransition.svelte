<script>
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { navigating } from '$app/stores';
  import AnimatedTransition from './AnimatedTransition.svelte';

  // ページ遷移アニメーションコンポーネント
  export let transitionType = 'fade';
  export let duration = 300;
  export let className = '';

  let isNavigating = false;
  let currentPath = '';
  let showContent = true;

  // ナビゲーション状態の監視
  $: if ($navigating) {
    isNavigating = true;
    showContent = false;
  } else if (isNavigating) {
    // ナビゲーション完了後、少し遅延してコンテンツを表示
    setTimeout(() => {
      showContent = true;
      isNavigating = false;
    }, 50);
  }

  // ページパスの変更を監視
  $: if ($page.url.pathname !== currentPath) {
    currentPath = $page.url.pathname;
    if (!isNavigating) {
      showContent = true;
    }
  }

  onMount(() => {
    showContent = true;
    currentPath = $page.url.pathname;
  });
</script>

<div class="page-transition-container {className}">
  <AnimatedTransition
    show={showContent}
    type={transitionType}
    {duration}
    direction="up"
    className="page-content"
  >
    <slot />
  </AnimatedTransition>

  {#if isNavigating}
    <div class="loading-overlay">
      <div class="loading-spinner"></div>
    </div>
  {/if}
</div>

<style>
  .page-transition-container {
    position: relative;
    min-height: 100%;
  }

  .page-content {
    width: 100%;
  }

  .loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(255, 255, 255, 0.8);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 9999;
    backdrop-filter: blur(2px);
  }

  .loading-spinner {
    width: 40px;
    height: 40px;
    border: 3px solid #f3f4f6;
    border-top: 3px solid #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .loading-overlay {
      background-color: rgba(31, 41, 55, 0.8);
    }

    .loading-spinner {
      border-color: #374151;
      border-top-color: #60a5fa;
    }
  }

  /* アニメーション無効化 */
  @media (prefers-reduced-motion: reduce) {
    .loading-spinner {
      animation: none;
    }
  }
</style>
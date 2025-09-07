<script>
  import { onMount, onDestroy } from 'svelte';
  import { writable } from 'svelte/store';

  // レスポンシブレイアウトコンポーネント
  export let breakpoints = {
    mobile: 768,
    tablet: 1024,
    desktop: 1200
  };
  
  export let container = true;
  export let fluid = false;
  export let padding = true;
  export let className = '';

  // 現在の画面サイズを管理するストア
  const screenSize = writable('desktop');
  const screenWidth = writable(0);

  let mounted = false;
  let resizeObserver;

  // 画面サイズの判定
  function getScreenSize(width) {
    if (width < breakpoints.mobile) return 'mobile';
    if (width < breakpoints.tablet) return 'tablet';
    if (width < breakpoints.desktop) return 'desktop';
    return 'large';
  }

  // ウィンドウサイズの更新
  function updateScreenSize() {
    if (typeof window !== 'undefined') {
      const width = window.innerWidth;
      screenWidth.set(width);
      screenSize.set(getScreenSize(width));
    }
  }

  // ResizeObserverを使用したサイズ監視
  function setupResizeObserver() {
    if (typeof window !== 'undefined' && 'ResizeObserver' in window) {
      resizeObserver = new ResizeObserver(() => {
        updateScreenSize();
      });
      resizeObserver.observe(document.documentElement);
    } else {
      // フォールバック: resize イベント
      window.addEventListener('resize', updateScreenSize);
    }
  }

  // クリーンアップ
  function cleanup() {
    if (resizeObserver) {
      resizeObserver.disconnect();
    } else if (typeof window !== 'undefined') {
      window.removeEventListener('resize', updateScreenSize);
    }
  }

  onMount(() => {
    mounted = true;
    updateScreenSize();
    setupResizeObserver();
  });

  onDestroy(() => {
    cleanup();
  });

  // クラス名の計算
  $: containerClass = [
    'responsive-layout',
    container && !fluid ? 'container' : '',
    fluid ? 'container-fluid' : '',
    padding ? 'with-padding' : '',
    `screen-${$screenSize}`,
    className
  ].filter(Boolean).join(' ');

  // 子コンポーネントに画面サイズ情報を提供
  export { screenSize, screenWidth };
</script>

<div class={containerClass} data-screen-size={$screenSize}>
  <slot {screenSize} {screenWidth} />
</div>

<style>
  .responsive-layout {
    width: 100%;
    position: relative;
  }

  .container {
    max-width: 1200px;
    margin: 0 auto;
  }

  .container-fluid {
    width: 100%;
  }

  .with-padding {
    padding: 0 1rem;
  }

  /* 画面サイズ別のスタイル */
  .screen-mobile.with-padding {
    padding: 0 0.5rem;
  }

  .screen-tablet.with-padding {
    padding: 0 1rem;
  }

  .screen-desktop.with-padding {
    padding: 0 1.5rem;
  }

  .screen-large.with-padding {
    padding: 0 2rem;
  }

  /* ブレークポイント別の最大幅 */
  @media (max-width: 767px) {
    .container {
      max-width: 100%;
    }
  }

  @media (min-width: 768px) and (max-width: 1023px) {
    .container {
      max-width: 750px;
    }
  }

  @media (min-width: 1024px) and (max-width: 1199px) {
    .container {
      max-width: 970px;
    }
  }

  @media (min-width: 1200px) {
    .container {
      max-width: 1170px;
    }
  }
</style>
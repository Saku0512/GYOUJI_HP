<!--
  遅延読み込みコンポーネント
  動的インポートを使用してコンポーネントを必要時にのみ読み込む
-->
<script>
  import { onMount } from 'svelte';
  import { lazyLoader, performanceMonitor } from '$lib/utils/performance.js';
  import LoadingSpinner from './LoadingSpinner.svelte';
  import ErrorBoundary from './ErrorBoundary.svelte';

  // Props
  export let componentImport; // () => import('./Component.svelte')
  export let componentName = 'LazyComponent';
  export let loadingText = '読み込み中...';
  export let errorTitle = 'コンポーネント読み込みエラー';
  export let errorMessage = 'コンポーネントの読み込みに失敗しました。';
  export let retryText = '再試行';
  export let showLoadingSpinner = true;
  export let minLoadingTime = 0; // 最小読み込み時間（ms）
  export let props = {}; // 遅延読み込みするコンポーネントに渡すprops

  // 状態管理
  let component = null;
  let loading = true;
  let error = null;
  let retryCount = 0;

  // コンポーネントの読み込み
  async function loadComponent() {
    if (!componentImport) {
      error = new Error('componentImport prop is required');
      loading = false;
      return;
    }

    try {
      loading = true;
      error = null;

      const startTime = performance.now();

      // 遅延読み込み実行
      const module = await lazyLoader.loadModule(componentImport, componentName);
      
      // 最小読み込み時間の確保（UX向上のため）
      if (minLoadingTime > 0) {
        const elapsed = performance.now() - startTime;
        if (elapsed < minLoadingTime) {
          await new Promise(resolve => setTimeout(resolve, minLoadingTime - elapsed));
        }
      }

      // デフォルトエクスポートまたは名前付きエクスポートを取得
      component = module.default || module[componentName] || module;
      
      if (!component) {
        throw new Error(`Component not found in module: ${componentName}`);
      }

      // パフォーマンスメトリクスを記録
      performanceMonitor.recordMetric(`lazy-component-${componentName}`, {
        type: 'component-load',
        loadTime: performance.now() - startTime,
        retryCount,
        timestamp: Date.now()
      });

    } catch (err) {
      console.error(`Failed to load component ${componentName}:`, err);
      error = err;
      
      // エラーメトリクスを記録
      performanceMonitor.recordMetric(`lazy-component-error-${componentName}`, {
        type: 'component-error',
        error: err.message,
        retryCount,
        timestamp: Date.now()
      });
    } finally {
      loading = false;
    }
  }

  // 再試行処理
  async function handleRetry() {
    retryCount++;
    await loadComponent();
  }

  // マウント時に読み込み開始
  onMount(() => {
    loadComponent();
  });
</script>

{#if loading}
  <div class="lazy-loading" role="status" aria-live="polite">
    {#if showLoadingSpinner}
      <LoadingSpinner size="medium" />
    {/if}
    <p class="loading-text">{loadingText}</p>
  </div>
{:else if error}
  <ErrorBoundary
    {errorTitle}
    {errorMessage}
    showRetry={true}
    {retryText}
    onRetry={handleRetry}
    error={error}
  />
{:else if component}
  <svelte:component this={component} {...props} />
{:else}
  <div class="lazy-error" role="alert">
    <p>コンポーネントが見つかりません</p>
  </div>
{/if}

<style>
  .lazy-loading {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    min-height: 200px;
  }

  .loading-text {
    margin-top: 1rem;
    color: #6c757d;
    font-size: 0.875rem;
    text-align: center;
  }

  .lazy-error {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem;
    min-height: 200px;
    background-color: #f8f9fa;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    color: #6c757d;
    text-align: center;
  }

  /* レスポンシブデザイン */
  @media (max-width: 768px) {
    .lazy-loading {
      padding: 1rem;
      min-height: 150px;
    }

    .lazy-error {
      padding: 1rem;
      min-height: 150px;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .loading-text {
      color: #adb5bd;
    }

    .lazy-error {
      background-color: #343a40;
      border-color: #495057;
      color: #adb5bd;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .lazy-loading {
      transition: none;
    }
  }
</style>
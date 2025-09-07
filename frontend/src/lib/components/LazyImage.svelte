<!--
  遅延読み込み画像コンポーネント
  パフォーマンス最適化とユーザーエクスペリエンス向上のための画像コンポーネント
-->
<script>
  import { onMount, onDestroy } from 'svelte';
  import { setupLazyImage, imageOptimizer } from '$lib/utils/image-optimization.js';
  import { performanceMonitor } from '$lib/utils/performance.js';

  // Props
  export let src = '';
  export let alt = '';
  export let width = undefined;
  export let height = undefined;
  export let placeholder = undefined;
  export let fallback = undefined;
  export let loading = 'lazy'; // 'lazy' | 'eager'
  export let className = '';
  export let style = '';
  export let sizes = undefined;
  export let srcset = undefined;
  export let objectFit = 'cover'; // 'cover' | 'contain' | 'fill' | 'scale-down' | 'none'
  export let objectPosition = 'center';
  export let priority = false; // 優先読み込み
  export let quality = 75; // 画質（将来の拡張用）
  export let blur = true; // プレースホルダーのぼかし効果
  export let fadeIn = true; // フェードイン効果
  export let onLoad = undefined; // 読み込み完了コールバック
  export let onError = undefined; // エラーコールバック

  // 内部状態
  let imgElement;
  let isLoaded = false;
  let isError = false;
  let isLoading = false;
  let currentSrc = '';

  // プレースホルダー画像の生成
  $: placeholderSrc = placeholder || generatePlaceholder(width, height);
  
  // CSS クラスの組み立て
  $: imgClasses = [
    'lazy-image',
    className,
    isLoading ? 'lazy-image--loading' : '',
    isLoaded ? 'lazy-image--loaded' : '',
    isError ? 'lazy-image--error' : '',
    fadeIn ? 'lazy-image--fade' : ''
  ].filter(Boolean).join(' ');

  // CSS スタイルの組み立て
  $: imgStyles = [
    style,
    `object-fit: ${objectFit}`,
    `object-position: ${objectPosition}`,
    width ? `width: ${typeof width === 'number' ? width + 'px' : width}` : '',
    height ? `height: ${typeof height === 'number' ? height + 'px' : height}` : ''
  ].filter(Boolean).join('; ');

  /**
   * プレースホルダー画像を生成
   */
  function generatePlaceholder(w, h) {
    const defaultWidth = w || 300;
    const defaultHeight = h || 200;
    const bgColor = '#f8f9fa';
    const textColor = '#6c757d';
    
    if (blur) {
      // ぼかし効果付きのプレースホルダー
      return `data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${defaultWidth} ${defaultHeight}"%3E%3Cdefs%3E%3Cfilter id="blur"%3E%3CfeGaussianBlur stdDeviation="2"/%3E%3C/filter%3E%3C/defs%3E%3Crect width="100%25" height="100%25" fill="${bgColor.replace('#', '%23')}" filter="url(%23blur)"/%3E%3C/svg%3E`;
    } else {
      // シンプルなプレースホルダー
      return `data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${defaultWidth} ${defaultHeight}"%3E%3Crect width="100%25" height="100%25" fill="${bgColor.replace('#', '%23')}"/%3E%3Ctext x="50%25" y="50%25" text-anchor="middle" dy=".3em" fill="${textColor.replace('#', '%23')}" font-family="system-ui,sans-serif" font-size="14"%3E読み込み中...%3C/text%3E%3C/svg%3E`;
    }
  }

  /**
   * エラー時のフォールバック画像を生成
   */
  function generateFallback(w, h) {
    const defaultWidth = w || 300;
    const defaultHeight = h || 200;
    const bgColor = '#f8d7da';
    const textColor = '#721c24';
    
    return fallback || `data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 ${defaultWidth} ${defaultHeight}"%3E%3Crect width="100%25" height="100%25" fill="${bgColor.replace('#', '%23')}"/%3E%3Ctext x="50%25" y="50%25" text-anchor="middle" dy=".3em" fill="${textColor.replace('#', '%23')}" font-family="system-ui,sans-serif" font-size="14"%3E画像を読み込めません%3C/text%3E%3C/svg%3E`;
  }

  /**
   * 画像読み込み処理
   */
  function handleImageLoad() {
    isLoaded = true;
    isLoading = false;
    isError = false;
    
    // パフォーマンスメトリクスを記録
    performanceMonitor.recordMetric('lazy-image-load-success', {
      type: 'component',
      src: currentSrc,
      timestamp: Date.now()
    });
    
    // コールバック実行
    if (onLoad) {
      onLoad({ src: currentSrc, element: imgElement });
    }
  }

  /**
   * 画像エラー処理
   */
  function handleImageError(event) {
    isError = true;
    isLoading = false;
    isLoaded = false;
    
    // フォールバック画像を設定
    if (imgElement) {
      imgElement.src = generateFallback(width, height);
    }
    
    // パフォーマンスメトリクスを記録
    performanceMonitor.recordMetric('lazy-image-load-error', {
      type: 'component-error',
      src: currentSrc,
      error: event.message || 'Image load failed',
      timestamp: Date.now()
    });
    
    // コールバック実行
    if (onError) {
      onError({ src: currentSrc, error: event, element: imgElement });
    }
  }

  /**
   * 画像読み込み開始処理
   */
  function handleImageLoadStart() {
    isLoading = true;
    isLoaded = false;
    isError = false;
  }

  /**
   * コンポーネントマウント時の処理
   */
  onMount(() => {
    if (!imgElement || !src) return;
    
    currentSrc = src;
    
    if (loading === 'eager' || priority) {
      // 即座に読み込み
      isLoading = true;
      imgElement.src = src;
    } else {
      // 遅延読み込みを設定
      setupLazyImage(imgElement, src, alt, {
        placeholder: placeholderSrc
      });
    }
  });

  /**
   * src が変更された時の処理
   */
  $: if (imgElement && src && src !== currentSrc) {
    currentSrc = src;
    isLoaded = false;
    isError = false;
    
    if (loading === 'eager' || priority) {
      isLoading = true;
      imgElement.src = src;
    } else {
      setupLazyImage(imgElement, src, alt, {
        placeholder: placeholderSrc
      });
    }
  }

  /**
   * コンポーネント破棄時の処理
   */
  onDestroy(() => {
    // 必要に応じてクリーンアップ処理
  });
</script>

<img
  bind:this={imgElement}
  {alt}
  {sizes}
  {srcset}
  class={imgClasses}
  style={imgStyles}
  src={placeholderSrc}
  loading={loading === 'eager' ? 'eager' : 'lazy'}
  decoding="async"
  on:load={handleImageLoad}
  on:error={handleImageError}
  on:loadstart={handleImageLoadStart}
  {...$$restProps}
/>

<style>
  .lazy-image {
    display: block;
    max-width: 100%;
    height: auto;
    transition: opacity 0.3s ease;
  }

  .lazy-image--fade {
    opacity: 0;
  }

  .lazy-image--loading {
    opacity: 0.7;
    background-color: #f8f9fa;
  }

  .lazy-image--loaded.lazy-image--fade {
    opacity: 1;
  }

  .lazy-image--error {
    opacity: 1;
    background-color: #f8d7da;
  }

  /* ローディング状態のアニメーション */
  .lazy-image--loading::after {
    content: '';
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background: linear-gradient(
      90deg,
      transparent,
      rgba(255, 255, 255, 0.4),
      transparent
    );
    animation: shimmer 1.5s infinite;
  }

  @keyframes shimmer {
    0% {
      transform: translateX(-100%);
    }
    100% {
      transform: translateX(100%);
    }
  }

  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .lazy-image {
      width: 100%;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .lazy-image {
      transition: none;
    }
    
    .lazy-image--loading::after {
      animation: none;
    }
  }

  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .lazy-image--loading {
      background-color: #000;
    }
    
    .lazy-image--error {
      background-color: #ff0000;
      border: 2px solid #000;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .lazy-image--loading {
      background-color: #343a40;
    }
    
    .lazy-image--error {
      background-color: #721c24;
    }
  }
</style>
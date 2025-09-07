<!--
  最適化された画像コンポーネント
  遅延読み込み、WebP対応、レスポンシブ画像をサポート
-->
<script>
  import { onMount, onDestroy } from 'svelte';
  import { lazyImageLoader, generateSrcSet, getOptimizedImageUrl } from '$lib/utils/imageOptimization.js';

  // Props
  export let src = '';
  export let alt = '';
  export let width = null;
  export let height = null;
  export let sizes = '100vw';
  export let lazy = true;
  export let quality = 85;
  export let placeholder = null;
  export let className = '';
  export let style = '';
  export let priority = false; // 重要な画像の場合はtrue
  export let responsive = true; // レスポンシブ画像を使用するか
  export let aspectRatio = null; // アスペクト比 (例: '16/9')

  // 内部状態
  let imgElement;
  let loaded = false;
  let error = false;
  let optimizedSrc = src;
  let srcset = '';

  // プレースホルダー画像の生成
  function generatePlaceholder(w = width || 400, h = height || 300) {
    if (placeholder) return placeholder;
    
    // SVGプレースホルダーを生成
    const svg = `
      <svg width="${w}" height="${h}" xmlns="http://www.w3.org/2000/svg">
        <rect width="100%" height="100%" fill="#f0f0f0"/>
        <text x="50%" y="50%" text-anchor="middle" dy=".3em" fill="#999" font-family="Arial, sans-serif" font-size="14">
          読み込み中...
        </text>
      </svg>
    `;
    return `data:image/svg+xml;base64,${btoa(svg)}`;
  }

  // 画像の最適化とsrcset生成
  async function setupImage() {
    if (!src) return;

    try {
      // 最適化されたURLを生成
      optimizedSrc = await getOptimizedImageUrl(src, {
        width,
        height,
        quality
      });

      // レスポンシブ画像のsrcsetを生成
      if (responsive && !srcset) {
        srcset = generateSrcSet(src);
      }
    } catch (error) {
      console.warn('Failed to optimize image:', error);
      optimizedSrc = src;
    }
  }

  // 画像の読み込み処理
  function handleImageLoad() {
    loaded = true;
    error = false;
  }

  function handleImageError() {
    error = true;
    loaded = false;
    console.warn(`Failed to load image: ${src}`);
  }

  // 遅延読み込みの設定
  function setupLazyLoading() {
    if (!lazy || priority) return;
    
    if (imgElement) {
      lazyImageLoader.observe(imgElement);
    }
  }

  // 重要な画像の即座読み込み
  function loadImmediately() {
    if (imgElement && (priority || !lazy)) {
      imgElement.src = optimizedSrc;
      if (srcset) {
        imgElement.srcset = srcset;
      }
    }
  }

  onMount(async () => {
    await setupImage();
    
    if (priority || !lazy) {
      loadImmediately();
    } else {
      setupLazyLoading();
    }
  });

  onDestroy(() => {
    if (imgElement && lazy && !priority) {
      lazyImageLoader.unobserve(imgElement);
    }
  });

  // リアクティブな更新
  $: if (imgElement && src) {
    setupImage().then(() => {
      if (priority || !lazy) {
        loadImmediately();
      } else {
        setupLazyLoading();
      }
    });
  }

  // アスペクト比のスタイル計算
  $: aspectRatioStyle = aspectRatio ? `aspect-ratio: ${aspectRatio};` : '';
  $: containerStyle = `${aspectRatioStyle} ${style}`;
</script>

<div 
  class="optimized-image-container {className}"
  style={containerStyle}
  class:loading={!loaded && !error}
  class:loaded
  class:error
>
  <!-- プレースホルダー -->
  {#if !loaded && !error && (lazy && !priority)}
    <img
      class="placeholder"
      src={generatePlaceholder()}
      alt=""
      aria-hidden="true"
    />
  {/if}

  <!-- メイン画像 -->
  <img
    bind:this={imgElement}
    class="main-image"
    class:lazy
    class:priority
    {alt}
    {width}
    {height}
    {sizes}
    data-src={lazy && !priority ? optimizedSrc : undefined}
    data-srcset={lazy && !priority && srcset ? srcset : undefined}
    src={!lazy || priority ? optimizedSrc : undefined}
    srcset={!lazy || priority && srcset ? srcset : undefined}
    loading={priority ? 'eager' : 'lazy'}
    decoding="async"
    on:load={handleImageLoad}
    on:error={handleImageError}
  />

  <!-- エラー時のフォールバック -->
  {#if error}
    <div class="error-placeholder">
      <svg width="48" height="48" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M21 19V5C21 3.9 20.1 3 19 3H5C3.9 3 3 3.9 3 5V19C3 20.1 3.9 21 5 21H19C20.1 21 21 20.1 21 19ZM8.5 13.5L11 16.51L14.5 12L19 18H5L8.5 13.5Z" fill="currentColor"/>
      </svg>
      <span>画像を読み込めませんでした</span>
    </div>
  {/if}

  <!-- 読み込み中のスピナー -->
  {#if !loaded && !error && (!lazy || priority)}
    <div class="loading-spinner">
      <div class="spinner"></div>
    </div>
  {/if}
</div>

<style>
  .optimized-image-container {
    position: relative;
    display: inline-block;
    overflow: hidden;
    background-color: #f5f5f5;
  }

  .main-image {
    width: 100%;
    height: auto;
    display: block;
    transition: opacity 0.3s ease-in-out;
  }

  .main-image.lazy:not(.loaded) {
    opacity: 0;
  }

  .main-image.loaded {
    opacity: 1;
  }

  .placeholder {
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    opacity: 0.7;
    z-index: 1;
  }

  .loading-spinner {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    z-index: 2;
  }

  .spinner {
    width: 24px;
    height: 24px;
    border: 2px solid #e0e0e0;
    border-top: 2px solid #007bff;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  .error-placeholder {
    position: absolute;
    top: 50%;
    left: 50%;
    transform: translate(-50%, -50%);
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    color: #666;
    font-size: 14px;
    text-align: center;
    z-index: 2;
  }

  .error-placeholder svg {
    opacity: 0.5;
  }

  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .optimized-image-container {
      width: 100%;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .main-image,
    .spinner {
      transition: none;
      animation: none;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .optimized-image-container {
      background-color: #2a2a2a;
    }

    .error-placeholder {
      color: #ccc;
    }

    .spinner {
      border-color: #444;
      border-top-color: #007bff;
    }
  }
</style>
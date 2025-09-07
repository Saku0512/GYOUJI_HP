/**
 * 画像最適化ユーティリティ
 * 遅延読み込み、WebP対応、レスポンシブ画像などの機能を提供
 */

/**
 * WebP対応チェック
 */
let webpSupported = null;

export function checkWebPSupport() {
  if (webpSupported !== null) {
    return Promise.resolve(webpSupported);
  }

  return new Promise((resolve) => {
    const webP = new Image();
    webP.onload = webP.onerror = () => {
      webpSupported = webP.height === 2;
      resolve(webpSupported);
    };
    webP.src = 'data:image/webp;base64,UklGRjoAAABXRUJQVlA4IC4AAACyAgCdASoCAAIALmk0mk0iIiIiIgBoSygABc6WWgAA/veff/0PP8bA//LwYAAA';
  });
}

/**
 * 画像の遅延読み込み用Intersection Observer
 */
class LazyImageLoader {
  constructor(options = {}) {
    this.options = {
      rootMargin: '50px 0px',
      threshold: 0.01,
      ...options
    };
    
    this.observer = null;
    this.images = new Set();
    this.init();
  }

  init() {
    if ('IntersectionObserver' in window) {
      this.observer = new IntersectionObserver(
        this.handleIntersection.bind(this),
        this.options
      );
    }
  }

  handleIntersection(entries) {
    entries.forEach(entry => {
      if (entry.isIntersecting) {
        const img = entry.target;
        this.loadImage(img);
        this.observer.unobserve(img);
        this.images.delete(img);
      }
    });
  }

  async loadImage(img) {
    const src = img.dataset.src;
    const srcset = img.dataset.srcset;
    
    if (!src && !srcset) return;

    // WebP対応チェック
    const supportsWebP = await checkWebPSupport();
    
    // WebP版があるかチェック
    let finalSrc = src;
    if (supportsWebP && src && !src.includes('.webp')) {
      const webpSrc = src.replace(/\.(jpg|jpeg|png)$/i, '.webp');
      if (await this.imageExists(webpSrc)) {
        finalSrc = webpSrc;
      }
    }

    // 画像の読み込み
    const tempImg = new Image();
    
    return new Promise((resolve, reject) => {
      tempImg.onload = () => {
        // フェードイン効果
        img.style.opacity = '0';
        img.style.transition = 'opacity 0.3s ease-in-out';
        
        if (srcset) img.srcset = srcset;
        if (finalSrc) img.src = finalSrc;
        
        // 読み込み完了後にフェードイン
        requestAnimationFrame(() => {
          img.style.opacity = '1';
          img.classList.add('loaded');
          img.classList.remove('loading');
        });
        
        resolve();
      };
      
      tempImg.onerror = () => {
        // フォールバック画像
        if (finalSrc !== src && src) {
          img.src = src;
        } else {
          img.classList.add('error');
        }
        reject(new Error(`Failed to load image: ${finalSrc}`));
      };
      
      tempImg.src = finalSrc;
    });
  }

  async imageExists(url) {
    try {
      const response = await fetch(url, { method: 'HEAD' });
      return response.ok;
    } catch {
      return false;
    }
  }

  observe(img) {
    if (!this.observer) {
      // Intersection Observer非対応の場合は即座に読み込み
      this.loadImage(img);
      return;
    }

    this.images.add(img);
    this.observer.observe(img);
    
    // 読み込み中のスタイルを適用
    img.classList.add('loading');
  }

  unobserve(img) {
    if (this.observer) {
      this.observer.unobserve(img);
    }
    this.images.delete(img);
  }

  disconnect() {
    if (this.observer) {
      this.observer.disconnect();
    }
    this.images.clear();
  }
}

// グローバルインスタンス
export const lazyImageLoader = new LazyImageLoader();

/**
 * レスポンシブ画像のsrcset生成
 */
export function generateSrcSet(basePath, sizes = [320, 640, 960, 1280, 1920]) {
  const extension = basePath.split('.').pop();
  const pathWithoutExt = basePath.replace(`.${extension}`, '');
  
  return sizes
    .map(size => `${pathWithoutExt}-${size}w.${extension} ${size}w`)
    .join(', ');
}

/**
 * 画像の最適化されたURLを生成
 */
export async function getOptimizedImageUrl(originalUrl, options = {}) {
  const {
    width,
    height,
    quality = 85,
    format = 'auto'
  } = options;

  // WebP対応チェック
  const supportsWebP = await checkWebPSupport();
  
  // URLパラメータを構築
  const params = new URLSearchParams();
  
  if (width) params.set('w', width);
  if (height) params.set('h', height);
  if (quality !== 85) params.set('q', quality);
  
  // フォーマット決定
  let targetFormat = format;
  if (format === 'auto') {
    targetFormat = supportsWebP ? 'webp' : 'jpg';
  }
  
  if (targetFormat !== 'auto') {
    params.set('f', targetFormat);
  }

  // 最適化されたURLを返す
  const separator = originalUrl.includes('?') ? '&' : '?';
  return `${originalUrl}${separator}${params.toString()}`;
}

/**
 * 画像のプリロード
 */
export function preloadImage(src, options = {}) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    
    if (options.crossOrigin) {
      img.crossOrigin = options.crossOrigin;
    }
    
    img.onload = () => resolve(img);
    img.onerror = reject;
    img.src = src;
  });
}

/**
 * 重要な画像のプリロード
 */
export async function preloadCriticalImages(imageUrls) {
  const promises = imageUrls.map(url => preloadImage(url));
  
  try {
    await Promise.all(promises);
    console.log('Critical images preloaded successfully');
  } catch (error) {
    console.warn('Some critical images failed to preload:', error);
  }
}

/**
 * 画像の圧縮（Canvas使用）
 */
export function compressImage(file, options = {}) {
  const {
    maxWidth = 1920,
    maxHeight = 1080,
    quality = 0.8,
    outputFormat = 'image/jpeg'
  } = options;

  return new Promise((resolve, reject) => {
    const canvas = document.createElement('canvas');
    const ctx = canvas.getContext('2d');
    const img = new Image();

    img.onload = () => {
      // アスペクト比を保持してリサイズ
      let { width, height } = img;
      
      if (width > maxWidth) {
        height = (height * maxWidth) / width;
        width = maxWidth;
      }
      
      if (height > maxHeight) {
        width = (width * maxHeight) / height;
        height = maxHeight;
      }

      canvas.width = width;
      canvas.height = height;

      // 画像を描画
      ctx.drawImage(img, 0, 0, width, height);

      // Blobとして出力
      canvas.toBlob(resolve, outputFormat, quality);
    };

    img.onerror = reject;
    img.src = URL.createObjectURL(file);
  });
}

/**
 * 画像のメタデータ取得
 */
export function getImageMetadata(file) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    
    img.onload = () => {
      resolve({
        width: img.naturalWidth,
        height: img.naturalHeight,
        aspectRatio: img.naturalWidth / img.naturalHeight,
        size: file.size,
        type: file.type,
        name: file.name
      });
    };
    
    img.onerror = reject;
    img.src = URL.createObjectURL(file);
  });
}

/**
 * 画像フォーマット変換
 */
export function convertImageFormat(imageUrl, targetFormat = 'webp', quality = 0.8) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.crossOrigin = 'anonymous';
    
    img.onload = () => {
      const canvas = document.createElement('canvas');
      const ctx = canvas.getContext('2d');
      
      canvas.width = img.width;
      canvas.height = img.height;
      
      ctx.drawImage(img, 0, 0);
      
      canvas.toBlob(resolve, `image/${targetFormat}`, quality);
    };
    
    img.onerror = reject;
    img.src = imageUrl;
  });
}
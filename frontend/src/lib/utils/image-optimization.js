/**
 * 画像最適化とアセット管理ユーティリティ
 * 遅延読み込み、最適化、キャッシュ戦略を提供
 */

import { performanceMonitor } from './performance.js';

/**
 * 画像最適化クラス
 */
export class ImageOptimizer {
  constructor() {
    this.loadedImages = new Set();
    this.imageCache = new Map();
    this.intersectionObserver = null;
    this.loadingImages = new Map();
    
    // Intersection Observer の初期化
    this.initIntersectionObserver();
  }

  /**
   * Intersection Observer の初期化
   */
  initIntersectionObserver() {
    if (typeof window === 'undefined' || !('IntersectionObserver' in window)) {
      return;
    }

    this.intersectionObserver = new IntersectionObserver(
      (entries) => {
        entries.forEach(entry => {
          if (entry.isIntersecting) {
            const img = entry.target;
            this.loadImage(img);
            this.intersectionObserver.unobserve(img);
          }
        });
      },
      {
        rootMargin: '50px 0px', // 50px手前で読み込み開始
        threshold: 0.01
      }
    );
  }

  /**
   * 遅延読み込み用の画像を登録
   * @param {HTMLImageElement} img - 画像要素
   * @param {string} src - 画像URL
   * @param {string} alt - 代替テキスト
   * @param {Object} options - オプション
   */
  registerLazyImage(img, src, alt = '', options = {}) {
    if (!img || !src) return;

    // データ属性に元のsrcを保存
    img.dataset.src = src;
    img.alt = alt;
    
    // プレースホルダー画像を設定
    if (options.placeholder) {
      img.src = options.placeholder;
    } else {
      // デフォルトのプレースホルダー（1x1の透明画像）
      img.src = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"%3E%3C/svg%3E';
    }
    
    // ローディング状態のクラスを追加
    img.classList.add('lazy-loading');
    
    // Intersection Observer に登録
    if (this.intersectionObserver) {
      this.intersectionObserver.observe(img);
    } else {
      // フォールバック: すぐに読み込み
      this.loadImage(img);
    }
  }

  /**
   * 画像を読み込み
   * @param {HTMLImageElement} img - 画像要素
   */
  async loadImage(img) {
    const src = img.dataset.src;
    if (!src || this.loadingImages.has(src)) {
      return;
    }

    this.loadingImages.set(src, true);
    const startTime = performance.now();

    try {
      // キャッシュから取得を試行
      let imageBlob = this.imageCache.get(src);
      
      if (!imageBlob) {
        // 新しい画像を読み込み
        const response = await fetch(src);
        if (!response.ok) {
          throw new Error(`Failed to load image: ${response.status}`);
        }
        
        imageBlob = await response.blob();
        
        // キャッシュに保存（サイズ制限あり）
        if (this.imageCache.size < 50) { // 最大50枚まで
          this.imageCache.set(src, imageBlob);
        }
      }

      // Blob URLを作成
      const blobUrl = URL.createObjectURL(imageBlob);
      
      // 画像の読み込み完了を待つ
      await new Promise((resolve, reject) => {
        const tempImg = new Image();
        tempImg.onload = () => {
          img.src = blobUrl;
          img.classList.remove('lazy-loading');
          img.classList.add('lazy-loaded');
          
          // パフォーマンスメトリクスを記録
          const loadTime = performance.now() - startTime;
          performanceMonitor.recordMetric('image-lazy-load', {
            type: 'asset',
            src,
            loadTime,
            fromCache: this.imageCache.has(src),
            timestamp: Date.now()
          });
          
          resolve();
        };
        
        tempImg.onerror = () => {
          img.classList.remove('lazy-loading');
          img.classList.add('lazy-error');
          
          // エラーメトリクスを記録
          performanceMonitor.recordMetric('image-lazy-load-error', {
            type: 'asset-error',
            src,
            loadTime: performance.now() - startTime,
            timestamp: Date.now()
          });
          
          reject(new Error(`Failed to load image: ${src}`));
        };
        
        tempImg.src = blobUrl;
      });

      this.loadedImages.add(src);
      
    } catch (error) {
      console.error('Image loading error:', error);
      
      // エラー時のフォールバック
      img.classList.remove('lazy-loading');
      img.classList.add('lazy-error');
      
      // エラー画像を表示
      img.src = 'data:image/svg+xml,%3Csvg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"%3E%3Crect width="100" height="100" fill="%23f8f9fa"/%3E%3Ctext x="50" y="50" text-anchor="middle" dy=".3em" fill="%236c757d"%3E画像エラー%3C/text%3E%3C/svg%3E';
      
    } finally {
      this.loadingImages.delete(src);
    }
  }

  /**
   * 画像のプリロード
   * @param {string[]} urls - プリロードする画像URL配列
   */
  async preloadImages(urls) {
    const promises = urls.map(url => this.preloadSingleImage(url));
    return Promise.allSettled(promises);
  }

  /**
   * 単一画像のプリロード
   * @param {string} url - 画像URL
   */
  preloadSingleImage(url) {
    return new Promise((resolve, reject) => {
      const img = new Image();
      img.onload = () => resolve(url);
      img.onerror = () => reject(new Error(`Failed to preload: ${url}`));
      img.src = url;
    });
  }

  /**
   * 画像キャッシュをクリア
   */
  clearImageCache() {
    // Blob URLsを解放
    for (const [url, blob] of this.imageCache) {
      if (blob instanceof Blob) {
        URL.revokeObjectURL(URL.createObjectURL(blob));
      }
    }
    
    this.imageCache.clear();
    this.loadedImages.clear();
  }

  /**
   * リソースのクリーンアップ
   */
  destroy() {
    if (this.intersectionObserver) {
      this.intersectionObserver.disconnect();
    }
    
    this.clearImageCache();
  }
}

/**
 * 静的アセット管理クラス
 */
export class AssetManager {
  constructor() {
    this.assetCache = new Map();
    this.cacheStrategies = new Map();
    
    // デフォルトのキャッシュ戦略を設定
    this.setupDefaultCacheStrategies();
  }

  /**
   * デフォルトのキャッシュ戦略を設定
   */
  setupDefaultCacheStrategies() {
    // 画像ファイル: 長期キャッシュ
    this.cacheStrategies.set(/\.(jpg|jpeg|png|gif|webp|svg)$/i, {
      maxAge: 7 * 24 * 60 * 60 * 1000, // 7日
      strategy: 'cache-first'
    });
    
    // CSSファイル: 中期キャッシュ
    this.cacheStrategies.set(/\.css$/i, {
      maxAge: 24 * 60 * 60 * 1000, // 1日
      strategy: 'stale-while-revalidate'
    });
    
    // JSファイル: 中期キャッシュ
    this.cacheStrategies.set(/\.js$/i, {
      maxAge: 24 * 60 * 60 * 1000, // 1日
      strategy: 'stale-while-revalidate'
    });
    
    // フォントファイル: 長期キャッシュ
    this.cacheStrategies.set(/\.(woff|woff2|ttf|eot)$/i, {
      maxAge: 30 * 24 * 60 * 60 * 1000, // 30日
      strategy: 'cache-first'
    });
  }

  /**
   * アセットを取得（キャッシュ戦略に基づく）
   * @param {string} url - アセットURL
   * @param {Object} options - オプション
   */
  async fetchAsset(url, options = {}) {
    const strategy = this.getCacheStrategy(url);
    const cacheKey = this.getCacheKey(url);
    
    switch (strategy.strategy) {
      case 'cache-first':
        return this.cacheFirstStrategy(url, cacheKey, strategy, options);
      
      case 'network-first':
        return this.networkFirstStrategy(url, cacheKey, strategy, options);
      
      case 'stale-while-revalidate':
        return this.staleWhileRevalidateStrategy(url, cacheKey, strategy, options);
      
      default:
        return this.networkOnlyStrategy(url, options);
    }
  }

  /**
   * Cache First戦略
   */
  async cacheFirstStrategy(url, cacheKey, strategy, options) {
    // キャッシュから取得を試行
    const cached = this.getCachedAsset(cacheKey);
    if (cached && !this.isCacheExpired(cached, strategy.maxAge)) {
      return cached.data;
    }
    
    // ネットワークから取得
    try {
      const response = await fetch(url, options);
      if (response.ok) {
        const data = await response.blob();
        this.setCachedAsset(cacheKey, data);
        return data;
      }
    } catch (error) {
      // ネットワークエラー時は期限切れキャッシュでも返す
      if (cached) {
        return cached.data;
      }
      throw error;
    }
  }

  /**
   * Network First戦略
   */
  async networkFirstStrategy(url, cacheKey, strategy, options) {
    try {
      const response = await fetch(url, options);
      if (response.ok) {
        const data = await response.blob();
        this.setCachedAsset(cacheKey, data);
        return data;
      }
    } catch (error) {
      // ネットワークエラー時はキャッシュから取得
      const cached = this.getCachedAsset(cacheKey);
      if (cached) {
        return cached.data;
      }
      throw error;
    }
  }

  /**
   * Stale While Revalidate戦略
   */
  async staleWhileRevalidateStrategy(url, cacheKey, strategy, options) {
    const cached = this.getCachedAsset(cacheKey);
    
    // キャッシュがある場合は即座に返す
    if (cached) {
      // バックグラウンドで更新
      if (this.isCacheExpired(cached, strategy.maxAge)) {
        this.updateAssetInBackground(url, cacheKey, options);
      }
      return cached.data;
    }
    
    // キャッシュがない場合はネットワークから取得
    const response = await fetch(url, options);
    if (response.ok) {
      const data = await response.blob();
      this.setCachedAsset(cacheKey, data);
      return data;
    }
    
    throw new Error(`Failed to fetch asset: ${url}`);
  }

  /**
   * Network Only戦略
   */
  async networkOnlyStrategy(url, options) {
    const response = await fetch(url, options);
    if (response.ok) {
      return response.blob();
    }
    throw new Error(`Failed to fetch asset: ${url}`);
  }

  /**
   * バックグラウンドでアセットを更新
   */
  async updateAssetInBackground(url, cacheKey, options) {
    try {
      const response = await fetch(url, options);
      if (response.ok) {
        const data = await response.blob();
        this.setCachedAsset(cacheKey, data);
      }
    } catch (error) {
      console.warn('Background asset update failed:', error);
    }
  }

  /**
   * URLに対するキャッシュ戦略を取得
   */
  getCacheStrategy(url) {
    for (const [pattern, strategy] of this.cacheStrategies) {
      if (pattern.test(url)) {
        return strategy;
      }
    }
    
    // デフォルト戦略
    return {
      maxAge: 60 * 60 * 1000, // 1時間
      strategy: 'network-first'
    };
  }

  /**
   * キャッシュキーを生成
   */
  getCacheKey(url) {
    return `asset_${btoa(url).replace(/[^a-zA-Z0-9]/g, '')}`;
  }

  /**
   * キャッシュからアセットを取得
   */
  getCachedAsset(cacheKey) {
    return this.assetCache.get(cacheKey);
  }

  /**
   * アセットをキャッシュに保存
   */
  setCachedAsset(cacheKey, data) {
    this.assetCache.set(cacheKey, {
      data,
      timestamp: Date.now()
    });
    
    // キャッシュサイズ制限
    if (this.assetCache.size > 100) {
      const oldestKey = this.assetCache.keys().next().value;
      this.assetCache.delete(oldestKey);
    }
  }

  /**
   * キャッシュが期限切れかチェック
   */
  isCacheExpired(cached, maxAge) {
    return Date.now() - cached.timestamp > maxAge;
  }

  /**
   * キャッシュをクリア
   */
  clearCache() {
    this.assetCache.clear();
  }
}

// グローバルインスタンス
export const imageOptimizer = new ImageOptimizer();
export const assetManager = new AssetManager();

/**
 * 遅延読み込み画像コンポーネント用のヘルパー関数
 */
export function setupLazyImage(imgElement, src, alt = '', options = {}) {
  if (typeof window === 'undefined') return;
  
  imageOptimizer.registerLazyImage(imgElement, src, alt, options);
}

/**
 * 画像のプリロード
 */
export function preloadImages(urls) {
  return imageOptimizer.preloadImages(urls);
}

/**
 * アセットの取得
 */
export function fetchAsset(url, options = {}) {
  return assetManager.fetchAsset(url, options);
}

/**
 * クリーンアップ関数
 */
export function cleanup() {
  imageOptimizer.destroy();
  assetManager.clearCache();
}
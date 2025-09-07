/**
 * アセットキャッシュ戦略
 * 静的アセットの効率的なキャッシュ管理を提供
 */

/**
 * キャッシュ設定
 */
const CACHE_CONFIG = {
  // 静的アセット（長期キャッシュ）
  static: {
    name: 'tournament-static-v1',
    maxAge: 7 * 24 * 60 * 60 * 1000, // 7日
    maxEntries: 100,
    patterns: [
      /\.(js|css|woff|woff2|ttf|eot|ico)$/,
      /\.(png|jpg|jpeg|gif|svg|webp)$/,
      /\/static\//
    ]
  },
  
  // 動的コンテンツ（短期キャッシュ）
  dynamic: {
    name: 'tournament-dynamic-v1',
    maxAge: 60 * 60 * 1000, // 1時間
    maxEntries: 50,
    patterns: [
      /\/api\/tournaments/,
      /\/api\/matches/
    ]
  },
  
  // ページキャッシュ
  pages: {
    name: 'tournament-pages-v1',
    maxAge: 24 * 60 * 60 * 1000, // 24時間
    maxEntries: 20,
    patterns: [
      /^\/$/, // ホームページ
      /^\/admin/,
      /^\/login/
    ]
  }
};

/**
 * メモリキャッシュクラス
 */
class MemoryCache {
  constructor(maxSize = 50) {
    this.cache = new Map();
    this.maxSize = maxSize;
  }

  get(key) {
    const item = this.cache.get(key);
    if (!item) return null;

    // 有効期限チェック
    if (Date.now() > item.expiry) {
      this.cache.delete(key);
      return null;
    }

    // LRU: アクセスされたアイテムを最後に移動
    this.cache.delete(key);
    this.cache.set(key, item);
    
    return item.data;
  }

  set(key, data, ttl = 5 * 60 * 1000) { // デフォルト5分
    // サイズ制限チェック
    if (this.cache.size >= this.maxSize) {
      // 最も古いアイテムを削除
      const firstKey = this.cache.keys().next().value;
      this.cache.delete(firstKey);
    }

    this.cache.set(key, {
      data,
      expiry: Date.now() + ttl,
      timestamp: Date.now()
    });
  }

  delete(key) {
    return this.cache.delete(key);
  }

  clear() {
    this.cache.clear();
  }

  size() {
    return this.cache.size;
  }

  // 期限切れアイテムのクリーンアップ
  cleanup() {
    const now = Date.now();
    for (const [key, item] of this.cache.entries()) {
      if (now > item.expiry) {
        this.cache.delete(key);
      }
    }
  }
}

/**
 * ブラウザキャッシュマネージャー
 */
class BrowserCacheManager {
  constructor() {
    this.memoryCache = new MemoryCache();
    this.setupCleanupInterval();
  }

  setupCleanupInterval() {
    // 5分ごとにメモリキャッシュをクリーンアップ
    setInterval(() => {
      this.memoryCache.cleanup();
    }, 5 * 60 * 1000);
  }

  /**
   * キャッシュからデータを取得
   */
  async get(key, options = {}) {
    const { useMemoryCache = true, useBrowserCache = true } = options;

    // メモリキャッシュから取得
    if (useMemoryCache) {
      const memoryData = this.memoryCache.get(key);
      if (memoryData) {
        return memoryData;
      }
    }

    // ブラウザキャッシュから取得
    if (useBrowserCache && 'caches' in window) {
      try {
        const cache = await caches.open(this.getCacheName(key));
        const response = await cache.match(key);
        
        if (response) {
          const data = await response.json();
          
          // メモリキャッシュにも保存
          if (useMemoryCache) {
            this.memoryCache.set(key, data);
          }
          
          return data;
        }
      } catch (error) {
        console.warn('Failed to get from browser cache:', error);
      }
    }

    return null;
  }

  /**
   * キャッシュにデータを保存
   */
  async set(key, data, options = {}) {
    const { 
      useMemoryCache = true, 
      useBrowserCache = true,
      ttl = 5 * 60 * 1000 // 5分
    } = options;

    // メモリキャッシュに保存
    if (useMemoryCache) {
      this.memoryCache.set(key, data, ttl);
    }

    // ブラウザキャッシュに保存
    if (useBrowserCache && 'caches' in window) {
      try {
        const cache = await caches.open(this.getCacheName(key));
        const response = new Response(JSON.stringify(data), {
          headers: {
            'Content-Type': 'application/json',
            'Cache-Control': `max-age=${Math.floor(ttl / 1000)}`,
            'X-Cache-Timestamp': Date.now().toString()
          }
        });
        
        await cache.put(key, response);
      } catch (error) {
        console.warn('Failed to set browser cache:', error);
      }
    }
  }

  /**
   * キャッシュからデータを削除
   */
  async delete(key) {
    // メモリキャッシュから削除
    this.memoryCache.delete(key);

    // ブラウザキャッシュから削除
    if ('caches' in window) {
      try {
        const cache = await caches.open(this.getCacheName(key));
        await cache.delete(key);
      } catch (error) {
        console.warn('Failed to delete from browser cache:', error);
      }
    }
  }

  /**
   * 全キャッシュをクリア
   */
  async clear() {
    // メモリキャッシュをクリア
    this.memoryCache.clear();

    // ブラウザキャッシュをクリア
    if ('caches' in window) {
      try {
        const cacheNames = await caches.keys();
        await Promise.all(
          cacheNames
            .filter(name => name.startsWith('tournament-'))
            .map(name => caches.delete(name))
        );
      } catch (error) {
        console.warn('Failed to clear browser cache:', error);
      }
    }
  }

  /**
   * キャッシュ名を取得
   */
  getCacheName(key) {
    // URLパターンに基づいてキャッシュ名を決定
    for (const [type, config] of Object.entries(CACHE_CONFIG)) {
      for (const pattern of config.patterns) {
        if (pattern.test(key)) {
          return config.name;
        }
      }
    }
    
    return CACHE_CONFIG.dynamic.name; // デフォルト
  }

  /**
   * キャッシュ統計を取得
   */
  async getStats() {
    const stats = {
      memory: {
        size: this.memoryCache.size(),
        maxSize: this.memoryCache.maxSize
      },
      browser: {}
    };

    if ('caches' in window) {
      try {
        const cacheNames = await caches.keys();
        
        for (const cacheName of cacheNames) {
          if (cacheName.startsWith('tournament-')) {
            const cache = await caches.open(cacheName);
            const keys = await cache.keys();
            stats.browser[cacheName] = keys.length;
          }
        }
      } catch (error) {
        console.warn('Failed to get cache stats:', error);
      }
    }

    return stats;
  }
}

/**
 * アセットプリローダー
 */
class AssetPreloader {
  constructor() {
    this.preloadQueue = [];
    this.preloading = false;
  }

  /**
   * 重要なアセットをプリロード
   */
  async preloadCriticalAssets(assets) {
    const promises = assets.map(asset => this.preloadAsset(asset));
    
    try {
      await Promise.all(promises);
      console.log('Critical assets preloaded successfully');
    } catch (error) {
      console.warn('Some critical assets failed to preload:', error);
    }
  }

  /**
   * 個別アセットのプリロード
   */
  preloadAsset(asset) {
    return new Promise((resolve, reject) => {
      const { url, type = 'fetch' } = typeof asset === 'string' ? { url: asset } : asset;

      switch (type) {
        case 'image':
          const img = new Image();
          img.onload = () => resolve(img);
          img.onerror = reject;
          img.src = url;
          break;

        case 'font':
          const font = new FontFace(asset.family, `url(${url})`);
          font.load().then(resolve).catch(reject);
          break;

        case 'fetch':
        default:
          fetch(url, { mode: 'no-cors' })
            .then(resolve)
            .catch(reject);
          break;
      }
    });
  }

  /**
   * バックグラウンドでアセットをプリロード
   */
  preloadInBackground(assets) {
    this.preloadQueue.push(...assets);
    
    if (!this.preloading) {
      this.processPreloadQueue();
    }
  }

  async processPreloadQueue() {
    this.preloading = true;

    while (this.preloadQueue.length > 0) {
      const asset = this.preloadQueue.shift();
      
      try {
        await this.preloadAsset(asset);
        
        // ブラウザをブロックしないように少し待機
        await new Promise(resolve => setTimeout(resolve, 10));
      } catch (error) {
        console.warn('Failed to preload asset:', asset, error);
      }
    }

    this.preloading = false;
  }
}

/**
 * リソースヒント生成
 */
export function generateResourceHints(assets) {
  const hints = [];

  assets.forEach(asset => {
    const { url, type, crossorigin } = typeof asset === 'string' ? { url: asset } : asset;

    // DNS prefetch
    if (url.startsWith('http')) {
      const hostname = new URL(url).hostname;
      hints.push(`<link rel="dns-prefetch" href="//${hostname}">`);
    }

    // Preload
    let as = '';
    if (type === 'image') as = 'image';
    else if (type === 'font') as = 'font';
    else if (type === 'script') as = 'script';
    else if (type === 'style') as = 'style';

    if (as) {
      const crossoriginAttr = crossorigin ? ` crossorigin="${crossorigin}"` : '';
      hints.push(`<link rel="preload" href="${url}" as="${as}"${crossoriginAttr}>`);
    }
  });

  return hints.join('\n');
}

// グローバルインスタンス
export const cacheManager = new BrowserCacheManager();
export const assetPreloader = new AssetPreloader();

/**
 * キャッシュ戦略のヘルパー関数
 */
export async function getCachedData(key, fetcher, options = {}) {
  const { ttl = 5 * 60 * 1000, forceRefresh = false } = options;

  // 強制リフレッシュでない場合はキャッシュから取得を試行
  if (!forceRefresh) {
    const cachedData = await cacheManager.get(key);
    if (cachedData) {
      return cachedData;
    }
  }

  // データを取得してキャッシュに保存
  try {
    const data = await fetcher();
    await cacheManager.set(key, data, { ttl });
    return data;
  } catch (error) {
    // エラー時は古いキャッシュがあれば返す
    const cachedData = await cacheManager.get(key);
    if (cachedData) {
      console.warn('Using stale cache due to fetch error:', error);
      return cachedData;
    }
    throw error;
  }
}

/**
 * キャッシュ無効化
 */
export async function invalidateCache(pattern) {
  if (typeof pattern === 'string') {
    await cacheManager.delete(pattern);
  } else if (pattern instanceof RegExp) {
    // パターンマッチングでキャッシュを無効化
    const stats = await cacheManager.getStats();
    
    for (const cacheName of Object.keys(stats.browser)) {
      if ('caches' in window) {
        const cache = await caches.open(cacheName);
        const keys = await cache.keys();
        
        for (const request of keys) {
          if (pattern.test(request.url)) {
            await cache.delete(request);
          }
        }
      }
    }
  }
}
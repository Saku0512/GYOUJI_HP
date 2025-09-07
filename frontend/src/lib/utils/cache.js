// データキャッシュシステム - メモリキャッシュとブラウザキャッシュの実装
import { writable, get } from 'svelte/store';

/**
 * キャッシュエントリの型定義
 */
class CacheEntry {
  constructor(data, options = {}) {
    this.data = data;
    this.timestamp = Date.now();
    this.ttl = options.ttl || 300000; // デフォルト5分
    this.tags = options.tags || [];
    this.priority = options.priority || 1;
    this.accessCount = 0;
    this.lastAccessed = this.timestamp;
  }

  /**
   * キャッシュエントリが有効かどうかをチェック
   */
  isValid() {
    return Date.now() - this.timestamp < this.ttl;
  }

  /**
   * キャッシュエントリにアクセス
   */
  access() {
    this.accessCount++;
    this.lastAccessed = Date.now();
    return this.data;
  }

  /**
   * 残り有効時間を取得（ミリ秒）
   */
  getRemainingTTL() {
    const remaining = this.ttl - (Date.now() - this.timestamp);
    return Math.max(0, remaining);
  }
}

/**
 * メモリキャッシュクラス
 */
export class MemoryCache {
  constructor(options = {}) {
    this.options = {
      maxSize: 100, // 最大エントリ数
      defaultTTL: 300000, // デフォルトTTL（5分）
      cleanupInterval: 60000, // クリーンアップ間隔（1分）
      enableLogging: false,
      ...options
    };

    this.cache = new Map();
    this.stats = {
      hits: 0,
      misses: 0,
      sets: 0,
      deletes: 0,
      evictions: 0
    };

    // 定期的なクリーンアップを開始
    this.startCleanup();
  }

  /**
   * データをキャッシュに保存
   * @param {string} key - キャッシュキー
   * @param {*} data - 保存するデータ
   * @param {Object} options - オプション（ttl, tags, priority）
   */
  set(key, data, options = {}) {
    if (!key) {
      throw new Error('Cache key is required');
    }

    // TTLの設定
    const ttl = options.ttl || this.options.defaultTTL;
    
    // キャッシュエントリを作成
    const entry = new CacheEntry(data, {
      ttl,
      tags: options.tags,
      priority: options.priority
    });

    // サイズ制限チェック
    if (this.cache.size >= this.options.maxSize && !this.cache.has(key)) {
      this.evictLRU();
    }

    this.cache.set(key, entry);
    this.stats.sets++;

    if (this.options.enableLogging) {
      console.log(`[MemoryCache] Set: ${key}, TTL: ${ttl}ms`);
    }

    return true;
  }

  /**
   * キャッシュからデータを取得
   * @param {string} key - キャッシュキー
   * @returns {*} キャッシュされたデータまたはnull
   */
  get(key) {
    if (!key) {
      return null;
    }

    const entry = this.cache.get(key);
    
    if (!entry) {
      this.stats.misses++;
      return null;
    }

    if (!entry.isValid()) {
      this.cache.delete(key);
      this.stats.misses++;
      
      if (this.options.enableLogging) {
        console.log(`[MemoryCache] Expired: ${key}`);
      }
      
      return null;
    }

    this.stats.hits++;
    
    if (this.options.enableLogging) {
      console.log(`[MemoryCache] Hit: ${key}`);
    }

    return entry.access();
  }

  /**
   * キャッシュからエントリを削除
   * @param {string} key - キャッシュキー
   */
  delete(key) {
    const deleted = this.cache.delete(key);
    if (deleted) {
      this.stats.deletes++;
      
      if (this.options.enableLogging) {
        console.log(`[MemoryCache] Deleted: ${key}`);
      }
    }
    return deleted;
  }

  /**
   * タグに基づいてキャッシュエントリを無効化
   * @param {string|Array} tags - 無効化するタグ
   */
  invalidateByTag(tags) {
    const tagsArray = Array.isArray(tags) ? tags : [tags];
    let invalidated = 0;

    for (const [key, entry] of this.cache.entries()) {
      if (entry.tags && entry.tags.some(tag => tagsArray.includes(tag))) {
        this.cache.delete(key);
        invalidated++;
      }
    }

    if (this.options.enableLogging && invalidated > 0) {
      console.log(`[MemoryCache] Invalidated ${invalidated} entries by tags: ${tagsArray.join(', ')}`);
    }

    return invalidated;
  }

  /**
   * キャッシュをクリア
   */
  clear() {
    const size = this.cache.size;
    this.cache.clear();
    
    if (this.options.enableLogging) {
      console.log(`[MemoryCache] Cleared ${size} entries`);
    }
    
    return size;
  }

  /**
   * LRU（Least Recently Used）アルゴリズムでエントリを削除
   */
  evictLRU() {
    let oldestKey = null;
    let oldestTime = Date.now();

    for (const [key, entry] of this.cache.entries()) {
      if (entry.lastAccessed <= oldestTime) {
        oldestTime = entry.lastAccessed;
        oldestKey = key;
      }
    }

    if (oldestKey) {
      this.cache.delete(oldestKey);
      this.stats.evictions++;
      
      if (this.options.enableLogging) {
        console.log(`[MemoryCache] Evicted LRU: ${oldestKey}`);
      }
    }
  }

  /**
   * 期限切れエントリのクリーンアップ
   */
  cleanup() {
    let cleaned = 0;
    
    for (const [key, entry] of this.cache.entries()) {
      if (!entry.isValid()) {
        this.cache.delete(key);
        cleaned++;
      }
    }

    if (this.options.enableLogging && cleaned > 0) {
      console.log(`[MemoryCache] Cleaned up ${cleaned} expired entries`);
    }

    return cleaned;
  }

  /**
   * 定期的なクリーンアップを開始
   */
  startCleanup() {
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
    }

    this.cleanupInterval = setInterval(() => {
      this.cleanup();
    }, this.options.cleanupInterval);
  }

  /**
   * クリーンアップを停止
   */
  stopCleanup() {
    if (this.cleanupInterval) {
      clearInterval(this.cleanupInterval);
      this.cleanupInterval = null;
    }
  }

  /**
   * キャッシュの統計情報を取得
   */
  getStats() {
    const hitRate = this.stats.hits + this.stats.misses > 0 
      ? (this.stats.hits / (this.stats.hits + this.stats.misses) * 100).toFixed(2)
      : 0;

    return {
      ...this.stats,
      size: this.cache.size,
      maxSize: this.options.maxSize,
      hitRate: `${hitRate}%`
    };
  }

  /**
   * キャッシュの詳細情報を取得
   */
  getInfo() {
    const entries = [];
    
    for (const [key, entry] of this.cache.entries()) {
      entries.push({
        key,
        size: JSON.stringify(entry.data).length,
        ttl: entry.ttl,
        remainingTTL: entry.getRemainingTTL(),
        accessCount: entry.accessCount,
        lastAccessed: entry.lastAccessed,
        tags: entry.tags,
        priority: entry.priority,
        isValid: entry.isValid()
      });
    }

    return {
      entries,
      stats: this.getStats(),
      options: this.options
    };
  }

  /**
   * リソースのクリーンアップ
   */
  destroy() {
    this.stopCleanup();
    this.clear();
  }
}

/**
 * ブラウザキャッシュクラス（LocalStorage/SessionStorage）
 */
export class BrowserCache {
  constructor(options = {}) {
    this.options = {
      storage: 'localStorage', // 'localStorage' または 'sessionStorage'
      prefix: 'cache_',
      defaultTTL: 3600000, // デフォルトTTL（1時間）
      enableLogging: false,
      ...options
    };

    this.storage = this.options.storage === 'sessionStorage' 
      ? sessionStorage 
      : localStorage;
  }

  /**
   * キーにプレフィックスを追加
   */
  getStorageKey(key) {
    return `${this.options.prefix}${key}`;
  }

  /**
   * データをブラウザキャッシュに保存
   * @param {string} key - キャッシュキー
   * @param {*} data - 保存するデータ
   * @param {Object} options - オプション（ttl）
   */
  set(key, data, options = {}) {
    if (!key) {
      throw new Error('Cache key is required');
    }

    try {
      const ttl = options.ttl || this.options.defaultTTL;
      const entry = {
        data,
        timestamp: Date.now(),
        ttl
      };

      const storageKey = this.getStorageKey(key);
      this.storage.setItem(storageKey, JSON.stringify(entry));

      if (this.options.enableLogging) {
        console.log(`[BrowserCache] Set: ${key}, TTL: ${ttl}ms`);
      }

      return true;
    } catch (error) {
      console.error(`[BrowserCache] Failed to set ${key}:`, error);
      return false;
    }
  }

  /**
   * ブラウザキャッシュからデータを取得
   * @param {string} key - キャッシュキー
   * @returns {*} キャッシュされたデータまたはnull
   */
  get(key) {
    if (!key) {
      return null;
    }

    try {
      const storageKey = this.getStorageKey(key);
      const item = this.storage.getItem(storageKey);
      
      if (!item) {
        return null;
      }

      const entry = JSON.parse(item);
      
      // TTLチェック
      if (Date.now() - entry.timestamp > entry.ttl) {
        this.storage.removeItem(storageKey);
        
        if (this.options.enableLogging) {
          console.log(`[BrowserCache] Expired: ${key}`);
        }
        
        return null;
      }

      if (this.options.enableLogging) {
        console.log(`[BrowserCache] Hit: ${key}`);
      }

      return entry.data;
    } catch (error) {
      console.error(`[BrowserCache] Failed to get ${key}:`, error);
      return null;
    }
  }

  /**
   * ブラウザキャッシュからエントリを削除
   * @param {string} key - キャッシュキー
   */
  delete(key) {
    try {
      const storageKey = this.getStorageKey(key);
      this.storage.removeItem(storageKey);
      
      if (this.options.enableLogging) {
        console.log(`[BrowserCache] Deleted: ${key}`);
      }
      
      return true;
    } catch (error) {
      console.error(`[BrowserCache] Failed to delete ${key}:`, error);
      return false;
    }
  }

  /**
   * プレフィックスに基づいてキャッシュをクリア
   */
  clear() {
    try {
      const keysToRemove = [];
      
      for (let i = 0; i < this.storage.length; i++) {
        const key = this.storage.key(i);
        if (key && key.startsWith(this.options.prefix)) {
          keysToRemove.push(key);
        }
      }

      keysToRemove.forEach(key => {
        this.storage.removeItem(key);
      });

      if (this.options.enableLogging) {
        console.log(`[BrowserCache] Cleared ${keysToRemove.length} entries`);
      }

      return keysToRemove.length;
    } catch (error) {
      console.error('[BrowserCache] Failed to clear cache:', error);
      return 0;
    }
  }

  /**
   * 期限切れエントリのクリーンアップ
   */
  cleanup() {
    try {
      const keysToRemove = [];
      
      for (let i = 0; i < this.storage.length; i++) {
        const key = this.storage.key(i);
        if (key && key.startsWith(this.options.prefix)) {
          try {
            const item = this.storage.getItem(key);
            if (item) {
              const entry = JSON.parse(item);
              if (Date.now() - entry.timestamp > entry.ttl) {
                keysToRemove.push(key);
              }
            }
          } catch (parseError) {
            // 無効なエントリは削除
            keysToRemove.push(key);
          }
        }
      }

      keysToRemove.forEach(key => {
        this.storage.removeItem(key);
      });

      if (this.options.enableLogging && keysToRemove.length > 0) {
        console.log(`[BrowserCache] Cleaned up ${keysToRemove.length} expired entries`);
      }

      return keysToRemove.length;
    } catch (error) {
      console.error('[BrowserCache] Failed to cleanup:', error);
      return 0;
    }
  }
}

/**
 * 統合キャッシュシステムクラス
 */
export class CacheSystem {
  constructor(options = {}) {
    this.options = {
      enableMemoryCache: true,
      enableBrowserCache: true,
      memoryFirst: true, // メモリキャッシュを優先
      ...options
    };

    // メモリキャッシュの初期化
    if (this.options.enableMemoryCache) {
      this.memoryCache = new MemoryCache(options.memory);
    }

    // ブラウザキャッシュの初期化
    if (this.options.enableBrowserCache) {
      this.browserCache = new BrowserCache(options.browser);
    }
  }

  /**
   * データをキャッシュに保存
   * @param {string} key - キャッシュキー
   * @param {*} data - 保存するデータ
   * @param {Object} options - オプション
   */
  async set(key, data, options = {}) {
    const results = {};

    // メモリキャッシュに保存
    if (this.memoryCache) {
      results.memory = this.memoryCache.set(key, data, options);
    }

    // ブラウザキャッシュに保存
    if (this.browserCache && options.persistent !== false) {
      results.browser = this.browserCache.set(key, data, options);
    }

    return results;
  }

  /**
   * キャッシュからデータを取得
   * @param {string} key - キャッシュキー
   * @returns {*} キャッシュされたデータまたはnull
   */
  async get(key) {
    // メモリキャッシュを優先
    if (this.options.memoryFirst && this.memoryCache) {
      const memoryData = this.memoryCache.get(key);
      if (memoryData !== null) {
        return memoryData;
      }
    }

    // ブラウザキャッシュから取得
    if (this.browserCache) {
      const browserData = this.browserCache.get(key);
      if (browserData !== null) {
        // メモリキャッシュにも保存（ホットキャッシュ）
        if (this.memoryCache) {
          this.memoryCache.set(key, browserData);
        }
        return browserData;
      }
    }

    // メモリキャッシュが優先でない場合の処理
    if (!this.options.memoryFirst && this.memoryCache) {
      return this.memoryCache.get(key);
    }

    return null;
  }

  /**
   * キャッシュからエントリを削除
   * @param {string} key - キャッシュキー
   */
  async delete(key) {
    const results = {};

    if (this.memoryCache) {
      results.memory = this.memoryCache.delete(key);
    }

    if (this.browserCache) {
      results.browser = this.browserCache.delete(key);
    }

    return results;
  }

  /**
   * タグに基づいてキャッシュを無効化
   * @param {string|Array} tags - 無効化するタグ
   */
  async invalidateByTag(tags) {
    const results = {};

    if (this.memoryCache) {
      results.memory = this.memoryCache.invalidateByTag(tags);
    }

    // ブラウザキャッシュはタグ機能をサポートしていない
    results.browser = 0;

    return results;
  }

  /**
   * 全キャッシュをクリア
   */
  async clear() {
    const results = {};

    if (this.memoryCache) {
      results.memory = this.memoryCache.clear();
    }

    if (this.browserCache) {
      results.browser = this.browserCache.clear();
    }

    return results;
  }

  /**
   * 期限切れエントリのクリーンアップ
   */
  async cleanup() {
    const results = {};

    if (this.memoryCache) {
      results.memory = this.memoryCache.cleanup();
    }

    if (this.browserCache) {
      results.browser = this.browserCache.cleanup();
    }

    return results;
  }

  /**
   * キャッシュの統計情報を取得
   */
  getStats() {
    return {
      memory: this.memoryCache ? this.memoryCache.getStats() : null,
      browser: this.browserCache ? 'Available' : null
    };
  }

  /**
   * リソースのクリーンアップ
   */
  destroy() {
    if (this.memoryCache) {
      this.memoryCache.destroy();
    }
  }
}

/**
 * デフォルトのキャッシュシステムインスタンス
 */
export const defaultCacheSystem = new CacheSystem({
  memory: {
    maxSize: 100,
    defaultTTL: 300000, // 5分
    enableLogging: import.meta.env.DEV
  },
  browser: {
    storage: 'localStorage',
    prefix: 'tournament_cache_',
    defaultTTL: 3600000, // 1時間
    enableLogging: import.meta.env.DEV
  }
});

/**
 * キャッシュシステムのファクトリー関数
 * @param {Object} options - キャッシュシステムのオプション
 * @returns {CacheSystem} 新しいキャッシュシステムインスタンス
 */
export function createCacheSystem(options = {}) {
  return new CacheSystem(options);
}

/**
 * キャッシュ付きデータフェッチ関数
 * @param {string} key - キャッシュキー
 * @param {Function} fetchFn - データ取得関数
 * @param {Object} options - オプション
 * @returns {Promise} データまたはキャッシュされたデータ
 */
export async function cachedFetch(key, fetchFn, options = {}) {
  const cache = options.cache || defaultCacheSystem;
  
  // キャッシュから取得を試行
  const cachedData = await cache.get(key);
  if (cachedData !== null) {
    return cachedData;
  }

  // キャッシュにない場合はデータを取得
  try {
    const data = await fetchFn();
    
    // 取得したデータをキャッシュに保存
    await cache.set(key, data, options);
    
    return data;
  } catch (error) {
    // エラー時はキャッシュされたデータがあれば返す（stale-while-revalidate）
    if (options.staleWhileRevalidate) {
      const staleData = await cache.get(key);
      if (staleData !== null) {
        return staleData;
      }
    }
    
    throw error;
  }
}
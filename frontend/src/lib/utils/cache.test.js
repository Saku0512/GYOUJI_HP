// キャッシュシステムのテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { 
  MemoryCache, 
  BrowserCache, 
  CacheSystem, 
  createCacheSystem, 
  cachedFetch 
} from './cache.js';

describe('MemoryCache', () => {
  let memoryCache;

  beforeEach(() => {
    vi.useFakeTimers();
    memoryCache = new MemoryCache({
      maxSize: 3,
      defaultTTL: 1000,
      cleanupInterval: 500,
      enableLogging: false
    });
  });

  afterEach(() => {
    memoryCache.destroy();
    vi.useRealTimers();
  });

  describe('基本機能', () => {
    it('データを保存・取得できる', () => {
      memoryCache.set('key1', 'value1');
      expect(memoryCache.get('key1')).toBe('value1');
    });

    it('存在しないキーはnullを返す', () => {
      expect(memoryCache.get('nonexistent')).toBeNull();
    });

    it('データを削除できる', () => {
      memoryCache.set('key1', 'value1');
      expect(memoryCache.delete('key1')).toBe(true);
      expect(memoryCache.get('key1')).toBeNull();
    });

    it('存在しないキーの削除はfalseを返す', () => {
      expect(memoryCache.delete('nonexistent')).toBe(false);
    });
  });

  describe('TTL機能', () => {
    it('TTL期限切れのデータはnullを返す', () => {
      memoryCache.set('key1', 'value1', { ttl: 1000 });
      
      // 1秒経過
      vi.advanceTimersByTime(1001);
      
      expect(memoryCache.get('key1')).toBeNull();
    });

    it('TTL内のデータは取得できる', () => {
      memoryCache.set('key1', 'value1', { ttl: 1000 });
      
      // 500ms経過
      vi.advanceTimersByTime(500);
      
      expect(memoryCache.get('key1')).toBe('value1');
    });
  });

  describe('サイズ制限', () => {
    it('最大サイズを超えるとLRUで削除される', () => {
      // 最大サイズ3に設定されている
      memoryCache.set('key1', 'value1');
      vi.advanceTimersByTime(10); // 時間を進める
      
      memoryCache.set('key2', 'value2');
      vi.advanceTimersByTime(10); // 時間を進める
      
      memoryCache.set('key3', 'value3');
      vi.advanceTimersByTime(10); // 時間を進める
      
      // key1にアクセスして最近使用済みにする
      memoryCache.get('key1');
      vi.advanceTimersByTime(10); // 時間を進める
      
      // 新しいエントリを追加（key2が最も古いので削除されるはず）
      memoryCache.set('key4', 'value4');
      
      expect(memoryCache.get('key1')).toBe('value1'); // 最近アクセスされたので残る
      expect(memoryCache.get('key2')).toBeNull(); // LRUで削除される
      expect(memoryCache.get('key3')).toBe('value3');
      expect(memoryCache.get('key4')).toBe('value4');
    });
  });

  describe('タグ機能', () => {
    it('タグでキャッシュを無効化できる', () => {
      memoryCache.set('key1', 'value1', { tags: ['tag1', 'tag2'] });
      memoryCache.set('key2', 'value2', { tags: ['tag2'] });
      memoryCache.set('key3', 'value3', { tags: ['tag3'] });
      
      const invalidated = memoryCache.invalidateByTag('tag2');
      
      expect(invalidated).toBe(2);
      expect(memoryCache.get('key1')).toBeNull();
      expect(memoryCache.get('key2')).toBeNull();
      expect(memoryCache.get('key3')).toBe('value3');
    });

    it('複数タグで無効化できる', () => {
      memoryCache.set('key1', 'value1', { tags: ['tag1'] });
      memoryCache.set('key2', 'value2', { tags: ['tag2'] });
      memoryCache.set('key3', 'value3', { tags: ['tag3'] });
      
      const invalidated = memoryCache.invalidateByTag(['tag1', 'tag3']);
      
      expect(invalidated).toBe(2);
      expect(memoryCache.get('key1')).toBeNull();
      expect(memoryCache.get('key2')).toBe('value2');
      expect(memoryCache.get('key3')).toBeNull();
    });
  });

  describe('統計情報', () => {
    it('正しい統計情報を返す', () => {
      memoryCache.set('key1', 'value1');
      memoryCache.get('key1'); // hit
      memoryCache.get('key2'); // miss
      
      const stats = memoryCache.getStats();
      
      expect(stats.hits).toBe(1);
      expect(stats.misses).toBe(1);
      expect(stats.sets).toBe(1);
      expect(stats.size).toBe(1);
      expect(stats.hitRate).toBe('50.00%');
    });
  });

  describe('クリーンアップ', () => {
    it('期限切れエントリを自動クリーンアップする', () => {
      memoryCache.set('key1', 'value1', { ttl: 100 });
      memoryCache.set('key2', 'value2', { ttl: 2000 });
      
      // 200ms経過（key1が期限切れ）
      vi.advanceTimersByTime(200);
      
      const cleaned = memoryCache.cleanup();
      
      expect(cleaned).toBe(1);
      expect(memoryCache.get('key1')).toBeNull();
      expect(memoryCache.get('key2')).toBe('value2');
    });
  });
});

describe('BrowserCache', () => {
  let browserCache;
  let mockStorage;

  beforeEach(() => {
    vi.useFakeTimers();
    
    // LocalStorageのモック
    mockStorage = {
      data: {},
      setItem: vi.fn((key, value) => {
        mockStorage.data[key] = value;
      }),
      getItem: vi.fn((key) => {
        return mockStorage.data[key] || null;
      }),
      removeItem: vi.fn((key) => {
        delete mockStorage.data[key];
      }),
      get length() {
        return Object.keys(mockStorage.data).length;
      },
      key: vi.fn((index) => {
        const keys = Object.keys(mockStorage.data);
        return keys[index] || null;
      })
    };

    // グローバルのlocalStorageを置き換え
    Object.defineProperty(window, 'localStorage', {
      value: mockStorage,
      writable: true
    });

    browserCache = new BrowserCache({
      prefix: 'test_',
      defaultTTL: 1000,
      enableLogging: false
    });
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  describe('基本機能', () => {
    it('データを保存・取得できる', () => {
      browserCache.set('key1', 'value1');
      expect(browserCache.get('key1')).toBe('value1');
    });

    it('存在しないキーはnullを返す', () => {
      expect(browserCache.get('nonexistent')).toBeNull();
    });

    it('データを削除できる', () => {
      browserCache.set('key1', 'value1');
      expect(browserCache.delete('key1')).toBe(true);
      expect(browserCache.get('key1')).toBeNull();
    });
  });

  describe('TTL機能', () => {
    it('TTL期限切れのデータはnullを返す', () => {
      browserCache.set('key1', 'value1', { ttl: 1000 });
      
      // 1秒経過
      vi.advanceTimersByTime(1001);
      
      expect(browserCache.get('key1')).toBeNull();
    });

    it('TTL内のデータは取得できる', () => {
      browserCache.set('key1', 'value1', { ttl: 1000 });
      
      // 500ms経過
      vi.advanceTimersByTime(500);
      
      expect(browserCache.get('key1')).toBe('value1');
    });
  });

  describe('プレフィックス機能', () => {
    it('正しいプレフィックス付きキーで保存される', () => {
      browserCache.set('key1', 'value1');
      
      expect(mockStorage.setItem).toHaveBeenCalledWith(
        'test_key1',
        expect.any(String)
      );
    });
  });

  describe('クリーンアップ', () => {
    it('期限切れエントリをクリーンアップできる', () => {
      browserCache.set('key1', 'value1', { ttl: 100 });
      browserCache.set('key2', 'value2', { ttl: 2000 });
      
      // 200ms経過
      vi.advanceTimersByTime(200);
      
      const cleaned = browserCache.cleanup();
      
      expect(cleaned).toBe(1);
      expect(browserCache.get('key1')).toBeNull();
      expect(browserCache.get('key2')).toBe('value2');
    });
  });
});

describe('CacheSystem', () => {
  let cacheSystem;
  let mockStorage;

  beforeEach(() => {
    vi.useFakeTimers();
    
    // LocalStorageのモック
    mockStorage = {
      data: {},
      setItem: vi.fn((key, value) => {
        mockStorage.data[key] = value;
      }),
      getItem: vi.fn((key) => {
        return mockStorage.data[key] || null;
      }),
      removeItem: vi.fn((key) => {
        delete mockStorage.data[key];
      }),
      get length() {
        return Object.keys(mockStorage.data).length;
      },
      key: vi.fn((index) => {
        const keys = Object.keys(mockStorage.data);
        return keys[index] || null;
      })
    };

    Object.defineProperty(window, 'localStorage', {
      value: mockStorage,
      writable: true
    });

    cacheSystem = new CacheSystem({
      memory: {
        maxSize: 5,
        defaultTTL: 1000,
        enableLogging: false
      },
      browser: {
        prefix: 'test_',
        defaultTTL: 2000,
        enableLogging: false
      }
    });
  });

  afterEach(() => {
    cacheSystem.destroy();
    vi.useRealTimers();
  });

  describe('統合機能', () => {
    it('メモリキャッシュを優先して取得する', async () => {
      await cacheSystem.set('key1', 'memory_value');
      
      // ブラウザキャッシュに直接異なる値を設定
      cacheSystem.browserCache.set('key1', 'browser_value');
      
      const result = await cacheSystem.get('key1');
      expect(result).toBe('memory_value');
    });

    it('メモリキャッシュにない場合はブラウザキャッシュから取得する', async () => {
      // ブラウザキャッシュのみに保存
      cacheSystem.browserCache.set('key1', 'browser_value');
      
      const result = await cacheSystem.get('key1');
      expect(result).toBe('browser_value');
    });

    it('両方のキャッシュからデータを削除する', async () => {
      await cacheSystem.set('key1', 'value1');
      
      const results = await cacheSystem.delete('key1');
      
      expect(results.memory).toBe(true);
      expect(results.browser).toBe(true);
      expect(await cacheSystem.get('key1')).toBeNull();
    });

    it('タグによる無効化はメモリキャッシュのみ対応', async () => {
      await cacheSystem.set('key1', 'value1', { tags: ['tag1'] });
      
      const results = await cacheSystem.invalidateByTag('tag1');
      
      expect(results.memory).toBe(1);
      expect(results.browser).toBe(0);
    });
  });

  describe('統計情報', () => {
    it('両方のキャッシュの統計情報を取得できる', () => {
      const stats = cacheSystem.getStats();
      
      expect(stats.memory).toBeDefined();
      expect(stats.browser).toBe('Available');
    });
  });
});

describe('cachedFetch', () => {
  let cacheSystem;
  let mockFetchFn;

  beforeEach(() => {
    vi.useFakeTimers();
    
    cacheSystem = createCacheSystem({
      memory: {
        maxSize: 5,
        defaultTTL: 1000,
        enableLogging: false
      },
      enableBrowserCache: false // テスト簡略化のため無効
    });

    mockFetchFn = vi.fn().mockResolvedValue('fetched_data');
  });

  afterEach(() => {
    cacheSystem.destroy();
    vi.useRealTimers();
  });

  describe('キャッシュ付きフェッチ', () => {
    it('初回はfetchFnを呼び出してデータを取得する', async () => {
      const result = await cachedFetch('key1', mockFetchFn, { cache: cacheSystem });
      
      expect(mockFetchFn).toHaveBeenCalledTimes(1);
      expect(result).toBe('fetched_data');
    });

    it('2回目以降はキャッシュからデータを取得する', async () => {
      await cachedFetch('key1', mockFetchFn, { cache: cacheSystem });
      const result = await cachedFetch('key1', mockFetchFn, { cache: cacheSystem });
      
      expect(mockFetchFn).toHaveBeenCalledTimes(1); // 1回のみ
      expect(result).toBe('fetched_data');
    });

    it('fetchFnがエラーの場合は例外を投げる', async () => {
      const errorFetchFn = vi.fn().mockRejectedValue(new Error('Fetch failed'));
      
      await expect(cachedFetch('key1', errorFetchFn, { cache: cacheSystem }))
        .rejects.toThrow('Fetch failed');
    });

    it('staleWhileRevalidateオプションでエラー時にキャッシュデータを返す', async () => {
      // 最初に成功してキャッシュに保存
      await cachedFetch('key1', mockFetchFn, { cache: cacheSystem });
      
      // 2回目はエラーだがキャッシュデータを返す
      const errorFetchFn = vi.fn().mockRejectedValue(new Error('Fetch failed'));
      const result = await cachedFetch('key1', errorFetchFn, { 
        cache: cacheSystem,
        staleWhileRevalidate: true 
      });
      
      expect(result).toBe('fetched_data');
    });
  });
});

describe('ファクトリー関数', () => {
  it('createCacheSystemが正しくインスタンスを作成する', () => {
    const instance = createCacheSystem({
      memory: { maxSize: 10 },
      browser: { prefix: 'custom_' }
    });
    
    expect(instance).toBeInstanceOf(CacheSystem);
    expect(instance.memoryCache.options.maxSize).toBe(10);
    expect(instance.browserCache.options.prefix).toBe('custom_');
    
    instance.destroy();
  });
});
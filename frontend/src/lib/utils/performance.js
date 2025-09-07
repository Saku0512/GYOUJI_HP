/**
 * パフォーマンス監視ユーティリティ
 * バンドルサイズ分析とパフォーマンステストのためのツール
 */

/**
 * パフォーマンスメトリクスを測定
 */
export class PerformanceMonitor {
  constructor() {
    this.metrics = new Map();
    this.observers = [];
  }

  /**
   * 処理時間を測定
   * @param {string} name - メトリクス名
   * @param {Function} fn - 測定する関数
   */
  async measureTime(name, fn) {
    const startTime = performance.now();
    try {
      const result = await fn();
      const endTime = performance.now();
      const duration = endTime - startTime;
      
      this.recordMetric(name, {
        type: 'timing',
        duration,
        timestamp: Date.now()
      });
      
      return result;
    } catch (error) {
      const endTime = performance.now();
      const duration = endTime - startTime;
      
      this.recordMetric(name, {
        type: 'timing',
        duration,
        timestamp: Date.now(),
        error: error.message
      });
      
      throw error;
    }
  }

  /**
   * メトリクスを記録
   * @param {string} name - メトリクス名
   * @param {Object} data - メトリクスデータ
   */
  recordMetric(name, data) {
    if (!this.metrics.has(name)) {
      this.metrics.set(name, []);
    }
    this.metrics.get(name).push(data);
    
    // 開発環境でのみログ出力
    if (import.meta.env.DEV) {
      console.log(`[Performance] ${name}:`, data);
    }
  }

  /**
   * バンドルサイズ情報を取得
   */
  getBundleInfo() {
    if (typeof window === 'undefined') return null;
    
    const scripts = Array.from(document.querySelectorAll('script[src]'));
    const styles = Array.from(document.querySelectorAll('link[rel="stylesheet"]'));
    
    return {
      scripts: scripts.map(script => ({
        src: script.src,
        async: script.async,
        defer: script.defer
      })),
      styles: styles.map(style => ({
        href: style.href
      })),
      totalScripts: scripts.length,
      totalStyles: styles.length
    };
  }

  /**
   * Core Web Vitals を測定
   */
  measureWebVitals() {
    if (typeof window === 'undefined') return;

    // Largest Contentful Paint (LCP)
    if ('PerformanceObserver' in window) {
      const lcpObserver = new PerformanceObserver((entryList) => {
        const entries = entryList.getEntries();
        const lastEntry = entries[entries.length - 1];
        
        this.recordMetric('LCP', {
          type: 'web-vital',
          value: lastEntry.startTime,
          timestamp: Date.now()
        });
      });
      
      lcpObserver.observe({ entryTypes: ['largest-contentful-paint'] });
      this.observers.push(lcpObserver);

      // First Input Delay (FID)
      const fidObserver = new PerformanceObserver((entryList) => {
        const entries = entryList.getEntries();
        entries.forEach(entry => {
          this.recordMetric('FID', {
            type: 'web-vital',
            value: entry.processingStart - entry.startTime,
            timestamp: Date.now()
          });
        });
      });
      
      fidObserver.observe({ entryTypes: ['first-input'] });
      this.observers.push(fidObserver);

      // Cumulative Layout Shift (CLS)
      let clsValue = 0;
      const clsObserver = new PerformanceObserver((entryList) => {
        const entries = entryList.getEntries();
        entries.forEach(entry => {
          if (!entry.hadRecentInput) {
            clsValue += entry.value;
          }
        });
        
        this.recordMetric('CLS', {
          type: 'web-vital',
          value: clsValue,
          timestamp: Date.now()
        });
      });
      
      clsObserver.observe({ entryTypes: ['layout-shift'] });
      this.observers.push(clsObserver);
    }

    // Navigation Timing
    if (window.performance && window.performance.timing) {
      const timing = window.performance.timing;
      const navigationStart = timing.navigationStart;
      
      this.recordMetric('page-load', {
        type: 'navigation',
        domContentLoaded: timing.domContentLoadedEventEnd - navigationStart,
        loadComplete: timing.loadEventEnd - navigationStart,
        firstPaint: timing.responseStart - navigationStart,
        timestamp: Date.now()
      });
    }
  }

  /**
   * メモリ使用量を測定
   */
  measureMemoryUsage() {
    if (typeof window === 'undefined' || !window.performance || !window.performance.memory) {
      return null;
    }

    const memory = window.performance.memory;
    const memoryInfo = {
      usedJSHeapSize: memory.usedJSHeapSize,
      totalJSHeapSize: memory.totalJSHeapSize,
      jsHeapSizeLimit: memory.jsHeapSizeLimit,
      timestamp: Date.now()
    };

    this.recordMetric('memory-usage', {
      type: 'memory',
      ...memoryInfo
    });

    return memoryInfo;
  }

  /**
   * 全メトリクスを取得
   */
  getAllMetrics() {
    const result = {};
    for (const [name, metrics] of this.metrics) {
      result[name] = metrics;
    }
    return result;
  }

  /**
   * メトリクスをクリア
   */
  clearMetrics() {
    this.metrics.clear();
  }

  /**
   * オブザーバーを停止
   */
  disconnect() {
    this.observers.forEach(observer => observer.disconnect());
    this.observers = [];
  }
}

// グローバルインスタンス
export const performanceMonitor = new PerformanceMonitor();

/**
 * 遅延読み込み用のユーティリティ
 */
export class LazyLoader {
  constructor() {
    this.loadedModules = new Set();
  }

  /**
   * 動的インポートでモジュールを遅延読み込み
   * @param {Function} importFn - import()関数
   * @param {string} moduleName - モジュール名（キャッシュ用）
   */
  async loadModule(importFn, moduleName) {
    if (this.loadedModules.has(moduleName)) {
      return;
    }

    return performanceMonitor.measureTime(`lazy-load-${moduleName}`, async () => {
      const module = await importFn();
      this.loadedModules.add(moduleName);
      return module;
    });
  }

  /**
   * 画像の遅延読み込み
   * @param {HTMLImageElement} img - 画像要素
   * @param {string} src - 画像URL
   */
  loadImage(img, src) {
    return new Promise((resolve, reject) => {
      const startTime = performance.now();
      
      img.onload = () => {
        const loadTime = performance.now() - startTime;
        performanceMonitor.recordMetric('image-load', {
          type: 'asset',
          src,
          loadTime,
          timestamp: Date.now()
        });
        resolve();
      };
      
      img.onerror = () => {
        const loadTime = performance.now() - startTime;
        performanceMonitor.recordMetric('image-load-error', {
          type: 'asset',
          src,
          loadTime,
          timestamp: Date.now()
        });
        reject(new Error(`Failed to load image: ${src}`));
      };
      
      img.src = src;
    });
  }
}

export const lazyLoader = new LazyLoader();

/**
 * パフォーマンステスト用のヘルパー関数
 */
export function runPerformanceTest() {
  if (typeof window === 'undefined') return;

  console.log('🚀 パフォーマンステストを開始します...');
  
  // Web Vitals測定開始
  performanceMonitor.measureWebVitals();
  
  // メモリ使用量測定
  const memoryInfo = performanceMonitor.measureMemoryUsage();
  if (memoryInfo) {
    console.log('💾 メモリ使用量:', {
      used: `${(memoryInfo.usedJSHeapSize / 1024 / 1024).toFixed(2)} MB`,
      total: `${(memoryInfo.totalJSHeapSize / 1024 / 1024).toFixed(2)} MB`,
      limit: `${(memoryInfo.jsHeapSizeLimit / 1024 / 1024).toFixed(2)} MB`
    });
  }
  
  // バンドル情報取得
  const bundleInfo = performanceMonitor.getBundleInfo();
  if (bundleInfo) {
    console.log('📦 バンドル情報:', bundleInfo);
  }
  
  // 5秒後に結果を出力
  setTimeout(() => {
    const metrics = performanceMonitor.getAllMetrics();
    console.log('📊 パフォーマンスメトリクス:', metrics);
  }, 5000);
}
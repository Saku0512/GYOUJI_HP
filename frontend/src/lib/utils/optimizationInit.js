/**
 * 最適化機能の初期化
 * アプリケーション起動時に実行される最適化設定
 */

import { performanceMonitor, initRUM } from './performanceMonitor.js';
import { cacheManager, assetPreloader } from './assetCache.js';
import { lazyImageLoader, preloadCriticalImages } from './imageOptimization.js';

/**
 * 最適化機能の初期化設定
 */
const OPTIMIZATION_CONFIG = {
  // パフォーマンス監視
  performance: {
    enabled: true,
    rumSampleRate: 0.1, // 10%のユーザーからデータを収集
    endpoint: null // 分析エンドポイント（設定されている場合）
  },
  
  // 画像最適化
  images: {
    lazyLoading: true,
    webpSupport: true,
    criticalImages: [
      '/icon-192x192.png',
      '/logo.svg'
    ]
  },
  
  // アセットキャッシュ
  cache: {
    enabled: true,
    preloadAssets: [
      '/manifest.json',
      '/sw.js'
    ]
  },
  
  // Service Worker
  serviceWorker: {
    enabled: true,
    updateInterval: 60000, // 1分ごとに更新チェック
    skipWaiting: false
  }
};

/**
 * Service Worker の登録と管理
 */
class ServiceWorkerManager {
  constructor() {
    this.registration = null;
    this.updateAvailable = false;
  }

  async register() {
    if (!('serviceWorker' in navigator)) {
      console.warn('Service Worker is not supported');
      return false;
    }

    try {
      this.registration = await navigator.serviceWorker.register('/sw.js', {
        scope: '/'
      });

      console.log('Service Worker registered successfully');

      // 更新の監視
      this.registration.addEventListener('updatefound', () => {
        const newWorker = this.registration.installing;
        
        newWorker.addEventListener('statechange', () => {
          if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
            this.updateAvailable = true;
            this.notifyUpdate();
          }
        });
      });

      // 定期的な更新チェック
      if (OPTIMIZATION_CONFIG.serviceWorker.updateInterval > 0) {
        setInterval(() => {
          this.registration.update();
        }, OPTIMIZATION_CONFIG.serviceWorker.updateInterval);
      }

      return true;
    } catch (error) {
      console.error('Service Worker registration failed:', error);
      return false;
    }
  }

  notifyUpdate() {
    // カスタムイベントを発火して UI に更新を通知
    window.dispatchEvent(new CustomEvent('sw-update-available', {
      detail: { registration: this.registration }
    }));
  }

  async skipWaiting() {
    if (this.registration && this.registration.waiting) {
      // Service Worker にスキップ待機メッセージを送信
      this.registration.waiting.postMessage({ type: 'SKIP_WAITING' });
      
      // ページをリロード
      window.location.reload();
    }
  }

  async getCacheInfo() {
    if (this.registration && this.registration.active) {
      return new Promise((resolve) => {
        const messageChannel = new MessageChannel();
        
        messageChannel.port1.onmessage = (event) => {
          if (event.data.type === 'CACHE_INFO') {
            resolve(event.data.payload);
          }
        };
        
        this.registration.active.postMessage(
          { type: 'GET_CACHE_INFO' },
          [messageChannel.port2]
        );
      });
    }
    return null;
  }

  async clearCache() {
    if (this.registration && this.registration.active) {
      return new Promise((resolve) => {
        const messageChannel = new MessageChannel();
        
        messageChannel.port1.onmessage = (event) => {
          if (event.data.type === 'CACHE_CLEARED') {
            resolve();
          }
        };
        
        this.registration.active.postMessage(
          { type: 'CLEAR_CACHE' },
          [messageChannel.port2]
        );
      });
    }
  }
}

/**
 * 重要なリソースのプリロード
 */
async function preloadCriticalResources() {
  const criticalAssets = [
    ...OPTIMIZATION_CONFIG.cache.preloadAssets,
    ...OPTIMIZATION_CONFIG.images.criticalImages
  ];

  try {
    await assetPreloader.preloadCriticalAssets(criticalAssets);
    console.log('Critical resources preloaded successfully');
  } catch (error) {
    console.warn('Failed to preload some critical resources:', error);
  }
}

/**
 * 画像最適化の初期化
 */
function initImageOptimization() {
  if (!OPTIMIZATION_CONFIG.images.lazyLoading) return;

  // 既存の画像に遅延読み込みを適用
  const images = document.querySelectorAll('img[data-src]');
  images.forEach(img => {
    lazyImageLoader.observe(img);
  });

  // 重要な画像のプリロード
  if (OPTIMIZATION_CONFIG.images.criticalImages.length > 0) {
    preloadCriticalImages(OPTIMIZATION_CONFIG.images.criticalImages);
  }
}

/**
 * パフォーマンス監視の初期化
 */
function initPerformanceMonitoring() {
  if (!OPTIMIZATION_CONFIG.performance.enabled) return;

  // RUM の初期化
  if (OPTIMIZATION_CONFIG.performance.endpoint) {
    initRUM({
      sampleRate: OPTIMIZATION_CONFIG.performance.rumSampleRate,
      endpoint: OPTIMIZATION_CONFIG.performance.endpoint
    });
  }

  // カスタムメトリクスの記録
  performanceMonitor.recordCustomMetric('app-init', Date.now(), {
    userAgent: navigator.userAgent,
    viewport: {
      width: window.innerWidth,
      height: window.innerHeight
    }
  });
}

/**
 * ネットワーク状態の監視
 */
function initNetworkMonitoring() {
  // オンライン/オフライン状態の監視
  function updateNetworkStatus() {
    const isOnline = navigator.onLine;
    
    performanceMonitor.recordCustomMetric('network-status', isOnline ? 'online' : 'offline');
    
    // オフライン時の処理
    if (!isOnline) {
      console.warn('Network is offline');
      // 必要に応じてオフラインページにリダイレクト
    }
  }

  window.addEventListener('online', updateNetworkStatus);
  window.addEventListener('offline', updateNetworkStatus);
  
  // 初期状態を記録
  updateNetworkStatus();

  // 接続品質の監視（対応ブラウザのみ）
  if ('connection' in navigator) {
    const connection = navigator.connection;
    
    function updateConnectionInfo() {
      performanceMonitor.recordCustomMetric('connection-info', {
        effectiveType: connection.effectiveType,
        downlink: connection.downlink,
        rtt: connection.rtt,
        saveData: connection.saveData
      });
    }
    
    connection.addEventListener('change', updateConnectionInfo);
    updateConnectionInfo();
  }
}

/**
 * バックグラウンドタスクの最適化
 */
function initBackgroundOptimization() {
  // ページ可視性の監視
  function handleVisibilityChange() {
    if (document.hidden) {
      // ページが非表示になった時の処理
      performanceMonitor.recordCustomMetric('page-hidden', Date.now());
      
      // 不要な処理を停止
      // 例: アニメーション、ポーリングなど
    } else {
      // ページが表示された時の処理
      performanceMonitor.recordCustomMetric('page-visible', Date.now());
      
      // 必要な処理を再開
    }
  }

  document.addEventListener('visibilitychange', handleVisibilityChange);

  // アイドル時の最適化（Idle Detection API対応ブラウザ）
  if ('requestIdleCallback' in window) {
    function performIdleTasks(deadline) {
      while (deadline.timeRemaining() > 0) {
        // アイドル時に実行する軽量なタスク
        // 例: キャッシュクリーンアップ、プリロードなど
        cacheManager.memoryCache.cleanup();
        break;
      }
      
      // 次のアイドル時間を予約
      requestIdleCallback(performIdleTasks);
    }
    
    requestIdleCallback(performIdleTasks);
  }
}

/**
 * エラー監視の初期化
 */
function initErrorMonitoring() {
  // JavaScript エラーの監視
  window.addEventListener('error', (event) => {
    performanceMonitor.recordCustomMetric('js-error', {
      message: event.message,
      filename: event.filename,
      lineno: event.lineno,
      colno: event.colno,
      stack: event.error?.stack
    });
  });

  // Promise rejection の監視
  window.addEventListener('unhandledrejection', (event) => {
    performanceMonitor.recordCustomMetric('unhandled-rejection', {
      reason: event.reason?.toString(),
      stack: event.reason?.stack
    });
  });

  // リソース読み込みエラーの監視
  window.addEventListener('error', (event) => {
    if (event.target !== window) {
      performanceMonitor.recordCustomMetric('resource-error', {
        tagName: event.target.tagName,
        src: event.target.src || event.target.href,
        type: event.target.type
      });
    }
  }, true);
}

/**
 * メイン初期化関数
 */
export async function initOptimizations(config = {}) {
  // 設定をマージ
  Object.assign(OPTIMIZATION_CONFIG, config);

  console.log('🚀 Initializing optimizations...');

  try {
    // Service Worker の登録
    const swManager = new ServiceWorkerManager();
    if (OPTIMIZATION_CONFIG.serviceWorker.enabled) {
      await swManager.register();
    }

    // 各種最適化機能の初期化
    await Promise.all([
      preloadCriticalResources(),
      initImageOptimization(),
      initPerformanceMonitoring(),
      initNetworkMonitoring(),
      initBackgroundOptimization(),
      initErrorMonitoring()
    ]);

    console.log('✅ Optimizations initialized successfully');

    // グローバルに Service Worker マネージャーを公開
    window.__swManager = swManager;

    return {
      swManager,
      performanceMonitor,
      cacheManager
    };

  } catch (error) {
    console.error('❌ Failed to initialize optimizations:', error);
    throw error;
  }
}

/**
 * 最適化設定の更新
 */
export function updateOptimizationConfig(newConfig) {
  Object.assign(OPTIMIZATION_CONFIG, newConfig);
}

/**
 * 現在の最適化設定を取得
 */
export function getOptimizationConfig() {
  return { ...OPTIMIZATION_CONFIG };
}
/**
 * Service Worker 登録と管理ユーティリティ
 * オフライン対応とキャッシュ管理を提供
 */

import { performanceMonitor } from './performance.js';

/**
 * Service Worker 管理クラス
 */
export class ServiceWorkerManager {
  constructor() {
    this.registration = null;
    this.isSupported = 'serviceWorker' in navigator;
    this.isOnline = navigator.onLine;
    this.updateAvailable = false;
    this.listeners = new Map();
    
    // オンライン/オフライン状態の監視
    this.setupNetworkListeners();
  }

  /**
   * Service Worker を登録
   */
  async register() {
    if (!this.isSupported) {
      console.warn('Service Worker is not supported in this browser');
      return null;
    }

    try {
      const startTime = performance.now();
      
      this.registration = await navigator.serviceWorker.register('/sw.js', {
        scope: '/'
      });

      console.log('Service Worker registered successfully:', this.registration);
      
      // パフォーマンスメトリクスを記録
      performanceMonitor.recordMetric('service-worker-registration', {
        type: 'service-worker',
        registrationTime: performance.now() - startTime,
        scope: this.registration.scope,
        timestamp: Date.now()
      });

      // 更新チェックの設定
      this.setupUpdateHandling();
      
      // メッセージハンドリングの設定
      this.setupMessageHandling();
      
      return this.registration;
      
    } catch (error) {
      console.error('Service Worker registration failed:', error);
      
      // エラーメトリクスを記録
      performanceMonitor.recordMetric('service-worker-registration-error', {
        type: 'service-worker-error',
        error: error.message,
        timestamp: Date.now()
      });
      
      throw error;
    }
  }

  /**
   * Service Worker の更新処理を設定
   */
  setupUpdateHandling() {
    if (!this.registration) return;

    // 新しい Service Worker が利用可能になった時
    this.registration.addEventListener('updatefound', () => {
      const newWorker = this.registration.installing;
      
      if (newWorker) {
        newWorker.addEventListener('statechange', () => {
          if (newWorker.state === 'installed' && navigator.serviceWorker.controller) {
            // 新しいバージョンが利用可能
            this.updateAvailable = true;
            this.emit('updateAvailable', { registration: this.registration });
          }
        });
      }
    });

    // Service Worker が制御を開始した時
    navigator.serviceWorker.addEventListener('controllerchange', () => {
      this.emit('controllerChange');
      
      // ページをリロードして新しい Service Worker を適用
      if (this.updateAvailable) {
        window.location.reload();
      }
    });
  }

  /**
   * メッセージハンドリングを設定
   */
  setupMessageHandling() {
    navigator.serviceWorker.addEventListener('message', (event) => {
      const { type, payload } = event.data;
      this.emit('message', { type, payload });
    });
  }

  /**
   * ネットワーク状態の監視を設定
   */
  setupNetworkListeners() {
    window.addEventListener('online', () => {
      this.isOnline = true;
      this.emit('online');
      
      performanceMonitor.recordMetric('network-status-change', {
        type: 'network',
        status: 'online',
        timestamp: Date.now()
      });
    });

    window.addEventListener('offline', () => {
      this.isOnline = false;
      this.emit('offline');
      
      performanceMonitor.recordMetric('network-status-change', {
        type: 'network',
        status: 'offline',
        timestamp: Date.now()
      });
    });
  }

  /**
   * Service Worker にメッセージを送信
   */
  async sendMessage(type, payload = {}) {
    if (!this.registration || !this.registration.active) {
      throw new Error('Service Worker is not active');
    }

    return new Promise((resolve, reject) => {
      const messageChannel = new MessageChannel();
      
      messageChannel.port1.onmessage = (event) => {
        resolve(event.data);
      };
      
      messageChannel.port1.onerror = (error) => {
        reject(error);
      };
      
      this.registration.active.postMessage(
        { type, payload },
        [messageChannel.port2]
      );
      
      // タイムアウト設定
      setTimeout(() => {
        reject(new Error('Message timeout'));
      }, 5000);
    });
  }

  /**
   * キャッシュ情報を取得
   */
  async getCacheInfo() {
    try {
      return await this.sendMessage('GET_CACHE_INFO');
    } catch (error) {
      console.error('Failed to get cache info:', error);
      return null;
    }
  }

  /**
   * キャッシュをクリア
   */
  async clearCache() {
    try {
      await this.sendMessage('CLEAR_CACHE');
      this.emit('cacheCleared');
      return true;
    } catch (error) {
      console.error('Failed to clear cache:', error);
      return false;
    }
  }

  /**
   * Service Worker を更新
   */
  async update() {
    if (!this.registration) {
      throw new Error('Service Worker is not registered');
    }

    try {
      await this.registration.update();
      this.emit('updateChecked');
    } catch (error) {
      console.error('Service Worker update failed:', error);
      throw error;
    }
  }

  /**
   * 新しい Service Worker をアクティベート
   */
  async skipWaiting() {
    if (!this.registration || !this.registration.waiting) {
      return;
    }

    try {
      await this.sendMessage('SKIP_WAITING');
    } catch (error) {
      console.error('Skip waiting failed:', error);
    }
  }

  /**
   * バックグラウンド同期を登録
   */
  async registerBackgroundSync(tag = 'background-sync') {
    if (!this.registration || !('sync' in this.registration)) {
      console.warn('Background Sync is not supported');
      return false;
    }

    try {
      await this.registration.sync.register(tag);
      return true;
    } catch (error) {
      console.error('Background sync registration failed:', error);
      return false;
    }
  }

  /**
   * イベントリスナーを追加
   */
  on(event, callback) {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event).push(callback);
  }

  /**
   * イベントリスナーを削除
   */
  off(event, callback) {
    if (!this.listeners.has(event)) return;
    
    const callbacks = this.listeners.get(event);
    const index = callbacks.indexOf(callback);
    
    if (index > -1) {
      callbacks.splice(index, 1);
    }
  }

  /**
   * イベントを発火
   */
  emit(event, data = {}) {
    if (!this.listeners.has(event)) return;
    
    this.listeners.get(event).forEach(callback => {
      try {
        callback(data);
      } catch (error) {
        console.error(`Event listener error for ${event}:`, error);
      }
    });
  }

  /**
   * Service Worker の状態を取得
   */
  getStatus() {
    return {
      isSupported: this.isSupported,
      isRegistered: !!this.registration,
      isOnline: this.isOnline,
      updateAvailable: this.updateAvailable,
      registration: this.registration
    };
  }

  /**
   * リソースのクリーンアップ
   */
  destroy() {
    this.listeners.clear();
    
    // イベントリスナーの削除
    window.removeEventListener('online', this.handleOnline);
    window.removeEventListener('offline', this.handleOffline);
  }
}

// グローバルインスタンス
export const serviceWorkerManager = new ServiceWorkerManager();

/**
 * Service Worker を初期化
 */
export async function initializeServiceWorker() {
  try {
    await serviceWorkerManager.register();
    
    // 定期的な更新チェック
    setInterval(() => {
      serviceWorkerManager.update().catch(console.error);
    }, 60000); // 1分間隔
    
    return serviceWorkerManager;
  } catch (error) {
    console.error('Service Worker initialization failed:', error);
    return null;
  }
}

/**
 * オフライン状態をチェック
 */
export function isOffline() {
  return !navigator.onLine;
}

/**
 * オンライン状態をチェック
 */
export function isOnline() {
  return navigator.onLine;
}

/**
 * ネットワーク状態の変更を監視
 */
export function watchNetworkStatus(callback) {
  const handleOnline = () => callback(true);
  const handleOffline = () => callback(false);
  
  window.addEventListener('online', handleOnline);
  window.addEventListener('offline', handleOffline);
  
  // 初期状態を通知
  callback(navigator.onLine);
  
  // クリーンアップ関数を返す
  return () => {
    window.removeEventListener('online', handleOnline);
    window.removeEventListener('offline', handleOffline);
  };
}

/**
 * キャッシュ統計を取得
 */
export async function getCacheStats() {
  try {
    const cacheInfo = await serviceWorkerManager.getCacheInfo();
    
    if (cacheInfo && cacheInfo.payload) {
      const stats = cacheInfo.payload;
      const totalEntries = Object.values(stats).reduce((sum, count) => sum + count, 0);
      
      return {
        caches: stats,
        totalEntries,
        timestamp: Date.now()
      };
    }
    
    return null;
  } catch (error) {
    console.error('Failed to get cache stats:', error);
    return null;
  }
}

/**
 * アプリケーション更新の通知
 */
export function notifyAppUpdate(callback) {
  serviceWorkerManager.on('updateAvailable', callback);
  
  return () => {
    serviceWorkerManager.off('updateAvailable', callback);
  };
}
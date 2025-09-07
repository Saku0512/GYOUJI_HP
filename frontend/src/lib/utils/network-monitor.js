// ネットワーク状態監視とオフライン対応
import { writable } from 'svelte/store';
import { uiActions } from '$lib/stores/ui.js';
import { globalErrorHandler, AppError, ERROR_TYPES, ERROR_LEVELS } from './error-handler.js';

// ネットワーク状態ストア
export const networkStore = writable({
  isOnline: typeof navigator !== 'undefined' ? navigator.onLine : true,
  connectionType: 'unknown',
  effectiveType: 'unknown',
  downlink: 0,
  rtt: 0,
  saveData: false,
  lastOnlineTime: null,
  lastOfflineTime: null
});

// ネットワーク監視クラス
class NetworkMonitor {
  constructor() {
    this.isInitialized = false;
    this.onlineListeners = [];
    this.offlineListeners = [];
    this.connectionChangeListeners = [];
    this.retryQueue = [];
    this.isProcessingQueue = false;
  }

  // 初期化
  initialize() {
    if (this.isInitialized || typeof window === 'undefined') {
      return;
    }

    // オンライン/オフラインイベントのリスナー設定
    window.addEventListener('online', this.handleOnline.bind(this));
    window.addEventListener('offline', this.handleOffline.bind(this));

    // Network Information API が利用可能な場合
    if ('connection' in navigator) {
      const connection = navigator.connection || navigator.mozConnection || navigator.webkitConnection;
      
      if (connection) {
        // 接続情報の初期化
        this.updateConnectionInfo(connection);
        
        // 接続変更イベントのリスナー設定
        connection.addEventListener('change', () => {
          this.updateConnectionInfo(connection);
          this.notifyConnectionChange();
        });
      }
    }

    // 初期状態の設定
    this.updateNetworkStatus(navigator.onLine);
    
    this.isInitialized = true;
    console.log('[NetworkMonitor] Network monitoring initialized');
  }

  // オンラインイベントハンドラー
  handleOnline() {
    console.log('[NetworkMonitor] Network connection restored');
    
    this.updateNetworkStatus(true);
    
    // オンライン復帰通知
    uiActions.showNotification('インターネット接続が復旧しました', 'success', 3000);
    
    // リスナーに通知
    this.onlineListeners.forEach(listener => {
      try {
        listener();
      } catch (error) {
        console.error('[NetworkMonitor] Error in online listener:', error);
      }
    });

    // 再試行キューの処理
    this.processRetryQueue();
  }

  // オフラインイベントハンドラー
  handleOffline() {
    console.log('[NetworkMonitor] Network connection lost');
    
    this.updateNetworkStatus(false);
    
    // オフライン通知
    uiActions.showNotification('インターネット接続が切断されました', 'warning', 5000);
    
    // リスナーに通知
    this.offlineListeners.forEach(listener => {
      try {
        listener();
      } catch (error) {
        console.error('[NetworkMonitor] Error in offline listener:', error);
      }
    });
  }

  // ネットワーク状態の更新
  updateNetworkStatus(isOnline) {
    const now = new Date().toISOString();
    
    networkStore.update(state => ({
      ...state,
      isOnline,
      lastOnlineTime: isOnline ? now : state.lastOnlineTime,
      lastOfflineTime: !isOnline ? now : state.lastOfflineTime
    }));
  }

  // 接続情報の更新
  updateConnectionInfo(connection) {
    networkStore.update(state => ({
      ...state,
      connectionType: connection.type || 'unknown',
      effectiveType: connection.effectiveType || 'unknown',
      downlink: connection.downlink || 0,
      rtt: connection.rtt || 0,
      saveData: connection.saveData || false
    }));
  }

  // 接続変更通知
  notifyConnectionChange() {
    this.connectionChangeListeners.forEach(listener => {
      try {
        listener();
      } catch (error) {
        console.error('[NetworkMonitor] Error in connection change listener:', error);
      }
    });
  }

  // オンラインリスナーの追加
  addOnlineListener(listener) {
    this.onlineListeners.push(listener);
  }

  // オフラインリスナーの追加
  addOfflineListener(listener) {
    this.offlineListeners.push(listener);
  }

  // 接続変更リスナーの追加
  addConnectionChangeListener(listener) {
    this.connectionChangeListeners.push(listener);
  }

  // リスナーの削除
  removeOnlineListener(listener) {
    const index = this.onlineListeners.indexOf(listener);
    if (index > -1) {
      this.onlineListeners.splice(index, 1);
    }
  }

  removeOfflineListener(listener) {
    const index = this.offlineListeners.indexOf(listener);
    if (index > -1) {
      this.offlineListeners.splice(index, 1);
    }
  }

  removeConnectionChangeListener(listener) {
    const index = this.connectionChangeListeners.indexOf(listener);
    if (index > -1) {
      this.connectionChangeListeners.splice(index, 1);
    }
  }

  // 再試行キューに追加
  addToRetryQueue(operation, context = {}) {
    const retryItem = {
      id: `retry-${Date.now()}-${Math.random()}`,
      operation,
      context,
      timestamp: Date.now(),
      attempts: 0,
      maxAttempts: context.maxAttempts || 3
    };

    this.retryQueue.push(retryItem);
    console.log(`[NetworkMonitor] Added operation to retry queue: ${retryItem.id}`);

    // オンラインの場合は即座に処理
    if (navigator.onLine) {
      this.processRetryQueue();
    }

    return retryItem.id;
  }

  // 再試行キューの処理
  async processRetryQueue() {
    if (this.isProcessingQueue || !navigator.onLine || this.retryQueue.length === 0) {
      return;
    }

    this.isProcessingQueue = true;
    console.log(`[NetworkMonitor] Processing retry queue (${this.retryQueue.length} items)`);

    const itemsToProcess = [...this.retryQueue];
    this.retryQueue = [];

    for (const item of itemsToProcess) {
      try {
        await item.operation();
        console.log(`[NetworkMonitor] Successfully retried operation: ${item.id}`);
      } catch (error) {
        item.attempts++;
        
        if (item.attempts < item.maxAttempts) {
          // 再試行回数が上限に達していない場合はキューに戻す
          this.retryQueue.push(item);
          console.log(`[NetworkMonitor] Retry failed, re-queued: ${item.id} (attempt ${item.attempts}/${item.maxAttempts})`);
        } else {
          // 上限に達した場合はエラーとして処理
          console.error(`[NetworkMonitor] Retry failed permanently: ${item.id}`, error);
          
          const retryError = new AppError(
            `ネットワーク復旧後の再試行に失敗しました: ${error.message}`,
            ERROR_TYPES.NETWORK,
            ERROR_LEVELS.MEDIUM,
            {
              retryId: item.id,
              attempts: item.attempts,
              originalError: error,
              context: item.context
            }
          );
          
          globalErrorHandler.handleError(retryError);
        }
      }
    }

    this.isProcessingQueue = false;
  }

  // 接続品質の評価
  getConnectionQuality() {
    const state = this.getCurrentState();
    
    if (!state.isOnline) {
      return 'offline';
    }

    // effectiveType に基づく品質判定
    switch (state.effectiveType) {
      case 'slow-2g':
        return 'poor';
      case '2g':
        return 'poor';
      case '3g':
        return 'good';
      case '4g':
        return 'excellent';
      default:
        // downlink に基づく判定
        if (state.downlink >= 10) {
          return 'excellent';
        } else if (state.downlink >= 1.5) {
          return 'good';
        } else if (state.downlink >= 0.5) {
          return 'fair';
        } else {
          return 'poor';
        }
    }
  }

  // 現在のネットワーク状態を取得
  getCurrentState() {
    let currentState;
    networkStore.subscribe(state => {
      currentState = state;
    })();
    return currentState;
  }

  // 接続テスト
  async testConnection(url = '/api/health') {
    try {
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), 5000);

      const response = await fetch(url, {
        method: 'HEAD',
        signal: controller.signal,
        cache: 'no-cache'
      });

      clearTimeout(timeoutId);
      return response.ok;
    } catch (error) {
      console.warn('[NetworkMonitor] Connection test failed:', error);
      return false;
    }
  }

  // クリーンアップ
  cleanup() {
    if (typeof window !== 'undefined') {
      window.removeEventListener('online', this.handleOnline.bind(this));
      window.removeEventListener('offline', this.handleOffline.bind(this));
    }

    this.onlineListeners = [];
    this.offlineListeners = [];
    this.connectionChangeListeners = [];
    this.retryQueue = [];
    this.isInitialized = false;
  }
}

// グローバルインスタンス
export const networkMonitor = new NetworkMonitor();

// 便利な関数をエクスポート
export const initializeNetworkMonitor = () => {
  networkMonitor.initialize();
};

export const addToRetryQueue = (operation, context) => {
  return networkMonitor.addToRetryQueue(operation, context);
};

export const getConnectionQuality = () => {
  return networkMonitor.getConnectionQuality();
};

export const testConnection = (url) => {
  return networkMonitor.testConnection(url);
};

// ネットワーク状態に基づく条件付き実行
export const executeWhenOnline = async (operation, options = {}) => {
  const { timeout = 30000, retryOnReconnect = true } = options;
  
  // オンラインの場合は即座に実行
  if (navigator.onLine) {
    try {
      return await operation();
    } catch (error) {
      if (retryOnReconnect && error.name === 'TypeError' && error.message.includes('fetch')) {
        // ネットワークエラーの場合は再試行キューに追加
        networkMonitor.addToRetryQueue(operation, options);
        throw error;
      }
      throw error;
    }
  }

  // オフラインの場合
  if (retryOnReconnect) {
    // 再試行キューに追加
    networkMonitor.addToRetryQueue(operation, options);
    
    // オフライン状態のエラーを投げる
    throw new AppError(
      'オフライン状態です。接続が復旧したら自動的に再試行されます。',
      ERROR_TYPES.NETWORK,
      ERROR_LEVELS.LOW
    );
  } else {
    throw new AppError(
      'インターネット接続が必要です。',
      ERROR_TYPES.NETWORK,
      ERROR_LEVELS.MEDIUM
    );
  }
};
// ポーリングシステム - リアルタイム更新機能
import { get } from 'svelte/store';
import { uiActions } from '../stores/ui.js';

/**
 * ポーリングシステムクラス
 * 定期的なデータ更新、ページ可視性制御、エラー時の再試行ロジックを提供
 */
export class PollingSystem {
  constructor(options = {}) {
    // デフォルト設定
    this.options = {
      interval: 30000, // ポーリング間隔（30秒）
      maxRetries: 3, // 最大再試行回数
      retryDelay: 5000, // 再試行間隔（5秒）
      exponentialBackoff: true, // 指数バックオフの使用
      respectVisibility: true, // ページ可視性の考慮
      enableLogging: true, // ログ出力の有効化
      ...options
    };

    // 内部状態
    this.intervalId = null;
    this.isPolling = false;
    this.retryCount = 0;
    this.lastSuccessTime = null;
    this.lastErrorTime = null;
    this.callbacks = new Map();
    this.errorCallbacks = new Map();
    
    // ページ可視性の監視
    this.isPageVisible = !document.hidden;
    this.setupVisibilityListener();
    
    // ネットワーク状態の監視
    this.isOnline = navigator.onLine;
    this.setupNetworkListener();
  }

  /**
   * ページ可視性変更の監視を設定
   */
  setupVisibilityListener() {
    if (typeof document !== 'undefined') {
      document.addEventListener('visibilitychange', () => {
        this.isPageVisible = !document.hidden;
        
        if (this.options.enableLogging) {
          console.log(`[Polling] Page visibility changed: ${this.isPageVisible ? 'visible' : 'hidden'}`);
        }
        
        // ページが表示されたときに即座にポーリングを実行
        if (this.isPageVisible && this.isPolling) {
          this.executePoll();
        }
      });
    }
  }

  /**
   * ネットワーク状態変更の監視を設定
   */
  setupNetworkListener() {
    if (typeof window !== 'undefined') {
      window.addEventListener('online', () => {
        this.isOnline = true;
        if (this.options.enableLogging) {
          console.log('[Polling] Network connection restored');
        }
        
        // オンラインに復帰したときに即座にポーリングを実行
        if (this.isPolling) {
          this.executePoll();
        }
      });
      
      window.addEventListener('offline', () => {
        this.isOnline = false;
        if (this.options.enableLogging) {
          console.log('[Polling] Network connection lost');
        }
      });
    }
  }

  /**
   * ポーリングコールバックを登録
   * @param {string} key - コールバックのキー
   * @param {Function} callback - 実行するコールバック関数
   */
  registerCallback(key, callback) {
    if (typeof callback !== 'function') {
      throw new Error('Callback must be a function');
    }
    
    this.callbacks.set(key, callback);
    
    if (this.options.enableLogging) {
      console.log(`[Polling] Registered callback: ${key}`);
    }
  }

  /**
   * エラーコールバックを登録
   * @param {string} key - コールバックのキー
   * @param {Function} callback - エラー時に実行するコールバック関数
   */
  registerErrorCallback(key, callback) {
    if (typeof callback !== 'function') {
      throw new Error('Error callback must be a function');
    }
    
    this.errorCallbacks.set(key, callback);
    
    if (this.options.enableLogging) {
      console.log(`[Polling] Registered error callback: ${key}`);
    }
  }

  /**
   * コールバックを削除
   * @param {string} key - 削除するコールバックのキー
   */
  unregisterCallback(key) {
    const removed = this.callbacks.delete(key);
    
    if (this.options.enableLogging && removed) {
      console.log(`[Polling] Unregistered callback: ${key}`);
    }
    
    return removed;
  }

  /**
   * エラーコールバックを削除
   * @param {string} key - 削除するエラーコールバックのキー
   */
  unregisterErrorCallback(key) {
    const removed = this.errorCallbacks.delete(key);
    
    if (this.options.enableLogging && removed) {
      console.log(`[Polling] Unregistered error callback: ${key}`);
    }
    
    return removed;
  }

  /**
   * ポーリングを開始
   */
  start() {
    if (this.isPolling) {
      if (this.options.enableLogging) {
        console.warn('[Polling] Already polling');
      }
      return;
    }

    this.isPolling = true;
    this.retryCount = 0;
    
    if (this.options.enableLogging) {
      console.log(`[Polling] Started with interval: ${this.options.interval}ms`);
    }
    
    // 定期的なポーリングを設定（即座に実行はしない）
    this.intervalId = setInterval(() => {
      this.executePoll();
    }, this.options.interval);
  }

  /**
   * ポーリングを停止
   */
  stop() {
    if (!this.isPolling) {
      return;
    }

    this.isPolling = false;
    
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
    
    if (this.options.enableLogging) {
      console.log('[Polling] Stopped');
    }
  }

  /**
   * ポーリングを一時停止
   */
  pause() {
    if (this.intervalId) {
      clearInterval(this.intervalId);
      this.intervalId = null;
    }
    
    if (this.options.enableLogging) {
      console.log('[Polling] Paused');
    }
  }

  /**
   * ポーリングを再開
   */
  resume() {
    if (this.isPolling && !this.intervalId) {
      this.intervalId = setInterval(() => {
        this.executePoll();
      }, this.options.interval);
      
      if (this.options.enableLogging) {
        console.log('[Polling] Resumed');
      }
    }
  }

  /**
   * 実際のポーリング処理を実行
   */
  async executePoll() {
    // ポーリングが停止されている場合は何もしない
    if (!this.isPolling) {
      return;
    }

    // ページが非表示でrespectVisibilityが有効な場合はスキップ
    if (this.options.respectVisibility && !this.isPageVisible) {
      if (this.options.enableLogging) {
        console.log('[Polling] Skipped due to page visibility');
      }
      return;
    }

    // オフライン状態の場合はスキップ
    if (!this.isOnline) {
      if (this.options.enableLogging) {
        console.log('[Polling] Skipped due to offline status');
      }
      return;
    }

    try {
      if (this.options.enableLogging) {
        console.log('[Polling] Executing poll...');
      }

      // 全てのコールバックを並列実行
      const promises = Array.from(this.callbacks.entries()).map(async ([key, callback]) => {
        try {
          await callback();
          return { key, success: true };
        } catch (error) {
          console.error(`[Polling] Callback error for ${key}:`, error);
          return { key, success: false, error };
        }
      });

      const results = await Promise.allSettled(promises);
      
      // 結果を処理
      let hasErrors = false;
      for (const result of results) {
        if (result.status === 'fulfilled' && !result.value.success) {
          hasErrors = true;
          
          // エラーコールバックを実行
          const errorCallback = this.errorCallbacks.get(result.value.key);
          if (errorCallback) {
            try {
              await errorCallback(result.value.error);
            } catch (callbackError) {
              console.error(`[Polling] Error callback failed for ${result.value.key}:`, callbackError);
            }
          }
        }
      }

      if (hasErrors) {
        throw new Error('One or more polling callbacks failed');
      }

      // 成功時の処理
      this.onPollSuccess();
      
    } catch (error) {
      this.onPollError(error);
    }
  }

  /**
   * ポーリング成功時の処理
   */
  onPollSuccess() {
    this.lastSuccessTime = Date.now();
    this.retryCount = 0; // 成功時はリトライカウントをリセット
    
    if (this.options.enableLogging) {
      console.log('[Polling] Poll completed successfully');
    }
  }

  /**
   * ポーリングエラー時の処理
   * @param {Error} error - 発生したエラー
   */
  async onPollError(error) {
    this.lastErrorTime = Date.now();
    this.retryCount++;
    
    console.error(`[Polling] Poll failed (attempt ${this.retryCount}/${this.options.maxRetries}):`, error);
    
    // 最大再試行回数に達した場合
    if (this.retryCount >= this.options.maxRetries) {
      console.error('[Polling] Max retries reached, showing error notification');
      
      // エラー通知を表示
      uiActions.showNotification(
        'データの更新に失敗しました。ネットワーク接続を確認してください。',
        'error',
        10000 // 10秒間表示
      );
      
      // リトライカウントをリセット（次回のポーリングで再試行可能にする）
      this.retryCount = 0;
      return;
    }

    // 再試行の実行
    const retryDelay = this.calculateRetryDelay();
    
    if (this.options.enableLogging) {
      console.log(`[Polling] Retrying in ${retryDelay}ms...`);
    }
    
    setTimeout(() => {
      if (this.isPolling) {
        this.executePoll();
      }
    }, retryDelay);
  }

  /**
   * 再試行間隔を計算（指数バックオフ対応）
   * @returns {number} 再試行間隔（ミリ秒）
   */
  calculateRetryDelay() {
    if (!this.options.exponentialBackoff) {
      return this.options.retryDelay;
    }
    
    // 指数バックオフ: delay * (2 ^ (retryCount - 1))
    const exponentialDelay = this.options.retryDelay * Math.pow(2, this.retryCount - 1);
    
    // 最大60秒に制限
    return Math.min(exponentialDelay, 60000);
  }

  /**
   * ポーリング間隔を動的に変更
   * @param {number} newInterval - 新しいポーリング間隔（ミリ秒）
   */
  setInterval(newInterval) {
    if (typeof newInterval !== 'number' || newInterval < 1000) {
      throw new Error('Interval must be a number >= 1000ms');
    }
    
    this.options.interval = newInterval;
    
    // 現在ポーリング中の場合は再起動
    if (this.isPolling) {
      this.pause();
      this.resume();
    }
    
    if (this.options.enableLogging) {
      console.log(`[Polling] Interval changed to ${newInterval}ms`);
    }
  }

  /**
   * 統計情報を取得
   * @returns {Object} ポーリングの統計情報
   */
  getStats() {
    return {
      isPolling: this.isPolling,
      isPageVisible: this.isPageVisible,
      isOnline: this.isOnline,
      interval: this.options.interval,
      retryCount: this.retryCount,
      lastSuccessTime: this.lastSuccessTime,
      lastErrorTime: this.lastErrorTime,
      callbackCount: this.callbacks.size,
      errorCallbackCount: this.errorCallbacks.size
    };
  }

  /**
   * リソースのクリーンアップ
   */
  destroy() {
    this.stop();
    this.callbacks.clear();
    this.errorCallbacks.clear();
    
    if (this.options.enableLogging) {
      console.log('[Polling] Destroyed');
    }
  }
}

/**
 * デフォルトのポーリングシステムインスタンス
 */
export const defaultPollingSystem = new PollingSystem({
  interval: 30000, // 30秒
  maxRetries: 3,
  retryDelay: 5000, // 5秒
  exponentialBackoff: true,
  respectVisibility: true,
  enableLogging: import.meta.env.DEV // 開発環境でのみログを有効化
});

/**
 * ポーリングシステムのファクトリー関数
 * @param {Object} options - ポーリングシステムのオプション
 * @returns {PollingSystem} 新しいポーリングシステムインスタンス
 */
export function createPollingSystem(options = {}) {
  return new PollingSystem(options);
}

/**
 * 簡単なポーリング関数（単発使用向け）
 * @param {Function} callback - 実行するコールバック関数
 * @param {Object} options - ポーリングオプション
 * @returns {Function} ポーリングを停止する関数
 */
export function startSimplePolling(callback, options = {}) {
  const pollingSystem = new PollingSystem(options);
  pollingSystem.registerCallback('simple', callback);
  pollingSystem.start();
  
  return () => pollingSystem.destroy();
}
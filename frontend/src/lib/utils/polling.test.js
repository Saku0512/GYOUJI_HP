// ポーリングシステムのテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { PollingSystem, createPollingSystem, startSimplePolling } from './polling.js';

// モック設定
vi.mock('../stores/ui.js', () => ({
  uiActions: {
    showNotification: vi.fn()
  }
}));

describe('PollingSystem', () => {
  let pollingSystem;
  let mockCallback;
  let mockErrorCallback;

  beforeEach(() => {
    // タイマーをモック化
    vi.useFakeTimers();
    
    // DOM APIをモック化
    Object.defineProperty(document, 'hidden', {
      writable: true,
      value: false
    });
    
    Object.defineProperty(navigator, 'onLine', {
      writable: true,
      value: true
    });

    // モックコールバック
    mockCallback = vi.fn().mockResolvedValue();
    mockErrorCallback = vi.fn().mockResolvedValue();

    // ポーリングシステムのインスタンス作成
    pollingSystem = new PollingSystem({
      interval: 1000,
      maxRetries: 2,
      retryDelay: 500,
      exponentialBackoff: false,
      enableLogging: false
    });
  });

  afterEach(() => {
    pollingSystem.destroy();
    vi.useRealTimers();
    vi.clearAllMocks();
    vi.resetAllMocks();
  });

  describe('基本機能', () => {
    it('コールバックを正しく登録できる', () => {
      pollingSystem.registerCallback('test', mockCallback);
      
      expect(pollingSystem.callbacks.has('test')).toBe(true);
      expect(pollingSystem.callbacks.get('test')).toBe(mockCallback);
    });

    it('エラーコールバックを正しく登録できる', () => {
      pollingSystem.registerErrorCallback('test', mockErrorCallback);
      
      expect(pollingSystem.errorCallbacks.has('test')).toBe(true);
      expect(pollingSystem.errorCallbacks.get('test')).toBe(mockErrorCallback);
    });

    it('コールバックを削除できる', () => {
      pollingSystem.registerCallback('test', mockCallback);
      const removed = pollingSystem.unregisterCallback('test');
      
      expect(removed).toBe(true);
      expect(pollingSystem.callbacks.has('test')).toBe(false);
    });

    it('存在しないコールバックの削除はfalseを返す', () => {
      const removed = pollingSystem.unregisterCallback('nonexistent');
      
      expect(removed).toBe(false);
    });

    it('無効なコールバックの登録はエラーを投げる', () => {
      expect(() => {
        pollingSystem.registerCallback('test', 'not a function');
      }).toThrow('Callback must be a function');
    });
  });

  describe('ポーリング制御', () => {
    it('ポーリングを開始できる', () => {
      pollingSystem.registerCallback('test', mockCallback);
      pollingSystem.start();
      
      expect(pollingSystem.isPolling).toBe(true);
      expect(pollingSystem.intervalId).not.toBeNull();
    });

    it('ポーリングを停止できる', () => {
      pollingSystem.registerCallback('test', mockCallback);
      pollingSystem.start();
      pollingSystem.stop();
      
      expect(pollingSystem.isPolling).toBe(false);
      expect(pollingSystem.intervalId).toBeNull();
    });

    it('既にポーリング中の場合は重複開始しない', () => {
      pollingSystem.registerCallback('test', mockCallback);
      pollingSystem.start();
      const firstIntervalId = pollingSystem.intervalId;
      
      pollingSystem.start(); // 2回目の開始
      
      expect(pollingSystem.intervalId).toBe(firstIntervalId);
    });

    it('ポーリングを一時停止・再開できる', () => {
      pollingSystem.registerCallback('test', mockCallback);
      pollingSystem.start();
      
      pollingSystem.pause();
      expect(pollingSystem.intervalId).toBeNull();
      expect(pollingSystem.isPolling).toBe(true);
      
      pollingSystem.resume();
      expect(pollingSystem.intervalId).not.toBeNull();
    });
  });

  describe('コールバック実行', () => {
    it('指定間隔でコールバックを実行する', async () => {
      pollingSystem.registerCallback('test', mockCallback);
      pollingSystem.start();
      
      // 最初は実行されない
      expect(mockCallback).toHaveBeenCalledTimes(0);
      
      // 1秒進める
      vi.advanceTimersByTime(1000);
      await vi.runOnlyPendingTimersAsync();
      expect(mockCallback).toHaveBeenCalledTimes(1);
      
      // さらに2秒進める
      vi.advanceTimersByTime(2000);
      await vi.runOnlyPendingTimersAsync();
      expect(mockCallback).toHaveBeenCalledTimes(3);
      
      pollingSystem.stop();
    });

    it('複数のコールバックを並列実行する', async () => {
      const mockCallback2 = vi.fn().mockResolvedValue();
      
      pollingSystem.registerCallback('test1', mockCallback);
      pollingSystem.registerCallback('test2', mockCallback2);
      pollingSystem.start();
      
      // 1秒進める
      vi.advanceTimersByTime(1000);
      await vi.runOnlyPendingTimersAsync();
      
      expect(mockCallback).toHaveBeenCalledTimes(1);
      expect(mockCallback2).toHaveBeenCalledTimes(1);
      
      pollingSystem.stop();
    });

    it('手動でポーリングを実行できる', async () => {
      pollingSystem.registerCallback('test', mockCallback);
      pollingSystem.isPolling = true; // 手動実行のためにポーリング状態を有効にする
      
      // 手動実行
      await pollingSystem.executePoll();
      
      expect(mockCallback).toHaveBeenCalledTimes(1);
    });
  });

  describe('エラーハンドリング', () => {
    it('コールバックエラー時にエラーコールバックを実行する', async () => {
      const errorCallback = vi.fn().mockRejectedValue(new Error('Test error'));
      
      pollingSystem.registerCallback('test', errorCallback);
      pollingSystem.registerErrorCallback('test', mockErrorCallback);
      pollingSystem.isPolling = true; // ポーリング状態を有効にする
      
      // 手動実行でエラーを発生させる
      await pollingSystem.executePoll();
      
      expect(mockErrorCallback).toHaveBeenCalledWith(expect.any(Error));
    });

    it('最大再試行回数まで再試行する', async () => {
      const errorCallback = vi.fn().mockRejectedValue(new Error('Test error'));
      
      pollingSystem.registerCallback('test', errorCallback);
      pollingSystem.isPolling = true; // ポーリング状態を有効にする
      
      // 手動実行でエラーを発生させる
      await pollingSystem.executePoll();
      expect(errorCallback).toHaveBeenCalledTimes(1);
      
      // 再試行1回目
      vi.advanceTimersByTime(500); // 再試行間隔
      await vi.runOnlyPendingTimersAsync();
      expect(errorCallback).toHaveBeenCalledTimes(2);
      
      // 再試行2回目
      vi.advanceTimersByTime(500);
      await vi.runOnlyPendingTimersAsync();
      expect(errorCallback).toHaveBeenCalledTimes(3);
    });

    it('指数バックオフが正しく動作する', () => {
      const exponentialPolling = new PollingSystem({
        interval: 1000,
        retryDelay: 1000,
        exponentialBackoff: true,
        enableLogging: false
      });
      
      exponentialPolling.retryCount = 1;
      expect(exponentialPolling.calculateRetryDelay()).toBe(1000);
      
      exponentialPolling.retryCount = 2;
      expect(exponentialPolling.calculateRetryDelay()).toBe(2000);
      
      exponentialPolling.retryCount = 3;
      expect(exponentialPolling.calculateRetryDelay()).toBe(4000);
      
      exponentialPolling.destroy();
    });
  });

  describe('ページ可視性制御', () => {
    it('ページが非表示の場合はポーリングをスキップする', async () => {
      Object.defineProperty(document, 'hidden', { value: true });
      pollingSystem.isPageVisible = false;
      
      pollingSystem.registerCallback('test', mockCallback);
      pollingSystem.start();
      
      // 1秒進める（非表示なのでスキップされる）
      vi.advanceTimersByTime(1000);
      await vi.runOnlyPendingTimersAsync();
      expect(mockCallback).toHaveBeenCalledTimes(0); // スキップされるので実行されない
      
      pollingSystem.stop();
    });

    it('respectVisibilityがfalseの場合は非表示でも実行する', async () => {
      const alwaysPolling = new PollingSystem({
        interval: 1000,
        respectVisibility: false,
        enableLogging: false
      });
      
      Object.defineProperty(document, 'hidden', { value: true });
      alwaysPolling.isPageVisible = false;
      
      alwaysPolling.registerCallback('test', mockCallback);
      alwaysPolling.start();
      
      // 1秒進める（respectVisibilityがfalseなので実行される）
      vi.advanceTimersByTime(1000);
      await vi.runOnlyPendingTimersAsync();
      expect(mockCallback).toHaveBeenCalledTimes(1);
      
      alwaysPolling.destroy();
    });
  });

  describe('ネットワーク状態制御', () => {
    it('オフライン時はポーリングをスキップする', async () => {
      Object.defineProperty(navigator, 'onLine', { value: false });
      pollingSystem.isOnline = false;
      
      pollingSystem.registerCallback('test', mockCallback);
      pollingSystem.start();
      
      // 1秒進める（オフラインなのでスキップされる）
      vi.advanceTimersByTime(1000);
      await vi.runOnlyPendingTimersAsync();
      expect(mockCallback).toHaveBeenCalledTimes(0); // スキップされるので実行されない
      
      pollingSystem.stop();
    });
  });

  describe('統計情報', () => {
    it('正しい統計情報を返す', () => {
      pollingSystem.registerCallback('test1', mockCallback);
      pollingSystem.registerCallback('test2', mockCallback);
      pollingSystem.registerErrorCallback('error1', mockErrorCallback);
      
      const stats = pollingSystem.getStats();
      
      expect(stats).toEqual({
        isPolling: false,
        isPageVisible: true,
        isOnline: true,
        interval: 1000,
        retryCount: 0,
        lastSuccessTime: null,
        lastErrorTime: null,
        callbackCount: 2,
        errorCallbackCount: 1
      });
    });
  });

  describe('間隔変更', () => {
    it('ポーリング間隔を動的に変更できる', () => {
      pollingSystem.setInterval(2000);
      
      expect(pollingSystem.options.interval).toBe(2000);
    });

    it('無効な間隔値はエラーを投げる', () => {
      expect(() => {
        pollingSystem.setInterval(500); // 1000ms未満
      }).toThrow('Interval must be a number >= 1000ms');
      
      expect(() => {
        pollingSystem.setInterval('invalid');
      }).toThrow('Interval must be a number >= 1000ms');
    });
  });

  describe('リソース管理', () => {
    it('destroyでリソースを正しくクリーンアップする', () => {
      pollingSystem.registerCallback('test', mockCallback);
      pollingSystem.registerErrorCallback('error', mockErrorCallback);
      pollingSystem.start();
      
      pollingSystem.destroy();
      
      expect(pollingSystem.isPolling).toBe(false);
      expect(pollingSystem.intervalId).toBeNull();
      expect(pollingSystem.callbacks.size).toBe(0);
      expect(pollingSystem.errorCallbacks.size).toBe(0);
    });
  });
});

describe('ファクトリー関数', () => {
  afterEach(() => {
    vi.useRealTimers();
  });

  it('createPollingSystemが正しくインスタンスを作成する', () => {
    const instance = createPollingSystem({
      interval: 5000,
      maxRetries: 5
    });
    
    expect(instance).toBeInstanceOf(PollingSystem);
    expect(instance.options.interval).toBe(5000);
    expect(instance.options.maxRetries).toBe(5);
    
    instance.destroy();
  });

  it('startSimplePollingが正しく動作する', async () => {
    vi.useFakeTimers();
    
    const mockCallback = vi.fn().mockResolvedValue();
    const stopPolling = startSimplePolling(mockCallback, {
      interval: 1000,
      enableLogging: false
    });
    
    // 1秒進める
    vi.advanceTimersByTime(1000);
    await vi.runOnlyPendingTimersAsync();
    
    expect(mockCallback).toHaveBeenCalledTimes(1);
    
    // 停止関数を呼び出し
    stopPolling();
    
    expect(typeof stopPolling).toBe('function');
  });
});
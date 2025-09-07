// ネットワークモニターの単体テスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { get } from 'svelte/store';
import {
  networkStore,
  networkMonitor,
  initializeNetworkMonitor,
  addToRetryQueue,
  getConnectionQuality,
  testConnection,
  executeWhenOnline
} from '../network-monitor.js';

describe('network-monitor', () => {
  let mockNavigator;
  let mockWindow;
  let mockFetch;

  beforeEach(() => {
    // Navigator のモック
    mockNavigator = {
      onLine: true,
      connection: {
        type: 'wifi',
        effectiveType: '4g',
        downlink: 10,
        rtt: 50,
        saveData: false,
        addEventListener: vi.fn()
      }
    };

    // Window のモック
    mockWindow = {
      addEventListener: vi.fn(),
      removeEventListener: vi.fn()
    };

    // Fetch のモック
    mockFetch = vi.fn();

    // グローバルオブジェクトの設定
    Object.defineProperty(global, 'navigator', {
      value: mockNavigator,
      writable: true
    });

    Object.defineProperty(global, 'window', {
      value: mockWindow,
      writable: true
    });

    global.fetch = mockFetch;

    // console のモック
    vi.spyOn(console, 'log').mockImplementation(() => {});
    vi.spyOn(console, 'warn').mockImplementation(() => {});
    vi.spyOn(console, 'error').mockImplementation(() => {});

    // ネットワークモニターをリセット
    networkMonitor.cleanup();
  });

  afterEach(() => {
    vi.restoreAllMocks();
    networkMonitor.cleanup();
  });

  describe('networkStore', () => {
    it('初期状態が正しく設定される', () => {
      const state = get(networkStore);
      
      expect(state.isOnline).toBe(true);
      expect(state.connectionType).toBe('unknown');
      expect(state.effectiveType).toBe('unknown');
      expect(state.downlink).toBe(0);
      expect(state.rtt).toBe(0);
      expect(state.saveData).toBe(false);
      expect(state.lastOnlineTime).toBeNull();
      expect(state.lastOfflineTime).toBeNull();
    });
  });

  describe('NetworkMonitor initialization', () => {
    it('正常に初期化される', () => {
      networkMonitor.initialize();

      expect(mockWindow.addEventListener).toHaveBeenCalledWith('online', expect.any(Function));
      expect(mockWindow.addEventListener).toHaveBeenCalledWith('offline', expect.any(Function));
      expect(mockNavigator.connection.addEventListener).toHaveBeenCalledWith('change', expect.any(Function));
    });

    it('重複初期化を防ぐ', () => {
      networkMonitor.initialize();
      networkMonitor.initialize();

      // 1回目の初期化のみ実行される
      expect(mockWindow.addEventListener).toHaveBeenCalledTimes(2); // online, offline
    });

    it('window が存在しない環境では初期化をスキップする', () => {
      Object.defineProperty(global, 'window', {
        value: undefined,
        writable: true
      });

      networkMonitor.initialize();

      // イベントリスナーが追加されないことを確認
      expect(mockWindow.addEventListener).not.toHaveBeenCalled();
    });
  });

  describe('オンライン/オフライン イベント', () => {
    beforeEach(() => {
      networkMonitor.initialize();
    });

    it('オンラインイベントを正しく処理する', () => {
      // オフライン状態から開始
      mockNavigator.onLine = false;
      networkMonitor.updateNetworkStatus(false);

      // オンラインイベントをシミュレート
      mockNavigator.onLine = true;
      const onlineHandler = mockWindow.addEventListener.mock.calls.find(
        call => call[0] === 'online'
      )[1];
      
      onlineHandler();

      const state = get(networkStore);
      expect(state.isOnline).toBe(true);
      expect(state.lastOnlineTime).toBeTruthy();
    });

    it('オフラインイベントを正しく処理する', () => {
      // オンライン状態から開始
      mockNavigator.onLine = true;
      networkMonitor.updateNetworkStatus(true);

      // オフラインイベントをシミュレート
      mockNavigator.onLine = false;
      const offlineHandler = mockWindow.addEventListener.mock.calls.find(
        call => call[0] === 'offline'
      )[1];
      
      offlineHandler();

      const state = get(networkStore);
      expect(state.isOnline).toBe(false);
      expect(state.lastOfflineTime).toBeTruthy();
    });
  });

  describe('接続情報の更新', () => {
    beforeEach(() => {
      networkMonitor.initialize();
    });

    it('接続情報を正しく更新する', () => {
      const connectionInfo = {
        type: 'cellular',
        effectiveType: '3g',
        downlink: 1.5,
        rtt: 200,
        saveData: true
      };

      networkMonitor.updateConnectionInfo(connectionInfo);

      const state = get(networkStore);
      expect(state.connectionType).toBe('cellular');
      expect(state.effectiveType).toBe('3g');
      expect(state.downlink).toBe(1.5);
      expect(state.rtt).toBe(200);
      expect(state.saveData).toBe(true);
    });
  });

  describe('リスナー管理', () => {
    beforeEach(() => {
      networkMonitor.initialize();
    });

    it('オンラインリスナーを正しく追加・削除する', () => {
      const listener = vi.fn();

      networkMonitor.addOnlineListener(listener);
      
      // オンラインイベントをトリガー
      const onlineHandler = mockWindow.addEventListener.mock.calls.find(
        call => call[0] === 'online'
      )[1];
      onlineHandler();

      expect(listener).toHaveBeenCalled();

      // リスナーを削除
      networkMonitor.removeOnlineListener(listener);
      listener.mockClear();

      // 再度オンラインイベントをトリガー
      onlineHandler();
      expect(listener).not.toHaveBeenCalled();
    });

    it('オフラインリスナーを正しく追加・削除する', () => {
      const listener = vi.fn();

      networkMonitor.addOfflineListener(listener);
      
      // オフラインイベントをトリガー
      const offlineHandler = mockWindow.addEventListener.mock.calls.find(
        call => call[0] === 'offline'
      )[1];
      offlineHandler();

      expect(listener).toHaveBeenCalled();

      // リスナーを削除
      networkMonitor.removeOfflineListener(listener);
      listener.mockClear();

      // 再度オフラインイベントをトリガー
      offlineHandler();
      expect(listener).not.toHaveBeenCalled();
    });
  });

  describe('再試行キュー', () => {
    beforeEach(() => {
      networkMonitor.initialize();
    });

    it('再試行キューに操作を追加できる', () => {
      const operation = vi.fn().mockResolvedValue('success');
      
      const retryId = networkMonitor.addToRetryQueue(operation);

      expect(retryId).toBeTruthy();
      expect(typeof retryId).toBe('string');
    });

    it('オンライン時に再試行キューを処理する', async () => {
      const operation = vi.fn().mockResolvedValue('success');
      
      networkMonitor.addToRetryQueue(operation);
      await networkMonitor.processRetryQueue();

      expect(operation).toHaveBeenCalled();
    });

    it('失敗した操作を再試行する', async () => {
      const operation = vi.fn()
        .mockRejectedValueOnce(new Error('First failure'))
        .mockResolvedValue('success');
      
      networkMonitor.addToRetryQueue(operation, { maxAttempts: 2 });
      await networkMonitor.processRetryQueue();

      expect(operation).toHaveBeenCalledTimes(2);
    });

    it('最大試行回数に達した操作を諦める', async () => {
      const operation = vi.fn().mockRejectedValue(new Error('Always fails'));
      
      networkMonitor.addToRetryQueue(operation, { maxAttempts: 2 });
      await networkMonitor.processRetryQueue();

      expect(operation).toHaveBeenCalledTimes(2);
    });
  });

  describe('接続品質評価', () => {
    beforeEach(() => {
      networkMonitor.initialize();
    });

    it('オフライン時は offline を返す', () => {
      networkMonitor.updateNetworkStatus(false);
      
      const quality = networkMonitor.getConnectionQuality();
      expect(quality).toBe('offline');
    });

    it('effectiveType に基づいて品質を判定する', () => {
      networkMonitor.updateNetworkStatus(true);
      
      // 4G接続
      networkMonitor.updateConnectionInfo({ effectiveType: '4g' });
      expect(networkMonitor.getConnectionQuality()).toBe('excellent');

      // 3G接続
      networkMonitor.updateConnectionInfo({ effectiveType: '3g' });
      expect(networkMonitor.getConnectionQuality()).toBe('good');

      // 2G接続
      networkMonitor.updateConnectionInfo({ effectiveType: '2g' });
      expect(networkMonitor.getConnectionQuality()).toBe('poor');
    });

    it('downlink に基づいて品質を判定する', () => {
      networkMonitor.updateNetworkStatus(true);
      networkMonitor.updateConnectionInfo({ effectiveType: 'unknown' });

      // 高速接続
      networkMonitor.updateConnectionInfo({ downlink: 15 });
      expect(networkMonitor.getConnectionQuality()).toBe('excellent');

      // 中速接続
      networkMonitor.updateConnectionInfo({ downlink: 2 });
      expect(networkMonitor.getConnectionQuality()).toBe('good');

      // 低速接続
      networkMonitor.updateConnectionInfo({ downlink: 0.3 });
      expect(networkMonitor.getConnectionQuality()).toBe('poor');
    });
  });

  describe('接続テスト', () => {
    it('成功時にtrueを返す', async () => {
      mockFetch.mockResolvedValue({ ok: true });

      const result = await networkMonitor.testConnection();

      expect(result).toBe(true);
      expect(mockFetch).toHaveBeenCalledWith('/api/health', {
        method: 'HEAD',
        signal: expect.any(AbortSignal),
        cache: 'no-cache'
      });
    });

    it('失敗時にfalseを返す', async () => {
      mockFetch.mockRejectedValue(new Error('Network error'));

      const result = await networkMonitor.testConnection();

      expect(result).toBe(false);
    });

    it('カスタムURLでテストできる', async () => {
      mockFetch.mockResolvedValue({ ok: true });

      await networkMonitor.testConnection('/custom/endpoint');

      expect(mockFetch).toHaveBeenCalledWith('/custom/endpoint', expect.any(Object));
    });
  });

  describe('便利関数', () => {
    describe('initializeNetworkMonitor', () => {
      it('ネットワークモニターを初期化する', () => {
        const initializeSpy = vi.spyOn(networkMonitor, 'initialize');
        
        initializeNetworkMonitor();

        expect(initializeSpy).toHaveBeenCalled();
      });
    });

    describe('addToRetryQueue', () => {
      it('再試行キューに操作を追加する', () => {
        const operation = vi.fn();
        const addSpy = vi.spyOn(networkMonitor, 'addToRetryQueue');
        
        addToRetryQueue(operation, { maxAttempts: 5 });

        expect(addSpy).toHaveBeenCalledWith(operation, { maxAttempts: 5 });
      });
    });

    describe('getConnectionQuality', () => {
      it('接続品質を取得する', () => {
        const qualitySpy = vi.spyOn(networkMonitor, 'getConnectionQuality');
        
        getConnectionQuality();

        expect(qualitySpy).toHaveBeenCalled();
      });
    });

    describe('testConnection', () => {
      it('接続テストを実行する', () => {
        const testSpy = vi.spyOn(networkMonitor, 'testConnection');
        
        testConnection('/test');

        expect(testSpy).toHaveBeenCalledWith('/test');
      });
    });
  });

  describe('executeWhenOnline', () => {
    it('オンライン時に操作を即座に実行する', async () => {
      mockNavigator.onLine = true;
      const operation = vi.fn().mockResolvedValue('success');

      const result = await executeWhenOnline(operation);

      expect(operation).toHaveBeenCalled();
      expect(result).toBe('success');
    });

    it('オフライン時にエラーを投げる', async () => {
      mockNavigator.onLine = false;
      const operation = vi.fn();

      await expect(executeWhenOnline(operation)).rejects.toThrow('オフライン状態です');
      expect(operation).not.toHaveBeenCalled();
    });

    it('ネットワークエラー時に再試行キューに追加する', async () => {
      mockNavigator.onLine = true;
      const operation = vi.fn().mockRejectedValue(new TypeError('Failed to fetch'));
      const addSpy = vi.spyOn(networkMonitor, 'addToRetryQueue');

      await expect(executeWhenOnline(operation, { retryOnReconnect: true })).rejects.toThrow();
      expect(addSpy).toHaveBeenCalledWith(operation, { retryOnReconnect: true });
    });
  });

  describe('クリーンアップ', () => {
    it('正しくクリーンアップされる', () => {
      networkMonitor.initialize();
      networkMonitor.cleanup();

      expect(mockWindow.removeEventListener).toHaveBeenCalledWith('online', expect.any(Function));
      expect(mockWindow.removeEventListener).toHaveBeenCalledWith('offline', expect.any(Function));
    });
  });
});
// UIストアの単体テスト
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { get } from 'svelte/store';
import { 
  uiStore, 
  uiActions, 
  showSuccessNotification,
  showErrorNotification,
  showWarningNotification,
  showInfoNotification
} from '../ui.js';

describe('UIストア', () => {
  beforeEach(() => {
    // 各テスト前にストアをリセット
    uiActions.reset();
    // タイマーをモック
    vi.useFakeTimers();
    // localStorageをモック
    Object.defineProperty(window, 'localStorage', {
      value: {
        getItem: vi.fn(),
        setItem: vi.fn(),
        removeItem: vi.fn(),
        clear: vi.fn(),
      },
      writable: true,
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  describe('初期状態', () => {
    it('正しい初期状態を持つ', () => {
      const state = get(uiStore);
      expect(state).toEqual({
        notifications: [],
        loading: false,
        theme: 'light'
      });
    });
  });

  describe('通知システム', () => {
    it('通知を追加できる', () => {
      const notificationId = uiActions.showNotification('テストメッセージ', 'info');
      const state = get(uiStore);
      
      expect(state.notifications).toHaveLength(1);
      expect(state.notifications[0]).toMatchObject({
        id: notificationId,
        message: 'テストメッセージ',
        type: 'info'
      });
      expect(state.notifications[0].timestamp).toBeTypeOf('number');
    });

    it('複数の通知を追加できる', () => {
      uiActions.showNotification('メッセージ1', 'info');
      uiActions.showNotification('メッセージ2', 'success');
      uiActions.showNotification('メッセージ3', 'error');
      
      const state = get(uiStore);
      expect(state.notifications).toHaveLength(3);
      expect(state.notifications[0].message).toBe('メッセージ1');
      expect(state.notifications[1].message).toBe('メッセージ2');
      expect(state.notifications[2].message).toBe('メッセージ3');
    });

    it('通知にユニークなIDが割り当てられる', () => {
      const id1 = uiActions.showNotification('メッセージ1');
      const id2 = uiActions.showNotification('メッセージ2');
      
      expect(id1).not.toBe(id2);
      expect(typeof id1).toBe('number');
      expect(typeof id2).toBe('number');
    });

    it('デフォルトの通知タイプは"info"', () => {
      uiActions.showNotification('テストメッセージ');
      const state = get(uiStore);
      
      expect(state.notifications[0].type).toBe('info');
    });

    it('特定の通知を削除できる', () => {
      const id1 = uiActions.showNotification('メッセージ1');
      const id2 = uiActions.showNotification('メッセージ2');
      const id3 = uiActions.showNotification('メッセージ3');
      
      uiActions.removeNotification(id2);
      
      const state = get(uiStore);
      expect(state.notifications).toHaveLength(2);
      expect(state.notifications.find(n => n.id === id1)).toBeDefined();
      expect(state.notifications.find(n => n.id === id2)).toBeUndefined();
      expect(state.notifications.find(n => n.id === id3)).toBeDefined();
    });

    it('存在しない通知IDを削除しても問題ない', () => {
      uiActions.showNotification('メッセージ1');
      const initialState = get(uiStore);
      
      uiActions.removeNotification(999);
      
      const finalState = get(uiStore);
      expect(finalState.notifications).toEqual(initialState.notifications);
    });

    it('全ての通知をクリアできる', () => {
      uiActions.showNotification('メッセージ1');
      uiActions.showNotification('メッセージ2');
      uiActions.showNotification('メッセージ3');
      
      uiActions.clearNotifications();
      
      const state = get(uiStore);
      expect(state.notifications).toHaveLength(0);
    });

    it('自動消去が設定された通知は指定時間後に削除される', () => {
      const notificationId = uiActions.showNotification('自動消去テスト', 'info', 3000);
      
      // 通知が追加されていることを確認
      let state = get(uiStore);
      expect(state.notifications).toHaveLength(1);
      expect(state.notifications[0].id).toBe(notificationId);
      
      // 3秒経過
      vi.advanceTimersByTime(3000);
      
      // 通知が削除されていることを確認
      state = get(uiStore);
      expect(state.notifications).toHaveLength(0);
    });

    it('duration=0の通知は自動消去されない', () => {
      uiActions.showNotification('永続通知', 'info', 0);
      
      // 長時間経過
      vi.advanceTimersByTime(10000);
      
      const state = get(uiStore);
      expect(state.notifications).toHaveLength(1);
    });
  });

  describe('ローディング状態', () => {
    it('ローディング状態をtrueに設定できる', () => {
      uiActions.setLoading(true);
      const state = get(uiStore);
      
      expect(state.loading).toBe(true);
    });

    it('ローディング状態をfalseに設定できる', () => {
      uiActions.setLoading(true);
      uiActions.setLoading(false);
      const state = get(uiStore);
      
      expect(state.loading).toBe(false);
    });

    it('非boolean値もboolean値に変換される', () => {
      uiActions.setLoading('true');
      let state = get(uiStore);
      expect(state.loading).toBe(true);
      
      uiActions.setLoading(0);
      state = get(uiStore);
      expect(state.loading).toBe(false);
      
      uiActions.setLoading(1);
      state = get(uiStore);
      expect(state.loading).toBe(true);
    });
  });

  describe('テーマ管理', () => {
    it('有効なテーマを設定できる', () => {
      uiActions.setTheme('dark');
      let state = get(uiStore);
      expect(state.theme).toBe('dark');
      
      uiActions.setTheme('light');
      state = get(uiStore);
      expect(state.theme).toBe('light');
    });

    it('無効なテーマは設定されない', () => {
      const consoleSpy = vi.spyOn(console, 'warn').mockImplementation(() => {});
      
      uiActions.setTheme('invalid');
      const state = get(uiStore);
      
      expect(state.theme).toBe('light'); // 初期値のまま
      expect(consoleSpy).toHaveBeenCalledWith('Invalid theme. Only "light" and "dark" are supported.');
      
      consoleSpy.mockRestore();
    });

    it('テーマ設定時にlocalStorageに保存される', () => {
      uiActions.setTheme('dark');
      
      expect(localStorage.setItem).toHaveBeenCalledWith('ui-theme', 'dark');
    });

    it('localStorageからテーマを読み込める', () => {
      localStorage.getItem.mockReturnValue('dark');
      
      uiActions.loadTheme();
      const state = get(uiStore);
      
      expect(state.theme).toBe('dark');
      expect(localStorage.getItem).toHaveBeenCalledWith('ui-theme');
    });

    it('無効なテーマがlocalStorageにある場合は読み込まない', () => {
      localStorage.getItem.mockReturnValue('invalid');
      
      uiActions.loadTheme();
      const state = get(uiStore);
      
      expect(state.theme).toBe('light'); // 初期値のまま
    });

    it('localStorageが利用できない環境でもエラーにならない', () => {
      // localStorageを一時的に無効化
      const originalLocalStorage = window.localStorage;
      Object.defineProperty(window, 'localStorage', {
        value: undefined,
        writable: true,
      });
      
      expect(() => {
        uiActions.setTheme('dark');
        uiActions.loadTheme();
      }).not.toThrow();
      
      // 元に戻す
      Object.defineProperty(window, 'localStorage', {
        value: originalLocalStorage,
        writable: true,
      });
    });
  });

  describe('リセット機能', () => {
    it('ストアを初期状態にリセットできる', () => {
      // 状態を変更
      uiActions.showNotification('テスト');
      uiActions.setLoading(true);
      uiActions.setTheme('dark');
      
      // リセット
      uiActions.reset();
      
      const state = get(uiStore);
      expect(state).toEqual({
        notifications: [],
        loading: false,
        theme: 'light'
      });
    });
  });

  describe('ヘルパー関数', () => {
    it('showSuccessNotificationが正しく動作する', () => {
      showSuccessNotification('成功メッセージ');
      const state = get(uiStore);
      
      expect(state.notifications).toHaveLength(1);
      expect(state.notifications[0].message).toBe('成功メッセージ');
      expect(state.notifications[0].type).toBe('success');
    });

    it('showErrorNotificationが正しく動作する', () => {
      showErrorNotification('エラーメッセージ');
      const state = get(uiStore);
      
      expect(state.notifications).toHaveLength(1);
      expect(state.notifications[0].message).toBe('エラーメッセージ');
      expect(state.notifications[0].type).toBe('error');
    });

    it('showWarningNotificationが正しく動作する', () => {
      showWarningNotification('警告メッセージ');
      const state = get(uiStore);
      
      expect(state.notifications).toHaveLength(1);
      expect(state.notifications[0].message).toBe('警告メッセージ');
      expect(state.notifications[0].type).toBe('warning');
    });

    it('showInfoNotificationが正しく動作する', () => {
      showInfoNotification('情報メッセージ');
      const state = get(uiStore);
      
      expect(state.notifications).toHaveLength(1);
      expect(state.notifications[0].message).toBe('情報メッセージ');
      expect(state.notifications[0].type).toBe('info');
    });

    it('ヘルパー関数でdurationを指定できる', () => {
      showSuccessNotification('テスト', 1000);
      
      // 1秒経過
      vi.advanceTimersByTime(1000);
      
      const state = get(uiStore);
      expect(state.notifications).toHaveLength(0);
    });
  });

  describe('ストアの購読', () => {
    it('ストアの変更を購読できる', () => {
      const mockCallback = vi.fn();
      const unsubscribe = uiStore.subscribe(mockCallback);
      
      // 初期状態で1回呼ばれる
      expect(mockCallback).toHaveBeenCalledTimes(1);
      
      // 状態を変更
      uiActions.showNotification('テスト');
      expect(mockCallback).toHaveBeenCalledTimes(2);
      
      uiActions.setLoading(true);
      expect(mockCallback).toHaveBeenCalledTimes(3);
      
      // 購読解除
      unsubscribe();
    });
  });
});
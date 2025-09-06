// UI状態管理ストア
import { writable } from 'svelte/store';

// UI状態の初期値
const initialUIState = {
  notifications: [],
  loading: false,
  theme: 'light'
};

// UIストアの作成
export const uiStore = writable(initialUIState);

// 通知のユニークIDを生成するためのカウンター
let notificationIdCounter = 0;

// UI関連のアクション
export const uiActions = {
  /**
   * 通知を表示する
   * @param {string} message - 通知メッセージ
   * @param {string} type - 通知タイプ ('success', 'error', 'warning', 'info')
   * @param {number} duration - 自動消去までの時間（ミリ秒）、0の場合は自動消去しない
   */
  showNotification: (message, type = 'info', duration = 5000) => {
    const notification = {
      id: ++notificationIdCounter,
      message,
      type,
      timestamp: Date.now()
    };

    uiStore.update(state => ({
      ...state,
      notifications: [...state.notifications, notification]
    }));

    // 自動消去の設定
    if (duration > 0) {
      setTimeout(() => {
        uiActions.removeNotification(notification.id);
      }, duration);
    }

    return notification.id;
  },

  /**
   * 特定の通知を削除する
   * @param {number} id - 削除する通知のID
   */
  removeNotification: (id) => {
    uiStore.update(state => ({
      ...state,
      notifications: state.notifications.filter(notification => notification.id !== id)
    }));
  },

  /**
   * ローディング状態を設定する
   * @param {boolean} state - ローディング状態
   */
  setLoading: (state) => {
    uiStore.update(currentState => ({
      ...currentState,
      loading: Boolean(state)
    }));
  },

  /**
   * 全ての通知をクリアする
   */
  clearNotifications: () => {
    uiStore.update(state => ({
      ...state,
      notifications: []
    }));
  },

  /**
   * テーマを設定する
   * @param {string} theme - テーマ ('light', 'dark')
   */
  setTheme: (theme) => {
    if (theme !== 'light' && theme !== 'dark') {
      console.warn('Invalid theme. Only "light" and "dark" are supported.');
      return;
    }

    uiStore.update(state => ({
      ...state,
      theme
    }));

    // ローカルストレージにテーマを保存
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem('ui-theme', theme);
    }
  },

  /**
   * ローカルストレージからテーマを読み込む
   */
  loadTheme: () => {
    if (typeof localStorage !== 'undefined') {
      const savedTheme = localStorage.getItem('ui-theme');
      if (savedTheme && (savedTheme === 'light' || savedTheme === 'dark')) {
        uiActions.setTheme(savedTheme);
      }
    }
  },

  /**
   * UIストアを初期状態にリセットする
   */
  reset: () => {
    uiStore.set(initialUIState);
  }
};

// 便利なヘルパー関数
export const showSuccessNotification = (message, duration) => 
  uiActions.showNotification(message, 'success', duration);

export const showErrorNotification = (message, duration) => 
  uiActions.showNotification(message, 'error', duration);

export const showWarningNotification = (message, duration) => 
  uiActions.showNotification(message, 'warning', duration);

export const showInfoNotification = (message, duration) => 
  uiActions.showNotification(message, 'info', duration);

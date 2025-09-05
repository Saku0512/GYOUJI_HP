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

// UI関連のアクション（後のタスクで実装）
export const uiActions = {
  showNotification: (message, type = 'info') => {
    // 通知表示は後のタスクで実装
    console.log('Show notification will be implemented in later tasks');
  },

  setLoading: (state) => {
    // ローディング状態設定は後のタスクで実装
    console.log('Set loading will be implemented in later tasks');
  },

  clearNotifications: () => {
    // 通知クリアは後のタスクで実装
    console.log('Clear notifications will be implemented in later tasks');
  }
};

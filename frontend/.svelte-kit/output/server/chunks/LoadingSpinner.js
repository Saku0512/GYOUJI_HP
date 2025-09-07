import { w as writable } from "./index.js";
import { E as fallback, F as attr_class, S as attr_style, I as bind_props, J as stringify } from "./index2.js";
/* empty css                                             */
const initialUIState = {
  notifications: [],
  loading: false,
  theme: "light"
};
const uiStore = writable(initialUIState);
let notificationIdCounter = 0;
const uiActions = {
  /**
   * 通知を表示する
   * @param {string} message - 通知メッセージ
   * @param {string} type - 通知タイプ ('success', 'error', 'warning', 'info')
   * @param {number} duration - 自動消去までの時間（ミリ秒）、0の場合は自動消去しない
   */
  showNotification: (message, type = "info", duration = 5e3) => {
    const notification = {
      id: ++notificationIdCounter,
      message,
      type,
      timestamp: Date.now()
    };
    uiStore.update((state) => ({
      ...state,
      notifications: [...state.notifications, notification]
    }));
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
    uiStore.update((state) => ({
      ...state,
      notifications: state.notifications.filter((notification) => notification.id !== id)
    }));
  },
  /**
   * ローディング状態を設定する
   * @param {boolean} state - ローディング状態
   */
  setLoading: (state) => {
    uiStore.update((currentState) => ({
      ...currentState,
      loading: Boolean(state)
    }));
  },
  /**
   * 全ての通知をクリアする
   */
  clearNotifications: () => {
    uiStore.update((state) => ({
      ...state,
      notifications: []
    }));
  },
  /**
   * テーマを設定する
   * @param {string} theme - テーマ ('light', 'dark')
   */
  setTheme: (theme) => {
    if (theme !== "light" && theme !== "dark") {
      console.warn('Invalid theme. Only "light" and "dark" are supported.');
      return;
    }
    uiStore.update((state) => ({
      ...state,
      theme
    }));
    if (typeof localStorage !== "undefined") {
      localStorage.setItem("ui-theme", theme);
    }
  },
  /**
   * ローカルストレージからテーマを読み込む
   */
  loadTheme: () => {
    if (typeof localStorage !== "undefined") {
      const savedTheme = localStorage.getItem("ui-theme");
      if (savedTheme && (savedTheme === "light" || savedTheme === "dark")) {
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
function LoadingSpinner($$payload, $$props) {
  let size = fallback(
    $$props["size"],
    "medium"
    // 'small', 'medium', 'large'
  );
  let color = fallback($$props["color"], "#007bff");
  $$payload.out.push(`<div class="spinner-container svelte-1yjjzjh"><div${attr_class(`spinner ${stringify(size)}`, "svelte-1yjjzjh")}${attr_style(`border-top-color: ${stringify(color)}`)}></div></div>`);
  bind_props($$props, { size, color });
}
export {
  LoadingSpinner as L,
  uiActions as a,
  uiStore as u
};

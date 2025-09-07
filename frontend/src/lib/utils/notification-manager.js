// 高度な通知管理システム
import { uiActions } from '$lib/stores/ui.js';

// 通知テンプレート
export const NOTIFICATION_TEMPLATES = {
  // 成功通知
  SUCCESS: {
    type: 'success',
    duration: 4000,
    dismissible: true
  },
  
  // エラー通知
  ERROR: {
    type: 'error',
    duration: 8000,
    dismissible: true
  },
  
  // 警告通知
  WARNING: {
    type: 'warning',
    duration: 6000,
    dismissible: true
  },
  
  // 情報通知
  INFO: {
    type: 'info',
    duration: 5000,
    dismissible: true
  },
  
  // 永続通知（手動で閉じるまで表示）
  PERSISTENT: {
    type: 'info',
    duration: 0,
    dismissible: true
  },
  
  // システム通知（重要、長時間表示）
  SYSTEM: {
    type: 'warning',
    duration: 10000,
    dismissible: true
  }
};

// 通知カテゴリ
export const NOTIFICATION_CATEGORIES = {
  AUTH: 'authentication',
  NETWORK: 'network',
  DATA: 'data',
  SYSTEM: 'system',
  USER_ACTION: 'user_action',
  VALIDATION: 'validation'
};

// 通知管理クラス
class NotificationManager {
  constructor() {
    this.activeNotifications = new Map();
    this.notificationQueue = [];
    this.isProcessingQueue = false;
    this.rateLimitMap = new Map();
    this.defaultRateLimit = {
      maxNotifications: 3,
      timeWindow: 5000 // 5秒
    };
  }

  // 基本通知表示
  show(message, options = {}) {
    const config = {
      ...NOTIFICATION_TEMPLATES.INFO,
      ...options,
      message,
      id: options.id || this.generateId(),
      timestamp: Date.now(),
      category: options.category || NOTIFICATION_CATEGORIES.SYSTEM
    };

    // レート制限チェック
    if (this.isRateLimited(config)) {
      console.warn('[NotificationManager] Rate limit exceeded for notification:', message);
      return null;
    }

    // 重複チェック
    if (this.isDuplicate(config)) {
      console.log('[NotificationManager] Duplicate notification ignored:', message);
      return this.activeNotifications.get(config.id);
    }

    // 通知を表示
    const notificationId = this.displayNotification(config);
    
    // アクティブ通知として記録
    this.activeNotifications.set(config.id, {
      ...config,
      notificationId,
      displayTime: Date.now()
    });

    return notificationId;
  }

  // 成功通知
  success(message, options = {}) {
    return this.show(message, {
      ...NOTIFICATION_TEMPLATES.SUCCESS,
      ...options,
      category: options.category || NOTIFICATION_CATEGORIES.USER_ACTION
    });
  }

  // エラー通知
  error(message, options = {}) {
    return this.show(message, {
      ...NOTIFICATION_TEMPLATES.ERROR,
      ...options,
      category: options.category || NOTIFICATION_CATEGORIES.SYSTEM
    });
  }

  // 警告通知
  warning(message, options = {}) {
    return this.show(message, {
      ...NOTIFICATION_TEMPLATES.WARNING,
      ...options,
      category: options.category || NOTIFICATION_CATEGORIES.SYSTEM
    });
  }

  // 情報通知
  info(message, options = {}) {
    return this.show(message, {
      ...NOTIFICATION_TEMPLATES.INFO,
      ...options,
      category: options.category || NOTIFICATION_CATEGORIES.SYSTEM
    });
  }

  // 永続通知
  persistent(message, options = {}) {
    return this.show(message, {
      ...NOTIFICATION_TEMPLATES.PERSISTENT,
      ...options
    });
  }

  // アクション付き通知
  withActions(message, actions, options = {}) {
    return this.show(message, {
      ...options,
      actions: actions.map(action => ({
        ...action,
        onClick: () => {
          try {
            action.onClick();
          } catch (error) {
            console.error('[NotificationManager] Action error:', error);
            this.error('アクションの実行中にエラーが発生しました');
          }
        }
      }))
    });
  }

  // 確認通知（Yes/No アクション付き）
  confirm(message, onConfirm, onCancel, options = {}) {
    const actions = [
      {
        label: options.confirmLabel || '確認',
        onClick: onConfirm,
        variant: 'primary'
      },
      {
        label: options.cancelLabel || 'キャンセル',
        onClick: onCancel,
        variant: 'secondary'
      }
    ];

    return this.withActions(message, actions, {
      ...NOTIFICATION_TEMPLATES.PERSISTENT,
      ...options,
      type: 'warning'
    });
  }

  // プログレス通知
  progress(message, options = {}) {
    const progressId = this.generateId();
    
    const notificationId = this.show(message, {
      ...NOTIFICATION_TEMPLATES.PERSISTENT,
      ...options,
      id: progressId,
      showProgress: true,
      type: 'info'
    });

    return {
      id: notificationId,
      update: (newMessage, progress) => {
        // プログレス更新の実装
        // 実際の実装では、通知の内容を動的に更新する機能が必要
        console.log(`[NotificationManager] Progress update: ${newMessage} (${progress}%)`);
      },
      complete: (successMessage) => {
        this.remove(notificationId);
        if (successMessage) {
          this.success(successMessage);
        }
      },
      error: (errorMessage) => {
        this.remove(notificationId);
        this.error(errorMessage);
      }
    };
  }

  // 通知の削除
  remove(notificationId) {
    uiActions.removeNotification(notificationId);
    
    // アクティブ通知から削除
    for (const [id, notification] of this.activeNotifications.entries()) {
      if (notification.notificationId === notificationId) {
        this.activeNotifications.delete(id);
        break;
      }
    }
  }

  // カテゴリ別通知削除
  removeByCategory(category) {
    const toRemove = [];
    
    for (const [id, notification] of this.activeNotifications.entries()) {
      if (notification.category === category) {
        toRemove.push(notification.notificationId);
        this.activeNotifications.delete(id);
      }
    }
    
    toRemove.forEach(notificationId => {
      uiActions.removeNotification(notificationId);
    });
  }

  // 全通知削除
  clear() {
    uiActions.clearNotifications();
    this.activeNotifications.clear();
  }

  // 通知の表示
  displayNotification(config) {
    return uiActions.showNotification(
      config.message,
      config.type,
      config.duration
    );
  }

  // ID生成
  generateId() {
    return `notification-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;
  }

  // レート制限チェック
  isRateLimited(config) {
    const key = `${config.category}-${config.type}`;
    const now = Date.now();
    const rateLimit = config.rateLimit || this.defaultRateLimit;
    
    if (!this.rateLimitMap.has(key)) {
      this.rateLimitMap.set(key, []);
    }
    
    const timestamps = this.rateLimitMap.get(key);
    
    // 古いタイムスタンプを削除
    const validTimestamps = timestamps.filter(
      timestamp => now - timestamp < rateLimit.timeWindow
    );
    
    this.rateLimitMap.set(key, validTimestamps);
    
    // 制限チェック
    if (validTimestamps.length >= rateLimit.maxNotifications) {
      return true;
    }
    
    // 新しいタイムスタンプを追加
    validTimestamps.push(now);
    this.rateLimitMap.set(key, validTimestamps);
    
    return false;
  }

  // 重複チェック
  isDuplicate(config) {
    if (!config.preventDuplicates) {
      return false;
    }
    
    const duplicateKey = config.duplicateKey || config.message;
    
    for (const notification of this.activeNotifications.values()) {
      const existingKey = notification.duplicateKey || notification.message;
      if (existingKey === duplicateKey) {
        return true;
      }
    }
    
    return false;
  }

  // 統計情報取得
  getStats() {
    const stats = {
      total: this.activeNotifications.size,
      byType: {},
      byCategory: {},
      averageDisplayTime: 0
    };
    
    let totalDisplayTime = 0;
    const now = Date.now();
    
    for (const notification of this.activeNotifications.values()) {
      // タイプ別統計
      stats.byType[notification.type] = (stats.byType[notification.type] || 0) + 1;
      
      // カテゴリ別統計
      stats.byCategory[notification.category] = (stats.byCategory[notification.category] || 0) + 1;
      
      // 表示時間統計
      totalDisplayTime += now - notification.displayTime;
    }
    
    if (stats.total > 0) {
      stats.averageDisplayTime = totalDisplayTime / stats.total;
    }
    
    return stats;
  }

  // デバッグ情報
  debug() {
    console.group('[NotificationManager] Debug Info');
    console.log('Active notifications:', this.activeNotifications.size);
    console.log('Rate limit map:', this.rateLimitMap);
    console.log('Stats:', this.getStats());
    console.groupEnd();
  }
}

// グローバルインスタンス
export const notificationManager = new NotificationManager();

// 便利な関数をエクスポート
export const showNotification = (message, options) => notificationManager.show(message, options);
export const showSuccess = (message, options) => notificationManager.success(message, options);
export const showError = (message, options) => notificationManager.error(message, options);
export const showWarning = (message, options) => notificationManager.warning(message, options);
export const showInfo = (message, options) => notificationManager.info(message, options);
export const showPersistent = (message, options) => notificationManager.persistent(message, options);
export const showWithActions = (message, actions, options) => notificationManager.withActions(message, actions, options);
export const showConfirm = (message, onConfirm, onCancel, options) => notificationManager.confirm(message, onConfirm, onCancel, options);
export const showProgress = (message, options) => notificationManager.progress(message, options);

// 特定用途の通知関数
export const showAuthError = (message = '認証エラーが発生しました') => {
  return notificationManager.error(message, {
    category: NOTIFICATION_CATEGORIES.AUTH,
    actions: [
      {
        label: 'ログイン',
        onClick: () => window.location.href = '/login',
        variant: 'primary'
      }
    ]
  });
};

export const showNetworkError = (message = 'ネットワークエラーが発生しました') => {
  return notificationManager.error(message, {
    category: NOTIFICATION_CATEGORIES.NETWORK,
    actions: [
      {
        label: '再試行',
        onClick: () => window.location.reload(),
        variant: 'primary'
      }
    ]
  });
};

export const showValidationError = (message, field) => {
  return notificationManager.warning(message, {
    category: NOTIFICATION_CATEGORIES.VALIDATION,
    duration: 6000,
    duplicateKey: `validation-${field}`,
    preventDuplicates: true
  });
};

export const showDataSaved = (message = 'データが保存されました') => {
  return notificationManager.success(message, {
    category: NOTIFICATION_CATEGORIES.DATA,
    duration: 3000
  });
};

export const showSystemMaintenance = (message = 'システムメンテナンス中です') => {
  return notificationManager.warning(message, {
    category: NOTIFICATION_CATEGORIES.SYSTEM,
    duration: 0, // 永続表示
    preventDuplicates: true,
    duplicateKey: 'system-maintenance'
  });
};
// NotificationManager のテスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { 
  notificationManager,
  showNotification,
  showSuccess,
  showError,
  showWarning,
  showInfo,
  showConfirm,
  showAuthError,
  showNetworkError,
  NOTIFICATION_TEMPLATES,
  NOTIFICATION_CATEGORIES
} from '../notification-manager.js';

// UIストアのモック
vi.mock('$lib/stores/ui.js', () => ({
  uiActions: {
    showNotification: vi.fn((message, type, duration) => {
      return `notification-${Date.now()}-${Math.random()}`;
    }),
    removeNotification: vi.fn(),
    clearNotifications: vi.fn()
  }
}));

describe('NotificationManager', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    notificationManager.activeNotifications.clear();
    notificationManager.rateLimitMap.clear();
  });

  describe('Basic Notification Display', () => {
    it('should show basic notification', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      const result = notificationManager.show('Test message');
      
      expect(uiActions.showNotification).toHaveBeenCalledWith(
        'Test message',
        'info',
        5000
      );
      expect(result).toBeTruthy();
      expect(notificationManager.activeNotifications.size).toBe(1);
    });

    it('should show notification with custom options', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      const options = {
        type: 'success',
        duration: 3000,
        category: NOTIFICATION_CATEGORIES.USER_ACTION
      };
      
      notificationManager.show('Success message', options);
      
      expect(uiActions.showNotification).toHaveBeenCalledWith(
        'Success message',
        'success',
        3000
      );
    });

    it('should generate unique IDs for notifications', () => {
      const id1 = notificationManager.generateId();
      const id2 = notificationManager.generateId();
      
      expect(id1).not.toBe(id2);
      expect(id1).toMatch(/^notification-\d+-[a-z0-9]+$/);
    });
  });

  describe('Notification Types', () => {
    it('should show success notification', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      notificationManager.success('Success message');
      
      expect(uiActions.showNotification).toHaveBeenCalledWith(
        'Success message',
        'success',
        4000
      );
    });

    it('should show error notification', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      notificationManager.error('Error message');
      
      expect(uiActions.showNotification).toHaveBeenCalledWith(
        'Error message',
        'error',
        8000
      );
    });

    it('should show warning notification', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      notificationManager.warning('Warning message');
      
      expect(uiActions.showNotification).toHaveBeenCalledWith(
        'Warning message',
        'warning',
        6000
      );
    });

    it('should show info notification', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      notificationManager.info('Info message');
      
      expect(uiActions.showNotification).toHaveBeenCalledWith(
        'Info message',
        'info',
        5000
      );
    });

    it('should show persistent notification', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      notificationManager.persistent('Persistent message');
      
      expect(uiActions.showNotification).toHaveBeenCalledWith(
        'Persistent message',
        'info',
        0
      );
    });
  });

  describe('Actions and Confirmation', () => {
    it('should show notification with actions', () => {
      const action1 = vi.fn();
      const action2 = vi.fn();
      
      const actions = [
        { label: 'Action 1', onClick: action1, variant: 'primary' },
        { label: 'Action 2', onClick: action2, variant: 'secondary' }
      ];
      
      notificationManager.withActions('Message with actions', actions);
      
      expect(notificationManager.activeNotifications.size).toBe(1);
      
      // アクションが適切にラップされていることを確認
      const notification = Array.from(notificationManager.activeNotifications.values())[0];
      expect(notification.actions).toHaveLength(2);
      expect(notification.actions[0].label).toBe('Action 1');
    });

    it('should handle action errors gracefully', () => {
      const consoleSpy = vi.spyOn(console, 'error').mockImplementation(() => {});
      const errorAction = vi.fn().mockImplementation(() => {
        throw new Error('Action failed');
      });
      
      const actions = [
        { label: 'Error Action', onClick: errorAction }
      ];
      
      notificationManager.withActions('Message', actions);
      
      const notification = Array.from(notificationManager.activeNotifications.values())[0];
      
      // アクションを実行してエラーハンドリングをテスト
      notification.actions[0].onClick();
      
      expect(errorAction).toHaveBeenCalled();
      expect(consoleSpy).toHaveBeenCalledWith('[NotificationManager] Action error:', expect.any(Error));
      
      consoleSpy.mockRestore();
    });

    it('should show confirmation notification', () => {
      const onConfirm = vi.fn();
      const onCancel = vi.fn();
      
      notificationManager.confirm('Confirm this action?', onConfirm, onCancel);
      
      const notification = Array.from(notificationManager.activeNotifications.values())[0];
      expect(notification.actions).toHaveLength(2);
      expect(notification.actions[0].label).toBe('確認');
      expect(notification.actions[1].label).toBe('キャンセル');
      expect(notification.type).toBe('warning');
      expect(notification.duration).toBe(0);
    });

    it('should show confirmation with custom labels', () => {
      const onConfirm = vi.fn();
      const onCancel = vi.fn();
      
      notificationManager.confirm('Custom confirm?', onConfirm, onCancel, {
        confirmLabel: 'はい',
        cancelLabel: 'いいえ'
      });
      
      const notification = Array.from(notificationManager.activeNotifications.values())[0];
      expect(notification.actions[0].label).toBe('はい');
      expect(notification.actions[1].label).toBe('いいえ');
    });
  });

  describe('Progress Notifications', () => {
    it('should create progress notification', () => {
      const progress = notificationManager.progress('Processing...');
      
      expect(progress).toHaveProperty('id');
      expect(progress).toHaveProperty('update');
      expect(progress).toHaveProperty('complete');
      expect(progress).toHaveProperty('error');
      expect(notificationManager.activeNotifications.size).toBe(1);
    });

    it('should handle progress completion', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      const progress = notificationManager.progress('Processing...');
      progress.complete('Process completed!');
      
      expect(uiActions.removeNotification).toHaveBeenCalled();
      expect(uiActions.showNotification).toHaveBeenCalledWith(
        'Process completed!',
        'success',
        4000
      );
    });

    it('should handle progress error', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      const progress = notificationManager.progress('Processing...');
      progress.error('Process failed!');
      
      expect(uiActions.removeNotification).toHaveBeenCalled();
      expect(uiActions.showNotification).toHaveBeenCalledWith(
        'Process failed!',
        'error',
        8000
      );
    });
  });

  describe('Notification Management', () => {
    it('should remove notification by ID', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      const notificationId = notificationManager.show('Test message');
      notificationManager.remove(notificationId);
      
      expect(uiActions.removeNotification).toHaveBeenCalledWith(notificationId);
      expect(notificationManager.activeNotifications.size).toBe(0);
    });

    it('should remove notifications by category', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      notificationManager.show('Auth message', { category: NOTIFICATION_CATEGORIES.AUTH });
      notificationManager.show('Network message', { category: NOTIFICATION_CATEGORIES.NETWORK });
      notificationManager.show('Another auth message', { category: NOTIFICATION_CATEGORIES.AUTH });
      
      expect(notificationManager.activeNotifications.size).toBe(3);
      
      notificationManager.removeByCategory(NOTIFICATION_CATEGORIES.AUTH);
      
      expect(uiActions.removeNotification).toHaveBeenCalledTimes(2);
      expect(notificationManager.activeNotifications.size).toBe(1);
    });

    it('should clear all notifications', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      notificationManager.show('Message 1');
      notificationManager.show('Message 2');
      notificationManager.show('Message 3');
      
      notificationManager.clear();
      
      expect(uiActions.clearNotifications).toHaveBeenCalled();
      expect(notificationManager.activeNotifications.size).toBe(0);
    });
  });

  describe('Rate Limiting', () => {
    beforeEach(() => {
      vi.useFakeTimers();
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it('should enforce rate limiting', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      // デフォルトのレート制限: 5秒間に3回まで
      const category = NOTIFICATION_CATEGORIES.SYSTEM;
      const type = 'info';
      
      // 3回まで表示される
      notificationManager.show('Message 1', { category, type });
      notificationManager.show('Message 2', { category, type });
      notificationManager.show('Message 3', { category, type });
      
      expect(uiActions.showNotification).toHaveBeenCalledTimes(3);
      
      // 4回目は制限される
      const result = notificationManager.show('Message 4', { category, type });
      expect(result).toBeNull();
      expect(uiActions.showNotification).toHaveBeenCalledTimes(3);
    });

    it('should reset rate limit after time window', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      const category = NOTIFICATION_CATEGORIES.SYSTEM;
      const type = 'info';
      
      // 制限まで表示
      notificationManager.show('Message 1', { category, type });
      notificationManager.show('Message 2', { category, type });
      notificationManager.show('Message 3', { category, type });
      
      // 時間を進める
      vi.advanceTimersByTime(6000); // 5秒 + 余裕
      
      // 再び表示できる
      const result = notificationManager.show('Message 4', { category, type });
      expect(result).toBeTruthy();
      expect(uiActions.showNotification).toHaveBeenCalledTimes(4);
    });

    it('should use custom rate limit', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      const customRateLimit = {
        maxNotifications: 1,
        timeWindow: 1000
      };
      
      notificationManager.show('Message 1', { rateLimit: customRateLimit });
      expect(uiActions.showNotification).toHaveBeenCalledTimes(1);
      
      const result = notificationManager.show('Message 2', { rateLimit: customRateLimit });
      expect(result).toBeNull();
      expect(uiActions.showNotification).toHaveBeenCalledTimes(1);
    });
  });

  describe('Duplicate Prevention', () => {
    it('should prevent duplicate notifications', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      const options = {
        preventDuplicates: true,
        duplicateKey: 'unique-key'
      };
      
      notificationManager.show('Duplicate message', options);
      expect(uiActions.showNotification).toHaveBeenCalledTimes(1);
      
      const result = notificationManager.show('Different message', options);
      expect(result).toBeTruthy(); // 既存の通知IDが返される
      expect(uiActions.showNotification).toHaveBeenCalledTimes(1); // 新しい通知は作成されない
    });

    it('should use message as duplicate key when not specified', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      const options = { preventDuplicates: true };
      
      notificationManager.show('Same message', options);
      notificationManager.show('Same message', options);
      
      expect(uiActions.showNotification).toHaveBeenCalledTimes(1);
    });
  });

  describe('Statistics and Debug', () => {
    it('should provide notification statistics', () => {
      notificationManager.show('Info message', { type: 'info', category: NOTIFICATION_CATEGORIES.SYSTEM });
      notificationManager.show('Error message', { type: 'error', category: NOTIFICATION_CATEGORIES.AUTH });
      notificationManager.show('Another info', { type: 'info', category: NOTIFICATION_CATEGORIES.DATA });
      
      const stats = notificationManager.getStats();
      
      expect(stats.total).toBe(3);
      expect(stats.byType.info).toBe(2);
      expect(stats.byType.error).toBe(1);
      expect(stats.byCategory[NOTIFICATION_CATEGORIES.SYSTEM]).toBe(1);
      expect(stats.byCategory[NOTIFICATION_CATEGORIES.AUTH]).toBe(1);
      expect(stats.byCategory[NOTIFICATION_CATEGORIES.DATA]).toBe(1);
      expect(stats.averageDisplayTime).toBeGreaterThan(0);
    });

    it('should provide debug information', () => {
      const consoleSpy = vi.spyOn(console, 'group').mockImplementation(() => {});
      const consoleLogSpy = vi.spyOn(console, 'log').mockImplementation(() => {});
      const consoleGroupEndSpy = vi.spyOn(console, 'groupEnd').mockImplementation(() => {});
      
      notificationManager.show('Test message');
      notificationManager.debug();
      
      expect(consoleSpy).toHaveBeenCalledWith('[NotificationManager] Debug Info');
      expect(consoleLogSpy).toHaveBeenCalledTimes(3);
      expect(consoleGroupEndSpy).toHaveBeenCalled();
      
      consoleSpy.mockRestore();
      consoleLogSpy.mockRestore();
      consoleGroupEndSpy.mockRestore();
    });
  });

  describe('Helper Functions', () => {
    it('should export convenience functions', () => {
      const { uiActions } = require('$lib/stores/ui.js');
      
      showSuccess('Success');
      showError('Error');
      showWarning('Warning');
      showInfo('Info');
      
      expect(uiActions.showNotification).toHaveBeenCalledTimes(4);
    });

    it('should show auth error with login action', () => {
      // window.location のモック
      delete window.location;
      window.location = { href: '' };
      
      showAuthError();
      
      const notification = Array.from(notificationManager.activeNotifications.values())[0];
      expect(notification.category).toBe(NOTIFICATION_CATEGORIES.AUTH);
      expect(notification.actions).toHaveLength(1);
      expect(notification.actions[0].label).toBe('ログイン');
    });

    it('should show network error with retry action', () => {
      // window.location のモック
      const mockReload = vi.fn();
      delete window.location;
      window.location = { reload: mockReload };
      
      showNetworkError();
      
      const notification = Array.from(notificationManager.activeNotifications.values())[0];
      expect(notification.category).toBe(NOTIFICATION_CATEGORIES.NETWORK);
      expect(notification.actions).toHaveLength(1);
      expect(notification.actions[0].label).toBe('再試行');
    });
  });
});
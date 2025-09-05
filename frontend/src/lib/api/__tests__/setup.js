// テストセットアップファイル
import { vi } from 'vitest';

// グローバルなモック設定
global.btoa = (str) => Buffer.from(str).toString('base64');
global.atob = (str) => Buffer.from(str, 'base64').toString();

// fetch のモック
global.fetch = vi.fn();

// window オブジェクトのモック
Object.defineProperty(window, 'dispatchEvent', {
  value: vi.fn()
});

// console のモック（テスト中のログを抑制）
global.console = {
  ...console,
  log: vi.fn(),
  error: vi.fn(),
  warn: vi.fn()
};
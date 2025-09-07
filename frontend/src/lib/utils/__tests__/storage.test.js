// ストレージユーティリティの単体テスト
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import {
  setStorageItem,
  getStorageItem,
  removeStorageItem,
  saveAuthToken,
  getAuthToken,
  removeAuthToken,
  saveUserData,
  getUserData,
  removeUserData,
  saveCurrentSport,
  getCurrentSport,
  saveTheme,
  getTheme,
  clearAuthData
} from '../storage.js';

describe('storage utilities', () => {
  let mockLocalStorage;

  beforeEach(() => {
    // localStorage のモック
    mockLocalStorage = {
      getItem: vi.fn(),
      setItem: vi.fn(),
      removeItem: vi.fn(),
      clear: vi.fn()
    };
    
    // グローバルのlocalStorageを置き換え
    Object.defineProperty(window, 'localStorage', {
      value: mockLocalStorage,
      writable: true
    });

    // console.error のモック
    vi.spyOn(console, 'error').mockImplementation(() => {});
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  describe('setStorageItem', () => {
    it('正常にアイテムを保存する', () => {
      const key = 'test-key';
      const value = { data: 'test' };

      const result = setStorageItem(key, value);

      expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
        key,
        JSON.stringify(value)
      );
      expect(result).toBe(true);
    });

    it('保存に失敗した場合はfalseを返す', () => {
      mockLocalStorage.setItem.mockImplementation(() => {
        throw new Error('Storage quota exceeded');
      });

      const result = setStorageItem('test-key', 'test-value');

      expect(result).toBe(false);
      expect(console.error).toHaveBeenCalledWith(
        'Failed to save to localStorage:',
        expect.any(Error)
      );
    });

    it('複雑なオブジェクトを正しくシリアライズする', () => {
      const complexObject = {
        id: 1,
        name: 'test',
        nested: {
          array: [1, 2, 3],
          boolean: true,
          null: null
        }
      };

      setStorageItem('complex', complexObject);

      expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
        'complex',
        JSON.stringify(complexObject)
      );
    });
  });

  describe('getStorageItem', () => {
    it('正常にアイテムを取得する', () => {
      const testData = { data: 'test' };
      mockLocalStorage.getItem.mockReturnValue(JSON.stringify(testData));

      const result = getStorageItem('test-key');

      expect(mockLocalStorage.getItem).toHaveBeenCalledWith('test-key');
      expect(result).toEqual(testData);
    });

    it('存在しないキーに対してデフォルト値を返す', () => {
      mockLocalStorage.getItem.mockReturnValue(null);

      const result = getStorageItem('non-existent', 'default');

      expect(result).toBe('default');
    });

    it('デフォルト値が指定されていない場合はnullを返す', () => {
      mockLocalStorage.getItem.mockReturnValue(null);

      const result = getStorageItem('non-existent');

      expect(result).toBeNull();
    });

    it('JSONパースエラー時にデフォルト値を返す', () => {
      mockLocalStorage.getItem.mockReturnValue('invalid json');

      const result = getStorageItem('invalid-key', 'default');

      expect(result).toBe('default');
      expect(console.error).toHaveBeenCalledWith(
        'Failed to get from localStorage:',
        expect.any(Error)
      );
    });
  });

  describe('removeStorageItem', () => {
    it('正常にアイテムを削除する', () => {
      const result = removeStorageItem('test-key');

      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('test-key');
      expect(result).toBe(true);
    });

    it('削除に失敗した場合はfalseを返す', () => {
      mockLocalStorage.removeItem.mockImplementation(() => {
        throw new Error('Remove failed');
      });

      const result = removeStorageItem('test-key');

      expect(result).toBe(false);
      expect(console.error).toHaveBeenCalledWith(
        'Failed to remove from localStorage:',
        expect.any(Error)
      );
    });
  });

  describe('認証トークン関連', () => {
    describe('saveAuthToken', () => {
      it('認証トークンを正しく保存する', () => {
        const token = 'test-auth-token';
        
        const result = saveAuthToken(token);

        expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
          'tournament_auth_token',
          JSON.stringify(token)
        );
        expect(result).toBe(true);
      });
    });

    describe('getAuthToken', () => {
      it('認証トークンを正しく取得する', () => {
        const token = 'test-auth-token';
        mockLocalStorage.getItem.mockReturnValue(JSON.stringify(token));

        const result = getAuthToken();

        expect(mockLocalStorage.getItem).toHaveBeenCalledWith('tournament_auth_token');
        expect(result).toBe(token);
      });

      it('トークンが存在しない場合はnullを返す', () => {
        mockLocalStorage.getItem.mockReturnValue(null);

        const result = getAuthToken();

        expect(result).toBeNull();
      });
    });

    describe('removeAuthToken', () => {
      it('認証トークンを正しく削除する', () => {
        const result = removeAuthToken();

        expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('tournament_auth_token');
        expect(result).toBe(true);
      });
    });
  });

  describe('ユーザーデータ関連', () => {
    describe('saveUserData', () => {
      it('ユーザーデータを正しく保存する', () => {
        const userData = { id: 1, name: 'Test User', role: 'admin' };
        
        const result = saveUserData(userData);

        expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
          'tournament_user_data',
          JSON.stringify(userData)
        );
        expect(result).toBe(true);
      });
    });

    describe('getUserData', () => {
      it('ユーザーデータを正しく取得する', () => {
        const userData = { id: 1, name: 'Test User', role: 'admin' };
        mockLocalStorage.getItem.mockReturnValue(JSON.stringify(userData));

        const result = getUserData();

        expect(mockLocalStorage.getItem).toHaveBeenCalledWith('tournament_user_data');
        expect(result).toEqual(userData);
      });
    });

    describe('removeUserData', () => {
      it('ユーザーデータを正しく削除する', () => {
        const result = removeUserData();

        expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('tournament_user_data');
        expect(result).toBe(true);
      });
    });
  });

  describe('現在のスポーツ関連', () => {
    describe('saveCurrentSport', () => {
      it('現在のスポーツを正しく保存する', () => {
        const sport = 'table_tennis';
        
        const result = saveCurrentSport(sport);

        expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
          'tournament_current_sport',
          JSON.stringify(sport)
        );
        expect(result).toBe(true);
      });
    });

    describe('getCurrentSport', () => {
      it('現在のスポーツを正しく取得する', () => {
        const sport = 'table_tennis';
        mockLocalStorage.getItem.mockReturnValue(JSON.stringify(sport));

        const result = getCurrentSport();

        expect(mockLocalStorage.getItem).toHaveBeenCalledWith('tournament_current_sport');
        expect(result).toBe(sport);
      });

      it('スポーツが設定されていない場合はデフォルト値を返す', () => {
        mockLocalStorage.getItem.mockReturnValue(null);

        const result = getCurrentSport();

        expect(result).toBe('volleyball');
      });
    });
  });

  describe('テーマ関連', () => {
    describe('saveTheme', () => {
      it('テーマを正しく保存する', () => {
        const theme = 'dark';
        
        const result = saveTheme(theme);

        expect(mockLocalStorage.setItem).toHaveBeenCalledWith(
          'tournament_theme',
          JSON.stringify(theme)
        );
        expect(result).toBe(true);
      });
    });

    describe('getTheme', () => {
      it('テーマを正しく取得する', () => {
        const theme = 'dark';
        mockLocalStorage.getItem.mockReturnValue(JSON.stringify(theme));

        const result = getTheme();

        expect(mockLocalStorage.getItem).toHaveBeenCalledWith('tournament_theme');
        expect(result).toBe(theme);
      });

      it('テーマが設定されていない場合はデフォルト値を返す', () => {
        mockLocalStorage.getItem.mockReturnValue(null);

        const result = getTheme();

        expect(result).toBe('light');
      });
    });
  });

  describe('clearAuthData', () => {
    it('認証関連データを全てクリアする', () => {
      clearAuthData();

      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('tournament_auth_token');
      expect(mockLocalStorage.removeItem).toHaveBeenCalledWith('tournament_user_data');
      expect(mockLocalStorage.removeItem).toHaveBeenCalledTimes(2);
    });
  });

  describe('エラーハンドリング統合テスト', () => {
    it('localStorage が利用できない環境でも正常に動作する', () => {
      // localStorage を undefined に設定
      Object.defineProperty(window, 'localStorage', {
        value: undefined,
        writable: true
      });

      // エラーが発生することを確認
      expect(() => setStorageItem('key', 'value')).not.toThrow();
      expect(() => getStorageItem('key')).not.toThrow();
      expect(() => removeStorageItem('key')).not.toThrow();
    });

    it('JSON.stringify が失敗した場合の処理', () => {
      // 循環参照オブジェクトを作成
      const circularObj = {};
      circularObj.self = circularObj;

      const result = setStorageItem('circular', circularObj);

      expect(result).toBe(false);
      expect(console.error).toHaveBeenCalled();
    });
  });
});
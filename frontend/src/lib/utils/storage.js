// ローカルストレージユーティリティ

// トークン関連のストレージキー
const STORAGE_KEYS = {
  AUTH_TOKEN: 'tournament_auth_token',
  USER_DATA: 'tournament_user_data',
  CURRENT_SPORT: 'tournament_current_sport',
  THEME: 'tournament_theme'
};

// ローカルストレージへの保存
export function setStorageItem(key, value) {
  try {
    const serializedValue = JSON.stringify(value);
    localStorage.setItem(key, serializedValue);
    return true;
  } catch (error) {
    console.error('Failed to save to localStorage:', error);
    return false;
  }
}

// ローカルストレージからの取得
export function getStorageItem(key, defaultValue = null) {
  try {
    const item = localStorage.getItem(key);
    if (item === null) {
      return defaultValue;
    }
    return JSON.parse(item);
  } catch (error) {
    console.error('Failed to get from localStorage:', error);
    return defaultValue;
  }
}

// ローカルストレージからの削除
export function removeStorageItem(key) {
  try {
    localStorage.removeItem(key);
    return true;
  } catch (error) {
    console.error('Failed to remove from localStorage:', error);
    return false;
  }
}

// 認証トークンの保存
export function saveAuthToken(token) {
  return setStorageItem(STORAGE_KEYS.AUTH_TOKEN, token);
}

// 認証トークンの取得
export function getAuthToken() {
  return getStorageItem(STORAGE_KEYS.AUTH_TOKEN);
}

// 認証トークンの削除
export function removeAuthToken() {
  return removeStorageItem(STORAGE_KEYS.AUTH_TOKEN);
}

// ユーザーデータの保存
export function saveUserData(userData) {
  return setStorageItem(STORAGE_KEYS.USER_DATA, userData);
}

// ユーザーデータの取得
export function getUserData() {
  return getStorageItem(STORAGE_KEYS.USER_DATA);
}

// ユーザーデータの削除
export function removeUserData() {
  return removeStorageItem(STORAGE_KEYS.USER_DATA);
}

// 現在のスポーツの保存
export function saveCurrentSport(sport) {
  return setStorageItem(STORAGE_KEYS.CURRENT_SPORT, sport);
}

// 現在のスポーツの取得
export function getCurrentSport() {
  return getStorageItem(STORAGE_KEYS.CURRENT_SPORT, 'volleyball');
}

// テーマの保存
export function saveTheme(theme) {
  return setStorageItem(STORAGE_KEYS.THEME, theme);
}

// テーマの取得
export function getTheme() {
  return getStorageItem(STORAGE_KEYS.THEME, 'light');
}

// 全ての認証関連データをクリア
export function clearAuthData() {
  removeAuthToken();
  removeUserData();
}

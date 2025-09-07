const STORAGE_KEYS = {
  AUTH_TOKEN: "tournament_auth_token",
  USER_DATA: "tournament_user_data"
};
function setStorageItem(key, value) {
  try {
    const serializedValue = JSON.stringify(value);
    localStorage.setItem(key, serializedValue);
    return true;
  } catch (error) {
    console.error("Failed to save to localStorage:", error);
    return false;
  }
}
function getStorageItem(key, defaultValue = null) {
  try {
    const item = localStorage.getItem(key);
    if (item === null) {
      return defaultValue;
    }
    return JSON.parse(item);
  } catch (error) {
    console.error("Failed to get from localStorage:", error);
    return defaultValue;
  }
}
function removeStorageItem(key) {
  try {
    localStorage.removeItem(key);
    return true;
  } catch (error) {
    console.error("Failed to remove from localStorage:", error);
    return false;
  }
}
function saveAuthToken(token) {
  return setStorageItem(STORAGE_KEYS.AUTH_TOKEN, token);
}
function getAuthToken() {
  return getStorageItem(STORAGE_KEYS.AUTH_TOKEN);
}
function removeAuthToken() {
  return removeStorageItem(STORAGE_KEYS.AUTH_TOKEN);
}
function saveUserData(userData) {
  return setStorageItem(STORAGE_KEYS.USER_DATA, userData);
}
function getUserData() {
  return getStorageItem(STORAGE_KEYS.USER_DATA);
}
function removeUserData() {
  return removeStorageItem(STORAGE_KEYS.USER_DATA);
}
function clearAuthData() {
  removeAuthToken();
  removeUserData();
}
export {
  getUserData as a,
  saveUserData as b,
  clearAuthData as c,
  getAuthToken as g,
  saveAuthToken as s
};

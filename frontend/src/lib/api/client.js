// HTTPクライアント設定
class APIClient {
  constructor(baseURL = '/api') {
    this.baseURL = baseURL;
    this.token = null;
  }

  // トークン設定
  setToken(token) {
    this.token = token;
  }

  // 共通リクエストヘッダーの取得
  getHeaders() {
    const headers = {
      'Content-Type': 'application/json'
    };

    if (this.token) {
      headers.Authorization = `Bearer ${this.token}`;
    }

    return headers;
  }

  // GET リクエスト
  async get(endpoint, options = {}) {
    // 実装は後のタスクで行う
    console.log('GET request will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  }

  // POST リクエスト
  async post(endpoint, data, options = {}) {
    // 実装は後のタスクで行う
    console.log('POST request will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  }

  // PUT リクエスト
  async put(endpoint, data, options = {}) {
    // 実装は後のタスクで行う
    console.log('PUT request will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  }

  // DELETE リクエスト
  async delete(endpoint, options = {}) {
    // 実装は後のタスクで行う
    console.log('DELETE request will be implemented in later tasks');
    return { success: false, message: 'Not implemented yet' };
  }

  // レスポンス処理
  handleResponse(response) {
    // 実装は後のタスクで行う
    console.log('Response handling will be implemented in later tasks');
  }

  // エラー処理
  handleError(error) {
    // 実装は後のタスクで行う
    console.log('Error handling will be implemented in later tasks');
  }
}

// デフォルトのAPIクライアントインスタンス
export const apiClient = new APIClient();
export default APIClient;

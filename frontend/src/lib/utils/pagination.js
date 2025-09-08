/**
 * ページネーション関連のユーティリティ関数
 */

/**
 * ページネーションパラメータを作成する
 * @param {number} page - ページ番号（1から開始）
 * @param {number} pageSize - ページサイズ
 * @returns {Object} ページネーションパラメータ
 */
export function createPaginationParams(page = 1, pageSize = 20) {
	return {
		page: Math.max(1, page),
		page_size: Math.max(1, Math.min(100, pageSize))
	};
}

/**
 * URLSearchParamsにページネーションパラメータを追加する
 * @param {URLSearchParams} params - URLSearchParams オブジェクト
 * @param {number} page - ページ番号
 * @param {number} pageSize - ページサイズ
 */
export function addPaginationToParams(params, page, pageSize) {
	const paginationParams = createPaginationParams(page, pageSize);
	params.set('page', paginationParams.page.toString());
	params.set('page_size', paginationParams.page_size.toString());
}

/**
 * ページネーションレスポンスを検証する
 * @param {Object} response - APIレスポンス
 * @returns {boolean} 有効なページネーションレスポンスかどうか
 */
export function validatePaginationResponse(response) {
	if (!response || typeof response !== 'object') {
		return false;
	}

	const { pagination } = response;
	if (!pagination || typeof pagination !== 'object') {
		return false;
	}

	const requiredFields = ['page', 'page_size', 'total_items', 'total_pages', 'has_next', 'has_prev'];
	return requiredFields.every(field => field in pagination);
}

/**
 * ページネーション情報からページ範囲を計算する
 * @param {Object} pagination - ページネーション情報
 * @returns {Object} ページ範囲情報
 */
export function calculatePageRange(pagination) {
	if (!pagination) {
		return { start: 0, end: 0, total: 0 };
	}

	const { page, page_size, total_items } = pagination;
	const start = total_items === 0 ? 0 : (page - 1) * page_size + 1;
	const end = Math.min(page * page_size, total_items);

	return {
		start,
		end,
		total: total_items
	};
}

/**
 * ページネーション状態を管理するクラス
 */
export class PaginationState {
	constructor(initialPage = 1, initialPageSize = 20) {
		this.page = initialPage;
		this.pageSize = initialPageSize;
		this.totalItems = 0;
		this.totalPages = 0;
		this.hasNext = false;
		this.hasPrev = false;
		this.loading = false;
		this.error = null;
	}

	/**
	 * ページネーション情報を更新する
	 * @param {Object} paginationData - ページネーション情報
	 */
	update(paginationData) {
		if (!paginationData) return;

		this.page = paginationData.page || this.page;
		this.pageSize = paginationData.page_size || this.pageSize;
		this.totalItems = paginationData.total_items || 0;
		this.totalPages = paginationData.total_pages || 0;
		this.hasNext = paginationData.has_next || false;
		this.hasPrev = paginationData.has_prev || false;
	}

	/**
	 * ページを変更する
	 * @param {number} newPage - 新しいページ番号
	 */
	setPage(newPage) {
		this.page = Math.max(1, Math.min(newPage, this.totalPages));
	}

	/**
	 * ページサイズを変更する
	 * @param {number} newPageSize - 新しいページサイズ
	 */
	setPageSize(newPageSize) {
		this.pageSize = Math.max(1, Math.min(100, newPageSize));
		this.page = 1; // ページサイズ変更時は1ページ目に戻る
	}

	/**
	 * 次のページに移動する
	 */
	nextPage() {
		if (this.hasNext) {
			this.page++;
		}
	}

	/**
	 * 前のページに移動する
	 */
	prevPage() {
		if (this.hasPrev) {
			this.page--;
		}
	}

	/**
	 * 最初のページに移動する
	 */
	firstPage() {
		this.page = 1;
	}

	/**
	 * 最後のページに移動する
	 */
	lastPage() {
		this.page = this.totalPages;
	}

	/**
	 * ローディング状態を設定する
	 * @param {boolean} loading - ローディング状態
	 */
	setLoading(loading) {
		this.loading = loading;
		if (loading) {
			this.error = null;
		}
	}

	/**
	 * エラー状態を設定する
	 * @param {Error|string} error - エラー情報
	 */
	setError(error) {
		this.error = error;
		this.loading = false;
	}

	/**
	 * エラーをクリアする
	 */
	clearError() {
		this.error = null;
	}

	/**
	 * 現在のページネーションパラメータを取得する
	 * @returns {Object} ページネーションパラメータ
	 */
	getParams() {
		return createPaginationParams(this.page, this.pageSize);
	}

	/**
	 * ページ範囲情報を取得する
	 * @returns {Object} ページ範囲情報
	 */
	getRange() {
		return calculatePageRange({
			page: this.page,
			page_size: this.pageSize,
			total_items: this.totalItems
		});
	}

	/**
	 * 状態をリセットする
	 */
	reset() {
		this.page = 1;
		this.totalItems = 0;
		this.totalPages = 0;
		this.hasNext = false;
		this.hasPrev = false;
		this.loading = false;
		this.error = null;
	}

	/**
	 * 状態をJSONオブジェクトとして取得する
	 * @returns {Object} 状態オブジェクト
	 */
	toJSON() {
		return {
			page: this.page,
			pageSize: this.pageSize,
			totalItems: this.totalItems,
			totalPages: this.totalPages,
			hasNext: this.hasNext,
			hasPrev: this.hasPrev,
			loading: this.loading,
			error: this.error
		};
	}
}

/**
 * ページネーション付きデータフェッチャー
 */
export class PaginatedDataFetcher {
	constructor(fetchFunction, options = {}) {
		this.fetchFunction = fetchFunction;
		this.options = {
			defaultPageSize: 20,
			maxRetries: 3,
			retryDelay: 1000,
			...options
		};
		this.pagination = new PaginationState(1, this.options.defaultPageSize);
		this.data = [];
		this.cache = new Map();
	}

	/**
	 * データを取得する
	 * @param {Object} params - 追加パラメータ
	 * @param {boolean} useCache - キャッシュを使用するか
	 * @returns {Promise<Object>} 取得結果
	 */
	async fetch(params = {}, useCache = true) {
		const paginationParams = this.pagination.getParams();
		const allParams = { ...params, ...paginationParams };
		const cacheKey = JSON.stringify(allParams);

		// キャッシュチェック
		if (useCache && this.cache.has(cacheKey)) {
			const cachedResult = this.cache.get(cacheKey);
			this.data = cachedResult.data;
			this.pagination.update(cachedResult.pagination);
			return cachedResult;
		}

		this.pagination.setLoading(true);

		try {
			const result = await this.fetchWithRetry(allParams);
			
			if (!validatePaginationResponse(result)) {
				throw new Error('無効なページネーションレスポンスです');
			}

			this.data = result.data || [];
			this.pagination.update(result.pagination);

			// キャッシュに保存
			this.cache.set(cacheKey, result);

			return result;
		} catch (error) {
			this.pagination.setError(error);
			throw error;
		} finally {
			this.pagination.setLoading(false);
		}
	}

	/**
	 * リトライ機能付きフェッチ
	 * @param {Object} params - パラメータ
	 * @returns {Promise<Object>} 取得結果
	 */
	async fetchWithRetry(params) {
		let lastError;
		
		for (let i = 0; i < this.options.maxRetries; i++) {
			try {
				return await this.fetchFunction(params);
			} catch (error) {
				lastError = error;
				
				if (i < this.options.maxRetries - 1) {
					await new Promise(resolve => 
						setTimeout(resolve, this.options.retryDelay * (i + 1))
					);
				}
			}
		}
		
		throw lastError;
	}

	/**
	 * ページを変更してデータを取得する
	 * @param {number} page - ページ番号
	 * @param {Object} params - 追加パラメータ
	 * @returns {Promise<Object>} 取得結果
	 */
	async goToPage(page, params = {}) {
		this.pagination.setPage(page);
		return this.fetch(params);
	}

	/**
	 * ページサイズを変更してデータを取得する
	 * @param {number} pageSize - ページサイズ
	 * @param {Object} params - 追加パラメータ
	 * @returns {Promise<Object>} 取得結果
	 */
	async changePageSize(pageSize, params = {}) {
		this.pagination.setPageSize(pageSize);
		return this.fetch(params);
	}

	/**
	 * データを更新する
	 * @param {Object} params - 追加パラメータ
	 * @returns {Promise<Object>} 取得結果
	 */
	async refresh(params = {}) {
		this.clearCache();
		return this.fetch(params, false);
	}

	/**
	 * キャッシュをクリアする
	 */
	clearCache() {
		this.cache.clear();
	}

	/**
	 * 状態をリセットする
	 */
	reset() {
		this.pagination.reset();
		this.data = [];
		this.clearCache();
	}
}

/**
 * デフォルトのページネーション設定
 */
export const DEFAULT_PAGINATION_CONFIG = {
	defaultPage: 1,
	defaultPageSize: 20,
	maxPageSize: 100,
	pageSizeOptions: [10, 20, 50, 100],
	maxVisiblePages: 5
};

/**
 * ページネーション設定を検証する
 * @param {Object} config - 設定オブジェクト
 * @returns {Object} 検証済み設定
 */
export function validatePaginationConfig(config = {}) {
	return {
		...DEFAULT_PAGINATION_CONFIG,
		...config,
		defaultPage: Math.max(1, config.defaultPage || DEFAULT_PAGINATION_CONFIG.defaultPage),
		defaultPageSize: Math.max(1, Math.min(100, config.defaultPageSize || DEFAULT_PAGINATION_CONFIG.defaultPageSize)),
		maxPageSize: Math.max(1, config.maxPageSize || DEFAULT_PAGINATION_CONFIG.maxPageSize),
		maxVisiblePages: Math.max(1, config.maxVisiblePages || DEFAULT_PAGINATION_CONFIG.maxVisiblePages)
	};
}
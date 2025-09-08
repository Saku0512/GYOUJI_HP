/**
 * レスポンス最適化ユーティリティ
 * データ転送量の削減とパフォーマンス向上のための機能
 */

/**
 * レスポンスデータを最適化する
 * @param {Object} data - 最適化対象のデータ
 * @param {Object} options - 最適化オプション
 * @returns {Object} 最適化されたデータ
 */
export function optimizeResponseData(data, options = {}) {
	const {
		removeNullValues = true,
		removeEmptyStrings = false,
		removeEmptyArrays = false,
		compressStrings = false,
		maxStringLength = 1000
	} = options;

	if (!data || typeof data !== 'object') {
		return data;
	}

	if (Array.isArray(data)) {
		return data
			.map(item => optimizeResponseData(item, options))
			.filter(item => {
				if (removeEmptyArrays && Array.isArray(item) && item.length === 0) {
					return false;
				}
				return true;
			});
	}

	const optimized = {};

	for (const [key, value] of Object.entries(data)) {
		// null値の除去
		if (removeNullValues && (value === null || value === undefined)) {
			continue;
		}

		// 空文字列の除去
		if (removeEmptyStrings && value === '') {
			continue;
		}

		// 文字列の最適化
		if (typeof value === 'string') {
			let optimizedString = value;

			// 文字列圧縮（長い文字列の場合）
			if (compressStrings && value.length > maxStringLength) {
				optimizedString = value.substring(0, maxStringLength) + '...';
			}

			optimized[key] = optimizedString;
		}
		// オブジェクトの再帰的最適化
		else if (typeof value === 'object') {
			const optimizedValue = optimizeResponseData(value, options);
			
			// 空配列の除去
			if (removeEmptyArrays && Array.isArray(optimizedValue) && optimizedValue.length === 0) {
				continue;
			}

			optimized[key] = optimizedValue;
		}
		// その他の値はそのまま
		else {
			optimized[key] = value;
		}
	}

	return optimized;
}

/**
 * フィールド選択によるデータ最適化
 * @param {Object} data - 元データ
 * @param {Array<string>} fields - 選択するフィールド
 * @returns {Object} 選択されたフィールドのみを含むデータ
 */
export function selectFields(data, fields) {
	if (!data || typeof data !== 'object' || !Array.isArray(fields)) {
		return data;
	}

	if (Array.isArray(data)) {
		return data.map(item => selectFields(item, fields));
	}

	const selected = {};
	
	for (const field of fields) {
		if (field.includes('.')) {
			// ネストされたフィールドの処理
			const [parent, ...rest] = field.split('.');
			if (data[parent] !== undefined) {
				if (!selected[parent]) {
					selected[parent] = {};
				}
				const nestedResult = selectFields(data[parent], [rest.join('.')]);
				Object.assign(selected[parent], nestedResult);
			}
		} else {
			// 単純なフィールドの処理
			if (data[field] !== undefined) {
				selected[field] = data[field];
			}
		}
	}

	return selected;
}

/**
 * レスポンスキャッシュマネージャー
 */
export class ResponseCache {
	constructor(options = {}) {
		this.cache = new Map();
		this.maxSize = options.maxSize || 100;
		this.defaultTTL = options.defaultTTL || 5 * 60 * 1000; // 5分
		this.cleanupInterval = options.cleanupInterval || 60 * 1000; // 1分

		// 定期的なクリーンアップを開始
		this.startCleanup();
	}

	/**
	 * キャッシュキーを生成する
	 * @param {string} url - リクエストURL
	 * @param {Object} params - リクエストパラメータ
	 * @returns {string} キャッシュキー
	 */
	generateKey(url, params = {}) {
		const sortedParams = Object.keys(params)
			.sort()
			.reduce((result, key) => {
				result[key] = params[key];
				return result;
			}, {});

		return `${url}:${JSON.stringify(sortedParams)}`;
	}

	/**
	 * キャッシュからデータを取得する
	 * @param {string} key - キャッシュキー
	 * @returns {Object|null} キャッシュされたデータまたはnull
	 */
	get(key) {
		const entry = this.cache.get(key);
		
		if (!entry) {
			return null;
		}

		// TTLチェック
		if (Date.now() > entry.expiry) {
			this.cache.delete(key);
			return null;
		}

		// アクセス時間を更新
		entry.lastAccess = Date.now();
		return entry.data;
	}

	/**
	 * データをキャッシュに保存する
	 * @param {string} key - キャッシュキー
	 * @param {Object} data - 保存するデータ
	 * @param {number} ttl - TTL（ミリ秒）
	 */
	set(key, data, ttl = this.defaultTTL) {
		// キャッシュサイズ制限チェック
		if (this.cache.size >= this.maxSize) {
			this.evictLRU();
		}

		const entry = {
			data: data,
			expiry: Date.now() + ttl,
			lastAccess: Date.now()
		};

		this.cache.set(key, entry);
	}

	/**
	 * LRU（Least Recently Used）アルゴリズムでエントリを削除
	 */
	evictLRU() {
		let oldestKey = null;
		let oldestTime = Date.now();

		for (const [key, entry] of this.cache.entries()) {
			if (entry.lastAccess < oldestTime) {
				oldestTime = entry.lastAccess;
				oldestKey = key;
			}
		}

		if (oldestKey) {
			this.cache.delete(oldestKey);
		}
	}

	/**
	 * 期限切れエントリのクリーンアップ
	 */
	cleanup() {
		const now = Date.now();
		const keysToDelete = [];

		for (const [key, entry] of this.cache.entries()) {
			if (now > entry.expiry) {
				keysToDelete.push(key);
			}
		}

		keysToDelete.forEach(key => this.cache.delete(key));
	}

	/**
	 * 定期的なクリーンアップを開始
	 */
	startCleanup() {
		this.cleanupTimer = setInterval(() => {
			this.cleanup();
		}, this.cleanupInterval);
	}

	/**
	 * クリーンアップを停止
	 */
	stopCleanup() {
		if (this.cleanupTimer) {
			clearInterval(this.cleanupTimer);
			this.cleanupTimer = null;
		}
	}

	/**
	 * キャッシュをクリア
	 */
	clear() {
		this.cache.clear();
	}

	/**
	 * キャッシュ統計を取得
	 * @returns {Object} キャッシュ統計
	 */
	getStats() {
		return {
			size: this.cache.size,
			maxSize: this.maxSize,
			hitRate: this.hitCount / (this.hitCount + this.missCount) || 0,
			hitCount: this.hitCount || 0,
			missCount: this.missCount || 0
		};
	}
}

/**
 * レスポンス圧縮ユーティリティ
 */
export class ResponseCompressor {
	/**
	 * JSON文字列を圧縮する（簡易版）
	 * @param {string} jsonString - 圧縮対象のJSON文字列
	 * @returns {string} 圧縮されたJSON文字列
	 */
	static compressJSON(jsonString) {
		try {
			const obj = JSON.parse(jsonString);
			// 不要な空白を除去
			return JSON.stringify(obj);
		} catch (error) {
			console.warn('JSON圧縮に失敗しました:', error);
			return jsonString;
		}
	}

	/**
	 * 配列データの重複を除去する
	 * @param {Array} array - 重複除去対象の配列
	 * @param {string} keyField - 重複判定に使用するキーフィールド
	 * @returns {Array} 重複が除去された配列
	 */
	static deduplicateArray(array, keyField = 'id') {
		if (!Array.isArray(array)) {
			return array;
		}

		const seen = new Set();
		return array.filter(item => {
			const key = item[keyField];
			if (seen.has(key)) {
				return false;
			}
			seen.add(key);
			return true;
		});
	}

	/**
	 * オブジェクトから空の値を除去する
	 * @param {Object} obj - 処理対象のオブジェクト
	 * @returns {Object} 空の値が除去されたオブジェクト
	 */
	static removeEmptyValues(obj) {
		if (!obj || typeof obj !== 'object') {
			return obj;
		}

		if (Array.isArray(obj)) {
			return obj
				.map(item => this.removeEmptyValues(item))
				.filter(item => item !== null && item !== undefined && item !== '');
		}

		const cleaned = {};
		for (const [key, value] of Object.entries(obj)) {
			if (value !== null && value !== undefined && value !== '') {
				if (typeof value === 'object') {
					const cleanedValue = this.removeEmptyValues(value);
					if (Array.isArray(cleanedValue) ? cleanedValue.length > 0 : Object.keys(cleanedValue).length > 0) {
						cleaned[key] = cleanedValue;
					}
				} else {
					cleaned[key] = value;
				}
			}
		}

		return cleaned;
	}
}

/**
 * バンドルサイズ最適化のためのデータ変換
 */
export class DataTransformer {
	/**
	 * 日時文字列を短縮形式に変換
	 * @param {string} dateString - ISO日時文字列
	 * @returns {string} 短縮された日時文字列
	 */
	static compressDate(dateString) {
		try {
			const date = new Date(dateString);
			// Unix timestamp（秒）に変換してサイズを削減
			return Math.floor(date.getTime() / 1000).toString();
		} catch (error) {
			return dateString;
		}
	}

	/**
	 * 短縮された日時文字列を元に戻す
	 * @param {string} compressedDate - 短縮された日時文字列
	 * @returns {string} ISO日時文字列
	 */
	static decompressDate(compressedDate) {
		try {
			const timestamp = parseInt(compressedDate) * 1000;
			return new Date(timestamp).toISOString();
		} catch (error) {
			return compressedDate;
		}
	}

	/**
	 * 列挙値を数値に変換してサイズを削減
	 * @param {string} enumValue - 列挙値
	 * @param {Object} enumMap - 列挙値マッピング
	 * @returns {number} 数値化された列挙値
	 */
	static compressEnum(enumValue, enumMap) {
		return enumMap[enumValue] !== undefined ? enumMap[enumValue] : enumValue;
	}

	/**
	 * 数値化された列挙値を元に戻す
	 * @param {number} compressedEnum - 数値化された列挙値
	 * @param {Object} enumMap - 列挙値マッピング
	 * @returns {string} 元の列挙値
	 */
	static decompressEnum(compressedEnum, enumMap) {
		const reverseMap = Object.fromEntries(
			Object.entries(enumMap).map(([key, value]) => [value, key])
		);
		return reverseMap[compressedEnum] !== undefined ? reverseMap[compressedEnum] : compressedEnum;
	}
}

/**
 * デフォルトのレスポンスキャッシュインスタンス
 */
export const defaultResponseCache = new ResponseCache({
	maxSize: 50,
	defaultTTL: 5 * 60 * 1000, // 5分
	cleanupInterval: 60 * 1000 // 1分
});

/**
 * レスポンス最適化設定
 */
export const OPTIMIZATION_PRESETS = {
	// 高速化重視（キャッシュ多用、データ圧縮）
	performance: {
		removeNullValues: true,
		removeEmptyStrings: true,
		removeEmptyArrays: true,
		compressStrings: true,
		maxStringLength: 500,
		useCache: true,
		cacheTTL: 10 * 60 * 1000 // 10分
	},

	// バンドルサイズ重視（データ最小化）
	size: {
		removeNullValues: true,
		removeEmptyStrings: true,
		removeEmptyArrays: true,
		compressStrings: true,
		maxStringLength: 200,
		useCache: false
	},

	// 開発用（最適化なし）
	development: {
		removeNullValues: false,
		removeEmptyStrings: false,
		removeEmptyArrays: false,
		compressStrings: false,
		useCache: false
	}
};

/**
 * プリセットを適用してレスポンスを最適化
 * @param {Object} data - 最適化対象のデータ
 * @param {string} preset - プリセット名
 * @returns {Object} 最適化されたデータ
 */
export function optimizeWithPreset(data, preset = 'performance') {
	const config = OPTIMIZATION_PRESETS[preset] || OPTIMIZATION_PRESETS.performance;
	return optimizeResponseData(data, config);
}
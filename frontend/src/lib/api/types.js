// 統一されたAPI型定義（JavaScript版）
// バックエンドと一致するデータ型の定義

// エラーコード列挙型
export const ErrorCode = {
  // 認証関連
  AUTH_INVALID_CREDENTIALS: 'AUTH_INVALID_CREDENTIALS',
  AUTH_TOKEN_EXPIRED: 'AUTH_TOKEN_EXPIRED',
  AUTH_TOKEN_INVALID: 'AUTH_TOKEN_INVALID',
  AUTH_UNAUTHORIZED: 'AUTH_UNAUTHORIZED',
  AUTH_FORBIDDEN: 'AUTH_FORBIDDEN',

  // バリデーション関連
  VALIDATION_REQUIRED_FIELD: 'VALIDATION_REQUIRED_FIELD',
  VALIDATION_INVALID_FORMAT: 'VALIDATION_INVALID_FORMAT',
  VALIDATION_OUT_OF_RANGE: 'VALIDATION_OUT_OF_RANGE',
  VALIDATION_DUPLICATE_VALUE: 'VALIDATION_DUPLICATE_VALUE',

  // リソース関連
  RESOURCE_NOT_FOUND: 'RESOURCE_NOT_FOUND',
  RESOURCE_ALREADY_EXISTS: 'RESOURCE_ALREADY_EXISTS',
  RESOURCE_CONFLICT: 'RESOURCE_CONFLICT',

  // ビジネスロジック関連
  BUSINESS_TOURNAMENT_COMPLETED: 'BUSINESS_TOURNAMENT_COMPLETED',
  BUSINESS_MATCH_ALREADY_COMPLETED: 'BUSINESS_MATCH_ALREADY_COMPLETED',
  BUSINESS_INVALID_MATCH_RESULT: 'BUSINESS_INVALID_MATCH_RESULT',

  // システム関連
  SYSTEM_DATABASE_ERROR: 'SYSTEM_DATABASE_ERROR',
  SYSTEM_NETWORK_ERROR: 'SYSTEM_NETWORK_ERROR',
  SYSTEM_TIMEOUT: 'SYSTEM_TIMEOUT',
  SYSTEM_UNKNOWN_ERROR: 'SYSTEM_UNKNOWN_ERROR'
};

// スポーツタイプ
export const SportType = {
  VOLLEYBALL: 'volleyball',
  TABLE_TENNIS: 'table_tennis',
  SOCCER: 'soccer'
};

// トーナメントステータス
export const TournamentStatus = {
  REGISTRATION: 'registration',
  ACTIVE: 'active',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled'
};

// 試合ステータス
export const MatchStatus = {
  PENDING: 'pending',
  IN_PROGRESS: 'in_progress',
  COMPLETED: 'completed',
  CANCELLED: 'cancelled'
};

// HTTPメソッド
export const HTTPMethod = {
  GET: 'GET',
  POST: 'POST',
  PUT: 'PUT',
  DELETE: 'DELETE',
  PATCH: 'PATCH',
  HEAD: 'HEAD',
  OPTIONS: 'OPTIONS'
};

// JSDocコメントで型情報を提供

/**
 * @typedef {Object} APIResponse
 * @property {boolean} success
 * @property {*} [data]
 * @property {string} [error]
 * @property {string} message
 * @property {number} code
 * @property {string} timestamp
 * @property {string} [request_id]
 */

/**
 * @typedef {Object} PaginationRequest
 * @property {number} page
 * @property {number} page_size
 */

/**
 * @typedef {Object} PaginationResponse
 * @property {number} page
 * @property {number} page_size
 * @property {number} total_items
 * @property {number} total_pages
 * @property {boolean} has_next
 * @property {boolean} has_prev
 */

/**
 * @typedef {APIResponse & {pagination?: PaginationResponse}} PaginatedAPIResponse
 */

/**
 * @typedef {Object} User
 * @property {number} id
 * @property {string} username
 * @property {string} role
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} LoginRequest
 * @property {string} username
 * @property {string} password
 */

/**
 * @typedef {Object} AuthResponse
 * @property {string} token
 * @property {string} [refresh_token]
 * @property {User} user
 */

/**
 * @typedef {Object} Tournament
 * @property {number} id
 * @property {string} sport
 * @property {string} format
 * @property {string} status
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} CreateTournamentRequest
 * @property {string} sport
 * @property {string} format
 * @property {string[]} [teams]
 */

/**
 * @typedef {Object} UpdateTournamentRequest
 * @property {string} [format]
 * @property {string} [status]
 */

/**
 * @typedef {Object} TournamentBracket
 * @property {number} tournament_id
 * @property {string} sport
 * @property {BracketRound[]} rounds
 * @property {string} updated_at
 */

/**
 * @typedef {Object} BracketRound
 * @property {string} round
 * @property {BracketMatch[]} matches
 */

/**
 * @typedef {Object} BracketMatch
 * @property {number} id
 * @property {string} team1
 * @property {string} team2
 * @property {number|null} score1
 * @property {number|null} score2
 * @property {string|null} winner
 * @property {string} status
 */

/**
 * @typedef {Object} Match
 * @property {number} id
 * @property {number} tournament_id
 * @property {string} round
 * @property {string} team1
 * @property {string} team2
 * @property {number|null} score1
 * @property {number|null} score2
 * @property {string|null} winner
 * @property {string} status
 * @property {string} scheduled_at
 * @property {string|null} completed_at
 * @property {string} created_at
 * @property {string} updated_at
 */

/**
 * @typedef {Object} CreateMatchRequest
 * @property {string} sport
 * @property {number} tournament_id
 * @property {string} round
 * @property {string} team1
 * @property {string} team2
 * @property {string} [scheduled_at]
 */

/**
 * @typedef {Object} UpdateMatchRequest
 * @property {string} [round]
 * @property {string} [team1]
 * @property {string} [team2]
 * @property {string} [scheduled_at]
 * @property {string} [status]
 */

/**
 * @typedef {Object} MatchResult
 * @property {number} score1
 * @property {number} score2
 * @property {string} winner
 */

/**
 * @typedef {Object} MatchFilters
 * @property {string} [status]
 * @property {string} [round]
 * @property {number} [limit]
 * @property {number} [offset]
 */

/**
 * @typedef {Object} APIError
 * @property {string} code
 * @property {string} message
 * @property {number} status_code
 * @property {Object} [details]
 */

/**
 * @typedef {Object} RequestOptions
 * @property {Object} [headers]
 * @property {number} [timeout]
 * @property {boolean} [retry]
 * @property {AbortSignal} [signal]
 */

/**
 * @typedef {Object} UpdateNotification
 * @property {string} type
 * @property {string} sport
 * @property {*} data
 * @property {string} timestamp
 */

/**
 * @typedef {Object} TournamentStats
 * @property {number} total_matches
 * @property {number} completed_matches
 * @property {number} pending_matches
 * @property {number} completion_rate
 * @property {string} last_updated
 */

/**
 * @typedef {Object} MatchStats
 * @property {number|null} duration_minutes
 * @property {number} total_points
 * @property {number} sets_played
 * @property {string} last_updated
 */
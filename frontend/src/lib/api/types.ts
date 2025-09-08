// 統一されたAPI型定義
// バックエンドと一致するデータ型の定義

// 基本的なレスポンス型
export interface APIResponse<T = any> {
  success: boolean;
  data?: T;
  error?: string;
  message: string;
  code: number;
  timestamp: string;
  request_id?: string;
}

// ページネーション関連
export interface PaginationRequest {
  page: number;
  page_size: number;
}

export interface PaginationResponse {
  page: number;
  page_size: number;
  total_items: number;
  total_pages: number;
  has_next: boolean;
  has_prev: boolean;
}

export interface PaginatedAPIResponse<T = any> extends APIResponse<T> {
  pagination?: PaginationResponse;
}

// 列挙型定義
export type SportType = 'volleyball' | 'table_tennis' | 'soccer';
export type TournamentStatus = 'registration' | 'active' | 'completed' | 'cancelled';
export type MatchStatus = 'pending' | 'in_progress' | 'completed' | 'cancelled';

// ユーザー関連型
export interface User {
  id: number;
  username: string;
  role: string;
  created_at: string;
  updated_at: string;
}

export interface LoginRequest {
  username: string;
  password: string;
}

export interface AuthResponse {
  token: string;
  refresh_token?: string;
  user: User;
}

// トーナメント関連型
export interface Tournament {
  id: number;
  sport: SportType;
  format: string;
  status: TournamentStatus;
  created_at: string;
  updated_at: string;
}

export interface CreateTournamentRequest {
  sport: SportType;
  format: string;
  teams?: string[];
}

export interface UpdateTournamentRequest {
  format?: string;
  status?: TournamentStatus;
}

export interface TournamentBracket {
  tournament_id: number;
  sport: SportType;
  rounds: BracketRound[];
  updated_at: string;
}

export interface BracketRound {
  round: string;
  matches: BracketMatch[];
}

export interface BracketMatch {
  id: number;
  team1: string;
  team2: string;
  score1: number | null;
  score2: number | null;
  winner: string | null;
  status: MatchStatus;
}

// 試合関連型
export interface Match {
  id: number;
  tournament_id: number;
  round: string;
  team1: string;
  team2: string;
  score1: number | null;
  score2: number | null;
  winner: string | null;
  status: MatchStatus;
  scheduled_at: string;
  completed_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface CreateMatchRequest {
  sport: SportType;
  tournament_id: number;
  round: string;
  team1: string;
  team2: string;
  scheduled_at?: string;
}

export interface UpdateMatchRequest {
  round?: string;
  team1?: string;
  team2?: string;
  scheduled_at?: string;
  status?: MatchStatus;
}

export interface MatchResult {
  score1: number;
  score2: number;
  winner: string;
}

export interface MatchFilters {
  status?: MatchStatus;
  round?: string;
  limit?: number;
  offset?: number;
}

// エラー関連型
export interface APIError {
  code: string;
  message: string;
  status_code: number;
  details?: Record<string, any>;
}

// リクエスト設定型
export interface RequestOptions {
  headers?: Record<string, string>;
  timeout?: number;
  retry?: boolean;
  signal?: AbortSignal;
}

// WebSocket関連型
export interface UpdateNotification {
  type: string;
  sport: SportType;
  data: any;
  timestamp: string;
}

// 統計情報型
export interface TournamentStats {
  total_matches: number;
  completed_matches: number;
  pending_matches: number;
  completion_rate: number;
  last_updated: string;
}

export interface MatchStats {
  duration_minutes: number | null;
  total_points: number;
  sets_played: number;
  last_updated: string;
}

// HTTP メソッド型
export type HTTPMethod = 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH' | 'HEAD' | 'OPTIONS';

// エラーコード列挙型
export enum ErrorCode {
  // 認証関連
  AUTH_INVALID_CREDENTIALS = 'AUTH_INVALID_CREDENTIALS',
  AUTH_TOKEN_EXPIRED = 'AUTH_TOKEN_EXPIRED',
  AUTH_TOKEN_INVALID = 'AUTH_TOKEN_INVALID',
  AUTH_UNAUTHORIZED = 'AUTH_UNAUTHORIZED',
  AUTH_FORBIDDEN = 'AUTH_FORBIDDEN',

  // バリデーション関連
  VALIDATION_REQUIRED_FIELD = 'VALIDATION_REQUIRED_FIELD',
  VALIDATION_INVALID_FORMAT = 'VALIDATION_INVALID_FORMAT',
  VALIDATION_OUT_OF_RANGE = 'VALIDATION_OUT_OF_RANGE',
  VALIDATION_DUPLICATE_VALUE = 'VALIDATION_DUPLICATE_VALUE',

  // リソース関連
  RESOURCE_NOT_FOUND = 'RESOURCE_NOT_FOUND',
  RESOURCE_ALREADY_EXISTS = 'RESOURCE_ALREADY_EXISTS',
  RESOURCE_CONFLICT = 'RESOURCE_CONFLICT',

  // ビジネスロジック関連
  BUSINESS_TOURNAMENT_COMPLETED = 'BUSINESS_TOURNAMENT_COMPLETED',
  BUSINESS_MATCH_ALREADY_COMPLETED = 'BUSINESS_MATCH_ALREADY_COMPLETED',
  BUSINESS_INVALID_MATCH_RESULT = 'BUSINESS_INVALID_MATCH_RESULT',

  // システム関連
  SYSTEM_DATABASE_ERROR = 'SYSTEM_DATABASE_ERROR',
  SYSTEM_NETWORK_ERROR = 'SYSTEM_NETWORK_ERROR',
  SYSTEM_TIMEOUT = 'SYSTEM_TIMEOUT',
  SYSTEM_UNKNOWN_ERROR = 'SYSTEM_UNKNOWN_ERROR'
}
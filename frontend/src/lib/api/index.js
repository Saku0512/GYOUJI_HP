// API統合エクスポート
// 統一APIクライアントと後方互換性のための既存APIクライアントをエクスポート

// 統一APIクライアント（推奨）
export { unifiedAPI, UnifiedAPI } from './unified-client.js';

// 型定義
export * from './types.js';

// 後方互換性のための既存APIクライアント
export { AuthAPI, authAPI } from './auth.js';
export { TournamentAPI, tournamentAPI } from './tournament.js';
export { MatchAPI, matchAPI } from './matches.js';

// 旧クライアント（非推奨だが互換性のため残す）
export { apiClient } from './client.js';

// デフォルトエクスポートは統一APIクライアント
export { unifiedAPI as default } from './unified-client.js';
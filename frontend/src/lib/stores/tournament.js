// トーナメント状態管理ストア
import { writable, get } from 'svelte/store';
import { tournamentAPI } from '../api/tournament.js';
import { matchAPI } from '../api/matches.js';

// トーナメント状態の初期値
const initialTournamentState = {
  tournaments: {}, // スポーツ別のトーナメントデータ
  currentSport: 'volleyball', // 現在選択されているスポーツ
  loading: false, // ローディング状態
  error: null, // エラー情報
  lastUpdated: null, // 最終更新時刻
  cache: {}, // データキャッシュ
  pollingInterval: null // ポーリング間隔ID
};

// トーナメントストアの作成
const { subscribe, set, update } = writable(initialTournamentState);

// データキャッシュの有効期限（5分）
const CACHE_DURATION = 5 * 60 * 1000;

// ポーリング間隔（30秒）
const POLLING_INTERVAL = 30 * 1000;

// サポートされているスポーツ
const SUPPORTED_SPORTS = ['volleyball', 'table_tennis', 'soccer'];

/**
 * ローディング状態を設定
 */
function setLoading(loading) {
  update(state => ({
    ...state,
    loading
  }));
}

/**
 * エラー状態を設定
 */
function setError(error) {
  update(state => ({
    ...state,
    error,
    loading: false
  }));
}

/**
 * エラー状態をクリア
 */
function clearError() {
  update(state => ({
    ...state,
    error: null
  }));
}

/**
 * スポーツ名の検証
 */
function validateSport(sport) {
  if (!sport) {
    throw new Error('スポーツ名が指定されていません');
  }
  
  if (!SUPPORTED_SPORTS.includes(sport)) {
    throw new Error(`サポートされていないスポーツです: ${sport}`);
  }
  
  return true;
}

/**
 * キャッシュの有効性をチェック
 */
function isCacheValid(sport) {
  const state = get({ subscribe });
  const cached = state.cache[sport];
  
  if (!cached || !cached.timestamp) {
    return false;
  }
  
  return Date.now() - cached.timestamp < CACHE_DURATION;
}

/**
 * キャッシュからデータを取得
 */
function getCachedData(sport) {
  const state = get({ subscribe });
  const cached = state.cache[sport];
  
  if (cached && isCacheValid(sport)) {
    return cached.data;
  }
  
  return null;
}

/**
 * データをキャッシュに保存
 */
function setCacheData(sport, data) {
  update(state => ({
    ...state,
    cache: {
      ...state.cache,
      [sport]: {
        data,
        timestamp: Date.now()
      }
    }
  }));
}

/**
 * トーナメントデータを取得
 */
async function fetchTournaments(sport = null, useCache = true) {
  try {
    clearError();
    
    // 特定のスポーツが指定された場合
    if (sport) {
      validateSport(sport);
      
      // キャッシュチェック
      if (useCache) {
        const cachedData = getCachedData(sport);
        if (cachedData) {
          update(state => ({
            ...state,
            tournaments: {
              ...state.tournaments,
              [sport]: cachedData
            }
          }));
          return {
            success: true,
            data: cachedData,
            message: `${sport}のトーナメントデータを取得しました（キャッシュ）`
          };
        }
      }
      
      setLoading(true);
      
      // APIからデータ取得
      const response = await tournamentAPI.getTournament(sport);
      
      if (response.success) {
        // データを状態とキャッシュに保存
        update(state => ({
          ...state,
          tournaments: {
            ...state.tournaments,
            [sport]: response.data
          },
          loading: false,
          lastUpdated: Date.now()
        }));
        
        setCacheData(sport, response.data);
        
        return response;
      } else {
        setError(response.message || 'トーナメントデータの取得に失敗しました');
        return response;
      }
    } else {
      // 全スポーツのデータを取得
      setLoading(true);
      
      const response = await tournamentAPI.getTournaments();
      
      if (response.success) {
        update(state => ({
          ...state,
          tournaments: response.data,
          loading: false,
          lastUpdated: Date.now()
        }));
        
        // 各スポーツのデータをキャッシュに保存
        if (response.data && typeof response.data === 'object') {
          Object.keys(response.data).forEach(sportKey => {
            setCacheData(sportKey, response.data[sportKey]);
          });
        }
        
        return response;
      } else {
        setError(response.message || 'トーナメントデータの取得に失敗しました');
        return response;
      }
    }
  } catch (error) {
    console.error('Fetch tournaments error:', error);
    setError(error.message || '予期しないエラーが発生しました');
    return {
      success: false,
      error: 'FETCH_TOURNAMENTS_ERROR',
      message: error.message || 'トーナメントデータの取得に失敗しました'
    };
  }
}

/**
 * 試合結果を更新
 */
async function updateMatch(matchId, result) {
  try {
    clearError();
    
    if (!matchId) {
      throw new Error('試合IDが指定されていません');
    }
    
    if (!result || typeof result !== 'object') {
      throw new Error('試合結果データが正しくありません');
    }
    
    setLoading(true);
    
    // APIで試合結果を更新
    const response = await matchAPI.updateMatch(matchId, result);
    
    if (response.success) {
      // 成功時は関連するトーナメントデータを再取得
      const state = get({ subscribe });
      const currentSport = state.currentSport;
      
      // キャッシュをクリアして最新データを取得
      await fetchTournaments(currentSport, false);
      
      return {
        success: true,
        data: response.data,
        message: '試合結果を更新しました'
      };
    } else {
      setError(response.message || '試合結果の更新に失敗しました');
      return response;
    }
  } catch (error) {
    console.error('Update match error:', error);
    setError(error.message || '予期しないエラーが発生しました');
    return {
      success: false,
      error: 'UPDATE_MATCH_ERROR',
      message: error.message || '試合結果の更新に失敗しました'
    };
  }
}

/**
 * スポーツを切り替え
 */
function switchSport(sport) {
  try {
    validateSport(sport);
    
    update(state => ({
      ...state,
      currentSport: sport,
      error: null
    }));
    
    // 新しいスポーツのデータを取得
    fetchTournaments(sport);
    
    return {
      success: true,
      message: `スポーツを${sport}に切り替えました`
    };
  } catch (error) {
    console.error('Switch sport error:', error);
    setError(error.message || 'スポーツの切り替えに失敗しました');
    return {
      success: false,
      error: 'SWITCH_SPORT_ERROR',
      message: error.message || 'スポーツの切り替えに失敗しました'
    };
  }
}

/**
 * データを強制的にリフレッシュ
 */
async function refreshData(sport = null) {
  try {
    clearError();
    
    const state = get({ subscribe });
    const targetSport = sport || state.currentSport;
    
    // キャッシュを無視して最新データを取得
    const response = await fetchTournaments(targetSport, false);
    
    return {
      success: true,
      message: 'データを更新しました'
    };
  } catch (error) {
    console.error('Refresh data error:', error);
    setError(error.message || 'データの更新に失敗しました');
    return {
      success: false,
      error: 'REFRESH_DATA_ERROR',
      message: error.message || 'データの更新に失敗しました'
    };
  }
}

/**
 * ポーリングを開始
 */
function startPolling() {
  const state = get({ subscribe });
  
  // 既にポーリングが開始されている場合は何もしない
  if (state.pollingInterval) {
    return;
  }
  
  const intervalId = setInterval(async () => {
    // ページが非表示の場合はポーリングをスキップ
    if (typeof document !== 'undefined' && document.hidden) {
      return;
    }
    
    const currentState = get({ subscribe });
    
    // ローディング中の場合はスキップ
    if (currentState.loading) {
      return;
    }
    
    try {
      await refreshData();
    } catch (error) {
      console.error('Polling error:', error);
    }
  }, POLLING_INTERVAL);
  
  update(state => ({
    ...state,
    pollingInterval: intervalId
  }));
}

/**
 * ポーリングを停止
 */
function stopPolling() {
  update(state => {
    if (state.pollingInterval) {
      clearInterval(state.pollingInterval);
    }
    
    return {
      ...state,
      pollingInterval: null
    };
  });
}

/**
 * ストアの初期化
 */
async function initialize() {
  try {
    // 初期データを取得
    await fetchTournaments();
    
    // ポーリングを開始
    startPolling();
    
    // ページの可視性変更イベントを監視
    if (typeof document !== 'undefined') {
      document.addEventListener('visibilitychange', () => {
        if (!document.hidden) {
          // ページが表示されたときにデータを更新
          refreshData();
        }
      });
    }
    
    return {
      success: true,
      message: 'トーナメントストアを初期化しました'
    };
  } catch (error) {
    console.error('Initialize tournament store error:', error);
    setError(error.message || '初期化に失敗しました');
    return {
      success: false,
      error: 'INITIALIZE_ERROR',
      message: error.message || '初期化に失敗しました'
    };
  }
}

/**
 * ストアのクリーンアップ
 */
function cleanup() {
  stopPolling();
  
  set(initialTournamentState);
  
  return {
    success: true,
    message: 'トーナメントストアをクリーンアップしました'
  };
}

/**
 * 現在のトーナメントデータを取得
 */
function getCurrentTournament() {
  const state = get({ subscribe });
  return state.tournaments[state.currentSport] || null;
}

/**
 * 特定スポーツのトーナメントデータを取得
 */
function getTournamentBySport(sport) {
  validateSport(sport);
  const state = get({ subscribe });
  return state.tournaments[sport] || null;
}

/**
 * サポートされているスポーツ一覧を取得
 */
function getSupportedSports() {
  return [...SUPPORTED_SPORTS];
}

// エクスポートするストアオブジェクト
export const tournamentStore = {
  subscribe,
  
  // アクション
  fetchTournaments,
  updateMatch,
  switchSport,
  refreshData,
  initialize,
  cleanup,
  
  // ポーリング制御
  startPolling,
  stopPolling,
  
  // ユーティリティ
  getCurrentTournament,
  getTournamentBySport,
  getSupportedSports,
  
  // 状態管理
  setLoading,
  setError,
  clearError
};

import { W as current_component } from "./index2.js";
import { w as writable, g as get } from "./index.js";
import { a as apiClient } from "./client.js";
function onDestroy(fn) {
  var context = (
    /** @type {Component} */
    current_component
  );
  (context.d ??= []).push(fn);
}
class TournamentAPI {
  constructor(client = apiClient) {
    this.client = client;
    this.supportedSports = ["volleyball", "table_tennis", "soccer"];
  }
  /**
   * スポーツ名の検証
   */
  validateSport(sport) {
    if (!sport) {
      throw new Error("スポーツ名が指定されていません");
    }
    if (!this.supportedSports.includes(sport)) {
      throw new Error(`サポートされていないスポーツです: ${sport}`);
    }
    return true;
  }
  /**
   * トーナメント一覧取得
   * 全スポーツのトーナメント情報を取得
   */
  async getTournaments() {
    try {
      const response = await this.client.get("/tournaments");
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: "トーナメント一覧を取得しました"
        };
      }
      return response;
    } catch (error) {
      console.error("Get tournaments error:", error);
      return {
        success: false,
        error: "GET_TOURNAMENTS_ERROR",
        message: "トーナメント一覧の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 特定スポーツのトーナメント取得
   * 指定されたスポーツのトーナメント情報を取得
   */
  async getTournament(sport) {
    try {
      this.validateSport(sport);
      const response = await this.client.get(`/tournaments/${sport}`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメント情報を取得しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Get tournament error:", error);
      return {
        success: false,
        error: "GET_TOURNAMENT_ERROR",
        message: "トーナメント情報の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * トーナメントブラケット取得
   * 指定されたスポーツのブラケット情報（試合組み合わせ）を取得
   */
  async getTournamentBracket(sport) {
    try {
      this.validateSport(sport);
      const response = await this.client.get(`/tournaments/${sport}/bracket`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のブラケット情報を取得しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Get tournament bracket error:", error);
      return {
        success: false,
        error: "GET_BRACKET_ERROR",
        message: "ブラケット情報の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * トーナメント形式更新
   * 卓球の晴天時/雨天時形式切り替えなど
   */
  async updateTournamentFormat(sport, format) {
    try {
      this.validateSport(sport);
      if (!format) {
        throw new Error("形式が指定されていません");
      }
      const response = await this.client.put(`/tournaments/${sport}/format`, {
        format
      });
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメント形式を${format}に更新しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Update tournament format error:", error);
      return {
        success: false,
        error: "UPDATE_FORMAT_ERROR",
        message: "トーナメント形式の更新に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * トーナメント作成
   * 新しいトーナメントを作成（管理者用）
   */
  async createTournament(tournamentData) {
    try {
      const { sport, format, teams } = tournamentData;
      this.validateSport(sport);
      if (!format) {
        throw new Error("トーナメント形式が指定されていません");
      }
      if (!teams || !Array.isArray(teams) || teams.length === 0) {
        throw new Error("参加チーム情報が正しくありません");
      }
      const response = await this.client.post("/tournaments", {
        sport,
        format,
        teams
      });
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメントを作成しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Create tournament error:", error);
      return {
        success: false,
        error: "CREATE_TOURNAMENT_ERROR",
        message: "トーナメントの作成に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * トーナメント削除
   * 指定されたトーナメントを削除（管理者用）
   */
  async deleteTournament(sport) {
    try {
      this.validateSport(sport);
      const response = await this.client.delete(`/tournaments/${sport}`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメントを削除しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Delete tournament error:", error);
      return {
        success: false,
        error: "DELETE_TOURNAMENT_ERROR",
        message: "トーナメントの削除に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * トーナメント状態更新
   * トーナメントの状態（開始、終了など）を更新
   */
  async updateTournamentStatus(sport, status) {
    try {
      this.validateSport(sport);
      const validStatuses = ["pending", "active", "completed", "cancelled"];
      if (!validStatuses.includes(status)) {
        throw new Error(`無効なステータスです: ${status}`);
      }
      const response = await this.client.patch(`/tournaments/${sport}/status`, {
        status
      });
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}のトーナメント状態を${status}に更新しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Update tournament status error:", error);
      return {
        success: false,
        error: "UPDATE_STATUS_ERROR",
        message: "トーナメント状態の更新に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * トーナメント統計情報取得
   * 試合数、完了率などの統計情報を取得
   */
  async getTournamentStats(sport) {
    try {
      this.validateSport(sport);
      const response = await this.client.get(`/tournaments/${sport}/stats`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の統計情報を取得しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Get tournament stats error:", error);
      return {
        success: false,
        error: "GET_STATS_ERROR",
        message: "統計情報の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 利用可能な形式一覧取得
   * 指定されたスポーツで利用可能なトーナメント形式を取得
   */
  async getAvailableFormats(sport) {
    try {
      this.validateSport(sport);
      const response = await this.client.get(`/tournaments/${sport}/formats`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の利用可能な形式一覧を取得しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Get available formats error:", error);
      return {
        success: false,
        error: "GET_FORMATS_ERROR",
        message: "形式一覧の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * サポートされているスポーツ一覧を取得
   */
  getSupportedSports() {
    return {
      success: true,
      data: this.supportedSports,
      message: "サポートされているスポーツ一覧"
    };
  }
}
const tournamentAPI = new TournamentAPI();
class MatchAPI {
  constructor(client = apiClient) {
    this.client = client;
    this.supportedSports = ["volleyball", "table_tennis", "soccer"];
    this.validStatuses = ["pending", "in_progress", "completed", "cancelled"];
  }
  /**
   * スポーツ名の検証
   */
  validateSport(sport) {
    if (!sport) {
      throw new Error("スポーツ名が指定されていません");
    }
    if (!this.supportedSports.includes(sport)) {
      throw new Error(`サポートされていないスポーツです: ${sport}`);
    }
    return true;
  }
  /**
   * 試合IDの検証
   */
  validateMatchId(matchId) {
    if (!matchId) {
      throw new Error("試合IDが指定されていません");
    }
    if (typeof matchId !== "number" && typeof matchId !== "string") {
      throw new Error("試合IDの形式が正しくありません");
    }
    return true;
  }
  /**
   * 試合結果データの検証
   */
  validateMatchResult(result) {
    if (!result || typeof result !== "object") {
      throw new Error("試合結果データが正しくありません");
    }
    const { score1, score2, winner } = result;
    if (score1 !== void 0 && (typeof score1 !== "number" || score1 < 0)) {
      throw new Error("チーム1のスコアが正しくありません");
    }
    if (score2 !== void 0 && (typeof score2 !== "number" || score2 < 0)) {
      throw new Error("チーム2のスコアが正しくありません");
    }
    if (score1 !== void 0 && score2 !== void 0 && winner !== void 0) {
      if (typeof winner !== "string" || winner.trim() === "") {
        throw new Error("勝者の情報が正しくありません");
      }
    }
    return true;
  }
  /**
   * 試合一覧取得
   * 指定されたスポーツの全試合を取得
   */
  async getMatches(sport, options = {}) {
    try {
      this.validateSport(sport);
      const queryParams = new URLSearchParams();
      if (options.status) {
        queryParams.append("status", options.status);
      }
      if (options.round) {
        queryParams.append("round", options.round);
      }
      if (options.limit) {
        queryParams.append("limit", options.limit.toString());
      }
      if (options.offset) {
        queryParams.append("offset", options.offset.toString());
      }
      const queryString = queryParams.toString();
      const endpoint = `/matches/${sport}${queryString ? `?${queryString}` : ""}`;
      const response = await this.client.get(endpoint);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の試合一覧を取得しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Get matches error:", error);
      return {
        success: false,
        error: "GET_MATCHES_ERROR",
        message: "試合一覧の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 特定試合の詳細取得
   * 指定されたIDの試合詳細情報を取得
   */
  async getMatch(matchId) {
    try {
      this.validateMatchId(matchId);
      const response = await this.client.get(`/matches/${matchId}`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: "試合詳細を取得しました"
        };
      }
      return response;
    } catch (error) {
      console.error("Get match error:", error);
      return {
        success: false,
        error: "GET_MATCH_ERROR",
        message: "試合詳細の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 試合結果更新
   * 指定された試合の結果を更新
   */
  async updateMatch(matchId, result) {
    try {
      this.validateMatchId(matchId);
      this.validateMatchResult(result);
      const response = await this.client.put(`/matches/${matchId}`, result);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: "試合結果を更新しました"
        };
      }
      return response;
    } catch (error) {
      console.error("Update match error:", error);
      return {
        success: false,
        error: "UPDATE_MATCH_ERROR",
        message: "試合結果の更新に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 新規試合作成
   * 新しい試合を作成（管理者用）
   */
  async createMatch(matchData) {
    try {
      if (!matchData || typeof matchData !== "object") {
        throw new Error("試合データが正しくありません");
      }
      const { sport, tournament_id, round, team1, team2, scheduled_at } = matchData;
      if (!sport) {
        throw new Error("スポーツが指定されていません");
      }
      this.validateSport(sport);
      if (!tournament_id) {
        throw new Error("トーナメントIDが指定されていません");
      }
      if (!round) {
        throw new Error("ラウンドが指定されていません");
      }
      if (!team1 || !team2) {
        throw new Error("対戦チームが正しく指定されていません");
      }
      if (team1 === team2) {
        throw new Error("同じチーム同士の試合は作成できません");
      }
      const response = await this.client.post("/matches", matchData);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: "新しい試合を作成しました"
        };
      }
      return response;
    } catch (error) {
      console.error("Create match error:", error);
      return {
        success: false,
        error: "CREATE_MATCH_ERROR",
        message: "試合の作成に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 試合削除
   * 指定された試合を削除（管理者用）
   */
  async deleteMatch(matchId) {
    try {
      this.validateMatchId(matchId);
      const response = await this.client.delete(`/matches/${matchId}`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: "試合を削除しました"
        };
      }
      return response;
    } catch (error) {
      console.error("Delete match error:", error);
      return {
        success: false,
        error: "DELETE_MATCH_ERROR",
        message: "試合の削除に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 試合状態更新
   * 試合の状態（開始、終了など）を更新
   */
  async updateMatchStatus(matchId, status) {
    try {
      this.validateMatchId(matchId);
      if (!this.validStatuses.includes(status)) {
        throw new Error(`無効なステータスです: ${status}`);
      }
      const response = await this.client.patch(`/matches/${matchId}/status`, {
        status
      });
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `試合状態を${status}に更新しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Update match status error:", error);
      return {
        success: false,
        error: "UPDATE_MATCH_STATUS_ERROR",
        message: "試合状態の更新に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 未完了試合一覧取得
   * 管理者ダッシュボード用の未完了試合一覧を取得
   */
  async getPendingMatches(sport) {
    try {
      this.validateSport(sport);
      const response = await this.client.get(`/matches/${sport}/pending`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の未完了試合一覧を取得しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Get pending matches error:", error);
      return {
        success: false,
        error: "GET_PENDING_MATCHES_ERROR",
        message: "未完了試合一覧の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 試合結果の一括更新
   * 複数の試合結果を一度に更新
   */
  async updateMultipleMatches(updates) {
    try {
      if (!Array.isArray(updates) || updates.length === 0) {
        throw new Error("更新データが正しくありません");
      }
      for (const update2 of updates) {
        if (!update2.matchId) {
          throw new Error("試合IDが指定されていない更新データがあります");
        }
        this.validateMatchId(update2.matchId);
        this.validateMatchResult(update2.result);
      }
      const response = await this.client.put("/matches/batch", {
        updates
      });
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${updates.length}件の試合結果を更新しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Update multiple matches error:", error);
      return {
        success: false,
        error: "UPDATE_MULTIPLE_MATCHES_ERROR",
        message: "試合結果の一括更新に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 試合統計情報取得
   * 指定された試合の統計情報を取得
   */
  async getMatchStats(matchId) {
    try {
      this.validateMatchId(matchId);
      const response = await this.client.get(`/matches/${matchId}/stats`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: "試合統計情報を取得しました"
        };
      }
      return response;
    } catch (error) {
      console.error("Get match stats error:", error);
      return {
        success: false,
        error: "GET_MATCH_STATS_ERROR",
        message: "試合統計情報の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * 次の試合取得
   * 指定されたスポーツの次に予定されている試合を取得
   */
  async getNextMatch(sport) {
    try {
      this.validateSport(sport);
      const response = await this.client.get(`/matches/${sport}/next`);
      if (response.success) {
        return {
          success: true,
          data: response.data,
          message: `${sport}の次の試合情報を取得しました`
        };
      }
      return response;
    } catch (error) {
      console.error("Get next match error:", error);
      return {
        success: false,
        error: "GET_NEXT_MATCH_ERROR",
        message: "次の試合情報の取得に失敗しました",
        details: error.message
      };
    }
  }
  /**
   * サポートされているスポーツ一覧を取得
   */
  getSupportedSports() {
    return {
      success: true,
      data: this.supportedSports,
      message: "サポートされているスポーツ一覧"
    };
  }
  /**
   * 有効なステータス一覧を取得
   */
  getValidStatuses() {
    return {
      success: true,
      data: this.validStatuses,
      message: "有効なステータス一覧"
    };
  }
}
const matchAPI = new MatchAPI();
const initialTournamentState = {
  tournaments: {},
  // スポーツ別のトーナメントデータ
  currentSport: "volleyball",
  // 現在選択されているスポーツ
  loading: false,
  // ローディング状態
  error: null,
  // エラー情報
  lastUpdated: null,
  // 最終更新時刻
  cache: {},
  // データキャッシュ
  pollingInterval: null
  // ポーリング間隔ID
};
const { subscribe, set, update } = writable(initialTournamentState);
const CACHE_DURATION = 5 * 60 * 1e3;
const POLLING_INTERVAL = 30 * 1e3;
const SUPPORTED_SPORTS = ["volleyball", "table_tennis", "soccer"];
function setLoading(loading) {
  update((state) => ({
    ...state,
    loading
  }));
}
function setError(error) {
  update((state) => ({
    ...state,
    error,
    loading: false
  }));
}
function clearError() {
  update((state) => ({
    ...state,
    error: null
  }));
}
function validateSport(sport) {
  if (!sport) {
    throw new Error("スポーツ名が指定されていません");
  }
  if (!SUPPORTED_SPORTS.includes(sport)) {
    throw new Error(`サポートされていないスポーツです: ${sport}`);
  }
  return true;
}
function isCacheValid(sport) {
  const state = get({ subscribe });
  const cached = state.cache[sport];
  if (!cached || !cached.timestamp) {
    return false;
  }
  return Date.now() - cached.timestamp < CACHE_DURATION;
}
function getCachedData(sport) {
  const state = get({ subscribe });
  const cached = state.cache[sport];
  if (cached && isCacheValid(sport)) {
    return cached.data;
  }
  return null;
}
function setCacheData(sport, data) {
  update((state) => ({
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
async function fetchTournaments(sport = null, useCache = true) {
  try {
    clearError();
    if (sport) {
      validateSport(sport);
      if (useCache) {
        const cachedData = getCachedData(sport);
        if (cachedData) {
          update((state) => ({
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
      const response = await tournamentAPI.getTournament(sport);
      if (response.success) {
        update((state) => ({
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
        setError(response.message || "トーナメントデータの取得に失敗しました");
        return response;
      }
    } else {
      setLoading(true);
      const response = await tournamentAPI.getTournaments();
      if (response.success) {
        update((state) => ({
          ...state,
          tournaments: response.data,
          loading: false,
          lastUpdated: Date.now()
        }));
        if (response.data && typeof response.data === "object") {
          Object.keys(response.data).forEach((sportKey) => {
            setCacheData(sportKey, response.data[sportKey]);
          });
        }
        return response;
      } else {
        setError(response.message || "トーナメントデータの取得に失敗しました");
        return response;
      }
    }
  } catch (error) {
    console.error("Fetch tournaments error:", error);
    setError(error.message || "予期しないエラーが発生しました");
    return {
      success: false,
      error: "FETCH_TOURNAMENTS_ERROR",
      message: error.message || "トーナメントデータの取得に失敗しました"
    };
  }
}
async function updateMatch(matchId, result) {
  try {
    clearError();
    if (!matchId) {
      throw new Error("試合IDが指定されていません");
    }
    if (!result || typeof result !== "object") {
      throw new Error("試合結果データが正しくありません");
    }
    setLoading(true);
    const response = await matchAPI.updateMatch(matchId, result);
    if (response.success) {
      const state = get({ subscribe });
      const currentSport = state.currentSport;
      await fetchTournaments(currentSport, false);
      return {
        success: true,
        data: response.data,
        message: "試合結果を更新しました"
      };
    } else {
      setError(response.message || "試合結果の更新に失敗しました");
      return response;
    }
  } catch (error) {
    console.error("Update match error:", error);
    setError(error.message || "予期しないエラーが発生しました");
    return {
      success: false,
      error: "UPDATE_MATCH_ERROR",
      message: error.message || "試合結果の更新に失敗しました"
    };
  }
}
function switchSport(sport) {
  try {
    validateSport(sport);
    update((state) => ({
      ...state,
      currentSport: sport,
      error: null
    }));
    fetchTournaments(sport);
    return {
      success: true,
      message: `スポーツを${sport}に切り替えました`
    };
  } catch (error) {
    console.error("Switch sport error:", error);
    setError(error.message || "スポーツの切り替えに失敗しました");
    return {
      success: false,
      error: "SWITCH_SPORT_ERROR",
      message: error.message || "スポーツの切り替えに失敗しました"
    };
  }
}
async function refreshData(sport = null) {
  try {
    clearError();
    const state = get({ subscribe });
    const targetSport = sport || state.currentSport;
    const response = await fetchTournaments(targetSport, false);
    return {
      success: true,
      message: "データを更新しました"
    };
  } catch (error) {
    console.error("Refresh data error:", error);
    setError(error.message || "データの更新に失敗しました");
    return {
      success: false,
      error: "REFRESH_DATA_ERROR",
      message: error.message || "データの更新に失敗しました"
    };
  }
}
function startPolling() {
  const state = get({ subscribe });
  if (state.pollingInterval) {
    return;
  }
  const intervalId = setInterval(async () => {
    if (typeof document !== "undefined" && document.hidden) {
      return;
    }
    const currentState = get({ subscribe });
    if (currentState.loading) {
      return;
    }
    try {
      await refreshData();
    } catch (error) {
      console.error("Polling error:", error);
    }
  }, POLLING_INTERVAL);
  update((state2) => ({
    ...state2,
    pollingInterval: intervalId
  }));
}
function stopPolling() {
  update((state) => {
    if (state.pollingInterval) {
      clearInterval(state.pollingInterval);
    }
    return {
      ...state,
      pollingInterval: null
    };
  });
}
async function initialize() {
  try {
    await fetchTournaments();
    startPolling();
    if (typeof document !== "undefined") {
      document.addEventListener("visibilitychange", () => {
        if (!document.hidden) {
          refreshData();
        }
      });
    }
    return {
      success: true,
      message: "トーナメントストアを初期化しました"
    };
  } catch (error) {
    console.error("Initialize tournament store error:", error);
    setError(error.message || "初期化に失敗しました");
    return {
      success: false,
      error: "INITIALIZE_ERROR",
      message: error.message || "初期化に失敗しました"
    };
  }
}
function cleanup() {
  stopPolling();
  set(initialTournamentState);
  return {
    success: true,
    message: "トーナメントストアをクリーンアップしました"
  };
}
function getCurrentTournament() {
  const state = get({ subscribe });
  return state.tournaments[state.currentSport] || null;
}
function getTournamentBySport(sport) {
  validateSport(sport);
  const state = get({ subscribe });
  return state.tournaments[sport] || null;
}
function getSupportedSports() {
  return [...SUPPORTED_SPORTS];
}
const tournamentStore = {
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
export {
  onDestroy as o,
  tournamentStore as t
};

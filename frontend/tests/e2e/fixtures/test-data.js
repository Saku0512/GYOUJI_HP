/**
 * E2Eテスト用のテストデータフィクスチャ
 */

export const mockUsers = {
  admin: {
    id: 1,
    username: 'admin',
    role: 'admin',
    token: 'mock-admin-jwt-token'
  },
  user: {
    id: 2,
    username: 'user',
    role: 'user',
    token: 'mock-user-jwt-token'
  }
};

export const mockTournaments = {
  volleyball: {
    tournament: {
      id: 1,
      sport: 'volleyball',
      format: 'single_elimination',
      status: 'active',
      created_at: '2024-01-15T09:00:00Z',
      updated_at: '2024-01-15T10:00:00Z'
    },
    bracket: {
      rounds: [
        {
          name: '準々決勝',
          matches: [
            {
              id: 1,
              tournament_id: 1,
              round: '準々決勝',
              team1: 'チームA',
              team2: 'チームB',
              score1: null,
              score2: null,
              winner: null,
              status: 'pending',
              scheduled_at: '2024-01-15T10:00:00Z'
            },
            {
              id: 2,
              tournament_id: 1,
              round: '準々決勝',
              team1: 'チームC',
              team2: 'チームD',
              score1: null,
              score2: null,
              winner: null,
              status: 'pending',
              scheduled_at: '2024-01-15T11:00:00Z'
            }
          ]
        },
        {
          name: '準決勝',
          matches: [
            {
              id: 3,
              tournament_id: 1,
              round: '準決勝',
              team1: 'TBD',
              team2: 'TBD',
              score1: null,
              score2: null,
              winner: null,
              status: 'pending',
              scheduled_at: '2024-01-15T14:00:00Z'
            }
          ]
        },
        {
          name: '決勝',
          matches: [
            {
              id: 4,
              tournament_id: 1,
              round: '決勝',
              team1: 'TBD',
              team2: 'TBD',
              score1: null,
              score2: null,
              winner: null,
              status: 'pending',
              scheduled_at: '2024-01-15T16:00:00Z'
            }
          ]
        }
      ]
    }
  },

  table_tennis: {
    tournament: {
      id: 2,
      sport: 'table_tennis',
      format: 'sunny_weather',
      status: 'active',
      created_at: '2024-01-15T09:00:00Z',
      updated_at: '2024-01-15T10:00:00Z'
    },
    bracket: {
      rounds: [
        {
          name: '1回戦',
          matches: [
            {
              id: 5,
              tournament_id: 2,
              round: '1回戦',
              team1: 'チームE',
              team2: 'チームF',
              score1: 2,
              score2: 0,
              winner: 'チームE',
              status: 'completed',
              scheduled_at: '2024-01-15T12:00:00Z',
              completed_at: '2024-01-15T12:30:00Z'
            },
            {
              id: 6,
              tournament_id: 2,
              round: '1回戦',
              team1: 'チームG',
              team2: 'チームH',
              score1: null,
              score2: null,
              winner: null,
              status: 'pending',
              scheduled_at: '2024-01-15T13:00:00Z'
            }
          ]
        },
        {
          name: '決勝',
          matches: [
            {
              id: 7,
              tournament_id: 2,
              round: '決勝',
              team1: 'チームE',
              team2: 'TBD',
              score1: null,
              score2: null,
              winner: null,
              status: 'pending',
              scheduled_at: '2024-01-15T15:00:00Z'
            }
          ]
        }
      ]
    }
  },

  soccer: {
    tournament: {
      id: 3,
      sport: 'soccer',
      format: 'single_elimination',
      status: 'active',
      created_at: '2024-01-15T09:00:00Z',
      updated_at: '2024-01-15T10:00:00Z'
    },
    bracket: {
      rounds: [
        {
          name: '1回戦',
          matches: [
            {
              id: 8,
              tournament_id: 3,
              round: '1回戦',
              team1: 'チームI',
              team2: 'チームJ',
              score1: null,
              score2: null,
              winner: null,
              status: 'pending',
              scheduled_at: '2024-01-15T14:00:00Z'
            },
            {
              id: 9,
              tournament_id: 3,
              round: '1回戦',
              team1: 'チームK',
              team2: 'チームL',
              score1: 1,
              score2: 2,
              winner: 'チームL',
              status: 'completed',
              scheduled_at: '2024-01-15T15:00:00Z',
              completed_at: '2024-01-15T16:00:00Z'
            }
          ]
        }
      ]
    }
  }
};

export const mockAdminTournaments = {
  volleyball: {
    tournament: mockTournaments.volleyball.tournament,
    pendingMatches: [
      {
        id: 1,
        tournament_id: 1,
        round: '準々決勝',
        team1: 'チームA',
        team2: 'チームB',
        score1: null,
        score2: null,
        winner: null,
        status: 'pending',
        scheduled_at: '2024-01-15T10:00:00Z'
      },
      {
        id: 2,
        tournament_id: 1,
        round: '準々決勝',
        team1: 'チームC',
        team2: 'チームD',
        score1: null,
        score2: null,
        winner: null,
        status: 'pending',
        scheduled_at: '2024-01-15T11:00:00Z'
      }
    ]
  },

  table_tennis: {
    tournament: mockTournaments.table_tennis.tournament,
    pendingMatches: [
      {
        id: 6,
        tournament_id: 2,
        round: '1回戦',
        team1: 'チームG',
        team2: 'チームH',
        score1: null,
        score2: null,
        winner: null,
        status: 'pending',
        scheduled_at: '2024-01-15T13:00:00Z'
      }
    ]
  },

  soccer: {
    tournament: mockTournaments.soccer.tournament,
    pendingMatches: [
      {
        id: 8,
        tournament_id: 3,
        round: '1回戦',
        team1: 'チームI',
        team2: 'チームJ',
        score1: null,
        score2: null,
        winner: null,
        status: 'pending',
        scheduled_at: '2024-01-15T14:00:00Z'
      }
    ]
  }
};

export const mockApiResponses = {
  // 認証関連
  loginSuccess: {
    success: true,
    data: {
      token: mockUsers.admin.token,
      user: mockUsers.admin
    }
  },

  loginFailure: {
    success: false,
    error: 'INVALID_CREDENTIALS',
    message: '認証情報が正しくありません'
  },

  logoutSuccess: {
    success: true,
    message: 'ログアウトしました'
  },

  tokenExpired: {
    success: false,
    error: 'TOKEN_EXPIRED',
    message: 'トークンが期限切れです'
  },

  // 試合結果更新
  matchUpdateSuccess: {
    success: true,
    message: '試合結果を更新しました'
  },

  matchUpdateFailure: {
    success: false,
    error: 'UPDATE_FAILED',
    message: '試合結果の更新に失敗しました'
  },

  // トーナメント形式変更
  formatChangeSuccess: {
    success: true,
    message: 'トーナメント形式を変更しました'
  },

  formatChangeFailure: {
    success: false,
    error: 'FORMAT_CHANGE_FAILED',
    message: 'トーナメント形式の変更に失敗しました'
  },

  // エラーレスポンス
  serverError: {
    success: false,
    error: 'SERVER_ERROR',
    message: 'サーバーエラーが発生しました'
  },

  networkError: {
    success: false,
    error: 'NETWORK_ERROR',
    message: 'ネットワークエラーが発生しました'
  },

  accessDenied: {
    success: false,
    error: 'ACCESS_DENIED',
    message: 'アクセス権限がありません'
  }
};

/**
 * 試合結果を更新したトーナメントデータを生成する
 * @param {string} sport - スポーツ名
 * @param {number} matchId - 試合ID
 * @param {number} score1 - チーム1のスコア
 * @param {number} score2 - チーム2のスコア
 * @returns {Object} 更新されたトーナメントデータ
 */
export function createUpdatedTournamentData(sport, matchId, score1, score2) {
  const tournamentData = JSON.parse(JSON.stringify(mockTournaments[sport]));
  
  // 指定された試合を見つけて更新
  for (const round of tournamentData.bracket.rounds) {
    const match = round.matches.find(m => m.id === matchId);
    if (match) {
      match.score1 = score1;
      match.score2 = score2;
      match.winner = score1 > score2 ? match.team1 : match.team2;
      match.status = 'completed';
      match.completed_at = new Date().toISOString();
      break;
    }
  }
  
  return tournamentData;
}

/**
 * 卓球の形式を変更したトーナメントデータを生成する
 * @param {string} format - 新しい形式 ('sunny_weather' | 'rainy_weather')
 * @returns {Object} 更新されたトーナメントデータ
 */
export function createUpdatedTableTennisFormat(format) {
  const tournamentData = JSON.parse(JSON.stringify(mockTournaments.table_tennis));
  tournamentData.tournament.format = format;
  tournamentData.tournament.updated_at = new Date().toISOString();
  
  return tournamentData;
}

/**
 * 管理者用の未完了試合リストから指定された試合を削除する
 * @param {string} sport - スポーツ名
 * @param {number} matchId - 完了した試合ID
 * @returns {Object} 更新された管理者用トーナメントデータ
 */
export function createUpdatedAdminTournamentData(sport, matchId) {
  const adminData = JSON.parse(JSON.stringify(mockAdminTournaments[sport]));
  adminData.pendingMatches = adminData.pendingMatches.filter(match => match.id !== matchId);
  
  return adminData;
}
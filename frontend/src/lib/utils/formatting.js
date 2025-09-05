// データフォーマットユーティリティ

// 日付フォーマット
export function formatDate(dateString) {
  if (!dateString) return '';

  const date = new Date(dateString);
  return date.toLocaleDateString('ja-JP', {
    year: 'numeric',
    month: 'long',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit'
  });
}

// スポーツ名の日本語変換
export function formatSportName(sport) {
  const sportNames = {
    volleyball: 'バレーボール',
    table_tennis: '卓球',
    soccer: 'サッカー'
  };

  return sportNames[sport] || sport;
}

// 試合ステータスの日本語変換
export function formatMatchStatus(status) {
  const statusNames = {
    pending: '未実施',
    in_progress: '進行中',
    completed: '完了',
    cancelled: 'キャンセル'
  };

  return statusNames[status] || status;
}

// スコア表示フォーマット
export function formatScore(score1, score2) {
  if (score1 === null || score1 === undefined || score2 === null || score2 === undefined) {
    return '未実施';
  }

  return `${score1} - ${score2}`;
}

// トーナメント形式の日本語変換
export function formatTournamentFormat(format) {
  const formatNames = {
    sunny: '晴天時形式',
    rainy: '雨天時形式',
    single_elimination: 'シングルエリミネーション',
    double_elimination: 'ダブルエリミネーション'
  };

  return formatNames[format] || format;
}

// チーム名の短縮表示
export function formatTeamName(teamName, maxLength = 10) {
  if (!teamName) return '';

  if (teamName.length <= maxLength) {
    return teamName;
  }

  return teamName.substring(0, maxLength - 3) + '...';
}

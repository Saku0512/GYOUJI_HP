<script>
  // TournamentBracket コンポーネント - ブラケット形式でのトーナメント表示
  import MatchCard from './MatchCard.svelte';
  
  export let sport = 'volleyball';
  export let matches = [];
  export let isAdmin = false;
  export let onEditMatch = null; // 管理者向け編集コールバック

  // スポーツ名の日本語表示マッピング
  const sportNames = {
    volleyball: 'バレーボール',
    table_tennis: '卓球',
    soccer: 'サッカー'
  };

  // ラウンド名の日本語表示マッピング
  const roundNames = {
    'round_1': '1回戦',
    'round_2': '2回戦',
    'quarterfinal': '準々決勝',
    'semifinal': '準決勝',
    'final': '決勝',
    'third_place': '3位決定戦'
  };

  // 試合をラウンド別にグループ化
  $: groupedMatches = groupMatchesByRound(matches);
  
  // ラウンドの順序を定義
  $: roundOrder = getRoundOrder(groupedMatches);

  /**
   * 試合をラウンド別にグループ化する関数
   */
  function groupMatchesByRound(matches) {
    if (!Array.isArray(matches)) {
      return {};
    }

    return matches.reduce((groups, match) => {
      if (!match || typeof match !== 'object') {
        return groups;
      }
      const round = match.round || 'unknown';
      if (!groups[round]) {
        groups[round] = [];
      }
      groups[round].push(match);
      return groups;
    }, {});
  }

  /**
   * ラウンドの表示順序を取得する関数
   */
  function getRoundOrder(groupedMatches) {
    const availableRounds = Object.keys(groupedMatches);
    const standardOrder = ['round_1', 'round_2', 'quarterfinal', 'semifinal', 'third_place', 'final'];
    
    // 標準順序に従ってソート
    return standardOrder.filter(round => availableRounds.includes(round))
      .concat(availableRounds.filter(round => !standardOrder.includes(round)));
  }

  /**
   * 管理者向け編集ボタンのクリックハンドラ
   */
  function handleEditMatch(match) {
    if (isAdmin && onEditMatch && typeof onEditMatch === 'function') {
      onEditMatch(match);
    }
  }

  /**
   * 試合の勝者を判定する関数
   */
  function getMatchWinner(match) {
    if (match.winner) {
      return match.winner;
    }
    
    if (match.score1 !== undefined && match.score2 !== undefined) {
      if (match.score1 > match.score2) {
        return match.team1;
      } else if (match.score2 > match.score1) {
        return match.team2;
      }
    }
    
    return null;
  }

  /**
   * 試合の状態を取得する関数
   */
  function getMatchStatus(match) {
    if (match.status === 'completed' || (match.score1 !== undefined && match.score2 !== undefined)) {
      return 'completed';
    } else if (match.status === 'in_progress') {
      return 'in_progress';
    } else {
      return 'pending';
    }
  }

  /**
   * レスポンシブ表示用のクラス名を取得
   */
  function getResponsiveClass() {
    if (typeof window !== 'undefined') {
      const width = window.innerWidth;
      if (width < 768) return 'mobile';
      if (width < 1024) return 'tablet';
      return 'desktop';
    }
    return 'desktop';
  }

  let responsiveClass = 'desktop';
  
  // ウィンドウリサイズイベントの監視
  if (typeof window !== 'undefined') {
    responsiveClass = getResponsiveClass();
    
    window.addEventListener('resize', () => {
      responsiveClass = getResponsiveClass();
    });
  }
</script>

<div class="tournament-bracket {responsiveClass}">
  <div class="bracket-header">
    <h2 class="tournament-title">
      {sportNames[sport] || sport} トーナメント
    </h2>
    {#if matches.length === 0}
      <p class="no-matches">試合データがありません</p>
    {/if}
  </div>

  {#if matches.length > 0}
    <div class="bracket-container">
      {#each roundOrder as round}
        <div class="round-column" data-round={round}>
          <h3 class="round-title">
            {roundNames[round] || round}
          </h3>
          <div class="matches-container">
            {#each groupedMatches[round] as match (match.id)}
              <div class="match-wrapper" data-status={getMatchStatus(match)}>
                <div class="match-card-container">
                  <MatchCard 
                    {match} 
                    editable={isAdmin}
                  />
                  
                  {#if getMatchWinner(match)}
                    <div class="winner-indicator">
                      勝者: {getMatchWinner(match)}
                    </div>
                  {/if}
                  
                  {#if isAdmin && getMatchStatus(match) === 'pending'}
                    <button 
                      class="edit-button"
                      on:click={() => handleEditMatch(match)}
                      aria-label="試合結果を編集"
                    >
                      結果入力
                    </button>
                  {/if}
                </div>
                
                <!-- 次のラウンドへの接続線 -->
                {#if round !== 'final' && round !== 'third_place'}
                  <div class="connection-line" aria-hidden="true"></div>
                {/if}
              </div>
            {/each}
          </div>
        </div>
      {/each}
    </div>
  {/if}
</div>

<style>
  .tournament-bracket {
    width: 100%;
    padding: 1rem;
    background-color: #f8f9fa;
    border-radius: 12px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  .bracket-header {
    text-align: center;
    margin-bottom: 2rem;
  }

  .tournament-title {
    font-size: 2rem;
    font-weight: bold;
    color: #2c3e50;
    margin: 0 0 1rem 0;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .no-matches {
    color: #6c757d;
    font-style: italic;
    font-size: 1.1rem;
  }

  .bracket-container {
    display: flex;
    justify-content: center;
    align-items: flex-start;
    gap: 2rem;
    overflow-x: auto;
    padding: 1rem;
    min-height: 400px;
  }

  .round-column {
    display: flex;
    flex-direction: column;
    align-items: center;
    min-width: 250px;
    position: relative;
  }

  .round-title {
    font-size: 1.2rem;
    font-weight: bold;
    color: #495057;
    margin: 0 0 1.5rem 0;
    padding: 0.5rem 1rem;
    background-color: #e9ecef;
    border-radius: 20px;
    border: 2px solid #dee2e6;
  }

  .matches-container {
    display: flex;
    flex-direction: column;
    gap: 2rem;
    width: 100%;
  }

  .match-wrapper {
    position: relative;
    display: flex;
    flex-direction: column;
    align-items: center;
  }

  .match-card-container {
    position: relative;
    width: 100%;
    background-color: white;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    transition: transform 0.2s ease, box-shadow 0.2s ease;
  }

  .match-card-container:hover {
    transform: translateY(-2px);
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .match-wrapper[data-status="completed"] .match-card-container {
    border-left: 4px solid #28a745;
  }

  .match-wrapper[data-status="in_progress"] .match-card-container {
    border-left: 4px solid #ffc107;
  }

  .match-wrapper[data-status="pending"] .match-card-container {
    border-left: 4px solid #6c757d;
  }

  .winner-indicator {
    background-color: #28a745;
    color: white;
    padding: 0.25rem 0.5rem;
    font-size: 0.8rem;
    font-weight: bold;
    text-align: center;
    border-radius: 0 0 8px 8px;
  }

  .edit-button {
    position: absolute;
    top: -10px;
    right: -10px;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 50%;
    width: 40px;
    height: 40px;
    font-size: 0.7rem;
    font-weight: bold;
    cursor: pointer;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.2);
    transition: all 0.2s ease;
    z-index: 10;
  }

  .edit-button:hover {
    background-color: #0056b3;
    transform: scale(1.1);
  }

  .edit-button:focus {
    outline: 2px solid #80bdff;
    outline-offset: 2px;
  }

  .connection-line {
    position: absolute;
    right: -2rem;
    top: 50%;
    width: 2rem;
    height: 2px;
    background-color: #dee2e6;
    transform: translateY(-50%);
  }

  .connection-line::after {
    content: '';
    position: absolute;
    right: -6px;
    top: -4px;
    width: 0;
    height: 0;
    border-left: 6px solid #dee2e6;
    border-top: 5px solid transparent;
    border-bottom: 5px solid transparent;
  }

  /* レスポンシブデザイン */
  .tournament-bracket.mobile {
    padding: 0.5rem;
  }

  .tournament-bracket.mobile .tournament-title {
    font-size: 1.5rem;
  }

  .tournament-bracket.mobile .bracket-container {
    flex-direction: column;
    gap: 1rem;
    align-items: center;
  }

  .tournament-bracket.mobile .round-column {
    min-width: 100%;
    max-width: 300px;
  }

  .tournament-bracket.mobile .connection-line {
    display: none;
  }

  .tournament-bracket.mobile .edit-button {
    position: static;
    margin-top: 0.5rem;
    border-radius: 4px;
    width: auto;
    height: auto;
    padding: 0.5rem 1rem;
  }

  .tournament-bracket.tablet .bracket-container {
    gap: 1.5rem;
  }

  .tournament-bracket.tablet .round-column {
    min-width: 200px;
  }

  .tournament-bracket.tablet .tournament-title {
    font-size: 1.8rem;
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .match-card-container {
      transition: none;
    }
    
    .edit-button {
      transition: none;
    }
    
    .match-card-container:hover {
      transform: none;
    }
    
    .edit-button:hover {
      transform: none;
    }
  }

  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .tournament-bracket {
      border: 2px solid #000;
    }
    
    .match-card-container {
      border: 1px solid #000;
    }
    
    .edit-button {
      border: 2px solid #fff;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .tournament-bracket {
      background-color: #2d3748;
      color: #e2e8f0;
    }
    
    .tournament-title {
      color: #f7fafc;
    }
    
    .round-title {
      background-color: #4a5568;
      color: #e2e8f0;
      border-color: #718096;
    }
    
    .match-card-container {
      background-color: #4a5568;
    }
    
    .connection-line {
      background-color: #718096;
    }
    
    .connection-line::after {
      border-left-color: #718096;
    }
  }

  /* 印刷スタイル */
  @media print {
    .tournament-bracket {
      background-color: white;
      box-shadow: none;
      padding: 0;
    }
    
    .edit-button {
      display: none;
    }
    
    .bracket-container {
      overflow: visible;
    }
  }
</style>

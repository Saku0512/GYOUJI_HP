<script>
  import { createEventDispatcher } from 'svelte';
  import { formatMatchStatus, formatScore, formatTeamName } from '../utils/formatting.js';
  
  // MatchCard コンポーネント - 個別試合情報表示
  export let match = {};
  export let editable = false;
  export let compact = false;
  
  const dispatch = createEventDispatcher();
  
  let isEditMode = false;
  let editScore1 = match.score1 || '';
  let editScore2 = match.score2 || '';
  
  // 試合の勝者を判定
  $: winner = getWinner(match);
  
  // クラス名の計算
  $: cardClasses = [
    'match-card',
    compact ? 'compact' : '',
    editable ? 'editable' : '',
    isEditMode ? 'edit-mode' : ''
  ].filter(Boolean).join(' ');
  
  $: team1Classes = ['team', winner === match.team1 ? 'winner' : ''].filter(Boolean).join(' ');
  $: team2Classes = ['team', winner === match.team2 ? 'winner' : ''].filter(Boolean).join(' ');
  
  // 編集モードの切り替え
  function toggleEditMode() {
    if (!editable) return;
    
    isEditMode = !isEditMode;
    if (isEditMode) {
      editScore1 = match.score1 || '';
      editScore2 = match.score2 || '';
    }
  }
  
  // スコア更新の送信
  function handleScoreSubmit() {
    const score1 = parseInt(editScore1);
    const score2 = parseInt(editScore2);
    
    if (isNaN(score1) || isNaN(score2) || score1 < 0 || score2 < 0) {
      dispatch('error', { message: 'スコアは0以上の数値を入力してください' });
      return;
    }
    
    dispatch('updateScore', {
      matchId: match.id,
      score1,
      score2
    });
    
    isEditMode = false;
  }
  
  // 編集キャンセル
  function handleCancel() {
    isEditMode = false;
    editScore1 = match.score1 || '';
    editScore2 = match.score2 || '';
  }
  
  // 勝者判定ロジック
  function getWinner(match) {
    if (match.winner) return match.winner;
    if (match.score1 !== null && match.score2 !== null && match.score1 !== undefined && match.score2 !== undefined) {
      if (match.score1 > match.score2) return match.team1;
      if (match.score2 > match.score1) return match.team2;
      return 'draw';
    }
    return null;
  }
  
  // キーボードイベントハンドラー
  function handleKeydown(event) {
    if (event.key === 'Enter') {
      handleScoreSubmit();
    } else if (event.key === 'Escape') {
      handleCancel();
    }
  }
</script>

<div 
  class={cardClasses}
  data-testid="match-card"
>
  <!-- 試合情報ヘッダー -->
  <div class="match-header">
    {#if match.round}
      <div class="round" data-testid="match-round">{match.round}</div>
    {/if}
    {#if match.scheduled_at}
      <div class="schedule" data-testid="match-schedule">
        {new Date(match.scheduled_at).toLocaleDateString('ja-JP', { 
          month: 'short', 
          day: 'numeric', 
          hour: '2-digit', 
          minute: '2-digit' 
        })}
      </div>
    {/if}
  </div>

  <!-- チーム表示 -->
  <div class="teams">
    <div class={team1Classes} data-testid="team1">
      <span class="team-name">{formatTeamName(match.team1 || 'Team A', compact ? 8 : 15)}</span>
      {#if winner === match.team1}
        <span class="winner-badge">🏆</span>
      {/if}
    </div>
    
    <div class="vs-section">
      <span class="vs">vs</span>
    </div>
    
    <div class={team2Classes} data-testid="team2">
      {#if winner === match.team2}
        <span class="winner-badge">🏆</span>
      {/if}
      <span class="team-name">{formatTeamName(match.team2 || 'Team B', compact ? 8 : 15)}</span>
    </div>
  </div>

  <!-- スコア表示/編集 -->
  <div class="score-section">
    {#if isEditMode}
      <div class="score-edit" data-testid="score-edit">
        <input 
          type="number" 
          bind:value={editScore1}
          min="0"
          class="score-input"
          data-testid="score1-input"
          on:keydown={handleKeydown}
          placeholder="0"
        />
        <span class="score-separator">-</span>
        <input 
          type="number" 
          bind:value={editScore2}
          min="0"
          class="score-input"
          data-testid="score2-input"
          on:keydown={handleKeydown}
          placeholder="0"
        />
      </div>
      <div class="edit-actions">
        <button 
          class="btn btn-primary btn-sm"
          on:click={handleScoreSubmit}
          data-testid="save-score-btn"
        >
          保存
        </button>
        <button 
          class="btn btn-secondary btn-sm"
          on:click={handleCancel}
          data-testid="cancel-edit-btn"
        >
          キャンセル
        </button>
      </div>
    {:else}
      <div class="score-display" data-testid="score-display">
        {#if match.score1 !== null && match.score1 !== undefined && match.score2 !== null && match.score2 !== undefined}
          <div class="score">{formatScore(match.score1, match.score2)}</div>
        {:else}
          <div class="status" data-testid="match-status">
            {formatMatchStatus(match.status || 'pending')}
          </div>
        {/if}
      </div>
      
      {#if editable && match.status !== 'completed'}
        <button 
          class="edit-btn"
          on:click={toggleEditMode}
          data-testid="edit-match-btn"
          aria-label="試合結果を編集"
        >
          ✏️ 編集
        </button>
      {/if}
    {/if}
  </div>

  <!-- 試合詳細情報 -->
  {#if !compact && (match.completed_at || match.status === 'completed')}
    <div class="match-details">
      <div class="completion-time" data-testid="completion-time">
        完了: {new Date(match.completed_at).toLocaleDateString('ja-JP', { 
          month: 'short', 
          day: 'numeric', 
          hour: '2-digit', 
          minute: '2-digit' 
        })}
      </div>
    </div>
  {/if}
</div>

<style>
  .match-card {
    border: 1px solid #ddd;
    border-radius: 12px;
    padding: 1rem;
    margin: 0.5rem;
    background-color: white;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    transition: all 0.2s ease;
    position: relative;
    min-height: 120px;
  }

  .match-card:hover {
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.15);
  }

  .match-card.compact {
    padding: 0.75rem;
    margin: 0.25rem;
    min-height: 80px;
  }

  .match-card.editable {
    cursor: pointer;
  }

  .match-card.edit-mode {
    border-color: #007bff;
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }

  .match-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
    font-size: 0.8rem;
    color: #666;
  }

  .round {
    font-weight: 600;
    color: #495057;
  }

  .schedule {
    font-size: 0.75rem;
  }

  .teams {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    gap: 0.5rem;
  }

  .team {
    flex: 1;
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.5rem;
    border-radius: 8px;
    transition: background-color 0.2s ease;
  }

  .team.winner {
    background-color: #d4edda;
    border: 1px solid #c3e6cb;
  }

  .team:first-child {
    justify-content: flex-start;
  }

  .team:last-child {
    justify-content: flex-end;
    flex-direction: row-reverse;
  }

  .team-name {
    font-weight: 600;
    color: #333;
    font-size: 0.9rem;
  }

  .winner-badge {
    font-size: 1rem;
    animation: bounce 0.5s ease-in-out;
  }

  @keyframes bounce {
    0%, 20%, 50%, 80%, 100% { transform: translateY(0); }
    40% { transform: translateY(-3px); }
    60% { transform: translateY(-2px); }
  }

  .vs-section {
    display: flex;
    align-items: center;
    justify-content: center;
    min-width: 40px;
  }

  .vs {
    color: #666;
    font-size: 0.8rem;
    font-weight: 500;
  }

  .score-section {
    text-align: center;
    margin-bottom: 0.5rem;
  }

  .score-display {
    margin-bottom: 0.5rem;
  }

  .score {
    font-size: 1.4rem;
    font-weight: bold;
    color: #007bff;
    margin-bottom: 0.25rem;
  }

  .status {
    color: #6c757d;
    font-style: italic;
    font-size: 0.9rem;
  }

  .score-edit {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
  }

  .score-input {
    width: 60px;
    padding: 0.5rem;
    border: 2px solid #ced4da;
    border-radius: 6px;
    text-align: center;
    font-size: 1.1rem;
    font-weight: 600;
    transition: border-color 0.2s ease;
  }

  .score-input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }

  .score-separator {
    font-size: 1.2rem;
    font-weight: bold;
    color: #495057;
  }

  .edit-actions {
    display: flex;
    justify-content: center;
    gap: 0.5rem;
  }

  .btn {
    padding: 0.5rem 1rem;
    border: none;
    border-radius: 6px;
    font-size: 0.85rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    min-width: 44px;
    min-height: 44px;
  }

  .btn-primary {
    background-color: #007bff;
    color: white;
  }

  .btn-primary:hover {
    background-color: #0056b3;
  }

  .btn-secondary {
    background-color: #6c757d;
    color: white;
  }

  .btn-secondary:hover {
    background-color: #545b62;
  }

  .btn-sm {
    padding: 0.375rem 0.75rem;
    font-size: 0.8rem;
    min-height: 36px;
  }

  .edit-btn {
    background: #f8f9fa;
    border: 1px solid #dee2e6;
    border-radius: 6px;
    padding: 0.5rem 0.75rem;
    font-size: 0.8rem;
    color: #495057;
    cursor: pointer;
    transition: all 0.2s ease;
    min-width: 44px;
    min-height: 44px;
  }

  .edit-btn:hover {
    background-color: #e9ecef;
    border-color: #adb5bd;
  }

  .match-details {
    border-top: 1px solid #e9ecef;
    padding-top: 0.5rem;
    margin-top: 0.5rem;
  }

  .completion-time {
    font-size: 0.75rem;
    color: #6c757d;
    text-align: center;
  }

  /* レスポンシブデザイン */
  @media (max-width: 768px) {
    .match-card {
      padding: 0.75rem;
      margin: 0.25rem 0;
    }

    .teams {
      flex-direction: column;
      gap: 0.75rem;
    }

    .team {
      justify-content: center !important;
      flex-direction: row !important;
    }

    .vs-section {
      order: -1;
      margin: 0.25rem 0;
    }

    .score {
      font-size: 1.2rem;
    }

    .score-input {
      width: 50px;
      font-size: 1rem;
    }

    .btn {
      min-width: 48px;
      min-height: 48px;
      font-size: 0.9rem;
    }

    .edit-btn {
      min-width: 48px;
      min-height: 48px;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .match-card {
      background-color: #2d3748;
      border-color: #4a5568;
      color: #e2e8f0;
    }

    .team-name {
      color: #e2e8f0;
    }

    .team.winner {
      background-color: #2f855a;
      border-color: #38a169;
    }

    .score {
      color: #63b3ed;
    }

    .status {
      color: #a0aec0;
    }

    .edit-btn {
      background-color: #4a5568;
      border-color: #2d3748;
      color: #e2e8f0;
    }

    .edit-btn:hover {
      background-color: #2d3748;
    }
  }

  /* アクセシビリティ */
  @media (prefers-reduced-motion: reduce) {
    .match-card,
    .btn,
    .edit-btn,
    .score-input,
    .team {
      transition: none;
    }

    .winner-badge {
      animation: none;
    }
  }

  /* 高コントラストモード */
  @media (prefers-contrast: high) {
    .match-card {
      border-width: 2px;
    }

    .team.winner {
      border-width: 2px;
    }

    .btn {
      border: 2px solid currentColor;
    }
  }
</style>
<script>
  // MatchCard コンポーネント - 個別試合情報表示
  export let match = {};
  export let editable = false;
</script>

<div class="match-card">
  <div class="teams">
    <span class="team">{match.team1 || 'Team A'}</span>
    <span class="vs">vs</span>
    <span class="team">{match.team2 || 'Team B'}</span>
  </div>
  {#if match.score1 !== undefined && match.score2 !== undefined}
    <div class="score">
      {match.score1} - {match.score2}
    </div>
  {:else}
    <div class="status">未実施</div>
  {/if}
</div>

<style>
  .match-card {
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 1rem;
    margin: 0.5rem;
    background-color: white;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .teams {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .team {
    font-weight: bold;
    color: #333;
  }

  .vs {
    color: #666;
    font-size: 0.9rem;
  }

  .score {
    text-align: center;
    font-size: 1.2rem;
    font-weight: bold;
    color: #007bff;
  }

  .status {
    text-align: center;
    color: #666;
    font-style: italic;
  }
</style>

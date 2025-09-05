<script>
  // AdminMatchForm コンポーネント - 試合結果入力フォーム
  export let match = {};
  export let onSubmit = () => {};

  let score1 = match.score1 || '';
  let score2 = match.score2 || '';
</script>

<form class="admin-match-form" on:submit|preventDefault={() => onSubmit({ score1, score2 })}>
  <h3>試合結果入力</h3>
  <div class="match-info">
    <span class="team">{match.team1 || 'Team A'}</span>
    <span class="vs">vs</span>
    <span class="team">{match.team2 || 'Team B'}</span>
  </div>

  <div class="score-inputs">
    <div class="score-group">
      <label for="score1">{match.team1 || 'Team A'} スコア</label>
      <input type="number" id="score1" bind:value={score1} min="0" required />
    </div>
    <div class="score-group">
      <label for="score2">{match.team2 || 'Team B'} スコア</label>
      <input type="number" id="score2" bind:value={score2} min="0" required />
    </div>
  </div>

  <button type="submit">結果を保存</button>
</form>

<style>
  .admin-match-form {
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 1.5rem;
    background-color: white;
    max-width: 400px;
    margin: 0 auto;
  }

  h3 {
    text-align: center;
    margin-bottom: 1rem;
    color: #333;
  }

  .match-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1.5rem;
    padding: 0.5rem;
    background-color: #f8f9fa;
    border-radius: 4px;
  }

  .team {
    font-weight: bold;
    color: #333;
  }

  .vs {
    color: #666;
  }

  .score-inputs {
    display: flex;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .score-group {
    flex: 1;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    color: #555;
    font-size: 0.9rem;
  }

  input {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
  }

  button {
    width: 100%;
    padding: 0.75rem;
    background-color: #28a745;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
  }

  button:hover {
    background-color: #218838;
  }
</style>

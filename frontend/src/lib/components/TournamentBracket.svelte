<script>
  // TournamentBracket コンポーネント - ブラケット形式でのトーナメント表示
  export let sport = 'volleyball';
  export let matches = [];
  export let isAdmin = false;
</script>

<div class="tournament-bracket">
  <h2>{sport} トーナメント</h2>
  <div class="bracket-container">
    <!-- ブラケット表示ロジックは後のタスクで実装 -->
    <p>Tournament bracket will be implemented in later tasks</p>
  </div>
</div>

<style>
  .tournament-bracket {
    width: 100%;
    padding: 1rem;
  }

  h2 {
    text-align: center;
    margin-bottom: 1rem;
    color: #333;
  }

  .bracket-container {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 200px;
    border: 1px solid #ddd;
    border-radius: 8px;
    background-color: #f9f9f9;
  }
</style>

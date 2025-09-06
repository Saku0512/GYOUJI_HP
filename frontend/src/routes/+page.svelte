<script>
  import { onMount, onDestroy } from 'svelte';
  import { tournamentStore } from '$lib/stores/tournament.js';
  import { uiStore, uiActions } from '$lib/stores/ui.js';
  import TournamentBracket from '$lib/components/TournamentBracket.svelte';
  import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';

  // ストアの状態を購読
  $: tournament = $tournamentStore;
  $: ui = $uiStore;

  // スポーツタブの定義
  const sports = [
    { key: 'volleyball', name: 'バレーボール', icon: '🏐' },
    { key: 'table_tennis', name: '卓球', icon: '🏓' },
    { key: 'soccer', name: 'サッカー', icon: '⚽' }
  ];

  // 現在選択されているスポーツのトーナメントデータ
  $: currentTournamentData = tournament.tournaments[tournament.currentSport] || null;
  
  // 現在のスポーツの試合データ
  $: matches = currentTournamentData?.matches || [];

  // リアルタイム更新のポーリング間隔ID
  let pollingInterval = null;

  /**
   * コンポーネントのマウント時の処理
   */
  onMount(async () => {
    try {
      // トーナメントストアの初期化
      await tournamentStore.initialize();
      
      // 初期データの取得
      await loadTournamentData(tournament.currentSport);
      
      // リアルタイム更新の開始
      startRealtimeUpdates();
      
    } catch (error) {
      console.error('Homepage initialization error:', error);
      uiActions.showNotification('データの読み込みに失敗しました', 'error');
    }
  });

  /**
   * コンポーネントのアンマウント時の処理
   */
  onDestroy(() => {
    stopRealtimeUpdates();
  });

  /**
   * スポーツタブの切り替え処理
   */
  async function handleSportChange(sportKey) {
    try {
      if (sportKey === tournament.currentSport) {
        return; // 同じスポーツが選択された場合は何もしない
      }

      // スポーツを切り替え
      const result = tournamentStore.switchSport(sportKey);
      
      if (result.success) {
        // 新しいスポーツのデータを読み込み
        await loadTournamentData(sportKey);
        
        // 成功通知（オプション）
        // uiActions.showNotification(`${getSportName(sportKey)}に切り替えました`, 'success');
      } else {
        uiActions.showNotification(result.message || 'スポーツの切り替えに失敗しました', 'error');
      }
    } catch (error) {
      console.error('Sport change error:', error);
      uiActions.showNotification('スポーツの切り替えでエラーが発生しました', 'error');
    }
  }

  /**
   * トーナメントデータの読み込み
   */
  async function loadTournamentData(sport, showLoading = true) {
    try {
      if (showLoading) {
        uiActions.setLoading(true);
      }

      const result = await tournamentStore.fetchTournaments(sport);
      
      if (!result.success) {
        uiActions.showNotification(result.message || 'データの取得に失敗しました', 'error');
      }
      
      return result;
    } catch (error) {
      console.error('Load tournament data error:', error);
      uiActions.showNotification('データの読み込みでエラーが発生しました', 'error');
      return { success: false, message: error.message };
    } finally {
      if (showLoading) {
        uiActions.setLoading(false);
      }
    }
  }

  /**
   * データの手動更新
   */
  async function handleRefresh() {
    try {
      uiActions.setLoading(true);
      
      const result = await tournamentStore.refreshData(tournament.currentSport);
      
      if (result.success) {
        uiActions.showNotification('データを更新しました', 'success');
      } else {
        uiActions.showNotification(result.message || 'データの更新に失敗しました', 'error');
      }
    } catch (error) {
      console.error('Refresh error:', error);
      uiActions.showNotification('データの更新でエラーが発生しました', 'error');
    } finally {
      uiActions.setLoading(false);
    }
  }

  /**
   * リアルタイム更新の開始
   */
  function startRealtimeUpdates() {
    // 既存のポーリングがある場合は停止
    stopRealtimeUpdates();
    
    // 30秒ごとにデータを更新
    pollingInterval = setInterval(async () => {
      // ページが非表示の場合はスキップ
      if (typeof document !== 'undefined' && document.hidden) {
        return;
      }
      
      // ローディング中の場合はスキップ
      if (tournament.loading || ui.loading) {
        return;
      }
      
      try {
        // サイレントでデータを更新（ローディング表示なし）
        await loadTournamentData(tournament.currentSport, false);
      } catch (error) {
        console.error('Polling update error:', error);
      }
    }, 30000); // 30秒間隔
  }

  /**
   * リアルタイム更新の停止
   */
  function stopRealtimeUpdates() {
    if (pollingInterval) {
      clearInterval(pollingInterval);
      pollingInterval = null;
    }
  }

  /**
   * スポーツ名を取得
   */
  function getSportName(sportKey) {
    const sport = sports.find(s => s.key === sportKey);
    return sport ? sport.name : sportKey;
  }

  /**
   * ページの可視性変更時の処理
   */
  function handleVisibilityChange() {
    if (typeof document !== 'undefined') {
      if (!document.hidden) {
        // ページが表示されたときにデータを更新
        loadTournamentData(tournament.currentSport, false);
      }
    }
  }

  // ページの可視性変更イベントの監視
  if (typeof document !== 'undefined') {
    document.addEventListener('visibilitychange', handleVisibilityChange);
  }
</script>

<svelte:head>
  <title>トーナメント管理システム - {getSportName(tournament.currentSport)}</title>
  <meta name="description" content="バレーボール、卓球、サッカーのトーナメント結果をリアルタイムで確認できます" />
</svelte:head>

<div class="homepage">
  <div class="container">
    <!-- ページヘッダー -->
    <div class="page-header">
      <h1 class="page-title">トーナメント管理システム</h1>
      <p class="page-description">
        リアルタイムでトーナメントの進行状況を確認できます
      </p>
      
      <!-- 更新ボタン -->
      <div class="header-actions">
        <button 
          class="refresh-button"
          on:click={handleRefresh}
          disabled={tournament.loading || ui.loading}
          aria-label="データを更新"
        >
          <span class="refresh-icon" class:spinning={tournament.loading || ui.loading}>🔄</span>
          更新
        </button>
        
        {#if tournament.lastUpdated}
          <span class="last-updated">
            最終更新: {new Date(tournament.lastUpdated).toLocaleTimeString('ja-JP')}
          </span>
        {/if}
      </div>
    </div>

    <!-- スポーツタブ -->
    <div class="sports-tabs">
      <div class="tabs-container">
        {#each sports as sport}
          <button
            class="sport-tab"
            class:active={tournament.currentSport === sport.key}
            on:click={() => handleSportChange(sport.key)}
            disabled={tournament.loading}
            aria-label="{sport.name}のトーナメントを表示"
          >
            <span class="sport-icon">{sport.icon}</span>
            <span class="sport-name">{sport.name}</span>
          </button>
        {/each}
      </div>
    </div>

    <!-- メインコンテンツ -->
    <div class="main-content">
      {#if tournament.error}
        <!-- エラー表示 -->
        <div class="error-container">
          <div class="error-message">
            <h3>エラーが発生しました</h3>
            <p>{tournament.error}</p>
            <button class="retry-button" on:click={handleRefresh}>
              再試行
            </button>
          </div>
        </div>
      {:else if tournament.loading && !currentTournamentData}
        <!-- 初回ローディング -->
        <div class="loading-container">
          <LoadingSpinner size="large" />
          <p class="loading-text">トーナメントデータを読み込み中...</p>
        </div>
      {:else if currentTournamentData && matches.length > 0}
        <!-- トーナメントブラケット表示 -->
        <div class="tournament-container">
          <TournamentBracket 
            sport={tournament.currentSport}
            {matches}
            isAdmin={false}
          />
        </div>
      {:else}
        <!-- データなし表示 -->
        <div class="no-data-container">
          <div class="no-data-message">
            <h3>トーナメントデータがありません</h3>
            <p>{getSportName(tournament.currentSport)}のトーナメントはまだ開始されていません。</p>
            <button class="refresh-button" on:click={handleRefresh}>
              データを確認
            </button>
          </div>
        </div>
      {/if}
    </div>
  </div>
</div>

<style>
  .homepage {
    min-height: calc(100vh - 200px);
    background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  }

  .container {
    max-width: 1400px;
    margin: 0 auto;
    padding: 2rem;
  }

  /* ページヘッダー */
  .page-header {
    text-align: center;
    margin-bottom: 3rem;
  }

  .page-title {
    font-size: 2.5rem;
    font-weight: 700;
    color: #2c3e50;
    margin: 0 0 1rem 0;
    text-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .page-description {
    font-size: 1.1rem;
    color: #6c757d;
    margin: 0 0 2rem 0;
    max-width: 600px;
    margin-left: auto;
    margin-right: auto;
  }

  .header-actions {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .refresh-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    background-color: #007bff;
    color: white;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 8px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    box-shadow: 0 2px 4px rgba(0, 123, 255, 0.3);
  }

  .refresh-button:hover:not(:disabled) {
    background-color: #0056b3;
    transform: translateY(-1px);
    box-shadow: 0 4px 8px rgba(0, 123, 255, 0.4);
  }

  .refresh-button:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
  }

  .refresh-icon {
    font-size: 1rem;
    transition: transform 0.5s ease;
  }

  .refresh-icon.spinning {
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    from { transform: rotate(0deg); }
    to { transform: rotate(360deg); }
  }

  .last-updated {
    font-size: 0.9rem;
    color: #6c757d;
    font-style: italic;
  }

  /* スポーツタブ */
  .sports-tabs {
    margin-bottom: 3rem;
  }

  .tabs-container {
    display: flex;
    justify-content: center;
    gap: 0.5rem;
    background-color: #f8f9fa;
    padding: 0.5rem;
    border-radius: 12px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    max-width: 600px;
    margin: 0 auto;
  }

  .sport-tab {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    background: none;
    border: none;
    padding: 1rem 1.5rem;
    border-radius: 8px;
    cursor: pointer;
    transition: all 0.2s ease;
    flex: 1;
    min-width: 120px;
  }

  .sport-tab:hover:not(:disabled) {
    background-color: #e9ecef;
    transform: translateY(-2px);
  }

  .sport-tab.active {
    background-color: #007bff;
    color: white;
    box-shadow: 0 4px 8px rgba(0, 123, 255, 0.3);
  }

  .sport-tab:disabled {
    opacity: 0.6;
    cursor: not-allowed;
    transform: none;
  }

  .sport-icon {
    font-size: 2rem;
    margin-bottom: 0.25rem;
  }

  .sport-name {
    font-size: 0.9rem;
    font-weight: 600;
    text-align: center;
  }

  /* メインコンテンツ */
  .main-content {
    min-height: 400px;
  }

  .tournament-container {
    background-color: white;
    border-radius: 12px;
    padding: 2rem;
    box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  }

  /* ローディング表示 */
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    padding: 4rem 2rem;
    text-align: center;
  }

  .loading-text {
    margin-top: 1rem;
    color: #6c757d;
    font-size: 1.1rem;
  }

  /* エラー表示 */
  .error-container {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 4rem 2rem;
  }

  .error-message {
    background-color: #f8d7da;
    color: #721c24;
    padding: 2rem;
    border-radius: 8px;
    border: 1px solid #f5c6cb;
    text-align: center;
    max-width: 500px;
  }

  .error-message h3 {
    margin: 0 0 1rem 0;
    color: #721c24;
  }

  .error-message p {
    margin: 0 0 1.5rem 0;
  }

  .retry-button {
    background-color: #dc3545;
    color: white;
    border: none;
    padding: 0.75rem 1.5rem;
    border-radius: 4px;
    cursor: pointer;
    font-weight: 500;
    transition: background-color 0.2s ease;
  }

  .retry-button:hover {
    background-color: #c82333;
  }

  /* データなし表示 */
  .no-data-container {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 4rem 2rem;
  }

  .no-data-message {
    background-color: #d1ecf1;
    color: #0c5460;
    padding: 2rem;
    border-radius: 8px;
    border: 1px solid #bee5eb;
    text-align: center;
    max-width: 500px;
  }

  .no-data-message h3 {
    margin: 0 0 1rem 0;
    color: #0c5460;
  }

  .no-data-message p {
    margin: 0 0 1.5rem 0;
  }

  /* レスポンシブデザイン */
  @media (max-width: 768px) {
    .container {
      padding: 1rem;
    }

    .page-title {
      font-size: 2rem;
    }

    .page-description {
      font-size: 1rem;
    }

    .header-actions {
      flex-direction: column;
      gap: 0.5rem;
    }

    .tabs-container {
      flex-direction: column;
      gap: 0.25rem;
    }

    .sport-tab {
      flex-direction: row;
      justify-content: center;
      padding: 0.75rem 1rem;
    }

    .sport-icon {
      font-size: 1.5rem;
      margin-bottom: 0;
      margin-right: 0.5rem;
    }

    .tournament-container {
      padding: 1rem;
    }

    .loading-container,
    .error-container,
    .no-data-container {
      padding: 2rem 1rem;
    }
  }

  @media (max-width: 480px) {
    .page-title {
      font-size: 1.75rem;
    }

    .sport-name {
      font-size: 0.8rem;
    }

    .refresh-button {
      padding: 0.5rem 1rem;
      font-size: 0.9rem;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .refresh-button,
    .sport-tab,
    .refresh-icon {
      transition: none;
    }
    
    .refresh-button:hover,
    .sport-tab:hover {
      transform: none;
    }
    
    .refresh-icon.spinning {
      animation: none;
    }
  }

  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .sport-tab {
      border: 2px solid #000;
    }
    
    .sport-tab.active {
      border-color: #fff;
    }
    
    .tournament-container {
      border: 2px solid #000;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .homepage {
      background: linear-gradient(135deg, #2d3748 0%, #4a5568 100%);
    }
    
    .page-title {
      color: #f7fafc;
    }
    
    .page-description {
      color: #a0aec0;
    }
    
    .tabs-container {
      background-color: #4a5568;
    }
    
    .sport-tab:hover:not(:disabled) {
      background-color: #718096;
    }
    
    .tournament-container {
      background-color: #2d3748;
      color: #e2e8f0;
    }
    
    .last-updated {
      color: #a0aec0;
    }
  }
</style>

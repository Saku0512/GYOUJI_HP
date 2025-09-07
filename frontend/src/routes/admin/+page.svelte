<script>
  // 管理者ダッシュボード - 管理者専用ダッシュボード
  import { onMount, onDestroy } from 'svelte';
  import { authStore } from '../../lib/stores/auth.js';
  import { tournamentStore } from '../../lib/stores/tournament.js';
  import { uiActions } from '../../lib/stores/ui.js';
  import { matchAPI } from '../../lib/api/matches.js';
  import { tournamentAPI } from '../../lib/api/tournament.js';
  import AdminMatchForm from '../../lib/components/AdminMatchForm.svelte';
  import LoadingSpinner from '../../lib/components/LoadingSpinner.svelte';
  import Button from '../../lib/components/Button.svelte';
  import Select from '../../lib/components/Select.svelte';
  import ResponsiveLayout from '../../lib/components/ResponsiveLayout.svelte';
  import ResponsiveGrid from '../../lib/components/ResponsiveGrid.svelte';
  import AnimatedTransition from '../../lib/components/AnimatedTransition.svelte';
  import StaggeredList from '../../lib/components/StaggeredList.svelte';

  // データ状態
  let currentSport = 'volleyball';
  let pendingMatches = [];
  let availableFormats = [];
  let currentFormat = '';
  let selectedMatch = null;
  let isLoading = false;
  let isLoadingFormats = false;
  let isUpdatingFormat = false;
  let showMatchForm = false;
  let refreshInterval = null;

  // スポーツオプション
  const sportOptions = [
    { value: 'volleyball', label: 'バレーボール' },
    { value: 'table_tennis', label: '卓球' },
    { value: 'soccer', label: '8人制サッカー' }
  ];

  // ストアの購読
  let authState;
  let tournamentState;

  $: authState = $authStore;
  $: tournamentState = $tournamentStore;

  /**
   * コンポーネントマウント時の処理
   */
  onMount(async () => {
    // 認証状態をチェック
    if (!authState.isAuthenticated) {
      await authStore.checkAuthStatus();
    }

    // 初期データを読み込み
    await loadInitialData();

    // 定期的なデータ更新を開始（60秒間隔）
    startAutoRefresh();
  });

  /**
   * コンポーネント破棄時の処理
   */
  onDestroy(() => {
    stopAutoRefresh();
  });

  /**
   * 初期データの読み込み
   */
  async function loadInitialData() {
    await Promise.all([
      loadPendingMatches(),
      loadAvailableFormats()
    ]);
  }

  /**
   * 未完了試合一覧の読み込み
   */
  async function loadPendingMatches() {
    isLoading = true;
    
    try {
      const response = await matchAPI.getPendingMatches(currentSport);
      
      if (response.success) {
        pendingMatches = response.data || [];
        uiActions.showNotification(
          `${getSportLabel(currentSport)}の未完了試合を${pendingMatches.length}件取得しました`,
          'success',
          3000
        );
      } else {
        console.error('Failed to load pending matches:', response);
        uiActions.showNotification(
          response.message || '未完了試合の取得に失敗しました',
          'error'
        );
        pendingMatches = [];
      }
    } catch (error) {
      console.error('Load pending matches error:', error);
      uiActions.showNotification(
        '未完了試合の取得でエラーが発生しました',
        'error'
      );
      pendingMatches = [];
    } finally {
      isLoading = false;
    }
  }

  /**
   * 利用可能な形式一覧の読み込み
   */
  async function loadAvailableFormats() {
    isLoadingFormats = true;
    
    try {
      const response = await tournamentAPI.getAvailableFormats(currentSport);
      
      if (response.success) {
        availableFormats = response.data || [];
        
        // 現在の形式を取得
        const tournamentResponse = await tournamentAPI.getTournament(currentSport);
        if (tournamentResponse.success && tournamentResponse.data) {
          currentFormat = tournamentResponse.data.format || '';
        }
      } else {
        console.error('Failed to load available formats:', response);
        availableFormats = [];
        currentFormat = '';
      }
    } catch (error) {
      console.error('Load available formats error:', error);
      availableFormats = [];
      currentFormat = '';
    } finally {
      isLoadingFormats = false;
    }
  }

  /**
   * スポーツ切り替え処理
   */
  async function handleSportChange(event) {
    const newSport = event.target.value;
    
    if (newSport === currentSport) {
      return;
    }

    currentSport = newSport;
    selectedMatch = null;
    showMatchForm = false;

    // 新しいスポーツのデータを読み込み
    await loadInitialData();
    
    // トーナメントストアも更新
    tournamentStore.switchSport(currentSport);
  }

  /**
   * トーナメント形式変更処理
   */
  async function handleFormatChange(event) {
    const newFormat = event.target.value;
    
    if (newFormat === currentFormat) {
      return;
    }

    // 既存の試合結果がある場合は確認ダイアログを表示
    const hasCompletedMatches = pendingMatches.some(match => 
      match.status === 'completed' || (match.score1 !== null && match.score2 !== null)
    );

    if (hasCompletedMatches) {
      const confirmed = confirm(
        '既存の試合結果があります。形式を変更すると、これらの結果に影響する可能性があります。続行しますか？'
      );
      
      if (!confirmed) {
        // 変更をキャンセル - selectの値を元に戻す
        event.target.value = currentFormat;
        return;
      }
    }

    isUpdatingFormat = true;
    
    try {
      const response = await tournamentAPI.updateTournamentFormat(currentSport, newFormat);
      
      if (response.success) {
        currentFormat = newFormat;
        uiActions.showNotification(
          `${getSportLabel(currentSport)}の形式を${newFormat}に変更しました`,
          'success'
        );
        
        // データを再読み込み
        await loadInitialData();
        
        // トーナメントストアも更新
        await tournamentStore.refreshData(currentSport);
      } else {
        console.error('Failed to update tournament format:', response);
        uiActions.showNotification(
          response.message || 'トーナメント形式の変更に失敗しました',
          'error'
        );
        
        // selectの値を元に戻す
        event.target.value = currentFormat;
      }
    } catch (error) {
      console.error('Update tournament format error:', error);
      uiActions.showNotification(
        'トーナメント形式の変更でエラーが発生しました',
        'error'
      );
      
      // selectの値を元に戻す
      event.target.value = currentFormat;
    } finally {
      isUpdatingFormat = false;
    }
  }

  /**
   * 試合編集開始
   */
  function handleEditMatch(match) {
    selectedMatch = match;
    showMatchForm = true;
  }

  /**
   * 試合結果送信成功時の処理
   */
  async function handleMatchSubmitSuccess(event) {
    const { match: matchId, result } = event.detail;
    
    // 未完了試合一覧を更新
    await loadPendingMatches();
    
    // トーナメントデータも更新
    await tournamentStore.refreshData(currentSport);
    
    // フォームを閉じる
    showMatchForm = false;
    selectedMatch = null;
    
    uiActions.showNotification(
      '試合結果を更新し、トーナメントデータを同期しました',
      'success'
    );
  }

  /**
   * 試合結果送信エラー時の処理
   */
  function handleMatchSubmitError(event) {
    const { error } = event.detail;
    console.error('Match submit error:', error);
  }

  /**
   * 試合フォームキャンセル
   */
  function handleMatchFormCancel() {
    showMatchForm = false;
    selectedMatch = null;
  }

  /**
   * データの手動更新
   */
  async function handleRefresh() {
    await loadInitialData();
    await tournamentStore.refreshData(currentSport);
    uiActions.showNotification('データを更新しました', 'success', 2000);
  }

  /**
   * 自動更新の開始
   */
  function startAutoRefresh() {
    refreshInterval = setInterval(async () => {
      // ページが非表示の場合はスキップ
      if (document.hidden) {
        return;
      }
      
      // フォームが開いている場合はスキップ
      if (showMatchForm) {
        return;
      }
      
      try {
        await loadPendingMatches();
      } catch (error) {
        console.error('Auto refresh error:', error);
      }
    }, 60000); // 60秒間隔
  }

  /**
   * 自動更新の停止
   */
  function stopAutoRefresh() {
    if (refreshInterval) {
      clearInterval(refreshInterval);
      refreshInterval = null;
    }
  }

  /**
   * スポーツ名のラベル取得
   */
  function getSportLabel(sport) {
    const option = sportOptions.find(opt => opt.value === sport);
    return option ? option.label : sport;
  }

  /**
   * 試合ステータスのラベル取得
   */
  function getStatusLabel(status) {
    const statusMap = {
      'pending': '未実施',
      'in_progress': '進行中',
      'completed': '完了',
      'cancelled': 'キャンセル'
    };
    return statusMap[status] || status;
  }

  /**
   * 試合ステータスのクラス取得
   */
  function getStatusClass(status) {
    const classMap = {
      'pending': 'status-pending',
      'in_progress': 'status-progress',
      'completed': 'status-completed',
      'cancelled': 'status-cancelled'
    };
    return classMap[status] || 'status-default';
  }

  /**
   * ログアウト処理
   */
  async function handleLogout() {
    const confirmed = confirm('ログアウトしますか？');
    if (confirmed) {
      await authStore.logout();
      // ログアウト後はSvelteKitのナビゲーションでホームページに移動
      window.location.href = '/';
    }
  }
</script>

<svelte:head>
  <title>管理者ダッシュボード - Tournament Management System</title>
</svelte:head>

<ResponsiveLayout let:screenSize>
  <div class="admin-container">
    <ResponsiveLayout container={true} padding={true}>
      <!-- ヘッダー -->
      <header class="admin-header">
        <div class="header-content responsive-flex justify-between align-center">
          <h1 class="responsive-text size-3xl">管理者ダッシュボード</h1>
          <div class="header-actions responsive-flex">
            <Button 
              variant="outline" 
              size="small" 
              on:click={handleRefresh}
              disabled={isLoading}
            >
              更新
            </Button>
            <Button 
              variant="secondary" 
              size="small" 
              on:click={handleLogout}
            >
              ログアウト
            </Button>
          </div>
        </div>
        <p class="header-description responsive-text">試合結果の入力とトーナメント管理</p>
      </header>

      <!-- コントロールパネル -->
      <section class="control-panel">
        <ResponsiveGrid 
          cols={{ mobile: 1, tablet: 2, desktop: 2 }}
          gap="1rem"
          className="control-grid"
        >
          <div class="control-group">
            <label for="sport-select" class="responsive-text size-sm">スポーツ選択</label>
            <Select
              id="sport-select"
              bind:value={currentSport}
              options={sportOptions}
              on:change={handleSportChange}
              disabled={isLoading}
            />
          </div>

          {#if availableFormats.length > 0}
            <div class="control-group">
              <label for="format-select" class="responsive-text size-sm">
                トーナメント形式
                {#if isLoadingFormats}
                  <LoadingSpinner size="small" />
                {/if}
              </label>
              <Select
                id="format-select"
                bind:value={currentFormat}
                options={availableFormats.map(format => ({ value: format, label: format }))}
                on:change={handleFormatChange}
                disabled={isLoadingFormats || isUpdatingFormat}
              />
              {#if isUpdatingFormat}
                <span class="updating-indicator responsive-text size-xs">更新中...</span>
              {/if}
            </div>
          {/if}
        </ResponsiveGrid>
      </section>

      <!-- 未完了試合一覧 -->
      <section class="matches-section">
        <div class="section-header responsive-flex align-center">
          <h2 class="responsive-text size-2xl">{getSportLabel(currentSport)}の未完了試合</h2>
          {#if isLoading}
            <LoadingSpinner size="small" />
          {/if}
        </div>

        {#if isLoading}
          <div class="loading-container">
            <LoadingSpinner />
            <p class="responsive-text">試合データを読み込み中...</p>
          </div>
        {:else if pendingMatches.length === 0}
          <div class="empty-state">
            <p class="responsive-text size-lg">未完了の試合はありません</p>
            <Button variant="outline" on:click={handleRefresh}>
              データを更新
            </Button>
          </div>
        {:else}
          <StaggeredList 
            items={pendingMatches.map(match => ({ ...match, id: match.id }))}
            staggerDelay={100}
            animationType="fadeInUp"
            tag="div"
            itemTag="div"
            className="matches-grid"
            itemClassName="match-grid-item"
          >
            <div 
              slot="default"
              let:item={match}
              class="match-card hover-lift transition-all"
            >
              <div class="match-header responsive-flex justify-between align-center">
                <span class="round-label responsive-text size-sm">{match.round}</span>
                <span class="status-badge {getStatusClass(match.status)} responsive-text size-xs">
                  {getStatusLabel(match.status)}
                </span>
              </div>
              
              <div class="match-teams">
                <div class="team">
                  <span class="team-name responsive-text size-sm">{match.team1}</span>
                  {#if match.score1 !== null}
                    <span class="team-score responsive-text size-xl">{match.score1}</span>
                  {/if}
                </div>
                
                <div class="vs-divider responsive-text size-sm">vs</div>
                
                <div class="team">
                  <span class="team-name responsive-text size-sm">{match.team2}</span>
                  {#if match.score2 !== null}
                    <span class="team-score responsive-text size-xl">{match.score2}</span>
                  {/if}
                </div>
              </div>

              {#if match.winner}
                <AnimatedTransition show={true} type="scale" duration={300}>
                  <div class="winner-info">
                    <span class="winner-label responsive-text size-sm">勝者:</span>
                    <span class="winner-name responsive-text size-sm">{match.winner}</span>
                  </div>
                </AnimatedTransition>
              {/if}

              {#if match.scheduled_at}
                <div class="schedule-info">
                  <span class="schedule-label responsive-text size-xs">予定:</span>
                  <span class="schedule-time responsive-text size-xs">
                    {new Date(match.scheduled_at).toLocaleString('ja-JP')}
                  </span>
                </div>
              {/if}

              <div class="match-actions">
                <Button 
                  variant="primary" 
                  size="small"
                  on:click={() => handleEditMatch(match)}
                  disabled={showMatchForm}
                >
                  結果入力
                </Button>
              </div>
            </div>
          </StaggeredList>
        {/if}
      </section>
    </ResponsiveLayout>
  </div>

  <!-- 試合結果入力フォーム -->
  {#if showMatchForm && selectedMatch}
    <section class="form-section">
      <AnimatedTransition 
        show={true}
        type="fade"
        duration={200}
        className="form-overlay-wrapper"
      >
        <div 
          class="form-overlay" 
          role="button"
          tabindex="0"
          on:click={handleMatchFormCancel}
          on:keydown={(e) => e.key === 'Escape' && handleMatchFormCancel()}
          aria-label="フォームを閉じる"
        ></div>
      </AnimatedTransition>
      
      <AnimatedTransition 
        show={true}
        type="scale"
        duration={300}
        className="form-container-wrapper"
      >
        <div class="form-container">
          <AdminMatchForm 
            match={selectedMatch}
            on:success={handleMatchSubmitSuccess}
            on:error={handleMatchSubmitError}
            on:cancel={handleMatchFormCancel}
          />
        </div>
      </AnimatedTransition>
    </section>
  {/if}
</ResponsiveLayout>

<style>
  .admin-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
    min-height: 100vh;
  }

  /* ヘッダー */
  .admin-header {
    margin-bottom: 2rem;
    padding-bottom: 1rem;
    border-bottom: 2px solid #e9ecef;
  }

  .header-content {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 0.5rem;
  }

  .admin-header h1 {
    color: #333;
    margin: 0;
    font-size: 2rem;
  }

  .header-actions {
    display: flex;
    gap: 1rem;
  }

  .header-description {
    color: #666;
    margin: 0;
    font-size: 1rem;
  }

  /* コントロールパネル */
  .control-panel {
    display: flex;
    gap: 2rem;
    margin-bottom: 2rem;
    padding: 1.5rem;
    background-color: #f8f9fa;
    border-radius: 8px;
    border: 1px solid #dee2e6;
  }

  .control-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-width: 200px;
  }

  .control-group label {
    font-weight: 600;
    color: #495057;
    font-size: 0.9rem;
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .updating-indicator {
    font-size: 0.8rem;
    color: #007bff;
    font-style: italic;
  }

  /* セクションヘッダー */
  .section-header {
    display: flex;
    align-items: center;
    gap: 1rem;
    margin-bottom: 1.5rem;
  }

  .section-header h2 {
    color: #333;
    margin: 0;
    font-size: 1.5rem;
  }

  /* ローディング状態 */
  .loading-container {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 1rem;
    padding: 3rem;
    color: #666;
  }

  /* 空の状態 */
  .empty-state {
    text-align: center;
    padding: 3rem;
    color: #666;
  }

  .empty-state p {
    margin-bottom: 1rem;
    font-size: 1.1rem;
  }

  /* 試合グリッド */
  .matches-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
    gap: 1.5rem;
  }

  .match-grid-item {
    width: 100%;
  }

  /* 試合カード */
  .match-card {
    background: white;
    border: 1px solid #dee2e6;
    border-radius: 8px;
    padding: 1.5rem;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    transition: box-shadow 0.2s, transform 0.2s;
  }

  .match-card:hover {
    box-shadow: 0 4px 8px rgba(0, 0, 0, 0.15);
  }

  .match-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
  }

  .round-label {
    font-weight: 600;
    color: #495057;
    font-size: 0.9rem;
  }

  .status-badge {
    padding: 0.25rem 0.75rem;
    border-radius: 12px;
    font-size: 0.8rem;
    font-weight: 500;
  }

  .status-pending {
    background-color: #fff3cd;
    color: #856404;
  }

  .status-progress {
    background-color: #d1ecf1;
    color: #0c5460;
  }

  .status-completed {
    background-color: #d4edda;
    color: #155724;
  }

  .status-cancelled {
    background-color: #f8d7da;
    color: #721c24;
  }

  .status-default {
    background-color: #e9ecef;
    color: #495057;
  }

  .match-teams {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 1rem;
  }

  .team {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.25rem;
    flex: 1;
  }

  .team-name {
    font-weight: 600;
    color: #333;
    text-align: center;
    font-size: 0.95rem;
  }

  .team-score {
    font-size: 1.5rem;
    font-weight: bold;
    color: #007bff;
  }

  .vs-divider {
    color: #666;
    font-weight: 500;
    margin: 0 1rem;
  }

  .winner-info {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1rem;
    padding: 0.5rem;
    background-color: #d4edda;
    border-radius: 4px;
  }

  .winner-label {
    font-weight: 500;
    color: #155724;
  }

  .winner-name {
    font-weight: bold;
    color: #155724;
  }

  .schedule-info {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1rem;
    font-size: 0.9rem;
    color: #666;
  }

  .schedule-label {
    font-weight: 500;
  }

  .match-actions {
    display: flex;
    justify-content: center;
  }

  /* フォームセクション */
  .form-section {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    z-index: 1000;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  :global(.form-overlay-wrapper) {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
  }

  .form-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    cursor: pointer;
  }

  :global(.form-container-wrapper) {
    position: relative;
    z-index: 1001;
    max-width: 500px;
    width: 90%;
    max-height: 90vh;
  }

  .form-container {
    overflow-y: auto;
  }

  /* レスポンシブデザイン */
  @media (max-width: 768px) {
    .admin-container {
      padding: 1rem;
    }

    .header-content {
      flex-direction: column;
      align-items: flex-start;
      gap: 1rem;
    }

    .control-panel {
      flex-direction: column;
      gap: 1rem;
    }

    .control-group {
      min-width: auto;
    }

    .matches-grid {
      grid-template-columns: 1fr;
    }

    .match-teams {
      flex-direction: column;
      gap: 0.5rem;
    }

    .vs-divider {
      margin: 0;
    }

    .form-container {
      width: 95%;
    }
  }

  @media (max-width: 480px) {
    .admin-container {
      padding: 0.5rem;
    }

    .admin-header h1 {
      font-size: 1.5rem;
    }

    .section-header h2 {
      font-size: 1.25rem;
    }

    .match-card {
      padding: 1rem;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .match-card {
      transition: none;
    }
  }

  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .match-card {
      border: 2px solid #000;
    }
    
    .status-badge {
      border: 1px solid #000;
    }
  }
</style>

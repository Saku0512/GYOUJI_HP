<script>
  import '../app.css';
  import '../lib/styles/responsive.css';
  import { onMount, onDestroy } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { authStore } from '$lib/stores/auth.js';
  import { uiStore, uiActions } from '$lib/stores/ui.js';
  import { setupAuthMonitoring, performLogout } from '$lib/utils/auth-guard.js';
  import { initializeSecurity } from '$lib/utils/security.js';
  import { initializeErrorHandler } from '$lib/utils/error-handler.js';
  import { initializeNetworkMonitor, networkStore } from '$lib/utils/network-monitor.js';
  import { performanceMonitor, runPerformanceTest } from '$lib/utils/performance.js';
  import ErrorBoundary from '$lib/components/ErrorBoundary.svelte';
  import NotificationContainer from '$lib/components/NotificationContainer.svelte';
  import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
  import NotificationToast from '$lib/components/NotificationToast.svelte';
  import ResponsiveLayout from '$lib/components/ResponsiveLayout.svelte';
  import ResponsiveNavigation from '$lib/components/ResponsiveNavigation.svelte';
  import PageTransition from '$lib/components/PageTransition.svelte';

  // ストアの状態を購読
  $: auth = $authStore;
  $: ui = $uiStore;
  $: network = $networkStore;

  // 認証監視のクリーンアップ関数
  let authMonitoringCleanup;

  // ナビゲーションアイテムの定義
  $: navigationItems = [
    {
      key: 'home',
      label: 'ホーム',
      href: '/',
      icon: '🏠'
    },
    ...(auth.isAuthenticated ? [
      {
        key: 'admin',
        label: '管理ダッシュボード',
        href: '/admin',
        icon: '⚙️'
      },
      {
        key: 'logout',
        label: 'ログアウト',
        onClick: handleLogout,
        icon: '🚪'
      }
    ] : [
      {
        key: 'login',
        label: '管理者ログイン',
        href: '/login',
        icon: '🔑'
      }
    ])
  ];

  // 現在のアクティブページ
  $: activeNavItem = getActiveNavItem($page.url.pathname);

  // 初期化処理
  onMount(() => {
    // パフォーマンス監視の開始
    performanceMonitor.measureWebVitals();
    
    // 開発環境でのパフォーマンステスト実行
    if (import.meta.env.DEV) {
      runPerformanceTest();
    }
    
    // エラーハンドリングシステムの初期化
    initializeErrorHandler();
    
    // ネットワーク監視の初期化
    initializeNetworkMonitor();
    
    // セキュリティ機能の初期化
    initializeSecurity();
    
    // テーマの読み込み
    uiActions.loadTheme();
    
    // 認証状態の初期化
    authStore.initialize();
    
    // 認証監視の開始
    authMonitoringCleanup = setupAuthMonitoring();
  });

  // クリーンアップ処理
  onDestroy(() => {
    if (authMonitoringCleanup) {
      authMonitoringCleanup();
    }
    
    // パフォーマンス監視のクリーンアップ
    performanceMonitor.disconnect();
  });

  // ログアウト処理
  async function handleLogout() {
    try {
      uiActions.setLoading(true);
      uiActions.showNotification('ログアウト中...', 'info');
      
      // 認証ガードのログアウト処理を使用
      await performLogout('/');
    } catch (error) {
      console.error('Logout error:', error);
      uiActions.showNotification('ログアウト処理でエラーが発生しました', 'error');
      uiActions.setLoading(false);
    }
  }

  // 通知の削除処理
  function handleNotificationClose(notificationId) {
    uiActions.removeNotification(notificationId);
  }

  // 現在のアクティブページを取得
  function getActiveNavItem(pathname) {
    if (pathname === '/') return 'home';
    if (pathname.startsWith('/admin')) return 'admin';
    if (pathname.startsWith('/login')) return 'login';
    return '';
  }

  // ナビゲーションアイテムクリック処理
  function handleNavItemClick(event) {
    const { item } = event.detail;
    if (item.onClick) {
      item.onClick();
    }
  }
</script>

<ResponsiveLayout>
  <div class="app-layout">
    <!-- ヘッダー -->
    <header class="header">
      <ResponsiveNavigation
        brand="トーナメント管理"
        brandHref="/"
        items={navigationItems}
        activeItem={activeNavItem}
        on:itemClick={handleNavItemClick}
        className="main-navigation"
      />
    </header>

    <!-- メインコンテンツ -->
    <main class="main-content">
      <ErrorBoundary
        errorTitle="アプリケーションエラー"
        errorMessage="アプリケーションでエラーが発生しました。ページを再読み込みするか、しばらく待ってから再試行してください。"
        showRetry={true}
        retryText="再読み込み"
        onRetry={() => window.location.reload()}
      >
        <PageTransition transitionType="fade" duration={300}>
          <slot />
        </PageTransition>
      </ErrorBoundary>
    </main>

    <!-- フッター -->
    <footer class="footer">
      <ResponsiveLayout container={true} padding={true}>
        <div class="footer-content">
          <div class="footer-section">
            <h3 class="footer-title">トーナメント管理システム</h3>
            <p class="footer-description">
              バレーボール、卓球、8人制サッカーのトーナメント管理
            </p>
          </div>
          
          <div class="footer-section">
            <h4 class="footer-subtitle">リンク</h4>
            <ul class="footer-links">
              <li><a href="/">ホーム</a></li>
              {#if auth.isAuthenticated}
                <li><a href="/admin">管理ダッシュボード</a></li>
              {:else}
                <li><a href="/login">管理者ログイン</a></li>
              {/if}
            </ul>
          </div>
        </div>
        
        <div class="footer-bottom">
          <p>&copy; 2024 トーナメント管理システム. All rights reserved.</p>
        </div>
      </ResponsiveLayout>
    </footer>

    <!-- ローディングオーバーレイ -->
    {#if ui.loading || auth.loading}
      <div class="loading-overlay">
        <div class="loading-content">
          <LoadingSpinner size="large" />
          <p class="loading-text">処理中...</p>
        </div>
      </div>
    {/if}

    <!-- ネットワーク状態インジケーター -->
    {#if !network.isOnline}
      <div class="network-status offline" role="alert" aria-live="polite">
        <div class="network-status-content">
          <svg width="20" height="20" viewBox="0 0 20 20" fill="currentColor" class="network-icon">
            <path fill-rule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clip-rule="evenodd" />
            <path d="M3 16l14-14" stroke="currentColor" stroke-width="2"/>
          </svg>
          <span>オフライン - 接続が復旧したら自動的に同期されます</span>
        </div>
      </div>
    {/if}

    <!-- 通知システム -->
    <NotificationContainer
      position="top-right"
      maxNotifications={5}
      defaultDuration={5000}
      pauseOnHover={true}
      stackDirection="down"
    />
  </div>
</ResponsiveLayout>

<style>
  .app-layout {
    min-height: 100vh;
    display: flex;
    flex-direction: column;
  }

  /* ヘッダー */
  .header {
    background-color: #fff;
    border-bottom: 1px solid #e9ecef;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
    position: sticky;
    top: 0;
    z-index: 100;
  }

  /* メインコンテンツ */
  .main-content {
    flex: 1;
    padding: 2rem 0;
  }

  /* フッター */
  .footer {
    background-color: #343a40;
    color: #fff;
    margin-top: auto;
  }

  .footer-content {
    display: grid;
    grid-template-columns: 2fr 1fr;
    gap: 2rem;
    padding: 2rem 0;
  }

  .footer-section {
    margin-bottom: 1rem;
  }

  .footer-title {
    font-size: 1.25rem;
    font-weight: 600;
    margin-bottom: 0.5rem;
    color: #fff;
  }

  .footer-subtitle {
    font-size: 1rem;
    font-weight: 600;
    margin-bottom: 0.5rem;
    color: #adb5bd;
  }

  .footer-description {
    color: #adb5bd;
    margin-bottom: 0;
  }

  .footer-links {
    list-style: none;
    padding: 0;
    margin: 0;
  }

  .footer-links li {
    margin-bottom: 0.25rem;
  }

  .footer-links a {
    color: #adb5bd;
    text-decoration: none;
    transition: color 0.2s ease;
  }

  .footer-links a:hover {
    color: #fff;
    text-decoration: none;
  }

  .footer-bottom {
    border-top: 1px solid #495057;
    padding: 1rem 0;
    text-align: center;
    color: #adb5bd;
    font-size: 0.875rem;
  }

  /* ローディングオーバーレイ */
  .loading-overlay {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    justify-content: center;
    align-items: center;
    z-index: 9999;
  }

  .loading-content {
    background-color: #fff;
    padding: 2rem;
    border-radius: 8px;
    text-align: center;
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.2);
  }

  .loading-text {
    margin-top: 1rem;
    color: #495057;
    font-weight: 500;
  }

  /* ネットワーク状態インジケーター */
  .network-status {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    z-index: 1001;
    padding: 0.5rem;
    text-align: center;
    font-size: 0.875rem;
    font-weight: 500;
  }

  .network-status.offline {
    background-color: #f59e0b;
    color: #fff;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  .network-status-content {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
  }

  .network-icon {
    flex-shrink: 0;
  }



  /* レスポンシブデザイン */
  @media (max-width: 768px) {
    .footer-content {
      grid-template-columns: 1fr;
      gap: 1rem;
    }

    .main-content {
      padding: 1rem 0;
    }

    .notifications-container {
      left: 1rem;
      right: 1rem;
    }
  }

  @media (max-width: 480px) {
    .footer-content {
      padding: 1.5rem 0;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .header {
      background-color: #1f2937;
      border-bottom-color: #374151;
    }

    .footer {
      background-color: #111827;
    }

    .loading-content {
      background-color: #1f2937;
      color: #f9fafb;
    }

    .loading-text {
      color: #d1d5db;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    * {
      transition: none !important;
      animation: none !important;
    }
  }

  /* ハイコントラストモード */
  @media (prefers-contrast: high) {
    .header {
      border-bottom: 2px solid #000;
    }

    .footer {
      border-top: 2px solid #000;
    }
  }
</style>
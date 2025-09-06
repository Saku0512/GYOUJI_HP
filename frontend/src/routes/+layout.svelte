<script>
  import '../app.css';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { authStore } from '$lib/stores/auth.js';
  import { uiStore, uiActions } from '$lib/stores/ui.js';
  import LoadingSpinner from '$lib/components/LoadingSpinner.svelte';
  import NotificationToast from '$lib/components/NotificationToast.svelte';

  // ストアの状態を購読
  $: auth = $authStore;
  $: ui = $uiStore;

  // モバイルメニューの表示状態
  let mobileMenuOpen = false;

  // 初期化処理
  onMount(() => {
    // テーマの読み込み
    uiActions.loadTheme();
    
    // 認証状態の初期化
    authStore.initialize();
  });

  // ログアウト処理
  async function handleLogout() {
    try {
      uiActions.setLoading(true);
      const result = await authStore.logout();
      
      if (result.success) {
        uiActions.showNotification('ログアウトしました', 'success');
        goto('/');
      } else {
        uiActions.showNotification(result.message || 'ログアウトに失敗しました', 'error');
      }
    } catch (error) {
      console.error('Logout error:', error);
      uiActions.showNotification('ログアウト処理でエラーが発生しました', 'error');
    } finally {
      uiActions.setLoading(false);
    }
  }

  // モバイルメニューの切り替え
  function toggleMobileMenu() {
    mobileMenuOpen = !mobileMenuOpen;
  }

  // モバイルメニューを閉じる
  function closeMobileMenu() {
    mobileMenuOpen = false;
  }

  // 通知の削除処理
  function handleNotificationClose(notificationId) {
    uiActions.removeNotification(notificationId);
  }

  // キーボードナビゲーション
  function handleKeydown(event) {
    // Escapeキーでモバイルメニューを閉じる
    if (event.key === 'Escape' && mobileMenuOpen) {
      closeMobileMenu();
    }
  }

  // 現在のページがアクティブかどうかを判定
  function isActivePage(path) {
    return $page.url.pathname === path;
  }
</script>

<svelte:window on:keydown={handleKeydown} />

<div class="app-layout">
  <!-- ヘッダー -->
  <header class="header">
    <div class="container">
      <nav class="navbar">
        <!-- ロゴ -->
        <div class="navbar-brand">
          <a href="/" class="brand-link">
            <h1 class="brand-title">トーナメント管理</h1>
          </a>
        </div>

        <!-- デスクトップナビゲーション -->
        <div class="navbar-nav desktop-nav">
          <a 
            href="/" 
            class="nav-link"
            class:active={isActivePage('/')}
          >
            ホーム
          </a>
          
          {#if auth.isAuthenticated}
            <a 
              href="/admin" 
              class="nav-link"
              class:active={isActivePage('/admin')}
            >
              管理ダッシュボード
            </a>
            <button 
              class="nav-button logout-button"
              on:click={handleLogout}
              disabled={auth.loading}
            >
              ログアウト
            </button>
          {:else}
            <a 
              href="/login" 
              class="nav-link login-link"
              class:active={isActivePage('/login')}
            >
              管理者ログイン
            </a>
          {/if}
        </div>

        <!-- モバイルメニューボタン -->
        <button 
          class="mobile-menu-button"
          on:click={toggleMobileMenu}
          aria-label="メニューを開く"
          aria-expanded={mobileMenuOpen}
        >
          <span class="hamburger-line" class:open={mobileMenuOpen}></span>
          <span class="hamburger-line" class:open={mobileMenuOpen}></span>
          <span class="hamburger-line" class:open={mobileMenuOpen}></span>
        </button>
      </nav>
    </div>

    <!-- モバイルナビゲーション -->
    {#if mobileMenuOpen}
      <div class="mobile-nav" class:open={mobileMenuOpen}>
        <div class="mobile-nav-content">
          <a 
            href="/" 
            class="mobile-nav-link"
            class:active={isActivePage('/')}
            on:click={closeMobileMenu}
          >
            ホーム
          </a>
          
          {#if auth.isAuthenticated}
            <a 
              href="/admin" 
              class="mobile-nav-link"
              class:active={isActivePage('/admin')}
              on:click={closeMobileMenu}
            >
              管理ダッシュボード
            </a>
            <button 
              class="mobile-nav-button logout-button"
              on:click={() => { handleLogout(); closeMobileMenu(); }}
              disabled={auth.loading}
            >
              ログアウト
            </button>
          {:else}
            <a 
              href="/login" 
              class="mobile-nav-link login-link"
              class:active={isActivePage('/login')}
              on:click={closeMobileMenu}
            >
              管理者ログイン
            </a>
          {/if}
        </div>
      </div>
    {/if}
  </header>

  <!-- メインコンテンツ -->
  <main class="main-content">
    <slot />
  </main>

  <!-- フッター -->
  <footer class="footer">
    <div class="container">
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
    </div>
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

  <!-- 通知システム -->
  <div class="notifications-container">
    {#each ui.notifications as notification (notification.id)}
      <NotificationToast
        message={notification.message}
        type={notification.type}
        duration={0}
        on:close={() => handleNotificationClose(notification.id)}
      />
    {/each}
  </div>
</div>

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

  .navbar {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 1rem 0;
  }

  .navbar-brand {
    flex-shrink: 0;
  }

  .brand-link {
    text-decoration: none;
    color: inherit;
  }

  .brand-title {
    font-size: 1.5rem;
    font-weight: 700;
    color: #007bff;
    margin: 0;
  }

  /* デスクトップナビゲーション */
  .desktop-nav {
    display: flex;
    align-items: center;
    gap: 1.5rem;
  }

  .nav-link {
    color: #495057;
    text-decoration: none;
    font-weight: 500;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    transition: all 0.2s ease;
  }

  .nav-link:hover {
    color: #007bff;
    background-color: #f8f9fa;
    text-decoration: none;
  }

  .nav-link.active {
    color: #007bff;
    background-color: #e3f2fd;
  }

  .nav-button {
    background: none;
    border: 1px solid #007bff;
    color: #007bff;
    padding: 0.5rem 1rem;
    border-radius: 4px;
    font-weight: 500;
    transition: all 0.2s ease;
  }

  .nav-button:hover {
    background-color: #007bff;
    color: white;
  }

  .logout-button {
    border-color: #dc3545;
    color: #dc3545;
  }

  .logout-button:hover {
    background-color: #dc3545;
    color: white;
  }

  .login-link {
    background-color: #007bff;
    color: white !important;
  }

  .login-link:hover {
    background-color: #0056b3;
  }

  /* モバイルメニューボタン */
  .mobile-menu-button {
    display: none;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    width: 40px;
    height: 40px;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
  }

  .hamburger-line {
    width: 24px;
    height: 2px;
    background-color: #495057;
    transition: all 0.3s ease;
    margin: 2px 0;
  }

  .hamburger-line.open:nth-child(1) {
    transform: rotate(45deg) translate(5px, 5px);
  }

  .hamburger-line.open:nth-child(2) {
    opacity: 0;
  }

  .hamburger-line.open:nth-child(3) {
    transform: rotate(-45deg) translate(7px, -6px);
  }

  /* モバイルナビゲーション */
  .mobile-nav {
    display: none;
    background-color: #fff;
    border-top: 1px solid #e9ecef;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
  }

  .mobile-nav-content {
    padding: 1rem;
  }

  .mobile-nav-link {
    display: block;
    color: #495057;
    text-decoration: none;
    font-weight: 500;
    padding: 0.75rem 1rem;
    border-radius: 4px;
    margin-bottom: 0.5rem;
    transition: all 0.2s ease;
  }

  .mobile-nav-link:hover {
    color: #007bff;
    background-color: #f8f9fa;
    text-decoration: none;
  }

  .mobile-nav-link.active {
    color: #007bff;
    background-color: #e3f2fd;
  }

  .mobile-nav-button {
    display: block;
    width: 100%;
    background: none;
    border: 1px solid #007bff;
    color: #007bff;
    padding: 0.75rem 1rem;
    border-radius: 4px;
    font-weight: 500;
    margin-top: 0.5rem;
    transition: all 0.2s ease;
  }

  .mobile-nav-button:hover {
    background-color: #007bff;
    color: white;
  }

  .mobile-nav-button.logout-button {
    border-color: #dc3545;
    color: #dc3545;
  }

  .mobile-nav-button.logout-button:hover {
    background-color: #dc3545;
    color: white;
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

  /* 通知コンテナ */
  .notifications-container {
    position: fixed;
    top: 1rem;
    right: 1rem;
    z-index: 1000;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  /* レスポンシブデザイン */
  @media (max-width: 768px) {
    .desktop-nav {
      display: none;
    }

    .mobile-menu-button {
      display: flex;
    }

    .mobile-nav {
      display: block;
    }

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
    .navbar {
      padding: 0.75rem 0;
    }

    .brand-title {
      font-size: 1.25rem;
    }

    .footer-content {
      padding: 1.5rem 0;
    }
  }
</style>

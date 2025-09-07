<!--
  オフラインページ
  ネットワーク接続がない場合に表示される
-->
<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';

  let isOnline = true;
  let retryCount = 0;
  let maxRetries = 3;

  // オンライン状態の監視
  function updateOnlineStatus() {
    isOnline = navigator.onLine;
    
    if (isOnline && retryCount < maxRetries) {
      // オンラインに復帰したらホームページに戻る
      setTimeout(() => {
        goto('/');
      }, 1000);
    }
  }

  // 手動で再接続を試行
  async function retryConnection() {
    retryCount++;
    
    try {
      // 軽量なリクエストでネットワーク接続をテスト
      const response = await fetch('/api/health', {
        method: 'HEAD',
        cache: 'no-cache'
      });
      
      if (response.ok) {
        isOnline = true;
        goto('/');
      } else {
        throw new Error('Network test failed');
      }
    } catch (error) {
      console.warn('Retry connection failed:', error);
      
      if (retryCount >= maxRetries) {
        // 最大試行回数に達した場合は少し待ってからリセット
        setTimeout(() => {
          retryCount = 0;
        }, 30000); // 30秒後にリセット
      }
    }
  }

  onMount(() => {
    // オンライン/オフライン状態の監視
    window.addEventListener('online', updateOnlineStatus);
    window.addEventListener('offline', updateOnlineStatus);
    
    // 初期状態を設定
    updateOnlineStatus();
    
    return () => {
      window.removeEventListener('online', updateOnlineStatus);
      window.removeEventListener('offline', updateOnlineStatus);
    };
  });
</script>

<svelte:head>
  <title>オフライン - トーナメント管理システム</title>
  <meta name="description" content="現在オフラインです。ネットワーク接続を確認してください。" />
</svelte:head>

<div class="offline-container">
  <div class="offline-content">
    <!-- オフラインアイコン -->
    <div class="offline-icon">
      <svg width="80" height="80" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
        <path d="M23.64 7C23.89 7.24 24 7.62 24 8V16C24 16.38 23.89 16.76 23.64 17L17 10.5L23.64 7Z" fill="#666"/>
        <path d="M1.27 4.27L19.73 22.73C20.12 23.12 20.76 23.12 21.15 22.73C21.54 22.34 21.54 21.7 21.15 21.31L2.69 2.85C2.3 2.46 1.66 2.46 1.27 2.85C0.88 3.24 0.88 3.88 1.27 4.27Z" fill="#f44336"/>
        <path d="M.36 7C.11 7.24 0 7.62 0 8V16C0 17.1.9 18 2 18H22C22.05 18 22.1 18 22.14 17.99L18 13.85V16H6V10.14L.36 7Z" fill="#666"/>
      </svg>
    </div>

    <!-- メッセージ -->
    <h1>オフラインです</h1>
    <p class="offline-message">
      現在インターネット接続がありません。<br>
      ネットワーク接続を確認してから再度お試しください。
    </p>

    <!-- 接続状態 -->
    <div class="connection-status" class:online={isOnline} class:offline={!isOnline}>
      <div class="status-indicator"></div>
      <span>{isOnline ? 'オンライン' : 'オフライン'}</span>
    </div>

    <!-- 再試行ボタン -->
    <div class="retry-section">
      <button 
        class="retry-button" 
        on:click={retryConnection}
        disabled={retryCount >= maxRetries}
      >
        {#if retryCount >= maxRetries}
          しばらくお待ちください...
        {:else}
          再接続を試行 ({retryCount}/{maxRetries})
        {/if}
      </button>
      
      {#if retryCount > 0}
        <p class="retry-info">
          {retryCount}回試行しました。
          {#if retryCount >= maxRetries}
            30秒後に再度お試しいただけます。
          {/if}
        </p>
      {/if}
    </div>

    <!-- オフライン時の機能説明 -->
    <div class="offline-features">
      <h3>オフライン時でも利用可能な機能</h3>
      <ul>
        <li>キャッシュされたトーナメント情報の閲覧</li>
        <li>過去に表示したページの閲覧</li>
        <li>アプリケーションの基本機能</li>
      </ul>
      
      <p class="note">
        ※ 最新の情報を取得するには、インターネット接続が必要です。
      </p>
    </div>

    <!-- ホームに戻るリンク -->
    <div class="navigation">
      <a href="/" class="home-link">
        ホームページに戻る
      </a>
    </div>
  </div>
</div>

<style>
  .offline-container {
    min-height: 100vh;
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 20px;
    background: linear-gradient(135deg, #f5f7fa 0%, #c3cfe2 100%);
  }

  .offline-content {
    max-width: 500px;
    text-align: center;
    background: white;
    padding: 40px;
    border-radius: 12px;
    box-shadow: 0 10px 30px rgba(0, 0, 0, 0.1);
  }

  .offline-icon {
    margin-bottom: 24px;
    opacity: 0.8;
  }

  h1 {
    font-size: 2rem;
    color: #333;
    margin-bottom: 16px;
    font-weight: 600;
  }

  .offline-message {
    font-size: 1.1rem;
    color: #666;
    line-height: 1.6;
    margin-bottom: 32px;
  }

  .connection-status {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    margin-bottom: 32px;
    padding: 12px 20px;
    border-radius: 8px;
    font-weight: 500;
  }

  .connection-status.online {
    background: #e8f5e8;
    color: #2e7d32;
  }

  .connection-status.offline {
    background: #ffebee;
    color: #c62828;
  }

  .status-indicator {
    width: 12px;
    height: 12px;
    border-radius: 50%;
    animation: pulse 2s infinite;
  }

  .connection-status.online .status-indicator {
    background: #4caf50;
  }

  .connection-status.offline .status-indicator {
    background: #f44336;
  }

  @keyframes pulse {
    0% {
      opacity: 1;
      transform: scale(1);
    }
    50% {
      opacity: 0.7;
      transform: scale(1.1);
    }
    100% {
      opacity: 1;
      transform: scale(1);
    }
  }

  .retry-section {
    margin-bottom: 32px;
  }

  .retry-button {
    background: #007bff;
    color: white;
    border: none;
    padding: 12px 24px;
    border-radius: 8px;
    font-size: 1rem;
    cursor: pointer;
    transition: all 0.2s ease;
    margin-bottom: 12px;
  }

  .retry-button:hover:not(:disabled) {
    background: #0056b3;
    transform: translateY(-1px);
  }

  .retry-button:disabled {
    background: #ccc;
    cursor: not-allowed;
    transform: none;
  }

  .retry-info {
    font-size: 0.9rem;
    color: #666;
    margin: 0;
  }

  .offline-features {
    text-align: left;
    background: #f8f9fa;
    padding: 24px;
    border-radius: 8px;
    margin-bottom: 32px;
  }

  .offline-features h3 {
    margin: 0 0 16px 0;
    color: #333;
    font-size: 1.1rem;
  }

  .offline-features ul {
    margin: 0 0 16px 0;
    padding-left: 20px;
  }

  .offline-features li {
    margin-bottom: 8px;
    color: #555;
  }

  .note {
    font-size: 0.9rem;
    color: #666;
    font-style: italic;
    margin: 0;
  }

  .navigation {
    border-top: 1px solid #eee;
    padding-top: 24px;
  }

  .home-link {
    display: inline-block;
    color: #007bff;
    text-decoration: none;
    font-weight: 500;
    padding: 8px 16px;
    border: 2px solid #007bff;
    border-radius: 8px;
    transition: all 0.2s ease;
  }

  .home-link:hover {
    background: #007bff;
    color: white;
    transform: translateY(-1px);
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .offline-container {
      background: linear-gradient(135deg, #2c3e50 0%, #34495e 100%);
    }

    .offline-content {
      background: #2a2a2a;
      color: #fff;
    }

    h1 {
      color: #fff;
    }

    .offline-message {
      color: #ccc;
    }

    .offline-features {
      background: #333;
    }

    .offline-features h3 {
      color: #fff;
    }

    .offline-features li {
      color: #ccc;
    }

    .note {
      color: #999;
    }

    .navigation {
      border-top-color: #444;
    }
  }

  /* モバイル対応 */
  @media (max-width: 768px) {
    .offline-container {
      padding: 16px;
    }

    .offline-content {
      padding: 24px;
    }

    h1 {
      font-size: 1.5rem;
    }

    .offline-message {
      font-size: 1rem;
    }

    .offline-features {
      padding: 16px;
      text-align: center;
    }

    .offline-features ul {
      text-align: left;
    }
  }
</style>
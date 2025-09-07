<script>
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  
  // レスポンシブナビゲーションコンポーネント
  export let items = [];
  export let activeItem = '';
  export let brand = '';
  export let brandHref = '/';
  export let mobileBreakpoint = 768;
  export let className = '';

  const dispatch = createEventDispatcher();

  let isOpen = false;
  let isMobile = false;
  let navElement;

  // モバイル表示の判定
  function checkMobile() {
    if (typeof window !== 'undefined') {
      isMobile = window.innerWidth < mobileBreakpoint;
    }
  }

  // メニューの開閉
  function toggleMenu() {
    isOpen = !isOpen;
    dispatch('toggle', { isOpen });
  }

  // メニューを閉じる
  function closeMenu() {
    isOpen = false;
    dispatch('close');
  }

  // アイテムクリック処理
  function handleItemClick(item, event) {
    if (isMobile) {
      closeMenu();
    }
    dispatch('itemClick', { item, event });
  }

  // 外部クリックでメニューを閉じる
  function handleOutsideClick(event) {
    if (isOpen && navElement && !navElement.contains(event.target)) {
      closeMenu();
    }
  }

  // キーボードナビゲーション
  function handleKeydown(event) {
    if (event.key === 'Escape' && isOpen) {
      closeMenu();
    }
  }

  // リサイズイベントの処理
  function handleResize() {
    checkMobile();
    if (!isMobile && isOpen) {
      closeMenu();
    }
  }

  onMount(() => {
    checkMobile();
    window.addEventListener('resize', handleResize);
    document.addEventListener('click', handleOutsideClick);
    document.addEventListener('keydown', handleKeydown);
  });

  onDestroy(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', handleResize);
      document.removeEventListener('click', handleOutsideClick);
      document.removeEventListener('keydown', handleKeydown);
    }
  });

  $: navClass = [
    'responsive-nav',
    isOpen ? 'nav-open' : '',
    isMobile ? 'nav-mobile' : 'nav-desktop',
    className
  ].filter(Boolean).join(' ');
</script>

<nav class={navClass} bind:this={navElement}>
  <!-- ブランド/ロゴ -->
  {#if brand}
    <div class="nav-brand">
      <a href={brandHref} class="brand-link">
        {brand}
      </a>
    </div>
  {/if}

  <!-- モバイルメニューボタン -->
  {#if isMobile}
    <button 
      class="nav-toggle"
      class:active={isOpen}
      on:click={toggleMenu}
      aria-label="メニューを{isOpen ? '閉じる' : '開く'}"
      aria-expanded={isOpen}
    >
      <span class="hamburger-line"></span>
      <span class="hamburger-line"></span>
      <span class="hamburger-line"></span>
    </button>
  {/if}

  <!-- ナビゲーションメニュー -->
  <div class="nav-menu" class:show={isOpen || !isMobile}>
    <ul class="nav-list">
      {#each items as item}
        <li class="nav-item">
          {#if item.href}
            <a 
              href={item.href}
              class="nav-link"
              class:active={activeItem === item.key}
              on:click={(e) => handleItemClick(item, e)}
            >
              {#if item.icon}
                <span class="nav-icon">{item.icon}</span>
              {/if}
              <span class="nav-text">{item.label}</span>
            </a>
          {:else if item.onClick}
            <button 
              class="nav-button"
              class:active={activeItem === item.key}
              on:click={(e) => handleItemClick(item, e)}
            >
              {#if item.icon}
                <span class="nav-icon">{item.icon}</span>
              {/if}
              <span class="nav-text">{item.label}</span>
            </button>
          {/if}
        </li>
      {/each}
    </ul>
  </div>

  <!-- オーバーレイ (モバイル時) -->
  {#if isMobile && isOpen}
    <div 
      class="nav-overlay"
      on:click={closeMenu}
      role="button"
      tabindex="0"
      on:keydown={(e) => e.key === 'Enter' && closeMenu()}
      aria-label="メニューを閉じる"
    ></div>
  {/if}
</nav>

<style>
  .responsive-nav {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1rem;
    background-color: #ffffff;
    border-bottom: 1px solid #e5e7eb;
    position: relative;
    z-index: 100;
  }

  .nav-brand {
    flex-shrink: 0;
  }

  .brand-link {
    font-size: 1.5rem;
    font-weight: 700;
    color: #1f2937;
    text-decoration: none;
    transition: color 0.2s ease;
  }

  .brand-link:hover {
    color: #3b82f6;
  }

  /* デスクトップナビゲーション */
  .nav-desktop .nav-menu {
    display: flex;
  }

  .nav-desktop .nav-toggle {
    display: none;
  }

  .nav-list {
    display: flex;
    align-items: center;
    gap: 1rem;
    list-style: none;
    margin: 0;
    padding: 0;
  }

  .nav-item {
    position: relative;
  }

  .nav-link,
  .nav-button {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.5rem 1rem;
    color: #4b5563;
    text-decoration: none;
    border: none;
    background: none;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s ease;
    min-height: 44px;
  }

  .nav-link:hover,
  .nav-button:hover {
    color: #3b82f6;
    background-color: #f3f4f6;
  }

  .nav-link.active,
  .nav-button.active {
    color: #3b82f6;
    background-color: #dbeafe;
  }

  .nav-icon {
    font-size: 1rem;
  }

  .nav-text {
    white-space: nowrap;
  }

  /* モバイルナビゲーション */
  .nav-mobile .nav-menu {
    position: fixed;
    top: 0;
    right: -100%;
    width: 280px;
    height: 100vh;
    background-color: #ffffff;
    box-shadow: -4px 0 6px rgba(0, 0, 0, 0.1);
    transition: right 0.3s ease;
    z-index: 1001;
    overflow-y: auto;
  }

  .nav-mobile .nav-menu.show {
    right: 0;
  }

  .nav-mobile .nav-list {
    flex-direction: column;
    align-items: stretch;
    gap: 0;
    padding: 2rem 1rem;
  }

  .nav-mobile .nav-link,
  .nav-mobile .nav-button {
    justify-content: flex-start;
    padding: 1rem;
    border-radius: 8px;
    margin-bottom: 0.5rem;
    min-height: 48px;
  }

  /* モバイルメニューボタン */
  .nav-toggle {
    display: flex;
    flex-direction: column;
    justify-content: center;
    align-items: center;
    width: 44px;
    height: 44px;
    background: none;
    border: none;
    cursor: pointer;
    padding: 0;
    z-index: 1002;
  }

  .hamburger-line {
    width: 24px;
    height: 2px;
    background-color: #374151;
    transition: all 0.3s ease;
    margin: 2px 0;
  }

  .nav-toggle.active .hamburger-line:nth-child(1) {
    transform: rotate(45deg) translate(5px, 5px);
  }

  .nav-toggle.active .hamburger-line:nth-child(2) {
    opacity: 0;
  }

  .nav-toggle.active .hamburger-line:nth-child(3) {
    transform: rotate(-45deg) translate(7px, -6px);
  }

  /* オーバーレイ */
  .nav-overlay {
    position: fixed;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background-color: rgba(0, 0, 0, 0.5);
    z-index: 1000;
    cursor: pointer;
  }

  /* アクセシビリティ対応 */
  .nav-link:focus-visible,
  .nav-button:focus-visible,
  .nav-toggle:focus-visible {
    outline: 2px solid #3b82f6;
    outline-offset: 2px;
  }

  /* アニメーション無効化 */
  @media (prefers-reduced-motion: reduce) {
    .nav-menu,
    .hamburger-line,
    .nav-link,
    .nav-button {
      transition: none;
    }
  }

  /* ハイコントラストモード */
  @media (prefers-contrast: high) {
    .responsive-nav {
      border-bottom: 2px solid #000;
    }

    .nav-link,
    .nav-button {
      border: 1px solid transparent;
    }

    .nav-link:hover,
    .nav-button:hover,
    .nav-link.active,
    .nav-button.active {
      border-color: #000;
    }
  }

  /* ダークモード */
  @media (prefers-color-scheme: dark) {
    .responsive-nav {
      background-color: #1f2937;
      border-bottom-color: #374151;
    }

    .brand-link {
      color: #f9fafb;
    }

    .brand-link:hover {
      color: #60a5fa;
    }

    .nav-mobile .nav-menu {
      background-color: #1f2937;
    }

    .nav-link,
    .nav-button {
      color: #d1d5db;
    }

    .nav-link:hover,
    .nav-button:hover {
      color: #60a5fa;
      background-color: #374151;
    }

    .nav-link.active,
    .nav-button.active {
      color: #60a5fa;
      background-color: #1e3a8a;
    }

    .hamburger-line {
      background-color: #d1d5db;
    }
  }

  /* プリント用 */
  @media print {
    .nav-toggle,
    .nav-overlay {
      display: none !important;
    }

    .nav-mobile .nav-menu {
      position: static;
      width: auto;
      height: auto;
      box-shadow: none;
    }

    .nav-mobile .nav-list {
      flex-direction: row;
      padding: 0;
    }
  }
</style>
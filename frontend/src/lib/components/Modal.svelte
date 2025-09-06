<script>
  // Modal コンポーネント - モーダルダイアログ
  import { createEventDispatcher, onMount } from 'svelte';
  
  export let open = false;
  export let title = '';
  export let size = 'medium'; // 'small', 'medium', 'large', 'full'
  export let closable = true; // 閉じるボタンを表示するか
  export let closeOnBackdrop = true; // 背景クリックで閉じるか
  export let closeOnEscape = true; // Escapeキーで閉じるか
  
  const dispatch = createEventDispatcher();
  
  let modalElement;
  let previousActiveElement;
  
  // モーダルが開いたときの処理
  $: if (open) {
    handleOpen();
  } else {
    handleClose();
  }
  
  function handleOpen() {
    // 現在のフォーカス要素を保存
    previousActiveElement = document.activeElement;
    
    // body のスクロールを無効化
    document.body.style.overflow = 'hidden';
    
    // モーダルにフォーカスを移動
    setTimeout(() => {
      if (modalElement) {
        modalElement.focus();
      }
    }, 0);
  }
  
  function handleClose() {
    // body のスクロールを復元
    document.body.style.overflow = '';
    
    // 前のフォーカス要素に戻す
    if (previousActiveElement) {
      previousActiveElement.focus();
    }
  }
  
  function close() {
    dispatch('close');
  }
  
  function handleBackdropClick(event) {
    if (closeOnBackdrop && event.target === event.currentTarget) {
      close();
    }
  }
  
  function handleKeydown(event) {
    if (event.key === 'Escape' && closeOnEscape) {
      close();
    }
    
    // Tab キーでのフォーカストラップ
    if (event.key === 'Tab') {
      trapFocus(event);
    }
  }
  
  function trapFocus(event) {
    if (!modalElement) return;
    
    const focusableElements = modalElement.querySelectorAll(
      'button, [href], input, select, textarea, [tabindex]:not([tabindex="-1"])'
    );
    
    const firstElement = focusableElements[0];
    const lastElement = focusableElements[focusableElements.length - 1];
    
    if (event.shiftKey) {
      if (document.activeElement === firstElement) {
        event.preventDefault();
        lastElement.focus();
      }
    } else {
      if (document.activeElement === lastElement) {
        event.preventDefault();
        firstElement.focus();
      }
    }
  }
  
  onMount(() => {
    return () => {
      // コンポーネントが破棄されるときにスクロールを復元
      document.body.style.overflow = '';
    };
  });
</script>

{#if open}
  <div
    class="modal-backdrop"
    on:click={handleBackdropClick}
    on:keydown={handleKeydown}
    role="dialog"
    aria-modal="true"
    aria-labelledby={title ? 'modal-title' : undefined}
    tabindex="0"
  >
    <div
      bind:this={modalElement}
      class="modal modal-{size}"
      tabindex="-1"
    >
      {#if title || closable}
        <div class="modal-header">
          {#if title}
            <h2 id="modal-title" class="modal-title">{title}</h2>
          {/if}
          {#if closable}
            <button
              class="modal-close"
              on:click={close}
              aria-label="モーダルを閉じる"
              type="button"
            >
              <svg width="24" height="24" viewBox="0 0 20 20" fill="currentColor">
                <path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
              </svg>
            </button>
          {/if}
        </div>
      {/if}
      
      <div class="modal-body">
        <slot />
      </div>
      
      <slot name="footer" />
    </div>
  </div>
{/if}

<style>
  .modal-backdrop {
    position: fixed;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 1000;
    padding: 1rem;
    animation: fadeIn 0.2s ease-out;
  }
  
  .modal {
    background: white;
    border-radius: 0.5rem;
    box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.25);
    max-height: 90vh;
    overflow-y: auto;
    animation: scaleIn 0.2s ease-out;
  }
  
  .modal:focus {
    outline: none;
  }
  
  .modal-small {
    width: 100%;
    max-width: 400px;
  }
  
  .modal-medium {
    width: 100%;
    max-width: 600px;
  }
  
  .modal-large {
    width: 100%;
    max-width: 800px;
  }
  
  .modal-full {
    width: 95vw;
    height: 95vh;
    max-width: none;
    max-height: none;
  }
  
  .modal-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 1.5rem 1.5rem 0 1.5rem;
    border-bottom: 1px solid #e5e7eb;
    margin-bottom: 1rem;
    padding-bottom: 1rem;
  }
  
  .modal-title {
    margin: 0;
    font-size: 1.25rem;
    font-weight: 600;
    color: #111827;
  }
  
  .modal-close {
    background: none;
    border: none;
    cursor: pointer;
    padding: 0.5rem;
    border-radius: 0.25rem;
    color: #6b7280;
    transition: color 0.2s, background-color 0.2s;
  }
  
  .modal-close:hover {
    color: #374151;
    background-color: #f3f4f6;
  }
  
  .modal-close:focus {
    outline: 2px solid #007bff;
    outline-offset: 2px;
  }
  
  .modal-body {
    padding: 0 1.5rem 1.5rem 1.5rem;
  }
  
  @keyframes fadeIn {
    from {
      opacity: 0;
    }
    to {
      opacity: 1;
    }
  }
  
  @keyframes scaleIn {
    from {
      opacity: 0;
      transform: scale(0.95);
    }
    to {
      opacity: 1;
      transform: scale(1);
    }
  }
  
  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .modal-backdrop {
      padding: 0.5rem;
    }
    
    .modal-small,
    .modal-medium,
    .modal-large {
      max-width: none;
      width: 100%;
    }
    
    .modal-full {
      width: 100vw;
      height: 100vh;
      border-radius: 0;
    }
    
    .modal-header {
      padding: 1rem 1rem 0 1rem;
    }
    
    .modal-body {
      padding: 0 1rem 1rem 1rem;
    }
  }
</style>
<script>
  // ErrorBoundary コンポーネント - エラー境界の実装
  import { createEventDispatcher, onMount, onDestroy } from 'svelte';
  import { globalErrorHandler, AppError, ERROR_TYPES, ERROR_LEVELS } from '$lib/utils/error-handler.js';
  import { uiActions } from '$lib/stores/ui.js';
  import Button from './Button.svelte';
  import LoadingSpinner from './LoadingSpinner.svelte';

  // Props
  export let fallback = null; // カスタムフォールバックコンポーネント
  export let showRetry = true; // 再試行ボタンを表示するか
  export let retryText = '再試行'; // 再試行ボタンのテキスト
  export let errorTitle = 'エラーが発生しました'; // エラータイトル
  export let errorMessage = ''; // カスタムエラーメッセージ
  export let level = ERROR_LEVELS.MEDIUM; // エラーレベル
  export let onError = null; // エラー発生時のコールバック
  export let onRetry = null; // 再試行時のコールバック
  export let resetOnPropsChange = true; // プロパティ変更時にエラー状態をリセットするか

  const dispatch = createEventDispatcher();

  // 内部状態
  let hasError = false;
  let error = null;
  let retrying = false;
  let errorId = null;

  // エラーリスナー
  let errorListener = null;

  // 初期化
  onMount(() => {
    // エラーリスナーを登録
    errorListener = (capturedError) => {
      handleError(capturedError);
    };
    
    globalErrorHandler.addErrorListener(errorListener);
  });

  // クリーンアップ
  onDestroy(() => {
    if (errorListener) {
      globalErrorHandler.removeErrorListener(errorListener);
    }
  });

  // プロパティ変更時のエラー状態リセット
  $: if (resetOnPropsChange && hasError) {
    // プロパティが変更された場合はエラー状態をリセット
    resetError();
  }

  // エラーハンドリング
  function handleError(capturedError) {
    if (hasError) return; // 既にエラー状態の場合は無視

    error = capturedError;
    hasError = true;
    errorId = `error-boundary-${Date.now()}`;

    // カスタムエラーハンドラーを呼び出し
    if (onError) {
      try {
        onError(error);
      } catch (handlerError) {
        console.error('[ErrorBoundary] Error in custom error handler:', handlerError);
      }
    }

    // イベントを発火
    dispatch('error', { error, errorId });

    console.error('[ErrorBoundary] Caught error:', error);
  }

  // エラー状態のリセット
  function resetError() {
    hasError = false;
    error = null;
    retrying = false;
    errorId = null;

    // エラー境界をリセット
    if (errorId) {
      globalErrorHandler.resetErrorBoundary(errorId);
    }

    // イベントを発火
    dispatch('reset');
  }

  // 再試行処理
  async function handleRetry() {
    if (retrying) return;

    retrying = true;

    try {
      // カスタム再試行ハンドラーがある場合は実行
      if (onRetry) {
        await onRetry(error);
      }

      // 成功した場合はエラー状態をリセット
      resetError();
      
      // 成功通知
      uiActions.showNotification('再試行が成功しました', 'success');
      
      // イベントを発火
      dispatch('retry-success');
    } catch (retryError) {
      console.error('[ErrorBoundary] Retry failed:', retryError);
      
      // 再試行失敗の通知
      uiActions.showNotification('再試行に失敗しました', 'error');
      
      // イベントを発火
      dispatch('retry-failed', { error: retryError });
    } finally {
      retrying = false;
    }
  }

  // エラーレベルに基づくスタイルクラス
  $: errorLevelClass = `error-level-${level}`;

  // エラーメッセージの取得
  $: displayErrorMessage = errorMessage || (error?.message) || 'エラーが発生しました';

  // エラー詳細の表示判定
  $: showErrorDetails = error && (error.level === ERROR_LEVELS.HIGH || error.level === ERROR_LEVELS.CRITICAL);
</script>

{#if hasError}
  <!-- カスタムフォールバックがある場合 -->
  {#if fallback}
    <svelte:component 
      this={fallback} 
      {error} 
      {errorId}
      onRetry={handleRetry}
      onReset={resetError}
      {retrying}
    />
  {:else}
    <!-- デフォルトエラー表示 -->
    <div class="error-boundary {errorLevelClass}" role="alert" aria-live="assertive">
      <div class="error-content">
        <!-- エラーアイコン -->
        <div class="error-icon">
          {#if level === ERROR_LEVELS.CRITICAL}
            <svg width="48" height="48" viewBox="0 0 20 20" fill="currentColor" class="error-icon-critical">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7 4a1 1 0 11-2 0 1 1 0 012 0zm-1-9a1 1 0 00-1 1v4a1 1 0 102 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
            </svg>
          {:else if level === ERROR_LEVELS.HIGH}
            <svg width="48" height="48" viewBox="0 0 20 20" fill="currentColor" class="error-icon-high">
              <path fill-rule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clip-rule="evenodd" />
            </svg>
          {:else}
            <svg width="48" height="48" viewBox="0 0 20 20" fill="currentColor" class="error-icon-medium">
              <path fill-rule="evenodd" d="M18 10a8 8 0 11-16 0 8 8 0 0116 0zm-7-4a1 1 0 11-2 0 1 1 0 012 0zM9 9a1 1 0 000 2v3a1 1 0 001 1h1a1 1 0 100-2v-3a1 1 0 00-1-1H9z" clip-rule="evenodd" />
            </svg>
          {/if}
        </div>

        <!-- エラータイトル -->
        <h2 class="error-title">{errorTitle}</h2>

        <!-- エラーメッセージ -->
        <p class="error-message">{displayErrorMessage}</p>

        <!-- エラー詳細（高レベルエラーの場合のみ） -->
        {#if showErrorDetails}
          <details class="error-details">
            <summary class="error-details-summary">詳細情報</summary>
            <div class="error-details-content">
              <p><strong>エラータイプ:</strong> {error.type}</p>
              <p><strong>発生時刻:</strong> {new Date(error.timestamp).toLocaleString()}</p>
              {#if error.details}
                <p><strong>詳細:</strong></p>
                <pre class="error-details-json">{JSON.stringify(error.details, null, 2)}</pre>
              {/if}
            </div>
          </details>
        {/if}

        <!-- アクションボタン -->
        <div class="error-actions">
          {#if showRetry}
            <Button
              variant="primary"
              size="medium"
              disabled={retrying}
              on:click={handleRetry}
              class="retry-button"
            >
              {#if retrying}
                <LoadingSpinner size="small" />
                再試行中...
              {:else}
                {retryText}
              {/if}
            </Button>
          {/if}

          <Button
            variant="secondary"
            size="medium"
            on:click={resetError}
            class="reset-button"
          >
            閉じる
          </Button>
        </div>
      </div>
    </div>
  {/if}
{:else}
  <!-- 正常時はスロットコンテンツを表示 -->
  <slot />
{/if}

<style>
  .error-boundary {
    display: flex;
    justify-content: center;
    align-items: center;
    min-height: 200px;
    padding: 2rem;
    border-radius: 8px;
    border: 1px solid;
    background-color: #fff;
  }

  .error-level-low {
    border-color: #d1d5db;
    background-color: #f9fafb;
  }

  .error-level-medium {
    border-color: #fbbf24;
    background-color: #fffbeb;
  }

  .error-level-high {
    border-color: #f87171;
    background-color: #fef2f2;
  }

  .error-level-critical {
    border-color: #dc2626;
    background-color: #fef2f2;
    box-shadow: 0 0 0 1px #dc2626;
  }

  .error-content {
    text-align: center;
    max-width: 500px;
  }

  .error-icon {
    margin-bottom: 1rem;
    display: flex;
    justify-content: center;
  }

  .error-icon-critical {
    color: #dc2626;
  }

  .error-icon-high {
    color: #f59e0b;
  }

  .error-icon-medium {
    color: #6b7280;
  }

  .error-title {
    font-size: 1.5rem;
    font-weight: 600;
    margin-bottom: 0.5rem;
    color: #111827;
  }

  .error-message {
    font-size: 1rem;
    color: #6b7280;
    margin-bottom: 1.5rem;
    line-height: 1.5;
  }

  .error-details {
    margin-bottom: 1.5rem;
    text-align: left;
    border: 1px solid #e5e7eb;
    border-radius: 4px;
    overflow: hidden;
  }

  .error-details-summary {
    padding: 0.75rem;
    background-color: #f3f4f6;
    cursor: pointer;
    font-weight: 500;
    border-bottom: 1px solid #e5e7eb;
  }

  .error-details-summary:hover {
    background-color: #e5e7eb;
  }

  .error-details-content {
    padding: 1rem;
    font-size: 0.875rem;
  }

  .error-details-json {
    background-color: #f3f4f6;
    padding: 0.5rem;
    border-radius: 4px;
    font-family: 'Courier New', monospace;
    font-size: 0.75rem;
    overflow-x: auto;
    white-space: pre-wrap;
    word-break: break-all;
  }

  .error-actions {
    display: flex;
    gap: 1rem;
    justify-content: center;
    flex-wrap: wrap;
  }

  :global(.retry-button) {
    min-width: 120px;
  }

  :global(.reset-button) {
    min-width: 80px;
  }

  /* レスポンシブデザイン */
  @media (max-width: 768px) {
    .error-boundary {
      padding: 1rem;
      min-height: 150px;
    }

    .error-title {
      font-size: 1.25rem;
    }

    .error-message {
      font-size: 0.875rem;
    }

    .error-actions {
      flex-direction: column;
      align-items: center;
    }

    :global(.retry-button),
    :global(.reset-button) {
      width: 100%;
      max-width: 200px;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .error-boundary {
      background-color: #1f2937;
      border-color: #374151;
    }

    .error-level-low {
      background-color: #111827;
      border-color: #374151;
    }

    .error-level-medium {
      background-color: #1f2937;
      border-color: #d97706;
    }

    .error-level-high {
      background-color: #1f2937;
      border-color: #ef4444;
    }

    .error-level-critical {
      background-color: #1f2937;
      border-color: #dc2626;
    }

    .error-title {
      color: #f9fafb;
    }

    .error-message {
      color: #d1d5db;
    }

    .error-details {
      border-color: #374151;
    }

    .error-details-summary {
      background-color: #374151;
      border-bottom-color: #4b5563;
    }

    .error-details-summary:hover {
      background-color: #4b5563;
    }

    .error-details-json {
      background-color: #374151;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .error-boundary {
      transition: none;
    }
  }

  /* ハイコントラストモード */
  @media (prefers-contrast: high) {
    .error-boundary {
      border-width: 2px;
    }

    .error-title {
      font-weight: 700;
    }
  }
</style>
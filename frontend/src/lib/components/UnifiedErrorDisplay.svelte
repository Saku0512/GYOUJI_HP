<script>
  // UnifiedErrorDisplay コンポーネント - 統一エラー表示
  import { createEventDispatcher, onMount } from 'svelte';
  import { slide, fade } from 'svelte/transition';
  import ValidationMessage from './ValidationMessage.svelte';
  import { defaultErrorHandler, localizeFieldName } from '../utils/error-response-handler.js';
  
  // Props
  export let errors = [];              // エラー配列またはエラーレスポンス
  export let language = 'ja';          // 表示言語
  export let showSuggestions = true;   // 提案を表示するか
  export let showContext = false;      // コンテキスト情報を表示するか
  export let groupByField = true;      // フィールド別にグループ化するか
  export let maxErrors = 10;           // 最大表示エラー数
  export let dismissible = true;       // 個別エラーを閉じられるか
  export let showSummary = true;       // サマリーを表示するか
  export let animate = true;           // アニメーションを有効にするか
  export let size = 'medium';          // 'small', 'medium', 'large'
  export let variant = 'default';     // 'default', 'compact', 'detailed'
  
  const dispatch = createEventDispatcher();
  
  let processedErrors = [];
  let errorSummary = '';
  let consolidatedSuggestions = [];
  let dismissedErrors = new Set();
  
  // エラーの処理
  $: processErrors(errors);
  
  function processErrors(rawErrors) {
    if (!rawErrors) {
      processedErrors = [];
      errorSummary = '';
      consolidatedSuggestions = [];
      return;
    }
    
    let errorList = [];
    
    // エラー形式の正規化
    if (Array.isArray(rawErrors)) {
      errorList = rawErrors;
    } else if (rawErrors.errors && Array.isArray(rawErrors.errors)) {
      errorList = rawErrors.errors;
      errorSummary = rawErrors.summary || '';
      consolidatedSuggestions = rawErrors.suggestions || [];
    } else if (typeof rawErrors === 'object') {
      // 単一エラーオブジェクト
      errorList = [rawErrors];
    }
    
    // エラーの処理とフィルタリング
    processedErrors = errorList
      .filter(error => !dismissedErrors.has(getErrorId(error)))
      .slice(0, maxErrors)
      .map(error => ({
        ...error,
        id: getErrorId(error),
        localizedFieldName: error.field ? localizeFieldName(error.field, language) : null,
        displayMessage: error.localizedMessage || error.userMessage || error.message,
        severity: error.severity || 'error',
        suggestions: error.suggestions || []
      }));
    
    // フィールド別グループ化
    if (groupByField) {
      processedErrors = groupErrorsByField(processedErrors);
    }
    
    // サマリーの生成（提供されていない場合）
    if (!errorSummary && processedErrors.length > 0) {
      errorSummary = generateErrorSummary(processedErrors);
    }
    
    // 提案の統合（提供されていない場合）
    if (consolidatedSuggestions.length === 0) {
      consolidatedSuggestions = consolidateSuggestions(processedErrors);
    }
  }
  
  function getErrorId(error) {
    return `${error.field || 'general'}_${error.code || 'unknown'}_${error.message || ''}`;
  }
  
  function groupErrorsByField(errors) {
    const grouped = {};
    const ungrouped = [];
    
    errors.forEach(error => {
      if (error.field) {
        if (!grouped[error.field]) {
          grouped[error.field] = [];
        }
        grouped[error.field].push(error);
      } else {
        ungrouped.push(error);
      }
    });
    
    // グループ化されたエラーを配列に変換
    const result = [];
    
    // フィールドエラーを追加
    Object.entries(grouped).forEach(([field, fieldErrors]) => {
      result.push({
        isGroup: true,
        field,
        localizedFieldName: localizeFieldName(field, language),
        errors: fieldErrors,
        severity: fieldErrors.some(e => e.severity === 'error') ? 'error' : 'warning'
      });
    });
    
    // 一般エラーを追加
    ungrouped.forEach(error => {
      result.push(error);
    });
    
    return result;
  }
  
  function generateErrorSummary(errors) {
    if (errors.length === 0) return '';
    if (errors.length === 1) return errors[0].displayMessage;
    
    const fieldErrors = errors.filter(e => e.field || e.isGroup);
    const generalErrors = errors.filter(e => !e.field && !e.isGroup);
    
    if (fieldErrors.length > 0 && generalErrors.length === 0) {
      return language === 'ja' ? '入力内容に問題があります' : 'Input validation failed';
    }
    
    if (generalErrors.length > 0) {
      return generalErrors[0].displayMessage;
    }
    
    return language === 'ja' ? 
      `${errors.length}個のエラーが発生しました` : 
      `${errors.length} errors occurred`;
  }
  
  function consolidateSuggestions(errors) {
    const allSuggestions = errors.flatMap(error => 
      error.isGroup ? 
        error.errors.flatMap(e => e.suggestions) : 
        error.suggestions || []
    );
    return [...new Set(allSuggestions)];
  }
  
  function dismissError(errorId) {
    dismissedErrors.add(errorId);
    dismissedErrors = dismissedErrors; // リアクティブ更新をトリガー
    
    dispatch('dismiss', { errorId });
    
    // エラーを再処理
    processErrors(errors);
  }
  
  function dismissAllErrors() {
    processedErrors.forEach(error => {
      if (error.isGroup) {
        error.errors.forEach(e => dismissedErrors.add(e.id));
      } else {
        dismissedErrors.add(error.id);
      }
    });
    dismissedErrors = dismissedErrors;
    
    dispatch('dismissAll');
    
    // エラーを再処理
    processErrors(errors);
  }
  
  function handleErrorClick(error) {
    dispatch('errorClick', { error });
  }
  
  function handleSuggestionClick(suggestion) {
    dispatch('suggestionClick', { suggestion });
  }
  
  // バリアント別のクラス
  $: containerClasses = [
    'unified-error-display',
    `unified-error-display-${variant}`,
    `unified-error-display-${size}`,
    animate ? 'unified-error-display-animate' : ''
  ].filter(Boolean).join(' ');
  
  // 表示判定
  $: hasVisibleErrors = processedErrors.length > 0;
  $: hasVisibleSuggestions = showSuggestions && consolidatedSuggestions.length > 0;
</script>

{#if hasVisibleErrors}
  <div 
    class={containerClasses}
    role="alert"
    aria-live="polite"
    transition:slide={{ duration: animate ? 300 : 0 }}
  >
    <!-- エラーサマリー -->
    {#if showSummary && errorSummary}
      <div class="error-summary">
        <div class="error-summary-content">
          <span class="error-summary-icon" aria-hidden="true">⚠️</span>
          <span class="error-summary-text">{errorSummary}</span>
        </div>
        
        {#if dismissible && processedErrors.length > 1}
          <button 
            class="error-summary-dismiss-all"
            on:click={dismissAllErrors}
            aria-label="全てのエラーを閉じる"
            type="button"
          >
            全て閉じる
          </button>
        {/if}
      </div>
    {/if}
    
    <!-- エラー一覧 -->
    <div class="error-list">
      {#each processedErrors as error (error.id || error.field)}
        <div 
          class="error-item error-item-{error.severity}"
          transition:slide={{ duration: animate ? 200 : 0 }}
        >
          {#if error.isGroup}
            <!-- フィールドグループエラー -->
            <div class="error-group">
              <div class="error-group-header">
                <span class="error-group-field">{error.localizedFieldName}</span>
                <span class="error-group-count">({error.errors.length})</span>
              </div>
              
              <div class="error-group-items">
                {#each error.errors as fieldError}
                  <ValidationMessage
                    error={fieldError.displayMessage}
                    touched={true}
                    type={fieldError.severity}
                    {size}
                    {dismissible}
                    on:dismiss={() => dismissError(fieldError.id)}
                  />
                {/each}
              </div>
            </div>
          {:else}
            <!-- 単一エラー -->
            <ValidationMessage
              error={error.displayMessage}
              touched={true}
              type={error.severity}
              {size}
              {dismissible}
              on:dismiss={() => dismissError(error.id)}
            />
            
            <!-- コンテキスト情報 -->
            {#if showContext && error.context && variant === 'detailed'}
              <div class="error-context">
                <details>
                  <summary>詳細情報</summary>
                  <div class="error-context-content">
                    {#if error.context.httpStatus}
                      <div>ステータス: {error.context.httpStatus}</div>
                    {/if}
                    {#if error.context.requestId}
                      <div>リクエストID: {error.context.requestId}</div>
                    {/if}
                    {#if error.context.timestamp}
                      <div>発生時刻: {new Date(error.context.timestamp).toLocaleString()}</div>
                    {/if}
                  </div>
                </details>
              </div>
            {/if}
          {/if}
        </div>
      {/each}
    </div>
    
    <!-- 提案 -->
    {#if hasVisibleSuggestions && variant !== 'compact'}
      <div class="error-suggestions" transition:fade={{ duration: animate ? 200 : 0 }}>
        <div class="error-suggestions-header">
          <span class="error-suggestions-icon" aria-hidden="true">💡</span>
          <span class="error-suggestions-title">
            {language === 'ja' ? '解決のヒント' : 'Suggestions'}
          </span>
        </div>
        
        <ul class="error-suggestions-list">
          {#each consolidatedSuggestions as suggestion}
            <li class="error-suggestion-item">
              <button 
                class="error-suggestion-button"
                on:click={() => handleSuggestionClick(suggestion)}
                type="button"
              >
                {suggestion}
              </button>
            </li>
          {/each}
        </ul>
      </div>
    {/if}
  </div>
{/if}

<style>
  .unified-error-display {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    padding: 1rem;
    background-color: #fef2f2;
    border: 1px solid #fecaca;
    border-radius: 0.5rem;
    color: #dc2626;
  }
  
  .unified-error-display-animate {
    transition: all 0.2s ease-in-out;
  }
  
  /* サイズ */
  .unified-error-display-small {
    padding: 0.75rem;
    gap: 0.75rem;
    font-size: 0.875rem;
  }
  
  .unified-error-display-medium {
    padding: 1rem;
    gap: 1rem;
    font-size: 1rem;
  }
  
  .unified-error-display-large {
    padding: 1.25rem;
    gap: 1.25rem;
    font-size: 1.125rem;
  }
  
  /* バリアント */
  .unified-error-display-compact {
    padding: 0.75rem;
    gap: 0.5rem;
  }
  
  .unified-error-display-detailed {
    padding: 1.25rem;
    gap: 1.25rem;
  }
  
  /* エラーサマリー */
  .error-summary {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem;
    background-color: #fee2e2;
    border-radius: 0.375rem;
    border-left: 4px solid #dc2626;
  }
  
  .error-summary-content {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  
  .error-summary-icon {
    font-size: 1.25em;
  }
  
  .error-summary-text {
    font-weight: 500;
  }
  
  .error-summary-dismiss-all {
    background: none;
    border: 1px solid #dc2626;
    color: #dc2626;
    padding: 0.25rem 0.5rem;
    border-radius: 0.25rem;
    font-size: 0.875rem;
    cursor: pointer;
    transition: all 0.2s ease-in-out;
  }
  
  .error-summary-dismiss-all:hover {
    background-color: #dc2626;
    color: white;
  }
  
  /* エラー一覧 */
  .error-list {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }
  
  .error-item {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  
  /* エラーグループ */
  .error-group {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .error-group-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    font-weight: 500;
    color: #991b1b;
  }
  
  .error-group-field {
    font-size: 1.1em;
  }
  
  .error-group-count {
    font-size: 0.875em;
    opacity: 0.8;
  }
  
  .error-group-items {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin-left: 1rem;
  }
  
  /* コンテキスト情報 */
  .error-context {
    margin-top: 0.5rem;
    font-size: 0.875em;
    opacity: 0.8;
  }
  
  .error-context details {
    cursor: pointer;
  }
  
  .error-context summary {
    font-weight: 500;
    margin-bottom: 0.25rem;
  }
  
  .error-context-content {
    padding: 0.5rem;
    background-color: rgba(0, 0, 0, 0.05);
    border-radius: 0.25rem;
    font-family: monospace;
    font-size: 0.8em;
  }
  
  /* 提案 */
  .error-suggestions {
    padding: 0.75rem;
    background-color: #eff6ff;
    border: 1px solid #bfdbfe;
    border-radius: 0.375rem;
    color: #1e40af;
  }
  
  .error-suggestions-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 0.75rem;
    font-weight: 500;
  }
  
  .error-suggestions-icon {
    font-size: 1.25em;
  }
  
  .error-suggestions-list {
    list-style: none;
    padding: 0;
    margin: 0;
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }
  
  .error-suggestion-item {
    display: flex;
  }
  
  .error-suggestion-button {
    background: none;
    border: none;
    color: #2563eb;
    text-decoration: underline;
    cursor: pointer;
    text-align: left;
    padding: 0.25rem;
    border-radius: 0.25rem;
    transition: background-color 0.2s ease-in-out;
  }
  
  .error-suggestion-button:hover {
    background-color: rgba(37, 99, 235, 0.1);
  }
  
  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .unified-error-display {
      padding: 0.75rem;
      gap: 0.75rem;
    }
    
    .error-summary {
      flex-direction: column;
      align-items: flex-start;
      gap: 0.5rem;
    }
    
    .error-group-items {
      margin-left: 0.5rem;
    }
  }
  
  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .unified-error-display {
      background-color: #450a0a;
      border-color: #7f1d1d;
      color: #fca5a5;
    }
    
    .error-summary {
      background-color: #7f1d1d;
      border-left-color: #fca5a5;
    }
    
    .error-suggestions {
      background-color: #1e3a8a;
      border-color: #3730a3;
      color: #93c5fd;
    }
    
    .error-suggestion-button {
      color: #60a5fa;
    }
  }
  
  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .unified-error-display {
      border-width: 2px;
    }
    
    .error-summary {
      border-left-width: 6px;
    }
    
    .error-suggestions {
      border-width: 2px;
    }
  }
  
  /* 縮小モーション設定対応 */
  @media (prefers-reduced-motion: reduce) {
    .unified-error-display-animate {
      transition: none;
    }
    
    .error-summary-dismiss-all {
      transition: none;
    }
    
    .error-suggestion-button {
      transition: none;
    }
  }
</style>
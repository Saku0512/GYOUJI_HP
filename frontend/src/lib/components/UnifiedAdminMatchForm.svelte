<script>
  // UnifiedAdminMatchForm コンポーネント - 統一バリデーション使用版
  import { createEventDispatcher, onMount } from 'svelte';
  import { createMatchResultValidator, createRealtimeValidator, createSubmitHandler } from '../utils/unified-validation.js';
  import { createRealtimeValidator as createRTV } from '../utils/realtime-validation.js';
  import ValidatedInput from './ValidatedInput.svelte';
  import ValidationMessage from './ValidationMessage.svelte';
  import { csrfTokenManager, securityLogger } from '../utils/security.js';
  import { uiActions } from '../stores/ui.js';
  import { matchAPI } from '../api/matches.js';

  // Props
  export let match = {};
  export let onSubmit = () => {};
  export let disabled = false;

  // イベントディスパッチャー
  const dispatch = createEventDispatcher();

  // バリデーター設定
  let matchResultValidator;
  let realtimeValidator;
  let submitHandler;
  
  // フォーム状態
  let isSubmitting = false;
  let csrfToken = '';
  let winner = '';
  let formErrors = {};

  // バリデーターの初期化
  onMount(() => {
    // 試合結果バリデーターを作成
    matchResultValidator = createMatchResultValidator();
    
    // リアルタイムバリデーターを作成
    realtimeValidator = createRTV(matchResultValidator, {
      debounceMs: 300,
      validateOnChange: true,
      validateOnBlur: true
    });
    
    // 送信ハンドラーを作成
    submitHandler = createSubmitHandler(realtimeValidator, handleFormSubmit);
    
    // 初期値を設定
    realtimeValidator.setFieldValue('score1', match.score1 || '');
    realtimeValidator.setFieldValue('score2', match.score2 || '');
    realtimeValidator.setFieldValue('winner', determineWinner(match.score1 || '', match.score2 || ''));
    
    // CSRFトークンの取得
    if (typeof window !== 'undefined') {
      csrfToken = csrfTokenManager.getToken();
    }
  });

  // フォームデータの監視
  $: if (realtimeValidator) {
    realtimeValidator.formData.subscribe(formData => {
      winner = determineWinner(formData.score1 || '', formData.score2 || '');
      // 勝者フィールドを自動更新
      if (winner && winner !== formData.winner) {
        realtimeValidator.setFieldValue('winner', winner);
      }
    });
  }

  // フォームエラーの監視
  $: if (realtimeValidator) {
    realtimeValidator.errors.subscribe(errors => {
      formErrors = errors;
    });
  }

  // フォームの有効性
  $: isFormValid = realtimeValidator ? 
    Object.keys(formErrors).length === 0 && 
    winner !== '' : false;

  /**
   * 勝者を決定する
   */
  function determineWinner(s1, s2) {
    const num1 = Number(s1);
    const num2 = Number(s2);
    
    if (isNaN(num1) || isNaN(num2) || s1 === '' || s2 === '') {
      return '';
    }
    
    if (num1 > num2) {
      return match.team1 || 'Team A';
    } else if (num2 > num1) {
      return match.team2 || 'Team B';
    } else {
      return ''; // 引き分けは無効
    }
  }

  /**
   * フォーム送信処理
   */
  async function handleFormSubmit(sanitizedData) {
    isSubmitting = true;
    uiActions.setLoading(true);

    try {
      // 最終的な試合結果データ
      const matchResult = {
        score1: sanitizedData.score1,
        score2: sanitizedData.score2,
        winner: sanitizedData.winner,
        csrfToken: csrfToken
      };

      // セキュリティログに記録
      securityLogger.logEvent('MATCH_RESULT_SUBMISSION', {
        matchId: match.id,
        team1: match.team1,
        team2: match.team2,
        score1: matchResult.score1,
        score2: matchResult.score2,
        winner: matchResult.winner,
        timestamp: new Date().toISOString()
      });

      // APIを使用して試合結果を更新
      if (match.id) {
        const response = await matchAPI.updateMatch(match.id, matchResult);
        
        if (response.success) {
          uiActions.showNotification('試合結果を更新しました', 'success');
          
          // 成功ログ
          securityLogger.logEvent('MATCH_RESULT_UPDATE_SUCCESS', {
            matchId: match.id,
            timestamp: new Date().toISOString()
          });
          
          dispatch('success', { match: match.id, result: matchResult });
          
          // 親コンポーネントのonSubmitも呼び出す
          if (typeof onSubmit === 'function') {
            onSubmit(matchResult);
          }
        } else {
          throw new Error(response.message || '試合結果の更新に失敗しました');
        }
      } else {
        // match.idがない場合は親コンポーネントのonSubmitのみ呼び出す
        if (typeof onSubmit === 'function') {
          await onSubmit(matchResult);
        }
        uiActions.showNotification('試合結果を保存しました', 'success');
        dispatch('success', { result: matchResult });
      }

    } catch (error) {
      console.error('Submit error:', error);
      
      // エラーログ
      securityLogger.logEvent('MATCH_RESULT_SUBMISSION_ERROR', {
        matchId: match.id,
        error: error.message,
        timestamp: new Date().toISOString()
      });
      
      let errorMessage = '試合結果の保存に失敗しました';
      
      // エラーの種類に応じてメッセージを調整
      if (error.message && error.message.includes('CSRF')) {
        errorMessage = 'セキュリティトークンが無効です。ページを再読み込みしてください。';
        // CSRFトークンを更新
        csrfToken = csrfTokenManager.refreshToken();
      } else if (error.message) {
        errorMessage = error.message;
      }
      
      uiActions.showNotification(errorMessage, 'error');
      dispatch('error', { error: error.message });
      
      throw error; // submitHandlerに伝播
    } finally {
      isSubmitting = false;
      uiActions.setLoading(false);
    }
  }

  /**
   * フォームリセット
   */
  function resetForm() {
    if (realtimeValidator) {
      realtimeValidator.reset();
      realtimeValidator.setFieldValue('score1', match.score1 || '');
      realtimeValidator.setFieldValue('score2', match.score2 || '');
      realtimeValidator.setFieldValue('winner', '');
    }
    winner = '';
  }

  /**
   * キャンセル処理
   */
  function handleCancel() {
    resetForm();
    dispatch('cancel');
  }
</script>

<form class="unified-admin-match-form" on:submit={submitHandler}>
  <h3>試合結果入力</h3>
  
  <div class="match-info">
    <span class="team">{match.team1 || 'Team A'}</span>
    <span class="vs">vs</span>
    <span class="team">{match.team2 || 'Team B'}</span>
  </div>

  {#if match.round}
    <div class="round-info">
      <span class="round-label">ラウンド:</span>
      <span class="round-value">{match.round}</span>
    </div>
  {/if}

  <div class="score-inputs">
    <div class="score-group">
      <ValidatedInput
        {realtimeValidator}
        fieldName="score1"
        type="number"
        label="{match.team1 || 'Team A'} スコア"
        placeholder="0"
        min="0"
        max="999"
        required
        disabled={disabled || isSubmitting}
        fullWidth
        data-testid="score1"
        autocomplete="off"
      />
    </div>
    
    <div class="score-group">
      <ValidatedInput
        {realtimeValidator}
        fieldName="score2"
        type="number"
        label="{match.team2 || 'Team B'} スコア"
        placeholder="0"
        min="0"
        max="999"
        required
        disabled={disabled || isSubmitting}
        fullWidth
        data-testid="score2"
        autocomplete="off"
      />
    </div>
  </div>

  <!-- 勝者の自動表示 -->
  {#if winner && isFormValid}
    <div class="winner-preview">
      <span class="winner-label">勝者:</span>
      <span class="winner-value">{winner}</span>
    </div>
  {/if}

  <!-- 引き分けエラーの表示 -->
  {#if realtimeValidator}
    {#await realtimeValidator.formData.subscribe(formData => {
      const s1 = Number(formData.score1 || 0);
      const s2 = Number(formData.score2 || 0);
      if (formData.score1 !== '' && formData.score2 !== '' && s1 === s2) {
        return true;
      }
      return false;
    }) then isDraw}
      {#if isDraw}
        <ValidationMessage
          error="引き分けは許可されていません"
          touched={true}
          type="error"
          fullWidth
        />
      {/if}
    {/await}
  {/if}

  <!-- CSRFトークン -->
  <input type="hidden" name="csrf_token" value={csrfToken} />

  <div class="form-actions">
    <button 
      type="button" 
      class="cancel-button"
      on:click={handleCancel}
      disabled={isSubmitting}
    >
      キャンセル
    </button>
    
    <button 
      type="submit" 
      class="submit-button"
      disabled={!isFormValid || disabled || isSubmitting}
      data-testid="submit-result"
    >
      {#if isSubmitting}
        <span class="loading-spinner"></span>
        保存中...
      {:else}
        結果を保存
      {/if}
    </button>
  </div>
</form>

<style>
  .unified-admin-match-form {
    border: 1px solid #ddd;
    border-radius: 8px;
    padding: 1.5rem;
    background-color: white;
    max-width: 450px;
    margin: 0 auto;
    box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
  }

  h3 {
    text-align: center;
    margin-bottom: 1rem;
    color: #333;
    font-size: 1.25rem;
  }

  .match-info {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 1rem;
    padding: 0.75rem;
    background-color: #f8f9fa;
    border-radius: 6px;
    border-left: 4px solid #007bff;
  }

  .team {
    font-weight: bold;
    color: #333;
    font-size: 1rem;
  }

  .vs {
    color: #666;
    font-weight: 500;
  }

  .round-info {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    padding: 0.5rem;
    background-color: #e9ecef;
    border-radius: 4px;
  }

  .round-label {
    font-weight: 500;
    color: #495057;
  }

  .round-value {
    font-weight: bold;
    color: #007bff;
  }

  .score-inputs {
    display: flex;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .score-group {
    flex: 1;
  }

  .winner-preview {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 0.5rem;
    margin-bottom: 1.5rem;
    padding: 0.75rem;
    background-color: #d4edda;
    border: 1px solid #c3e6cb;
    border-radius: 4px;
  }

  .winner-label {
    font-weight: 500;
    color: #155724;
  }

  .winner-value {
    font-weight: bold;
    color: #155724;
  }

  .form-actions {
    display: flex;
    gap: 1rem;
    margin-top: 1.5rem;
  }

  .cancel-button {
    flex: 1;
    padding: 0.75rem;
    background-color: #6c757d;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
  }

  .cancel-button:hover:not(:disabled) {
    background-color: #5a6268;
  }

  .submit-button {
    flex: 2;
    padding: 0.75rem;
    background-color: #28a745;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
  }

  .submit-button:hover:not(:disabled) {
    background-color: #218838;
  }

  .submit-button:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
  }

  .loading-spinner {
    width: 16px;
    height: 16px;
    border: 2px solid transparent;
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% { transform: rotate(0deg); }
    100% { transform: rotate(360deg); }
  }

  /* レスポンシブデザイン */
  @media (max-width: 480px) {
    .unified-admin-match-form {
      padding: 1rem;
      margin: 0 1rem;
    }

    .score-inputs {
      flex-direction: column;
      gap: 1rem;
    }

    .form-actions {
      flex-direction: column;
    }

    .match-info {
      flex-direction: column;
      gap: 0.5rem;
      text-align: center;
    }

    .vs {
      order: -1;
      font-size: 0.9rem;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .loading-spinner {
      animation: none;
    }
    
    .submit-button, .cancel-button {
      transition: none;
    }
  }

  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .unified-admin-match-form {
      border: 2px solid #000;
    }
    
    .submit-button {
      border: 2px solid #000;
    }
    
    .cancel-button {
      border: 2px solid #000;
    }
  }
</style>
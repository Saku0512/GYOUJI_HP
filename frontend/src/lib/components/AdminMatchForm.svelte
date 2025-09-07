<script>
  // AdminMatchForm コンポーネント - 試合結果入力フォーム
  import { createEventDispatcher } from 'svelte';
  import { validateMatchResult, debounce } from '../utils/validation.js';
  import { csrfTokenManager, securityLogger, enforceInputLimits } from '../utils/security.js';
  import { uiActions } from '../stores/ui.js';
  import { matchAPI } from '../api/matches.js';

  // Props
  export let match = {};
  export let onSubmit = () => {};
  export let disabled = false;

  // イベントディスパッチャー
  const dispatch = createEventDispatcher();

  // フォーム状態
  let score1 = match.score1 || '';
  let score2 = match.score2 || '';
  let isSubmitting = false;
  let errors = {};
  let touched = { score1: false, score2: false };
  let csrfToken = '';
  let validationInProgress = false;

  // CSRFトークンの初期化
  $: if (typeof window !== 'undefined') {
    csrfToken = csrfTokenManager.getToken();
  }

  // デバウンス付きリアルタイム検証
  const debouncedValidation = debounce(() => {
    if (touched.score1 || touched.score2) {
      validationInProgress = true;
      
      // 入力値の長さ制限を適用
      const limitedScore1 = enforceInputLimits(score1.toString(), 10);
      const limitedScore2 = enforceInputLimits(score2.toString(), 10);
      
      // 制限が適用された場合は値を更新
      if (limitedScore1 !== score1.toString()) {
        score1 = limitedScore1;
        securityLogger.logEvent('INPUT_LENGTH_LIMITED', {
          field: 'score1',
          originalLength: score1.toString().length,
          limitedLength: limitedScore1.length
        });
      }
      
      if (limitedScore2 !== score2.toString()) {
        score2 = limitedScore2;
        securityLogger.logEvent('INPUT_LENGTH_LIMITED', {
          field: 'score2',
          originalLength: score2.toString().length,
          limitedLength: limitedScore2.length
        });
      }
      
      const validation = validateMatchResult(score1, score2, match.team1, match.team2);
      errors = validation.errors;
      
      // セキュリティ違反があった場合はログに記録
      if (!validation.isValid && (validation.errors.score1 || validation.errors.score2)) {
        const hasSecurityIssue = Object.values(validation.errors).some(error => 
          error.includes('不正') || error.includes('スクリプト') || error.includes('文字')
        );
        
        if (hasSecurityIssue) {
          securityLogger.logEvent('FORM_SECURITY_VIOLATION', {
            component: 'AdminMatchForm',
            matchId: match.id,
            errors: validation.errors,
            values: { score1, score2 }
          });
        }
      }
      
      validationInProgress = false;
    }
  }, 300);

  // リアクティブ検証の実行
  $: {
    if (touched.score1 || touched.score2) {
      debouncedValidation();
    }
  }

  // フォームの有効性チェック
  $: isFormValid = Object.keys(errors).length === 0 && score1 !== '' && score2 !== '';

  // 勝者の決定
  $: winner = determineWinner(score1, score2);

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
      return '引き分け';
    }
  }

  /**
   * 入力フィールドのフォーカス処理
   */
  function handleFieldTouch(field) {
    touched[field] = true;
  }

  /**
   * 入力値の変更処理（セキュリティチェック付き）
   */
  function handleInputChange(field, value) {
    // 入力値の長さ制限
    const limitedValue = enforceInputLimits(value.toString(), 10);
    
    if (limitedValue !== value.toString()) {
      securityLogger.logEvent('INPUT_LENGTH_EXCEEDED', {
        component: 'AdminMatchForm',
        field,
        originalLength: value.toString().length,
        limitedLength: limitedValue.length
      });
    }
    
    // 値を更新
    if (field === 'score1') {
      score1 = limitedValue;
    } else if (field === 'score2') {
      score2 = limitedValue;
    }
    
    // タッチ状態を更新
    handleFieldTouch(field);
  }

  /**
   * フォーム送信処理
   */
  async function handleSubmit() {
    // 全フィールドをタッチ済みにする
    touched = { score1: true, score2: true };
    
    // 最終検証（サニタイゼーション付き）
    const validation = validateMatchResult(score1, score2, match.team1, match.team2);
    if (!validation.isValid) {
      errors = validation.errors;
      uiActions.showNotification('入力内容を確認してください', 'error');
      
      // セキュリティ違反の可能性をログに記録
      securityLogger.logEvent('FORM_SUBMISSION_BLOCKED', {
        component: 'AdminMatchForm',
        matchId: match.id,
        errors: validation.errors,
        timestamp: new Date().toISOString()
      });
      
      return;
    }

    isSubmitting = true;
    uiActions.setLoading(true);

    try {
      // サニタイズされたデータを使用
      const sanitizedResult = {
        score1: validation.sanitizedData.score1,
        score2: validation.sanitizedData.score2,
        winner: winner,
        csrfToken: csrfToken
      };

      // セキュリティログに記録
      securityLogger.logEvent('MATCH_RESULT_SUBMISSION', {
        matchId: match.id,
        team1: match.team1,
        team2: match.team2,
        score1: sanitizedResult.score1,
        score2: sanitizedResult.score2,
        winner: sanitizedResult.winner,
        timestamp: new Date().toISOString()
      });

      // APIを使用して試合結果を更新
      if (match.id) {
        const response = await matchAPI.updateMatch(match.id, sanitizedResult);
        
        if (response.success) {
          uiActions.showNotification('試合結果を更新しました', 'success');
          
          // 成功ログ
          securityLogger.logEvent('MATCH_RESULT_UPDATE_SUCCESS', {
            matchId: match.id,
            timestamp: new Date().toISOString()
          });
          
          dispatch('success', { match: match.id, result: sanitizedResult });
          
          // 親コンポーネントのonSubmitも呼び出す
          if (typeof onSubmit === 'function') {
            onSubmit(sanitizedResult);
          }
        } else {
          throw new Error(response.message || '試合結果の更新に失敗しました');
        }
      } else {
        // match.idがない場合は親コンポーネントのonSubmitのみ呼び出す
        if (typeof onSubmit === 'function') {
          await onSubmit(sanitizedResult);
        }
        uiActions.showNotification('試合結果を保存しました', 'success');
        dispatch('success', { result: sanitizedResult });
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
    } finally {
      isSubmitting = false;
      uiActions.setLoading(false);
    }
  }

  /**
   * フォームリセット
   */
  function resetForm() {
    score1 = match.score1 || '';
    score2 = match.score2 || '';
    errors = {};
    touched = { score1: false, score2: false };
  }

  /**
   * キャンセル処理
   */
  function handleCancel() {
    resetForm();
    dispatch('cancel');
  }
</script>

<form class="admin-match-form" on:submit|preventDefault={handleSubmit}>
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
      <label for="score1">{match.team1 || 'Team A'} スコア</label>
      <input 
        type="number" 
        id="score1" 
        value={score1}
        on:input={(e) => handleInputChange('score1', e.target.value)}
        min="0" 
        max="999"
        required 
        disabled={disabled || isSubmitting || validationInProgress}
        class:error={errors.score1}
        on:blur={() => handleFieldTouch('score1')}
        data-testid="score1"
        autocomplete="off"
      />
      {#if errors.score1}
        <span class="error-message">{errors.score1}</span>
      {/if}
    </div>
    
    <div class="score-group">
      <label for="score2">{match.team2 || 'Team B'} スコア</label>
      <input 
        type="number" 
        id="score2" 
        value={score2}
        on:input={(e) => handleInputChange('score2', e.target.value)}
        min="0" 
        max="999"
        required 
        disabled={disabled || isSubmitting || validationInProgress}
        class:error={errors.score2}
        on:blur={() => handleFieldTouch('score2')}
        data-testid="score2"
        autocomplete="off"
      />
      {#if errors.score2}
        <span class="error-message">{errors.score2}</span>
      {/if}
    </div>
  </div>

  {#if winner && isFormValid}
    <div class="winner-preview">
      <span class="winner-label">勝者:</span>
      <span class="winner-value">{winner}</span>
    </div>
  {/if}

  <!-- CSRFトークン -->
  <input type="hidden" name="csrf_token" value={csrfToken} />

  <div class="form-actions">
    <button 
      type="button" 
      class="cancel-button"
      on:click={handleCancel}
      disabled={isSubmitting || validationInProgress}
    >
      キャンセル
    </button>
    
    <button 
      type="submit" 
      class="submit-button"
      disabled={!isFormValid || disabled || isSubmitting || validationInProgress}
      data-testid="submit-result"
    >
      {#if isSubmitting}
        <span class="loading-spinner"></span>
        保存中...
      {:else if validationInProgress}
        <span class="loading-spinner"></span>
        検証中...
      {:else}
        結果を保存
      {/if}
    </button>
  </div>
</form>

<style>
  .admin-match-form {
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

  label {
    display: block;
    margin-bottom: 0.5rem;
    color: #555;
    font-size: 0.9rem;
    font-weight: 500;
  }

  input {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
    transition: border-color 0.2s, box-shadow 0.2s;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 2px rgba(0, 123, 255, 0.25);
  }

  input.error {
    border-color: #dc3545;
    box-shadow: 0 0 0 2px rgba(220, 53, 69, 0.25);
  }

  input:disabled {
    background-color: #f8f9fa;
    color: #6c757d;
    cursor: not-allowed;
  }

  .error-message {
    display: block;
    margin-top: 0.25rem;
    color: #dc3545;
    font-size: 0.8rem;
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
    .admin-match-form {
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
    
    input {
      transition: none;
    }
    
    .submit-button, .cancel-button {
      transition: none;
    }
  }

  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .admin-match-form {
      border: 2px solid #000;
    }
    
    input {
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

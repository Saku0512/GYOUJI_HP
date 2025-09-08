<script>
  // UnifiedLoginForm コンポーネント - 統一バリデーション使用版
  import { createEventDispatcher, onMount } from 'svelte';
  import { createLoginValidator } from '../utils/unified-validation.js';
  import { createRealtimeValidator, createSubmitHandler } from '../utils/realtime-validation.js';
  import ValidatedInput from './ValidatedInput.svelte';
  import ValidationMessage from './ValidationMessage.svelte';
  import { authAPI } from '../api/auth.js';
  import { uiActions } from '../stores/ui.js';
  import { csrfTokenManager, securityLogger } from '../utils/security.js';

  // Props
  export let disabled = false;
  export let showRememberMe = true;
  export let redirectUrl = '/';

  // イベントディスパッチャー
  const dispatch = createEventDispatcher();

  // バリデーター設定
  let loginValidator;
  let realtimeValidator;
  let submitHandler;
  
  // フォーム状態
  let isSubmitting = false;
  let rememberMe = false;
  let csrfToken = '';
  let formErrors = {};
  let generalError = '';

  // バリデーターの初期化
  onMount(() => {
    // ログインバリデーターを作成
    loginValidator = createLoginValidator();
    
    // リアルタイムバリデーターを作成
    realtimeValidator = createRealtimeValidator(loginValidator, {
      debounceMs: 300,
      validateOnChange: true,
      validateOnBlur: true
    });
    
    // 送信ハンドラーを作成
    submitHandler = createSubmitHandler(realtimeValidator, handleFormSubmit);
    
    // CSRFトークンの取得
    if (typeof window !== 'undefined') {
      csrfToken = csrfTokenManager.getToken();
    }
  });

  // フォームエラーの監視
  $: if (realtimeValidator) {
    realtimeValidator.errors.subscribe(errors => {
      formErrors = errors;
    });
  }

  // フォームの有効性
  $: isFormValid = realtimeValidator ? 
    Object.keys(formErrors).length === 0 : false;

  /**
   * フォーム送信処理
   */
  async function handleFormSubmit(sanitizedData) {
    isSubmitting = true;
    generalError = '';
    uiActions.setLoading(true);

    try {
      // ログインデータ
      const loginData = {
        username: sanitizedData.username,
        password: sanitizedData.password,
        remember_me: rememberMe,
        csrfToken: csrfToken
      };

      // セキュリティログに記録（パスワードは除く）
      securityLogger.logEvent('LOGIN_ATTEMPT', {
        username: sanitizedData.username,
        rememberMe: rememberMe,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent
      });

      // APIを使用してログイン
      const response = await authAPI.login(loginData);
      
      if (response.success) {
        uiActions.showNotification('ログインしました', 'success');
        
        // 成功ログ
        securityLogger.logEvent('LOGIN_SUCCESS', {
          username: sanitizedData.username,
          timestamp: new Date().toISOString()
        });
        
        dispatch('success', { 
          user: response.data.user,
          token: response.data.token,
          redirectUrl 
        });
        
      } else {
        throw new Error(response.message || 'ログインに失敗しました');
      }

    } catch (error) {
      console.error('Login error:', error);
      
      // エラーログ
      securityLogger.logEvent('LOGIN_ERROR', {
        username: sanitizedData.username,
        error: error.message,
        timestamp: new Date().toISOString()
      });
      
      let errorMessage = 'ログインに失敗しました';
      
      // エラーの種類に応じてメッセージを調整
      if (error.message) {
        if (error.message.includes('認証情報')) {
          errorMessage = 'ユーザー名またはパスワードが正しくありません';
        } else if (error.message.includes('CSRF')) {
          errorMessage = 'セキュリティトークンが無効です。ページを再読み込みしてください。';
          // CSRFトークンを更新
          csrfToken = csrfTokenManager.refreshToken();
        } else if (error.message.includes('アカウント')) {
          errorMessage = 'アカウントがロックされています。しばらく時間をおいてから再試行してください。';
        } else {
          errorMessage = error.message;
        }
      }
      
      generalError = errorMessage;
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
    }
    rememberMe = false;
    generalError = '';
  }

  /**
   * パスワードリセットリンク
   */
  function handleForgotPassword() {
    dispatch('forgotPassword');
  }

  /**
   * 新規登録リンク
   */
  function handleSignUp() {
    dispatch('signUp');
  }
</script>

<form class="unified-login-form" on:submit={submitHandler}>
  <div class="form-header">
    <h2>ログイン</h2>
    <p class="form-description">アカウントにログインしてください</p>
  </div>

  <!-- 全般的なエラーメッセージ -->
  {#if generalError}
    <ValidationMessage
      error={generalError}
      touched={true}
      type="error"
      fullWidth
      dismissible
      on:dismiss={() => generalError = ''}
    />
  {/if}

  <div class="form-fields">
    <ValidatedInput
      {realtimeValidator}
      fieldName="username"
      type="text"
      label="ユーザー名"
      placeholder="ユーザー名を入力"
      required
      disabled={disabled || isSubmitting}
      fullWidth
      autocomplete="username"
      data-testid="username"
    />

    <ValidatedInput
      {realtimeValidator}
      fieldName="password"
      type="password"
      label="パスワード"
      placeholder="パスワードを入力"
      required
      disabled={disabled || isSubmitting}
      fullWidth
      autocomplete="current-password"
      data-testid="password"
    />

    {#if showRememberMe}
      <div class="remember-me-container">
        <label class="remember-me-label">
          <input
            type="checkbox"
            bind:checked={rememberMe}
            disabled={disabled || isSubmitting}
            class="remember-me-checkbox"
          />
          <span class="remember-me-text">ログイン状態を保持する</span>
        </label>
      </div>
    {/if}
  </div>

  <!-- CSRFトークン -->
  <input type="hidden" name="csrf_token" value={csrfToken} />

  <div class="form-actions">
    <button 
      type="submit" 
      class="login-button"
      disabled={!isFormValid || disabled || isSubmitting}
      data-testid="login-submit"
    >
      {#if isSubmitting}
        <span class="loading-spinner"></span>
        ログイン中...
      {:else}
        ログイン
      {/if}
    </button>
  </div>

  <div class="form-links">
    <button 
      type="button" 
      class="link-button"
      on:click={handleForgotPassword}
      disabled={isSubmitting}
    >
      パスワードを忘れた方
    </button>
    
    <button 
      type="button" 
      class="link-button"
      on:click={handleSignUp}
      disabled={isSubmitting}
    >
      新規アカウント作成
    </button>
  </div>
</form>

<style>
  .unified-login-form {
    max-width: 400px;
    margin: 0 auto;
    padding: 2rem;
    background-color: white;
    border-radius: 8px;
    box-shadow: 0 4px 6px rgba(0, 0, 0, 0.1);
    border: 1px solid #e5e7eb;
  }

  .form-header {
    text-align: center;
    margin-bottom: 2rem;
  }

  .form-header h2 {
    font-size: 1.5rem;
    font-weight: 600;
    color: #111827;
    margin: 0 0 0.5rem 0;
  }

  .form-description {
    color: #6b7280;
    font-size: 0.875rem;
    margin: 0;
  }

  .form-fields {
    display: flex;
    flex-direction: column;
    gap: 1.5rem;
    margin-bottom: 1.5rem;
  }

  .remember-me-container {
    display: flex;
    align-items: center;
  }

  .remember-me-label {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    cursor: pointer;
    font-size: 0.875rem;
    color: #374151;
  }

  .remember-me-checkbox {
    width: 1rem;
    height: 1rem;
    border: 1px solid #d1d5db;
    border-radius: 0.25rem;
    cursor: pointer;
  }

  .remember-me-checkbox:checked {
    background-color: #3b82f6;
    border-color: #3b82f6;
  }

  .remember-me-checkbox:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }

  .remember-me-text {
    user-select: none;
  }

  .form-actions {
    margin-bottom: 1.5rem;
  }

  .login-button {
    width: 100%;
    padding: 0.75rem 1rem;
    background-color: #3b82f6;
    color: white;
    border: none;
    border-radius: 0.375rem;
    font-size: 1rem;
    font-weight: 500;
    cursor: pointer;
    transition: background-color 0.2s;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
  }

  .login-button:hover:not(:disabled) {
    background-color: #2563eb;
  }

  .login-button:disabled {
    background-color: #9ca3af;
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

  .form-links {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
    align-items: center;
  }

  .link-button {
    background: none;
    border: none;
    color: #3b82f6;
    font-size: 0.875rem;
    cursor: pointer;
    text-decoration: underline;
    padding: 0.25rem;
    transition: color 0.2s;
  }

  .link-button:hover:not(:disabled) {
    color: #2563eb;
  }

  .link-button:disabled {
    color: #9ca3af;
    cursor: not-allowed;
  }

  /* レスポンシブデザイン */
  @media (max-width: 480px) {
    .unified-login-form {
      margin: 0 1rem;
      padding: 1.5rem;
    }

    .form-header h2 {
      font-size: 1.25rem;
    }

    .form-fields {
      gap: 1.25rem;
    }
  }

  /* アクセシビリティ対応 */
  @media (prefers-reduced-motion: reduce) {
    .loading-spinner {
      animation: none;
    }
    
    .login-button, .link-button {
      transition: none;
    }
  }

  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .unified-login-form {
      border: 2px solid #000;
    }
    
    .login-button {
      border: 2px solid #000;
    }
    
    .remember-me-checkbox {
      border: 2px solid #000;
    }
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .unified-login-form {
      background-color: #1f2937;
      border-color: #374151;
    }
    
    .form-header h2 {
      color: #f9fafb;
    }
    
    .form-description {
      color: #9ca3af;
    }
    
    .remember-me-label {
      color: #d1d5db;
    }
  }
</style>
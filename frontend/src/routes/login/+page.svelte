<script>
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { authStore } from '$lib/stores/auth.js';
  import { uiActions, showErrorNotification, showSuccessNotification } from '$lib/stores/ui.js';
  import { validateLoginCredentials, debounce } from '$lib/utils/validation.js';
  import { csrfTokenManager, defaultRateLimiter, securityLogger } from '$lib/utils/security.js';
  
  // ページデータを取得
  export let data;

  // フォームデータ
  let formData = {
    username: '',
    password: ''
  };

  // バリデーションエラー
  let validationErrors = {};

  // フォーム状態
  let isSubmitting = false;
  let showPassword = false;
  let loginAttempts = 0;
  let isRateLimited = false;
  
  // CSRFトークン
  let csrfToken = '';
  
  // エラー表示用
  let errorMessage = '';
  let successMessage = '';

  // 認証ストアの状態を購読
  let authState = {};
  authStore.subscribe((state) => {
    authState = state;
  });

  // 初期化処理
  onMount(() => {
    // CSRFトークンを取得
    csrfToken = csrfTokenManager.getToken();
    
    // URLパラメータからエラー情報を表示
    if (data.errorType) {
      displayErrorMessage(data.errorType);
    }
    
    // 既にログイン済みの場合は管理者ダッシュボードにリダイレクト
    if (authState.isAuthenticated) {
      const redirectTo = data.redirectTarget || '/admin';
      goto(redirectTo);
    }
  });
  
  // エラーメッセージの表示
  function displayErrorMessage(errorType) {
    const errorMessages = {
      expired: 'セッションの有効期限が切れました。再度ログインしてください。',
      invalid: 'トークンが無効です。再度ログインしてください。',
      unauthorized: 'アクセス権限がありません。',
      network_error: 'ネットワークエラーが発生しました。',
      logout_error: 'ログアウト処理でエラーが発生しましたが、認証情報はクリアされました。',
      session_expired: 'セッションが期限切れになりました。',
      error: 'エラーが発生しました。再度お試しください。'
    };
    
    errorMessage = errorMessages[errorType] || 'エラーが発生しました。';
    
    // セキュリティログに記録
    securityLogger.logEvent('LOGIN_PAGE_ERROR', {
      errorType,
      userAgent: navigator.userAgent,
      timestamp: new Date().toISOString()
    });
  }

  // リアルタイムバリデーション（デバウンス付き）
  const debouncedValidation = debounce((fieldName) => {
    const validation = validateLoginCredentials(formData.username, formData.password);

    if (validation.errors[fieldName]) {
      validationErrors[fieldName] = validation.errors[fieldName];
    } else {
      delete validationErrors[fieldName];
    }

    // リアクティブ更新をトリガー
    validationErrors = { ...validationErrors };
  }, 300);

  function validateField(fieldName) {
    debouncedValidation(fieldName);
  }

  // フォーム送信処理
  async function handleSubmit(event) {
    event.preventDefault();

    // 既に送信中の場合は処理をスキップ
    if (isSubmitting) return;

    // レート制限チェック
    const clientId = `login_${formData.username || 'anonymous'}`;
    if (!defaultRateLimiter.isAllowed(clientId)) {
      isRateLimited = true;
      errorMessage = 'ログイン試行回数が上限に達しました。しばらく待ってから再試行してください。';
      
      securityLogger.logEvent('RATE_LIMIT_EXCEEDED', {
        username: formData.username,
        attempts: loginAttempts,
        clientId
      });
      
      return;
    }

    // フォーム全体のバリデーション
    const validation = validateLoginCredentials(formData.username, formData.password);

    if (!validation.isValid) {
      validationErrors = validation.errors;
      errorMessage = '入力内容を確認してください';
      
      securityLogger.logEvent('LOGIN_VALIDATION_FAILED', {
        username: formData.username,
        errors: validation.errors
      });
      
      return;
    }

    // バリデーションエラーをクリア
    validationErrors = {};
    errorMessage = '';
    isSubmitting = true;
    loginAttempts++;

    try {
      // サニタイズされたデータを使用
      const sanitizedCredentials = {
        username: validation.sanitizedData.username,
        password: validation.sanitizedData.password,
        csrfToken: csrfToken
      };

      // セキュリティログに記録
      securityLogger.logEvent('LOGIN_ATTEMPT', {
        username: sanitizedCredentials.username,
        timestamp: new Date().toISOString(),
        userAgent: navigator.userAgent
      });

      // ログイン処理を実行
      const result = await authStore.login(sanitizedCredentials);

      if (result.success) {
        // ログイン成功
        successMessage = 'ログインに成功しました';
        loginAttempts = 0; // 成功時はカウンターをリセット
        
        securityLogger.logEvent('LOGIN_SUCCESS', {
          username: sanitizedCredentials.username,
          timestamp: new Date().toISOString()
        });

        // リダイレクト先を決定
        const redirectTo = data.redirectTarget || '/admin';

        // 管理者ダッシュボードまたは指定されたページにリダイレクト
        setTimeout(() => {
          goto(redirectTo);
        }, 1000);
      } else {
        // ログイン失敗
        let failureMessage = 'ログインに失敗しました';

        // エラーの種類に応じてメッセージを調整
        if (result.error === 'INVALID_CREDENTIALS') {
          failureMessage = 'ユーザー名またはパスワードが正しくありません';
        } else if (result.error === 'ACCOUNT_LOCKED') {
          failureMessage = 'アカウントがロックされています';
        } else if (result.error === 'CSRF_TOKEN_MISMATCH') {
          failureMessage = 'セキュリティトークンが無効です。ページを再読み込みしてください。';
          // CSRFトークンを更新
          csrfToken = csrfTokenManager.refreshToken();
        } else if (result.message) {
          failureMessage = result.message;
        }

        errorMessage = failureMessage;

        // セキュリティログに記録
        securityLogger.logEvent('LOGIN_FAILED', {
          username: sanitizedCredentials.username,
          error: result.error,
          attempts: loginAttempts,
          timestamp: new Date().toISOString()
        });

        // パスワードフィールドをクリア
        formData.password = '';

        // 連続失敗時の追加セキュリティ
        if (loginAttempts >= 3) {
          securityLogger.logEvent('MULTIPLE_LOGIN_FAILURES', {
            username: sanitizedCredentials.username,
            attempts: loginAttempts,
            timestamp: new Date().toISOString()
          });
        }
      }
    } catch (error) {
      console.error('Login submission error:', error);
      errorMessage = 'ログイン処理でエラーが発生しました';

      securityLogger.logEvent('LOGIN_ERROR', {
        username: formData.username,
        error: error.message,
        attempts: loginAttempts,
        timestamp: new Date().toISOString()
      });

      // パスワードフィールドをクリア
      formData.password = '';
    } finally {
      isSubmitting = false;
    }
  }

  // パスワード表示切り替え
  function togglePasswordVisibility() {
    showPassword = !showPassword;
  }

  // Enterキーでのフォーム送信
  function handleKeydown(event) {
    if (event.key === 'Enter' && !isSubmitting) {
      handleSubmit(event);
    }
  }
</script>

<svelte:head>
  <title>管理者ログイン - Tournament Management System</title>
</svelte:head>

<div class="login-container">
  <div class="login-card">
    <h1>管理者ログイン</h1>
    <p class="login-description">
      トーナメント管理システムの管理者ダッシュボードにアクセスするには、認証情報を入力してください。
    </p>

    <!-- エラー・成功メッセージ表示 -->
    {#if errorMessage}
      <div class="alert alert-error" role="alert">
        {errorMessage}
      </div>
    {/if}
    
    {#if successMessage}
      <div class="alert alert-success" role="alert">
        {successMessage}
      </div>
    {/if}
    
    {#if isRateLimited}
      <div class="alert alert-warning" role="alert">
        セキュリティのため、ログイン試行が一時的に制限されています。
        残り試行回数: {defaultRateLimiter.getRemainingRequests(formData.username || 'anonymous')}
      </div>
    {/if}

    <form on:submit={handleSubmit} novalidate>
      <!-- ユーザー名フィールド -->
      <div class="form-group">
        <label for="username" class:error={validationErrors.username}> ユーザー名 </label>
        <input
          type="text"
          id="username"
          name="username"
          bind:value={formData.username}
          on:blur={() => validateField('username')}
          on:input={() => validateField('username')}
          on:keydown={handleKeydown}
          class:error={validationErrors.username}
          disabled={isSubmitting}
          autocomplete="username"
          data-testid="username"
          required
        />
        {#if validationErrors.username}
          <span class="error-message">{validationErrors.username}</span>
        {/if}
      </div>

      <!-- パスワードフィールド -->
      <div class="form-group">
        <label for="password" class:error={validationErrors.password}> パスワード </label>
        <div class="password-input-container">
          <input
            type={showPassword ? 'text' : 'password'}
            id="password"
            name="password"
            bind:value={formData.password}
            on:blur={() => validateField('password')}
            on:input={() => validateField('password')}
            on:keydown={handleKeydown}
            class:error={validationErrors.password}
            disabled={isSubmitting}
            autocomplete="current-password"
            data-testid="password"
            required
          />
          <button
            type="button"
            class="password-toggle"
            on:click={togglePasswordVisibility}
            disabled={isSubmitting}
            aria-label={showPassword ? 'パスワードを隠す' : 'パスワードを表示'}
          >
            {showPassword ? '🙈' : '👁️'}
          </button>
        </div>
        {#if validationErrors.password}
          <span class="error-message">{validationErrors.password}</span>
        {/if}
      </div>

      <!-- CSRFトークン -->
      <input type="hidden" name="csrf_token" value={csrfToken} />

      <!-- 送信ボタン -->
      <button
        type="submit"
        class="login-button"
        disabled={isSubmitting || authState.loading || isRateLimited}
        data-testid="login-button"
      >
        {#if isSubmitting || authState.loading}
          <span class="loading-spinner"></span>
          ログイン中...
        {:else if isRateLimited}
          制限中
        {:else}
          ログイン
        {/if}
      </button>
    </form>

    <!-- ローディング状態の表示 -->
    {#if authState.loading}
      <div class="loading-overlay">
        <div class="loading-content">
          <span class="loading-spinner large"></span>
          <p>認証中...</p>
        </div>
      </div>
    {/if}
  </div>
</div>

<style>
  .login-container {
    min-height: calc(100vh - 200px);
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2rem 1rem;
  }

  .login-card {
    width: 100%;
    max-width: 400px;
    background: #fff;
    border-radius: 12px;
    box-shadow: 0 10px 25px rgba(0, 0, 0, 0.1);
    padding: 2.5rem;
    border: 1px solid #e9ecef;
  }

  h1 {
    text-align: center;
    margin-bottom: 0.5rem;
    color: #212529;
    font-size: 1.75rem;
    font-weight: 600;
  }

  .login-description {
    text-align: center;
    color: #6c757d;
    margin-bottom: 2rem;
    font-size: 0.875rem;
    line-height: 1.5;
  }

  .alert {
    padding: 0.75rem 1rem;
    margin-bottom: 1rem;
    border-radius: 6px;
    font-size: 0.875rem;
    font-weight: 500;
    border: 1px solid transparent;
  }

  .alert-error {
    color: #721c24;
    background-color: #f8d7da;
    border-color: #f5c6cb;
  }

  .alert-success {
    color: #155724;
    background-color: #d4edda;
    border-color: #c3e6cb;
  }

  .alert-warning {
    color: #856404;
    background-color: #fff3cd;
    border-color: #ffeaa7;
  }

  .form-group {
    margin-bottom: 1.5rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    color: #495057;
    font-weight: 500;
    font-size: 0.875rem;
  }

  label.error {
    color: #dc3545;
  }

  input {
    width: 100%;
    padding: 0.75rem 1rem;
    border: 2px solid #e9ecef;
    border-radius: 8px;
    font-size: 1rem;
    transition: all 0.2s ease;
    background-color: #fff;
    box-sizing: border-box;
  }

  input:focus {
    outline: none;
    border-color: #007bff;
    box-shadow: 0 0 0 3px rgba(0, 123, 255, 0.1);
  }

  input.error {
    border-color: #dc3545;
  }

  input.error:focus {
    border-color: #dc3545;
    box-shadow: 0 0 0 3px rgba(220, 53, 69, 0.1);
  }

  input:disabled {
    background-color: #f8f9fa;
    color: #6c757d;
    cursor: not-allowed;
  }

  .password-input-container {
    position: relative;
  }

  .password-toggle {
    position: absolute;
    right: 12px;
    top: 50%;
    transform: translateY(-50%);
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    color: #6c757d;
    font-size: 1rem;
    transition: color 0.2s ease;
    width: auto;
  }

  .password-toggle:hover {
    color: #495057;
    background-color: #f8f9fa;
  }

  .password-toggle:disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }

  .error-message {
    display: block;
    color: #dc3545;
    font-size: 0.75rem;
    margin-top: 0.25rem;
    font-weight: 500;
  }

  .login-button {
    width: 100%;
    padding: 0.875rem 1rem;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 8px;
    font-size: 1rem;
    font-weight: 600;
    cursor: pointer;
    transition: all 0.2s ease;
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    min-height: 48px;
  }

  .login-button:hover:not(:disabled) {
    background-color: #0056b3;
    transform: translateY(-1px);
    box-shadow: 0 4px 12px rgba(0, 123, 255, 0.3);
  }

  .login-button:active:not(:disabled) {
    transform: translateY(0);
  }

  .login-button:disabled {
    background-color: #6c757d;
    cursor: not-allowed;
    transform: none;
    box-shadow: none;
  }

  .loading-spinner {
    width: 16px;
    height: 16px;
    border: 2px solid transparent;
    border-top: 2px solid currentColor;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  .loading-spinner.large {
    width: 32px;
    height: 32px;
    border-width: 3px;
  }

  .loading-overlay {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    background-color: rgba(255, 255, 255, 0.9);
    display: flex;
    align-items: center;
    justify-content: center;
    border-radius: 12px;
    z-index: 10;
  }

  .loading-content {
    text-align: center;
    color: #495057;
  }

  .loading-content p {
    margin-top: 1rem;
    font-weight: 500;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  /* レスポンシブデザイン */
  @media (max-width: 768px) {
    .login-container {
      padding: 1rem;
      min-height: calc(100vh - 160px);
    }

    .login-card {
      padding: 2rem 1.5rem;
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
    }

    h1 {
      font-size: 1.5rem;
    }

    .login-description {
      font-size: 0.8rem;
    }
  }

  @media (max-width: 480px) {
    .login-card {
      padding: 1.5rem 1rem;
      margin: 0.5rem;
    }

    h1 {
      font-size: 1.375rem;
    }

    .form-group {
      margin-bottom: 1.25rem;
    }
  }

  /* アクセシビリティ */
  @media (prefers-reduced-motion: reduce) {
    .login-button:hover {
      transform: none;
    }

    .loading-spinner {
      animation: none;
    }
  }

  /* ハイコントラストモード対応 */
  @media (prefers-contrast: high) {
    .login-card {
      border: 2px solid #000;
    }

    input {
      border-width: 2px;
    }

    .login-button {
      border: 2px solid #007bff;
    }
  }
</style>

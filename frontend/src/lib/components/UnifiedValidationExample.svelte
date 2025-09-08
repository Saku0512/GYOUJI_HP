<script>
  // UnifiedValidationExample - 統一バリデーションシステムの使用例
  import { onMount } from 'svelte';
  import { createLoginValidator } from '../utils/unified-validation.js';
  import { createRealtimeValidator } from '../utils/realtime-validation.js';
  import { handleValidationError } from '../utils/error-response-handler.js';
  import ValidatedInput from './ValidatedInput.svelte';
  import UnifiedErrorDisplay from './UnifiedErrorDisplay.svelte';
  import ValidationMessage from './ValidationMessage.svelte';

  // バリデーター設定
  let loginValidator;
  let realtimeValidator;
  
  // フォーム状態
  let isSubmitting = false;
  let submitResult = null;
  let validationErrors = null;
  let showExample = true;

  // 例用のデータ
  let exampleData = {
    validInput: { username: 'admin', password: 'password123' },
    invalidInput: { username: '', password: 'short' }
  };

  onMount(() => {
    // バリデーターの初期化
    loginValidator = createLoginValidator();
    realtimeValidator = createRealtimeValidator(loginValidator, {
      debounceMs: 300,
      validateOnChange: true,
      validateOnBlur: true
    });
  });

  // 有効な入力例を設定
  function setValidExample() {
    realtimeValidator.setFieldValue('username', exampleData.validInput.username);
    realtimeValidator.setFieldValue('password', exampleData.validInput.password);
    realtimeValidator.touchField('username');
    realtimeValidator.touchField('password');
    validationErrors = null;
    submitResult = null;
  }

  // 無効な入力例を設定
  function setInvalidExample() {
    realtimeValidator.setFieldValue('username', exampleData.invalidInput.username);
    realtimeValidator.setFieldValue('password', exampleData.invalidInput.password);
    realtimeValidator.touchField('username');
    realtimeValidator.touchField('password');
    
    // バリデーションエラーを生成
    const result = realtimeValidator.validateAll();
    if (!result.isValid) {
      validationErrors = handleValidationError(result, { language: 'ja' });
    }
    submitResult = null;
  }

  // フォームクリア
  function clearForm() {
    realtimeValidator.reset();
    validationErrors = null;
    submitResult = null;
  }

  // フォーム送信のシミュレーション
  async function handleSubmit(event) {
    event.preventDefault();
    isSubmitting = true;
    
    try {
      const result = realtimeValidator.validateAll();
      
      if (result.isValid) {
        // 成功のシミュレーション
        await new Promise(resolve => setTimeout(resolve, 1000));
        submitResult = {
          success: true,
          message: 'ログインに成功しました',
          data: result.sanitizedData
        };
        validationErrors = null;
      } else {
        // バリデーションエラーの処理
        validationErrors = handleValidationError(result, { language: 'ja' });
        submitResult = null;
      }
    } catch (error) {
      submitResult = {
        success: false,
        message: 'エラーが発生しました: ' + error.message
      };
    } finally {
      isSubmitting = false;
    }
  }

  // エラー表示の切り替え
  function toggleExample() {
    showExample = !showExample;
  }
</script>

<div class="unified-validation-example">
  <div class="example-header">
    <h2>統一バリデーションシステム デモ</h2>
    <p>フロントエンドバリデーションの統一システムの動作例です</p>
  </div>

  <div class="example-controls">
    <button type="button" on:click={setValidExample} class="control-button valid">
      有効な入力例
    </button>
    <button type="button" on:click={setInvalidExample} class="control-button invalid">
      無効な入力例
    </button>
    <button type="button" on:click={clearForm} class="control-button clear">
      クリア
    </button>
    <button type="button" on:click={toggleExample} class="control-button toggle">
      {showExample ? '非表示' : '表示'}
    </button>
  </div>

  {#if showExample && realtimeValidator}
    <div class="example-content">
      <!-- フォーム例 -->
      <div class="form-section">
        <h3>リアルタイムバリデーション付きフォーム</h3>
        
        <form on:submit={handleSubmit} class="validation-form">
          <div class="form-fields">
            <ValidatedInput
              {realtimeValidator}
              fieldName="username"
              type="text"
              label="ユーザー名"
              placeholder="ユーザー名を入力"
              required
              disabled={isSubmitting}
              fullWidth
            />

            <ValidatedInput
              {realtimeValidator}
              fieldName="password"
              type="password"
              label="パスワード"
              placeholder="パスワードを入力"
              required
              disabled={isSubmitting}
              fullWidth
            />
          </div>

          <div class="form-actions">
            <button 
              type="submit" 
              class="submit-button"
              disabled={isSubmitting}
            >
              {#if isSubmitting}
                送信中...
              {:else}
                ログイン
              {/if}
            </button>
          </div>
        </form>
      </div>

      <!-- エラー表示例 -->
      {#if validationErrors}
        <div class="error-section">
          <h3>統一エラー表示</h3>
          <UnifiedErrorDisplay
            errors={validationErrors}
            language="ja"
            showSuggestions={true}
            showContext={false}
            groupByField={true}
            dismissible={true}
            showSummary={true}
            variant="detailed"
          />
        </div>
      {/if}

      <!-- 送信結果表示 -->
      {#if submitResult}
        <div class="result-section">
          <h3>送信結果</h3>
          {#if submitResult.success}
            <ValidationMessage
              error={null}
              touched={true}
              type="success"
              fullWidth
            >
              <div slot="default">
                <strong>{submitResult.message}</strong>
                <pre>{JSON.stringify(submitResult.data, null, 2)}</pre>
              </div>
            </ValidationMessage>
          {:else}
            <ValidationMessage
              error={submitResult.message}
              touched={true}
              type="error"
              fullWidth
            />
          {/if}
        </div>
      {/if}

      <!-- バリデーション状態表示 -->
      <div class="status-section">
        <h3>バリデーション状態</h3>
        <div class="status-grid">
          <div class="status-item">
            <label>フォーム有効性:</label>
            <span class="status-value" class:valid={$realtimeValidator.isValid} class:invalid={!$realtimeValidator.isValid}>
              {$realtimeValidator.isValid ? '有効' : '無効'}
            </span>
          </div>
          
          <div class="status-item">
            <label>エラー数:</label>
            <span class="status-value">
              {Object.keys($realtimeValidator.errors).length}
            </span>
          </div>
          
          <div class="status-item">
            <label>タッチ済みフィールド:</label>
            <span class="status-value">
              {Object.keys($realtimeValidator.touched).filter(key => $realtimeValidator.touched[key]).length}
            </span>
          </div>
          
          <div class="status-item">
            <label>検証中:</label>
            <span class="status-value">
              {$realtimeValidator.isValidating ? 'はい' : 'いいえ'}
            </span>
          </div>
        </div>

        <!-- 詳細状態 -->
        <details class="status-details">
          <summary>詳細状態</summary>
          <div class="status-detail-content">
            <div class="status-detail-section">
              <h4>フォームデータ:</h4>
              <pre>{JSON.stringify($realtimeValidator.formData, null, 2)}</pre>
            </div>
            
            <div class="status-detail-section">
              <h4>エラー:</h4>
              <pre>{JSON.stringify($realtimeValidator.errors, null, 2)}</pre>
            </div>
            
            <div class="status-detail-section">
              <h4>タッチ状態:</h4>
              <pre>{JSON.stringify($realtimeValidator.touched, null, 2)}</pre>
            </div>
            
            <div class="status-detail-section">
              <h4>サニタイズ済みデータ:</h4>
              <pre>{JSON.stringify($realtimeValidator.sanitizedData, null, 2)}</pre>
            </div>
          </div>
        </details>
      </div>
    </div>
  {/if}
</div>

<style>
  .unified-validation-example {
    max-width: 800px;
    margin: 0 auto;
    padding: 2rem;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }

  .example-header {
    text-align: center;
    margin-bottom: 2rem;
  }

  .example-header h2 {
    color: #1f2937;
    margin-bottom: 0.5rem;
  }

  .example-header p {
    color: #6b7280;
    font-size: 1rem;
  }

  .example-controls {
    display: flex;
    gap: 1rem;
    justify-content: center;
    margin-bottom: 2rem;
    flex-wrap: wrap;
  }

  .control-button {
    padding: 0.5rem 1rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    background-color: white;
    color: #374151;
    cursor: pointer;
    font-size: 0.875rem;
    transition: all 0.2s ease-in-out;
  }

  .control-button:hover {
    background-color: #f9fafb;
  }

  .control-button.valid {
    border-color: #10b981;
    color: #10b981;
  }

  .control-button.valid:hover {
    background-color: #ecfdf5;
  }

  .control-button.invalid {
    border-color: #ef4444;
    color: #ef4444;
  }

  .control-button.invalid:hover {
    background-color: #fef2f2;
  }

  .control-button.clear {
    border-color: #6b7280;
    color: #6b7280;
  }

  .control-button.clear:hover {
    background-color: #f3f4f6;
  }

  .control-button.toggle {
    border-color: #3b82f6;
    color: #3b82f6;
  }

  .control-button.toggle:hover {
    background-color: #eff6ff;
  }

  .example-content {
    display: flex;
    flex-direction: column;
    gap: 2rem;
  }

  .form-section,
  .error-section,
  .result-section,
  .status-section {
    padding: 1.5rem;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    background-color: white;
  }

  .form-section h3,
  .error-section h3,
  .result-section h3,
  .status-section h3 {
    margin: 0 0 1rem 0;
    color: #1f2937;
    font-size: 1.125rem;
  }

  .validation-form {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-fields {
    display: flex;
    flex-direction: column;
    gap: 1rem;
  }

  .form-actions {
    display: flex;
    justify-content: center;
    margin-top: 1rem;
  }

  .submit-button {
    padding: 0.75rem 2rem;
    background-color: #3b82f6;
    color: white;
    border: none;
    border-radius: 0.375rem;
    font-size: 1rem;
    cursor: pointer;
    transition: background-color 0.2s ease-in-out;
  }

  .submit-button:hover:not(:disabled) {
    background-color: #2563eb;
  }

  .submit-button:disabled {
    background-color: #9ca3af;
    cursor: not-allowed;
  }

  .status-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 1rem;
    margin-bottom: 1rem;
  }

  .status-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 0.75rem;
    background-color: #f9fafb;
    border-radius: 0.375rem;
  }

  .status-item label {
    font-weight: 500;
    color: #374151;
  }

  .status-value {
    font-weight: 600;
    color: #1f2937;
  }

  .status-value.valid {
    color: #10b981;
  }

  .status-value.invalid {
    color: #ef4444;
  }

  .status-details {
    margin-top: 1rem;
  }

  .status-details summary {
    cursor: pointer;
    font-weight: 500;
    color: #374151;
    padding: 0.5rem;
    background-color: #f3f4f6;
    border-radius: 0.25rem;
  }

  .status-detail-content {
    padding: 1rem;
    background-color: #f9fafb;
    border-radius: 0.25rem;
    margin-top: 0.5rem;
  }

  .status-detail-section {
    margin-bottom: 1rem;
  }

  .status-detail-section:last-child {
    margin-bottom: 0;
  }

  .status-detail-section h4 {
    margin: 0 0 0.5rem 0;
    color: #374151;
    font-size: 0.875rem;
    font-weight: 600;
  }

  .status-detail-section pre {
    background-color: #1f2937;
    color: #f9fafb;
    padding: 0.75rem;
    border-radius: 0.25rem;
    font-size: 0.75rem;
    overflow-x: auto;
    margin: 0;
  }

  /* レスポンシブ対応 */
  @media (max-width: 768px) {
    .unified-validation-example {
      padding: 1rem;
    }

    .example-controls {
      flex-direction: column;
      align-items: center;
    }

    .control-button {
      width: 100%;
      max-width: 200px;
    }

    .status-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
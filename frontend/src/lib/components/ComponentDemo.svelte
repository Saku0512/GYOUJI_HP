<script>
  // コンポーネントデモ - 全ての共通UIコンポーネントの動作確認
  import Button from './Button.svelte';
  import Input from './Input.svelte';
  import Select from './Select.svelte';
  import LoadingSpinner from './LoadingSpinner.svelte';
  import NotificationToast from './NotificationToast.svelte';
  import Modal from './Modal.svelte';
  
  // デモ用の状態
  let inputValue = '';
  let selectValue = '';
  let showModal = false;
  let showNotification = false;
  let notificationMessage = '';
  let notificationType = 'info';
  let loading = false;
  
  // セレクトボックスのオプション
  const sportOptions = [
    { value: 'volleyball', label: 'バレーボール' },
    { value: 'table_tennis', label: '卓球' },
    { value: 'soccer', label: 'サッカー' }
  ];
  
  // イベントハンドラー
  function handleButtonClick() {
    showNotification = true;
    notificationMessage = 'ボタンがクリックされました！';
    notificationType = 'success';
  }
  
  function handleLoadingToggle() {
    loading = !loading;
    showNotification = true;
    notificationMessage = loading ? 'ローディング開始' : 'ローディング終了';
    notificationType = 'info';
  }
  
  function handleModalOpen() {
    showModal = true;
  }
  
  function handleModalClose() {
    showModal = false;
    showNotification = true;
    notificationMessage = 'モーダルが閉じられました';
    notificationType = 'info';
  }
  
  function handleNotificationClose() {
    showNotification = false;
  }
  
  function handleInputChange(event) {
    inputValue = event.detail.value;
  }
  
  function handleSelectChange(event) {
    selectValue = event.detail.value;
    showNotification = true;
    notificationMessage = `${event.detail.value} が選択されました`;
    notificationType = 'info';
  }
</script>

<div class="demo-container">
  <h1>共通UIコンポーネント デモ</h1>
  
  <!-- Button コンポーネントのデモ -->
  <section class="demo-section">
    <h2>Button コンポーネント</h2>
    <div class="demo-grid">
      <Button variant="primary" on:click={handleButtonClick}>
        Primary Button
      </Button>
      
      <Button variant="secondary" outline>
        Secondary Outline
      </Button>
      
      <Button variant="success" size="small">
        Small Success
      </Button>
      
      <Button variant="danger" size="large">
        Large Danger
      </Button>
      
      <Button variant="warning" {loading} on:click={handleLoadingToggle}>
        {loading ? 'Loading...' : 'Toggle Loading'}
      </Button>
      
      <Button variant="info" disabled>
        Disabled Button
      </Button>
    </div>
  </section>
  
  <!-- Input コンポーネントのデモ -->
  <section class="demo-section">
    <h2>Input コンポーネント</h2>
    <div class="demo-grid">
      <Input
        label="基本入力"
        bind:value={inputValue}
        placeholder="テキストを入力"
        helperText="何でも入力してください"
        on:input={handleInputChange}
      />
      
      <Input
        type="email"
        label="メールアドレス"
        placeholder="email@example.com"
        variant="success"
      />
      
      <Input
        type="password"
        label="パスワード"
        placeholder="パスワードを入力"
        required
      />
      
      <Input
        label="エラー例"
        variant="error"
        errorMessage="この項目は必須です"
        value="invalid input"
      />
    </div>
  </section>
  
  <!-- Select コンポーネントのデモ -->
  <section class="demo-section">
    <h2>Select コンポーネント</h2>
    <div class="demo-grid">
      <Select
        label="スポーツ選択"
        bind:value={selectValue}
        options={sportOptions}
        placeholder="スポーツを選択してください"
        on:change={handleSelectChange}
      />
      
      <Select
        label="サイズ選択"
        options={[
          { value: 'small', label: 'Small' },
          { value: 'medium', label: 'Medium' },
          { value: 'large', label: 'Large' }
        ]}
        size="small"
      />
      
      <Select
        label="無効な選択"
        options={sportOptions}
        disabled
        helperText="この選択は無効です"
      />
    </div>
  </section>
  
  <!-- LoadingSpinner コンポーネントのデモ -->
  <section class="demo-section">
    <h2>LoadingSpinner コンポーネント</h2>
    <div class="demo-grid">
      <div class="spinner-demo">
        <p>Small Spinner</p>
        <LoadingSpinner size="small" />
      </div>
      
      <div class="spinner-demo">
        <p>Medium Spinner</p>
        <LoadingSpinner size="medium" />
      </div>
      
      <div class="spinner-demo">
        <p>Large Spinner</p>
        <LoadingSpinner size="large" color="#28a745" />
      </div>
    </div>
  </section>
  
  <!-- Modal コンポーネントのデモ -->
  <section class="demo-section">
    <h2>Modal コンポーネント</h2>
    <div class="demo-grid">
      <Button variant="primary" on:click={handleModalOpen}>
        モーダルを開く
      </Button>
    </div>
  </section>
  
  <!-- 現在の状態表示 -->
  <section class="demo-section">
    <h2>現在の状態</h2>
    <div class="state-display">
      <p><strong>Input Value:</strong> {inputValue || '(空)'}</p>
      <p><strong>Select Value:</strong> {selectValue || '(未選択)'}</p>
      <p><strong>Loading:</strong> {loading ? 'Yes' : 'No'}</p>
      <p><strong>Modal Open:</strong> {showModal ? 'Yes' : 'No'}</p>
    </div>
  </section>
</div>

<!-- Modal -->
<Modal
  bind:open={showModal}
  title="デモモーダル"
  size="medium"
  on:close={handleModalClose}
>
  <p>これはモーダルダイアログのデモです。</p>
  <p>背景をクリックするかEscapeキーで閉じることができます。</p>
  
  <div slot="footer">
    <Button variant="secondary" on:click={() => showModal = false}>
      キャンセル
    </Button>
    <Button variant="primary" on:click={() => showModal = false}>
      OK
    </Button>
  </div>
</Modal>

<!-- Notification Toast -->
{#if showNotification}
  <NotificationToast
    message={notificationMessage}
    type={notificationType}
    duration={3000}
    on:close={handleNotificationClose}
  />
{/if}

<style>
  .demo-container {
    max-width: 1200px;
    margin: 0 auto;
    padding: 2rem;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }
  
  .demo-section {
    margin-bottom: 3rem;
    padding: 1.5rem;
    border: 1px solid #e5e7eb;
    border-radius: 0.5rem;
    background-color: #f9fafb;
  }
  
  .demo-section h2 {
    margin-top: 0;
    margin-bottom: 1.5rem;
    color: #111827;
    font-size: 1.5rem;
    font-weight: 600;
  }
  
  .demo-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
    gap: 1rem;
  }
  
  .spinner-demo {
    display: flex;
    flex-direction: column;
    align-items: center;
    padding: 1rem;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
    background-color: white;
  }
  
  .spinner-demo p {
    margin: 0 0 1rem 0;
    font-weight: 500;
  }
  
  .state-display {
    padding: 1rem;
    background-color: white;
    border: 1px solid #d1d5db;
    border-radius: 0.375rem;
  }
  
  .state-display p {
    margin: 0.5rem 0;
  }
  
  h1 {
    text-align: center;
    color: #111827;
    margin-bottom: 2rem;
  }
  
  @media (max-width: 768px) {
    .demo-container {
      padding: 1rem;
    }
    
    .demo-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
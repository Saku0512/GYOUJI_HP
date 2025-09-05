<script>
  // ログインページ - 管理者ログインフォーム
</script>

<svelte:head>
  <title>Login - Tournament Management System</title>
</svelte:head>

<div class="login-container">
  <h1>管理者ログイン</h1>
  <form>
    <div class="form-group">
      <label for="username">ユーザー名</label>
      <input type="text" id="username" name="username" required />
    </div>
    <div class="form-group">
      <label for="password">パスワード</label>
      <input type="password" id="password" name="password" required />
    </div>
    <button type="submit">ログイン</button>
  </form>
</div>

<style>
  .login-container {
    max-width: 400px;
    margin: 2rem auto;
    padding: 2rem;
    border: 1px solid #ddd;
    border-radius: 8px;
  }

  h1 {
    text-align: center;
    margin-bottom: 2rem;
    color: #333;
  }

  .form-group {
    margin-bottom: 1rem;
  }

  label {
    display: block;
    margin-bottom: 0.5rem;
    color: #555;
  }

  input {
    width: 100%;
    padding: 0.75rem;
    border: 1px solid #ddd;
    border-radius: 4px;
    font-size: 1rem;
  }

  button {
    width: 100%;
    padding: 0.75rem;
    background-color: #007bff;
    color: white;
    border: none;
    border-radius: 4px;
    font-size: 1rem;
    cursor: pointer;
  }

  button:hover {
    background-color: #0056b3;
  }
</style>

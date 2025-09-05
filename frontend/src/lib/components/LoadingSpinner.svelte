<script>
  // LoadingSpinner コンポーネント - ローディング表示
  export let size = 'medium'; // 'small', 'medium', 'large'
  export let color = '#007bff';
</script>

<div class="spinner-container">
  <div class="spinner {size}" style="border-top-color: {color}"></div>
</div>

<style>
  .spinner-container {
    display: flex;
    justify-content: center;
    align-items: center;
    padding: 1rem;
  }

  .spinner {
    border: 3px solid #f3f3f3;
    border-radius: 50%;
    border-top: 3px solid #007bff;
    animation: spin 1s linear infinite;
  }

  .spinner.small {
    width: 20px;
    height: 20px;
    border-width: 2px;
  }

  .spinner.medium {
    width: 40px;
    height: 40px;
    border-width: 3px;
  }

  .spinner.large {
    width: 60px;
    height: 60px;
    border-width: 4px;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }
</style>

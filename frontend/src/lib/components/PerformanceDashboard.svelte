<!--
  パフォーマンスダッシュボードコンポーネント
  リアルタイムのパフォーマンスメトリクスを表示
-->
<script>
  import { onMount, onDestroy } from 'svelte';
  import { performanceMonitor } from '$lib/utils/performanceMonitor.js';
  import { cacheManager } from '$lib/utils/assetCache.js';

  // Props
  export let visible = false;
  export let position = 'bottom-right'; // 'bottom-right', 'bottom-left', 'top-right', 'top-left'

  // 内部状態
  let metrics = {};
  let cacheStats = {};
  let updateInterval;
  let expanded = false;

  // パフォーマンススコアの色分け
  function getScoreColor(score) {
    if (score >= 90) return '#4caf50'; // 緑
    if (score >= 70) return '#ff9800'; // オレンジ
    return '#f44336'; // 赤
  }

  // バイト数のフォーマット
  function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  // 時間のフォーマット
  function formatTime(ms) {
    if (ms < 1000) return `${Math.round(ms)}ms`;
    return `${(ms / 1000).toFixed(2)}s`;
  }

  // メトリクスの更新
  async function updateMetrics() {
    try {
      const report = performanceMonitor.generateReport();
      metrics = report;
      
      cacheStats = await cacheManager.getStats();
    } catch (error) {
      console.warn('Failed to update performance metrics:', error);
    }
  }

  // Core Web Vitals の評価
  function evaluateCWV(metric, value) {
    if (!value) return 'unknown';
    
    switch (metric) {
      case 'lcp':
        return value <= 2500 ? 'good' : value <= 4000 ? 'needs-improvement' : 'poor';
      case 'fid':
        return value <= 100 ? 'good' : value <= 300 ? 'needs-improvement' : 'poor';
      case 'cls':
        return value <= 0.1 ? 'good' : value <= 0.25 ? 'needs-improvement' : 'poor';
      default:
        return 'unknown';
    }
  }

  // キャッシュクリア
  async function clearCache() {
    try {
      await cacheManager.clear();
      await updateMetrics();
    } catch (error) {
      console.error('Failed to clear cache:', error);
    }
  }

  onMount(() => {
    if (visible) {
      updateMetrics();
      updateInterval = setInterval(updateMetrics, 5000); // 5秒ごとに更新
    }
  });

  onDestroy(() => {
    if (updateInterval) {
      clearInterval(updateInterval);
    }
  });

  // 可視性の変更を監視
  $: if (visible) {
    updateMetrics();
    if (!updateInterval) {
      updateInterval = setInterval(updateMetrics, 5000);
    }
  } else {
    if (updateInterval) {
      clearInterval(updateInterval);
      updateInterval = null;
    }
  }
</script>

{#if visible}
  <div class="performance-dashboard" class:expanded class:position-{position}>
    <!-- ヘッダー -->
    <div class="header" on:click={() => expanded = !expanded}>
      <div class="title">
        <span class="icon">📊</span>
        パフォーマンス
      </div>
      
      {#if metrics.score !== undefined}
        <div class="score" style="color: {getScoreColor(metrics.score)}">
          {Math.round(metrics.score)}
        </div>
      {/if}
      
      <button class="toggle" class:expanded>
        <svg width="12" height="12" viewBox="0 0 12 12">
          <path d="M2 4l4 4 4-4" stroke="currentColor" stroke-width="2" fill="none"/>
        </svg>
      </button>
    </div>

    {#if expanded}
      <div class="content">
        <!-- Core Web Vitals -->
        {#if metrics.coreWebVitals}
          <div class="section">
            <h4>Core Web Vitals</h4>
            <div class="metrics-grid">
              {#if metrics.coreWebVitals.lcp}
                <div class="metric" class:good={evaluateCWV('lcp', metrics.coreWebVitals.lcp.value) === 'good'} 
                     class:needs-improvement={evaluateCWV('lcp', metrics.coreWebVitals.lcp.value) === 'needs-improvement'}
                     class:poor={evaluateCWV('lcp', metrics.coreWebVitals.lcp.value) === 'poor'}>
                  <div class="label">LCP</div>
                  <div class="value">{formatTime(metrics.coreWebVitals.lcp.value)}</div>
                </div>
              {/if}
              
              {#if metrics.coreWebVitals.fid}
                <div class="metric" class:good={evaluateCWV('fid', metrics.coreWebVitals.fid.value) === 'good'}
                     class:needs-improvement={evaluateCWV('fid', metrics.coreWebVitals.fid.value) === 'needs-improvement'}
                     class:poor={evaluateCWV('fid', metrics.coreWebVitals.fid.value) === 'poor'}>
                  <div class="label">FID</div>
                  <div class="value">{formatTime(metrics.coreWebVitals.fid.value)}</div>
                </div>
              {/if}
              
              {#if metrics.coreWebVitals.cls}
                <div class="metric" class:good={evaluateCWV('cls', metrics.coreWebVitals.cls.value) === 'good'}
                     class:needs-improvement={evaluateCWV('cls', metrics.coreWebVitals.cls.value) === 'needs-improvement'}
                     class:poor={evaluateCWV('cls', metrics.coreWebVitals.cls.value) === 'poor'}>
                  <div class="label">CLS</div>
                  <div class="value">{metrics.coreWebVitals.cls.value.toFixed(3)}</div>
                </div>
              {/if}
            </div>
          </div>
        {/if}

        <!-- リソース統計 -->
        {#if metrics.resources}
          <div class="section">
            <h4>リソース</h4>
            <div class="resource-stats">
              {#each Object.entries(metrics.resources) as [type, stats]}
                <div class="resource-item">
                  <div class="resource-type">{type}</div>
                  <div class="resource-info">
                    <span>{stats.count}個</span>
                    <span>{formatBytes(stats.totalSize)}</span>
                    {#if stats.cached > 0}
                      <span class="cached">{stats.cached}個キャッシュ済み</span>
                    {/if}
                  </div>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- キャッシュ統計 -->
        {#if cacheStats}
          <div class="section">
            <h4>キャッシュ</h4>
            <div class="cache-stats">
              <div class="cache-item">
                <span>メモリ:</span>
                <span>{cacheStats.memory?.size || 0}/{cacheStats.memory?.maxSize || 0}</span>
              </div>
              {#each Object.entries(cacheStats.browser || {}) as [cacheName, count]}
                <div class="cache-item">
                  <span>{cacheName.replace('tournament-', '')}:</span>
                  <span>{count}個</span>
                </div>
              {/each}
            </div>
            
            <button class="clear-cache-btn" on:click={clearCache}>
              キャッシュクリア
            </button>
          </div>
        {/if}

        <!-- Long Tasks -->
        {#if metrics.longTasks && metrics.longTasks.length > 0}
          <div class="section">
            <h4>Long Tasks</h4>
            <div class="long-tasks">
              {#each metrics.longTasks.slice(-3) as task}
                <div class="task-item">
                  <span>{formatTime(task.duration)}</span>
                  <span class="task-time">{new Date(task.timestamp).toLocaleTimeString()}</span>
                </div>
              {/each}
            </div>
          </div>
        {/if}

        <!-- 推奨事項 -->
        {#if metrics.recommendations && metrics.recommendations.length > 0}
          <div class="section">
            <h4>推奨事項</h4>
            <div class="recommendations">
              {#each metrics.recommendations as rec}
                <div class="recommendation" class:high={rec.priority === 'high'} class:medium={rec.priority === 'medium'}>
                  <div class="rec-priority">{rec.priority}</div>
                  <div class="rec-message">{rec.message}</div>
                </div>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {/if}
  </div>
{/if}

<style>
  .performance-dashboard {
    position: fixed;
    z-index: 10000;
    background: rgba(255, 255, 255, 0.95);
    backdrop-filter: blur(10px);
    border: 1px solid #e0e0e0;
    border-radius: 8px;
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    font-size: 12px;
    min-width: 280px;
    max-width: 400px;
    max-height: 80vh;
    overflow: hidden;
  }

  /* 位置設定 */
  .position-bottom-right {
    bottom: 20px;
    right: 20px;
  }

  .position-bottom-left {
    bottom: 20px;
    left: 20px;
  }

  .position-top-right {
    top: 20px;
    right: 20px;
  }

  .position-top-left {
    top: 20px;
    left: 20px;
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 12px 16px;
    cursor: pointer;
    user-select: none;
    border-bottom: 1px solid #e0e0e0;
  }

  .title {
    display: flex;
    align-items: center;
    gap: 8px;
    font-weight: 600;
  }

  .icon {
    font-size: 16px;
  }

  .score {
    font-weight: bold;
    font-size: 14px;
  }

  .toggle {
    background: none;
    border: none;
    cursor: pointer;
    padding: 4px;
    border-radius: 4px;
    transition: transform 0.2s ease;
  }

  .toggle:hover {
    background: #f5f5f5;
  }

  .toggle.expanded {
    transform: rotate(180deg);
  }

  .content {
    max-height: 60vh;
    overflow-y: auto;
    padding: 16px;
  }

  .section {
    margin-bottom: 16px;
  }

  .section:last-child {
    margin-bottom: 0;
  }

  h4 {
    margin: 0 0 8px 0;
    font-size: 13px;
    font-weight: 600;
    color: #333;
  }

  .metrics-grid {
    display: grid;
    grid-template-columns: repeat(3, 1fr);
    gap: 8px;
  }

  .metric {
    padding: 8px;
    border-radius: 4px;
    text-align: center;
    border: 1px solid #e0e0e0;
  }

  .metric.good {
    background: #e8f5e8;
    border-color: #4caf50;
  }

  .metric.needs-improvement {
    background: #fff3e0;
    border-color: #ff9800;
  }

  .metric.poor {
    background: #ffebee;
    border-color: #f44336;
  }

  .metric .label {
    font-size: 10px;
    color: #666;
    margin-bottom: 2px;
  }

  .metric .value {
    font-weight: 600;
    font-size: 11px;
  }

  .resource-stats,
  .cache-stats {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .resource-item,
  .cache-item {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 4px 0;
  }

  .resource-type {
    font-weight: 500;
    text-transform: capitalize;
  }

  .resource-info {
    display: flex;
    gap: 8px;
    font-size: 11px;
    color: #666;
  }

  .cached {
    color: #4caf50;
  }

  .clear-cache-btn {
    margin-top: 8px;
    padding: 6px 12px;
    background: #f5f5f5;
    border: 1px solid #ddd;
    border-radius: 4px;
    cursor: pointer;
    font-size: 11px;
    width: 100%;
  }

  .clear-cache-btn:hover {
    background: #e0e0e0;
  }

  .long-tasks {
    display: flex;
    flex-direction: column;
    gap: 4px;
  }

  .task-item {
    display: flex;
    justify-content: space-between;
    padding: 4px 8px;
    background: #fff3e0;
    border-radius: 4px;
    font-size: 11px;
  }

  .task-time {
    color: #666;
  }

  .recommendations {
    display: flex;
    flex-direction: column;
    gap: 8px;
  }

  .recommendation {
    padding: 8px;
    border-radius: 4px;
    border-left: 3px solid #ddd;
  }

  .recommendation.high {
    background: #ffebee;
    border-left-color: #f44336;
  }

  .recommendation.medium {
    background: #fff3e0;
    border-left-color: #ff9800;
  }

  .rec-priority {
    font-size: 10px;
    font-weight: 600;
    text-transform: uppercase;
    margin-bottom: 4px;
    color: #666;
  }

  .rec-message {
    font-size: 11px;
    line-height: 1.4;
  }

  /* ダークモード対応 */
  @media (prefers-color-scheme: dark) {
    .performance-dashboard {
      background: rgba(30, 30, 30, 0.95);
      border-color: #444;
      color: #fff;
    }

    .header {
      border-bottom-color: #444;
    }

    h4 {
      color: #fff;
    }

    .metric {
      border-color: #444;
    }

    .clear-cache-btn {
      background: #444;
      border-color: #555;
      color: #fff;
    }

    .clear-cache-btn:hover {
      background: #555;
    }

    .toggle:hover {
      background: #444;
    }
  }

  /* モバイル対応 */
  @media (max-width: 768px) {
    .performance-dashboard {
      position: fixed;
      bottom: 0;
      left: 0;
      right: 0;
      border-radius: 8px 8px 0 0;
      max-width: none;
    }

    .metrics-grid {
      grid-template-columns: repeat(2, 1fr);
    }
  }
</style>
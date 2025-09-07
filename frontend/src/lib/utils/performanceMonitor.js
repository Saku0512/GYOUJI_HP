/**
 * パフォーマンス監視ユーティリティ
 * Core Web Vitals、リソース読み込み時間、ユーザーエクスペリエンス指標を監視
 */

/**
 * パフォーマンスメトリクス収集クラス
 */
class PerformanceMonitor {
  constructor() {
    this.metrics = new Map();
    this.observers = new Map();
    this.isSupported = this.checkSupport();
    
    if (this.isSupported) {
      this.init();
    }
  }

  checkSupport() {
    return (
      'performance' in window &&
      'PerformanceObserver' in window &&
      'getEntriesByType' in performance
    );
  }

  init() {
    this.observeNavigationTiming();
    this.observeResourceTiming();
    this.observeLargestContentfulPaint();
    this.observeFirstInputDelay();
    this.observeCumulativeLayoutShift();
    this.observeLongTasks();
    this.setupVisibilityChangeListener();
  }

  /**
   * ナビゲーションタイミングの監視
   */
  observeNavigationTiming() {
    if ('navigation' in performance.getEntriesByType('navigation')[0]) {
      const navigation = performance.getEntriesByType('navigation')[0];
      
      this.metrics.set('navigationTiming', {
        domContentLoaded: navigation.domContentLoadedEventEnd - navigation.domContentLoadedEventStart,
        loadComplete: navigation.loadEventEnd - navigation.loadEventStart,
        domInteractive: navigation.domInteractive - navigation.navigationStart,
        firstByte: navigation.responseStart - navigation.requestStart,
        dnsLookup: navigation.domainLookupEnd - navigation.domainLookupStart,
        tcpConnect: navigation.connectEnd - navigation.connectStart,
        serverResponse: navigation.responseEnd - navigation.responseStart,
        domProcessing: navigation.domComplete - navigation.domLoading,
        timestamp: Date.now()
      });
    }
  }

  /**
   * リソースタイミングの監視
   */
  observeResourceTiming() {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      
      entries.forEach(entry => {
        if (entry.entryType === 'resource') {
          this.processResourceEntry(entry);
        }
      });
    });

    observer.observe({ entryTypes: ['resource'] });
    this.observers.set('resource', observer);
  }

  processResourceEntry(entry) {
    const resourceType = this.getResourceType(entry.name);
    const timing = {
      name: entry.name,
      type: resourceType,
      duration: entry.duration,
      size: entry.transferSize || 0,
      cached: entry.transferSize === 0 && entry.decodedBodySize > 0,
      timestamp: Date.now()
    };

    // リソースタイプ別の統計を更新
    const resourceStats = this.metrics.get('resourceStats') || {};
    if (!resourceStats[resourceType]) {
      resourceStats[resourceType] = {
        count: 0,
        totalDuration: 0,
        totalSize: 0,
        cached: 0
      };
    }

    const stats = resourceStats[resourceType];
    stats.count++;
    stats.totalDuration += timing.duration;
    stats.totalSize += timing.size;
    if (timing.cached) stats.cached++;

    this.metrics.set('resourceStats', resourceStats);
  }

  getResourceType(url) {
    if (url.match(/\.(js)$/)) return 'script';
    if (url.match(/\.(css)$/)) return 'stylesheet';
    if (url.match(/\.(png|jpg|jpeg|gif|svg|webp|ico)$/)) return 'image';
    if (url.match(/\.(woff|woff2|ttf|eot)$/)) return 'font';
    if (url.includes('/api/')) return 'api';
    return 'other';
  }

  /**
   * Largest Contentful Paint (LCP) の監視
   */
  observeLargestContentfulPaint() {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      const lastEntry = entries[entries.length - 1];
      
      this.metrics.set('lcp', {
        value: lastEntry.startTime,
        element: lastEntry.element?.tagName || 'unknown',
        url: lastEntry.url || '',
        timestamp: Date.now()
      });
    });

    observer.observe({ entryTypes: ['largest-contentful-paint'] });
    this.observers.set('lcp', observer);
  }

  /**
   * First Input Delay (FID) の監視
   */
  observeFirstInputDelay() {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      
      entries.forEach(entry => {
        if (entry.entryType === 'first-input') {
          this.metrics.set('fid', {
            value: entry.processingStart - entry.startTime,
            eventType: entry.name,
            timestamp: Date.now()
          });
        }
      });
    });

    observer.observe({ entryTypes: ['first-input'] });
    this.observers.set('fid', observer);
  }

  /**
   * Cumulative Layout Shift (CLS) の監視
   */
  observeCumulativeLayoutShift() {
    let clsValue = 0;
    let sessionValue = 0;
    let sessionEntries = [];

    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      
      entries.forEach(entry => {
        if (!entry.hadRecentInput) {
          const firstSessionEntry = sessionEntries[0];
          const lastSessionEntry = sessionEntries[sessionEntries.length - 1];

          // セッションが5秒以上空いている、または5秒以上続いている場合は新しいセッション
          if (sessionValue &&
              entry.startTime - lastSessionEntry.startTime < 1000 &&
              entry.startTime - firstSessionEntry.startTime < 5000) {
            sessionValue += entry.value;
            sessionEntries.push(entry);
          } else {
            sessionValue = entry.value;
            sessionEntries = [entry];
          }

          if (sessionValue > clsValue) {
            clsValue = sessionValue;
            
            this.metrics.set('cls', {
              value: clsValue,
              entries: sessionEntries.length,
              timestamp: Date.now()
            });
          }
        }
      });
    });

    observer.observe({ entryTypes: ['layout-shift'] });
    this.observers.set('cls', observer);
  }

  /**
   * Long Tasks の監視
   */
  observeLongTasks() {
    const observer = new PerformanceObserver((list) => {
      const entries = list.getEntries();
      
      entries.forEach(entry => {
        if (entry.entryType === 'longtask') {
          const longTasks = this.metrics.get('longTasks') || [];
          longTasks.push({
            duration: entry.duration,
            startTime: entry.startTime,
            timestamp: Date.now()
          });
          
          // 最新の10個のみ保持
          if (longTasks.length > 10) {
            longTasks.shift();
          }
          
          this.metrics.set('longTasks', longTasks);
        }
      });
    });

    observer.observe({ entryTypes: ['longtask'] });
    this.observers.set('longtask', observer);
  }

  /**
   * ページ可視性変更の監視
   */
  setupVisibilityChangeListener() {
    let visibilityStart = Date.now();
    let totalVisibleTime = 0;

    const handleVisibilityChange = () => {
      if (document.hidden) {
        totalVisibleTime += Date.now() - visibilityStart;
      } else {
        visibilityStart = Date.now();
      }

      this.metrics.set('visibility', {
        totalVisibleTime,
        isVisible: !document.hidden,
        timestamp: Date.now()
      });
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
  }

  /**
   * カスタムメトリクスの記録
   */
  recordCustomMetric(name, value, metadata = {}) {
    this.metrics.set(name, {
      value,
      metadata,
      timestamp: Date.now()
    });
  }

  /**
   * 時間計測の開始
   */
  startTiming(name) {
    performance.mark(`${name}-start`);
  }

  /**
   * 時間計測の終了
   */
  endTiming(name) {
    performance.mark(`${name}-end`);
    performance.measure(name, `${name}-start`, `${name}-end`);
    
    const measure = performance.getEntriesByName(name, 'measure')[0];
    if (measure) {
      this.recordCustomMetric(name, measure.duration);
    }
  }

  /**
   * 全メトリクスの取得
   */
  getAllMetrics() {
    const metrics = {};
    
    for (const [key, value] of this.metrics.entries()) {
      metrics[key] = value;
    }

    return metrics;
  }

  /**
   * Core Web Vitals の取得
   */
  getCoreWebVitals() {
    return {
      lcp: this.metrics.get('lcp'),
      fid: this.metrics.get('fid'),
      cls: this.metrics.get('cls')
    };
  }

  /**
   * パフォーマンススコアの計算
   */
  calculatePerformanceScore() {
    const cwv = this.getCoreWebVitals();
    let score = 100;

    // LCP スコア (2.5秒以下が良好)
    if (cwv.lcp) {
      const lcpValue = cwv.lcp.value / 1000; // ミリ秒を秒に変換
      if (lcpValue > 4) score -= 30;
      else if (lcpValue > 2.5) score -= 15;
    }

    // FID スコア (100ms以下が良好)
    if (cwv.fid) {
      const fidValue = cwv.fid.value;
      if (fidValue > 300) score -= 30;
      else if (fidValue > 100) score -= 15;
    }

    // CLS スコア (0.1以下が良好)
    if (cwv.cls) {
      const clsValue = cwv.cls.value;
      if (clsValue > 0.25) score -= 30;
      else if (clsValue > 0.1) score -= 15;
    }

    // Long Tasks ペナルティ
    const longTasks = this.metrics.get('longTasks') || [];
    const recentLongTasks = longTasks.filter(task => 
      Date.now() - task.timestamp < 60000 // 過去1分間
    );
    score -= Math.min(recentLongTasks.length * 5, 20);

    return Math.max(score, 0);
  }

  /**
   * パフォーマンスレポートの生成
   */
  generateReport() {
    const metrics = this.getAllMetrics();
    const cwv = this.getCoreWebVitals();
    const score = this.calculatePerformanceScore();

    return {
      timestamp: Date.now(),
      score,
      coreWebVitals: cwv,
      navigation: metrics.navigationTiming,
      resources: metrics.resourceStats,
      longTasks: metrics.longTasks || [],
      visibility: metrics.visibility,
      recommendations: this.generateRecommendations(cwv, metrics)
    };
  }

  /**
   * パフォーマンス改善の推奨事項を生成
   */
  generateRecommendations(cwv, metrics) {
    const recommendations = [];

    // LCP の改善提案
    if (cwv.lcp && cwv.lcp.value > 2500) {
      recommendations.push({
        type: 'lcp',
        priority: 'high',
        message: 'Largest Contentful Paint が遅いです。画像の最適化やクリティカルリソースの優先読み込みを検討してください。'
      });
    }

    // FID の改善提案
    if (cwv.fid && cwv.fid.value > 100) {
      recommendations.push({
        type: 'fid',
        priority: 'high',
        message: 'First Input Delay が長いです。JavaScriptの実行時間を短縮してください。'
      });
    }

    // CLS の改善提案
    if (cwv.cls && cwv.cls.value > 0.1) {
      recommendations.push({
        type: 'cls',
        priority: 'medium',
        message: 'Cumulative Layout Shift が大きいです。画像や広告のサイズを事前に指定してください。'
      });
    }

    // Long Tasks の改善提案
    const longTasks = metrics.longTasks || [];
    if (longTasks.length > 3) {
      recommendations.push({
        type: 'longtask',
        priority: 'medium',
        message: 'Long Tasks が多く検出されています。JavaScriptの処理を分割してください。'
      });
    }

    // リソースサイズの改善提案
    const resources = metrics.resourceStats || {};
    if (resources.script && resources.script.totalSize > 1024 * 1024) { // 1MB以上
      recommendations.push({
        type: 'bundle-size',
        priority: 'medium',
        message: 'JavaScriptバンドルサイズが大きいです。コード分割を検討してください。'
      });
    }

    return recommendations;
  }

  /**
   * メトリクスの送信（分析サービスへ）
   */
  async sendMetrics(endpoint) {
    if (!endpoint) return;

    const report = this.generateReport();
    
    try {
      await fetch(endpoint, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json'
        },
        body: JSON.stringify(report)
      });
    } catch (error) {
      console.warn('Failed to send performance metrics:', error);
    }
  }

  /**
   * 監視の停止
   */
  disconnect() {
    for (const observer of this.observers.values()) {
      observer.disconnect();
    }
    this.observers.clear();
    this.metrics.clear();
  }
}

/**
 * リアルユーザーモニタリング (RUM)
 */
class RealUserMonitoring {
  constructor(options = {}) {
    this.options = {
      sampleRate: 0.1, // 10%のユーザーからデータを収集
      endpoint: null,
      ...options
    };
    
    this.performanceMonitor = new PerformanceMonitor();
    this.sessionId = this.generateSessionId();
    this.pageLoadTime = Date.now();
    
    if (Math.random() < this.options.sampleRate) {
      this.init();
    }
  }

  generateSessionId() {
    return Date.now().toString(36) + Math.random().toString(36).substr(2);
  }

  init() {
    // ページ離脱時にメトリクスを送信
    window.addEventListener('beforeunload', () => {
      this.sendMetrics();
    });

    // 定期的にメトリクスを送信
    setInterval(() => {
      this.sendMetrics();
    }, 30000); // 30秒ごと
  }

  async sendMetrics() {
    if (!this.options.endpoint) return;

    const report = this.performanceMonitor.generateReport();
    const payload = {
      sessionId: this.sessionId,
      url: window.location.href,
      userAgent: navigator.userAgent,
      viewport: {
        width: window.innerWidth,
        height: window.innerHeight
      },
      connection: this.getConnectionInfo(),
      sessionDuration: Date.now() - this.pageLoadTime,
      ...report
    };

    try {
      // Beacon API を使用（ページ離脱時でも送信可能）
      if ('sendBeacon' in navigator) {
        navigator.sendBeacon(
          this.options.endpoint,
          JSON.stringify(payload)
        );
      } else {
        await fetch(this.options.endpoint, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(payload)
        });
      }
    } catch (error) {
      console.warn('Failed to send RUM data:', error);
    }
  }

  getConnectionInfo() {
    if ('connection' in navigator) {
      const conn = navigator.connection;
      return {
        effectiveType: conn.effectiveType,
        downlink: conn.downlink,
        rtt: conn.rtt,
        saveData: conn.saveData
      };
    }
    return null;
  }
}

// グローバルインスタンス
export const performanceMonitor = new PerformanceMonitor();

// RUM の初期化（環境変数で制御）
export function initRUM(options = {}) {
  return new RealUserMonitoring(options);
}

// パフォーマンス計測のヘルパー関数
export function measureAsync(name, asyncFn) {
  return async (...args) => {
    performanceMonitor.startTiming(name);
    try {
      const result = await asyncFn(...args);
      return result;
    } finally {
      performanceMonitor.endTiming(name);
    }
  };
}

export function measureSync(name, fn) {
  return (...args) => {
    performanceMonitor.startTiming(name);
    try {
      const result = fn(...args);
      return result;
    } finally {
      performanceMonitor.endTiming(name);
    }
  };
}
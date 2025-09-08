package metrics

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"time"
)

// DefaultCollector はデフォルトのメトリクス収集器実装
type DefaultCollector struct {
	counters   map[string]*Counter
	gauges     map[string]*Gauge
	histograms map[string]*Histogram
	summaries  map[string]*Summary
	mutex      sync.RWMutex

	// 事前定義されたメトリクス
	httpRequestTotal    *Counter
	httpRequestDuration *Histogram
	httpRequestSize     *Histogram
	httpResponseSize    *Histogram
	httpErrorTotal      *Counter

	dbQueryTotal    *Counter
	dbQueryDuration *Histogram
	dbErrorTotal    *Counter

	activeUsers          *Gauge
	tournamentCount      *Gauge
	matchCount           *Gauge
	websocketConnections *Gauge

	memoryUsage    *Gauge
	cpuUsage       *Gauge
	goroutineCount *Gauge
}

// NewDefaultCollector は新しいDefaultCollectorを作成する
func NewDefaultCollector() *DefaultCollector {
	collector := &DefaultCollector{
		counters:   make(map[string]*Counter),
		gauges:     make(map[string]*Gauge),
		histograms: make(map[string]*Histogram),
		summaries:  make(map[string]*Summary),
	}

	// 事前定義されたメトリクスを初期化
	collector.initPredefinedMetrics()

	return collector
}

// initPredefinedMetrics は事前定義されたメトリクスを初期化する
func (c *DefaultCollector) initPredefinedMetrics() {
	// HTTPメトリクス
	c.httpRequestTotal = c.RegisterCounter(
		"http_requests_total",
		"Total number of HTTP requests",
		Labels{},
	)

	c.httpRequestDuration = c.RegisterHistogram(
		"http_request_duration_seconds",
		"HTTP request duration in seconds",
		Labels{},
		[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	)

	c.httpRequestSize = c.RegisterHistogram(
		"http_request_size_bytes",
		"HTTP request size in bytes",
		Labels{},
		[]float64{100, 1000, 10000, 100000, 1000000},
	)

	c.httpResponseSize = c.RegisterHistogram(
		"http_response_size_bytes",
		"HTTP response size in bytes",
		Labels{},
		[]float64{100, 1000, 10000, 100000, 1000000},
	)

	c.httpErrorTotal = c.RegisterCounter(
		"http_errors_total",
		"Total number of HTTP errors",
		Labels{},
	)

	// データベースメトリクス
	c.dbQueryTotal = c.RegisterCounter(
		"db_queries_total",
		"Total number of database queries",
		Labels{},
	)

	c.dbQueryDuration = c.RegisterHistogram(
		"db_query_duration_seconds",
		"Database query duration in seconds",
		Labels{},
		[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	)

	c.dbErrorTotal = c.RegisterCounter(
		"db_errors_total",
		"Total number of database errors",
		Labels{},
	)

	// アプリケーションメトリクス
	c.activeUsers = c.RegisterGauge(
		"active_users",
		"Number of active users",
		Labels{},
	)

	c.tournamentCount = c.RegisterGauge(
		"tournaments_total",
		"Total number of tournaments",
		Labels{},
	)

	c.matchCount = c.RegisterGauge(
		"matches_total",
		"Total number of matches",
		Labels{},
	)

	c.websocketConnections = c.RegisterGauge(
		"websocket_connections",
		"Number of active WebSocket connections",
		Labels{},
	)

	// システムメトリクス
	c.memoryUsage = c.RegisterGauge(
		"memory_usage_bytes",
		"Memory usage in bytes",
		Labels{},
	)

	c.cpuUsage = c.RegisterGauge(
		"cpu_usage_percent",
		"CPU usage percentage",
		Labels{},
	)

	c.goroutineCount = c.RegisterGauge(
		"goroutines_total",
		"Number of goroutines",
		Labels{},
	)
}

// RegisterCounter はカウンターを登録する
func (c *DefaultCollector) RegisterCounter(name, help string, labels Labels) *Counter {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.generateKey(name, labels)
	if existing, exists := c.counters[key]; exists {
		return existing
	}

	counter := NewCounter(name, help, labels)
	c.counters[key] = counter
	return counter
}

// RegisterGauge はゲージを登録する
func (c *DefaultCollector) RegisterGauge(name, help string, labels Labels) *Gauge {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.generateKey(name, labels)
	if existing, exists := c.gauges[key]; exists {
		return existing
	}

	gauge := NewGauge(name, help, labels)
	c.gauges[key] = gauge
	return gauge
}

// RegisterHistogram はヒストグラムを登録する
func (c *DefaultCollector) RegisterHistogram(name, help string, labels Labels, buckets []float64) *Histogram {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.generateKey(name, labels)
	if existing, exists := c.histograms[key]; exists {
		return existing
	}

	histogram := NewHistogram(name, help, labels, buckets)
	c.histograms[key] = histogram
	return histogram
}

// RegisterSummary はサマリーを登録する
func (c *DefaultCollector) RegisterSummary(name, help string, labels Labels, objectives map[float64]float64) *Summary {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	key := c.generateKey(name, labels)
	if existing, exists := c.summaries[key]; exists {
		return existing
	}

	summary := NewSummary(name, help, labels, objectives)
	c.summaries[key] = summary
	return summary
}

// GetMetrics は全てのメトリクスを取得する
func (c *DefaultCollector) GetMetrics() []Metric {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	var metrics []Metric

	// カウンター
	for _, counter := range c.counters {
		metrics = append(metrics, counter.ToMetric())
	}

	// ゲージ
	for _, gauge := range c.gauges {
		metrics = append(metrics, gauge.ToMetric())
	}

	// ヒストグラム
	for _, histogram := range c.histograms {
		metrics = append(metrics, histogram.ToMetric())
	}

	// サマリー
	for _, summary := range c.summaries {
		metrics = append(metrics, summary.ToMetric())
	}

	return metrics
}

// GetMetricsByType は指定されたタイプのメトリクスを取得する
func (c *DefaultCollector) GetMetricsByType(metricType MetricType) []Metric {
	metrics := c.GetMetrics()
	var filtered []Metric

	for _, metric := range metrics {
		if metric.Type == metricType {
			filtered = append(filtered, metric)
		}
	}

	return filtered
}

// GetMetricsByName は指定された名前のメトリクスを取得する
func (c *DefaultCollector) GetMetricsByName(name string) []Metric {
	metrics := c.GetMetrics()
	var filtered []Metric

	for _, metric := range metrics {
		if metric.Name == name {
			filtered = append(filtered, metric)
		}
	}

	return filtered
}

// RecordHTTPRequest はHTTPリクエストのメトリクスを記録する
func (c *DefaultCollector) RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	labels := Labels{
		"method": method,
		"path":   path,
		"status": fmt.Sprintf("%d", statusCode),
	}

	// リクエスト総数
	counter := c.RegisterCounter("http_requests_total", "Total HTTP requests", labels)
	counter.Inc()

	// レスポンス時間
	histogram := c.RegisterHistogram(
		"http_request_duration_seconds",
		"HTTP request duration",
		labels,
		[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	)
	histogram.Observe(duration.Seconds())

	// リクエストサイズ
	if requestSize > 0 {
		reqSizeHist := c.RegisterHistogram(
			"http_request_size_bytes",
			"HTTP request size",
			labels,
			[]float64{100, 1000, 10000, 100000, 1000000},
		)
		reqSizeHist.Observe(float64(requestSize))
	}

	// レスポンスサイズ
	if responseSize > 0 {
		respSizeHist := c.RegisterHistogram(
			"http_response_size_bytes",
			"HTTP response size",
			labels,
			[]float64{100, 1000, 10000, 100000, 1000000},
		)
		respSizeHist.Observe(float64(responseSize))
	}

	// エラー数
	if statusCode >= 400 {
		errorCounter := c.RegisterCounter("http_errors_total", "Total HTTP errors", labels)
		errorCounter.Inc()
	}
}

// RecordDBQuery はデータベースクエリのメトリクスを記録する
func (c *DefaultCollector) RecordDBQuery(operation string, duration time.Duration, success bool) {
	labels := Labels{
		"operation": operation,
		"success":   fmt.Sprintf("%t", success),
	}

	// クエリ総数
	counter := c.RegisterCounter("db_queries_total", "Total database queries", labels)
	counter.Inc()

	// クエリ実行時間
	histogram := c.RegisterHistogram(
		"db_query_duration_seconds",
		"Database query duration",
		labels,
		[]float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
	)
	histogram.Observe(duration.Seconds())

	// エラー数
	if !success {
		errorCounter := c.RegisterCounter("db_errors_total", "Total database errors", labels)
		errorCounter.Inc()
	}
}

// RecordWebSocketConnection はWebSocket接続のメトリクスを記録する
func (c *DefaultCollector) RecordWebSocketConnection(action string) {
	switch action {
	case "connect":
		c.websocketConnections.Inc()
	case "disconnect":
		c.websocketConnections.Dec()
	}
}

// SetActiveUsers はアクティブユーザー数を設定する
func (c *DefaultCollector) SetActiveUsers(count int) {
	c.activeUsers.Set(float64(count))
}

// SetTournamentCount はトーナメント数を設定する
func (c *DefaultCollector) SetTournamentCount(count int) {
	c.tournamentCount.Set(float64(count))
}

// SetMatchCount は試合数を設定する
func (c *DefaultCollector) SetMatchCount(count int) {
	c.matchCount.Set(float64(count))
}

// Collect はメトリクスを収集する
func (c *DefaultCollector) Collect(ctx context.Context) []Metric {
	// システムメトリクスを更新
	c.updateSystemMetrics()

	return c.GetMetrics()
}

// updateSystemMetrics はシステムメトリクスを更新する
func (c *DefaultCollector) updateSystemMetrics() {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// メモリ使用量
	c.memoryUsage.Set(float64(m.Alloc))

	// Goroutine数
	c.goroutineCount.Set(float64(runtime.NumGoroutine()))

	// CPU使用量は簡易実装（実際の実装では適切なライブラリを使用）
	// ここでは固定値を設定
	c.cpuUsage.Set(0.0)
}

// generateKey はメトリクスのキーを生成する
func (c *DefaultCollector) generateKey(name string, labels Labels) string {
	key := name
	for k, v := range labels {
		key += fmt.Sprintf("_%s_%s", k, v)
	}
	return key
}

// グローバルコレクター
var globalCollector MetricsCollector

// Init はグローバルメトリクスコレクターを初期化する
func Init() {
	globalCollector = NewDefaultCollector()
}

// GetCollector はグローバルメトリクスコレクターを取得する
func GetCollector() MetricsCollector {
	if globalCollector == nil {
		Init()
	}
	return globalCollector
}

// 便利関数
func RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, requestSize, responseSize int64) {
	GetCollector().RecordHTTPRequest(method, path, statusCode, duration, requestSize, responseSize)
}

func RecordDBQuery(operation string, duration time.Duration, success bool) {
	GetCollector().RecordDBQuery(operation, duration, success)
}

func RecordWebSocketConnection(action string) {
	GetCollector().RecordWebSocketConnection(action)
}

func SetActiveUsers(count int) {
	GetCollector().SetActiveUsers(count)
}

func SetTournamentCount(count int) {
	GetCollector().SetTournamentCount(count)
}

func SetMatchCount(count int) {
	GetCollector().SetMatchCount(count)
}
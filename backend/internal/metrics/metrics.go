package metrics

import (
	"context"
	"sync"
	"time"
)

// MetricType はメトリクスの種類を表す
type MetricType string

const (
	// HTTP関連メトリクス
	HTTPRequestTotal    MetricType = "http_request_total"
	HTTPRequestDuration MetricType = "http_request_duration"
	HTTPRequestSize     MetricType = "http_request_size"
	HTTPResponseSize    MetricType = "http_response_size"
	HTTPErrorTotal      MetricType = "http_error_total"

	// データベース関連メトリクス
	DBConnectionTotal    MetricType = "db_connection_total"
	DBQueryDuration      MetricType = "db_query_duration"
	DBQueryTotal         MetricType = "db_query_total"
	DBTransactionTotal   MetricType = "db_transaction_total"
	DBConnectionPoolSize MetricType = "db_connection_pool_size"

	// アプリケーション関連メトリクス
	ActiveUsers          MetricType = "active_users"
	TournamentTotal      MetricType = "tournament_total"
	MatchTotal           MetricType = "match_total"
	WebSocketConnections MetricType = "websocket_connections"

	// システム関連メトリクス
	MemoryUsage    MetricType = "memory_usage"
	CPUUsage       MetricType = "cpu_usage"
	GoroutineCount MetricType = "goroutine_count"
)

// Labels はメトリクスのラベルを表す
type Labels map[string]string

// Metric はメトリクスデータを表す構造体
type Metric struct {
	Type      MetricType             `json:"type"`
	Name      string                 `json:"name"`
	Value     float64                `json:"value"`
	Labels    Labels                 `json:"labels"`
	Timestamp time.Time              `json:"timestamp"`
	Unit      string                 `json:"unit"`
	Help      string                 `json:"help"`
	Tags      map[string]interface{} `json:"tags,omitempty"`
}

// Counter はカウンターメトリクス
type Counter struct {
	name   string
	help   string
	labels Labels
	value  float64
	mutex  sync.RWMutex
}

// NewCounter は新しいCounterを作成する
func NewCounter(name, help string, labels Labels) *Counter {
	return &Counter{
		name:   name,
		help:   help,
		labels: labels,
		value:  0,
	}
}

// Inc はカウンターを1増加させる
func (c *Counter) Inc() {
	c.Add(1)
}

// Add はカウンターに値を追加する
func (c *Counter) Add(value float64) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.value += value
}

// Get は現在の値を取得する
func (c *Counter) Get() float64 {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.value
}

// ToMetric はMetric構造体に変換する
func (c *Counter) ToMetric() Metric {
	return Metric{
		Name:      c.name,
		Value:     c.Get(),
		Labels:    c.labels,
		Timestamp: time.Now(),
		Unit:      "count",
		Help:      c.help,
	}
}

// Gauge はゲージメトリクス
type Gauge struct {
	name   string
	help   string
	labels Labels
	value  float64
	mutex  sync.RWMutex
}

// NewGauge は新しいGaugeを作成する
func NewGauge(name, help string, labels Labels) *Gauge {
	return &Gauge{
		name:   name,
		help:   help,
		labels: labels,
		value:  0,
	}
}

// Set はゲージの値を設定する
func (g *Gauge) Set(value float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value = value
}

// Inc はゲージを1増加させる
func (g *Gauge) Inc() {
	g.Add(1)
}

// Dec はゲージを1減少させる
func (g *Gauge) Dec() {
	g.Add(-1)
}

// Add はゲージに値を追加する
func (g *Gauge) Add(value float64) {
	g.mutex.Lock()
	defer g.mutex.Unlock()
	g.value += value
}

// Get は現在の値を取得する
func (g *Gauge) Get() float64 {
	g.mutex.RLock()
	defer g.mutex.RUnlock()
	return g.value
}

// ToMetric はMetric構造体に変換する
func (g *Gauge) ToMetric() Metric {
	return Metric{
		Name:      g.name,
		Value:     g.Get(),
		Labels:    g.labels,
		Timestamp: time.Now(),
		Unit:      "gauge",
		Help:      g.help,
	}
}

// Histogram はヒストグラムメトリクス
type Histogram struct {
	name    string
	help    string
	labels  Labels
	buckets []float64
	counts  []uint64
	sum     float64
	count   uint64
	mutex   sync.RWMutex
}

// NewHistogram は新しいHistogramを作成する
func NewHistogram(name, help string, labels Labels, buckets []float64) *Histogram {
	return &Histogram{
		name:    name,
		help:    help,
		labels:  labels,
		buckets: buckets,
		counts:  make([]uint64, len(buckets)+1), // +1 for +Inf bucket
	}
}

// Observe は値を観測する
func (h *Histogram) Observe(value float64) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.sum += value
	h.count++

	// バケットに値を分類
	for i, bucket := range h.buckets {
		if value <= bucket {
			h.counts[i]++
		}
	}
	// +Inf bucket
	h.counts[len(h.buckets)]++
}

// GetSum は合計値を取得する
func (h *Histogram) GetSum() float64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.sum
}

// GetCount は観測回数を取得する
func (h *Histogram) GetCount() uint64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.count
}

// GetBuckets はバケット情報を取得する
func (h *Histogram) GetBuckets() map[string]uint64 {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	buckets := make(map[string]uint64)
	for i, bucket := range h.buckets {
		buckets[formatFloat(bucket)] = h.counts[i]
	}
	buckets["+Inf"] = h.counts[len(h.buckets)]
	return buckets
}

// ToMetric はMetric構造体に変換する
func (h *Histogram) ToMetric() Metric {
	return Metric{
		Name:      h.name,
		Value:     h.GetSum(),
		Labels:    h.labels,
		Timestamp: time.Now(),
		Unit:      "histogram",
		Help:      h.help,
		Tags: map[string]interface{}{
			"count":   h.GetCount(),
			"buckets": h.GetBuckets(),
		},
	}
}

// Summary はサマリーメトリクス
type Summary struct {
	name       string
	help       string
	labels     Labels
	objectives map[float64]float64 // quantile -> error
	values     []float64
	sum        float64
	count      uint64
	mutex      sync.RWMutex
}

// NewSummary は新しいSummaryを作成する
func NewSummary(name, help string, labels Labels, objectives map[float64]float64) *Summary {
	return &Summary{
		name:       name,
		help:       help,
		labels:     labels,
		objectives: objectives,
		values:     make([]float64, 0),
	}
}

// Observe は値を観測する
func (s *Summary) Observe(value float64) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.values = append(s.values, value)
	s.sum += value
	s.count++

	// メモリ使用量を制限するため、古い値を削除
	if len(s.values) > 1000 {
		s.values = s.values[len(s.values)-1000:]
	}
}

// GetSum は合計値を取得する
func (s *Summary) GetSum() float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.sum
}

// GetCount は観測回数を取得する
func (s *Summary) GetCount() uint64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.count
}

// GetQuantiles は分位数を取得する
func (s *Summary) GetQuantiles() map[float64]float64 {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if len(s.values) == 0 {
		return make(map[float64]float64)
	}

	// 値をソート
	sorted := make([]float64, len(s.values))
	copy(sorted, s.values)
	
	// 簡単なソート（実際の実装では効率的なアルゴリズムを使用）
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	quantiles := make(map[float64]float64)
	for q := range s.objectives {
		index := int(q * float64(len(sorted)-1))
		if index >= len(sorted) {
			index = len(sorted) - 1
		}
		quantiles[q] = sorted[index]
	}

	return quantiles
}

// ToMetric はMetric構造体に変換する
func (s *Summary) ToMetric() Metric {
	return Metric{
		Name:      s.name,
		Value:     s.GetSum(),
		Labels:    s.labels,
		Timestamp: time.Now(),
		Unit:      "summary",
		Help:      s.help,
		Tags: map[string]interface{}{
			"count":     s.GetCount(),
			"quantiles": s.GetQuantiles(),
		},
	}
}

// MetricsCollector はメトリクス収集器のインターフェース
type MetricsCollector interface {
	RegisterCounter(name, help string, labels Labels) *Counter
	RegisterGauge(name, help string, labels Labels) *Gauge
	RegisterHistogram(name, help string, labels Labels, buckets []float64) *Histogram
	RegisterSummary(name, help string, labels Labels, objectives map[float64]float64) *Summary
	GetMetrics() []Metric
	GetMetricsByType(metricType MetricType) []Metric
	GetMetricsByName(name string) []Metric
	RecordHTTPRequest(method, path string, statusCode int, duration time.Duration, requestSize, responseSize int64)
	RecordDBQuery(operation string, duration time.Duration, success bool)
	RecordWebSocketConnection(action string) // "connect" or "disconnect"
	SetActiveUsers(count int)
	SetTournamentCount(count int)
	SetMatchCount(count int)
	Collect(ctx context.Context) []Metric
}

// formatFloat は浮動小数点数を文字列に変換する
func formatFloat(f float64) string {
	if f == float64(int64(f)) {
		return string(rune(int64(f)))
	}
	return string(rune(f))
}
package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCounter(t *testing.T) {
	counter := NewCounter("test_counter", "Test counter", Labels{"env": "test"})

	// 初期値は0
	assert.Equal(t, 0.0, counter.Get())

	// Inc()で1増加
	counter.Inc()
	assert.Equal(t, 1.0, counter.Get())

	// Add()で値を追加
	counter.Add(5.5)
	assert.Equal(t, 6.5, counter.Get())

	// ToMetric()でMetric構造体に変換
	metric := counter.ToMetric()
	assert.Equal(t, "test_counter", metric.Name)
	assert.Equal(t, 6.5, metric.Value)
	assert.Equal(t, "count", metric.Unit)
	assert.Equal(t, "test", metric.Labels["env"])
}

func TestGauge(t *testing.T) {
	gauge := NewGauge("test_gauge", "Test gauge", Labels{"type": "memory"})

	// 初期値は0
	assert.Equal(t, 0.0, gauge.Get())

	// Set()で値を設定
	gauge.Set(100.0)
	assert.Equal(t, 100.0, gauge.Get())

	// Inc()で1増加
	gauge.Inc()
	assert.Equal(t, 101.0, gauge.Get())

	// Dec()で1減少
	gauge.Dec()
	assert.Equal(t, 100.0, gauge.Get())

	// Add()で値を追加
	gauge.Add(-10.0)
	assert.Equal(t, 90.0, gauge.Get())

	// ToMetric()でMetric構造体に変換
	metric := gauge.ToMetric()
	assert.Equal(t, "test_gauge", metric.Name)
	assert.Equal(t, 90.0, metric.Value)
	assert.Equal(t, "gauge", metric.Unit)
	assert.Equal(t, "memory", metric.Labels["type"])
}

func TestHistogram(t *testing.T) {
	buckets := []float64{0.1, 0.5, 1.0, 2.5, 5.0}
	histogram := NewHistogram("test_histogram", "Test histogram", Labels{"method": "GET"}, buckets)

	// 値を観測
	histogram.Observe(0.05) // bucket 0.1
	histogram.Observe(0.3)  // bucket 0.5
	histogram.Observe(1.5)  // bucket 2.5
	histogram.Observe(10.0) // +Inf bucket

	// 合計値と観測回数をチェック
	assert.Equal(t, 4.0, histogram.GetSum())
	assert.Equal(t, uint64(4), histogram.GetCount())

	// バケット情報をチェック
	bucketCounts := histogram.GetBuckets()
	assert.Contains(t, bucketCounts, "+Inf")

	// ToMetric()でMetric構造体に変換
	metric := histogram.ToMetric()
	assert.Equal(t, "test_histogram", metric.Name)
	assert.Equal(t, 4.0, metric.Value) // sum
	assert.Equal(t, "histogram", metric.Unit)
	assert.Equal(t, "GET", metric.Labels["method"])
	assert.Contains(t, metric.Tags, "count")
	assert.Contains(t, metric.Tags, "buckets")
}

func TestSummary(t *testing.T) {
	objectives := map[float64]float64{
		0.5:  0.05,
		0.9:  0.01,
		0.99: 0.001,
	}
	summary := NewSummary("test_summary", "Test summary", Labels{"endpoint": "/api"}, objectives)

	// 値を観測
	values := []float64{1.0, 2.0, 3.0, 4.0, 5.0}
	for _, v := range values {
		summary.Observe(v)
	}

	// 合計値と観測回数をチェック
	assert.Equal(t, 15.0, summary.GetSum())
	assert.Equal(t, uint64(5), summary.GetCount())

	// 分位数をチェック
	quantiles := summary.GetQuantiles()
	assert.Contains(t, quantiles, 0.5)
	assert.Contains(t, quantiles, 0.9)
	assert.Contains(t, quantiles, 0.99)

	// ToMetric()でMetric構造体に変換
	metric := summary.ToMetric()
	assert.Equal(t, "test_summary", metric.Name)
	assert.Equal(t, 15.0, metric.Value) // sum
	assert.Equal(t, "summary", metric.Unit)
	assert.Equal(t, "/api", metric.Labels["endpoint"])
	assert.Contains(t, metric.Tags, "count")
	assert.Contains(t, metric.Tags, "quantiles")
}

func TestDefaultCollector(t *testing.T) {
	collector := NewDefaultCollector()

	// カウンターの登録
	counter := collector.RegisterCounter("test_counter", "Test counter", Labels{"app": "test"})
	assert.NotNil(t, counter)

	// 同じ名前とラベルで再登録すると同じインスタンスが返される
	counter2 := collector.RegisterCounter("test_counter", "Test counter", Labels{"app": "test"})
	assert.Equal(t, counter, counter2)

	// ゲージの登録
	gauge := collector.RegisterGauge("test_gauge", "Test gauge", Labels{"type": "cpu"})
	assert.NotNil(t, gauge)

	// ヒストグラムの登録
	histogram := collector.RegisterHistogram("test_histogram", "Test histogram", Labels{}, []float64{0.1, 1.0, 10.0})
	assert.NotNil(t, histogram)

	// サマリーの登録
	summary := collector.RegisterSummary("test_summary", "Test summary", Labels{}, map[float64]float64{0.5: 0.05})
	assert.NotNil(t, summary)

	// メトリクスの取得
	metrics := collector.GetMetrics()
	assert.True(t, len(metrics) > 0)
}

func TestCollectorHTTPMetrics(t *testing.T) {
	collector := NewDefaultCollector()

	// HTTPリクエストメトリクスを記録
	collector.RecordHTTPRequest("GET", "/api/users", 200, 100*time.Millisecond, 1024, 2048)
	collector.RecordHTTPRequest("POST", "/api/users", 201, 200*time.Millisecond, 2048, 1024)
	collector.RecordHTTPRequest("GET", "/api/users", 404, 50*time.Millisecond, 512, 256)

	// メトリクスを取得
	metrics := collector.GetMetrics()
	assert.True(t, len(metrics) > 0)

	// 特定のメトリクスが存在することを確認
	found := false
	for _, metric := range metrics {
		if metric.Name == "http_requests_total" {
			found = true
			break
		}
	}
	assert.True(t, found, "http_requests_total メトリクスが見つかりません")
}

func TestCollectorDBMetrics(t *testing.T) {
	collector := NewDefaultCollector()

	// データベースクエリメトリクスを記録
	collector.RecordDBQuery("SELECT", 50*time.Millisecond, true)
	collector.RecordDBQuery("INSERT", 100*time.Millisecond, true)
	collector.RecordDBQuery("UPDATE", 75*time.Millisecond, false) // エラー

	// メトリクスを取得
	metrics := collector.GetMetrics()
	assert.True(t, len(metrics) > 0)

	// データベースメトリクスが存在することを確認
	found := false
	for _, metric := range metrics {
		if metric.Name == "db_queries_total" {
			found = true
			break
		}
	}
	assert.True(t, found, "db_queries_total メトリクスが見つかりません")
}

func TestCollectorWebSocketMetrics(t *testing.T) {
	collector := NewDefaultCollector()

	// WebSocket接続メトリクスを記録
	collector.RecordWebSocketConnection("connect")
	collector.RecordWebSocketConnection("connect")
	collector.RecordWebSocketConnection("disconnect")

	// メトリクスを取得
	metrics := collector.GetMetrics()
	
	// WebSocket接続数メトリクスを確認
	found := false
	for _, metric := range metrics {
		if metric.Name == "websocket_connections" {
			found = true
			assert.Equal(t, 1.0, metric.Value) // 2接続 - 1切断 = 1
			break
		}
	}
	assert.True(t, found, "websocket_connections メトリクスが見つかりません")
}

func TestCollectorApplicationMetrics(t *testing.T) {
	collector := NewDefaultCollector()

	// アプリケーションメトリクスを設定
	collector.SetActiveUsers(100)
	collector.SetTournamentCount(25)
	collector.SetMatchCount(150)

	// メトリクスを取得
	metrics := collector.GetMetrics()

	// 各メトリクスの値を確認
	metricsMap := make(map[string]float64)
	for _, metric := range metrics {
		metricsMap[metric.Name] = metric.Value
	}

	assert.Equal(t, 100.0, metricsMap["active_users"])
	assert.Equal(t, 25.0, metricsMap["tournaments_total"])
	assert.Equal(t, 150.0, metricsMap["matches_total"])
}

func TestGlobalCollector(t *testing.T) {
	// グローバルコレクターをリセット
	globalCollector = nil

	// 初回取得時に初期化される
	collector := GetCollector()
	assert.NotNil(t, collector)

	// 2回目の取得で同じインスタンスが返される
	collector2 := GetCollector()
	assert.Equal(t, collector, collector2)

	// 便利関数のテスト
	RecordHTTPRequest("GET", "/test", 200, 100*time.Millisecond, 1024, 2048)
	RecordDBQuery("SELECT", 50*time.Millisecond, true)
	RecordWebSocketConnection("connect")
	SetActiveUsers(50)
	SetTournamentCount(10)
	SetMatchCount(75)

	// メトリクスが記録されていることを確認
	metrics := collector.GetMetrics()
	assert.True(t, len(metrics) > 0)
}

func TestMetricsFiltering(t *testing.T) {
	collector := NewDefaultCollector()

	// 複数のメトリクスを登録
	counter1 := collector.RegisterCounter("counter1", "Counter 1", Labels{"type": "test"})
	counter2 := collector.RegisterCounter("counter2", "Counter 2", Labels{"type": "prod"})
	gauge1 := collector.RegisterGauge("gauge1", "Gauge 1", Labels{"type": "test"})

	counter1.Inc()
	counter2.Add(5)
	gauge1.Set(10)

	// 名前でフィルタリング
	counter1Metrics := collector.GetMetricsByName("counter1")
	assert.Len(t, counter1Metrics, 1)
	assert.Equal(t, "counter1", counter1Metrics[0].Name)

	// 全メトリクス取得
	allMetrics := collector.GetMetrics()
	assert.True(t, len(allMetrics) >= 3)
}
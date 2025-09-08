package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsMiddleware はHTTPメトリクスを自動収集するミドルウェア
func MetricsMiddleware() gin.HandlerFunc {
	return MetricsMiddlewareWithCollector(GetCollector())
}

// MetricsMiddlewareWithCollector は指定されたコレクターでメトリクスを収集するミドルウェア
func MetricsMiddlewareWithCollector(collector MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// リクエストサイズを取得
		requestSize := c.Request.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		c.Next()

		// レスポンス情報を取得
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		responseSize := int64(c.Writer.Size())

		// メトリクスを記録
		collector.RecordHTTPRequest(
			method,
			path,
			statusCode,
			duration,
			requestSize,
			responseSize,
		)

		// 追加のメトリクス情報をコンテキストに設定（ログで使用可能）
		c.Set("metrics_duration", duration)
		c.Set("metrics_status_code", statusCode)
		c.Set("metrics_request_size", requestSize)
		c.Set("metrics_response_size", responseSize)
	}
}

// MetricsConfig はメトリクスミドルウェアの設定
type MetricsConfig struct {
	SkipPaths      []string // メトリクス収集をスキップするパス
	PathNormalizer func(string) string // パスを正規化する関数
}

// DefaultMetricsConfig はデフォルトのメトリクス設定を返す
func DefaultMetricsConfig() MetricsConfig {
	return MetricsConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		PathNormalizer: func(path string) string {
			// デフォルトではパスをそのまま返す
			return path
		},
	}
}

// MetricsMiddlewareWithConfig は設定を指定してメトリクスミドルウェアを作成する
func MetricsMiddlewareWithConfig(config MetricsConfig, collector MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// スキップパスのチェック
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}

		// パスの正規化
		normalizedPath := config.PathNormalizer(path)

		// リクエストサイズを取得
		requestSize := c.Request.ContentLength
		if requestSize < 0 {
			requestSize = 0
		}

		c.Next()

		// レスポンス情報を取得
		duration := time.Since(start)
		statusCode := c.Writer.Status()
		responseSize := int64(c.Writer.Size())

		// メトリクスを記録
		collector.RecordHTTPRequest(
			method,
			normalizedPath,
			statusCode,
			duration,
			requestSize,
			responseSize,
		)

		// 追加のメトリクス情報をコンテキストに設定
		c.Set("metrics_duration", duration)
		c.Set("metrics_status_code", statusCode)
		c.Set("metrics_request_size", requestSize)
		c.Set("metrics_response_size", responseSize)
		c.Set("metrics_normalized_path", normalizedPath)
	}
}

// DatabaseMetricsWrapper はデータベース操作のメトリクスを収集するラッパー
type DatabaseMetricsWrapper struct {
	collector MetricsCollector
}

// NewDatabaseMetricsWrapper は新しいDatabaseMetricsWrapperを作成する
func NewDatabaseMetricsWrapper(collector MetricsCollector) *DatabaseMetricsWrapper {
	return &DatabaseMetricsWrapper{
		collector: collector,
	}
}

// WrapQuery はクエリ実行をラップしてメトリクスを収集する
func (w *DatabaseMetricsWrapper) WrapQuery(operation string, fn func() error) error {
	start := time.Now()
	err := fn()
	duration := time.Since(start)
	success := err == nil

	w.collector.RecordDBQuery(operation, duration, success)
	return err
}

// WrapQueryWithResult はクエリ実行をラップしてメトリクスを収集する（結果付き）
func (w *DatabaseMetricsWrapper) WrapQueryWithResult(operation string, fn func() (interface{}, error)) (interface{}, error) {
	start := time.Now()
	result, err := fn()
	duration := time.Since(start)
	success := err == nil

	w.collector.RecordDBQuery(operation, duration, success)
	return result, err
}

// WebSocketMetricsWrapper はWebSocket接続のメトリクスを収集するラッパー
type WebSocketMetricsWrapper struct {
	collector MetricsCollector
}

// NewWebSocketMetricsWrapper は新しいWebSocketMetricsWrapperを作成する
func NewWebSocketMetricsWrapper(collector MetricsCollector) *WebSocketMetricsWrapper {
	return &WebSocketMetricsWrapper{
		collector: collector,
	}
}

// OnConnect は接続時にメトリクスを記録する
func (w *WebSocketMetricsWrapper) OnConnect() {
	w.collector.RecordWebSocketConnection("connect")
}

// OnDisconnect は切断時にメトリクスを記録する
func (w *WebSocketMetricsWrapper) OnDisconnect() {
	w.collector.RecordWebSocketConnection("disconnect")
}

// PathNormalizers は一般的なパス正規化関数
var PathNormalizers = struct {
	// IDパラメータを正規化する
	NormalizeIDs func(string) string
	// クエリパラメータを除去する
	RemoveQuery func(string) string
}{
	NormalizeIDs: func(path string) string {
		// 簡易実装: 数値IDを:idに置換
		// 実際の実装では正規表現を使用
		return path
	},
	RemoveQuery: func(path string) string {
		// クエリパラメータを除去
		if idx := len(path); idx > 0 {
			for i, char := range path {
				if char == '?' {
					return path[:i]
				}
			}
		}
		return path
	},
}

// MetricsHandler はメトリクス情報を返すHTTPハンドラー
func MetricsHandler() gin.HandlerFunc {
	return MetricsHandlerWithCollector(GetCollector())
}

// MetricsHandlerWithCollector は指定されたコレクターでメトリクス情報を返すHTTPハンドラー
func MetricsHandlerWithCollector(collector MetricsCollector) gin.HandlerFunc {
	return func(c *gin.Context) {
		metrics := collector.GetMetrics()
		
		// フォーマットパラメータをチェック
		format := c.DefaultQuery("format", "json")
		
		switch format {
		case "prometheus":
			// Prometheus形式での出力
			c.Header("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
			c.String(200, formatPrometheus(metrics))
		default:
			// JSON形式での出力
			c.JSON(200, gin.H{
				"metrics": metrics,
				"timestamp": time.Now(),
				"count": len(metrics),
			})
		}
	}
}

// formatPrometheus はメトリクスをPrometheus形式でフォーマットする
func formatPrometheus(metrics []Metric) string {
	var result string
	
	for _, metric := range metrics {
		// HELP行
		result += "# HELP " + metric.Name + " " + metric.Help + "\n"
		
		// TYPE行
		result += "# TYPE " + metric.Name + " " + metric.Unit + "\n"
		
		// メトリクス行
		result += metric.Name
		
		// ラベルの追加
		if len(metric.Labels) > 0 {
			result += "{"
			first := true
			for k, v := range metric.Labels {
				if !first {
					result += ","
				}
				result += k + "=\"" + v + "\""
				first = false
			}
			result += "}"
		}
		
		result += " " + strconv.FormatFloat(metric.Value, 'f', -1, 64)
		result += " " + strconv.FormatInt(metric.Timestamp.UnixMilli(), 10)
		result += "\n"
	}
	
	return result
}
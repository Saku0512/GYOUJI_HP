package logger

import (
	"bytes"
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDMiddleware はリクエストIDを生成し、コンテキストに設定するミドルウェア
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// リクエストIDを生成
		requestID := uuid.New().String()
		
		// ヘッダーからリクエストIDを取得（既に設定されている場合）
		if existingID := c.GetHeader("X-Request-ID"); existingID != "" {
			requestID = existingID
		}
		
		// コンテキストにリクエストIDを設定
		ctx := context.WithValue(c.Request.Context(), "request_id", requestID)
		c.Request = c.Request.WithContext(ctx)
		
		// Ginコンテキストにも設定（他のミドルウェアで使用可能）
		c.Set("request_id", requestID)
		
		// レスポンスヘッダーにリクエストIDを設定
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// responseWriter はレスポンスボディをキャプチャするためのラッパー
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// LoggingConfig はログミドルウェアの設定
type LoggingConfig struct {
	SkipPaths      []string // ログをスキップするパス
	LogRequestBody bool     // リクエストボディをログに記録するか
	LogResponseBody bool    // レスポンスボディをログに記録するか
	MaxBodySize    int64    // ログに記録する最大ボディサイズ
}

// DefaultLoggingConfig はデフォルトのログ設定を返す
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		LogRequestBody:  false,
		LogResponseBody: false,
		MaxBodySize:     1024, // 1KB
	}
}

// LoggingMiddleware はHTTPリクエストをログに記録するミドルウェア
func LoggingMiddleware() gin.HandlerFunc {
	return LoggingMiddlewareWithConfig(DefaultLoggingConfig())
}

// LoggingMiddlewareWithConfig は設定を指定してログミドルウェアを作成する
func LoggingMiddlewareWithConfig(config LoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		
		// スキップパスのチェック
		for _, skipPath := range config.SkipPaths {
			if path == skipPath {
				c.Next()
				return
			}
		}
		
		raw := c.Request.URL.RawQuery
		
		// リクエストIDを取得
		requestID := c.Request.Context().Value("request_id")
		logger := GetLogger().WithComponent("http")
		if requestID != nil {
			logger = logger.WithRequestID(requestID.(string))
		}
		
		// 認証情報をコンテキストから取得してログに追加
		logger = logger.WithContext(c.Request.Context())
		
		// リクエストボディの読み取り（必要な場合）
		var requestBody string
		if config.LogRequestBody && c.Request.Body != nil {
			bodyBytes, err := io.ReadAll(io.LimitReader(c.Request.Body, config.MaxBodySize))
			if err == nil {
				requestBody = string(bodyBytes)
				// ボディを復元
				c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
		
		// レスポンスボディをキャプチャ（必要な場合）
		var responseBody *bytes.Buffer
		if config.LogResponseBody {
			responseBody = &bytes.Buffer{}
			c.Writer = &responseWriter{
				ResponseWriter: c.Writer,
				body:          responseBody,
			}
		}
		
		// リクエスト開始ログ
		requestFields := []Field{
			String("event", "request_start"),
			String("method", c.Request.Method),
			String("path", path),
			String("query", raw),
			String("user_agent", c.Request.UserAgent()),
			String("client_ip", c.ClientIP()),
			String("remote_addr", c.Request.RemoteAddr),
			String("referer", c.Request.Referer()),
			Int64("content_length", c.Request.ContentLength),
		}
		
		if requestBody != "" {
			requestFields = append(requestFields, String("request_body", requestBody))
		}
		
		logger.Info("HTTPリクエスト開始", requestFields...)
		
		c.Next()
		
		// レスポンス完了ログ
		latency := time.Since(start)
		status := c.Writer.Status()
		responseSize := c.Writer.Size()
		
		// ログレベルの決定
		logLevel := "info"
		if status >= 400 && status < 500 {
			logLevel = "warn"
		} else if status >= 500 {
			logLevel = "error"
		}
		
		responseFields := []Field{
			String("event", "request_complete"),
			String("method", c.Request.Method),
			String("path", path),
			StatusCode(status),
			Latency(latency),
			Int("response_size", responseSize),
			Float64("latency_seconds", latency.Seconds()),
		}
		
		// エラー情報の追加
		if len(c.Errors) > 0 {
			errorMessages := make([]string, len(c.Errors))
			for i, err := range c.Errors {
				errorMessages[i] = err.Error()
			}
			responseFields = append(responseFields, String("errors", strings.Join(errorMessages, "; ")))
		}
		
		// レスポンスボディの追加（必要な場合）
		if responseBody != nil && responseBody.Len() > 0 {
			body := responseBody.String()
			if len(body) > int(config.MaxBodySize) {
				body = body[:config.MaxBodySize] + "..."
			}
			responseFields = append(responseFields, String("response_body", body))
		}
		
		// パフォーマンス分析用の追加フィールド
		if latency > time.Second {
			responseFields = append(responseFields, Bool("slow_request", true))
		}
		
		switch logLevel {
		case "warn":
			logger.Warn("HTTPリクエスト完了", responseFields...)
		case "error":
			logger.Error("HTTPリクエスト完了", responseFields...)
		default:
			logger.Info("HTTPリクエスト完了", responseFields...)
		}
		
		// メトリクス用のログ（後でメトリクス収集に使用）
		if status >= 200 && status < 300 {
			logger.Debug("メトリクス: 成功リクエスト",
				String("metric_type", "http_request_success"),
				String("endpoint", path),
				String("method", c.Request.Method),
				Latency(latency),
			)
		} else if status >= 400 {
			logger.Debug("メトリクス: エラーリクエスト",
				String("metric_type", "http_request_error"),
				String("endpoint", path),
				String("method", c.Request.Method),
				StatusCode(status),
				Latency(latency),
			)
		}
	}
}

// ErrorLoggingMiddleware はエラーを構造化ログに記録するミドルウェア
func ErrorLoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		
		// エラーが発生した場合のログ記録
		if len(c.Errors) > 0 {
			logger := GetLogger().WithComponent("error_handler").WithContext(c.Request.Context())
			
			for _, ginErr := range c.Errors {
				fields := []Field{
					String("error_type", ginErr.Type.String()),
					String("error_message", ginErr.Error()),
					String("path", c.Request.URL.Path),
					String("method", c.Request.Method),
				}
				
				// エラーの詳細情報を追加
				if ginErr.Meta != nil {
					fields = append(fields, Any("error_meta", ginErr.Meta))
				}
				
				logger.Error("アプリケーションエラー", fields...)
			}
		}
	}
}

// PanicRecoveryMiddleware はパニックを捕捉してログに記録するミドルウェア
func PanicRecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		logger := GetLogger().WithComponent("panic_recovery").WithContext(c.Request.Context())
		
		logger.Error("パニックが発生しました",
			Any("panic", recovered),
			String("path", c.Request.URL.Path),
			String("method", c.Request.Method),
			Stack(),
		)
		
		c.AbortWithStatus(500)
	})
}
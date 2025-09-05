package logger

import (
	"context"
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
		
		// レスポンスヘッダーにリクエストIDを設定
		c.Header("X-Request-ID", requestID)
		
		c.Next()
	}
}

// LoggingMiddleware はHTTPリクエストをログに記録するミドルウェア
func LoggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		
		// リクエストIDを取得
		requestID := c.Request.Context().Value("request_id")
		logger := GetLogger()
		if requestID != nil {
			logger = logger.WithRequestID(requestID.(string))
		}
		
		// リクエスト開始ログ
		logger.Info("HTTPリクエスト開始",
			String("method", c.Request.Method),
			String("path", path),
			String("query", raw),
			String("user_agent", c.Request.UserAgent()),
			String("client_ip", c.ClientIP()),
		)
		
		c.Next()
		
		// レスポンス完了ログ
		latency := time.Since(start)
		status := c.Writer.Status()
		
		logLevel := "info"
		if status >= 400 && status < 500 {
			logLevel = "warn"
		} else if status >= 500 {
			logLevel = "error"
		}
		
		fields := []Field{
			String("method", c.Request.Method),
			String("path", path),
			Int("status", status),
			Any("latency", latency),
			Int("response_size", c.Writer.Size()),
		}
		
		switch logLevel {
		case "warn":
			logger.Warn("HTTPリクエスト完了", fields...)
		case "error":
			logger.Error("HTTPリクエスト完了", fields...)
		default:
			logger.Info("HTTPリクエスト完了", fields...)
		}
	}
}
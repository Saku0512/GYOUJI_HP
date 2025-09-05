package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name     string
		env      string
		logLevel string
		wantJSON bool
	}{
		{
			name:     "開発環境でのテキスト形式",
			env:      "development",
			logLevel: "info",
			wantJSON: false,
		},
		{
			name:     "本番環境でのJSON形式",
			env:      "production",
			logLevel: "debug",
			wantJSON: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 環境変数を設定
			os.Setenv("GO_ENV", tt.env)
			os.Setenv("LOG_LEVEL", tt.logLevel)
			defer func() {
				os.Unsetenv("GO_ENV")
				os.Unsetenv("LOG_LEVEL")
			}()

			logger := NewLogger()
			assert.NotNil(t, logger)

			// 型アサーション
			logrusLogger, ok := logger.(*logrusLogger)
			assert.True(t, ok)

			// フォーマッターの確認
			if tt.wantJSON {
				_, ok := logrusLogger.logger.Formatter.(*logrus.JSONFormatter)
				assert.True(t, ok, "本番環境ではJSONフォーマッターが使用されるべき")
			} else {
				_, ok := logrusLogger.logger.Formatter.(*logrus.TextFormatter)
				assert.True(t, ok, "開発環境ではテキストフォーマッターが使用されるべき")
			}
		})
	}
}

func TestLoggerMethods(t *testing.T) {
	// テスト用のバッファを作成
	var buf bytes.Buffer
	
	// logrusロガーを作成してバッファに出力
	logrusLog := logrus.New()
	logrusLog.SetOutput(&buf)
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	logrusLog.SetLevel(logrus.DebugLevel) // デバッグレベルを設定
	
	logger := &logrusLogger{
		logger: logrusLog,
		entry:  logrus.NewEntry(logrusLog),
	}

	tests := []struct {
		name     string
		logFunc  func()
		wantMsg  string
		wantLevel string
	}{
		{
			name: "Debugログ",
			logFunc: func() {
				logger.Debug("デバッグメッセージ", String("key", "value"))
			},
			wantMsg:  "デバッグメッセージ",
			wantLevel: "debug",
		},
		{
			name: "Infoログ",
			logFunc: func() {
				logger.Info("情報メッセージ", Int("count", 42))
			},
			wantMsg:  "情報メッセージ",
			wantLevel: "info",
		},
		{
			name: "Warnログ",
			logFunc: func() {
				logger.Warn("警告メッセージ", String("warning", "test"))
			},
			wantMsg:  "警告メッセージ",
			wantLevel: "warning",
		},
		{
			name: "Errorログ",
			logFunc: func() {
				logger.Error("エラーメッセージ", Err(assert.AnError))
			},
			wantMsg:  "エラーメッセージ",
			wantLevel: "error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()

			// ログ出力を解析
			var logEntry map[string]interface{}
			err := json.Unmarshal(buf.Bytes(), &logEntry)
			assert.NoError(t, err)

			assert.Equal(t, tt.wantMsg, logEntry["msg"])
			assert.Equal(t, tt.wantLevel, logEntry["level"])
		})
	}
}

func TestWithRequestID(t *testing.T) {
	var buf bytes.Buffer
	
	logrusLog := logrus.New()
	logrusLog.SetOutput(&buf)
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	
	logger := &logrusLogger{
		logger: logrusLog,
		entry:  logrus.NewEntry(logrusLog),
	}

	requestID := "test-request-id"
	loggerWithID := logger.WithRequestID(requestID)
	
	loggerWithID.Info("テストメッセージ")

	// ログ出力を解析
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, requestID, logEntry["request_id"])
	assert.Equal(t, "テストメッセージ", logEntry["msg"])
}

func TestWithContext(t *testing.T) {
	var buf bytes.Buffer
	
	logrusLog := logrus.New()
	logrusLog.SetOutput(&buf)
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	
	logger := &logrusLogger{
		logger: logrusLog,
		entry:  logrus.NewEntry(logrusLog),
	}

	// コンテキストにリクエストIDを設定
	requestID := "context-request-id"
	ctx := context.WithValue(context.Background(), "request_id", requestID)
	
	loggerWithCtx := logger.WithContext(ctx)
	loggerWithCtx.Info("コンテキストテスト")

	// ログ出力を解析
	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, requestID, logEntry["request_id"])
	assert.Equal(t, "コンテキストテスト", logEntry["msg"])
}

func TestFieldHelpers(t *testing.T) {
	tests := []struct {
		name      string
		field     Field
		wantKey   string
		wantValue interface{}
	}{
		{
			name:      "Stringフィールド",
			field:     String("name", "test"),
			wantKey:   "name",
			wantValue: "test",
		},
		{
			name:      "Intフィールド",
			field:     Int("count", 42),
			wantKey:   "count",
			wantValue: 42,
		},
		{
			name:      "Errorフィールド",
			field:     Err(assert.AnError),
			wantKey:   "error",
			wantValue: assert.AnError,
		},
		{
			name:      "Anyフィールド",
			field:     Any("data", map[string]string{"key": "value"}),
			wantKey:   "data",
			wantValue: map[string]string{"key": "value"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKey, tt.field.Key)
			assert.Equal(t, tt.wantValue, tt.field.Value)
		})
	}
}

func TestGlobalLogger(t *testing.T) {
	// グローバルロガーをリセット
	globalLogger = nil
	
	logger := GetLogger()
	assert.NotNil(t, logger)
	
	// 2回目の呼び出しで同じインスタンスが返されることを確認
	logger2 := GetLogger()
	assert.Equal(t, logger, logger2)
}

func TestFieldsToMap(t *testing.T) {
	fields := []Field{
		String("name", "test"),
		Int("count", 42),
		Any("data", map[string]string{"key": "value"}),
	}

	result := fieldsToMap(fields)
	
	expected := map[string]interface{}{
		"name":  "test",
		"count": 42,
		"data":  map[string]string{"key": "value"},
	}

	assert.Equal(t, expected, result)
}
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
}func Test
NewLoggerWithConfig(t *testing.T) {
	tests := []struct {
		name   string
		config LogConfig
	}{
		{
			name: "JSON形式設定",
			config: LogConfig{
				Level:       "debug",
				Format:      "json",
				Output:      "stdout",
				ServiceName: "test-service",
				Version:     "1.0.0",
			},
		},
		{
			name: "テキスト形式設定",
			config: LogConfig{
				Level:       "info",
				Format:      "text",
				Output:      "stderr",
				ServiceName: "test-service",
				Version:     "2.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLoggerWithConfig(tt.config)
			assert.NotNil(t, logger)

			logrusLogger, ok := logger.(*logrusLogger)
			assert.True(t, ok)

			if tt.config.Format == "json" {
				_, ok := logrusLogger.logger.Formatter.(*logrus.JSONFormatter)
				assert.True(t, ok)
			} else {
				_, ok := logrusLogger.logger.Formatter.(*logrus.TextFormatter)
				assert.True(t, ok)
			}
		})
	}
}

func TestEnhancedFieldHelpers(t *testing.T) {
	tests := []struct {
		name      string
		field     Field
		wantKey   string
		wantValue interface{}
	}{
		{
			name:      "Int64フィールド",
			field:     Int64("id", 1234567890),
			wantKey:   "id",
			wantValue: int64(1234567890),
		},
		{
			name:      "Float64フィールド",
			field:     Float64("price", 99.99),
			wantKey:   "price",
			wantValue: 99.99,
		},
		{
			name:      "Boolフィールド",
			field:     Bool("active", true),
			wantKey:   "active",
			wantValue: true,
		},
		{
			name:      "Componentフィールド",
			field:     Component("auth"),
			wantKey:   "component",
			wantValue: "auth",
		},
		{
			name:      "UserIDフィールド",
			field:     UserID(123),
			wantKey:   "user_id",
			wantValue: 123,
		},
		{
			name:      "RequestIDフィールド",
			field:     RequestID("req-123"),
			wantKey:   "request_id",
			wantValue: "req-123",
		},
		{
			name:      "Methodフィールド",
			field:     Method("POST"),
			wantKey:   "method",
			wantValue: "POST",
		},
		{
			name:      "Pathフィールド",
			field:     Path("/api/users"),
			wantKey:   "path",
			wantValue: "/api/users",
		},
		{
			name:      "StatusCodeフィールド",
			field:     StatusCode(200),
			wantKey:   "status_code",
			wantValue: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantKey, tt.field.Key)
			assert.Equal(t, tt.wantValue, tt.field.Value)
		})
	}
}

func TestWithComponent(t *testing.T) {
	var buf bytes.Buffer
	
	logrusLog := logrus.New()
	logrusLog.SetOutput(&buf)
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	
	logger := &logrusLogger{
		logger: logrusLog,
		entry:  logrus.NewEntry(logrusLog),
	}

	componentLogger := logger.WithComponent("auth")
	componentLogger.Info("認証処理開始")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "auth", logEntry["component"])
	assert.Equal(t, "認証処理開始", logEntry["msg"])
}

func TestWithUserID(t *testing.T) {
	var buf bytes.Buffer
	
	logrusLog := logrus.New()
	logrusLog.SetOutput(&buf)
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	
	logger := &logrusLogger{
		logger: logrusLog,
		entry:  logrus.NewEntry(logrusLog),
	}

	userLogger := logger.WithUserID(123)
	userLogger.Info("ユーザー操作")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, float64(123), logEntry["user_id"]) // JSONでは数値はfloat64になる
	assert.Equal(t, "ユーザー操作", logEntry["msg"])
}

func TestWithFields(t *testing.T) {
	var buf bytes.Buffer
	
	logrusLog := logrus.New()
	logrusLog.SetOutput(&buf)
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	
	logger := &logrusLogger{
		logger: logrusLog,
		entry:  logrus.NewEntry(logrusLog),
	}

	fieldsLogger := logger.WithFields(
		String("operation", "create"),
		Int("count", 5),
		Bool("success", true),
	)
	fieldsLogger.Info("複数フィールドテスト")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "create", logEntry["operation"])
	assert.Equal(t, float64(5), logEntry["count"])
	assert.Equal(t, true, logEntry["success"])
	assert.Equal(t, "複数フィールドテスト", logEntry["msg"])
}

func TestWithContextEnhanced(t *testing.T) {
	var buf bytes.Buffer
	
	logrusLog := logrus.New()
	logrusLog.SetOutput(&buf)
	logrusLog.SetFormatter(&logrus.JSONFormatter{})
	
	logger := &logrusLogger{
		logger: logrusLog,
		entry:  logrus.NewEntry(logrusLog),
	}

	// 複数の値をコンテキストに設定
	ctx := context.WithValue(context.Background(), "request_id", "req-123")
	ctx = context.WithValue(ctx, "user_id", 456)
	ctx = context.WithValue(ctx, "username", "testuser")
	ctx = context.WithValue(ctx, "role", "admin")
	
	contextLogger := logger.WithContext(ctx)
	contextLogger.Info("拡張コンテキストテスト")

	var logEntry map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logEntry)
	assert.NoError(t, err)

	assert.Equal(t, "req-123", logEntry["request_id"])
	assert.Equal(t, float64(456), logEntry["user_id"])
	assert.Equal(t, "testuser", logEntry["username"])
	assert.Equal(t, "admin", logEntry["role"])
	assert.Equal(t, "拡張コンテキストテスト", logEntry["msg"])
}

func TestGetDefaultConfig(t *testing.T) {
	// 環境変数をクリア
	os.Unsetenv("LOG_LEVEL")
	os.Unsetenv("LOG_FORMAT")
	os.Unsetenv("LOG_OUTPUT")
	os.Unsetenv("SERVICE_NAME")
	os.Unsetenv("SERVICE_VERSION")

	config := getDefaultConfig()
	
	assert.Equal(t, "info", config.Level)
	assert.Equal(t, "text", config.Format)
	assert.Equal(t, "stdout", config.Output)
	assert.Equal(t, "tournament-api", config.ServiceName)
	assert.Equal(t, "1.0.0", config.Version)
}

func TestGetDefaultConfigWithEnv(t *testing.T) {
	// 環境変数を設定
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("SERVICE_NAME", "custom-service")
	defer func() {
		os.Unsetenv("LOG_LEVEL")
		os.Unsetenv("LOG_FORMAT")
		os.Unsetenv("SERVICE_NAME")
	}()

	config := getDefaultConfig()
	
	assert.Equal(t, "debug", config.Level)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, "custom-service", config.ServiceName)
}

func TestStackField(t *testing.T) {
	field := Stack()
	
	assert.Equal(t, "stack", field.Key)
	assert.NotEmpty(t, field.Value)
	
	stackTrace, ok := field.Value.(string)
	assert.True(t, ok)
	assert.Contains(t, stackTrace, "TestStackField")
}
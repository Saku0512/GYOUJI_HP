package logger

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger は構造化ログのインターフェース
type Logger interface {
	Debug(msg string, fields ...Field)
	Info(msg string, fields ...Field)
	Warn(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	Fatal(msg string, fields ...Field)
	WithRequestID(requestID string) Logger
	WithContext(ctx context.Context) Logger
	WithComponent(component string) Logger
	WithUserID(userID int) Logger
	WithFields(fields ...Field) Logger
}

// Field はログフィールドを表す構造体
type Field struct {
	Key   string
	Value interface{}
}

// String はStringフィールドを作成する
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

// Int はIntフィールドを作成する
func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

// Int64 はInt64フィールドを作成する
func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

// Float64 はFloat64フィールドを作成する
func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

// Bool はBoolフィールドを作成する
func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

// Duration はDurationフィールドを作成する
func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

// Time はTimeフィールドを作成する
func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value.Format(time.RFC3339)}
}

// Err はErrorフィールドを作成する
func Err(err error) Field {
	return Field{Key: "error", Value: err.Error()}
}

// Stack はスタックトレースフィールドを作成する
func Stack() Field {
	buf := make([]byte, 1024)
	n := runtime.Stack(buf, false)
	return Field{Key: "stack", Value: string(buf[:n])}
}

// Any は任意の値のフィールドを作成する
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Component はコンポーネント名フィールドを作成する
func Component(name string) Field {
	return Field{Key: "component", Value: name}
}

// UserID はユーザーIDフィールドを作成する
func UserID(id int) Field {
	return Field{Key: "user_id", Value: id}
}

// RequestID はリクエストIDフィールドを作成する
func RequestID(id string) Field {
	return Field{Key: "request_id", Value: id}
}

// Method はHTTPメソッドフィールドを作成する
func Method(method string) Field {
	return Field{Key: "method", Value: method}
}

// Path はパスフィールドを作成する
func Path(path string) Field {
	return Field{Key: "path", Value: path}
}

// StatusCode はステータスコードフィールドを作成する
func StatusCode(code int) Field {
	return Field{Key: "status_code", Value: code}
}

// Latency はレイテンシフィールドを作成する
func Latency(duration time.Duration) Field {
	return Field{Key: "latency_ms", Value: float64(duration.Nanoseconds()) / 1e6}
}

// logrusLogger はlogrusを使用したLogger実装
type logrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// LogConfig はログ設定を表す構造体
type LogConfig struct {
	Level       string `json:"level"`
	Format      string `json:"format"`      // "json" or "text"
	Output      string `json:"output"`      // "stdout", "stderr", or file path
	ServiceName string `json:"service_name"`
	Version     string `json:"version"`
}

// NewLogger は新しいLoggerインスタンスを作成する
func NewLogger() Logger {
	return NewLoggerWithConfig(getDefaultConfig())
}

// NewLoggerWithConfig は設定を指定してLoggerインスタンスを作成する
func NewLoggerWithConfig(config LogConfig) Logger {
	logger := logrus.New()
	
	// ログフォーマットの設定
	if config.Format == "json" || os.Getenv("GO_ENV") == "production" {
		// JSON形式（本番環境推奨）
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339Nano,
			FieldMap: logrus.FieldMap{
				logrus.FieldKeyTime:  "timestamp",
				logrus.FieldKeyLevel: "level",
				logrus.FieldKeyMsg:   "message",
				logrus.FieldKeyFunc:  "caller",
			},
		})
	} else {
		// テキスト形式（開発環境）
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05.000",
			ForceColors:     true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				filename := strings.Split(f.File, "/")
				return fmt.Sprintf("%s()", f.Function), fmt.Sprintf("%s:%d", filename[len(filename)-1], f.Line)
			},
		})
	}

	// ログレベルの設定
	level := config.Level
	if level == "" {
		level = os.Getenv("LOG_LEVEL")
	}
	switch strings.ToLower(level) {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn", "warning":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	case "fatal":
		logger.SetLevel(logrus.FatalLevel)
	case "panic":
		logger.SetLevel(logrus.PanicLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	// 出力先の設定
	if config.Output != "" && config.Output != "stdout" {
		if config.Output == "stderr" {
			logger.SetOutput(os.Stderr)
		} else {
			// ファイル出力の場合は実装を追加可能
			logger.SetOutput(os.Stdout)
		}
	}

	// 呼び出し元情報の報告を有効化
	logger.SetReportCaller(true)

	// 基本フィールドを設定
	entry := logger.WithFields(logrus.Fields{
		"service": config.ServiceName,
		"version": config.Version,
	})

	return &logrusLogger{
		logger: logger,
		entry:  entry,
	}
}

// getDefaultConfig はデフォルト設定を取得する
func getDefaultConfig() LogConfig {
	return LogConfig{
		Level:       getEnvOrDefault("LOG_LEVEL", "info"),
		Format:      getEnvOrDefault("LOG_FORMAT", "text"),
		Output:      getEnvOrDefault("LOG_OUTPUT", "stdout"),
		ServiceName: getEnvOrDefault("SERVICE_NAME", "tournament-api"),
		Version:     getEnvOrDefault("SERVICE_VERSION", "1.0.0"),
	}
}

// getEnvOrDefault は環境変数を取得し、存在しない場合はデフォルト値を返す
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// fieldsToMap はFieldスライスをmapに変換する
func fieldsToMap(fields []Field) map[string]interface{} {
	m := make(map[string]interface{})
	for _, field := range fields {
		m[field.Key] = field.Value
	}
	return m
}

func (l *logrusLogger) Debug(msg string, fields ...Field) {
	l.entry.WithFields(fieldsToMap(fields)).Debug(msg)
}

func (l *logrusLogger) Info(msg string, fields ...Field) {
	l.entry.WithFields(fieldsToMap(fields)).Info(msg)
}

func (l *logrusLogger) Warn(msg string, fields ...Field) {
	l.entry.WithFields(fieldsToMap(fields)).Warn(msg)
}

func (l *logrusLogger) Error(msg string, fields ...Field) {
	l.entry.WithFields(fieldsToMap(fields)).Error(msg)
}

func (l *logrusLogger) Fatal(msg string, fields ...Field) {
	l.entry.WithFields(fieldsToMap(fields)).Fatal(msg)
}

func (l *logrusLogger) WithRequestID(requestID string) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField("request_id", requestID),
	}
}

func (l *logrusLogger) WithContext(ctx context.Context) Logger {
	newEntry := l.entry
	
	// コンテキストからリクエストIDを取得
	if requestID := ctx.Value("request_id"); requestID != nil {
		newEntry = newEntry.WithField("request_id", requestID.(string))
	}
	
	// コンテキストからユーザーIDを取得
	if userID := ctx.Value("user_id"); userID != nil {
		newEntry = newEntry.WithField("user_id", userID)
	}
	
	// コンテキストからユーザー名を取得
	if username := ctx.Value("username"); username != nil {
		newEntry = newEntry.WithField("username", username.(string))
	}
	
	// コンテキストからロールを取得
	if role := ctx.Value("role"); role != nil {
		newEntry = newEntry.WithField("role", role.(string))
	}
	
	return &logrusLogger{
		logger: l.logger,
		entry:  newEntry,
	}
}

func (l *logrusLogger) WithComponent(component string) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField("component", component),
	}
}

func (l *logrusLogger) WithUserID(userID int) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithField("user_id", userID),
	}
}

func (l *logrusLogger) WithFields(fields ...Field) Logger {
	return &logrusLogger{
		logger: l.logger,
		entry:  l.entry.WithFields(fieldsToMap(fields)),
	}
}

// グローバルロガーインスタンス
var globalLogger Logger

// Init はグローバルロガーを初期化する
func Init() {
	globalLogger = NewLogger()
}

// GetLogger はグローバルロガーを取得する
func GetLogger() Logger {
	if globalLogger == nil {
		Init()
	}
	return globalLogger
}

// 便利関数
func Debug(msg string, fields ...Field) {
	GetLogger().Debug(msg, fields...)
}

func Info(msg string, fields ...Field) {
	GetLogger().Info(msg, fields...)
}

func Warn(msg string, fields ...Field) {
	GetLogger().Warn(msg, fields...)
}

func ErrorMsg(msg string, fields ...Field) {
	GetLogger().Error(msg, fields...)
}

func Fatal(msg string, fields ...Field) {
	GetLogger().Fatal(msg, fields...)
}
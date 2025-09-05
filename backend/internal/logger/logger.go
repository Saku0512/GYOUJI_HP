package logger

import (
	"context"
	"os"

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

// Err はErrorフィールドを作成する
func Err(err error) Field {
	return Field{Key: "error", Value: err}
}

// Any は任意の値のフィールドを作成する
func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// logrusLogger はlogrusを使用したLogger実装
type logrusLogger struct {
	logger *logrus.Logger
	entry  *logrus.Entry
}

// NewLogger は新しいLoggerインスタンスを作成する
func NewLogger() Logger {
	logger := logrus.New()
	
	// 環境に応じてログフォーマットを設定
	env := os.Getenv("GO_ENV")
	if env == "production" {
		// 本番環境: JSON形式
		logger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		// 開発環境: テキスト形式
		logger.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: "2006-01-02 15:04:05",
			ForceColors:     true,
		})
	}

	// ログレベルの設定
	level := os.Getenv("LOG_LEVEL")
	switch level {
	case "debug":
		logger.SetLevel(logrus.DebugLevel)
	case "info":
		logger.SetLevel(logrus.InfoLevel)
	case "warn":
		logger.SetLevel(logrus.WarnLevel)
	case "error":
		logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.SetLevel(logrus.InfoLevel)
	}

	return &logrusLogger{
		logger: logger,
		entry:  logrus.NewEntry(logger),
	}
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
	// コンテキストからリクエストIDを取得
	if requestID := ctx.Value("request_id"); requestID != nil {
		return l.WithRequestID(requestID.(string))
	}
	return l
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
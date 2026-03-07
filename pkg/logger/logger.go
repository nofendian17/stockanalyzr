package logger

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// log is used by the package-level wrapper functions (Info, Error, etc.)
	// It has caller skip = 1 so the caller is accurately reported.
	log *zap.Logger

	// sugar is the sugared version of log, used for formatted logging (Infof, etc.)
	sugar *zap.SugaredLogger

	// baseLog is the logger without caller skip,
	// returned by Get() for those who want to use a *zap.Logger directly.
	baseLog *zap.Logger
)

func init() {
	Init(os.Getenv("APP_ENV"))
}

// Init initializes the logger based on the environment.
func Init(env string) {
	var config zap.Config

	if env == "production" || env == "prod" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// build wrapper logger
	wrapperLogger, err := config.Build(zap.AddCallerSkip(1))
	if err != nil {
		wrapperLogger = zap.NewNop()
	}

	log = wrapperLogger
	sugar = wrapperLogger.Sugar()

	// build base logger
	baseLogger, err := config.Build()
	if err != nil {
		baseLogger = zap.NewNop()
	}
	baseLog = baseLogger

	zap.ReplaceGlobals(baseLogger)
}

// Get returns the underlying zap logger without the caller skip.
// This is useful for passing to other libraries or embedding in structs.
func Get() *zap.Logger {
	return baseLog
}

// Sync flushes any buffered log entries.
func Sync() error {
	return baseLog.Sync()
}

// Info logs a message at InfoLevel.
func Info(msg string, fields ...zap.Field) {
	log.Info(msg, fields...)
}

// Error logs a message at ErrorLevel.
func Error(msg string, fields ...zap.Field) {
	log.Error(msg, fields...)
}

// Debug logs a message at DebugLevel.
func Debug(msg string, fields ...zap.Field) {
	log.Debug(msg, fields...)
}

// Warn logs a message at WarnLevel.
func Warn(msg string, fields ...zap.Field) {
	log.Warn(msg, fields...)
}

// Fatal logs a message at FatalLevel.
func Fatal(msg string, fields ...zap.Field) {
	log.Fatal(msg, fields...)
}

// With returns a logger with the given fields (no caller skip added).
func With(fields ...zap.Field) *zap.Logger {
	return baseLog.With(fields...)
}

// Infof uses sugared logger to format a message at InfoLevel.
func Infof(template string, args ...interface{}) {
	sugar.Infof(template, args...)
}

// Errorf uses sugared logger to format a message at ErrorLevel.
func Errorf(template string, args ...interface{}) {
	sugar.Errorf(template, args...)
}

// Debugf uses sugared logger to format a message at DebugLevel.
func Debugf(template string, args ...interface{}) {
	sugar.Debugf(template, args...)
}

// Warnf uses sugared logger to format a message at WarnLevel.
func Warnf(template string, args ...interface{}) {
	sugar.Warnf(template, args...)
}

// Fatalf uses sugared logger to format a message at FatalLevel.
func Fatalf(template string, args ...interface{}) {
	sugar.Fatalf(template, args...)
}

// WithContext can be extended to extract values from context like trace IDs.
func WithContext(ctx context.Context) *zap.Logger {
	// e.g. extract RequestID and return baseLog.With(zap.String("req_id", reqID))
	return baseLog
}

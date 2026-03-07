package logger

import (
	"context"
	"testing"

	"go.uber.org/zap"
)

func TestLogger(t *testing.T) {
	Init("development")

	// Test package-level functions
	Info("this is an info message", zap.String("key", "value"))
	Debug("this is a debug message", zap.Int("count", 42))

	// Test sugared logger
	Infof("this is a formatted %s message", "info")
	Debugf("this is a formatted debug message, count: %d", 42)

	// Test Get()
	log := Get()
	if log == nil {
		t.Fatal("Get() should not return nil")
	}
	log.Info("logging from Get()", zap.String("source", "Get"))

	// Test With()
	withLog := With(zap.String("context_key", "context_value"))
	if withLog == nil {
		t.Fatal("With() should not return nil")
	}
	withLog.Info("logging with fields")

	// Test WithContext
	ctxLog := WithContext(context.Background())
	if ctxLog == nil {
		t.Fatal("WithContext() should not return nil")
	}
	ctxLog.Info("logging with context")

	// Note: We skip Error/Fatal tests to avoid stopping the test run or polluting it with false errors.
}

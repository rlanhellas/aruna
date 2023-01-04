package logger

import (
	"context"
	"github.com/rlanhellas/aruna/global"
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

// SetLogger set the logger to be used
func SetLogger(l *zap.SugaredLogger) {
	logger = l
}

// Info log info level
func Info(ctx context.Context, msg string, args ...interface{}) {
	enrichedLogger(ctx).Infof(msg, args...)
}

// Debug log debug level
func Debug(ctx context.Context, msg string, args ...interface{}) {
	enrichedLogger(ctx).Debugf(msg, args...)
}

// Warn log warn level
func Warn(ctx context.Context, msg string, args ...interface{}) {
	enrichedLogger(ctx).Warnf(msg, args...)
}

// Error log error level
func Error(ctx context.Context, msg string, args ...interface{}) {
	enrichedLogger(ctx).Errorf(msg, args...)
}

func enrichedLogger(ctx context.Context) *zap.SugaredLogger {
	if ctx.Value(global.CorrelationID) != nil {
		return logger.With(zap.String(global.CorrelationID, ctx.Value(global.CorrelationID).(string)))
	} else {
		return logger
	}
}

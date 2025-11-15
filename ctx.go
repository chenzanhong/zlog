package zlog

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey string

const (
	RequestIDKey ctxKey = "request_id"
	UserIDKey    ctxKey = "user_id"
	TraceIDKey   ctxKey = "trace_id"
)

func loggerWithContext(ctx context.Context) *zap.Logger {
	logger := Logger()

	var extraFields []zap.Field

	if reqID, ok := ctx.Value(RequestIDKey).(string); ok && reqID != "" {
		extraFields = append(extraFields, zap.String("request_id", reqID))
	}
	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		extraFields = append(extraFields, zap.String("user_id", userID))
	}
	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		extraFields = append(extraFields, zap.String("trace_id", traceID))
	}

	if len(extraFields) > 0 {
		logger = logger.With(extraFields...)
	}
	return logger
}

func InfoCtx(ctx context.Context, msg string, fields ...Field) {
	logWithFieldsCtx(ctx, InfoLevel, msg, fields)
}

func logWithFieldsCtx(ctx context.Context, level Level, msg string, fields []Field) {
	executeHooks(level, msg, fields)

	zapLogger := loggerWithContext(ctx) // internal use only
	switch level {
	case DebugLevel:
		zapLogger.Debug(msg, fields...)
	case WarnLevel:
		zapLogger.Warn(msg, fields...)
	case ErrorLevel:
		zapLogger.Error(msg, fields...)
	case FatalLevel:
		zapLogger.Fatal(msg, fields...)
	default:
		zapLogger.Info(msg, fields...)
	}
}

func DebugCtx(ctx context.Context, msg string, fields ...Field) {
	logWithFieldsCtx(ctx, DebugLevel, msg, fields)
}

func WarnCtx(ctx context.Context, msg string, fields ...Field) {
	logWithFieldsCtx(ctx, WarnLevel, msg, fields)
}

func ErrorCtx(ctx context.Context, msg string, fields ...Field) {
	logWithFieldsCtx(ctx, ErrorLevel, msg, fields)
}

func FatalCtx(ctx context.Context, msg string, fields ...Field) {
	logWithFieldsCtx(ctx, FatalLevel, msg, fields)
}

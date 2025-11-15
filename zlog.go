package zlog

// ========== Structured Logging (High Performance, Recommended for Production) ==========
// Structured logging functions: parameters are []zlog.Field
func Debug(msg string, fields ...Field) { logWithFields(DebugLevel, msg, fields) }
func Info(msg string, fields ...Field)  { logWithFields(InfoLevel, msg, fields) }
func Warn(msg string, fields ...Field)  { logWithFields(WarnLevel, msg, fields) }
func Error(msg string, fields ...Field) { logWithFields(ErrorLevel, msg, fields) }
func Panic(msg string, fields ...Field) { logWithFields(PanicLevel, msg, fields) }
func Fatal(msg string, fields ...Field) { logWithFields(FatalLevel, msg, fields) }

func logWithFields(level Level, msg string, fields []Field) {
	executeHooks(level, msg, fields)

	logger := Logger()
	switch level {
	case DebugLevel:
		logger.Debug(msg, fields...)
	case InfoLevel:
		logger.Info(msg, fields...)
	case WarnLevel:
		logger.Warn(msg, fields...)
	case ErrorLevel:
		logger.Error(msg, fields...)
	case PanicLevel:
		logger.Panic(msg, fields...)
	case FatalLevel:
		logger.Fatal(msg, fields...)
	}
}

// ========== Key-Value Logging (Easy to Use, Suitable for Rapid Development) ==========
func Debugw(msg string, keysAndValues ...interface{}) { Sugar().Debugw(msg, keysAndValues...) }
func Infow(msg string, keysAndValues ...interface{})  { Sugar().Infow(msg, keysAndValues...) }
func Warnw(msg string, keysAndValues ...interface{})  { Sugar().Warnw(msg, keysAndValues...) }
func Errorw(msg string, keysAndValues ...interface{}) { Sugar().Errorw(msg, keysAndValues...) }
func Panicw(msg string, keysAndValues ...interface{}) { Sugar().Panicw(msg, keysAndValues...) }
func Fatalw(msg string, keysAndValues ...interface{}) { Sugar().Fatalw(msg, keysAndValues...) }

// ========== Formatted Logging (fmt Style Compatible) ==========
func Debugf(format string, args ...interface{}) { Sugar().Debugf(format, args...) }
func Infof(format string, args ...interface{})  { Sugar().Infof(format, args...) }
func Warnf(format string, args ...interface{})  { Sugar().Warnf(format, args...) }
func Errorf(format string, args ...interface{}) { Sugar().Errorf(format, args...) }
func Panicf(format string, args ...interface{}) { Sugar().Panicf(format, args...) }
func Fatalf(format string, args ...interface{}) { Sugar().Fatalf(format, args...) }

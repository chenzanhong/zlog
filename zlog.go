package zlog

// ========== 结构化日志（高性能，推荐生产环境使用）==========
// 结构化日志函数：参数是 []zlog.Field
func Debug(msg string, fields ...Field) { logWithFields(DebugLevel, msg, fields) }
func Info(msg string, fields ...Field)  { logWithFields(InfoLevel, msg, fields) }
func Warn(msg string, fields ...Field)  { logWithFields(WarnLevel, msg, fields) }
func Error(msg string, fields ...Field) { logWithFields(ErrorLevel, msg, fields) }
func Panic(msg string, fields ...Field) { logWithFields(PanicLevel, msg, fields) }
func Fatal(msg string, fields ...Field) { logWithFields(FatalLevel, msg, fields) }

func logWithFields(level Level, msg string, fields []Field) {
	executeHooks(level, msg, fields)

	logger := Logger()
	zapFields := toZapFields(fields)
	switch level {
	case DebugLevel:
		logger.Debug(msg, zapFields...)
	case InfoLevel:
		logger.Info(msg, zapFields...)
	case WarnLevel:
		logger.Warn(msg, zapFields...)
	case ErrorLevel:
		logger.Error(msg, zapFields...)
	case PanicLevel:
		logger.Panic(msg, zapFields...)
	case FatalLevel:
		logger.Fatal(msg, zapFields...)
	}
}

// ========== 键值对日志（易用，适合快速开发）==========
func Debugw(msg string, keysAndValues ...interface{}) { Sugar().Debugw(msg, keysAndValues...) }
func Infow(msg string, keysAndValues ...interface{})  { Sugar().Infow(msg, keysAndValues...) }
func Warnw(msg string, keysAndValues ...interface{})  { Sugar().Warnw(msg, keysAndValues...) }
func Errorw(msg string, keysAndValues ...interface{}) { Sugar().Errorw(msg, keysAndValues...) }
func Panicw(msg string, keysAndValues ...interface{}) { Sugar().Panicw(msg, keysAndValues...) }
func Fatalw(msg string, keysAndValues ...interface{}) { Sugar().Fatalw(msg, keysAndValues...) }

// ========== 格式化日志（兼容 fmt 风格）==========
func Debugf(format string, args ...interface{}) { Sugar().Debugf(format, args...) }
func Infof(format string, args ...interface{})  { Sugar().Infof(format, args...) }
func Warnf(format string, args ...interface{})  { Sugar().Warnf(format, args...) }
func Errorf(format string, args ...interface{}) { Sugar().Errorf(format, args...) }
func Panicf(format string, args ...interface{}) { Sugar().Panicf(format, args...) }
func Fatalf(format string, args ...interface{}) { Sugar().Fatalf(format, args...) }

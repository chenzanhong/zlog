package zlog

import "go.uber.org/zap/zapcore"

// Level 表示日志级别，对外隐藏 zapcore.Level
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	PanicLevel
	FatalLevel
)

// String 返回可读的级别名称
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	case PanicLevel:
		return "panic"
	case FatalLevel:
		return "fatal"
	default:
		return "unknown"
	}
}

// toZapCoreLevel 转换为 zapcore.Level（内部使用）
func (l Level) toZapCoreLevel() zapcore.Level {
	switch l {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case PanicLevel:
		return zapcore.PanicLevel
	case FatalLevel:
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

// fromZapCoreLevel 从 zapcore.Level 转换回来（如需要）
func fromZapCoreLevel(l zapcore.Level) Level {
	switch l {
	case zapcore.DebugLevel:
		return DebugLevel
	case zapcore.InfoLevel:
		return InfoLevel
	case zapcore.WarnLevel:
		return WarnLevel
	case zapcore.ErrorLevel:
		return ErrorLevel
	case zapcore.PanicLevel:
		return PanicLevel
	case zapcore.FatalLevel:
		return FatalLevel
	default:
		return InfoLevel
	}
}

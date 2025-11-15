package zlog

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerConfig struct {
	Level      Level  `yaml:"level"`
	Output     string `yaml:"output"`
	Format     string `yaml:"format"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
	Sampling   bool   `yaml:"sampling"`
}

func (c *LoggerConfig) Validate() error {
	if c.MaxSize <= 0 {
		c.MaxSize = 100
	}
	if c.MaxBackups < 0 {
		c.MaxBackups = 10
	}
	if c.MaxAge < 0 {
		c.MaxAge = 30
	}
	if (c.Output == "file" || c.Output == "both") && c.FilePath == "" {
		return fmt.Errorf("FilePath is required when Output='file'")
	}
	return nil
}

func DefaultConfig() LoggerConfig {
	return LoggerConfig{
		Level:      InfoLevel,
		Output:     "console",
		Format:     "console",
		FilePath:   "",
		MaxSize:    100, // MB
		MaxBackups: 10,
		MaxAge:     30, // days
		Compress:   true,
		Sampling:   false,
	}
}

type fieldType int

const (
	fieldAny fieldType = iota
	fieldString
	fieldInt
	fieldInt64
	fieldBool
	fieldFloat64
	fieldDuration
	fieldTime
)

// Field is zlog's custom log field type, hiding zap.Field internally
type Field struct {
	key   string
	value interface{}
	typ   fieldType
}

// Constructor functions
func String(key, val string) Field          { return Field{key: key, value: val, typ: fieldString} }
func Int(key string, val int) Field         { return Field{key: key, value: val, typ: fieldInt} }
func Int64(key string, val int64) Field     { return Field{key: key, value: val, typ: fieldInt64} }
func Bool(key string, val bool) Field       { return Field{key: key, value: val, typ: fieldBool} }
func Float64(key string, val float64) Field { return Field{key: key, value: val, typ: fieldFloat64} }
func Duration(key string, val time.Duration) Field {
	return Field{key: key, value: val, typ: fieldDuration}
}
func Time(key string, val time.Time) Field  { return Field{key: key, value: val, typ: fieldTime} }
func Any(key string, val interface{}) Field { return Field{key: key, value: val, typ: fieldAny} }

var (
	globalHooks []LogHook
	hooksMutex  sync.RWMutex
)

type LogHook interface {
	OnLog(level Level, msg string, fields []Field) error
}

func RegisterLogHook(hook LogHook) {
	hooksMutex.Lock()
	defer hooksMutex.Unlock()
	globalHooks = append(globalHooks, hook)
}

// executeHooks is called within logWithFields
func executeHooks(zlogLevel Level, msg string, fields []Field) {
	hooksMutex.RLock()
	hooks := make([]LogHook, len(globalHooks))
	copy(hooks, globalHooks)
	hooksMutex.RUnlock()

	for _, hook := range hooks {
		if err := hook.OnLog(zlogLevel, msg, fields); err != nil {
			fmt.Fprintf(os.Stderr, "[zlog] LogHook error: %v\n", err)
		}
	}
}

// Level represents log level, hiding zapcore.Level internally
type Level string

const (
	DebugLevel Level = "debug"
	InfoLevel  Level = "info"
	WarnLevel  Level = "warn"
	ErrorLevel Level = "error"
	PanicLevel Level = "panic"
	FatalLevel Level = "fatal"
)

// String returns human-readable level name
func (l Level) String() string {
	return string(l)
}

// Valid checks if the level is one of the predefined valid levels.
func (l Level) Valid() bool {
	switch l {
	case DebugLevel, InfoLevel, WarnLevel, ErrorLevel, PanicLevel, FatalLevel:
		return true
	default:
		return false
	}
}

// UnmarshalText implements encoding.TextUnmarshaler
// Supports parsing from YAML, JSON, TOML, env vars, etc.
func (l *Level) UnmarshalText(text []byte) error {
	levelStr := strings.ToLower(string(text))
	switch levelStr {
	case "debug", "d":
		*l = DebugLevel
	case "info", "i":
		*l = InfoLevel
	case "warn", "warning", "w":
		*l = WarnLevel
	case "error", "err", "e":
		*l = ErrorLevel
	case "panic", "p":
		*l = PanicLevel
	case "fatal", "f":
		*l = FatalLevel
	default:
		return fmt.Errorf("invalid log level: %q", string(text))
	}
	return nil
}

func (l Level) MarshalText() ([]byte, error) {
	if !l.Valid() {
		return []byte("info"), nil // safe default
	}
	return []byte(l), nil
}

// toZapCoreLevel converts to zapcore.Level (internal use)
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

// fromZapCoreLevel converts from zapcore.Level (if needed)
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

// Global instances (for backward compatibility)
var (
	globalLogger        *zap.Logger
	globalSugaredLogger *zap.SugaredLogger
	once                sync.Once
)

// newLogger creates a new zap.Logger instance with automatic config validation,
// default value filling, and path resolution.
// internal helper, not exported
func newLogger(config LoggerConfig) (*zap.Logger, error) {
	cfg := config

	// Normalize log level
	if cfg.Level < DebugLevel || cfg.Level > FatalLevel {
		cfg.Level = InfoLevel
	}

	// Normalize output destination
	switch cfg.Output {
	case "console", "file", "both":
		// valid
	default:
		cfg.Output = "console"
	}

	// Normalize format
	if cfg.Format != "json" && cfg.Format != "console" {
		cfg.Format = "console"
	}

	// Validate file path when needed
	if (cfg.Output == "file" || cfg.Output == "both") && cfg.FilePath == "" {
		return nil, fmt.Errorf("file path is required when output is 'file' or 'both'")
	}

	// Apply reasonable defaults for rotation settings
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = 100 // MB
	}
	if cfg.MaxBackups < 0 {
		cfg.MaxBackups = 10
	}
	if cfg.MaxAge < 0 {
		cfg.MaxAge = 30 // days
	}

	// Resolve relative file path to absolute
	if cfg.FilePath != "" && !filepath.IsAbs(cfg.FilePath) {
		wd, err := os.Getwd()
		if err != nil {
			return nil, fmt.Errorf("failed to get working directory: %w", err)
		}
		cfg.FilePath = filepath.Join(wd, cfg.FilePath)
	}

	// Create log directory if needed
	if cfg.FilePath != "" {
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory %q: %w", dir, err)
		}
	}

	// 4. Build encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// 5. Build cores
	var cores []zapcore.Core
	zapLevel := cfg.Level.toZapCoreLevel()

	// Console output
	if cfg.Output == "console" || cfg.Output == "both" {
		var enc zapcore.Encoder
		consoleEncCfg := encoderConfig
		if cfg.Format == "json" {
			enc = zapcore.NewJSONEncoder(consoleEncCfg)
		} else {
			consoleEncCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
			enc = zapcore.NewConsoleEncoder(consoleEncCfg)
		}
		cores = append(cores, zapcore.NewCore(enc, zapcore.Lock(os.Stdout), zapLevel))
	}

	// File output (always JSON)
	if cfg.Output == "file" || cfg.Output == "both" {
		writer := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
		enc := zapcore.NewJSONEncoder(encoderConfig)
		cores = append(cores, zapcore.NewCore(enc, zapcore.AddSync(writer), zapLevel))
	}

	if len(cores) == 0 {
		return nil, fmt.Errorf("no valid log output configured")
	}

	// 6. Build logger
	core := zapcore.NewTee(cores...)
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.ErrorOutput(zapcore.Lock(os.Stderr)),
	}

	if cfg.Sampling {
		options = append(options, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
		}))
	}

	logger := zap.New(core, options...)

	return logger, nil
}

// InitLogger initializes global logger (thread-safe)
func InitLogger(config LoggerConfig) error {
	var err error
	once.Do(func() {
		globalLogger, err = newLogger(config)
		if err == nil {
			globalSugaredLogger = globalLogger.Sugar()
			globalLogger = globalLogger.WithOptions(zap.AddCallerSkip(1))
		}
	})
	return err
}

// Logger returns global zap.Logger
func Logger() *zap.Logger {
	if globalLogger == nil {
		once.Do(func() {
			cfg := DefaultConfig()
			globalLogger, _ = newLogger(cfg)
			globalSugaredLogger = globalLogger.Sugar()
			globalLogger = globalLogger.WithOptions(zap.AddCallerSkip(1))
		})
	}
	return globalLogger
}

// Sugar returns global SugaredLogger
func Sugar() *zap.SugaredLogger {
	_ = Logger() // Trigger initialization
	return globalSugaredLogger
}

// InitDefault initializes with default configuration
func InitDefault() error {
	return InitLogger(DefaultConfig())
}

// MustInitDefault panics if default logger fails to initialize.
// Useful in main() for fail-fast behavior.
func MustInitDefault() {
	if err := InitDefault(); err != nil {
		panic(fmt.Sprintf("failed to init default logger: %v", err))
	}
}

// Sync ensures logs are flushed to disk
func Sync() error {
	logger := Logger() // Trigger default initialization if not already initialized
	return logger.Sync()
}

// toZapFields converts zlog.Field to zap.Field
func toZapFields(fields []Field) []zap.Field {
	if len(fields) == 0 {
		return nil
	}
	zfs := make([]zap.Field, 0, len(fields))
	for _, f := range fields {
		switch f.typ {
		case fieldString:
			zfs = append(zfs, zap.String(f.key, f.value.(string)))
		case fieldInt:
			zfs = append(zfs, zap.Int(f.key, f.value.(int)))
		case fieldInt64:
			zfs = append(zfs, zap.Int64(f.key, f.value.(int64)))
		case fieldBool:
			zfs = append(zfs, zap.Bool(f.key, f.value.(bool)))
		case fieldFloat64:
			zfs = append(zfs, zap.Float64(f.key, f.value.(float64)))
		case fieldDuration:
			zfs = append(zfs, zap.Duration(f.key, f.value.(time.Duration)))
		case fieldTime:
			zfs = append(zfs, zap.Time(f.key, f.value.(time.Time)))
		default:
			zfs = append(zfs, zap.Any(f.key, f.value))
		}
	}
	return zfs
}

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

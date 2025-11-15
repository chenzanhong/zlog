package zlog

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// 全局实例（兼容旧用法）
var (
	globalLogger *zap.Logger
	globalSugar  *zap.SugaredLogger
	once         sync.Once
)

// NewLogger 创建一个新的 Logger 实例
func NewLogger(config *LoggerConfig) (*zap.Logger, error) {
	// 创建日志目录
	if config.FilePath != "" {
		dir := filepath.Dir(config.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("创建日志目录失败: %v", err)
		}
	}

	// 设置日志级别
	level := zapcore.InfoLevel
	if config.Level != "" {
		var err error
		level, err = zapcore.ParseLevel(config.Level)
		if err != nil {
			log.Printf("无效的日志级别 %s，使用默认级别 info", config.Level)
		}
	}

	// 配置编码器
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

	// 创建输出核心
	var cores []zapcore.Core

	// 根据output参数控制输出目的地，format参数控制终端输出格式
	// 文件输出始终使用JSON格式
	switch config.Output {
	case "console":
		// 仅控制台输出，根据format决定格式
		var consoleEncoder zapcore.Encoder
		consoleEncoderConfig := encoderConfig

		if config.Format == "json" {
			// 使用JSON格式输出到控制台
			consoleEncoder = zapcore.NewJSONEncoder(consoleEncoderConfig)
		} else {
			// 默认使用带颜色的控制台格式
			consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
		}

		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level)
		cores = append(cores, consoleCore)

	case "file":
		// 仅文件输出，强制使用JSON格式
		if config.FilePath != "" {
			writer := &lumberjack.Logger{
				Filename:   config.FilePath,
				MaxSize:    config.MaxSize,
				MaxBackups: config.MaxBackups,
				MaxAge:     config.MaxAge,
				Compress:   config.Compress,
			}
			fileEncoder := zapcore.NewJSONEncoder(encoderConfig) // 文件始终使用JSON格式
			fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(writer), level)
			cores = append(cores, fileCore)
		} else {
			return nil, fmt.Errorf("日志参数output值为file，但是未指定日志文件路径")
		}

	case "both":
		fallthrough
	default:
		// 默认使用both模式：同时输出到文件和控制台
		// 文件输出强制使用JSON格式
		if config.FilePath != "" {
			writer := &lumberjack.Logger{
				Filename:   config.FilePath,
				MaxSize:    config.MaxSize,
				MaxBackups: config.MaxBackups,
				MaxAge:     config.MaxAge,
				Compress:   config.Compress,
			}
			fileEncoder := zapcore.NewJSONEncoder(encoderConfig) // 文件始终使用JSON格式
			fileCore := zapcore.NewCore(fileEncoder, zapcore.AddSync(writer), level)
			cores = append(cores, fileCore)
		} else {
			return nil, fmt.Errorf("日志参数output值为file，但是未指定日志文件路径")
		}

		// 控制台输出，根据format决定格式
		var consoleEncoder zapcore.Encoder
		consoleEncoderConfig := encoderConfig

		if config.Format == "json" {
			// 使用JSON格式输出到控制台
			consoleEncoder = zapcore.NewJSONEncoder(consoleEncoderConfig)
		} else {
			// 默认使用带颜色的控制台格式
			consoleEncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
			consoleEncoder = zapcore.NewConsoleEncoder(consoleEncoderConfig)
		}

		consoleCore := zapcore.NewCore(consoleEncoder, zapcore.Lock(os.Stdout), level)
		cores = append(cores, consoleCore)
	}

	// 如果没有配置输出，报错
	if len(cores) == 0 {
		return nil, fmt.Errorf("未配置任何日志输出")
	}

	// 创建核心
	core := zapcore.NewTee(cores...)

	// 创建日志选项
	options := []zap.Option{
		zap.AddCaller(),
		zap.AddCallerSkip(0), // 设置为0以正确显示实际调用日志的文件
		zap.AddStacktrace(zapcore.ErrorLevel),
		zap.ErrorOutput(zapcore.Lock(os.Stderr)),
	}

	// 添加采样
	if config.Sampling {
		options = append(options, zap.WrapCore(func(core zapcore.Core) zapcore.Core {
			return zapcore.NewSamplerWithOptions(core, time.Second, 100, 100)
		}))
	}

	// 创建Logger
	logger := zap.New(core, options...)

	// 记录初始化日志
	logger.Info("日志系统初始化完成",
		zap.String("level", level.String()),
		zap.String("output", config.Output),
		zap.String("format", config.Format),
		zap.String("file_path", config.FilePath),
	)
	return logger, nil
}

// InitLogger 初始化全局日志（线程安全）
func InitLogger(config *LoggerConfig) error {
	var err error
	once.Do(func() {
		globalLogger, err = NewLogger(config)
		if err == nil {
			globalSugar = globalLogger.Sugar()
		}
	})
	return err
}

// Logger 返回全局 zap.Logger
func Logger() *zap.Logger {
	if globalLogger == nil {
		once.Do(func() {
			cfg := defaultConfig()
			globalLogger, _ = NewLogger(cfg)
			globalSugar = globalLogger.Sugar()
		})
	}
	return globalLogger
}

// Sugar 返回全局 SugaredLogger
func Sugar() *zap.SugaredLogger {
	_ = Logger() // 触发初始化
	return globalSugar
}

// InitLoggerDefault 使用默认配置
func InitLoggerDefault() error {
	wd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前工作目录失败: %w", err)
	}
	// 日志文件默认路径
	defaultLogFilePath := filepath.Join(wd, "logs", "app.log")

	// 使用默认配置初始化
	config := &LoggerConfig{
		Level:      "info",
		Output:     "both",    // 默认为同时输出到文件和控制台
		Format:     "console", // 默认为控制台格式
		FilePath:   defaultLogFilePath,
		MaxSize:    20,
		MaxBackups: 5,
		MaxAge:     60,
		Compress:   false,
		Sampling:   false,
	}

	err = InitLogger(config)
	if err != nil {
		return fmt.Errorf("日志系统初始化失败: %w", err)
	}
	return nil
}

// FromEnv 从环境变量初始化全局日志
func InitFromEnv() error {
	cfg := &LoggerConfig{
		Level:      getEnv("LOG_LEVEL", "info"),
		Output:     getEnv("LOG_OUTPUT", "both"),
		Format:     getEnv("LOG_FORMAT", "console"),
		FilePath:   getEnv("LOG_FILE_PATH", ""),
		MaxSize:    getEnvInt("LOG_MAX_SIZE", 100),
		MaxBackups: getEnvInt("LOG_MAX_BACKUPS", 10),
		MaxAge:     getEnvInt("LOG_MAX_AGE", 30),
		Compress:   getEnvBool("LOG_COMPRESS", true),
		Sampling:   getEnvBool("LOG_SAMPLING", false),
	}
	if cfg.FilePath == "" {
		wd, _ := os.Getwd()
		cfg.FilePath = filepath.Join(wd, "logs", "app.log")
	}
	return InitLogger(cfg)
}

// 确保日志落盘
func Sync() error {
	logger := Logger() // 触发默认初始化（如果还没初始化）
	return logger.Sync()
}

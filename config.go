package zlog

import (
	"fmt"
	"os"
	"strconv"
)

type LoggerConfig struct {
	Level      Level `yaml:"level"`
	Output     string `yaml:"output"`
	Format     string `yaml:"format"`
	FilePath   string `yaml:"file_path"`
	MaxSize    int    `yaml:"max_size"`
	MaxBackups int    `yaml:"max_backups"`
	MaxAge     int    `yaml:"max_age"`
	Compress   bool   `yaml:"compress"`
	Sampling   bool   `yaml:"sampling"`
}

func (c *LoggerConfig) validate() error {
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

func getEnv(key, defaultValue string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if v, ok := os.LookupEnv(key); ok {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if v, ok := os.LookupEnv(key); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return defaultValue
}

func defaultConfig() *LoggerConfig {
	return &LoggerConfig{
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

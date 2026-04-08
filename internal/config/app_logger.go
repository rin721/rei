package config

import "fmt"

// LoggerConfig 描述日志配置。
type LoggerConfig struct {
	Level           string `yaml:"level" env:"LOG_LEVEL"`
	Format          string `yaml:"format" env:"LOG_FORMAT"`
	OutputPath      string `yaml:"output_path" env:"LOG_OUTPUT_PATH"`
	ErrorOutputPath string `yaml:"error_output_path" env:"LOG_ERROR_OUTPUT_PATH"`
	Filename        string `yaml:"filename" env:"LOG_FILENAME"`
	MaxSizeMB       int    `yaml:"max_size_mb" env:"LOG_MAX_SIZE_MB"`
	MaxBackups      int    `yaml:"max_backups" env:"LOG_MAX_BACKUPS"`
	MaxAgeDays      int    `yaml:"max_age_days" env:"LOG_MAX_AGE_DAYS"`
	Compress        bool   `yaml:"compress" env:"LOG_COMPRESS"`
}

// ValidateName 返回配置域名。
func (c LoggerConfig) ValidateName() string {
	return "logger"
}

// ValidateRequired 返回该配置域是否必需。
func (c LoggerConfig) ValidateRequired() bool {
	return true
}

// Validate 校验 LoggerConfig。
func (c LoggerConfig) Validate() error {
	switch c.Level {
	case "", "debug", "info", "warn", "warning", "error", "fatal":
	default:
		return fmt.Errorf("unsupported level %q", c.Level)
	}
	return nil
}

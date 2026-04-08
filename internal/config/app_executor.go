package config

import "fmt"

// ExecutorConfig 描述执行器配置。
type ExecutorConfig struct {
	Enabled         bool `yaml:"enabled" env:"EXECUTOR_ENABLED"`
	DefaultPoolSize int  `yaml:"default_pool_size" env:"EXECUTOR_DEFAULT_POOL_SIZE"`
}

// ValidateName 返回配置域名。
func (c ExecutorConfig) ValidateName() string {
	return "executor"
}

// ValidateRequired 返回该配置域是否必需。
func (c ExecutorConfig) ValidateRequired() bool {
	return true
}

// Validate 校验 ExecutorConfig。
func (c ExecutorConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.DefaultPoolSize <= 0 {
		return fmt.Errorf("default_pool_size must be greater than 0")
	}
	return nil
}

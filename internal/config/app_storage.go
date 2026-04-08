package config

import "fmt"

// StorageConfig 描述存储配置。
type StorageConfig struct {
	Enabled bool   `yaml:"enabled" env:"STORAGE_ENABLED"`
	Driver  string `yaml:"driver" env:"STORAGE_DRIVER"`
	RootDir string `yaml:"root_dir" env:"STORAGE_ROOT_DIR"`
}

// ValidateName 返回配置域名。
func (c StorageConfig) ValidateName() string {
	return "storage"
}

// ValidateRequired 返回该配置域是否必需。
func (c StorageConfig) ValidateRequired() bool {
	return true
}

// Validate 校验 StorageConfig。
func (c StorageConfig) Validate() error {
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}
	if c.RootDir == "" {
		return fmt.Errorf("root_dir is required")
	}
	return nil
}

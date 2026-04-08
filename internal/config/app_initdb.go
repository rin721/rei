package config

import "fmt"

// InitDBConfig 描述 initdb 模式配置。
type InitDBConfig struct {
	Enabled   bool   `yaml:"enabled" env:"INITDB_ENABLED"`
	Driver    string `yaml:"driver" env:"INITDB_DRIVER"`
	OutputDir string `yaml:"output_dir" env:"INITDB_OUTPUT_DIR"`
	LockFile  string `yaml:"lock_file" env:"INITDB_LOCK_FILE"`
}

// ValidateName 返回配置域名。
func (c InitDBConfig) ValidateName() string {
	return "initdb"
}

// ValidateRequired 返回该配置域是否必需。
func (c InitDBConfig) ValidateRequired() bool {
	return true
}

// Validate 校验 InitDBConfig。
func (c InitDBConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.OutputDir == "" {
		return fmt.Errorf("output_dir is required")
	}
	if c.LockFile == "" {
		return fmt.Errorf("lock_file is required")
	}
	return nil
}

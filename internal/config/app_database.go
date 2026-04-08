package config

import "fmt"

// DatabaseConfig 描述数据库连接配置。
type DatabaseConfig struct {
	Enabled      bool   `yaml:"enabled" env:"DB_ENABLED"`
	Driver       string `yaml:"driver" env:"DB_DRIVER"`
	DSN          string `yaml:"dsn" env:"DB_DSN"`
	Host         string `yaml:"host" env:"DB_HOST"`
	Port         int    `yaml:"port" env:"DB_PORT"`
	Name         string `yaml:"name" env:"DB_NAME"`
	User         string `yaml:"user" env:"DB_USER"`
	Password     string `yaml:"password" env:"DB_PASSWORD"`
	SSLMode      string `yaml:"ssl_mode" env:"DB_SSL_MODE"`
	MaxOpenConns int    `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns int    `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
}

// ValidateName 返回配置域名。
func (c DatabaseConfig) ValidateName() string {
	return "database"
}

// ValidateRequired 返回该配置域是否必需。
func (c DatabaseConfig) ValidateRequired() bool {
	return false
}

// Validate 校验 DatabaseConfig。
func (c DatabaseConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Driver == "" {
		return fmt.Errorf("driver is required")
	}
	if c.DSN == "" {
		if c.Name == "" {
			return fmt.Errorf("name is required when dsn is empty")
		}
		if c.Driver != "sqlite" && c.Host == "" {
			return fmt.Errorf("host is required when dsn is empty")
		}
	}
	if c.Port < 0 {
		return fmt.Errorf("port must be non-negative")
	}
	return nil
}

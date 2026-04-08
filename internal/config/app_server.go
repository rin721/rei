package config

import "fmt"

// ServerConfig 描述 HTTPServer 运行配置。
type ServerConfig struct {
	Host         string `yaml:"host" env:"SERVER_HOST"`
	Port         int    `yaml:"port" env:"SERVER_PORT"`
	Mode         string `yaml:"mode" env:"SERVER_MODE"`
	ReadTimeout  int    `yaml:"read_timeout" env:"SERVER_READ_TIMEOUT"`
	WriteTimeout int    `yaml:"write_timeout" env:"SERVER_WRITE_TIMEOUT"`
	IdleTimeout  int    `yaml:"idle_timeout" env:"SERVER_IDLE_TIMEOUT"`
}

// ValidateName 返回配置域名。
func (c ServerConfig) ValidateName() string {
	return "server"
}

// ValidateRequired 返回该配置域是否必需。
func (c ServerConfig) ValidateRequired() bool {
	return true
}

// Validate 校验 ServerConfig。
func (c ServerConfig) Validate() error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}
	if c.Port <= 0 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if c.ReadTimeout < 0 || c.WriteTimeout < 0 || c.IdleTimeout < 0 {
		return fmt.Errorf("timeouts must be non-negative")
	}
	return nil
}

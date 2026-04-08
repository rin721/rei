package config

import "fmt"

// RedisConfig 描述缓存后端相关配置。
type RedisConfig struct {
	Enabled  bool   `yaml:"enabled" env:"REDIS_ENABLED"`
	Addr     string `yaml:"addr" env:"REDIS_ADDR"`
	Password string `yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `yaml:"db" env:"REDIS_DB"`
}

// ValidateName 返回配置域名。
func (c RedisConfig) ValidateName() string {
	return "redis"
}

// ValidateRequired 返回该配置域是否必需。
func (c RedisConfig) ValidateRequired() bool {
	return false
}

// Validate 校验 RedisConfig。
func (c RedisConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Addr == "" {
		return fmt.Errorf("addr is required")
	}
	if c.DB < 0 {
		return fmt.Errorf("db must be non-negative")
	}
	return nil
}

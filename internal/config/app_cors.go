package config

import "fmt"

// CORSConfig 描述跨域配置。
type CORSConfig struct {
	Enabled      bool     `yaml:"enabled" env:"CORS_ENABLED"`
	AllowOrigins []string `yaml:"allow_origins" env:"CORS_ALLOW_ORIGINS"`
	AllowMethods []string `yaml:"allow_methods" env:"CORS_ALLOW_METHODS"`
	AllowHeaders []string `yaml:"allow_headers" env:"CORS_ALLOW_HEADERS"`
}

// ValidateName 返回配置域名。
func (c CORSConfig) ValidateName() string {
	return "cors"
}

// ValidateRequired 返回该配置域是否必需。
func (c CORSConfig) ValidateRequired() bool {
	return true
}

// Validate 校验 CORSConfig。
func (c CORSConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if len(c.AllowOrigins) == 0 {
		return fmt.Errorf("allow_origins must not be empty")
	}
	return nil
}

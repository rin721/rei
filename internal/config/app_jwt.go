package config

import "fmt"

// JWTConfig 描述 JWT 配置。
type JWTConfig struct {
	Enabled               bool   `yaml:"enabled" env:"JWT_ENABLED"`
	Issuer                string `yaml:"issuer" env:"JWT_ISSUER"`
	Secret                string `yaml:"secret" env:"JWT_SECRET"`
	AccessTokenTTLMinutes int    `yaml:"access_token_ttl_minutes" env:"JWT_ACCESS_TOKEN_TTL_MINUTES"`
	RefreshTokenTTLHours  int    `yaml:"refresh_token_ttl_hours" env:"JWT_REFRESH_TOKEN_TTL_HOURS"`
}

// ValidateName 返回配置域名。
func (c JWTConfig) ValidateName() string {
	return "jwt"
}

// ValidateRequired 返回该配置域是否必需。
func (c JWTConfig) ValidateRequired() bool {
	return true
}

// Validate 校验 JWTConfig。
func (c JWTConfig) Validate() error {
	if !c.Enabled {
		return nil
	}
	if c.Secret == "" {
		return fmt.Errorf("secret is required")
	}
	if c.AccessTokenTTLMinutes <= 0 || c.RefreshTokenTTLHours <= 0 {
		return fmt.Errorf("token ttl must be greater than 0")
	}
	return nil
}

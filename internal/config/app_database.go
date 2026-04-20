package config

import "fmt"

// DatabaseConfig describes database connectivity and migration settings.
type DatabaseConfig struct {
	Enabled       bool   `yaml:"enabled" env:"DB_ENABLED"`
	Driver        string `yaml:"driver" env:"DB_DRIVER"`
	DSN           string `yaml:"dsn" env:"DB_DSN"`
	Host          string `yaml:"host" env:"DB_HOST"`
	Port          int    `yaml:"port" env:"DB_PORT"`
	Name          string `yaml:"name" env:"DB_NAME"`
	User          string `yaml:"user" env:"DB_USER"`
	Password      string `yaml:"password" env:"DB_PASSWORD"`
	SSLMode       string `yaml:"ssl_mode" env:"DB_SSL_MODE"`
	MigrationsDir string `yaml:"migrations_dir" env:"DB_MIGRATIONS_DIR"`
	MaxOpenConns  int    `yaml:"max_open_conns" env:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns  int    `yaml:"max_idle_conns" env:"DB_MAX_IDLE_CONNS"`
}

func (c DatabaseConfig) ValidateName() string {
	return "database"
}

func (c DatabaseConfig) ValidateRequired() bool {
	return false
}

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

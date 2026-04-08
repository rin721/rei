package config

import (
	"fmt"
)

type domainConfig interface {
	ValidateName() string
	ValidateRequired() bool
	Validate() error
}

// Config 描述应用的顶层配置结构。
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Logger   LoggerConfig   `yaml:"logger"`
	I18n     I18nConfig     `yaml:"i18n"`
	InitDB   InitDBConfig   `yaml:"initdb"`
	Executor ExecutorConfig `yaml:"executor"`
	JWT      JWTConfig      `yaml:"jwt"`
	RBAC     RBACConfig     `yaml:"rbac"`
	Storage  StorageConfig  `yaml:"storage"`
	CORS     CORSConfig     `yaml:"cors"`
}

// Default 返回一份安全且可运行的默认配置。
func Default() Config {
	return Config{
		Server: ServerConfig{
			Host:         "0.0.0.0",
			Port:         9999,
			Mode:         "debug",
			ReadTimeout:  10,
			WriteTimeout: 10,
			IdleTimeout:  60,
		},
		Database: DatabaseConfig{
			Enabled:      true,
			Driver:       "sqlite",
			Host:         "",
			Port:         0,
			Name:         "tmp/go_scaffold2.db",
			User:         "",
			Password:     "",
			SSLMode:      "disable",
			MaxOpenConns: 20,
			MaxIdleConns: 10,
		},
		Redis: RedisConfig{
			Enabled: false,
			Addr:    "127.0.0.1:6379",
			DB:      0,
		},
		Logger: LoggerConfig{
			Level:           "info",
			Format:          "json",
			OutputPath:      "stdout",
			ErrorOutputPath: "stderr",
			Filename:        "logs/app.log",
			MaxSizeMB:       100,
			MaxBackups:      10,
			MaxAgeDays:      7,
			Compress:        false,
		},
		I18n: I18nConfig{
			DefaultLocale:  "zh-CN",
			FallbackLocale: "en-US",
			LocaleDir:      "configs/locales",
		},
		InitDB: InitDBConfig{
			Enabled:   true,
			Driver:    "sqlite",
			OutputDir: "scripts/initdb",
			LockFile:  "scripts/initdb/.initdb.lock",
		},
		Executor: ExecutorConfig{
			Enabled:         true,
			DefaultPoolSize: 32,
		},
		JWT: JWTConfig{
			Enabled:               true,
			Issuer:                "go-scaffold2",
			Secret:                "replace_me_with_a_private_secret",
			AccessTokenTTLMinutes: 60,
			RefreshTokenTTLHours:  72,
		},
		RBAC: RBACConfig{
			Enabled:     true,
			ModelPath:   "configs/rbac_model.conf",
			PolicyTable: "casbin_rule",
			AutoSave:    true,
		},
		Storage: StorageConfig{
			Enabled: false,
			Driver:  "local",
			RootDir: "tmp/storage",
		},
		CORS: CORSConfig{
			Enabled: true,
			AllowOrigins: []string{
				"*",
			},
			AllowMethods: []string{
				"GET",
				"POST",
				"PUT",
				"PATCH",
				"DELETE",
				"OPTIONS",
			},
			AllowHeaders: []string{
				"Authorization",
				"Content-Type",
				"X-Trace-ID",
			},
		},
	}
}

// Clone 返回配置副本，避免热重载时共享切片引用。
func (c Config) Clone() Config {
	cloned := c
	cloned.CORS.AllowOrigins = append([]string(nil), c.CORS.AllowOrigins...)
	cloned.CORS.AllowMethods = append([]string(nil), c.CORS.AllowMethods...)
	cloned.CORS.AllowHeaders = append([]string(nil), c.CORS.AllowHeaders...)
	return cloned
}

// Validate 校验所有配置域。
func (c Config) Validate() error {
	domains := []domainConfig{
		c.Server,
		c.Database,
		c.Redis,
		c.Logger,
		c.I18n,
		c.InitDB,
		c.Executor,
		c.JWT,
		c.RBAC,
		c.Storage,
		c.CORS,
	}

	for _, domain := range domains {
		if err := domain.Validate(); err != nil {
			return fmt.Errorf("%s: %w", domain.ValidateName(), err)
		}
	}

	return nil
}

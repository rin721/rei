package app

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rei0721/go-scaffold2/internal/config"
	pkgcache "github.com/rei0721/go-scaffold2/pkg/cache"
	pkgdatabase "github.com/rei0721/go-scaffold2/pkg/database"
	pkgexecutor "github.com/rei0721/go-scaffold2/pkg/executor"
	pkghttpserver "github.com/rei0721/go-scaffold2/pkg/httpserver"
	pkgi18n "github.com/rei0721/go-scaffold2/pkg/i18n"
	pkgjwt "github.com/rei0721/go-scaffold2/pkg/jwt"
	pkglogger "github.com/rei0721/go-scaffold2/pkg/logger"
	pkgrbac "github.com/rei0721/go-scaffold2/pkg/rbac"
	pkgstorage "github.com/rei0721/go-scaffold2/pkg/storage"
)

func toLoggerConfig(cfg config.LoggerConfig) pkglogger.Config {
	return pkglogger.Config{
		Level:  cfg.Level,
		Prefix: "go-scaffold2",
	}
}

func toI18nConfig(cfg config.I18nConfig) pkgi18n.Config {
	return pkgi18n.Config{
		DefaultLocale:  cfg.DefaultLocale,
		FallbackLocale: cfg.FallbackLocale,
		LocaleDir:      cfg.LocaleDir,
	}
}

func toCacheConfig(config.RedisConfig) pkgcache.Config {
	return pkgcache.Config{
		DefaultTTL: 5 * time.Minute,
	}
}

func toDatabaseConfig(cfg config.DatabaseConfig) pkgdatabase.Config {
	return pkgdatabase.Config{
		Driver:          cfg.Driver,
		DSN:             buildDSN(cfg),
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: time.Hour,
	}
}

func toExecutorConfig(cfg config.ExecutorConfig) pkgexecutor.Config {
	return pkgexecutor.Config{
		DefaultPool: pkgexecutor.DefaultPoolName,
		Pools: []pkgexecutor.PoolConfig{
			{
				Name:      pkgexecutor.DefaultPoolName,
				Workers:   cfg.DefaultPoolSize,
				QueueSize: cfg.DefaultPoolSize,
			},
		},
	}
}

func toJWTConfig(cfg config.JWTConfig) pkgjwt.Config {
	return pkgjwt.Config{
		Issuer:     cfg.Issuer,
		Secret:     cfg.Secret,
		AccessTTL:  time.Duration(cfg.AccessTokenTTLMinutes) * time.Minute,
		RefreshTTL: time.Duration(cfg.RefreshTokenTTLHours) * time.Hour,
	}
}

func toStorageConfig(cfg config.StorageConfig) pkgstorage.Config {
	return pkgstorage.Config{
		RootDir: cfg.RootDir,
	}
}

func toHTTPServerConfig(cfg config.ServerConfig) pkghttpserver.Config {
	return pkghttpserver.Config{
		Address:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		ReadTimeout:     time.Duration(cfg.ReadTimeout) * time.Second,
		WriteTimeout:    time.Duration(cfg.WriteTimeout) * time.Second,
		IdleTimeout:     time.Duration(cfg.IdleTimeout) * time.Second,
		ShutdownTimeout: defaultShutdownTimeout,
	}
}

func toRBACConfig(cfg config.RBACConfig) (pkgrbac.Config, error) {
	modelText, err := readOptionalModel(cfg.ModelPath)
	if err != nil {
		return pkgrbac.Config{}, err
	}

	return pkgrbac.Config{
		ModelText: modelText,
		AutoSave:  cfg.AutoSave,
	}, nil
}

func buildDSN(cfg config.DatabaseConfig) string {
	if cfg.DSN != "" {
		return cfg.DSN
	}

	switch cfg.Driver {
	case "sqlite":
		if cfg.Name != "" {
			return cfg.Name
		}
		return "file:go_scaffold2?mode=memory&cache=shared"
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
	default:
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name)
	}
}

func readOptionalModel(path string) (string, error) {
	if path == "" {
		return "", nil
	}

	cleaned := filepath.Clean(path)
	content, err := os.ReadFile(cleaned)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	return string(content), nil
}

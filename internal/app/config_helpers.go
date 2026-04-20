package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rin721/rei/internal/config"
	pkgcache "github.com/rin721/rei/pkg/cache"
	pkgdatabase "github.com/rin721/rei/pkg/database"
	pkgexecutor "github.com/rin721/rei/pkg/executor"
	pkghttpserver "github.com/rin721/rei/pkg/httpserver"
	pkgi18n "github.com/rin721/rei/pkg/i18n"
	pkgjwt "github.com/rin721/rei/pkg/jwt"
	pkglogger "github.com/rin721/rei/pkg/logger"
	pkgrbac "github.com/rin721/rei/pkg/rbac"
	pkgstorage "github.com/rin721/rei/pkg/storage"
)

func toLoggerConfig(cfg config.LoggerConfig) *pkglogger.Config {
	output := strings.ToLower(strings.TrimSpace(cfg.OutputPath))
	if output == "" || (output != pkglogger.OutputStdout && output != pkglogger.OutputFile && output != pkglogger.OutputBoth) {
		output = pkglogger.OutputStdout
	}

	return &pkglogger.Config{
		Level:      cfg.Level,
		Format:     cfg.Format,
		Output:     output,
		FilePath:   cfg.Filename,
		MaxSize:    cfg.MaxSizeMB,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAgeDays,
	}
}

func toI18nConfig(cfg config.I18nConfig) *pkgi18n.Config {
	supported := []string{cfg.DefaultLocale}
	if cfg.FallbackLocale != "" && cfg.FallbackLocale != cfg.DefaultLocale {
		supported = append(supported, cfg.FallbackLocale)
	}

	return &pkgi18n.Config{
		DefaultLanguage:    cfg.DefaultLocale,
		SupportedLanguages: supported,
		MessagesDir:        cfg.LocaleDir,
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

func toExecutorConfig(cfg config.ExecutorConfig) []pkgexecutor.Config {
	return []pkgexecutor.Config{
		{
			Name:        defaultExecutorPoolName,
			Size:        cfg.DefaultPoolSize,
			Expiry:      pkgexecutor.DefaultWorkerExpiry,
			NonBlocking: pkgexecutor.DefaultNonBlocking,
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

func toStorageConfig(cfg config.StorageConfig) *pkgstorage.Config {
	fsType := pkgstorage.FSTypeBasePathFS
	switch strings.ToLower(strings.TrimSpace(cfg.Driver)) {
	case "", "local":
		fsType = pkgstorage.FSTypeBasePathFS
	case string(pkgstorage.FSTypeOS):
		fsType = pkgstorage.FSTypeOS
	case string(pkgstorage.FSTypeMemory):
		fsType = pkgstorage.FSTypeMemory
	case string(pkgstorage.FSTypeReadOnly):
		fsType = pkgstorage.FSTypeReadOnly
	case string(pkgstorage.FSTypeBasePathFS):
		fsType = pkgstorage.FSTypeBasePathFS
	}

	return &pkgstorage.Config{
		FSType:          fsType,
		BasePath:        cfg.RootDir,
		EnableWatch:     cfg.Enabled,
		WatchBufferSize: 100,
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

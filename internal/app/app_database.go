package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/rin721/rei/internal/config"
	pkgdatabase "github.com/rin721/rei/pkg/database"
)

func (p infrastructureProvisioning) initDatabase(ctx context.Context) error {
	if p.infra.database != nil || !p.cfg.Database.Enabled {
		return nil
	}
	if err := ensureSQLitePath(p.cfg.Database); err != nil {
		return fmt.Errorf("prepare sqlite path: %w", err)
	}

	store, err := pkgdatabase.New(toDatabaseConfig(p.cfg.Database))
	if err != nil {
		return fmt.Errorf("init database: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := store.Ping(pingCtx); err != nil {
		_ = store.Close()
		return fmt.Errorf("ping database: %w", err)
	}

	p.infra.database = store
	return nil
}

func ensureSQLitePath(cfg config.DatabaseConfig) error {
	if cfg.Driver != "sqlite" {
		return nil
	}

	target := strings.TrimSpace(cfg.Name)
	if strings.TrimSpace(cfg.DSN) != "" {
		target = cfg.DSN
	}
	if target == "" || target == ":memory:" || strings.Contains(target, "mode=memory") {
		return nil
	}

	if strings.HasPrefix(target, "file:") {
		target = strings.TrimPrefix(target, "file:")
		if index := strings.Index(target, "?"); index >= 0 {
			target = target[:index]
		}
	}

	if target == "" {
		return nil
	}

	dir := filepath.Dir(filepath.Clean(target))
	if dir == "." || dir == "" {
		return nil
	}

	return os.MkdirAll(dir, 0o755)
}

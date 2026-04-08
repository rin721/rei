package config

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManagerLoadAppliesEnvExpansionAndOverride(t *testing.T) {
	root := t.TempDir()
	localeDir := filepath.Join(root, "locales")
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	configPath := filepath.Join(root, "config.yaml")
	content := []byte(`
server:
  host: 127.0.0.1
  port: ${SERVER_PORT:9999}
i18n:
  locale_dir: ${I18N_LOCALE_DIR:configs/locales}
jwt:
  secret: ${JWT_SECRET:placeholder}
`)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	t.Setenv("SERVER_PORT", "12345")
	t.Setenv("I18N_LOCALE_DIR", localeDir)
	t.Setenv("JWT_SECRET", "env-secret")

	manager := NewManager(Options{Path: configPath})
	cfg, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	if cfg.Server.Port != 12345 {
		t.Fatalf("Server.Port = %d, want 12345", cfg.Server.Port)
	}
	if cfg.I18n.LocaleDir != localeDir {
		t.Fatalf("I18n.LocaleDir = %q, want %q", cfg.I18n.LocaleDir, localeDir)
	}
	if cfg.JWT.Secret != "env-secret" {
		t.Fatalf("JWT.Secret = %q, want %q", cfg.JWT.Secret, "env-secret")
	}
}

func TestManagerStartReloadsConfig(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	configPath := filepath.Join(root, "config.yaml")
	localeDir := filepath.Join(root, "locales")
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}

	writeConfig := func(port int) {
		content := []byte(
			"server:\n  host: 127.0.0.1\n  port: " + itoa(port) + "\n" +
				"i18n:\n  locale_dir: " + localeDir + "\n" +
				"jwt:\n  secret: reload-secret\n",
		)
		if err := os.WriteFile(configPath, content, 0o644); err != nil {
			t.Fatalf("WriteFile() returned error: %v", err)
		}
	}

	writeConfig(9999)

	manager := NewManager(Options{
		Path:          configPath,
		WatchInterval: 20 * time.Millisecond,
	})
	if _, err := manager.Load(); err != nil {
		t.Fatalf("Load() returned error: %v", err)
	}

	reloaded := make(chan struct{}, 1)
	manager.RegisterReloadHook("test", func(_ context.Context, oldCfg, newCfg Config) error {
		if oldCfg.Server.Port == 9999 && newCfg.Server.Port == 10001 {
			select {
			case reloaded <- struct{}{}:
			default:
			}
		}
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := manager.Start(ctx); err != nil {
		t.Fatalf("Start() returned error: %v", err)
	}
	defer func() {
		_ = manager.Stop()
	}()

	writeConfig(10001)

	select {
	case <-reloaded:
	case <-time.After(2 * time.Second):
		t.Fatal("config reload hook was not triggered")
	}

	if got := manager.Current().Server.Port; got != 10001 {
		t.Fatalf("Current().Server.Port = %d, want 10001", got)
	}
}

func itoa(value int) string {
	return fmt.Sprintf("%d", value)
}

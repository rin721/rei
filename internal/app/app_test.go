package app

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAppRunServerDryRun(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t)

	application, err := New(Options{
		Mode:       ModeServer,
		ConfigPath: configPath,
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}
}

func TestAppRunInitDBGeneratesSQLAndLock(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t)

	application, err := New(Options{
		Mode:       ModeInitDB,
		ConfigPath: configPath,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	root := filepath.Dir(filepath.Dir(configPath))
	outputDir := filepath.Join(root, "scripts", "initdb")
	if _, err := os.Stat(outputDir); err != nil {
		t.Fatalf("Stat() returned error: %v", err)
	}

	scriptPath := filepath.Join(outputDir, "initdb.sqlite.sql")
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}
	if !strings.Contains(string(content), "CREATE TABLE IF NOT EXISTS `users`") {
		t.Fatalf("script content does not contain users table DDL")
	}

	lockPath := filepath.Join(outputDir, ".initdb.lock")
	if _, err := os.Stat(lockPath); err != nil {
		t.Fatalf("Stat(lockPath) returned error: %v", err)
	}
}

func TestAppRunInitDBDryRunSkipsLockFile(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t)

	application, err := New(Options{
		Mode:       ModeInitDB,
		ConfigPath: configPath,
		DryRun:     true,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	outputDir := filepath.Join(filepath.Dir(filepath.Dir(configPath)), "scripts", "initdb")
	if _, err := os.Stat(filepath.Join(outputDir, "initdb.sqlite.sql")); err != nil {
		t.Fatalf("Stat(script) returned error: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outputDir, ".initdb.lock")); !os.IsNotExist(err) {
		t.Fatalf("Stat(lock) error = %v, want IsNotExist", err)
	}
}

func writeTestConfig(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	localeDir := filepath.Join(root, "configs", "locales")
	if err := os.MkdirAll(localeDir, 0o755); err != nil {
		t.Fatalf("MkdirAll() returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localeDir, "zh-CN.yaml"), []byte("message: 你好\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(localeDir, "en-US.yaml"), []byte("message: hello\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	modelPath := filepath.Join(root, "configs", "rbac_model.conf")
	if err := os.WriteFile(modelPath, []byte("[request_definition]\nr = sub, obj, act\n[policy_definition]\np = sub, obj, act\n[role_definition]\ng = _, _\n[policy_effect]\ne = some(where (p.eft == allow))\n[matchers]\nm = (g(r.sub, p.sub) || r.sub == p.sub) && r.obj == p.obj && r.act == p.act\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	configPath := filepath.Join(root, "configs", "config.yaml")
	content := []byte(
		"server:\n  host: 127.0.0.1\n  port: 18080\n  mode: debug\n" +
			"database:\n  enabled: true\n  driver: sqlite\n  name: " + filepath.Join(root, "tmp", "app.db") + "\n" +
			"i18n:\n  default_locale: zh-CN\n  fallback_locale: en-US\n  locale_dir: " + localeDir + "\n" +
			"jwt:\n  enabled: true\n  issuer: go-scaffold2-test\n  secret: test-secret\n  access_token_ttl_minutes: 60\n  refresh_token_ttl_hours: 72\n" +
			"rbac:\n  enabled: true\n  model_path: " + modelPath + "\n  auto_save: true\n" +
			"initdb:\n  enabled: true\n  driver: sqlite\n  output_dir: " + filepath.Join(root, "scripts", "initdb") + "\n  lock_file: " + filepath.Join(root, "scripts", "initdb", ".initdb.lock") + "\n" +
			"storage:\n  enabled: true\n  driver: local\n  root_dir: " + filepath.Join(root, "storage") + "\n",
	)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	return configPath
}

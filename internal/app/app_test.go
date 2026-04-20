package app

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	pkgdatabase "github.com/rin721/rei/pkg/database"
)

func TestAppRunServerDryRun(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t)
	prepareTestSchema(t, configPath)

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

func TestAppRunDBMigrateAppliesPendingMigrations(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t)

	application, err := New(Options{
		Mode:       ModeDB,
		ConfigPath: configPath,
		DBOptions: DBOptions{
			Action: DBActionMigrate,
		},
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	dbPath := testDBPath(configPath)
	if !sqliteTableExists(t, dbPath, "users") {
		t.Fatal("users table was not created by db migrate")
	}
	if !migrationRecordExists(t, dbPath, "20260420_001") {
		t.Fatal("schema_migrations does not contain 20260420_001")
	}
}

func TestAppRunDBMigrateDryRunSkipsChanges(t *testing.T) {
	t.Parallel()

	configPath := writeTestConfig(t)

	application, err := New(Options{
		Mode:       ModeDB,
		ConfigPath: configPath,
		DryRun:     true,
		DBOptions: DBOptions{
			Action: DBActionMigrate,
			DryRun: true,
		},
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() returned error: %v", err)
	}

	dbPath := testDBPath(configPath)
	if sqliteTableExists(t, dbPath, "users") {
		t.Fatal("users table should not be created during dry-run")
	}
	if sqliteTableExists(t, dbPath, "schema_migrations") {
		t.Fatal("schema_migrations should not be created during dry-run")
	}
}

func prepareTestSchema(t *testing.T, configPath string) {
	t.Helper()

	application, err := New(Options{
		Mode:       ModeDB,
		ConfigPath: configPath,
		DBOptions: DBOptions{
			Action: DBActionMigrate,
		},
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	if err := application.Run(context.Background()); err != nil {
		t.Fatalf("Run() returned error: %v", err)
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
			"database:\n  enabled: true\n  driver: sqlite\n  name: " + filepath.Join(root, "tmp", "app.db") + "\n  migrations_dir: " + testMigrationsDir(t) + "\n" +
			"i18n:\n  default_locale: zh-CN\n  fallback_locale: en-US\n  locale_dir: " + localeDir + "\n" +
			"jwt:\n  enabled: true\n  issuer: go-scaffold2-test\n  secret: test-secret\n  access_token_ttl_minutes: 60\n  refresh_token_ttl_hours: 72\n" +
			"rbac:\n  enabled: true\n  model_path: " + modelPath + "\n  auto_save: true\n" +
			"storage:\n  enabled: true\n  driver: local\n  root_dir: " + filepath.Join(root, "storage") + "\n",
	)
	if err := os.WriteFile(configPath, content, 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	return configPath
}

func testMigrationsDir(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	return filepath.Join(filepath.Dir(filepath.Dir(filepath.Dir(file))), "scripts", "migrations")
}

func testDBPath(configPath string) string {
	return filepath.Join(filepath.Dir(filepath.Dir(configPath)), "tmp", "app.db")
}

func sqliteTableExists(t *testing.T, dbPath, table string) bool {
	t.Helper()

	store, err := pkgdatabase.New(pkgdatabase.Config{
		Driver:          "sqlite",
		DSN:             dbPath,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
	})
	if err != nil {
		t.Fatalf("pkgdatabase.New() returned error: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	var count int64
	if err := store.DB().
		Raw("SELECT COUNT(1) FROM sqlite_master WHERE type = 'table' AND name = ?", table).
		Scan(&count).Error; err != nil {
		t.Fatalf("table existence query returned error: %v", err)
	}

	return count > 0
}

func migrationRecordExists(t *testing.T, dbPath, version string) bool {
	t.Helper()

	if !sqliteTableExists(t, dbPath, "schema_migrations") {
		return false
	}

	store, err := pkgdatabase.New(pkgdatabase.Config{
		Driver:          "sqlite",
		DSN:             dbPath,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 0,
	})
	if err != nil {
		t.Fatalf("pkgdatabase.New() returned error: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	var count int64
	if err := store.DB().
		Raw("SELECT COUNT(1) FROM schema_migrations WHERE version = ?", version).
		Scan(&count).Error; err != nil {
		t.Fatalf("migration record query returned error: %v", err)
	}

	return count > 0
}

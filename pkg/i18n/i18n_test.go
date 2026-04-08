package i18n

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManagerLocalizeAndReload(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "zh-CN.yaml"), []byte("welcome: \"你好，{{.Name}}\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "en-US.yaml"), []byte("welcome: \"Hello, {{.Name}}\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	manager, err := New(Config{
		DefaultLocale:  "zh-CN",
		FallbackLocale: "en-US",
		LocaleDir:      dir,
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	got, err := manager.Localize("en-US", "welcome", map[string]any{"Name": "Vibe"})
	if err != nil {
		t.Fatalf("Localize() returned error: %v", err)
	}
	if got != "Hello, Vibe" {
		t.Fatalf("Localize() = %q, want %q", got, "Hello, Vibe")
	}

	if err := os.WriteFile(filepath.Join(dir, "en-US.yaml"), []byte("welcome: \"Hi, {{.Name}}\"\n"), 0o644); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	if err := manager.Reload(Config{
		DefaultLocale:  "zh-CN",
		FallbackLocale: "en-US",
		LocaleDir:      dir,
	}); err != nil {
		t.Fatalf("Reload() returned error: %v", err)
	}

	got, err = manager.Localize("en-US", "welcome", map[string]any{"Name": "Vibe"})
	if err != nil {
		t.Fatalf("Localize() after reload returned error: %v", err)
	}
	if got != "Hi, Vibe" {
		t.Fatalf("Localize() after reload = %q, want %q", got, "Hi, Vibe")
	}
}

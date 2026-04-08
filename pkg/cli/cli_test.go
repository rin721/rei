package cli

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

func TestRegisterRejectsDuplicateCommand(t *testing.T) {
	t.Parallel()

	app := New("test", "test cli")

	err := app.Register(Command{
		Name:        "server",
		Description: "first registration",
		Run: func(context.Context, []string) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() first call returned error: %v", err)
	}

	err = app.Register(Command{
		Name:        "server",
		Description: "duplicate registration",
		Run: func(context.Context, []string) error {
			return nil
		},
	})
	if err == nil {
		t.Fatal("Register() duplicate call returned nil error")
	}
}

func TestRunExecutesKnownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := New("test", "test cli")
	app.SetOutput(&stdout, &stderr)

	err := app.Register(Command{
		Name:        "server",
		Description: "runs server",
		Run: func(_ context.Context, args []string) error {
			if len(args) != 1 || args[0] != "--dry-run" {
				t.Fatalf("args = %#v, want %#v", args, []string{"--dry-run"})
			}
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	code := app.Run(context.Background(), []string{"server", "--dry-run"})

	if code != ExitCodeSuccess {
		t.Fatalf("Run() code = %d, want %d", code, ExitCodeSuccess)
	}

	if stderr.Len() != 0 {
		t.Fatalf("stderr = %q, want empty", stderr.String())
	}
}

func TestRunShowsUsageForUnknownCommand(t *testing.T) {
	t.Parallel()

	var stdout bytes.Buffer
	var stderr bytes.Buffer

	app := New("test", "test cli")
	app.SetOutput(&stdout, &stderr)

	err := app.Register(Command{
		Name:        "initdb",
		Description: "runs initdb",
		Run: func(context.Context, []string) error {
			return nil
		},
	})
	if err != nil {
		t.Fatalf("Register() returned error: %v", err)
	}

	code := app.Run(context.Background(), []string{"unknown"})

	if code != ExitCodeUsage {
		t.Fatalf("Run() code = %d, want %d", code, ExitCodeUsage)
	}

	if !strings.Contains(stderr.String(), "unknown command: unknown") {
		t.Fatalf("stderr = %q, want unknown command message", stderr.String())
	}
}

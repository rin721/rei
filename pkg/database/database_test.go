package database

import (
	"context"
	"testing"
)

func TestDatabasePingAndReload(t *testing.T) {
	t.Parallel()

	store, err := New(Config{
		Driver: "sqlite",
		DSN:    "file:phase2_db_1?mode=memory&cache=shared",
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() returned error: %v", err)
	}

	if err := store.Reload(Config{
		Driver: "sqlite",
		DSN:    "file:phase2_db_2?mode=memory&cache=shared",
	}); err != nil {
		t.Fatalf("Reload() returned error: %v", err)
	}

	if err := store.Ping(context.Background()); err != nil {
		t.Fatalf("Ping() after reload returned error: %v", err)
	}
}

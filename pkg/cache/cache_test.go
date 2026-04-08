package cache

import (
	"context"
	"testing"
	"time"
)

func TestMemoryCacheSetGetAndExpire(t *testing.T) {
	t.Parallel()

	store := New(Config{DefaultTTL: time.Second})

	if err := store.Set(context.Background(), "name", "vibe", 50*time.Millisecond); err != nil {
		t.Fatalf("Set() returned error: %v", err)
	}

	value, ok := store.Get(context.Background(), "name")
	if !ok || value != "vibe" {
		t.Fatalf("Get() = (%v, %v), want (%q, true)", value, ok, "vibe")
	}

	time.Sleep(70 * time.Millisecond)

	if _, ok := store.Get(context.Background(), "name"); ok {
		t.Fatal("Get() should report expired item as missing")
	}
}

func TestMemoryCacheIncr(t *testing.T) {
	t.Parallel()

	store := New(Config{})

	got, err := store.Incr(context.Background(), "counter", 2)
	if err != nil {
		t.Fatalf("Incr() returned error: %v", err)
	}
	if got != 2 {
		t.Fatalf("Incr() = %d, want 2", got)
	}

	got, err = store.Incr(context.Background(), "counter", 3)
	if err != nil {
		t.Fatalf("Incr() returned error: %v", err)
	}
	if got != 5 {
		t.Fatalf("Incr() = %d, want 5", got)
	}
}

package executor

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestManagerSubmitDefault(t *testing.T) {
	t.Parallel()

	manager, err := New(Config{
		Pools: []PoolConfig{{Name: DefaultPoolName, Workers: 2, QueueSize: 2}},
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() {
		_ = manager.Shutdown(context.Background())
	}()

	var executed atomic.Int64
	done := make(chan struct{})

	err = manager.SubmitDefault(context.Background(), func() {
		executed.Add(1)
		close(done)
	})
	if err != nil {
		t.Fatalf("SubmitDefault() returned error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("task was not executed in time")
	}

	if executed.Load() != 1 {
		t.Fatalf("executed = %d, want 1", executed.Load())
	}
}

func TestManagerReload(t *testing.T) {
	t.Parallel()

	manager, err := New(Config{
		Pools: []PoolConfig{{Name: DefaultPoolName, Workers: 1, QueueSize: 1}},
	})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() {
		_ = manager.Shutdown(context.Background())
	}()

	err = manager.Reload(Config{
		DefaultPool: "jobs",
		Pools: []PoolConfig{
			{Name: "jobs", Workers: 1, QueueSize: 1},
		},
	})
	if err != nil {
		t.Fatalf("Reload() returned error: %v", err)
	}

	done := make(chan struct{})
	err = manager.SubmitDefault(context.Background(), func() {
		close(done)
	})
	if err != nil {
		t.Fatalf("SubmitDefault() after reload returned error: %v", err)
	}

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("reloaded pool did not execute task in time")
	}
}

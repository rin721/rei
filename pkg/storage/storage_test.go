package storage

import (
	"context"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestStorageWriteReadExistsAndImageInfo(t *testing.T) {
	t.Parallel()

	store, err := New(Config{RootDir: t.TempDir()})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	if err := store.WriteFile(context.Background(), "docs/info.txt", []byte("hello")); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	content, err := store.ReadFile(context.Background(), "docs/info.txt")
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}
	if string(content) != "hello" {
		t.Fatalf("ReadFile() = %q, want %q", string(content), "hello")
	}

	exists, err := store.Exists(context.Background(), "docs/info.txt")
	if err != nil {
		t.Fatalf("Exists() returned error: %v", err)
	}
	if !exists {
		t.Fatal("Exists() = false, want true")
	}

	imagePath := filepath.Join(t.TempDir(), "sample.png")
	file, err := os.Create(imagePath)
	if err != nil {
		t.Fatalf("Create() returned error: %v", err)
	}
	if err := png.Encode(file, image.NewRGBA(image.Rect(0, 0, 2, 3))); err != nil {
		t.Fatalf("Encode() returned error: %v", err)
	}
	if err := file.Close(); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}

	imageBytes, err := os.ReadFile(imagePath)
	if err != nil {
		t.Fatalf("ReadFile() returned error: %v", err)
	}
	if err := store.WriteFile(context.Background(), "images/sample.png", imageBytes); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	info, err := store.ImageInfo(context.Background(), "images/sample.png")
	if err != nil {
		t.Fatalf("ImageInfo() returned error: %v", err)
	}
	if info.Width != 2 || info.Height != 3 || info.Format != "png" {
		t.Fatalf("ImageInfo() = %#v, want width=2 height=3 format=png", info)
	}
}

func TestStorageWatch(t *testing.T) {
	t.Parallel()

	store, err := New(Config{RootDir: t.TempDir()})
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}
	defer func() {
		_ = store.Close()
	}()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events := make(chan Event, 1)
	stop, err := store.Watch(ctx, "watch.txt", 10*time.Millisecond, func(event Event) {
		select {
		case events <- event:
		default:
		}
	})
	if err != nil {
		t.Fatalf("Watch() returned error: %v", err)
	}
	defer func() {
		_ = stop()
	}()

	if err := store.WriteFile(context.Background(), "watch.txt", []byte("changed")); err != nil {
		t.Fatalf("WriteFile() returned error: %v", err)
	}

	select {
	case event := <-events:
		if event.Op != "create" && event.Op != "write" {
			t.Fatalf("event.Op = %q, want create or write", event.Op)
		}
	case <-time.After(time.Second):
		t.Fatal("watch event was not delivered in time")
	}
}

package config

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"sync"
	"time"
)

// Watcher 负责监听配置文件变化。
type Watcher struct {
	path     string
	interval time.Duration
	stopCh   chan struct{}
	stopOnce sync.Once
	wg       sync.WaitGroup
}

// NewWatcher 创建一个新的配置文件监听器。
func NewWatcher(path string, interval time.Duration) *Watcher {
	if interval <= 0 {
		interval = 200 * time.Millisecond
	}

	return &Watcher{
		path:     path,
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

// Start 启动监听。
func (w *Watcher) Start(ctx context.Context, onChange func() error) error {
	if onChange == nil {
		return fmt.Errorf("onChange is required")
	}

	lastSignature := readSignature(w.path)
	w.wg.Add(1)

	go func() {
		defer w.wg.Done()

		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-w.stopCh:
				return
			case <-ticker.C:
				signature := readSignature(w.path)
				if signature == lastSignature {
					continue
				}

				lastSignature = signature
				_ = onChange()
			}
		}
	}()

	return nil
}

// Stop 停止监听。
func (w *Watcher) Stop() error {
	w.stopOnce.Do(func() {
		close(w.stopCh)
	})
	w.wg.Wait()
	return nil
}

func readSignature(path string) string {
	content, err := os.ReadFile(path)
	if err != nil {
		return "missing"
	}

	sum := sha256.Sum256(content)
	return fmt.Sprintf("%x", sum)
}

package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/afero"
)

// Watch 监听文件或目录的变化
func (i *impl) Watch(path string, handler WatchHandler) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 检查是否启用监听功能
	if i.watcher == nil {
		return fmt.Errorf("Storage: watch is not enabled")
	}

	// 检查路径是否已被监听
	if _, exists := i.watches[path]; exists {
		return ErrWatcherAlreadyExists
	}

	// 检查路径是否存在
	exists, err := afero.Exists(i.fs, path)
	if err != nil {
		return fmt.Errorf("Storage: failed to check path: %w", err)
	}
	if !exists {
		return fmt.Errorf("%w: %s", ErrPathNotFound, path)
	}

	// 添加到 watcher
	if err := i.watcher.Add(path); err != nil {
		return fmt.Errorf("Storage: failed to add watcher: %w", err)
	}

	// 创建取消上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 创建监听条目
	entry := &watchEntry{
		path:    path,
		handler: handler,
		cancel:  cancel,
	}
	i.watches[path] = entry

	// 启动事件处理 goroutine
	go i.handleWatchEvents(ctx, entry)

	return nil
}

// handleWatchEvents 处理文件监听事件
func (i *impl) handleWatchEvents(ctx context.Context, entry *watchEntry) {
	for {
		select {
		case <-ctx.Done():
			// 监听被取消
			return

		case event, ok := <-i.watcher.Events:
			if !ok {
				return
			}

			// 检查事件路径是否匹配
			if event.Name != entry.path {
				continue
			}

			// 转换为 WatchEvent
			watchEvent := i.convertFsnotifyEvent(event)

			// 调用处理函数
			entry.handler(watchEvent)

		case err, ok := <-i.watcher.Errors:
			if !ok {
				return
			}

			// 发送错误事件
			watchEvent := WatchEvent{
				Path:  entry.path,
				Op:    "ERROR",
				Time:  time.Now(),
				IsDir: false,
			}
			entry.handler(watchEvent)

			// 记录错误(可选)
			_ = err // 忽略错误,或者可以通过日志记录
		}
	}
}

// convertFsnotifyEvent 转换 fsnotify 事件为 WatchEvent
func (i *impl) convertFsnotifyEvent(event fsnotify.Event) WatchEvent {
	var op string
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		op = WatchEventCreate
	case event.Op&fsnotify.Write == fsnotify.Write:
		op = WatchEventWrite
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		op = WatchEventRemove
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		op = WatchEventRename
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		op = WatchEventChmod
	default:
		op = "UNKNOWN"
	}

	// 检查是否为目录
	isDir := false
	if info, err := i.fs.Stat(event.Name); err == nil {
		isDir = info.IsDir()
	}

	return WatchEvent{
		Path:  event.Name,
		Op:    op,
		Time:  time.Now(),
		IsDir: isDir,
	}
}

// StopWatch 停止监听指定路径
func (i *impl) StopWatch(path string) error {
	i.mu.Lock()
	defer i.mu.Unlock()

	entry, exists := i.watches[path]
	if !exists {
		return ErrWatcherNotFound
	}

	// 取消事件处理 goroutine
	entry.cancel()

	// 从 watcher 中移除
	if err := i.watcher.Remove(path); err != nil {
		return fmt.Errorf("Storage: failed to remove watcher: %w", err)
	}

	// 从 map 中删除
	delete(i.watches, path)

	return nil
}

// StopAllWatch 停止所有监听
func (i *impl) StopAllWatch() {
	i.mu.Lock()
	defer i.mu.Unlock()

	// 取消所有监听
	for path, entry := range i.watches {
		entry.cancel()
		i.watcher.Remove(path)
	}

	// 清空 map
	i.watches = make(map[string]*watchEntry)
}

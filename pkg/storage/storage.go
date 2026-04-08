package storage

import (
	"context"
	"errors"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// Config 描述本地存储配置。
type Config struct {
	RootDir string
}

// Event 描述文件变更事件。
type Event struct {
	Path string
	Op   string
}

// ImageInfo 描述图片的基本元信息。
type ImageInfo struct {
	Width  int
	Height int
	Format string
}

// Storage 提供最小本地文件系统封装。
type Storage struct {
	mu         sync.RWMutex
	rootDir    string
	closed     bool
	watchers   map[int]chan struct{}
	nextID     int
	watchersWG sync.WaitGroup
}

// New 创建一个新的本地存储实例。
func New(cfg Config) (*Storage, error) {
	rootDir, err := normalizeRoot(cfg.RootDir)
	if err != nil {
		return nil, err
	}

	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return nil, err
	}

	return &Storage{
		rootDir:  rootDir,
		watchers: make(map[int]chan struct{}),
	}, nil
}

// WriteFile 将数据写入相对路径文件。
func (s *Storage) WriteFile(ctx context.Context, relativePath string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.ensureOpen(); err != nil {
		return err
	}

	target, err := s.resolvePath(relativePath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}

	return os.WriteFile(target, data, 0o644)
}

// ReadFile 读取相对路径文件。
func (s *Storage) ReadFile(ctx context.Context, relativePath string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := s.ensureOpen(); err != nil {
		return nil, err
	}

	target, err := s.resolvePath(relativePath)
	if err != nil {
		return nil, err
	}

	return os.ReadFile(target)
}

// Delete 删除相对路径文件。
func (s *Storage) Delete(ctx context.Context, relativePath string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := s.ensureOpen(); err != nil {
		return err
	}

	target, err := s.resolvePath(relativePath)
	if err != nil {
		return err
	}

	if err := os.Remove(target); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}

// Exists 判断相对路径文件是否存在。
func (s *Storage) Exists(ctx context.Context, relativePath string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	if err := s.ensureOpen(); err != nil {
		return false, err
	}

	target, err := s.resolvePath(relativePath)
	if err != nil {
		return false, err
	}

	_, err = os.Stat(target)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

// Watch 轮询监听相对路径文件变化。
func (s *Storage) Watch(ctx context.Context, relativePath string, interval time.Duration, handler func(Event)) (func() error, error) {
	if handler == nil {
		return nil, errors.New("watch handler is required")
	}
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	if err := s.ensureOpen(); err != nil {
		return nil, err
	}

	target, err := s.resolvePath(relativePath)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, errors.New("storage is closed")
	}
	id := s.nextID
	s.nextID++
	stopCh := make(chan struct{})
	s.watchers[id] = stopCh
	s.watchersWG.Add(1)
	s.mu.Unlock()

	exists, modTime := fileState(target)
	var once sync.Once

	go func() {
		defer s.watchersWG.Done()
		defer s.unregisterWatcher(id)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-stopCh:
				return
			case <-ticker.C:
				nextExists, nextModTime := fileState(target)
				switch {
				case !exists && nextExists:
					handler(Event{Path: relativePath, Op: "create"})
				case exists && !nextExists:
					handler(Event{Path: relativePath, Op: "delete"})
				case exists && nextExists && nextModTime.After(modTime):
					handler(Event{Path: relativePath, Op: "write"})
				}
				exists = nextExists
				modTime = nextModTime
			}
		}
	}()

	return func() error {
		once.Do(func() {
			safeClose(stopCh)
		})
		return nil
	}, nil
}

// ImageInfo 返回图片尺寸与格式信息。
func (s *Storage) ImageInfo(ctx context.Context, relativePath string) (ImageInfo, error) {
	if err := ctx.Err(); err != nil {
		return ImageInfo{}, err
	}
	if err := s.ensureOpen(); err != nil {
		return ImageInfo{}, err
	}

	target, err := s.resolvePath(relativePath)
	if err != nil {
		return ImageInfo{}, err
	}

	file, err := os.Open(target)
	if err != nil {
		return ImageInfo{}, err
	}
	defer file.Close()

	config, format, err := image.DecodeConfig(file)
	if err != nil {
		return ImageInfo{}, err
	}

	return ImageInfo{
		Width:  config.Width,
		Height: config.Height,
		Format: format,
	}, nil
}

// Reload 更新存储根目录配置。
func (s *Storage) Reload(cfg Config) error {
	rootDir, err := normalizeRoot(cfg.RootDir)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(rootDir, 0o755); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("storage is closed")
	}
	s.rootDir = rootDir
	return nil
}

// Close 关闭存储并停止所有 watcher。
func (s *Storage) Close() error {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil
	}
	s.closed = true
	watchers := s.watchers
	s.watchers = make(map[int]chan struct{})
	s.mu.Unlock()

	for _, stopCh := range watchers {
		safeClose(stopCh)
	}

	s.watchersWG.Wait()
	return nil
}

func (s *Storage) ensureOpen() error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed {
		return errors.New("storage is closed")
	}
	return nil
}

func (s *Storage) resolvePath(relativePath string) (string, error) {
	s.mu.RLock()
	rootDir := s.rootDir
	s.mu.RUnlock()

	cleaned := filepath.Clean(relativePath)
	target := filepath.Join(rootDir, cleaned)
	targetAbs, err := filepath.Abs(target)
	if err != nil {
		return "", err
	}

	if targetAbs != rootDir && !strings.HasPrefix(targetAbs, rootDir+string(filepath.Separator)) {
		return "", errors.New("path escapes storage root")
	}

	return targetAbs, nil
}

func (s *Storage) unregisterWatcher(id int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.watchers, id)
}

func normalizeRoot(rootDir string) (string, error) {
	if rootDir == "" {
		rootDir = "tmp/storage"
	}
	return filepath.Abs(rootDir)
}

func fileState(path string) (bool, time.Time) {
	info, err := os.Stat(path)
	if err != nil {
		return false, time.Time{}
	}
	return true, info.ModTime()
}

func safeClose(ch chan struct{}) {
	defer func() {
		_ = recover()
	}()
	close(ch)
}

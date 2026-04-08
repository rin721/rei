package config

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// ReloadHook 定义配置切换成功前的组件重载钩子。
type ReloadHook func(context.Context, Config, Config) error

type reloadHookEntry struct {
	name string
	fn   ReloadHook
}

// Options 描述配置管理器选项。
type Options struct {
	Path          string
	WatchInterval time.Duration
}

// Manager 管理配置加载、覆盖和热重载。
type Manager struct {
	path          string
	watchInterval time.Duration

	mu      sync.RWMutex
	current Config
	loaded  bool
	hooks   []reloadHookEntry
	watcher *Watcher
}

// NewManager 创建一个新的配置管理器。
func NewManager(options Options) *Manager {
	if options.WatchInterval <= 0 {
		options.WatchInterval = 200 * time.Millisecond
	}

	return &Manager{
		path:          options.Path,
		watchInterval: options.WatchInterval,
	}
}

// Load 从文件加载并验证配置。
func (m *Manager) Load() (Config, error) {
	cfg, err := m.loadCandidate()
	if err != nil {
		return Config{}, err
	}

	m.mu.Lock()
	m.current = cfg.Clone()
	m.loaded = true
	m.mu.Unlock()

	return cfg.Clone(), nil
}

// Current 返回当前配置快照。
func (m *Manager) Current() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current.Clone()
}

// RegisterReloadHook 注册一个配置热重载钩子。
func (m *Manager) RegisterReloadHook(name string, fn ReloadHook) {
	if fn == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.hooks = append(m.hooks, reloadHookEntry{
		name: name,
		fn:   fn,
	})
}

// Start 启动配置文件监听。
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	if !m.loaded {
		m.mu.Unlock()
		if _, err := m.Load(); err != nil {
			return err
		}
		m.mu.Lock()
	}
	if m.watcher != nil {
		m.mu.Unlock()
		return nil
	}

	watcher := NewWatcher(m.path, m.watchInterval)
	m.watcher = watcher
	m.mu.Unlock()

	return watcher.Start(ctx, func() error {
		return m.reload(context.Background())
	})
}

// Stop 停止配置文件监听。
func (m *Manager) Stop() error {
	m.mu.Lock()
	watcher := m.watcher
	m.watcher = nil
	m.mu.Unlock()

	if watcher == nil {
		return nil
	}

	return watcher.Stop()
}

func (m *Manager) reload(ctx context.Context) error {
	next, err := m.loadCandidate()
	if err != nil {
		return err
	}

	m.mu.RLock()
	current := m.current.Clone()
	hooks := append([]reloadHookEntry(nil), m.hooks...)
	m.mu.RUnlock()

	for _, hook := range hooks {
		if err := hook.fn(ctx, current, next); err != nil {
			return fmt.Errorf("reload hook %q: %w", hook.name, err)
		}
	}

	m.mu.Lock()
	m.current = next.Clone()
	m.loaded = true
	m.mu.Unlock()
	return nil
}

func (m *Manager) loadCandidate() (Config, error) {
	if m.path == "" {
		return Config{}, fmt.Errorf("config path is required")
	}

	content, err := os.ReadFile(m.path)
	if err != nil {
		return Config{}, err
	}

	cfg := Default()
	content = ExpandEnvPlaceholders(content)
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return Config{}, err
	}
	if err := ApplyEnvOverrides(&cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg.Clone(), nil
}

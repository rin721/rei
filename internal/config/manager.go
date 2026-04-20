package config

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

// ReloadHook defines a callback that runs before a new config snapshot becomes active.
type ReloadHook func(context.Context, Config, Config) error

type reloadHookEntry struct {
	name string
	fn   ReloadHook
}

// Options describes manager setup.
type Options struct {
	Path          string
	WatchInterval time.Duration
}

// Manager loads, validates, and hot-reloads config files.
type Manager struct {
	path          string
	watchInterval time.Duration

	mu      sync.RWMutex
	current Config
	loaded  bool
	hooks   []reloadHookEntry
	watcher *Watcher
}

func NewManager(options Options) *Manager {
	if options.WatchInterval <= 0 {
		options.WatchInterval = 200 * time.Millisecond
	}

	return &Manager{
		path:          options.Path,
		watchInterval: options.WatchInterval,
	}
}

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

func (m *Manager) Current() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.current.Clone()
}

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
	if err := rejectLegacySchemaConfig(content); err != nil {
		return Config{}, err
	}
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

func rejectLegacySchemaConfig(content []byte) error {
	var node yaml.Node
	decoder := yaml.NewDecoder(bytes.NewReader(content))
	if err := decoder.Decode(&node); err != nil {
		return err
	}

	if len(node.Content) == 0 {
		return nil
	}

	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil
	}

	for i := 0; i+1 < len(root.Content); i += 2 {
		key := root.Content[i]
		if key.Value == "initdb" {
			return fmt.Errorf("legacy schema config field %q has been removed; use database.migrations_dir with cmd/db instead", key.Value)
		}
	}

	return nil
}

package executor

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// DefaultPoolName 是默认任务池名称。
const DefaultPoolName = "default"

// PoolConfig 描述单个任务池的配置。
type PoolConfig struct {
	Name      string
	Workers   int
	QueueSize int
}

// Config 描述执行器管理器配置。
type Config struct {
	DefaultPool string
	Pools       []PoolConfig
}

// Manager 管理多个任务池。
type Manager struct {
	mu          sync.RWMutex
	defaultPool string
	pools       map[string]*pool
	closed      bool
}

type pool struct {
	closeOnce sync.Once
	tasks     chan func()
	wg        sync.WaitGroup
}

// New 创建一个新的执行器管理器。
func New(cfg Config) (*Manager, error) {
	normalized, err := normalizeConfig(cfg)
	if err != nil {
		return nil, err
	}

	manager := &Manager{
		defaultPool: normalized.DefaultPool,
		pools:       make(map[string]*pool, len(normalized.Pools)),
	}

	for _, poolCfg := range normalized.Pools {
		manager.pools[poolCfg.Name] = newPool(poolCfg)
	}

	return manager, nil
}

// Submit 向指定任务池提交任务。
func (m *Manager) Submit(ctx context.Context, poolName string, task func()) error {
	if task == nil {
		return errors.New("task is required")
	}

	m.mu.RLock()
	if m.closed {
		m.mu.RUnlock()
		return errors.New("executor is closed")
	}

	poolRef, exists := m.pools[poolName]
	m.mu.RUnlock()
	if !exists {
		return fmt.Errorf("pool %q not found", poolName)
	}

	return poolRef.submit(ctx, task)
}

// SubmitDefault 向默认任务池提交任务。
func (m *Manager) SubmitDefault(ctx context.Context, task func()) error {
	m.mu.RLock()
	defaultPool := m.defaultPool
	m.mu.RUnlock()

	return m.Submit(ctx, defaultPool, task)
}

// Reload 使用新配置重建任务池并原子替换旧实例。
func (m *Manager) Reload(cfg Config) error {
	next, err := New(cfg)
	if err != nil {
		return err
	}

	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		_ = next.Shutdown(context.Background())
		return errors.New("executor is closed")
	}

	oldPools := m.pools
	m.defaultPool = next.defaultPool
	m.pools = next.pools
	m.mu.Unlock()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for _, poolRef := range oldPools {
		if err := poolRef.shutdown(shutdownCtx); err != nil {
			return err
		}
	}

	return nil
}

// Shutdown 优雅关闭所有任务池。
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	if m.closed {
		m.mu.Unlock()
		return nil
	}

	m.closed = true
	pools := m.pools
	m.mu.Unlock()

	for _, poolRef := range pools {
		if err := poolRef.shutdown(ctx); err != nil {
			return err
		}
	}

	return nil
}

func normalizeConfig(cfg Config) (Config, error) {
	if len(cfg.Pools) == 0 {
		cfg.Pools = []PoolConfig{{Name: DefaultPoolName, Workers: 1, QueueSize: 1}}
	}

	if cfg.DefaultPool == "" {
		cfg.DefaultPool = cfg.Pools[0].Name
	}

	seen := make(map[string]struct{}, len(cfg.Pools))
	for index, poolCfg := range cfg.Pools {
		if poolCfg.Name == "" {
			return Config{}, fmt.Errorf("pool[%d] name is required", index)
		}
		if _, exists := seen[poolCfg.Name]; exists {
			return Config{}, fmt.Errorf("pool %q duplicated", poolCfg.Name)
		}
		seen[poolCfg.Name] = struct{}{}

		if poolCfg.Workers <= 0 {
			poolCfg.Workers = 1
		}
		if poolCfg.QueueSize <= 0 {
			poolCfg.QueueSize = poolCfg.Workers
		}
		cfg.Pools[index] = poolCfg
	}

	if _, exists := seen[cfg.DefaultPool]; !exists {
		return Config{}, fmt.Errorf("default pool %q not found", cfg.DefaultPool)
	}

	return cfg, nil
}

func newPool(cfg PoolConfig) *pool {
	poolRef := &pool{
		tasks: make(chan func(), cfg.QueueSize),
	}

	for index := 0; index < cfg.Workers; index++ {
		poolRef.wg.Add(1)
		go func() {
			defer poolRef.wg.Done()
			for task := range poolRef.tasks {
				if task != nil {
					task()
				}
			}
		}()
	}

	return poolRef
}

func (p *pool) submit(ctx context.Context, task func()) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case p.tasks <- task:
		return nil
	}
}

func (p *pool) shutdown(ctx context.Context) error {
	p.closeOnce.Do(func() {
		close(p.tasks)
	})

	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-done:
		return nil
	}
}

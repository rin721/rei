package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

// Config 描述缓存实例配置。
type Config struct {
	DefaultTTL time.Duration
}

type item struct {
	value     any
	expiresAt time.Time
}

// MemoryCache 提供最小可用的内存缓存。
type MemoryCache struct {
	mu         sync.RWMutex
	items      map[string]item
	defaultTTL time.Duration
	closed     bool
}

// New 创建一个新的内存缓存。
func New(cfg Config) *MemoryCache {
	return &MemoryCache{
		items:      make(map[string]item),
		defaultTTL: cfg.DefaultTTL,
	}
}

// Get 获取指定 key 的值。
func (c *MemoryCache) Get(ctx context.Context, key string) (any, bool) {
	if err := ctx.Err(); err != nil {
		return nil, false
	}
	if err := c.ensureOpen(); err != nil {
		return nil, false
	}

	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || isExpired(entry) {
		if ok {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
		}
		return nil, false
	}

	return entry.value, true
}

// Set 写入指定 key 的值。
func (c *MemoryCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := c.ensureOpen(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.items[key] = item{
		value:     value,
		expiresAt: c.expiration(ttl),
	}

	return nil
}

// Delete 删除指定 key。
func (c *MemoryCache) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := c.ensureOpen(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.items, key)
	return nil
}

// Exists 判断指定 key 是否存在。
func (c *MemoryCache) Exists(ctx context.Context, key string) bool {
	_, ok := c.Get(ctx, key)
	return ok
}

// MGet 批量获取多个 key。
func (c *MemoryCache) MGet(ctx context.Context, keys ...string) map[string]any {
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		if value, ok := c.Get(ctx, key); ok {
			result[key] = value
		}
	}
	return result
}

// MSet 批量写入多个 key。
func (c *MemoryCache) MSet(ctx context.Context, values map[string]any, ttl time.Duration) error {
	for key, value := range values {
		if err := c.Set(ctx, key, value, ttl); err != nil {
			return err
		}
	}
	return nil
}

// Expire 更新指定 key 的过期时间。
func (c *MemoryCache) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := c.ensureOpen(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry, ok := c.items[key]
	if !ok || isExpired(entry) {
		delete(c.items, key)
		return fmt.Errorf("cache key %q not found", key)
	}

	entry.expiresAt = c.expiration(ttl)
	c.items[key] = entry
	return nil
}

// TTL 返回指定 key 的剩余过期时间。
//
// 当 key 永不过期时，返回 `-1` 和 `true`。
func (c *MemoryCache) TTL(ctx context.Context, key string) (time.Duration, bool) {
	if err := ctx.Err(); err != nil {
		return 0, false
	}

	c.mu.RLock()
	entry, ok := c.items[key]
	c.mu.RUnlock()
	if !ok || isExpired(entry) {
		if ok {
			c.mu.Lock()
			delete(c.items, key)
			c.mu.Unlock()
		}
		return 0, false
	}

	if entry.expiresAt.IsZero() {
		return -1, true
	}

	return time.Until(entry.expiresAt), true
}

// Incr 对整数值执行增量操作。
func (c *MemoryCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if err := c.ensureOpen(); err != nil {
		return 0, err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	entry := c.items[key]
	if isExpired(entry) {
		entry = item{}
	}

	var current int64
	switch value := entry.value.(type) {
	case nil:
		current = 0
	case int:
		current = int64(value)
	case int64:
		current = value
	default:
		return 0, fmt.Errorf("cache key %q is not an integer", key)
	}

	current += delta
	entry.value = current
	if entry.expiresAt.IsZero() {
		entry.expiresAt = c.expiration(0)
	}
	c.items[key] = entry
	return current, nil
}

// Reload 原子更新缓存配置。
func (c *MemoryCache) Reload(cfg Config) error {
	if err := c.ensureOpen(); err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.defaultTTL = cfg.DefaultTTL
	return nil
}

// Close 关闭缓存并清理所有数据。
func (c *MemoryCache) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.closed = true
	c.items = make(map[string]item)
	return nil
}

func (c *MemoryCache) ensureOpen() error {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.closed {
		return errors.New("cache is closed")
	}
	return nil
}

func (c *MemoryCache) expiration(ttl time.Duration) time.Time {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}
	if ttl <= 0 {
		return time.Time{}
	}
	return time.Now().Add(ttl)
}

func isExpired(entry item) bool {
	return !entry.expiresAt.IsZero() && time.Now().After(entry.expiresAt)
}

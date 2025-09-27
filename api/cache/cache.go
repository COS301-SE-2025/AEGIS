package cache

import (
	"context"
	"sync"
	"time"
)

type Client interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) (int, error)
}

// ---------------- In-Memory (fallback) ----------------

type memoryItem struct {
	value string
	exp   time.Time
}

type Memory struct {
	mu sync.RWMutex
	m  map[string]memoryItem
}

func NewMemory() *Memory { return &Memory{m: make(map[string]memoryItem)} }

func (c *Memory) Get(_ context.Context, key string) (string, bool, error) {
	c.mu.RLock()
	it, ok := c.m[key]
	c.mu.RUnlock()
	if !ok {
		return "", false, nil
	}
	if !it.exp.IsZero() && time.Now().After(it.exp) {
		c.mu.Lock()
		// recheck & delete under write lock
		if cur, ok := c.m[key]; ok && cur.exp.Equal(it.exp) {
			delete(c.m, key)
		}
		c.mu.Unlock()
		return "", false, nil
	}
	return it.value, true, nil
}

func (c *Memory) Set(_ context.Context, key string, value string, ttl time.Duration) error {
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}
	c.mu.Lock()
	c.m[key] = memoryItem{value: value, exp: exp}
	c.mu.Unlock()
	return nil
}

func (c *Memory) Del(_ context.Context, keys ...string) (int, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := 0
	for _, k := range keys {
		if _, ok := c.m[k]; ok {
			delete(c.m, k)
			count++
		}
	}
	return count, nil
}

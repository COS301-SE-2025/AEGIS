package unit_tests

import (
	"aegis-api/cache"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestMemory_SetGet_NoTTL(t *testing.T) {
	c := cache.NewMemory()
	ctx := context.Background()

	err := c.Set(ctx, "foo", "bar", 0)
	require.NoError(t, err)

	got, ok, err := c.Get(ctx, "foo")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "bar", got)
}

func TestMemory_SetGet_WithTTL_Expires(t *testing.T) {
	c := cache.NewMemory()
	ctx := context.Background()

	require.NoError(t, c.Set(ctx, "temp", "v", 20*time.Millisecond))

	// Immediately present
	_, ok, err := c.Get(ctx, "temp")
	require.NoError(t, err)
	require.True(t, ok)

	time.Sleep(30 * time.Millisecond)

	// Should be expired and evicted on read
	_, ok, err = c.Get(ctx, "temp")
	require.NoError(t, err)
	require.False(t, ok)

	// Re-set to ensure cache still usable
	require.NoError(t, c.Set(ctx, "temp", "v2", 0))
	got, ok, err := c.Get(ctx, "temp")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "v2", got)
}

func TestMemory_Del(t *testing.T) {
	c := cache.NewMemory()
	ctx := context.Background()

	_ = c.Set(ctx, "a", "1", 0)
	_ = c.Set(ctx, "b", "2", 0)
	_ = c.Set(ctx, "c", "3", 0)

	n, err := c.Del(ctx, "a", "c", "z") // z doesn't exist
	require.NoError(t, err)
	require.Equal(t, 2, n)

	_, ok, _ := c.Get(ctx, "a")
	require.False(t, ok)
	_, ok, _ = c.Get(ctx, "c")
	require.False(t, ok)
	_, ok, _ = c.Get(ctx, "b")
	require.True(t, ok)
}

func TestMemory_DeleteByPrefixes_ThroughPublicInvalidators(t *testing.T) {
	m := cache.NewMemory()
	ctx := context.Background()

	// Seed a bunch of keys (t1 + t2 + unrelated)
	keys := []string{
		"cases:t1:active:q=abc",
		"cases:t1:closed:q=xyz",
		"cases:t1:all:q=123",
		"cases:t1:byUser:u1:q=111",
		"cases:t2:active:q=abc",
		"unrelated:key",
	}
	for _, k := range keys {
		_ = m.Set(ctx, k, "v", 0)
	}

	// Use the public helper which routes to memoryDelByPrefixes internally
	cache.InvalidateTenantLists(ctx, m, "t1")

	// Keys that should be gone
	_, ok, _ := m.Get(ctx, "cases:t1:active:q=abc")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "cases:t1:closed:q=xyz")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "cases:t1:all:q=123")
	require.False(t, ok)

	// Survivors
	_, ok, _ = m.Get(ctx, "cases:t1:byUser:u1:q=111") // not part of tenant-list invalidator
	require.True(t, ok)
	_, ok, _ = m.Get(ctx, "cases:t2:active:q=abc") // different tenant
	require.True(t, ok)
	_, ok, _ = m.Get(ctx, "unrelated:key")
	require.True(t, ok)
}

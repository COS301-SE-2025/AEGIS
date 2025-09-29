//go:build redis_integration
// +build redis_integration

package integration_test

import (
	"aegis-api/cache"
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRedis_BasicOps(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	pass := os.Getenv("REDIS_PASS")
	db := 0
	if s := os.Getenv("REDIS_DB"); s != "" {
		if v, err := strconv.Atoi(s); err == nil {
			db = v
		}
	}

	r := cache.NewRedis(addr, pass, db)
	ctx := context.Background()

	// Quick connectivity probe so we can skip gracefully if Redis isn't up.
	if err := r.Set(ctx, "it:probe", "ok", time.Second); err != nil {
		t.Skipf("Skipping Redis integration test; cannot connect to %s: %v", addr, err)
	}

	// Clean up keys used by test just in case
	_, _ = r.Del(ctx, "it:key", "it:ttl", "it:probe")

	// Set/Get
	require.NoError(t, r.Set(ctx, "it:key", "val", 0))
	val, ok, err := r.Get(ctx, "it:key")
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "val", val)

	// TTL
	require.NoError(t, r.Set(ctx, "it:ttl", "v", 50*time.Millisecond))
	_, ok, _ = r.Get(ctx, "it:ttl")
	require.True(t, ok)
	time.Sleep(60 * time.Millisecond)
	_, ok, _ = r.Get(ctx, "it:ttl")
	require.False(t, ok)

	// Del
	n, err := r.Del(ctx, "it:key", "it:ttl")
	require.NoError(t, err)
	require.Equal(t, 2, n)
}

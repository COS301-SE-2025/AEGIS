package unit_tests

import (
	"aegis-api/cache"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInvalidateTenantLists(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// matching (t1)
	_ = m.Set(ctx, "cases:t1:active:q=aaa", "v", 0)
	_ = m.Set(ctx, "cases:t1:closed:q=bbb", "v", 0)
	_ = m.Set(ctx, "cases:t1:all:q=ccc", "v", 0)

	// non-matching
	_ = m.Set(ctx, "cases:t2:active:q=zzz", "v", 0)
	_ = m.Set(ctx, "cases:t1:byUser:u1:q=kkk", "v", 0)

	cache.InvalidateTenantLists(ctx, m, "t1")

	_, ok, _ := m.Get(ctx, "cases:t1:active:q=aaa")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "cases:t1:closed:q=bbb")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "cases:t1:all:q=ccc")
	require.False(t, ok)

	_, ok, _ = m.Get(ctx, "cases:t2:active:q=zzz")
	require.True(t, ok)
	_, ok, _ = m.Get(ctx, "cases:t1:byUser:u1:q=kkk")
	require.True(t, ok)
}

func TestInvalidateByUserLists(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	_ = m.Set(ctx, "cases:t1:byUser:u1:q=kkk", "v", 0)
	_ = m.Set(ctx, "cases:t1:byUser:u2:q=lll", "v", 0)
	_ = m.Set(ctx, "cases:t1:active:q=aaa", "v", 0)

	cache.InvalidateByUserLists(ctx, m, "t1", "u1", "u2")

	_, ok, _ := m.Get(ctx, "cases:t1:byUser:u1:q=kkk")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "cases:t1:byUser:u2:q=lll")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "cases:t1:active:q=aaa")
	require.True(t, ok)
}

func TestInvalidateCaseSpecific(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()
	hk := cache.CaseHeaderKey("t1", "c9")
	ck := cache.CaseCollabsKey("t1", "c9")

	_ = m.Set(ctx, hk, "hv", 0)
	_ = m.Set(ctx, ck, "cv", 0)

	cache.InvalidateCaseHeader(ctx, m, "t1", "c9")
	cache.InvalidateCaseCollabs(ctx, m, "t1", "c9")

	_, ok, _ := m.Get(ctx, hk)
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, ck)
	require.False(t, ok)
}

func TestInvalidateDashboardTotals(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	_ = m.Set(ctx, "dashboard:t1:totals:user:u1:q=abc", "v", 0)
	_ = m.Set(ctx, "dashboard:t1:totals:user:u2:q=xyz", "v", 0)
	_ = m.Set(ctx, "dashboard:t2:totals:user:u1:q=abc", "v", 0)

	cache.InvalidateDashboardTotals(ctx, m, "t1", "u1", "u2")

	_, ok, _ := m.Get(ctx, "dashboard:t1:totals:user:u1:q=abc")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "dashboard:t1:totals:user:u2:q=xyz")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "dashboard:t2:totals:user:u1:q=abc")
	require.True(t, ok)
}

func TestInvalidateDashboardTotalsTenant(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	_ = m.Set(ctx, "dashboard:t1:totals:user:u1:q=abc", "v", 0)
	_ = m.Set(ctx, "dashboard:t1:totals:user:u2:q=xyz", "v", 0)
	_ = m.Set(ctx, "dashboard:t1:totals:tenant:q=all", "v", 0)
	_ = m.Set(ctx, "dashboard:t2:totals:user:u1:q=abc", "v", 0)

	cache.InvalidateDashboardTotalsTenant(ctx, m, "t1")

	_, ok, _ := m.Get(ctx, "dashboard:t1:totals:user:u1:q=abc")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "dashboard:t1:totals:user:u2:q=xyz")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "dashboard:t1:totals:tenant:q=all")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "dashboard:t2:totals:user:u1:q=abc")
	require.True(t, ok)
}

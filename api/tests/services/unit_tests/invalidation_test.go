package unit_tests

import (
	"aegis-api/cache"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockRedisClient is a mock for the Redis client interface
type MockRedisClient struct {
	mock.Mock
}

func (m *MockRedisClient) Scan(ctx context.Context, cursor uint64, match string, count int64) *redis.ScanCmd {
	args := m.Called(ctx, cursor, match, count)
	keys, _ := args.Get(0).([]string)
	nextCursor, _ := args.Get(1).(uint64)
	err, _ := args.Get(2).(error)
	cmd := redis.NewScanCmd(ctx, nil, "SCAN", cursor, "MATCH", match, "COUNT", count)
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(keys, nextCursor)
	}
	return cmd
}
func (m *MockRedisClient) Del(ctx context.Context, keys ...string) *redis.IntCmd {
	args := m.Called(ctx, keys)
	count, _ := args.Get(0).(int64)
	err, _ := args.Get(1).(error)
	cmd := redis.NewIntCmd(ctx, nil, "DEL", keys)
	if err != nil {
		cmd.SetErr(err)
	} else {
		cmd.SetVal(count)
	}
	return cmd
}
func TestInvalidateEvidenceCount(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Set up test data
	evidenceCountKey := "evidence:count:tenant:t1"
	_ = m.Set(ctx, evidenceCountKey, `{"count": 42}`, 0)
	_ = m.Set(ctx, "evidence:count:tenant:t2", `{"count": 99}`, 0) // Non-matching key
	_ = m.Set(ctx, "other:key", "value", 0)                        // Unrelated key

	cache.InvalidateEvidenceCount(ctx, m, "t1")

	// Verify that evidence count key for t1 is deleted
	_, ok, _ := m.Get(ctx, evidenceCountKey)
	require.False(t, ok, "evidence count key for t1 should be deleted")

	// Verify that non-matching keys remain
	_, ok, _ = m.Get(ctx, "evidence:count:tenant:t2")
	require.True(t, ok, "evidence count key for t2 should remain")
	_, ok, _ = m.Get(ctx, "other:key")
	require.True(t, ok, "unrelated key should remain")
}
func TestRefreshEvidenceCount(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Mock getCount function for success case
	getCountSuccess := func(tenantID string) (int, error) {
		if tenantID == "t1" {
			return 42, nil
		}
		return 0, errors.New("invalid tenant")
	}

	// Mock getCount function for error case
	getCountError := func(tenantID string) (int, error) {
		return 0, errors.New("failed to get count")
	}

	// Test case 1: Successful count refresh
	ttl := time.Second * 60
	cache.RefreshEvidenceCount(ctx, m, getCountSuccess, "t1", ttl)

	// Verify the cached value
	value, ok, _ := m.Get(ctx, "evidence:count:tenant:t1")
	require.True(t, ok, "evidence count key should exist after successful refresh")
	var payload struct {
		Count int `json:"count"`
	}
	err := json.Unmarshal([]byte(value), &payload)
	require.NoError(t, err, "should unmarshal cached value without error")
	require.Equal(t, 42, payload.Count, "cached count should match expected value")

	// Test case 2: Error case
	_ = m.Set(ctx, "evidence:count:tenant:t2", `{"count": 99}`, 0) // Pre-existing key
	cache.RefreshEvidenceCount(ctx, m, getCountError, "t2", ttl)

	// Verify that the key is deleted on error
	_, ok, _ = m.Get(ctx, "evidence:count:tenant:t2")
	require.False(t, ok, "evidence count key should be deleted on error")

	// Test case 3: Empty cache before refresh
	m = cache.NewMemory()
	cache.RefreshEvidenceCount(ctx, m, getCountSuccess, "t1", ttl)

	// Verify the cached value again
	value, ok, _ = m.Get(ctx, "evidence:count:tenant:t1")
	require.True(t, ok, "evidence count key should exist after refresh in empty cache")
	err = json.Unmarshal([]byte(value), &payload)
	require.NoError(t, err, "should unmarshal cached value without error")
	require.Equal(t, 42, payload.Count, "cached count should match expected value")
}
func TestInvalidationHelpers_Comprehensive(t *testing.T) {
	ctx := context.Background()

	t.Run("TestAllDeleteByPrefixesPaths", func(t *testing.T) {
		// Test Memory client path
		memoryClient := cache.NewMemory()

		// Set up multiple keys with different prefixes
		prefixes := []string{
			"cases:t1:active:q=",
			"cases:t1:closed:q=",
			"cases:t1:all:q=",
			"ev:list:t1:c1:q=",
		}

		for _, prefix := range prefixes {
			_ = memoryClient.Set(ctx, prefix+"hash1", "value1", 0)
			_ = memoryClient.Set(ctx, prefix+"hash2", "value2", 0)
		}

		// Call functions that use deleteByPrefixes
		cache.InvalidateTenantLists(ctx, memoryClient, "t1")
		cache.InvalidateEvidenceListsForCase(ctx, memoryClient, "t1", "c1")

		// Verify all are deleted
		for _, prefix := range prefixes {
			_, ok, _ := memoryClient.Get(ctx, prefix+"hash1")
			require.False(t, ok)
			_, ok, _ = memoryClient.Get(ctx, prefix+"hash2")
			require.False(t, ok)
		}
	})

	t.Run("TestEmptyAndEdgeCases", func(t *testing.T) {
		memoryClient := cache.NewMemory()

		// Test with empty cache - should not panic
		cache.InvalidateTenantLists(ctx, memoryClient, "nonexistent")
		cache.InvalidateByUserLists(ctx, memoryClient, "t1") // empty userIDs
		cache.InvalidateEvidenceListsForCase(ctx, memoryClient, "t1", "c1")

		// Test with single key
		_ = memoryClient.Set(ctx, "cases:t1:active:q=abc", "value", 0)
		cache.InvalidateTenantLists(ctx, memoryClient, "t1")
		_, ok, _ := memoryClient.Get(ctx, "cases:t1:active:q=abc")
		require.False(t, ok)
	})
}

func TestDashboardInvalidation_PrefixDeletion(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Set up dashboard keys that use prefix deletion
	_ = m.Set(ctx, "dashboard:t1:totals:user:u1:q=abc", "v1", 0)
	_ = m.Set(ctx, "dashboard:t1:totals:user:u1:q=def", "v2", 0)
	_ = m.Set(ctx, "dashboard:t1:totals:user:u2:q=xyz", "v3", 0)
	_ = m.Set(ctx, "dashboard:t1:totals:tenant:q=all", "v4", 0)

	// These use deleteByPrefixes internally
	cache.InvalidateDashboardTotals(ctx, m, "t1", "u1", "u2")
	cache.InvalidateDashboardTotalsTenant(ctx, m, "t1")

	// Verify all are deleted
	keys := []string{
		"dashboard:t1:totals:user:u1:q=abc",
		"dashboard:t1:totals:user:u1:q=def",
		"dashboard:t1:totals:user:u2:q=xyz",
		"dashboard:t1:totals:tenant:q=all",
	}

	for _, key := range keys {
		_, ok, _ := m.Get(ctx, key)
		require.False(t, ok, "key %s should be deleted", key)
	}
}
func TestInvalidateByUserLists(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Set up test data
	_ = m.Set(ctx, "cases:t1:byUser:u1:q=kkk", "v", 0)
	_ = m.Set(ctx, "cases:t1:byUser:u2:q=lll", "v", 0)
	_ = m.Set(ctx, "cases:t1:active:q=aaa", "v", 0) // Non-matching key

	cache.InvalidateByUserLists(ctx, m, "t1", "u1", "u2")

	// Verify matching keys are deleted
	_, ok, _ := m.Get(ctx, "cases:t1:byUser:u1:q=kkk")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "cases:t1:byUser:u2:q=lll")
	require.False(t, ok)
	// Verify non-matching key remains
	_, ok, _ = m.Get(ctx, "cases:t1:active:q=aaa")
	require.True(t, ok)
}
func TestInvalidateTenantLists(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Set up test data with various prefixes
	_ = m.Set(ctx, "cases:t1:active:q=abc123", "active1", 0)
	_ = m.Set(ctx, "cases:t1:active:q=def456", "active2", 0)
	_ = m.Set(ctx, "cases:t1:closed:q=xyz789", "closed1", 0)
	_ = m.Set(ctx, "cases:t1:all:q=aaa111", "all1", 0)
	_ = m.Set(ctx, "cases:t2:active:q=bbb222", "other_tenant", 0) // Should remain
	_ = m.Set(ctx, "other:key", "unrelated", 0)                   // Should remain

	// This should call deleteByPrefixes internally
	cache.InvalidateTenantLists(ctx, m, "t1")

	// Verify all matching keys are deleted
	_, ok, _ := m.Get(ctx, "cases:t1:active:q=abc123")
	require.False(t, ok, "active list key should be deleted")
	_, ok, _ = m.Get(ctx, "cases:t1:active:q=def456")
	require.False(t, ok, "active list key should be deleted")
	_, ok, _ = m.Get(ctx, "cases:t1:closed:q=xyz789")
	require.False(t, ok, "closed list key should be deleted")
	_, ok, _ = m.Get(ctx, "cases:t1:all:q=aaa111")
	require.False(t, ok, "all list key should be deleted")

	// Verify non-matching keys remain
	_, ok, _ = m.Get(ctx, "cases:t2:active:q=bbb222")
	require.True(t, ok, "other tenant's keys should remain")
	_, ok, _ = m.Get(ctx, "other:key")
	require.True(t, ok, "unrelated keys should remain")
}

func TestInvalidateEvidenceListsForCase(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Set up test data with evidence list prefixes
	_ = m.Set(ctx, "ev:list:t1:c1:q=abc123", "list1", 0)
	_ = m.Set(ctx, "ev:list:t1:c1:q=def456", "list2", 0)
	_ = m.Set(ctx, "ev:list:t1:c2:q=xyz789", "other_case", 0) // Should remain
	_ = m.Set(ctx, "ev:item:t1:e1", "item", 0)                // Should remain

	// This should call deleteByPrefixes internally
	cache.InvalidateEvidenceListsForCase(ctx, m, "t1", "c1")

	// Verify matching keys are deleted
	_, ok, _ := m.Get(ctx, "ev:list:t1:c1:q=abc123")
	require.False(t, ok, "evidence list key should be deleted")
	_, ok, _ = m.Get(ctx, "ev:list:t1:c1:q=def456")
	require.False(t, ok, "evidence list key should be deleted")

	// Verify non-matching keys remain
	_, ok, _ = m.Get(ctx, "ev:list:t1:c2:q=xyz789")
	require.True(t, ok, "other case's list keys should remain")
	_, ok, _ = m.Get(ctx, "ev:item:t1:e1")
	require.True(t, ok, "evidence item keys should remain")
}
func TestInvalidateCaseHeader(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Set up test data
	headerKey := cache.CaseHeaderKey("t1", "c9")
	_ = m.Set(ctx, headerKey, "hv", 0)
	_ = m.Set(ctx, cache.CaseCollabsKey("t1", "c9"), "cv", 0) // Non-matching key

	cache.InvalidateCaseHeader(ctx, m, "t1", "c9")

	// Verify header key is deleted
	_, ok, _ := m.Get(ctx, headerKey)
	require.False(t, ok)
	// Verify non-matching key remains
	_, ok, _ = m.Get(ctx, cache.CaseCollabsKey("t1", "c9"))
	require.True(t, ok)
}

func TestInvalidateCaseCollabs(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Set up test data
	collabsKey := cache.CaseCollabsKey("t1", "c9")
	_ = m.Set(ctx, collabsKey, "cv", 0)
	_ = m.Set(ctx, cache.CaseHeaderKey("t1", "c9"), "hv", 0) // Non-matching key

	cache.InvalidateCaseCollabs(ctx, m, "t1", "c9")

	// Verify collabs key is deleted
	_, ok, _ := m.Get(ctx, collabsKey)
	require.False(t, ok)
	// Verify non-matching key remains
	_, ok, _ = m.Get(ctx, cache.CaseHeaderKey("t1", "c9"))
	require.True(t, ok)
}

func TestInvalidateDashboardTotals(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Set up test data
	_ = m.Set(ctx, "dashboard:t1:totals:user:u1:q=abc", "v", 0)
	_ = m.Set(ctx, "dashboard:t1:totals:user:u2:q=xyz", "v", 0)
	_ = m.Set(ctx, "dashboard:t2:totals:user:u1:q=abc", "v", 0) // Non-matching key

	cache.InvalidateDashboardTotals(ctx, m, "t1", "u1", "u2")

	// Verify matching keys are deleted
	_, ok, _ := m.Get(ctx, "dashboard:t1:totals:user:u1:q=abc")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "dashboard:t1:totals:user:u2:q=xyz")
	require.False(t, ok)
	// Verify non-matching key remains
	_, ok, _ = m.Get(ctx, "dashboard:t2:totals:user:u1:q=abc")
	require.True(t, ok)
}

func TestInvalidateDashboardTotalsTenant(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Set up test data
	_ = m.Set(ctx, "dashboard:t1:totals:user:u1:q=abc", "v", 0)
	_ = m.Set(ctx, "dashboard:t1:totals:tenant:q=all", "v", 0)
	_ = m.Set(ctx, "dashboard:t2:totals:user:u1:q=abc", "v", 0) // Non-matching key

	cache.InvalidateDashboardTotalsTenant(ctx, m, "t1")

	// Verify matching keys are deleted
	_, ok, _ := m.Get(ctx, "dashboard:t1:totals:user:u1:q=abc")
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "dashboard:t1:totals:tenant:q=all")
	require.False(t, ok)
	// Verify non-matching key remains
	_, ok, _ = m.Get(ctx, "dashboard:t2:totals:user:u1:q=abc")
	require.True(t, ok)
}

func TestInvalidateEvidenceAll(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Test case 1: Normal case with keys present
	evidenceItemKey := cache.EvidenceItemKey("t1", "e1")
	evidenceTagsKey := cache.EvidenceTagsKey("t1", "e1")
	evidenceListKey := "ev:list:t1:c1:q=abc"
	_ = m.Set(ctx, evidenceItemKey, "item", 0)
	_ = m.Set(ctx, evidenceTagsKey, "tags", 0)
	_ = m.Set(ctx, evidenceListKey, "list", 0)
	_ = m.Set(ctx, "ev:list:t2:c2:q=xyz", "other", 0) // Non-matching key

	cache.InvalidateEvidenceAll(ctx, m, "t1", "c1", "e1")

	_, ok, _ := m.Get(ctx, evidenceItemKey)
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, evidenceTagsKey)
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, evidenceListKey)
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, "ev:list:t2:c2:q=xyz")
	require.True(t, ok)

	// Test case 2: Empty cache
	m = cache.NewMemory()
	cache.InvalidateEvidenceAll(ctx, m, "t1", "c1", "e1")
	_, ok, _ = m.Get(ctx, evidenceItemKey)
	require.False(t, ok) // Should not error or add keys
}

func TestInvalidateEvidenceItem(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Test case 1: Normal case with key present
	evidenceItemKey := cache.EvidenceItemKey("t1", "e1")
	_ = m.Set(ctx, evidenceItemKey, "item", 0)
	_ = m.Set(ctx, cache.EvidenceTagsKey("t1", "e1"), "tags", 0) // Non-matching key

	cache.InvalidateEvidenceItem(ctx, m, "t1", "e1")

	_, ok, _ := m.Get(ctx, evidenceItemKey)
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, cache.EvidenceTagsKey("t1", "e1"))
	require.True(t, ok)

	// Test case 2: Key does not exist
	m = cache.NewMemory()
	cache.InvalidateEvidenceItem(ctx, m, "t1", "e1")
	_, ok, _ = m.Get(ctx, evidenceItemKey)
	require.False(t, ok) // Should not error or add keys
}

func TestInvalidateEvidenceTags(t *testing.T) {
	ctx := context.Background()
	m := cache.NewMemory()

	// Test case 1: Normal case with key present
	evidenceTagsKey := cache.EvidenceTagsKey("t1", "e1")
	_ = m.Set(ctx, evidenceTagsKey, "tags", 0)
	_ = m.Set(ctx, cache.EvidenceItemKey("t1", "e1"), "item", 0) // Non-matching key

	cache.InvalidateEvidenceTags(ctx, m, "t1", "e1")

	_, ok, _ := m.Get(ctx, evidenceTagsKey)
	require.False(t, ok)
	_, ok, _ = m.Get(ctx, cache.EvidenceItemKey("t1", "e1"))
	require.True(t, ok)

	// Test case 2: Key does not exist
	m = cache.NewMemory()
	cache.InvalidateEvidenceTags(ctx, m, "t1", "e1")
	_, ok, _ = m.Get(ctx, evidenceTagsKey)
	require.False(t, ok) // Should not error or add keys
}

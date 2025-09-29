package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// On case create/update/status-change, blow tenant list caches
func InvalidateTenantLists(ctx context.Context, c Client, tenantID string) {
	prefixes := []string{
		"cases:" + tenantID + ":active:q=",
		"cases:" + tenantID + ":closed:q=",
		"cases:" + tenantID + ":all:q=",
	}
	deleteByPrefixes(ctx, c, prefixes...)
}

func InvalidateByUserLists(ctx context.Context, c Client, tenantID string, userIDs ...string) {
	var prefixes []string
	for _, u := range userIDs {
		prefixes = append(prefixes, "cases:"+tenantID+":byUser:"+u+":q=")
	}
	deleteByPrefixes(ctx, c, prefixes...)
}

func InvalidateCaseHeader(ctx context.Context, c Client, tenantID, caseID string) {
	c.Del(ctx, CaseHeaderKey(tenantID, caseID))
}

func InvalidateCaseCollabs(ctx context.Context, c Client, tenantID, caseID string) {
	c.Del(ctx, CaseCollabsKey(tenantID, caseID))
}

// Dashboard totals invalidation (user-scoped)
func InvalidateDashboardTotals(ctx context.Context, c Client, tenantID string, userIDs ...string) {
	var prefixes []string
	for _, u := range userIDs {
		prefixes = append(prefixes, "dashboard:"+tenantID+":totals:user:"+u+":q=")
	}
	deleteByPrefixes(ctx, c, prefixes...)
}

// If you have tenant-wide totals dashboards later:
func InvalidateDashboardTotalsTenant(ctx context.Context, c Client, tenantID string) {
	deleteByPrefixes(ctx, c, "dashboard:"+tenantID+":totals:")
}

// Blow item + tags + all list variants for a case (on add/update/delete/tag change)
func InvalidateEvidenceAll(ctx context.Context, c Client, tenantID, caseID, evidenceID string) {
	c.Del(ctx,
		EvidenceItemKey(tenantID, evidenceID),
		EvidenceTagsKey(tenantID, evidenceID),
	)
	InvalidateEvidenceListsForCase(ctx, c, tenantID, caseID)
}

func InvalidateEvidenceItem(ctx context.Context, c Client, tenantID, evidenceID string) {
	c.Del(ctx, EvidenceItemKey(tenantID, evidenceID))
}

func InvalidateEvidenceTags(ctx context.Context, c Client, tenantID, evidenceID string) {
	c.Del(ctx, EvidenceTagsKey(tenantID, evidenceID))
}

func InvalidateEvidenceListsForCase(ctx context.Context, c Client, tenantID, caseID string) {
	// delete all q variants
	prefix := fmt.Sprintf("ev:list:%s:%s:q=", tenantID, caseID)
	deleteByPrefixes(ctx, c, prefix) // uses your internal Redis/Memory prefix scanners
}

/* -------------------- Evidence Count -------------------- */

// Delete cached evidence count for a tenant
func InvalidateEvidenceCount(ctx context.Context, c Client, tenantID string) {
	c.Del(ctx, "evidence:count:tenant:"+tenantID)
}

// Recompute and cache evidence count using a caller-provided function
// getCount should be something like h.service.GetEvidenceCount
func RefreshEvidenceCount(
	ctx context.Context,
	c Client,
	getCount func(tenantID string) (int, error),
	tenantID string,
	ttl time.Duration,
) {
	n, err := getCount(tenantID)
	if err != nil {
		// Fall back to cold cache on error
		c.Del(ctx, "evidence:count:tenant:"+tenantID)
		return
	}
	payload := struct {
		Count int `json:"count"`
	}{Count: n}
	b, _ := json.Marshal(payload)
	_ = c.Set(ctx, "evidence:count:tenant:"+tenantID, string(b), ttl)
}

/* -------------------- helpers -------------------- */

func deleteByPrefixes(ctx context.Context, c Client, prefixes ...string) {
	switch cc := c.(type) {
	case *Redis:
		redisDelByPrefixes(ctx, cc, prefixes...)
	case *Memory:
		memoryDelByPrefixes(cc, prefixes...)
	default:
		// best-effort no-op for unknown clients
	}
}

func redisDelByPrefixes(ctx context.Context, r *Redis, prefixes ...string) {
	for _, p := range prefixes {
		var cursor uint64
		for {
			keys, next, err := r.rdb.Scan(ctx, cursor, p+"*", 1000).Result()
			if err != nil {
				break // avoid tight error loops; best-effort
			}
			cursor = next
			if len(keys) > 0 {
				_ = r.rdb.Del(ctx, keys...).Err()
			}
			if cursor == 0 {
				break
			}
		}
	}
}

func memoryDelByPrefixes(m *Memory, prefixes ...string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for k := range m.m {
		for _, p := range prefixes {
			if strings.HasPrefix(k, p) {
				delete(m.m, k)
				break
			}
		}
	}
}

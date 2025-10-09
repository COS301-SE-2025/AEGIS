package unit_tests

import (
	"aegis-api/cache"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestListKey(t *testing.T) {
	tenant := "t1"
	scope := cache.ScopeActive
	qSig := `{"Page":"1","PageSize":"20","Sort":"updatedAt","Order":"desc","Extra":{"status":["open"]}}`

	h := sha256.Sum256([]byte(qSig))
	wantHash := hex.EncodeToString(h[:])
	want := "cases:" + tenant + ":active:q=" + wantHash

	require.Equal(t, want, cache.ListKey(tenant, scope, qSig))
}

func TestListByUserKey(t *testing.T) {
	tenant := "t1"
	user := "u123"
	qSig := "xyz"

	h := sha256.Sum256([]byte(qSig))
	wantHash := hex.EncodeToString(h[:])
	want := "cases:" + tenant + ":byUser:" + user + ":q=" + wantHash

	require.Equal(t, want, cache.ListByUserKey(tenant, user, qSig))
}

func TestCaseHeaderKey(t *testing.T) {
	require.Equal(t, "case:t1:c9:header", cache.CaseHeaderKey("t1", "c9"))
}

func TestCaseCollabsKey(t *testing.T) {
	require.Equal(t, "case:t1:c9:collabs", cache.CaseCollabsKey("t1", "c9"))
}

func TestEvidenceListKey(t *testing.T) {
	tenant := "t1"
	caseID := "c123"
	qSig := `{"filters":{"type":["log","file"]},"sort":"name","page":1,"pageSize":20}`

	h := sha256.Sum256([]byte(qSig))
	wantHash := hex.EncodeToString(h[:])
	want := "ev:list:" + tenant + ":" + caseID + ":q=" + wantHash

	require.Equal(t, want, cache.EvidenceListKey(tenant, caseID, qSig))
}

func TestEvidenceItemKey(t *testing.T) {
	require.Equal(t, "ev:item:t1:e456", cache.EvidenceItemKey("t1", "e456"))
}

func TestEvidenceTagsKey(t *testing.T) {
	require.Equal(t, "ev:tags:t1:e456", cache.EvidenceTagsKey("t1", "e456"))
}

// Alternative if you can't export shaQSIG
func TestEvidenceListKey_IndirectlyTestsShaQSIG(t *testing.T) {
	tenant := "t1"
	caseID := "c123"
	qSig := "test query"

	// Calculate expected hash manually
	h := sha256.Sum256([]byte(qSig))
	expectedHash := hex.EncodeToString(h[:])
	expectedKey := "ev:list:" + tenant + ":" + caseID + ":q=" + expectedHash

	require.Equal(t, expectedKey, cache.EvidenceListKey(tenant, caseID, qSig))
}

// Test different query signatures produce different keys
func TestEvidenceListKey_QueryVariations(t *testing.T) {
	tenant := "t1"
	caseID := "c123"

	key1 := cache.EvidenceListKey(tenant, caseID, `{"page":1}`)
	key2 := cache.EvidenceListKey(tenant, caseID, `{"page":2}`)
	key3 := cache.EvidenceListKey(tenant, caseID, `{"page":1}`)

	// Same query should produce same key
	require.Equal(t, key1, key3)
	// Different queries should produce different keys
	require.NotEqual(t, key1, key2)
}

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

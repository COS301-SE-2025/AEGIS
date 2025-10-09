package unit_tests

import (
	"aegis-api/cache"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuildQuerySig_JSONShape(t *testing.T) {
	extra := map[string]any{
		"status": []string{"open", "pending"},
		"tags":   []string{"p1", "vip"},
	}
	sig := cache.BuildQuerySig("1", "25", "updatedAt", "desc", extra)
	require.NotEmpty(t, sig)

	var out struct {
		Page     string         `json:"Page"`
		PageSize string         `json:"PageSize"`
		Sort     string         `json:"Sort"`
		Order    string         `json:"Order"`
		Extra    map[string]any `json:"Extra"`
	}
	require.NoError(t, json.Unmarshal([]byte(sig), &out))
	require.Equal(t, "1", out.Page)
	require.Equal(t, "25", out.PageSize)
	require.Equal(t, "updatedAt", out.Sort)
	require.Equal(t, "desc", out.Order)
	require.Contains(t, out.Extra, "status")
	require.Contains(t, out.Extra, "tags")
}

func TestEvidenceQSIG(t *testing.T) {
	filters := map[string]any{
		"type":   []string{"log", "file"},
		"status": []string{"processed"},
	}

	sig := cache.EvidenceQSIG("1", "20", "name", "asc", filters)
	require.NotEmpty(t, sig)

	// Parse and verify the structure
	var out struct {
		Page     string         `json:"Page"`
		PageSize string         `json:"PageSize"`
		Sort     string         `json:"Sort"`
		Order    string         `json:"Order"`
		Extra    map[string]any `json:"Extra"`
	}
	require.NoError(t, json.Unmarshal([]byte(sig), &out))

	require.Equal(t, "1", out.Page)
	require.Equal(t, "20", out.PageSize)
	require.Equal(t, "name", out.Sort)
	require.Equal(t, "asc", out.Order)

	// Check that filters are properly nested under Extra
	require.Contains(t, out.Extra, "filters")
	filtersMap, ok := out.Extra["filters"].(map[string]any)
	require.True(t, ok, "filters should be a map")
	require.Contains(t, filtersMap, "type")
	require.Contains(t, filtersMap, "status")
}

func TestEvidenceQSIG_EmptyFilters(t *testing.T) {
	sig := cache.EvidenceQSIG("1", "20", "createdAt", "desc", nil)
	require.NotEmpty(t, sig)

	var out struct {
		Page     string         `json:"Page"`
		PageSize string         `json:"PageSize"`
		Sort     string         `json:"Sort"`
		Order    string         `json:"Order"`
		Extra    map[string]any `json:"Extra"`
	}
	require.NoError(t, json.Unmarshal([]byte(sig), &out))

	require.Equal(t, "1", out.Page)
	require.Equal(t, "20", out.PageSize)
	require.Equal(t, "createdAt", out.Sort)
	require.Equal(t, "desc", out.Order)
	require.Contains(t, out.Extra, "filters")
}

func TestBuildQuerySig_EmptyExtra(t *testing.T) {
	sig := cache.BuildQuerySig("1", "10", "name", "asc", nil)
	require.NotEmpty(t, sig)

	var out struct {
		Page     string         `json:"Page"`
		PageSize string         `json:"PageSize"`
		Sort     string         `json:"Sort"`
		Order    string         `json:"Order"`
		Extra    map[string]any `json:"Extra"`
	}
	require.NoError(t, json.Unmarshal([]byte(sig), &out))

	require.Equal(t, "1", out.Page)
	require.Equal(t, "10", out.PageSize)
	require.Equal(t, "name", out.Sort)
	require.Equal(t, "asc", out.Order)

	// When extra is nil, it should unmarshal as nil, not an empty map
	// So we should check that Extra exists but is nil
	require.Nil(t, out.Extra)
}

func TestBuildQuerySig_EmptyStrings(t *testing.T) {
	extra := map[string]any{"test": "value"}
	sig := cache.BuildQuerySig("", "", "", "", extra)
	require.NotEmpty(t, sig)

	var out struct {
		Page     string         `json:"Page"`
		PageSize string         `json:"PageSize"`
		Sort     string         `json:"Sort"`
		Order    string         `json:"Order"`
		Extra    map[string]any `json:"Extra"`
	}
	require.NoError(t, json.Unmarshal([]byte(sig), &out))

	require.Equal(t, "", out.Page)
	require.Equal(t, "", out.PageSize)
	require.Equal(t, "", out.Sort)
	require.Equal(t, "", out.Order)
	require.Equal(t, "value", out.Extra["test"])
}

func TestStableJSON(t *testing.T) {
	// Test that StableJSON produces consistent output
	data := map[string]any{
		"b": "second",
		"a": "first",
		"c": "third",
	}

	result1 := cache.StableJSON(data)
	result2 := cache.StableJSON(data)

	require.Equal(t, result1, result2)
	require.True(t, json.Valid(result1))

	// Parse back and verify content is preserved
	var parsed map[string]any
	require.NoError(t, json.Unmarshal(result1, &parsed))
	require.Equal(t, "first", parsed["a"])
	require.Equal(t, "second", parsed["b"])
	require.Equal(t, "third", parsed["c"])
}

// Test that different inputs produce different signatures
func TestBuildQuerySig_DifferentInputs(t *testing.T) {
	extra1 := map[string]any{"status": "open"}
	extra2 := map[string]any{"status": "closed"}

	sig1 := cache.BuildQuerySig("1", "10", "name", "asc", extra1)
	sig2 := cache.BuildQuerySig("1", "10", "name", "asc", extra2)
	sig3 := cache.BuildQuerySig("2", "10", "name", "asc", extra1)

	require.NotEqual(t, sig1, sig2)
	require.NotEqual(t, sig1, sig3)
	require.NotEqual(t, sig2, sig3)
}

// Test that same inputs produce identical signatures
func TestBuildQuerySig_Consistency(t *testing.T) {
	extra := map[string]any{"filters": map[string]any{"type": "file"}}

	sig1 := cache.BuildQuerySig("1", "20", "date", "desc", extra)
	sig2 := cache.BuildQuerySig("1", "20", "date", "desc", extra)

	require.Equal(t, sig1, sig2)
}

func TestEvidenceQSIG_Consistency(t *testing.T) {
	filters := map[string]any{"type": []string{"log"}}

	sig1 := cache.EvidenceQSIG("1", "10", "name", "asc", filters)
	sig2 := cache.EvidenceQSIG("1", "10", "name", "asc", filters)

	require.Equal(t, sig1, sig2)
}

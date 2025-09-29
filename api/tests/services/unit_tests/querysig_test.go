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

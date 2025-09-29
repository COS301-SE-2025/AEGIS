package unit_tests

import (
	"aegis-api/cache"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEntityETag(t *testing.T) {
	body := []byte("hello world")
	sum := sha256.Sum256(body)
	want := `W/"` + hex.EncodeToString(sum[:]) + `"`

	got := cache.EntityETag(body)
	require.Equal(t, want, got)
}

func TestListETag_EqualsEntityETag(t *testing.T) {
	data := []byte(`{"list":[1,2,3],"meta":{"page":1}}`)
	require.Equal(t, cache.EntityETag(data), cache.ListETag(data))
}

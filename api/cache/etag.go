package cache

import (
	"crypto/sha256"
	"encoding/hex"
)

// EntityETag: hash over stable serialization of a single case (updatedAt, id, title, status, etc.)
func EntityETag(stableBytes []byte) string {
	h := sha256.Sum256(stableBytes)
	return `W/"` + hex.EncodeToString(h[:]) + `"` // Weak ETag is fine for DB-backed resources
}

// ListETag: hash over list payload, including pagination meta
func ListETag(stableBytes []byte) string { return EntityETag(stableBytes) }

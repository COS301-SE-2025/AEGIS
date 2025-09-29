package middleware

import (
	"fmt"
	"net/http"
)

func SetCacheControl(w http.ResponseWriter, maxAgeSeconds int) {
	w.Header().Set("Cache-Control", "private, max-age="+itoa(maxAgeSeconds))
}

func itoa(i int) string { return fmt.Sprintf("%d", i) }

// IfNoneMatch handles ETag conditional requests. Return true if 304 was written.
func IfNoneMatch(w http.ResponseWriter, r *http.Request, etag string) bool {
	if inm := r.Header.Get("If-None-Match"); inm != "" && inm == etag {
		w.WriteHeader(http.StatusNotModified)
		return true
	}
	w.Header().Set("ETag", etag)
	return false
}

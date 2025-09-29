// api/cache/querysig.go
package cache

import "encoding/json"

func StableJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

// Build a signature for pagination/sorting + extras
func BuildQuerySig(page, pageSize, sort, order string, extra map[string]any) string {
	type sig struct {
		Page, PageSize, Sort, Order string
		Extra                       map[string]any
	}
	return string(StableJSON(sig{page, pageSize, sort, order, extra}))
}

func EvidenceQSIG(page, pageSize, sort, order string, filters map[string]any) string {
	payload := map[string]any{
		"page":     page,
		"pageSize": pageSize,
		"sort":     sort,
		"order":    order,
		"filters":  filters,
	}
	//_, _ := json.Marshal(payload)
	// reuse your existing signature builder
	return BuildQuerySig(page, pageSize, sort, order, map[string]any{"filters": payload["filters"]})
}

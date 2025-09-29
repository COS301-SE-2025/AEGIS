package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type ListScope string

const (
	ScopeActive ListScope = "active"
	ScopeClosed ListScope = "closed"
	ScopeAll    ListScope = "all"
)

func shaQuery(serialized string) string {
	h := sha256.Sum256([]byte(serialized))
	return hex.EncodeToString(h[:])
}

// qSig should capture filters + sort + page + pageSize in a stable serialized form (e.g., JSON with ordered keys)
func ListKey(tenantID string, scope ListScope, qSig string) string {
	return fmt.Sprintf("cases:%s:%s:q=%s", tenantID, scope, shaQuery(qSig))
}

func ListByUserKey(tenantID, userID, qSig string) string {
	return fmt.Sprintf("cases:%s:byUser:%s:q=%s", tenantID, userID, shaQuery(qSig))
}

func CaseHeaderKey(tenantID, caseID string) string {
	return fmt.Sprintf("case:%s:%s:header", tenantID, caseID)
}

func CaseCollabsKey(tenantID, caseID string) string {
	return fmt.Sprintf("case:%s:%s:collabs", tenantID, caseID)
}

// ev:list:<tenantId>:<caseId>:q=<sha>
func EvidenceListKey(tenantID, caseID, qsig string) string {
	return fmt.Sprintf("ev:list:%s:%s:q=%s", tenantID, caseID, shaQSIG(qsig))
}

// ev:item:<tenantId>:<evidenceId>
func EvidenceItemKey(tenantID, evidenceID string) string {
	return fmt.Sprintf("ev:item:%s:%s", tenantID, evidenceID)
}

// ev:tags:<tenantId>:<evidenceId>
func EvidenceTagsKey(tenantID, evidenceID string) string {
	return fmt.Sprintf("ev:tags:%s:%s", tenantID, evidenceID)
}

// If you want to reuse your BuildQuerySig output directly, we still hash it to keep keys compact.
func shaQSIG(s string) string {
	h := sha256.Sum256([]byte(s))
	return hex.EncodeToString(h[:])
}

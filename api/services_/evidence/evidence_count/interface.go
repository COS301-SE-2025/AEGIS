package evidencecount

// Provides the interface for fetching total evidence count
type EvidenceRepository interface {
	GetEvidenceCountByTenantID(tenantID string) (int64, error)
}

type EvidenceService interface {
	GetEvidenceCount(tenantID string) (int64, error)
}

package listArchiveCases

// ArchiveCaseLister defines the interface for listing archived cases
type ArchiveCaseLister interface {
	ListArchivedCases(userID, tenantID, teamID string) ([]ArchivedCase, error)
}

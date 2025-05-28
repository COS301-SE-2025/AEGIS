package case_status_update

type CaseStatusRepository interface {
	UpdateStatus(caseID string, newStatus string) error
}

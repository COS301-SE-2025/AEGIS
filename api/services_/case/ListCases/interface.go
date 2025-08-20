package ListCases

import (
	"aegis-api/services_/case/case_creation"
)

type CaseQueryRepository interface {
	QueryCases(filter CaseFilter) ([]Case, error)
	GetAllCases(tenantID string) ([]case_creation.Case, error)
	GetCasesByUser(userID string, tenantID string) ([]case_creation.Case, error)
	GetCaseByID(caseID string, tenantID string) (*case_creation.Case, error)
}

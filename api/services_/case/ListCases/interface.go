package ListCases

import (
	"aegis-api/services_/case/case_creation"
)

type CaseQueryRepository interface {
	QueryCases(filter CaseFilter) ([]Case, error)
	GetAllCases() ([]case_creation.Case, error)
	GetCasesByUser(userID string) ([]case_creation.Case, error)
}

package Case_Evidence_Totals


import (
    "aegis-api/repositories"
)

type DashboardService interface {
    GetCounts() (caseCount int64, evidenceCount int64, err error)
}

type dashboardService struct {
    statsRepo repositories.CountCasesEvidenceRepo
}

func NewDashboardService(statsRepo repositories.CountCasesEvidenceRepo) DashboardService {
    return &dashboardService{
        statsRepo: statsRepo,
    }
}

func (s *dashboardService) GetCounts() (int64, int64, error) {
    caseCount, err := s.statsRepo.CountCases()
    if err != nil {
        return 0, 0, err
    }

    evidenceCount, err := s.statsRepo.CountEvidence()
    if err != nil {
        return 0, 0, err
    }

    return caseCount, evidenceCount, nil
}

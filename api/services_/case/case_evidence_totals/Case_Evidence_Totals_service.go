package case_evidence_totals

type DashboardService interface {
    GetCounts(string,[]string) (caseCount int64, evidenceCount int64, err error)
}

type dashboardService struct {
    statsRepo CountCasesEvidenceRepo
}

func NewDashboardService(statsRepo CountCasesEvidenceRepo) DashboardService {
    return &dashboardService{
        statsRepo: statsRepo,
    }
}


func (s *dashboardService) GetCounts(userID string, statuses []string) (int64, int64, error) {
	caseCount, err := s.statsRepo.CountCases(userID, statuses)
	if err != nil {
		return 0, 0, err
	}

	evidenceCount, err := s.statsRepo.CountEvidence(userID)
	if err != nil {
		return 0, 0, err
	}

	return caseCount, evidenceCount, nil
}


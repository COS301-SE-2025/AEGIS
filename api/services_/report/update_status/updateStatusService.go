package update_status

import (
	"context"
	

	"github.com/google/uuid"
)

type ReportStatusService interface {
	UpdateStatus(ctx context.Context, reportID uuid.UUID, status ReportStatus) (*Report, error)
}

type ReportStatusServiceImpl struct {
	repo ReportStatusRepository
}

func NewReportStatusService(repo ReportStatusRepository) ReportStatusService {
	return &ReportStatusServiceImpl{repo: repo}
}

func (s *ReportStatusServiceImpl) UpdateStatus(ctx context.Context, reportID uuid.UUID, status ReportStatus) (*Report, error) {
	
	return s.repo.UpdateReportStatus(ctx, reportID, status)
}

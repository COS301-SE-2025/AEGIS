package evidence_tag

import (
	"context"

	"github.com/google/uuid"
	
)

type EvidenceTagService interface {
	TagEvidence(ctx context.Context, userID, evidenceID uuid.UUID, tags []string) error
	UntagEvidence(ctx context.Context, userID, evidenceID uuid.UUID, tags []string) error
	GetEvidenceTags(ctx context.Context, evidenceID uuid.UUID) ([]string, error)
}

type evidenceTagService struct {
	repo EvidenceTagRepository
}

func NewEvidenceTagService(repo EvidenceTagRepository) EvidenceTagService {
	return &evidenceTagService{repo: repo}
}

func (s *evidenceTagService) TagEvidence(ctx context.Context, userID, evidenceID uuid.UUID, tags []string) error {
	return s.repo.AddTagsToEvidence(ctx, userID, evidenceID, tags)
}

func (s *evidenceTagService) UntagEvidence(ctx context.Context, userID, evidenceID uuid.UUID, tags []string) error {
	return s.repo.RemoveTagsFromEvidence(ctx, userID, evidenceID, tags)
}

func (s *evidenceTagService) GetEvidenceTags(ctx context.Context, evidenceID uuid.UUID) ([]string, error) {
	return s.repo.GetTagsForEvidence(ctx, evidenceID)
}

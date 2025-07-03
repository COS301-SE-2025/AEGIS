package case_tags

import (
	"context"

	"github.com/google/uuid"
	
)

type CaseTagService interface {
	TagCase(ctx context.Context, userID uuid.UUID, caseID uuid.UUID, tags []string) error
	UntagCase(ctx context.Context, userID uuid.UUID, caseID uuid.UUID, tags []string) error
	GetTags(ctx context.Context, caseID uuid.UUID) ([]string, error)
}

type caseTagService struct {
	repo CaseTagRepository
}

func NewCaseTagService(repo CaseTagRepository) CaseTagService {
	return &caseTagService{repo: repo}
}

func (s *caseTagService) TagCase(ctx context.Context, userID uuid.UUID, caseID uuid.UUID, tags []string) error {
	// You could add permission checking here before delegating
	return s.repo.AddTagsToCase(ctx, userID, caseID, tags)
}

func (s *caseTagService) UntagCase(ctx context.Context, userID uuid.UUID, caseID uuid.UUID, tags []string) error {
	return s.repo.RemoveTagsFromCase(ctx, userID, caseID, tags)
}

func (s *caseTagService) GetTags(ctx context.Context, caseID uuid.UUID) ([]string, error) {
	return s.repo.GetTagsForCase(ctx, caseID)
}

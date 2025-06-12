package metadata

func NewMetadataService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) SaveEvidenceMetadata(e *Evidence, tags []string) error {
	if err := s.repo.SaveMetadata(e); err != nil {
		return err
	}
	return s.repo.AttachTags(e, tags)
}

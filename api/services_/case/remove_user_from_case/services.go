package remove_user_from_case

import "errors"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) RemoveUser(req RemoveUserRequest) error {
	isAdmin, err := s.repo.IsAdmin(req.AdminID)
	if err != nil {
		return err
	}
	if !isAdmin {
		return errors.New("unauthorized: only admins can remove users from a case")
	}
	return s.repo.RemoveUserFromCase(req.UserID, req.CaseID)
}

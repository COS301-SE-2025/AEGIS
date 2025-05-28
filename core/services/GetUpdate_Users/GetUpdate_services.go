package GetUpdate_Users

import (
	"errors"
	"aegis-api/models"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetProfile(userID string) (*models.UserDTO, error) {
	user, err := s.repo.GetUserByID(userID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (s *UserService) UpdateProfile(userID string, updates map[string]interface{}) error {
	return s.repo.UpdateUser(userID, updates)
}

func (s *UserService) Authenticate(email, password string) (*models.UserDTO, error) {
	user, err := s.repo.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid credentials")
	}
	// Add your password verification here (e.g., bcrypt.CompareHashAndPassword)
	return user, nil
}

func (s *UserService) GetUserRoles(userID string) ([]string, error) {
	return s.repo.GetUserRoles(userID)
}

func (s *UserService) GetRepo() UserRepository {
    return s.repo
}
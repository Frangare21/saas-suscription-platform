package service

import (
	"saas-subscription-platform/services/user-service/internal/model"
	"saas-subscription-platform/services/user-service/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{
		repo: repo,
	}
}

func (s *UserService) CreateUser(email, password string) (model.User, error) {
	return s.repo.Create(email, password)
}

func (s *UserService) GetUserByEmail(email string) (model.User, error) {
	return s.repo.GetByEmail(email)
}

func (s *UserService) GetUserByID(userID string) (model.User, error) {
	return s.repo.GetByID(userID)
}

func (s *UserService) UpdateUser(userID string, email *string, password *string) error {
	return s.repo.UpdateFields(userID, email, password)
}

func (s *UserService) DeleteUser(userID string) error {
	return s.repo.Delete(userID)
}

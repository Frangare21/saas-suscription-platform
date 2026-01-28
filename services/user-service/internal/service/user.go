package service

import (
	"saas-subscription-platform/services/user-service/internal/model"
)

// UserStore define las operaciones que la capa de servicio necesita del repositorio.
type UserStore interface {
	Create(email, password string) (model.User, error)
	GetByEmail(email string) (model.User, error)
	GetByID(userID string) (model.User, error)
	UpdateFields(userID string, email, password *string) error
	Delete(userID string) error
}

type UserService struct {
	repo UserStore
}

func NewUserService(repo UserStore) *UserService {
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

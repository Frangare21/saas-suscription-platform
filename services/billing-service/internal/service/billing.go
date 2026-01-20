package service

import (
	"saas-subscription-platform/services/billing-service/internal/model"
	"saas-subscription-platform/services/billing-service/internal/repository"
	"time"
)

type BillingService struct {
	repo *repository.InvoiceRepository
}

func NewBillingService(repo *repository.InvoiceRepository) *BillingService {
	return &BillingService{repo: repo}
}

func (s *BillingService) CreateInvoice(userID string, amount float64) (*model.Invoice, error) {
	invoice := &model.Invoice{
		UserID:    userID,
		Amount:    amount,
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.repo.CreateInvoice(invoice); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s *BillingService) GetInvoices() ([]*model.Invoice, error) {
	return s.repo.GetInvoices()
}

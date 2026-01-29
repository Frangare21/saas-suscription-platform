package service

import (
	"fmt"
	"time"

	"saas-subscription-platform/services/billing-service/internal/model"
	"saas-subscription-platform/services/billing-service/internal/repository"
)

// InvoiceStore define las operaciones que la capa de servicio necesita del repositorio.
type InvoiceStore interface {
	CreateInvoice(invoice *model.Invoice) error
	GetInvoiceByID(userID string, id int) (*model.Invoice, error)
	GetInvoices(filter repository.InvoiceFilter) ([]*model.Invoice, error)
}

type BillingService struct {
	repo InvoiceStore
}

func NewBillingService(repo InvoiceStore) *BillingService {
	return &BillingService{repo: repo}
}

func (s *BillingService) CreateInvoice(userID string, amountCents int64, currency string) (*model.Invoice, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	if amountCents <= 0 {
		return nil, fmt.Errorf("amount must be positive")
	}
	if currency == "" {
		currency = "USD"
	}

	now := time.Now()
	invoice := &model.Invoice{
		UserID:      userID,
		AmountCents: amountCents,
		Currency:    currency,
		Status:      "pending",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.CreateInvoice(invoice); err != nil {
		return nil, err
	}
	return invoice, nil
}

func (s *BillingService) GetInvoiceByID(userID string, id int) (*model.Invoice, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	return s.repo.GetInvoiceByID(userID, id)
}

func (s *BillingService) ListInvoices(userID, status string, limit, offset int) ([]*model.Invoice, error) {
	if userID == "" {
		return nil, fmt.Errorf("user id is required")
	}
	filter := repository.InvoiceFilter{
		UserID: userID,
		Status: status,
		Limit:  limit,
		Offset: offset,
	}
	return s.repo.GetInvoices(filter)
}

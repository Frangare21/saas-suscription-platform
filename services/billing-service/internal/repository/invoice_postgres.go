package repository

import (
	"database/sql"
	"fmt"
	"saas-subscription-platform/services/billing-service/internal/model"
)

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

func (r *InvoiceRepository) CreateInvoice(invoice *model.Invoice) error {
	query := `INSERT INTO invoices (user_id, amount, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5) RETURNING id`
	err := r.db.QueryRow(query, invoice.UserID, invoice.Amount, invoice.Status, invoice.CreatedAt, invoice.UpdatedAt).Scan(&invoice.ID)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	return nil
}

func (r *InvoiceRepository) GetInvoices() ([]*model.Invoice, error) {
	query := `SELECT id, user_id, amount, status, created_at, updated_at FROM invoices`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoices: %w", err)
	}
	defer rows.Close()

	var invoices []*model.Invoice
	for rows.Next() {
		invoice := &model.Invoice{}
		if err := rows.Scan(&invoice.ID, &invoice.UserID, &invoice.Amount, &invoice.Status, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

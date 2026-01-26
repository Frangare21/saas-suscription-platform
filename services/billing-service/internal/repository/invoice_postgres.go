package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"saas-subscription-platform/services/billing-service/internal/model"
)

type InvoiceRepository struct {
	db *sql.DB
}

func NewInvoiceRepository(db *sql.DB) *InvoiceRepository {
	return &InvoiceRepository{db: db}
}

// InvoiceFilter permite acotar listados por usuario, estado y paginación.
type InvoiceFilter struct {
	UserID string
	Status string
	Limit  int
	Offset int
}

func (r *InvoiceRepository) CreateInvoice(invoice *model.Invoice) error {
	query := `INSERT INTO invoices (user_id, amount_cents, currency, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`
	err := r.db.QueryRow(query, invoice.UserID, invoice.AmountCents, invoice.Currency, invoice.Status, invoice.CreatedAt, invoice.UpdatedAt).Scan(&invoice.ID)
	if err != nil {
		return fmt.Errorf("failed to create invoice: %w", err)
	}
	return nil
}

func (r *InvoiceRepository) GetInvoiceByID(userID string, id int) (*model.Invoice, error) {
	query := `SELECT id, user_id, amount_cents, currency, status, created_at, updated_at FROM invoices WHERE id = $1 AND user_id = $2`
	invoice := &model.Invoice{}
	if err := r.db.QueryRow(query, id, userID).Scan(&invoice.ID, &invoice.UserID, &invoice.AmountCents, &invoice.Currency, &invoice.Status, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch invoice: %w", err)
	}
	return invoice, nil
}

func (r *InvoiceRepository) GetInvoices(filter InvoiceFilter) ([]*model.Invoice, error) {
	base := `SELECT id, user_id, amount_cents, currency, status, created_at, updated_at FROM invoices`
	args := make([]interface{}, 0, 4)
	clauses := make([]string, 0, 2)

	if filter.UserID != "" {
		args = append(args, filter.UserID)
		clauses = append(clauses, fmt.Sprintf("user_id = $%d", len(args)))
	}
	if filter.Status != "" {
		args = append(args, filter.Status)
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}

	var sb strings.Builder
	sb.WriteString(base)
	if len(clauses) > 0 {
		sb.WriteString(" WHERE ")
		sb.WriteString(strings.Join(clauses, " AND "))
	}

	limit := filter.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	args = append(args, limit)
	sb.WriteString(fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", len(args)))
	args = append(args, offset)
	sb.WriteString(fmt.Sprintf(" OFFSET $%d", len(args)))

	rows, err := r.db.Query(sb.String(), args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch invoices: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var invoices []*model.Invoice
	for rows.Next() {
		invoice := &model.Invoice{}
		if err := rows.Scan(&invoice.ID, &invoice.UserID, &invoice.AmountCents, &invoice.Currency, &invoice.Status, &invoice.CreatedAt, &invoice.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan invoice: %w", err)
		}
		invoices = append(invoices, invoice)
	}
	return invoices, nil
}

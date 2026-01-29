package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"saas-subscription-platform/services/billing-service/internal/model"
	"saas-subscription-platform/services/billing-service/internal/repository"
	"saas-subscription-platform/services/billing-service/internal/service"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

type stubInvoiceStore struct {
	createFn func(inv *model.Invoice) error
	getByID  func(userID string, id int) (*model.Invoice, error)
	listFn   func(filter repository.InvoiceFilter) ([]*model.Invoice, error)
}

func (s stubInvoiceStore) CreateInvoice(inv *model.Invoice) error {
	if s.createFn == nil {
		return nil
	}
	return s.createFn(inv)
}

func (s stubInvoiceStore) GetInvoiceByID(userID string, id int) (*model.Invoice, error) {
	if s.getByID == nil {
		return nil, nil
	}
	return s.getByID(userID, id)
}

func (s stubInvoiceStore) GetInvoices(filter repository.InvoiceFilter) ([]*model.Invoice, error) {
	if s.listFn == nil {
		return nil, nil
	}
	return s.listFn(filter)
}

func newHandler(store service.InvoiceStore) *BillingHandler {
	svc := service.NewBillingService(store)
	return NewBillingHandler(svc)
}

func TestCreateInvoiceHandler_Success(t *testing.T) {
	h := newHandler(stubInvoiceStore{
		createFn: func(inv *model.Invoice) error {
			inv.ID = 1
			return nil
		},
	})

	body := bytes.NewBufferString(`{"amount_cents":1500,"currency":"USD"}`)
	req := httptest.NewRequest(http.MethodPost, "/invoices", body)
	req.Header.Set("X-Internal-User-ID", "user-1")
	rr := httptest.NewRecorder()

	h.CreateInvoice(rr, req)

	require.Equal(t, http.StatusCreated, rr.Code)
	var resp model.Invoice
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 1, resp.ID)
	require.Equal(t, "user-1", resp.UserID)
	require.Equal(t, int64(1500), resp.AmountCents)
}

func TestCreateInvoiceHandler_Invalid(t *testing.T) {
	h := newHandler(stubInvoiceStore{})

	req := httptest.NewRequest(http.MethodPost, "/invoices", bytes.NewBufferString("invalid"))
	req.Header.Set("X-Internal-User-ID", "user-1")
	rr := httptest.NewRecorder()

	h.CreateInvoice(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestCreateInvoiceHandler_MissingUser(t *testing.T) {
	h := newHandler(stubInvoiceStore{})
	req := httptest.NewRequest(http.MethodPost, "/invoices", bytes.NewBufferString(`{"amount_cents":100}`))
	rr := httptest.NewRecorder()

	h.CreateInvoice(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestCreateInvoiceHandler_RepoError(t *testing.T) {
	h := newHandler(stubInvoiceStore{
		createFn: func(inv *model.Invoice) error { return errors.New("db error") },
	})
	req := httptest.NewRequest(http.MethodPost, "/invoices", bytes.NewBufferString(`{"amount_cents":100}`))
	req.Header.Set("X-Internal-User-ID", "user-1")
	rr := httptest.NewRecorder()

	h.CreateInvoice(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetInvoicesHandler_Success(t *testing.T) {
	now := time.Now()
	h := newHandler(stubInvoiceStore{
		listFn: func(filter repository.InvoiceFilter) ([]*model.Invoice, error) {
			return []*model.Invoice{{ID: 1, UserID: filter.UserID, CreatedAt: now}}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/invoices?status=pending&limit=5&offset=0", nil)
	req.Header.Set("X-Internal-User-ID", "user-1")
	rr := httptest.NewRecorder()

	h.GetInvoices(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp []model.Invoice
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Len(t, resp, 1)
	require.Equal(t, 1, resp[0].ID)
}

func TestGetInvoicesHandler_Error(t *testing.T) {
	h := newHandler(stubInvoiceStore{
		listFn: func(filter repository.InvoiceFilter) ([]*model.Invoice, error) {
			return nil, errors.New("db error")
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/invoices", nil)
	req.Header.Set("X-Internal-User-ID", "user-1")
	rr := httptest.NewRecorder()

	h.GetInvoices(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetInvoiceByIDHandler_Success(t *testing.T) {
	h := newHandler(stubInvoiceStore{
		getByID: func(userID string, id int) (*model.Invoice, error) {
			return &model.Invoice{ID: id, UserID: userID}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/invoices/1", nil)
	req.Header.Set("X-Internal-User-ID", "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	h.GetInvoiceByID(rr, req)

	require.Equal(t, http.StatusOK, rr.Code)
	var resp model.Invoice
	require.NoError(t, json.NewDecoder(rr.Body).Decode(&resp))
	require.Equal(t, 1, resp.ID)
}

func TestGetInvoiceByIDHandler_InvalidID(t *testing.T) {
	h := newHandler(stubInvoiceStore{})

	req := httptest.NewRequest(http.MethodGet, "/invoices/abc", nil)
	req.Header.Set("X-Internal-User-ID", "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": "abc"})
	rr := httptest.NewRecorder()

	h.GetInvoiceByID(rr, req)

	require.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestGetInvoiceByIDHandler_NotFound(t *testing.T) {
	h := newHandler(stubInvoiceStore{
		getByID: func(userID string, id int) (*model.Invoice, error) { return nil, nil },
	})

	req := httptest.NewRequest(http.MethodGet, "/invoices/99", nil)
	req.Header.Set("X-Internal-User-ID", "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": "99"})
	rr := httptest.NewRecorder()

	h.GetInvoiceByID(rr, req)

	require.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetInvoiceByIDHandler_Error(t *testing.T) {
	h := newHandler(stubInvoiceStore{
		getByID: func(userID string, id int) (*model.Invoice, error) { return nil, errors.New("db error") },
	})

	req := httptest.NewRequest(http.MethodGet, "/invoices/1", nil)
	req.Header.Set("X-Internal-User-ID", "user-1")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	rr := httptest.NewRecorder()

	h.GetInvoiceByID(rr, req)

	require.Equal(t, http.StatusInternalServerError, rr.Code)
}

func TestGetInvoiceByIDHandler_MissingUser(t *testing.T) {
	h := newHandler(stubInvoiceStore{})

	req := httptest.NewRequest(http.MethodGet, "/invoices/1", nil)
	rr := httptest.NewRecorder()

	h.GetInvoiceByID(rr, req)

	require.Equal(t, http.StatusUnauthorized, rr.Code)
}

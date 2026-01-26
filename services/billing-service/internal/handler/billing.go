package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"saas-subscription-platform/services/billing-service/internal/service"
)

type BillingHandler struct {
	service *service.BillingService
}

func NewBillingHandler(service *service.BillingService) *BillingHandler {
	return &BillingHandler{service: service}
}

func (h *BillingHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-Internal-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("missing internal user id"))
		return
	}

	var req struct {
		AmountCents int64  `json:"amount_cents"`
		Currency    string `json:"currency"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request payload"))
		return
	}

	invoice, err := h.service.CreateInvoice(userID, req.AmountCents, req.Currency)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(err.Error()))
		return
	}

	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(invoice)
}

func (h *BillingHandler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-Internal-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("missing internal user id"))
		return
	}

	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	invoices, err := h.service.ListInvoices(userID, status, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to fetch invoices"))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(invoices)
}

func (h *BillingHandler) GetInvoiceByID(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-Internal-User-ID")
	if userID == "" {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("missing internal user id"))
		return
	}

	vars := mux.Vars(r)
	idStr := vars["id"]
	invoiceID, err := strconv.Atoi(idStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid invoice id"))
		return
	}

	invoice, err := h.service.GetInvoiceByID(userID, invoiceID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("failed to fetch invoice"))
		return
	}
	if invoice == nil {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("invoice not found"))
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(invoice)
}

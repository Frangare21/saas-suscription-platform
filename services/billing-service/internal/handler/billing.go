package handler

import (
	"encoding/json"
	"net/http"
	"saas-subscription-platform/services/billing-service/internal/service"
)

type BillingHandler struct {
	service *service.BillingService
}

func NewBillingHandler(service *service.BillingService) *BillingHandler {
	return &BillingHandler{service: service}
}

func (h *BillingHandler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string  `json:"user_id"`
		Amount float64 `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid request payload"))
		return
	}

	invoice, err := h.service.CreateInvoice(req.UserID, req.Amount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to create invoice"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(invoice)
}

func (h *BillingHandler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	invoices, err := h.service.GetInvoices()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to fetch invoices"))
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(invoices)
}

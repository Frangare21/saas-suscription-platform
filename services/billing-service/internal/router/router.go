package router

import (
	"database/sql"
	"net/http"

	"github.com/gorilla/mux"

	"saas-subscription-platform/services/billing-service/internal/handler"
	"saas-subscription-platform/services/billing-service/internal/middleware"
	"saas-subscription-platform/services/billing-service/internal/repository"
	"saas-subscription-platform/services/billing-service/internal/service"
)

// NewRouter construye el router HTTP del billing-service.
// La conexión a DB debe venir inyectada (configurada por env en server).
func NewRouter(db *sql.DB) *mux.Router {
	repo := repository.NewInvoiceRepository(db)
	billingService := service.NewBillingService(repo)
	h := handler.NewBillingHandler(billingService)

	r := mux.NewRouter()

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}).Methods(http.MethodGet)

	protected := r.NewRoute().Subrouter()
	protected.Use(middleware.InternalAuthMux)
	protected.HandleFunc("/invoices", h.CreateInvoice).Methods(http.MethodPost)
	protected.HandleFunc("/invoices", h.GetInvoices).Methods(http.MethodGet)

	return r
}

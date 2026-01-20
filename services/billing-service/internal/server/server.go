package server

import (
	"database/sql"
	"log"
	"net/http"

	_ "github.com/lib/pq"

	"saas-subscription-platform/services/billing-service/internal/config"
	"saas-subscription-platform/services/billing-service/internal/router"
)

func Run() error {
	cfg := config.Load()

	db, err := sql.Open("postgres", cfg.DBDSN)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	if err := db.Ping(); err != nil {
		return err
	}

	r := router.NewRouter(db)
	log.Printf("Starting billing-service on %s", cfg.HTTPAddr)
	return http.ListenAndServe(cfg.HTTPAddr, r)
}

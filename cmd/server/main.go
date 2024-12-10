package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/config"
	"github.com/Axel791/metricsalert/internal/server/handlers"
	"github.com/Axel791/metricsalert/internal/server/repositories"
	"github.com/Axel791/metricsalert/internal/shared/validatiors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg, err := config.ServerLoadConfig()
	if err != nil {
		fmt.Printf("error")
		//log.Fatalf("error loading config: %v", err)
	}

	addr := config.ParseFlags(cfg)

	if !validatiors.IsValidAddress(addr, false) {
		log.Fatalf("invalid address: %s\n", addr)
	}

	router := chi.NewRouter()

	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	storage := repositories.NewMetricRepository()

	router.Method(
		http.MethodPost, "/update/{metricType}/{name}/{value}", handlers.NewUpdateMetricHandler(storage),
	)
	router.Method(http.MethodGet, "/value/{metricType}/{name}", handlers.NewGetMetricHandler(storage))
	router.Method(http.MethodGet, "/", handlers.NewGetMetricsHTMLHandler(storage))

	err = http.ListenAndServe(addr, router)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}

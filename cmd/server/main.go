package main

import (
	"flag"
	"fmt"
	"github.com/Axel791/metricsalert/internal/server/config"
	"github.com/Axel791/metricsalert/internal/server/handlers"
	"github.com/Axel791/metricsalert/internal/server/storage/repositories"
	"github.com/Axel791/metricsalert/internal/shared/validatiors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
)

func parseFlags(cfg *config.Config) string {
	addr := flag.String("a", cfg.Address, "HTTP server address")
	flag.Parse()
	return *addr
}

func main() {
	cfg, err := config.ServerLoadConfig()
	if err != nil {
		fmt.Printf("error loading config: %v", err)
	}

	addr := parseFlags(cfg)

	if !validatiors.IsValidAddress(addr, false) {
		fmt.Printf("invalid address: %s\n", addr)
		return
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
		panic(err)
	}
}

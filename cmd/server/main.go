package main

import (
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/config"
	"github.com/Axel791/metricsalert/internal/server/handlers"
	"github.com/Axel791/metricsalert/internal/server/middleware"
	"github.com/Axel791/metricsalert/internal/server/repositories"
	"github.com/Axel791/metricsalert/internal/server/services"
	"github.com/Axel791/metricsalert/internal/shared/logger"
	"github.com/Axel791/metricsalert/internal/shared/validatiors"

	"github.com/go-chi/chi/v5"
	log "github.com/sirupsen/logrus"
)

func main() {
	logger.InitLogger()

	cfg, err := config.ServerLoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	addr := config.ParseFlags(cfg)

	if !validatiors.IsValidAddress(addr, false) {
		log.Fatalf("invalid address: %s\n", addr)
	}

	router := chi.NewRouter()

	router.Use(middleware.WithLogging)
	router.Use(middleware.GzipMiddleware)

	storage := repositories.NewMetricRepository()
	metricsService := services.NewMetricsService(storage)

	router.Method(
		http.MethodPost,
		"/update",
		handlers.NewUpdateMetricHandler(metricsService),
	)
	router.Method(
		http.MethodPost,
		"/value",
		handlers.NewGetMetricHandler(metricsService),
	)
	router.Method(
		http.MethodGet,
		"/",
		handlers.NewGetMetricsHTMLHandler(metricsService),
	)

	log.Infof("server started on %s", addr)
	err = http.ListenAndServe(addr, router)

	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}

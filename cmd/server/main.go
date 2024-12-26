package main

import (
	"github.com/Axel791/metricsalert/internal/server/handlers/deprecated"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/config"
	"github.com/Axel791/metricsalert/internal/server/handlers"
	serverMiddleware "github.com/Axel791/metricsalert/internal/server/middleware"
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

	router.Use(serverMiddleware.WithLogging)
	router.Use(middleware.StripSlashes)

	storage := repositories.NewMetricRepository()
	metricsService := services.NewMetricsService(storage)

	// Актуальные маршруты
	//router.Method(
	//	http.MethodPost,
	//	"/update/",
	//	handlers.NewUpdateMetricHandler(metricsService),
	//)
	router.Method(
		http.MethodPost,
		"/update",
		handlers.NewUpdateMetricHandler(metricsService),
	)
	//router.Method(
	//	http.MethodPost,
	//	"/value/",
	//	handlers.NewGetMetricHandler(metricsService),
	//)
	router.Method(
		http.MethodPost,
		"/value",
		handlers.NewGetMetricHandler(metricsService),
	)
	router.Get(
		"/healthcheck/",
		handlers.HealthCheckHandler,
	)
	router.Method(
		http.MethodGet,
		"/",
		handlers.NewGetMetricsHTMLHandler(metricsService),
	)

	// Устаревшие маршруты
	router.Method(
		http.MethodPost,
		"/update/{metricType}/{name}/{value}",
		deprecated.NewUpdateMetricHandler(storage),
	)
	router.Method(
		http.MethodGet,
		"/value/{metricType}/{name}",
		deprecated.NewGetMetricHandler(storage),
	)
	log.Infof("server started on %s", addr)
	err = http.ListenAndServe(addr, router)

	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}

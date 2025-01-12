package main

import (
	"context"
	"github.com/go-chi/chi/v5/middleware"
	"net/http"
	"time"

	"github.com/Axel791/metricsalert/internal/server/handlers/deprecated"

	"github.com/Axel791/metricsalert/internal/server/config"
	"github.com/Axel791/metricsalert/internal/server/handlers"
	serverMiddleware "github.com/Axel791/metricsalert/internal/server/middleware"
	"github.com/Axel791/metricsalert/internal/server/repositories"
	"github.com/Axel791/metricsalert/internal/server/services"
	"github.com/Axel791/metricsalert/internal/shared/validators"

	"github.com/go-chi/chi/v5"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(logrus.InfoLevel)

	cfg, err := config.ServerLoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	addr, storeIntervalFlag, filePathFlag, restoreFlag := config.ParseFlags(cfg)

	if !validators.IsValidAddress(addr, false) {
		log.Fatalf("invalid address: %s\n", addr)
	}

	router := chi.NewRouter()

	router.Use(serverMiddleware.WithLogging)
	router.Use(serverMiddleware.GzipMiddleware)
	router.Use(middleware.StripSlashes)

	opts := repositories.StoreOptions{
		FilePath:        filePathFlag,
		RestoreFromFile: restoreFlag,
		StoreInterval:   time.Duration(storeIntervalFlag) * time.Second,
		UseFileStore:    cfg.UseFileStorage,
	}

	storage, err := repositories.StoreFactory(context.Background(), opts)
	if err != nil {
		log.Fatalf("error creating storage: %v", err)
	}

	metricsService := services.NewMetricsService(storage)

	// Актуальные маршруты
	router.Method(
		http.MethodPost,
		"/update",
		handlers.NewUpdateMetricHandler(metricsService, log),
	)
	router.Method(
		http.MethodPost,
		"/value",
		handlers.NewGetMetricHandler(metricsService, log),
	)
	router.Get(
		"/healthcheck",
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

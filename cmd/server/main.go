package main

import (
	"context"
	"github.com/Axel791/metricsalert/internal/server/db"
	"github.com/go-chi/chi/v5"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/Axel791/metricsalert/internal/server/handlers/deprecated"

	"github.com/Axel791/metricsalert/internal/server/config"
	"github.com/Axel791/metricsalert/internal/server/handlers"
	serverMiddleware "github.com/Axel791/metricsalert/internal/server/middleware"
	"github.com/Axel791/metricsalert/internal/server/repositories"
	"github.com/Axel791/metricsalert/internal/server/services"
	"github.com/Axel791/metricsalert/internal/shared/validators"

	_ "github.com/lib/pq"
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

	addr, databaseDSN, storeIntervalFlag, filePathFlag, restoreFlag, key := config.ParseFlags(cfg)

	cfg.Address = addr
	cfg.DatabaseDSN = databaseDSN
	cfg.StoreInterval = storeIntervalFlag
	cfg.FileStoragePath = filePathFlag
	cfg.Restore = restoreFlag
	cfg.Key = key

	log.Infof("start server with key: %s", cfg.Key)

	if !validators.IsValidAddress(cfg.Address, false) {
		log.Fatalf("invalid address: %s\n", cfg.Address)
	}

	dbConn, err := db.ConnectDB(cfg.DatabaseDSN, cfg)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer func() {
		if dbConn != nil {
			_ = dbConn.Close()
		}
	}()

	router := chi.NewRouter()
	router.Use(serverMiddleware.WithLogging)
	router.Use(serverMiddleware.GzipMiddleware)
	router.Use(middleware.StripSlashes)

	opts := repositories.StoreOptions{
		FilePath:        cfg.FileStoragePath,
		RestoreFromFile: cfg.Restore,
		StoreInterval:   time.Duration(cfg.StoreInterval) * time.Second,
		UseFileStore:    cfg.UseFileStorage,
	}

	storage, err := repositories.StoreFactory(context.Background(), dbConn, opts)
	if err != nil {
		log.Fatalf("error creating storage: %v", err)
	}
	metricsService := services.NewMetricsService(storage)
	authService := services.NewAuthService(key)

	// Актуальные маршруты
	router.Method(
		http.MethodPost,
		"/update",
		handlers.NewUpdateMetricHandler(metricsService, authService, log),
	)
	router.Method(
		http.MethodPost,
		"/updates",
		handlers.NewUpdatesMetricsHandler(metricsService, authService, log),
	)
	router.Method(
		http.MethodPost,
		"/value",
		handlers.NewGetMetricHandler(metricsService, authService, log),
	)
	router.Get(
		"/healthcheck",
		handlers.NewHealthCheckHandler,
	)
	router.Method(
		http.MethodGet,
		"/",
		handlers.NewGetMetricsHTMLHandler(metricsService),
	)
	router.Method(
		http.MethodGet,
		"/ping",
		handlers.NewDatabaseHealthCheckHandler(cfg.DatabaseDSN),
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

	log.Infof("server started on %s", cfg.Address)
	err = http.ListenAndServe(cfg.Address, router)
	if err != nil {
		log.Fatalf("error starting server: %v", err)
	}
}

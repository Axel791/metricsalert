package main

import (
	"context"
	"net/http"
	"time"

	"github.com/Axel791/metricsalert/internal/shared"

	"github.com/go-chi/chi/v5"

	"github.com/Axel791/metricsalert/internal/server/db"

	"github.com/go-chi/chi/v5/middleware"

	"github.com/Axel791/metricsalert/internal/server/handlers/deprecated"

	"github.com/Axel791/metricsalert/internal/server/config"
	"github.com/Axel791/metricsalert/internal/server/handlers"
	serverMiddleware "github.com/Axel791/metricsalert/internal/server/middleware"
	"github.com/Axel791/metricsalert/internal/server/repositories"
	"github.com/Axel791/metricsalert/internal/server/services"
	"github.com/Axel791/metricsalert/internal/shared/validators"

	_ "net/http/pprof"

	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(logrus.InfoLevel)

	path := "server_config.json"
	shared.LoadEnvFromFile(log, path)

	ctx := shared.CatchShutdown()

	log.Infof("Build version: %s", buildVersion)
	log.Infof("Build date:    %s", buildDate)
	log.Infof("Build commit:  %s", buildCommit)

	cfg, err := config.ServerLoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v", err)
	}

	addr, databaseDSN, storeIntervalFlag, filePathFlag, restoreFlag, key, cryptoKey, trustedSubnet := config.ParseFlags(cfg)

	cfg.Address = addr
	cfg.DatabaseDSN = databaseDSN
	cfg.StoreInterval = storeIntervalFlag
	cfg.FileStoragePath = filePathFlag
	cfg.Restore = restoreFlag
	cfg.Key = key
	cfg.CryptoKey = cryptoKey
	cfg.TrustedSubnet = trustedSubnet

	if !validators.IsValidAddress(cfg.Address, false) {
		log.Fatalf("invalid address: %s", cfg.Address)
	}

	// --- подключаем БД --------------------------------------------------
	dbConn, err := db.ConnectDB(cfg.DatabaseDSN, cfg)
	if err != nil {
		log.Fatalf("error connecting to database: %v", err)
	}
	defer func() {
		if dbConn != nil {
			_ = dbConn.Close()
		}
	}()

	// --- создаём сервисы безопасности ----------------------------------
	var (
		cryptoSvc services.CryptoService // RSA-расшифровка
		signSvc   services.SignService   // HMAC-подписи
	)

	if cfg.CryptoKey != "" {
		cryptoSvc, err = services.NewCryptoService(cfg.CryptoKey)
		if err != nil {
			log.Fatalf("crypto key error: %v", err)
		}
		log.Info("RSA decryption enabled")
	}

	// HMAC-подпись используется, если RSA не включён
	if cfg.Key != "" && cryptoSvc == nil {
		signSvc = services.NewSignService(cfg.Key)
		log.Info("HMAC signature enabled")
	}

	// --- роутер и middleware -------------------------------------------
	router := chi.NewRouter()
	router.Use(serverMiddleware.WithLogging)

	if cryptoSvc != nil {
		router.Use(serverMiddleware.CryptoMiddleware(cryptoSvc))
	}
	if signSvc != nil {
		router.Use(serverMiddleware.SignatureMiddleware(signSvc))
	}
	if trustedSubnet != "" {
		router.Use(serverMiddleware.TrustedSubnetMiddleware(trustedSubnet))
	}

	router.Use(serverMiddleware.GzipMiddleware)
	router.Use(middleware.StripSlashes)

	// --- хранилище и сервис метрик -------------------------------------
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

	// --- актуальные маршруты -------------------------------------------
	router.Method(http.MethodPost, "/update",
		handlers.NewUpdateMetricHandler(metricsService, log))
	router.Method(http.MethodPost, "/updates",
		handlers.NewUpdatesMetricsHandler(metricsService, log))
	router.Method(http.MethodPost, "/value",
		handlers.NewGetMetricHandler(metricsService, log))
	router.Get("/healthcheck", handlers.NewHealthCheckHandler)
	router.Method(http.MethodGet, "/",
		handlers.NewGetMetricsHTMLHandler(metricsService))
	router.Method(http.MethodGet, "/ping",
		handlers.NewDatabaseHealthCheckHandler(cfg.DatabaseDSN))

	// --- устаревшие маршруты -------------------------------------------
	router.Method(http.MethodPost, "/update/{metricType}/{name}/{value}",
		deprecated.NewUpdateMetricHandler(storage))
	router.Method(http.MethodGet, "/value/{metricType}/{name}",
		deprecated.NewGetMetricHandler(storage))

	// --- pprof ----------------------------------------------------------
	go func() {
		if err = http.ListenAndServe("localhost:6060", nil); err != nil {
			log.Errorf("pprof server: %v", err)
		}
	}()

	// --- старт ----------------------------------------------------------
	log.Infof("server started on %s", cfg.Address)
	if err = http.ListenAndServe(cfg.Address, router); err != nil {
		log.Fatalf("error starting server: %v", err)
	}

	<-ctx.Done()
	_, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer cancel()

	log.Info("server stopped gracefully")
}

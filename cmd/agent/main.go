package main

import (
	"context"
	"crypto/rsa"
	"math/rand"
	"sync"
	"time"

	"github.com/Axel791/metricsalert/internal/shared"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/Axel791/metricsalert/internal/agent/sender"
	"github.com/Axel791/metricsalert/internal/agent/services"
	"github.com/Axel791/metricsalert/internal/shared/validators"

	cryptoutil "github.com/Axel791/metricsalert/internal/agent/crypto"
	"github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/agent/collector"
	"github.com/Axel791/metricsalert/internal/agent/config"
	"github.com/Axel791/metricsalert/internal/agent/model/api"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

// collectMetricsLoop собирает runtime-метрики.
func collectMetricsLoop(
	ctx context.Context,
	pollInterval time.Duration,
	mu *sync.RWMutex,
	metrics *api.Metrics,
	pollCount *int64,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			metric := collector.Collector()

			mu.Lock()
			*pollCount++
			randomValue := rand.Float64() * 100.0

			metrics.Alloc = float64(metric.Alloc) / 1024
			metrics.BuckHashSys = float64(metric.BuckHashSys) / 1024
			metrics.Frees = float64(metric.Frees)
			metrics.GCCPUFraction = metric.GCCPUFraction
			metrics.GCSys = float64(metric.GCSys) / 1024
			metrics.HeapAlloc = float64(metric.HeapAlloc) / 1024
			metrics.HeapIdle = float64(metric.HeapIdle) / 1024
			metrics.HeapInuse = float64(metric.HeapInuse) / 1024
			metrics.HeapObjects = float64(metric.HeapObjects)
			metrics.HeapReleased = float64(metric.HeapReleased) / 1024
			metrics.HeapSys = float64(metric.HeapSys) / 1024
			metrics.LastGC = float64(metric.LastGC)
			metrics.Lookups = float64(metric.Lookups)
			metrics.MCacheInuse = float64(metric.MCacheInuse) / 1024
			metrics.MSpanInuse = float64(metric.MSpanInuse) / 1024
			metrics.MSpanSys = float64(metric.MSpanSys) / 1024
			metrics.Mallocs = float64(metric.Mallocs)
			metrics.NextGC = float64(metric.NextGC) / 1024
			metrics.NumGC = float64(metric.NumGC)
			metrics.NumForcedGC = float64(metric.NumForcedGC)
			metrics.OtherSys = float64(metric.OtherSys) / 1024
			metrics.PauseTotalNs = float64(metric.PauseTotalNs)
			metrics.StackInuse = float64(metric.StackInuse) / 1024
			metrics.Sys = float64(metric.Sys) / 1024
			metrics.TotalAlloc = float64(metric.TotalAlloc) / 1024
			metrics.MCacheSys = float64(metric.MCacheSys) / 1024
			metrics.StackSys = float64(metric.StackSys) / 1024
			metrics.PollCount = *pollCount
			metrics.RandomValue = randomValue
			mu.Unlock()
		}
	}
}

// reportMetricsLoop делает снимок метрик и кладёт его в канал.
func reportMetricsLoop(
	ctx context.Context,
	reportInterval time.Duration,
	mu *sync.RWMutex,
	metrics *api.Metrics,
	sendCh chan<- api.Metrics,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			mu.RLock()
			currentMetrics := *metrics
			mu.RUnlock()

			sendCh <- currentMetrics
		}
	}
}

// startWorkerPool запускает воркеров, которые читают из sendCh.
func startWorkerPool(
	ctx context.Context,
	rateLimit int,
	sendCh <-chan api.Metrics,
	metricClient *sender.MetricClient,
	log *logrus.Logger,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	var innerWG sync.WaitGroup
	innerWG.Add(rateLimit)

	for i := 0; i < rateLimit; i++ {
		go func(workerID int) {
			defer innerWG.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case m, ok := <-sendCh:
					if !ok { // канал закрыт
						return
					}
					if err := metricClient.SendMetrics(m); err != nil {
						log.Errorf("Worker %d: error sending metrics: %v", workerID, err)
					}
				}
			}
		}(i)
	}

	innerWG.Wait()
}

// collectSystemMetricsLoop собирает системные метрики.
func collectSystemMetricsLoop(
	ctx context.Context,
	pollInterval time.Duration,
	mu *sync.RWMutex,
	metrics *api.Metrics,
	log *logrus.Logger,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			vmStat, err := mem.VirtualMemory()
			if err != nil {
				log.Errorf("error getting virtual memory info: %v", err)
				continue
			}
			cpuPercents, err := cpu.Percent(0, true)
			if err != nil {
				log.Errorf("error getting cpu percents: %v", err)
				continue
			}

			mu.Lock()
			metrics.TotalMemory = float64(vmStat.Total) / 1024
			metrics.FreeMemory = float64(vmStat.Free) / 1024
			metrics.CPUutilization1 = cpuPercents
			mu.Unlock()
		}
	}
}

// runAgent объединяет запуск сборщиков метрик, worker pool и т.д.
func runAgent(
	address string,
	reportInterval, pollInterval time.Duration,
	log *logrus.Logger,
	cryptoKey, key string,
	rateLimit int,
	useGrpc bool,
) {
	if !validators.IsValidAddress(address, true) {
		log.Fatalf("invalid address: %s\n", address)
	}

	authService := services.NewAuthServiceHandler(key)
	var rsaPub *rsa.PublicKey
	if cryptoKey != "" {
		var err error
		rsaPub, err = cryptoutil.LoadPublic(cryptoKey)
		if err != nil {
			log.Fatalf("RSA key error: %v", err)
		}
		log.Info("RSA encryption enabled")
	}

	metricClient := sender.NewMetricClient(address, log, authService, rsaPub, useGrpc)

	sendCh := make(chan api.Metrics, rateLimit)

	var (
		mu         sync.RWMutex
		metricsDTO api.Metrics
		pollCount  int64
	)

	// ------------- ДОБАВЛЕНО -----------------------------------
	ctx := shared.CatchShutdown() // слушаем SIGINT/SIGTERM/SIGQUIT
	wg := &sync.WaitGroup{}       // ждём завершения всех воркеров
	// -----------------------------------------------------------

	// Запускаем горутину по сбору runtime-метрик.
	wg.Add(1)
	go collectMetricsLoop(ctx, pollInterval, &mu, &metricsDTO, &pollCount, wg)

	// Запускаем горутину по сбору системных метрик.
	wg.Add(1)
	go collectSystemMetricsLoop(ctx, pollInterval, &mu, &metricsDTO, log, wg)

	// Запускаем горутину формирования отчётов.
	wg.Add(1)
	go reportMetricsLoop(ctx, reportInterval, &mu, &metricsDTO, sendCh, wg)

	// Запускаем worker pool для отправки метрик.
	wg.Add(1)
	go startWorkerPool(ctx, rateLimit, sendCh, metricClient, log, wg)

	// Ждём отмены контекста (первый сигнал).
	<-ctx.Done()

	close(sendCh)

	// Дожидаемся корректного завершения всех горутин.
	wg.Wait()

	log.Info("agent stopped gracefully")
}

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(logrus.InfoLevel)

	log.Infof("Build version: %s", buildVersion)
	log.Infof("Build date:    %s", buildDate)
	log.Infof("Build commit:  %s", buildCommit)

	path := "agent_config.json"
	shared.LoadEnvFromFile(log, path)

	cfg, err := config.AgentLoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v\n", err)
	}
	address, reportInterval, pollInterval, cryptoKey, key, rateLimit := config.ParseFlags(cfg)

	cfg.CryptoKey = cryptoKey
	cfg.Key = key
	cfg.RateLimit = rateLimit

	runAgent(address, reportInterval, pollInterval, log, cryptoKey, key, rateLimit, cfg.UseGRPC)
}

package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"

	"github.com/Axel791/metricsalert/internal/agent/sender"
	"github.com/Axel791/metricsalert/internal/agent/services"
	"github.com/Axel791/metricsalert/internal/shared/validators"

	"github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/agent/collector"
	"github.com/Axel791/metricsalert/internal/agent/config"
	"github.com/Axel791/metricsalert/internal/agent/model/api"
)

// collectMetricsLoop собирает runtime-метрики и обновляет только соответствующие поля.
func collectMetricsLoop(pollInterval time.Duration, mu *sync.RWMutex, metrics *api.Metrics, pollCount *int64) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
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

// reportMetricsLoop делает снимок текущих метрик и отправляет его в канал задач.
func reportMetricsLoop(reportInterval time.Duration, mu *sync.RWMutex, metrics *api.Metrics, sendCh chan<- api.Metrics) {
	ticker := time.NewTicker(reportInterval)
	defer ticker.Stop()

	for range ticker.C {
		mu.RLock()
		currentMetrics := *metrics
		mu.RUnlock()

		sendCh <- currentMetrics
	}
}

// startWorkerPool запускает пул воркеров, каждый из которых забирать задачу из sendCh и отправляет метрики.
func startWorkerPool(rateLimit int, sendCh <-chan api.Metrics, metricClient *sender.MetricClient, log *logrus.Logger) {
	for i := 0; i < rateLimit; i++ {
		go func(workerID int) {
			for m := range sendCh {
				if err := metricClient.SendMetrics(m); err != nil {
					log.Errorf("Worker %d: error sending metrics: %v", workerID, err)
				}
			}
		}(i)
	}
}

// collectSystemMetricsLoop собирает дополнительные системные метрики с помощью gopsutil.
func collectSystemMetricsLoop(pollInterval time.Duration, mu *sync.RWMutex, metrics *api.Metrics, log *logrus.Logger) {
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
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

// runAgent объединяет запуск сборщиков метрик, worker pool и т.д.
func runAgent(address string, reportInterval, pollInterval time.Duration, log *logrus.Logger, key string, rateLimit int) {
	if !validators.IsValidAddress(address, true) {
		log.Fatalf("invalid address: %s\n", address)
	}

	authService := services.NewAuthServiceHandler(key)
	metricClient := sender.NewMetricClient(address, log, authService)

	sendCh := make(chan api.Metrics, rateLimit)

	var (
		mu         sync.RWMutex
		metricsDTO api.Metrics
		pollCount  int64
	)

	// Запускаем горутину по сбору runtime-метрик.
	go collectMetricsLoop(pollInterval, &mu, &metricsDTO, &pollCount)
	// Запускаем горутину по сбору системных метрик (TotalMemory, FreeMemory, CPUutilization1).
	go collectSystemMetricsLoop(pollInterval, &mu, &metricsDTO, log)
	// Запускаем горутину формирования отчётов.
	go reportMetricsLoop(reportInterval, &mu, &metricsDTO, sendCh)
	// Запускаем worker pool для отправки метрик.
	startWorkerPool(rateLimit, sendCh, metricClient, log)

	select {}
}

func main() {
	log := logrus.New()
	log.SetFormatter(&logrus.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})
	log.SetLevel(logrus.InfoLevel)

	log.Infof("agent started")

	cfg, err := config.AgentLoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v\n", err)
	}
	address, reportInterval, pollInterval, key, rateLimit := config.ParseFlags(cfg)

	cfg.Key = key
	cfg.RateLimit = rateLimit

	runAgent(address, reportInterval, pollInterval, log, key, rateLimit)
}

package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/agent/collector"
	"github.com/Axel791/metricsalert/internal/agent/config"
	"github.com/Axel791/metricsalert/internal/agent/model/api"
	"github.com/Axel791/metricsalert/internal/agent/sender"
	"github.com/Axel791/metricsalert/internal/shared/validators"
)

func runAgent(address string, reportInterval, pollInterval time.Duration, log *logrus.Logger) {
	if !validators.IsValidAddress(address, true) {
		log.Fatalf("invalid address: %s\n", address)
	}

	tickerCollector := time.NewTicker(pollInterval)
	tickerSender := time.NewTicker(reportInterval)

	metricClient := sender.NewMetricClient(address, log)

	defer tickerCollector.Stop()
	defer tickerSender.Stop()

	var metricsDTO api.Metrics
	var pollCount int64
	var mu sync.Mutex

	for {
		select {
		case <-tickerCollector.C:
			mu.Lock()
			pollCount++
			randomValue := rand.Float64() * 100.0

			metric := collector.Collector()

			metricsDTO = api.Metrics{
				Alloc:         float64(metric.Alloc) / 1024,
				BuckHashSys:   float64(metric.BuckHashSys) / 1024,
				Frees:         float64(metric.Frees),
				GCCPUFraction: metric.GCCPUFraction,
				GCSys:         float64(metric.GCSys) / 1024,
				HeapAlloc:     float64(metric.HeapAlloc) / 1024,
				HeapIdle:      float64(metric.HeapIdle) / 1024,
				HeapInuse:     float64(metric.HeapInuse) / 1024,
				HeapObjects:   float64(metric.HeapObjects),
				HeapReleased:  float64(metric.HeapReleased) / 1024,
				HeapSys:       float64(metric.HeapSys) / 1024,
				LastGC:        float64(metric.LastGC),
				Lookups:       float64(metric.Lookups),
				MCacheInuse:   float64(metric.MCacheInuse) / 1024,
				MSpanInuse:    float64(metric.MSpanInuse) / 1024,
				MSpanSys:      float64(metric.MSpanSys) / 1024,
				Mallocs:       float64(metric.Mallocs),
				NextGC:        float64(metric.NextGC) / 1024,
				NumGC:         float64(metric.NumGC),
				NumForcedGC:   float64(metric.NumForcedGC),
				OtherSys:      float64(metric.OtherSys) / 1024,
				PauseTotalNs:  float64(metric.PauseTotalNs),
				StackInuse:    float64(metric.StackInuse) / 1024,
				Sys:           float64(metric.Sys) / 1024,
				TotalAlloc:    float64(metric.TotalAlloc) / 1024,
				MCacheSys:     float64(metric.MCacheSys) / 1024,
				StackSys:      float64(metric.StackSys) / 1024,
				PollCount:     pollCount,
				RandomValue:   randomValue,
			}
			mu.Unlock()

		case <-tickerSender.C:
			mu.Lock()
			err := metricClient.SendMetrics(metricsDTO)
			mu.Unlock()
			if err != nil {
				log.Errorf("error sending metrics: %v\n", err)
			}
		}
	}
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

	address, reportInterval, pollInterval := config.ParseFlags(cfg)

	runAgent(address, reportInterval, pollInterval, log)
}

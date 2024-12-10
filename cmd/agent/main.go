package main

import (
	"log"
	"time"

	"github.com/Axel791/metricsalert/internal/agent/collector"
	"github.com/Axel791/metricsalert/internal/agent/config"
	"github.com/Axel791/metricsalert/internal/agent/model/dto"
	"github.com/Axel791/metricsalert/internal/agent/sender"
	"github.com/Axel791/metricsalert/internal/shared/validatiors"
)

func runAgent(address string, reportInterval, pollInterval time.Duration) {
	if !validatiors.IsValidAddress(address, true) {
		log.Printf("invalid address: %s\n", address)
		return
	}

	tickerCollector := time.NewTicker(pollInterval)
	tickerSender := time.NewTicker(reportInterval)

	metricClient := sender.NewMetricClient(address)

	defer tickerCollector.Stop()
	defer tickerSender.Stop()

	var metricsDTO dto.Metrics

	for {
		select {
		case <-tickerCollector.C:
			metric := collector.Collector()

			metricsDTO = dto.Metrics{
				Alloc:         float64(metric.Alloc) / 1024,
				BuckHashSys:   float64(metric.BuckHashSys) / 1024,
				Frees:         float64(metric.Frees),
				GCCPUFraction: metric.GCCPUFraction,
			}

		case <-tickerSender.C:
			err := metricClient.SendMetrics(metricsDTO)
			if err != nil {
				log.Printf("error sending metrics: %v\n", err)
			}
		}
	}
}

func main() {
	cfg, err := config.AgentLoadConfig()
	if err != nil {
		log.Fatalf("error loading config: %v\n", err)
	}

	address, reportInterval, pollInterval := config.ParseFlags(cfg)

	runAgent(address, reportInterval, pollInterval)
}

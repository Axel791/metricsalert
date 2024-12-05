package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

	"github.com/Axel791/metricsalert/internal/agent/collector"
	"github.com/Axel791/metricsalert/internal/agent/config"
	"github.com/Axel791/metricsalert/internal/agent/model/dto"
	"github.com/Axel791/metricsalert/internal/agent/sender"
	"github.com/Axel791/metricsalert/internal/shared/validatiors"
)

func parseFlags(cfg *config.Config) (string, time.Duration, time.Duration) {
	address := flag.String("a", cfg.Address, "HTTP server address")
	reportInterval := flag.Int(
		"r",
		int(cfg.ReportInterval),
		"Frequency of sending metrics to the server (in seconds)",
	)
	pollInterval := flag.Int(
		"p",
		int(cfg.PollInterval),
		"Frequency of collecting metrics from runtime (in seconds)",
	)

	flag.Parse()

	addr := *address
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}

	return addr, time.Duration(*reportInterval) * time.Second, time.Duration(*pollInterval) * time.Second
}

func runAgent(address string, reportInterval, pollInterval time.Duration) {
	if !validatiors.IsValidAddress(address, true) {
		fmt.Printf("invalid address: %s\n", address)
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
				fmt.Printf("error sending metrics: %v\n", err)
			}
		}
	}
}

func main() {
	cfg, err := config.AgentLoadConfig()
	if err != nil {
		fmt.Printf("error loading config: %v\n", err)
		return
	}

	address, reportInterval, pollInterval := parseFlags(cfg)

	runAgent(address, reportInterval, pollInterval)
}

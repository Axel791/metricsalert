package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"github.com/Axel791/metricsalert/internal/agent/model/api"
	"github.com/gojek/heimdall/v7/httpclient"
	log "github.com/sirupsen/logrus"
	"io"
	"math"
	"net/http"
	"net/url"
	"time"
)

const (
	maxRetries  = 5
	minInterval = 1 * time.Second
	maxInterval = 5 * time.Second
)

type MetricClient struct {
	httpClient *httpclient.Client
	baseURL    string
}

func NewMetricClient(baseURL string) *MetricClient {
	client := httpclient.NewClient()
	return &MetricClient{
		httpClient: client,
		baseURL:    baseURL,
	}
}

func (client *MetricClient) SendMetrics(metrics api.Metrics) error {
	metricsList := []api.MetricPost{
		{ID: "Alloc", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.Alloc)},
		{ID: "BuckHashSys", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.BuckHashSys)},
		{ID: "Frees", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.Frees)},
		{ID: "GCCPUFraction", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.GCCPUFraction)},
		{ID: "GCSys", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.GCSys)},
		{ID: "HeapAlloc", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.HeapAlloc)},
		{ID: "HeapIdle", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.HeapIdle)},
		{ID: "HeapInuse", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.HeapInuse)},
		{ID: "HeapObjects", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.HeapObjects)},
		{ID: "HeapReleased", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.HeapReleased)},
		{ID: "HeapSys", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.HeapSys)},
		{ID: "LastGC", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.LastGC)},
		{ID: "Lookups", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.Lookups)},
		{ID: "MCacheInuse", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.MCacheInuse)},
		{ID: "MSpanInuse", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.MSpanInuse)},
		{ID: "MSpanSys", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.MSpanSys)},
		{ID: "Mallocs", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.Mallocs)},
		{ID: "NextGC", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.NextGC)},
		{ID: "NumGC", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.NumGC)},
		{ID: "NumForcedGC", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.NumForcedGC)},
		{ID: "OtherSys", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.OtherSys)},
		{ID: "PauseTotalNs", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.PauseTotalNs)},
		{ID: "StackInuse", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.StackInuse)},
		{ID: "Sys", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.Sys)},
		{ID: "TotalAlloc", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.TotalAlloc)},
		{ID: "PollCount", MType: "counter", Delta: metrics.PollCount},
		{ID: "RandomValue", MType: "gauge", Value: roundToSixDecimalPlaces(metrics.RandomValue)},
	}

	if err := client.healthCheck(); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	for _, metric := range metricsList {
		log.Infof(
			"Sending metric: %s %s %v %d", metric.ID, metric.MType, metric.Value, metric.Delta,
		)
		err := client.sendMetric(metric)
		if err != nil {
			log.Errorf("failed to send metric %s: %v", metric.ID, err)
		}
	}

	return nil
}

func (client *MetricClient) sendMetric(metric api.MetricPost) error {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Accept-Encoding", "gzip")

	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	u, err := url.Parse(fmt.Sprintf("%s/update", client.baseURL))
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, u.String(), bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header = headers

	rsp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send metrics %s: %w", metric.ID, err)
	}
	defer rsp.Body.Close()

	var responseBody []byte

	if rsp.Header.Get("Content-Encoding") == "gzip" {
		gzipReader, err := gzip.NewReader(rsp.Body)
		if err != nil {
			return fmt.Errorf("failed to create gzip reader: %w", err)
		}
		defer gzipReader.Close()

		responseBody, err = io.ReadAll(gzipReader)
		if err != nil {
			return fmt.Errorf("failed to read gzip response: %w", err)
		}
	} else {
		responseBody, err = io.ReadAll(rsp.Body)
		if err != nil {
			return fmt.Errorf("failed to read response body: %w", err)
		}
	}

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d, response: %s", rsp.StatusCode, string(responseBody))
	}

	return nil
}

func (client *MetricClient) healthCheck() error {
	u, err := url.Parse(fmt.Sprintf("%s/healthcheck/", client.baseURL))
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	retries := 0
	interval := minInterval

	for retries < maxRetries {
		rsp, err := client.httpClient.Do(req)
		if err != nil {
			log.Infof("failed to send healthcheck request (attempt %d/%d): %v", retries+1, maxRetries, err)
		} else {
			defer rsp.Body.Close()

			if rsp.StatusCode == http.StatusOK {
				return nil
			} else {
				log.Infof("unexpected status code during health check: %d (attempt %d/%d)", rsp.StatusCode, retries+1, maxRetries)
			}
		}

		retries++

		if interval < maxInterval {
			interval += time.Duration(1 / 2)
			if interval > maxInterval {
				interval = maxInterval
			}
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("health check failed after %d attempts", retries)
}

func roundToSixDecimalPlaces(value float64) float64 {
	return math.Round(value*1e6) / 1e6
}

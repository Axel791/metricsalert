package sender

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/agent/model/api"
	"github.com/gojek/heimdall/v7/httpclient"
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
		{ID: "Alloc", MType: "gauge", Value: metrics.Alloc},
		{ID: "BuckHashSys", MType: "gauge", Value: metrics.BuckHashSys},
		{ID: "Frees", MType: "gauge", Value: metrics.Frees},
		{ID: "GCCPUFraction", MType: "gauge", Value: metrics.GCCPUFraction},
		{ID: "GCSys", MType: "gauge", Value: metrics.GCSys},
		{ID: "HeapAlloc", MType: "gauge", Value: metrics.HeapAlloc},
		{ID: "HeapIdle", MType: "gauge", Value: metrics.HeapIdle},
		{ID: "HeapInuse", MType: "gauge", Value: metrics.HeapInuse},
		{ID: "HeapObjects", MType: "gauge", Value: metrics.HeapObjects},
		{ID: "HeapReleased", MType: "gauge", Value: metrics.HeapReleased},
		{ID: "HeapSys", MType: "gauge", Value: metrics.HeapSys},
		{ID: "LastGC", MType: "gauge", Value: metrics.LastGC},
		{ID: "Lookups", MType: "gauge", Value: metrics.Lookups},
		{ID: "MCacheInuse", MType: "gauge", Value: metrics.MCacheInuse},
		{ID: "MSpanInuse", MType: "gauge", Value: metrics.MSpanInuse},
		{ID: "MSpanSys", MType: "gauge", Value: metrics.MSpanSys},
		{ID: "Mallocs", MType: "gauge", Value: metrics.Mallocs},
		{ID: "NextGC", MType: "gauge", Value: metrics.NextGC},
		{ID: "NumGC", MType: "gauge", Value: metrics.NumGC},
		{ID: "NumForcedGC", MType: "gauge", Value: metrics.NumForcedGC},
		{ID: "OtherSys", MType: "gauge", Value: metrics.OtherSys},
		{ID: "PauseTotalNs", MType: "gauge", Value: metrics.PauseTotalNs},
		{ID: "StackInuse", MType: "gauge", Value: metrics.StackInuse},
		{ID: "Sys", MType: "gauge", Value: metrics.Sys},
		{ID: "TotalAlloc", MType: "gauge", Value: metrics.TotalAlloc},
		{ID: "PollCount", MType: "counter", Delta: metrics.PollCount},
		{ID: "RandomValue", MType: "gauge", Value: metrics.RandomValue},
	}

	for _, metric := range metricsList {
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

func (client *MetricClient) HealthCheck() error {
	u, err := url.Parse(fmt.Sprintf("%s/healthcheck/", client.baseURL))
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	rsp, err := client.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send healthcheck request: %w", err)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", rsp.StatusCode)
	}

	var response map[string]interface{}
	if err := json.NewDecoder(rsp.Body).Decode(&response); err != nil {
		return fmt.Errorf("failed to decode healthcheck response: %w", err)
	}

	status, ok := response["status"].(string)
	if !ok || status != "true" {
		return fmt.Errorf("unexpected healthcheck response: %v", response)
	}

	return nil
}

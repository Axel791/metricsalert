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
		{ID: "alloc", MType: "gauge", Value: metrics.Alloc},
		{ID: "buckHashSys", MType: "gauge", Value: metrics.BuckHashSys},
		{ID: "frees", MType: "gauge", Value: metrics.Frees},
		{ID: "gcCPUFraction", MType: "gauge", Value: metrics.GCCPUFraction},
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

package sender

import (
	"bytes"
	"encoding/json"
	"fmt"
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

	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("failed to marshal metric: %w", err)
	}

	u, err := url.Parse(fmt.Sprintf("%s/update", client.baseURL))
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	rsp, err := client.httpClient.Post(u.String(), bytes.NewBuffer(body), headers)
	if err != nil {
		return fmt.Errorf("failed to send metrics %s: %w", metric.ID, err)
	}

	defer rsp.Body.Close()
	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", rsp.StatusCode)
	}

	return nil
}

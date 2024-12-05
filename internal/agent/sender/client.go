package sender

import (
	"bytes"
	"fmt"
	"github.com/Axel791/metricsalert/internal/agent/model/dto"
	"github.com/gojek/heimdall/v7/httpclient"
	"github.com/pkg/errors"
	"net/http"
	"net/url"
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

func (client *MetricClient) SendMetrics(metrics dto.Metrics) error {
	metricsMap := map[string]interface{}{
		"alloc":         metrics.Alloc,
		"buckHashSys":   metrics.BuckHashSys,
		"frees":         metrics.Frees,
		"gcCPUFraction": metrics.GCCPUFraction,
	}

	for name, value := range metricsMap {
		metricType := "counter"

		if _, ok := value.(float64); ok {
			metricType = "gauge"
		}

		err := client.sendMetric(name, metricType, value)

		if err != nil {
			return errors.Wrapf(err, "failed to send metric %s", name)
		}
	}

	return nil
}

func (client *MetricClient) sendMetric(name, metricType string, value interface{}) error {
	headers := http.Header{}
	headers.Set("Content-Type", "text/plain")

	u, err := url.Parse(
		fmt.Sprintf("%s/update/%s/%s/%v", client.baseURL, metricType, name, value),
	)

	if err != nil {
		return errors.Wrap(err, "failed to parse URL")
	}

	rsp, err := client.httpClient.Post(u.String(), bytes.NewBuffer(nil), headers)

	if err != nil {
		return errors.Wrap(err, "failed to send metrics")
	}

	defer rsp.Body.Close()
	return nil
}

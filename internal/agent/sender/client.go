package sender

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"net/url"
	"strings"
	"time"

	pb "github.com/Axel791/metricsalert/internal/metricsgrpc"
	"google.golang.org/grpc"

	"github.com/Axel791/metricsalert/internal/agent/services"

	"github.com/gojek/heimdall/v7/httpclient"
	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/agent/model/api"
)

const (
	maxRetries  = 3
	minInterval = 1 * time.Second
	maxInterval = 5 * time.Second
)

type MetricClient struct {
	httpClient  *httpclient.Client
	grpcClient  pb.MetricsServiceClient
	grpcConn    *grpc.ClientConn
	logger      *log.Logger
	authService services.AuthService
	baseURL     string
	pubKey      *rsa.PublicKey
	useGrpc     bool
}

func NewMetricClient(
	baseURL string,
	logger *log.Logger,
	authService services.AuthService,
	pubKey *rsa.PublicKey,
	useGrpc bool,
) *MetricClient {

	if strings.HasPrefix(baseURL, "grpc://") && useGrpc {
		addr := strings.TrimPrefix(baseURL, "grpc://")

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		//nolint:staticcheck // DialContext и WithBlock помечены как deprecated, но сохраняются до конца 1.x
		conn, err := grpc.DialContext(
			ctx,
			addr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			logger.Fatalf("gRPC dial error: %v", err)
		}

		return &MetricClient{
			grpcClient: pb.NewMetricsServiceClient(conn),
			grpcConn:   conn,
			logger:     logger,
			baseURL:    baseURL,
			pubKey:     pubKey,
		}
	}
	return &MetricClient{
		httpClient:  httpclient.NewClient(),
		authService: authService,
		baseURL:     baseURL,
		logger:      logger,
		pubKey:      pubKey,
	}
}

func (client *MetricClient) SendMetrics(metrics api.Metrics) error {
	metricsList := []api.MetricPost{
		{ID: "Alloc", MType: "gauge", Value: &metrics.Alloc},
		{ID: "BuckHashSys", MType: "gauge", Value: &metrics.BuckHashSys},
		{ID: "Frees", MType: "gauge", Value: &metrics.Frees},
		{ID: "GCCPUFraction", MType: "gauge", Value: &metrics.GCCPUFraction},
		{ID: "GCSys", MType: "gauge", Value: &metrics.GCSys},
		{ID: "HeapAlloc", MType: "gauge", Value: &metrics.HeapAlloc},
		{ID: "HeapIdle", MType: "gauge", Value: &metrics.HeapIdle},
		{ID: "HeapInuse", MType: "gauge", Value: &metrics.HeapInuse},
		{ID: "HeapObjects", MType: "gauge", Value: &metrics.HeapObjects},
		{ID: "HeapReleased", MType: "gauge", Value: &metrics.HeapReleased},
		{ID: "HeapSys", MType: "gauge", Value: &metrics.HeapSys},
		{ID: "LastGC", MType: "gauge", Value: &metrics.LastGC},
		{ID: "Lookups", MType: "gauge", Value: &metrics.Lookups},
		{ID: "MCacheInuse", MType: "gauge", Value: &metrics.MCacheInuse},
		{ID: "MSpanInuse", MType: "gauge", Value: &metrics.MSpanInuse},
		{ID: "MSpanSys", MType: "gauge", Value: &metrics.MSpanSys},
		{ID: "Mallocs", MType: "gauge", Value: &metrics.Mallocs},
		{ID: "NextGC", MType: "gauge", Value: &metrics.NextGC},
		{ID: "NumGC", MType: "gauge", Value: &metrics.NumGC},
		{ID: "NumForcedGC", MType: "gauge", Value: &metrics.NumForcedGC},
		{ID: "OtherSys", MType: "gauge", Value: &metrics.OtherSys},
		{ID: "PauseTotalNs", MType: "gauge", Value: &metrics.PauseTotalNs},
		{ID: "StackInuse", MType: "gauge", Value: &metrics.StackInuse},
		{ID: "Sys", MType: "gauge", Value: &metrics.Sys},
		{ID: "MCacheSys", MType: "gauge", Value: &metrics.MCacheSys},
		{ID: "StackSys", MType: "gauge", Value: &metrics.StackSys},
		{ID: "TotalAlloc", MType: "gauge", Value: &metrics.TotalAlloc},
		{ID: "PollCount", MType: "counter", Delta: &metrics.PollCount},
		{ID: "RandomValue", MType: "gauge", Value: &metrics.RandomValue},
		{ID: "TotalMemory", MType: "gauge", Value: &metrics.TotalMemory},
		{ID: "FreeMemory", MType: "gauge", Value: &metrics.FreeMemory},
	}

	if client.grpcClient != nil {
		return client.sendMetricsBatchGRPC(metricsList)
	}

	if err := client.healthCheck(); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}

	if err := client.healthCheck(); err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	return client.sendMetricsBatch(metricsList)
}

func (client *MetricClient) sendMetricsBatch(metricsList []api.MetricPost) error {
	body, err := json.Marshal(metricsList)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics batch: %w", err)
	}
	compressedBody, err := compressData(body)
	if err != nil {
		return fmt.Errorf("failed to compress metrics batch: %w", err)
	}

	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	headers.Set("Content-Encoding", "gzip")

	payload := compressedBody

	if client.pubKey != nil {
		payload, err = rsa.EncryptOAEP(
			sha256.New(), rand.Reader, client.pubKey, compressedBody, nil,
		)
		if err != nil {
			return fmt.Errorf("failed to encrypt metrics batch: %w", err)
		}
		headers.Set("Content-Encryption", "rsa-oaep-sha256")
	} else {
		token := client.authService.ComputeHash(compressedBody)
		if token != "" {
			headers.Set("HashSHA256", token)
		}
	}

	u, err := url.Parse(fmt.Sprintf("%s/updates", client.baseURL))
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	rsp, err := client.httpClient.Post(u.String(), bytes.NewBuffer(payload), headers)
	if err != nil {
		return fmt.Errorf("failed to send metrics batch: %w", err)
	}
	defer rsp.Body.Close()

	if rsp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", rsp.StatusCode)
	}

	client.logger.Infof("Successfully sent metrics batch: %d metrics", len(metricsList))
	return nil
}

func (client *MetricClient) sendMetricsBatchGRPC(list []api.MetricPost) error {
	pbList := make([]*pb.Metric, 0, len(list))
	for _, m := range list {
		p := &pb.Metric{Id: m.ID, MType: m.MType}
		if m.MType == "gauge" && m.Value != nil {
			p.Value = *m.Value
		}
		if m.MType == "counter" && m.Delta != nil {
			p.Delta = *m.Delta
		}
		pbList = append(pbList, p)
	}
	_, err := client.grpcClient.UpdateBatch(context.Background(),
		&pb.MetricsBatch{Metrics: pbList})
	if err == nil {
		client.logger.Infof("gRPC: sent %d metrics", len(list))
	}
	return err
}

func (client *MetricClient) healthCheck() error {
	u, err := url.Parse(fmt.Sprintf("%s/healthcheck", client.baseURL))
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
			}
			log.Infof(
				"unexpected status code during health check: %d (attempt %d/%d)",
				rsp.StatusCode,
				retries+1,
				maxRetries,
			)
		}

		retries++

		if interval < maxInterval {
			interval += 2 * time.Second
			if interval > maxInterval {
				interval = maxInterval
			}
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("health check failed after %d attempts", retries)
}

func compressData(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)

	if _, err := gz.Write(data); err != nil {
		return nil, fmt.Errorf("failed to write to gzip writer: %w", err)
	}
	if err := gz.Close(); err != nil {
		return nil, fmt.Errorf("failed to close gzip writer: %w", err)
	}

	return buf.Bytes(), nil
}

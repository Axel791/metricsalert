package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/Axel791/metricsalert/internal/server/model/dto"
	"gopkg.in/guregu/null.v4"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	log "github.com/sirupsen/logrus"
)

type stubMetricService struct {
	store map[string]api.Metrics
}

func newStubService() *stubMetricService {
	return &stubMetricService{store: make(map[string]api.Metrics)}
}

// GetMetric возвращает сохранённую метрику по имени и типу.
func (s *stubMetricService) GetMetric(_ context.Context, metricType, name string) (dto.Metrics, error) {
	m, ok := s.store[name]
	if !ok || m.MType != metricType {
		return dto.Metrics{}, fmt.Errorf("not found")
	}
	return dtoFromAPI(m), nil
}

// CreateOrUpdateMetric создаёт либо обновляет запись.
func (s *stubMetricService) CreateOrUpdateMetric(_ context.Context, m api.Metrics) (dto.Metrics, error) {
	s.store[m.ID] = m
	return dtoFromAPI(m), nil
}

// GetAllMetric возвращает все сохранённые метрики.
func (s *stubMetricService) GetAllMetric(_ context.Context) ([]dto.Metrics, error) {
	res := make([]dto.Metrics, 0, len(s.store))
	for _, m := range s.store {
		res = append(res, dtoFromAPI(m))
	}
	return res, nil
}

// BatchMetricsUpdate сохраняет несколько метрик.
func (s *stubMetricService) BatchMetricsUpdate(_ context.Context, metrics []api.Metrics) error {
	for _, m := range metrics {
		s.store[m.ID] = m
	}
	return nil
}

// dtoFromAPI конвертирует api.Metrics → dto.Metrics c использованием null.Int/Float.
func dtoFromAPI(m api.Metrics) dto.Metrics {
	var d dto.Metrics
	d.ID = m.ID
	d.MType = m.MType
	if m.Delta != nil {
		d.Delta = null.IntFrom(*m.Delta)
	}
	if m.Value != nil {
		d.Value = null.FloatFrom(*m.Value)
	}
	return d
}

func ExampleUpdateMetricHandler() {
	svc := newStubService()
	h := NewUpdateMetricHandler(svc, log.New())

	payload := api.Metrics{ID: "Alloc", MType: "gauge", Value: floatPtr(6.27)}
	body, _ := json.Marshal(payload)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))

	h.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	// Output:
	// 200
}

func ExampleGetMetricHandler() {
	svc := newStubService()
	_, _ = svc.CreateOrUpdateMetric(context.Background(), api.Metrics{ID: "Alloc", MType: "gauge", Value: floatPtr(6.27)})

	h := NewGetMetricHandler(svc, log.New())

	reqBody, _ := json.Marshal(api.GetMetric{ID: "Alloc", MType: "gauge"})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(reqBody))

	h.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	// Output:
	// 200
}

func ExampleUpdatesMetricsHandler() {
	svc := newStubService()
	h := NewUpdatesMetricsHandler(svc, log.New())

	batch := []api.Metrics{
		{ID: "Alloc", MType: "gauge", Value: floatPtr(7.01)},
		{ID: "PollCount", MType: "counter", Delta: intPtr(5)},
	}
	body, _ := json.Marshal(batch)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))

	h.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	// Output:
	// 200
}

func ExampleGetMetricsHTMLHandler() {
	svc := newStubService()
	_, _ = svc.CreateOrUpdateMetric(context.Background(), api.Metrics{ID: "Alloc", MType: "gauge", Value: floatPtr(6.27)})
	_ = svc.BatchMetricsUpdate(context.Background(), []api.Metrics{{ID: "PollCount", MType: "counter", Delta: intPtr(3)}})

	h := NewGetMetricsHTMLHandler(svc)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)

	h.ServeHTTP(rr, req)
	fmt.Println(rr.Code)
	// Output:
	// 200
}

func intPtr(v int64) *int64       { return &v }
func floatPtr(v float64) *float64 { return &v }

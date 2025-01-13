package deprecated

import (
	"flag"
	"net/http"
	"net/http/httptest"
	"testing"

	"gopkg.in/guregu/null.v4"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/repositories/mocks"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestUpdateMetricHandler(t *testing.T) {
	originalFlagSet := flag.CommandLine
	defer func() {
		flag.CommandLine = originalFlagSet
	}()

	flag.CommandLine = flag.NewFlagSet("test", flag.ContinueOnError)
	flag.String("a", "localhost:8080", "HTTP server address")

	mockStore := new(mocks.MockStore)

	handler := NewUpdateMetricHandler(mockStore)

	router := chi.NewRouter()
	router.Post("/update/{metricType}/{name}/{value}", handler.ServeHTTP)

	tests := []struct {
		name           string
		urlPath        string
		method         string
		expectedStatus int
		mockBehavior   func()
	}{
		{
			name:           "Valid Gauge Update",
			urlPath:        "/update/gauge/testMetric/123.45",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			mockBehavior: func() {
				mockStore.
					On("UpdateGauge", mock.Anything, "testMetric", 123.45).
					Return(domain.Metrics{
						ID:    "testMetric",
						MType: Gauge,
						Value: null.NewFloat(123.45, true),
					}, nil).
					Once()
			},
		},
		{
			name:           "Valid Counter Update",
			urlPath:        "/update/counter/testMetric/10",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			mockBehavior: func() {
				mockStore.
					On("UpdateCounter", mock.Anything, "testMetric", int64(10)).
					Return(domain.Metrics{
						ID:    "testMetric",
						MType: Counter,
						Delta: null.NewInt(10, true),
					}, nil).
					Once()
			},
		},
		{
			name:           "Invalid Metric Type",
			urlPath:        "/update/invalid/testMetric/123",
			method:         http.MethodPost,
			expectedStatus: http.StatusBadRequest,
			mockBehavior:   func() {},
		},
		{
			name:           "Invalid HTTP Method",
			urlPath:        "/update/gauge/testMetric/123.45",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
			mockBehavior:   func() {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockBehavior()

			req := httptest.NewRequest(tt.method, tt.urlPath, nil)
			rec := httptest.NewRecorder()

			router.ServeHTTP(rec, req)

			require.Equal(t, tt.expectedStatus, rec.Code)

			mockStore.AssertExpectations(t)
		})
	}
}

package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/model/dto"
	"github.com/Axel791/metricsalert/internal/server/services/mock"
)

func TestUpdateMetricHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		input          api.Metrics
		mockSetup      func(*mock.MockMetric)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful counter update",
			input: api.Metrics{
				ID:    "testCounter",
				MType: domain.Counter,
				Delta: int64Ptr(42),
			},
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().CreateOrUpdateMetric(gomock.Any(), api.Metrics{
					ID:    "testCounter",
					MType: domain.Counter,
					Delta: int64Ptr(42),
				}).Return(dto.Metrics{
					ID:    "testCounter",
					MType: domain.Counter,
					Delta: null.Int{NullInt64: sql.NullInt64{Int64: 42, Valid: true}},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: dto.Metrics{
				ID:    "testCounter",
				MType: domain.Counter,
				Delta: null.Int{NullInt64: sql.NullInt64{Int64: 42, Valid: true}},
			},
		},
		{
			name: "successful gauge update",
			input: api.Metrics{
				ID:    "testGauge",
				MType: domain.Gauge,
				Value: float64Ptr(3.14),
			},
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().CreateOrUpdateMetric(gomock.Any(), api.Metrics{
					ID:    "testGauge",
					MType: domain.Gauge,
					Value: float64Ptr(3.14),
				}).Return(dto.Metrics{
					ID:    "testGauge",
					MType: domain.Gauge,
					Value: null.Float{NullFloat64: sql.NullFloat64{Float64: 3.14, Valid: true}},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: dto.Metrics{
				ID:    "testGauge",
				MType: domain.Gauge,
				Value: null.Float{NullFloat64: sql.NullFloat64{Float64: 3.14, Valid: true}},
			},
		},
		{
			name: "invalid metric type",
			input: api.Metrics{
				ID:    "invalid",
				MType: "invalidType",
			},
			mockSetup:      func(_ *mock.MockMetric) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name: "missing required fields",
			input: api.Metrics{
				ID:    "",
				MType: domain.Gauge,
			},
			mockSetup:      func(_ *mock.MockMetric) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
		{
			name: "service error",
			input: api.Metrics{
				ID:    "errorCase",
				MType: domain.Counter,
				Delta: int64Ptr(1),
			},
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().CreateOrUpdateMetric(gomock.Any(), api.Metrics{
					ID:    "errorCase",
					MType: domain.Counter,
					Delta: int64Ptr(1),
				}).Return(dto.Metrics{}, errors.New("service error"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMetric := mock.NewMockMetric(ctrl)
			tt.mockSetup(mockMetric)

			handler := NewUpdateMetricHandler(mockMetric, nil)

			reqBody, err := json.Marshal(tt.input)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response dto.Metrics
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			} else {
				assert.Contains(t, rr.Body.String(), tt.expectedBody.(string))
			}
		})
	}
}

func TestUpdateMetricHandler_EncodingError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetric := mock.NewMockMetric(ctrl)
	mockMetric.EXPECT().CreateOrUpdateMetric(gomock.Any(), gomock.Any()).
		Return(dto.Metrics{
			ID:    "testCounter",
			MType: domain.Counter,
			Delta: null.Int{NullInt64: sql.NullInt64{Int64: 42, Valid: true}},
		}, nil)

	handler := NewUpdateMetricHandler(mockMetric, nil)

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"id":"testCounter","type":"counter","delta":42}`))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	w := &errorResponseWriter{ResponseWriter: rr, failAfter: 0}
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "failed to encode response")
}

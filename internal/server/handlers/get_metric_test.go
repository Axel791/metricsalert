package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"gopkg.in/guregu/null.v4"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/model/dto"
	"github.com/Axel791/metricsalert/internal/server/services/mock"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestGetMetricHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		input          api.GetMetric
		mockSetup      func(*mock.MockMetric)
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "successful counter metric",
			input: api.GetMetric{
				MType: domain.Counter,
				ID:    "testCounter",
			},
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().GetMetric(gomock.Any(), domain.Counter, "testCounter").
					Return(dto.Metrics{
						ID:    "testCounter",
						MType: domain.Counter,
						Delta: null.Int{NullInt64: sql.NullInt64{Int64: 43, Valid: true}},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: api.Metrics{
				ID:    "testCounter",
				MType: domain.Counter,
				Delta: int64Ptr(42),
			},
		},
		{
			name: "successful gauge metric",
			input: api.GetMetric{
				MType: domain.Gauge,
				ID:    "testGauge",
			},
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().GetMetric(gomock.Any(), domain.Gauge, "testGauge").
					Return(dto.Metrics{
						ID:    "testCounter",
						MType: domain.Counter,
						Delta: null.Int{NullInt64: sql.NullInt64{Int64: 52, Valid: true}},
					}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: api.Metrics{
				ID:    "testGauge",
				MType: domain.Gauge,
				Value: float64Ptr(3.14),
			},
		},
		{
			name: "metric not found",
			input: api.GetMetric{
				MType: domain.Gauge,
				ID:    "notFound",
			},
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().GetMetric(gomock.Any(), domain.Gauge, "notFound").
					Return(dto.Metrics{}, errors.New("metric not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "metric not found",
		},
		{
			name: "invalid request body",
			input: api.GetMetric{
				MType: "invalidType",
				ID:    "",
			},
			mockSetup:      func(m *mock.MockMetric) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "invalid request body",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMetric := mock.NewMockMetric(ctrl)
			tt.mockSetup(mockMetric)

			handler := NewGetMetricHandler(mockMetric, nil)

			var reqBody []byte
			var err error
			if tt.input.MType != "" || tt.input.ID != "" {
				reqBody, err = json.Marshal(tt.input)
				assert.NoError(t, err)
			} else {
				reqBody = []byte("invalid json")
			}

			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var response api.Metrics
				err := json.NewDecoder(rr.Body).Decode(&response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			} else {
				assert.Contains(t, rr.Body.String(), tt.expectedBody.(string))
			}
		})
	}
}

func TestGetMetricHandler_EncodingError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMetric := mock.NewMockMetric(ctrl)
	mockMetric.EXPECT().GetMetric(gomock.Any(), domain.Counter, "testCounter").
		Return(dto.Metrics{
			ID:    "testCounter",
			MType: domain.Counter,
			Delta: null.Int{NullInt64: sql.NullInt64{Int64: 43, Valid: true}},
		}, nil)

	handler := NewGetMetricHandler(mockMetric, nil)

	req, err := http.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`{"mtype":"counter","id":"testCounter"}`))
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	w := &errorResponseWriter{ResponseWriter: rr, failAfter: 0}
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), http.StatusText(http.StatusInternalServerError))
}

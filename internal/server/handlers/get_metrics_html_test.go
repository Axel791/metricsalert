package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"gopkg.in/guregu/null.v4"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/model/dto"
	"github.com/Axel791/metricsalert/internal/server/services/mock"
)

func TestGetMetricsHTMLHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*mock.MockMetric)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful with metrics",
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().GetAllMetric(gomock.Any()).Return([]dto.Metrics{
					{
						ID:    "testCounter",
						MType: domain.Counter,
						Delta: null.Int{NullInt64: sql.NullInt64{Int64: 42, Valid: true}},
					},
					{
						ID:    "testGauge",
						MType: domain.Gauge,
						Value: null.Float{NullFloat64: sql.NullFloat64{Float64: 3.14, Valid: true}},
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `<td>testCounter</td>.*<td>int64</td>.*<td>42</td>.*<td>testGauge</td>.*<td>float64</td>.*<td>3.14</td>`,
		},
		{
			name: "successful with empty metrics",
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().GetAllMetric(gomock.Any()).Return([]dto.Metrics{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `<h1>Список метрик</h1>`,
		},
		{
			name: "service error",
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().GetAllMetric(gomock.Any()).Return(nil, errors.New("service error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "service error",
		},
		{
			name: "unknown metric type",
			mockSetup: func(m *mock.MockMetric) {
				m.EXPECT().GetAllMetric(gomock.Any()).Return([]dto.Metrics{
					{
						ID:    "unknown",
						MType: "unknownType",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `<td>unknown</td>.*<td>string</td>.*<td>unknown</td>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockMetric := mock.NewMockMetric(ctrl)
			tt.mockSetup(mockMetric)

			handler := NewGetMetricsHTMLHandler(mockMetric)

			req, err := http.NewRequest(http.MethodGet, "/", nil)
			assert.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Equal(t, "text/html; charset=utf-8", rr.Header().Get("Content-Type"))

			if tt.expectedStatus == http.StatusOK {
				assert.Regexp(t, tt.expectedBody, rr.Body.String())
			} else {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}
		})
	}
}

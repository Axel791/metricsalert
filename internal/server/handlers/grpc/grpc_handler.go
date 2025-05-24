package server

import (
	"context"
	"fmt"

	"google.golang.org/protobuf/types/known/emptypb"

	pb "github.com/Axel791/metricsalert/internal/metricsgrpc"
	"github.com/Axel791/metricsalert/internal/server/model/api"
	"github.com/Axel791/metricsalert/internal/server/services"
)

const (
	Counter = "counter"
	Gauge   = "gauge"
)

// grpcHandler реализует сгенерированный интерфейс.
type grpcHandler struct {
	pb.UnimplementedMetricsServiceServer
	metricsSvc *services.MetricsService
}

// NewGRPCHandler создаёт адаптер поверх MetricsService.
func NewGRPCHandler(svc *services.MetricsService) pb.MetricsServiceServer {
	return &grpcHandler{metricsSvc: svc}
}

// UpdateBatch приходит от агента.
func (h *grpcHandler) UpdateBatch(ctx context.Context, in *pb.MetricsBatch) (*emptypb.Empty, error) {
	if in == nil || len(in.Metrics) == 0 {
		return &emptypb.Empty{}, nil
	}

	apiList := make([]api.Metrics, 0, len(in.Metrics))
	for _, m := range in.Metrics {
		item := api.Metrics{
			ID:    m.Id,
			MType: m.MType,
		}
		switch m.MType {
		case Gauge:
			v := m.Value
			item.Value = &v
		case Counter:
			d := m.Delta
			item.Delta = &d
		default:
			return nil, fmt.Errorf("unknown metric type: %s", m.MType)
		}
		apiList = append(apiList, item)
	}

	if err := h.metricsSvc.BatchMetricsUpdate(ctx, apiList); err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, nil
}

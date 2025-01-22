package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/pkg/errors"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
)

type FileStoreHandler struct {
	memoryStore Store
	filePath    string
	mutex       *sync.Mutex
}

// NewFileStore создает новый экземпляр FileStoreHandler.
func NewFileStore(
	ctx context.Context,
	memoryStore Store,
	filePath string,
	restoreFlag bool,
	storeInterval time.Duration,
) (*FileStoreHandler, error) {
	fs := &FileStoreHandler{
		memoryStore: memoryStore,
		filePath:    filePath,
		mutex:       &sync.Mutex{},
	}
	if restoreFlag {
		if err := fs.load(ctx); err != nil {
			log.Warnf("failed to load metrics from file %q: %v", filePath, err)
			return nil, fmt.Errorf("failed to load metrics from file %q: %w", filePath, err)
		}
	}

	fs.startAutoSave(ctx, storeInterval)

	return fs, nil
}

// UpdateGauge обновляет Gauge в памяти и возвращает обновлённую метрику
func (fs *FileStoreHandler) UpdateGauge(ctx context.Context, name string, value float64) (domain.Metrics, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	metric, err := fs.memoryStore.UpdateGauge(ctx, name, value)
	if err != nil {
		return domain.Metrics{}, fmt.Errorf("failed to update counter %q: %v", name, err)
	}

	_ = fs.saveToFile(ctx)
	return metric, nil
}

// UpdateCounter обновляет Counter в памяти и возвращает обновлённую метрику
func (fs *FileStoreHandler) UpdateCounter(ctx context.Context, name string, value int64) (domain.Metrics, error) {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	metric, err := fs.memoryStore.UpdateCounter(ctx, name, value)
	if err != nil {
		return domain.Metrics{}, fmt.Errorf("failed to update counter %q: %v", name, err)
	}
	_ = fs.saveToFile(ctx)
	return metric, nil
}

// GetMetric возвращает метрику из памяти
func (fs *FileStoreHandler) GetMetric(ctx context.Context, metricsDomain domain.Metrics) (domain.Metrics, error) {
	metric, err := fs.memoryStore.GetMetric(ctx, metricsDomain)
	if err != nil {
		return domain.Metrics{}, fmt.Errorf("failed to get metric %s: %v", metricsDomain.Name, err)
	}
	return metric, nil
}

// GetAllMetrics возвращает все метрики из памяти
func (fs *FileStoreHandler) GetAllMetrics(ctx context.Context) (map[string]domain.Metrics, error) {
	metric, err := fs.memoryStore.GetAllMetrics(ctx)
	if err != nil {
		log.Warnf("failed to get all metrics: %v", err)
		return nil, err
	}
	return metric, nil
}

func (fs *FileStoreHandler) BatchUpdateMetrics(ctx context.Context, m []domain.Metrics) error {
	return fs.memoryStore.BatchUpdateMetrics(ctx, m)
}

// Load загружает метрики из файла в память
func (fs *FileStoreHandler) load(ctx context.Context) error {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	file, err := os.Open(fs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			log.Infof("File %s does not exist, skipping load", fs.filePath)
			return nil
		}
		return err
	}
	defer file.Close()

	var data map[string]domain.Metrics
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return err
	}

	for name, metric := range data {
		switch metric.MType {
		case domain.Counter:
			_, err := fs.memoryStore.UpdateCounter(ctx, name, metric.Delta.Int64)
			if err != nil {
				return fmt.Errorf("failed to update metric %q: %v", name, err)
			}
		case domain.Gauge:

			_, err := fs.memoryStore.UpdateGauge(ctx, name, metric.Value.Float64)
			if err != nil {
				return fmt.Errorf("failed to update metric %q: %v", name, err)
			}
		}
	}
	return nil
}

// StartAutoSave запускает периодическое сохранение.
func (fs *FileStoreHandler) startAutoSave(ctx context.Context, storeInterval time.Duration) {
	if storeInterval <= 0 {
		log.Infof("Auto-save disabled (storeInterval=%v)", storeInterval)
		return
	}

	ticker := time.NewTicker(storeInterval)
	log.Infof("Starting auto-save every %s seconds", storeInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info("Auto-save context canceled, stopping...")
				return
			case <-ticker.C:
				if err := fs.saveToFile(ctx); err != nil {
					log.Errorf("failed to auto-save metrics: %v", err)
				}
			}
		}
	}()
}

// SaveToFile сохраняет все метрики в файл
func (fs *FileStoreHandler) saveToFile(ctx context.Context) error {
	file, err := os.Create(fs.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := fs.memoryStore.GetAllMetrics(ctx)
	if err != nil {
		return errors.New(fmt.Sprintf("failed to get all metrics: %v", err))
	}
	return json.NewEncoder(file).Encode(data)
}

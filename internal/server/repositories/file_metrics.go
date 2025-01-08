package repositories

import (
	"encoding/json"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
)

type FileStoreHandler struct {
	memoryStore Store
	filePath    string
	mutex       *sync.Mutex
}

// NewFileStore создает новый экземпляр FileStoreHandler.
func NewFileStore(memoryStore Store, filePath string) *FileStoreHandler {
	return &FileStoreHandler{
		memoryStore: memoryStore,
		filePath:    filePath,
		mutex:       &sync.Mutex{},
	}
}

// UpdateGauge обновляет Gauge в памяти и возвращает обновлённую метрику
func (fs *FileStoreHandler) UpdateGauge(name string, value float64) domain.Metrics {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	metric := fs.memoryStore.UpdateGauge(name, value)
	_ = fs.SaveToFile()
	return metric
}

// UpdateCounter обновляет Counter в памяти и возвращает обновлённую метрику
func (fs *FileStoreHandler) UpdateCounter(name string, value int64) domain.Metrics {
	fs.mutex.Lock()
	defer fs.mutex.Unlock()

	metric := fs.memoryStore.UpdateCounter(name, value)
	_ = fs.SaveToFile()
	return metric
}

// GetMetric возвращает метрику из памяти
func (fs *FileStoreHandler) GetMetric(metricsDomain domain.Metrics) domain.Metrics {
	return fs.memoryStore.GetMetric(metricsDomain)
}

// GetAllMetrics возвращает все метрики из памяти
func (fs *FileStoreHandler) GetAllMetrics() map[string]domain.Metrics {
	return fs.memoryStore.GetAllMetrics()
}

// Load загружает метрики из файла в память
func (fs *FileStoreHandler) Load() error {
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
			fs.memoryStore.UpdateCounter(name, metric.Delta.Int64)
		case domain.Gauge:
			fs.memoryStore.UpdateGauge(name, metric.Value.Float64)
		}
	}
	return nil
}

// SaveToFile сохраняет все метрики в файл
func (fs *FileStoreHandler) SaveToFile() error {
	file, err := os.Create(fs.filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	data := fs.memoryStore.GetAllMetrics()
	return json.NewEncoder(file).Encode(data)
}

package services

import (
	"context"
	"encoding/json"
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/server/model/domain"
	"github.com/Axel791/metricsalert/internal/server/repositories"
)

type FileStorageService struct {
	repo          repositories.Store
	filePath      string
	storeInterval time.Duration
}

// NewFileStorageService создаёт новый сервис для работы с файлами
func NewFileStorageService(
	repo repositories.Store,
	filePath string,
	storeInterval time.Duration,
) *FileStorageService {
	return &FileStorageService{
		repo:          repo,
		filePath:      filePath,
		storeInterval: storeInterval,
	}
}

// Load загружает ранее сохранённые метрики из файла (если файл существует).
func (s *FileStorageService) Load() error {
	log.Infof("Loading metrics from file: %s", s.filePath)

	file, err := os.Open(s.filePath)
	if err != nil {
		log.Warnf("Can't open file %s for loading: %v", s.filePath, err)
		return err
	}
	defer file.Close()

	var data map[string]domain.Metrics
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		log.Warnf("Can't decode JSON from file %s: %v", s.filePath, err)
		return err
	}

	for name, m := range data {
		switch m.MType {
		case domain.Counter:
			s.repo.UpdateCounter(name, m.Delta.Int64)
		case domain.Gauge:
			s.repo.UpdateGauge(name, m.Value.Float64)
		default:
			log.Warnf("Unknown metric type %s for metric %s", m.MType, name)
		}
	}

	log.Infof("Metrics loaded successfully from %s", s.filePath)
	return nil
}

// Save сохраняет метрики в файл
func (s *FileStorageService) Save() error {
	log.Infof("Saving metrics to file: %s", s.filePath)

	file, err := os.Create(s.filePath)
	if err != nil {
		log.Errorf("Failed to create file %s: %v", s.filePath, err)
		return err
	}
	defer file.Close()

	data := s.repo.GetAllMetrics()

	if err := json.NewEncoder(file).Encode(data); err != nil {
		log.Errorf("Failed to encode metrics to file %s: %v", s.filePath, err)
		return err
	}

	log.Infof("Metrics saved successfully to %s", s.filePath)
	return nil
}

// StartAutoSave запускает отдельную горутину, которая каждые storeInterval
// вызывает Save(). Если storeInterval <= 0, автосохранение не запускается.
func (s *FileStorageService) StartAutoSave(ctx context.Context) {
	if s.storeInterval <= 0 {
		log.Infof("Auto-save disabled (storeInterval=%v)", s.storeInterval)
		return
	}

	ticker := time.NewTicker(s.storeInterval * time.Second)
	log.Infof("Starting auto-save every %s", s.storeInterval)

	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				log.Info("Auto-save context done, stopping...")
				return
			case <-ticker.C:
				_ = s.Save()
			}
		}
	}()
}

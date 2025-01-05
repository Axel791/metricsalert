package services

import (
	"context"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/Axel791/metricsalert/internal/server/repositories"
)

type FileStorageService struct {
	store         repositories.FileStore
	storeInterval time.Duration
}

// NewFileStorageService создаёт новый сервис для работы с файлами
func NewFileStorageService(
	store repositories.FileStore,
	storeInterval time.Duration,
) *FileStorageService {
	return &FileStorageService{
		store:         store,
		storeInterval: storeInterval,
	}
}

// Load загружает метрики из файла через FileStore
func (s *FileStorageService) Load() error {
	return s.store.Load()
}

// Save сохраняет метрики через FileStore
func (s *FileStorageService) Save() error {
	return s.store.SaveToFile()
}

// StartAutoSave запускает периодическое сохранение
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

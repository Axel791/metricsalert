package services

import (
	"context"
	"time"

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
		return
	}

	ticker := time.NewTicker(s.storeInterval)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = s.Save()
			}
		}
	}()
}

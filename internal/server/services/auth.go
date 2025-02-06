package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"

	log "github.com/sirupsen/logrus"
)

// AuthServiceHandler - структура сервиса auth
type AuthServiceHandler struct {
	key string
}

// NewAuthService - инициализация сервиса auth
func NewAuthService(key string) *AuthServiceHandler {
	return &AuthServiceHandler{key: key}
}

// Validate - валидация входящего токена
func (s *AuthServiceHandler) Validate(token string) error {
	if s.key == "" {
		return nil
	}

	log.Infof("token: %s", token)
	log.Infof("key server: %s", s.key)

	hash := sha256.Sum256([]byte(s.key))
	computedHash := hex.EncodeToString(hash[:])

	if token != computedHash {
		return fmt.Errorf("invalid token")
	}
	return nil
}

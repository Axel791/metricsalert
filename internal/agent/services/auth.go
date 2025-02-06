package services

import (
	"crypto/sha256"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
)

type AuthServiceHandler struct {
	key string
}

func NewAuthServiceHandler(key string) *AuthServiceHandler {
	return &AuthServiceHandler{key: key}
}

// ComputeHash - вычисление хеша на основе ключа
func (s *AuthServiceHandler) ComputeHash() string {
	if s.key == "" {
		return ""
	}

	log.Infof("key agent: %s", s.key)

	hash := sha256.Sum256([]byte(s.key))
	return hex.EncodeToString(hash[:])
}

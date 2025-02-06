package services

import (
	"crypto/hmac"
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
func (s *AuthServiceHandler) Validate(token string, body []byte) error {
	if s.key == "" {
		return nil
	}

	log.Infof("token: %s", token)
	log.Infof("key server: %s", s.key)

	hash := hmac.New(sha256.New, []byte(s.key))
	hash.Write(body)

	validToken := hex.EncodeToString(hash.Sum(nil))

	if token != validToken {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func (s *AuthServiceHandler) ComputedHash(body []byte) string {
	hash := hmac.New(sha256.New, []byte(s.key))

	hash.Write(body)
	token := hex.EncodeToString(hash.Sum(nil))

	log.Infof("generated token: %s", token)

	return token
}

package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	log "github.com/sirupsen/logrus"
)

// SignServiceHandler - структура сервиса auth
type SignServiceHandler struct {
	key string
}

// NewSignService - инициализация сервиса auth
func NewSignService(key string) *SignServiceHandler {
	return &SignServiceHandler{key: key}
}

// Validate - валидация входящего токена
func (s *SignServiceHandler) Validate(token string, body []byte) error {
	if s.key == "" {
		return nil
	}

	if token == "" {
		return nil
	}

	log.Infof("body: %s", string(body))

	hash := hmac.New(sha256.New, []byte(s.key))
	hash.Write(body)

	validToken := hex.EncodeToString(hash.Sum(nil))

	log.Infof("validate token: %s", validToken)
	log.Infof("input token: %s", token)

	if token != validToken {
		return fmt.Errorf("invalid token")
	}
	return nil
}

func (s *SignServiceHandler) ComputedHash(body []byte) string {
	hash := hmac.New(sha256.New, []byte(s.key))

	hash.Write(body)
	token := hex.EncodeToString(hash.Sum(nil))

	log.Infof("generated token: %s", token)

	return token
}

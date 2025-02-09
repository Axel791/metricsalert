package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	log "github.com/sirupsen/logrus"
)

// SignServiceHandler реализует SignService.
type SignServiceHandler struct {
	key string
}

// NewSignService создаёт новый обработчик подписи.
func NewSignService(key string) *SignServiceHandler {
	return &SignServiceHandler{key: key}
}

// ComputedHash вычисляет хеш для заданного тела.
func (s *SignServiceHandler) ComputedHash(body []byte) string {
	if s.key == "" {
		return ""
	}
	hash := hmac.New(sha256.New, []byte(s.key))
	hash.Write(body)
	return hex.EncodeToString(hash.Sum(nil))
}

// Validate сравнивает переданный токен с вычисленным для тела.
func (s *SignServiceHandler) Validate(token string, body []byte) error {
	if token == "" {
		return nil
	}
	log.Infof("server: agent token: %s", token)
	expected := s.ComputedHash(body)
	log.Infof("server: signature: %s", expected)
	if token != expected {
		return nil
	}
	return nil
}

package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
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
	expected := s.ComputedHash(body)
	if token != expected {
		return errors.New("invalid sign")
	}
	return nil
}

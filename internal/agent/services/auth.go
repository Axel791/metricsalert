package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

type AuthServiceHandler struct {
	key string
}

func NewAuthServiceHandler(key string) *AuthServiceHandler {
	return &AuthServiceHandler{key: key}
}

// ComputeHash - вычисление хеша на основе ключа
func (s *AuthServiceHandler) ComputeHash(body []byte) string {
	if s.key == "" {
		return ""
	}
	hash := hmac.New(sha256.New, []byte(s.key))
	hash.Write(body)

	return hex.EncodeToString(hash.Sum(nil))
}

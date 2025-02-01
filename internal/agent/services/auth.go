package services

import (
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
func (s *AuthServiceHandler) ComputeHash() string {
	if s.key == "" {
		return ""
	}

	hash := sha256.Sum256([]byte(s.key))
	return hex.EncodeToString(hash[:])
}

package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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

	return base64.StdEncoding.EncodeToString(hash.Sum(nil))
}

package services

import (
	"crypto/hmac"
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
func (s *AuthServiceHandler) ComputeHash(body []byte) string {
	if s.key == "" {
		return ""
	}
	log.Infof("agent key: %s", s.key)
	log.Infof("agent body: %s", string(body))

	hash := hmac.New(sha256.New, []byte(s.key))
	log.Infof("hash before add bosy agent: %s", hash)

	hash.Write(body)

	return hex.EncodeToString(hash.Sum(nil))
}

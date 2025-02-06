package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

	log.Infof("body: %s", string(body))

	normalizedBody, err := s.normalizeBody(body)

	if err != nil {
		return err
	}

	hash := hmac.New(sha256.New, []byte(s.key))
	hash.Write(normalizedBody)

	validToken := hex.EncodeToString(hash.Sum(nil))

	log.Infof("validate token: %s", validToken)
	log.Infof("input token: %s", token)

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

func (s *AuthServiceHandler) normalizeBody(body []byte) ([]byte, error) {
	var rawMessage json.RawMessage
	if err := json.Unmarshal(body, &rawMessage); err != nil {
		return nil, fmt.Errorf("failed to unmarshal body: %w", err)
	}

	normalizedBody, err := json.Marshal(rawMessage)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal normalized body: %w", err)
	}

	log.Infof("normalized body batch update metrics: %s", string(normalizedBody))
	return normalizedBody, nil
}

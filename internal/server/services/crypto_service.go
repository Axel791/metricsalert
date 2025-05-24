package services

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
)

// CryptoServiceHandler — реализация.
type CryptoServiceHandler struct {
	priv *rsa.PrivateKey
}

// NewCryptoService создаёт сервис из PEM-файла с приватным ключом.
func NewCryptoService(pemPath string) (*CryptoServiceHandler, error) {
	data, err := os.ReadFile(pemPath)
	if err != nil {
		return nil, fmt.Errorf("read key: %w", err)
	}
	block, _ := pem.Decode(data)
	if block == nil {
		return nil, fmt.Errorf("no PEM block")
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, fmt.Errorf("parse key: %w", err)
	}
	return &CryptoServiceHandler{priv: priv}, nil
}

// Decrypt раскрывает шифротекст.
func (c *CryptoServiceHandler) Decrypt(cipherText []byte) ([]byte, error) {
	plain, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, c.priv, cipherText, nil)
	if err != nil {
		return nil, fmt.Errorf("decrypt: %w", err)
	}
	return plain, nil
}

package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/Axel791/metricsalert/internal/server/services"
)

// CryptoMiddleware расшифровывает тело, если заголовок
// Content-Encryption: rsa-oaep-sha256 присутствует.
func CryptoMiddleware(cryptoSvc services.CryptoService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			if r.Header.Get("Content-Encryption") == "rsa-oaep-sha256" {
				cipherBody, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "read body error", http.StatusBadRequest)
					return
				}
				plain, err := cryptoSvc.Decrypt(cipherBody)
				if err != nil {
					http.Error(w, "decrypt error: "+err.Error(), http.StatusBadRequest)
					return
				}
				r.Body = io.NopCloser(bytes.NewReader(plain))
				r.Header.Del("Content-Encryption")
			}

			next.ServeHTTP(w, r)
		})
	}
}

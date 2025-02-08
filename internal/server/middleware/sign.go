package middleware

import (
	"bytes"
	"fmt"
	"github.com/Axel791/metricsalert/internal/server/services"
	"io"
	"net/http"
)

// responseCapture реализует http.ResponseWriter и перехватывает данные,
// записанные обработчиком.
type responseCapture struct {
	header     http.Header
	body       bytes.Buffer
	statusCode int
}

// newResponseCapture инициализирует новый responseCapture.
func newResponseCapture() *responseCapture {
	return &responseCapture{
		header:     make(http.Header),
		statusCode: http.StatusOK,
	}
}

func (rc *responseCapture) Header() http.Header {
	return rc.header
}

// Write записывает данные в буфер.
func (rc *responseCapture) Write(b []byte) (int, error) {
	return rc.body.Write(b)
}

// WriteHeader сохраняет код статуса.
func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
}

// SignatureMiddleware возвращает middleware, которое принимает SignServiceHandler
// и оборачивает следующий http.Handler, выполняя валидацию входящего запроса
// (для методов POST, PUT, PATCH) и подписывая ответ.
func SignatureMiddleware(signService services.SignService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
				body, err := validateBody(r)
				if err != nil {
					http.Error(w, "error read body", http.StatusBadRequest)
					return
				}

				token := r.Header.Get("HashSHA256")

				if err := signService.Validate(token, body); err != nil {
					http.Error(w, fmt.Sprintf("invalid sign: %v", err), http.StatusBadRequest)
					return
				}
			}

			rc := newResponseCapture()

			next.ServeHTTP(rc, r)

			newToken := signService.ComputedHash(rc.body.Bytes())

			for key, values := range rc.header {
				for _, v := range values {
					w.Header().Add(key, v)
				}
			}

			w.Header().Set("HashSHA256", newToken)

			w.WriteHeader(rc.statusCode)
			_, _ = w.Write(rc.body.Bytes())
		})
	}
}

// validateBody - валидация тела запроса
func validateBody(r *http.Request) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewBuffer(body))
	if len(body) == 0 {
		return nil, nil
	}
	return body, nil
}

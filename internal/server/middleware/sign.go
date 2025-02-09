package middleware

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/Axel791/metricsalert/internal/server/services"
	"io"
	"net"
	"net/http"
)

// responseCapture перехватывает запись ответа и буферизует тело,
// а также использует оригинальный заголовочный набор.
type responseCapture struct {
	rw         http.ResponseWriter // оригинальный writer
	body       bytes.Buffer
	statusCode int
}

// newResponseCapture создаёт новый перехватчик, устанавливая начальный код статуса.
func newResponseCapture(rw http.ResponseWriter) *responseCapture {
	return &responseCapture{
		rw:         rw,
		statusCode: http.StatusOK,
	}
}

// Header возвращает заголовки оригинального writer’а.
func (rc *responseCapture) Header() http.Header {
	return rc.rw.Header()
}

// Write записывает данные в буфер (не отправляя их сразу).
func (rc *responseCapture) Write(b []byte) (int, error) {
	return rc.body.Write(b)
}

// WriteHeader сохраняет код статуса (фактическая отправка заголовков откладывается).
func (rc *responseCapture) WriteHeader(statusCode int) {
	rc.statusCode = statusCode
}

// Flush делегирует вызов оригинальному ResponseWriter, если он поддерживает http.Flusher.
func (rc *responseCapture) Flush() {
	if flusher, ok := rc.rw.(http.Flusher); ok {
		flusher.Flush()
	}
}

// Hijack делегирует вызов оригинальному ResponseWriter, если он поддерживает http.Hijacker.
func (rc *responseCapture) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hijacker, ok := rc.rw.(http.Hijacker); ok {
		return hijacker.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying ResponseWriter does not support Hijacker")
}

// SignatureMiddleware проверяет входящую подпись (для методов с изменением) и
// после формирования полного ответа вычисляет подпись для него, устанавливая заголовок HashSHA256.
func SignatureMiddleware(signService services.SignService) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Для методов, изменяющих состояние, выполняется валидация подписи запроса.
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

			// Вместо немедленной отправки ответа, перехватываем его.
			rc := newResponseCapture(w)
			next.ServeHTTP(rc, r)

			// Вычисляем подпись на основе полного тела ответа.
			newToken := signService.ComputedHash(rc.body.Bytes())

			// Удаляем старый Content-Length, чтобы избежать несоответствия.
			rc.rw.Header().Del("Content-Length")
			// Устанавливаем заголовок с вычисленной подписью.
			rc.rw.Header().Set("HashSHA256", newToken)

			// Отправляем заголовки с сохранённым статусом и записываем тело.
			rc.rw.WriteHeader(rc.statusCode)
			_, _ = rc.rw.Write(rc.body.Bytes())
		})
	}
}

// validateBody читает и возвращает тело запроса, а затем восстанавливает r.Body.
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

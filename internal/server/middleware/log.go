package middleware

import (
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

// ResponseWriter структура ответа
type ResponseWriter struct {
	http.ResponseWriter
	StatusCode int
	Size       int
}

// WriteHeader перехватывает статус код
func (rw *ResponseWriter) WriteHeader(statusCode int) {
	rw.StatusCode = statusCode
	rw.ResponseWriter.WriteHeader(statusCode)
}

// Write перехватывает размер содержимого
func (rw *ResponseWriter) Write(data []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(data)
	rw.Size += size
	return size, err
}

// WithLogging middleware для получения информации о запросе
func WithLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		rw := &ResponseWriter{
			ResponseWriter: w,
			StatusCode:     http.StatusOK,
		}

		log.Infof("Started %s %s", r.Method, r.RequestURI)

		next.ServeHTTP(rw, r)

		duration := time.Since(start)
		log.Infof(
			"Completed %d %s in %v (%d bytes)",
			rw.StatusCode, http.StatusText(rw.StatusCode), duration, rw.Size,
		)
	})
}

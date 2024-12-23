package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// GzipMiddleware обрабатывает входящие и исходящие gzip-запросы/ответы.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.Header.Get("Content-Encoding"), "gzip") {
			gzipReader, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "failed to decode gzip body", http.StatusBadRequest)
				return
			}
			defer gzipReader.Close()
			r.Body = io.NopCloser(gzipReader)
		}

		if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			gzWriter := gzip.NewWriter(w)
			defer gzWriter.Close()

			w.Header().Set("Content-Encoding", "gzip")
			w = &gzipResponseWriter{ResponseWriter: w, Writer: gzWriter}
		}

		next.ServeHTTP(w, r)
	})
}

// gzipResponseWriter добавляет поддержку gzip для записи ответа.
type gzipResponseWriter struct {
	http.ResponseWriter
	Writer io.Writer
}

// Write оборачивает запись в gzip.Writer.
func (w *gzipResponseWriter) Write(b []byte) (int, error) {
	contentType := w.Header().Get("Content-Type")
	if strings.HasPrefix(contentType, "application/json") || strings.HasPrefix(contentType, "text/html") {
		return w.Writer.Write(b)
	}
	return w.ResponseWriter.Write(b)
}

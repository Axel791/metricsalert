package middleware

import (
	"compress/gzip"
	"net/http"
	"strings"
)

// GzipMiddleware обрабатывает gzip на входе и выходе.
func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.EqualFold(r.Header.Get("Content-Encoding"), "gzip") {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "can't create gzip reader", http.StatusBadRequest)
				return
			}
			defer gz.Close()
			r.Body = gz
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			// Если клиент не поддерживает gzip, то просто вызываем следующий обработчик без обёртки.
			next.ServeHTTP(w, r)
			return
		}

		gzrw := newGzipResponseWriter(w)
		defer gzrw.Close()

		// Передаём наш "прокси" в следующий обработчик
		next.ServeHTTP(gzrw, r)
	})
}

// gzipResponseWriter — обёртка над http.ResponseWriter,
// которая сжимает тело ответа через gzip.Writer,
// но только если Content-Type — application/json или text/html.
type gzipResponseWriter struct {
	http.ResponseWriter
	gzWriter    *gzip.Writer
	gzipStarted bool
}

// newGzipResponseWriter создаёт новую обёртку над ResponseWriter.
func newGzipResponseWriter(w http.ResponseWriter) *gzipResponseWriter {
	return &gzipResponseWriter{
		ResponseWriter: w,
	}
}

// WriteHeader перехватывает установку заголовков.
// Если Content-Type — application/json или text/html,
// мы устанавливаем Content-Encoding: gzip и начинаем сжимать тело.
func (g *gzipResponseWriter) WriteHeader(statusCode int) {
	contentType := g.Header().Get("Content-Type")

	if strings.Contains(contentType, "application/json") ||
		strings.Contains(contentType, "text/html") {

		g.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(g.ResponseWriter)
		g.gzWriter = gz
		g.gzipStarted = true
	}

	g.ResponseWriter.WriteHeader(statusCode)
}

// Write пишет тело ответа.
// Если мы включили gzip (gzipStarted == true),
// то идёт сжатие через g.gzWriter.
func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	if g.gzipStarted {
		return g.gzWriter.Write(b)
	}
	// Если Content-Type не попал под нужные типы,
	// то пишем «как есть» в исходный ResponseWriter.
	return g.ResponseWriter.Write(b)
}

// Close нужен, чтобы закрыть gzip.Writer и «дописать» сжатый поток.
func (g *gzipResponseWriter) Close() error {
	if g.gzipStarted && g.gzWriter != nil {
		return g.gzWriter.Close()
	}
	return nil
}

package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
	"sync"
)

const minSize = 1024

var gzipWriterPool = sync.Pool{
	New: func() any {
		w, _ := gzip.NewWriterLevel(io.Discard, gzip.BestSpeed)
		return w
	},
}
var gzipReaderPool = sync.Pool{
	New: func() any { return new(gzip.Reader) },
}

func GzipMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.EqualFold(r.Header.Get("Content-Encoding"), "gzip") {
			gr := gzipReaderPool.Get().(*gzip.Reader)

			if err := gr.Reset(r.Body); err != nil {
				gzipReaderPool.Put(gr)
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			r.Header.Del("Content-Encoding")
			r.Body = struct {
				io.Reader
				io.Closer
			}{gr, io.NopCloser(nil)}

			defer func() {
				gr.Close()
				gzipReaderPool.Put(gr)
			}()
		}

		if !strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
			next.ServeHTTP(w, r)
			return
		}

		w.Header().Add("Vary", "Accept-Encoding")

		gzrw := &gzipResponseWriter{ResponseWriter: w}
		defer gzrw.closeAsync() // синхронно, до возврата

		next.ServeHTTP(gzrw, r)
	})
}

type gzipResponseWriter struct {
	http.ResponseWriter
	gz   *gzip.Writer
	size int
}

func (g *gzipResponseWriter) Write(b []byte) (int, error) {
	g.size += len(b)

	if g.gz == nil && g.size > minSize && isCompressible(g.Header().Get("Content-Type")) {
		g.Header().Set("Content-Encoding", "gzip")
		gz := gzipWriterPool.Get().(*gzip.Writer)
		gz.Reset(g.ResponseWriter)
		g.gz = gz
	}

	if g.gz != nil {
		return g.gz.Write(b)
	}
	return g.ResponseWriter.Write(b)
}

func (g *gzipResponseWriter) closeAsync() {
	if g.gz != nil {
		g.gz.Close()
		gzipWriterPool.Put(g.gz)
	}
}

func isCompressible(ct string) bool {
	return strings.Contains(ct, "json") || strings.Contains(ct, "text/html")
}

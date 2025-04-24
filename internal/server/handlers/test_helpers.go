package handlers

import (
	"errors"
	"net/http"
)

func int64Ptr(i int64) *int64 {
	return &i
}

func float64Ptr(f float64) *float64 {
	return &f
}

func (w *errorResponseWriter) Write(p []byte) (int, error) {
	if w.bytesWritten >= w.failAfter {
		return 0, errors.New("simulated write error")
	}
	n, err := w.ResponseWriter.Write(p)
	w.bytesWritten += n
	return n, err
}

type errorResponseWriter struct {
	http.ResponseWriter
	failAfter    int
	bytesWritten int
}

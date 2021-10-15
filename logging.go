package main

import (
	"log"
	"net/http"
	"time"

	"github.com/docker/go-units"
)

// LoggingResponseWriter is a wrapper for a response writer to record the response header written by a handler.
type LoggingResponseWriter struct {
	http.ResponseWriter
	StatusCode   int
	BytesWritten int
}

// NewLoggingResponseWriter records the HTTP header.
func NewLoggingResponseWriter(w http.ResponseWriter) *LoggingResponseWriter {
	return &LoggingResponseWriter{w, http.StatusOK, 0}
}

// WriteHeader saves the status code, then passes it on to the underlying response writer.
func (lrw *LoggingResponseWriter) WriteHeader(code int) {
	lrw.StatusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func (lrw *LoggingResponseWriter) Write(b []byte) (int, error) {
	i, err := lrw.ResponseWriter.Write(b)
	lrw.BytesWritten += i
	return i, err
}

// LoggingMiddleware will handle every HTTP request and log using logfmt
func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := NewLoggingResponseWriter(wr)
		next.ServeHTTP(lrw, r)
		end := time.Now()
		duration := end.Sub(start)

		log.Printf("method=%s duration=%s size=%s size_bytes=%d status=%d path=%q time=%d",
			r.Method,
			duration,
			units.HumanSize(float64(lrw.BytesWritten)),
			lrw.BytesWritten,
			lrw.StatusCode,
			r.URL.Path,
			end.UnixNano(),
		)
	})
}

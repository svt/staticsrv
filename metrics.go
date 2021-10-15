package main

import (
	"fmt"
	"log"
	"math"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var codes = []int{}

var (
	// HTTPRequestsTotalCounter is the global http request counter
	HTTPRequestsTotalCounter *prometheus.CounterVec

	// HTTPRequestsDuration is the global http request duration histogram
	HTTPRequestsDuration *prometheus.HistogramVec

	// HTTPRequestsSize is the global http request size histogram
	HTTPRequestsSize *prometheus.HistogramVec
)

func init() {
	HTTPRequestsTotalCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "staticsrv_http_requests_total",
		Help: "The total number of processed HTTP requests for staticsrv",
	}, []string{"method", "status"})
	if err := prometheus.Register(HTTPRequestsTotalCounter); err != nil {
		log.Fatal(err)
	}

	HTTPRequestsDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "staticsrv_http_requests_duration_seconds",
		Help: "Time it has taken to process HTTP requests for staticsrv",
		Buckets: []float64{
			0.0001,
			0.0005,
			0.001,
			0.005,
			0.01,
			0.05,
			0.1,
			1,
			2,
			5,
			10,
			math.Inf(+1),
		},
	}, []string{})
	if err := prometheus.Register(HTTPRequestsDuration); err != nil {
		log.Fatal(err)
	}

	HTTPRequestsSize = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "staticsrv_http_requests_size_bytes",
		Help: "Bytes written to the client over HTTP for staticsrv",
		Buckets: []float64{
			0,
			2,
			4,
			8,
			16,
			32,
			64,
			128,
			256,
			512,
			1024,
			2048,
			4094,
			8192,
			math.Inf(+1),
		},
	}, []string{})
	if err := prometheus.Register(HTTPRequestsSize); err != nil {
		log.Fatal(err)
	}
}

// MetricsMiddleware will handle every HTTP request and increment the request counter.
func MetricsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, r *http.Request) {
		start := time.Now()
		lrw := NewLoggingResponseWriter(wr)

		next.ServeHTTP(lrw, r)

		end := time.Now()
		duration := end.Sub(start)

		HTTPRequestsTotalCounter.With(prometheus.Labels{
			"status": fmt.Sprintf("%d", lrw.StatusCode),
			"method": r.Method,
		}).Inc()

		HTTPRequestsDuration.WithLabelValues().Observe(duration.Seconds())
		HTTPRequestsSize.WithLabelValues().Observe(float64(lrw.BytesWritten))
	})
}

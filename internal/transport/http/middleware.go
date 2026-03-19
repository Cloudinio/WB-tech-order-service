package http

import (
	nethttp "net/http"
	"strconv"

	"github.com/Cloudinio/wb-tech-order-service/internal/metrics"
)

type statusRecorder struct {
	nethttp.ResponseWriter
	statusCode int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.statusCode = code
	r.ResponseWriter.WriteHeader(code)
}

func MetricsMiddleware(m *metrics.Metrics) func(next nethttp.Handler) nethttp.Handler {
	return func(next nethttp.Handler) nethttp.Handler {
		return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
			rec := &statusRecorder{
				ResponseWriter: w,
				statusCode:     nethttp.StatusOK,
			}

			next.ServeHTTP(rec, r)

			m.HTTPRequestTotal.WithLabelValues(
				r.Method,
				r.URL.Path,
				strconv.Itoa(rec.statusCode),
			).Inc()
		})
	}
}
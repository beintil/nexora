package middleware

import (
	"net/http"
	"telephony/pkg/logger"
	"time"

	"github.com/gorilla/mux"
)

// responseWriter is a wrapper around http.ResponseWriter that captures the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// AccessLog returns a middleware that logs HTTP requests with method, path, status code, latency and client IP.
func AccessLog(log logger.Logger) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{w, http.StatusOK}

			next.ServeHTTP(rw, r)

			latency := time.Since(start)
			log.Infof("%s %s %d %s %s",
				r.Method,
				r.URL.Path,
				rw.statusCode,
				latency,
				extractIP(r),
			)
		})
	}
}

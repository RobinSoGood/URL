package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			lw := &loggingResponseWriter{
				ResponseWriter: w,
			}

			h.ServeHTTP(lw, r)

			duration := time.Since(start)

			logger.Info("HTTP Request",
				zap.String("method", r.Method),
				zap.String("uri", r.RequestURI),
				zap.Int("status", lw.status),
				zap.Int64("size", lw.size),
				zap.Duration("duration", duration),
			)
		})
	}
}

type loggingResponseWriter struct {
	http.ResponseWriter
	status int
	size   int64
}

func (lw *loggingResponseWriter) WriteHeader(status int) {
	lw.status = status
	lw.ResponseWriter.WriteHeader(status)
}

func (lw *loggingResponseWriter) Write(b []byte) (int, error) {
	n, err := lw.ResponseWriter.Write(b)
	lw.size += int64(n)
	return n, err
}

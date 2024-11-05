package middleware

import (
	"net/http"

	"go.uber.org/zap"
)

func LoggerMiddleware(logger *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger.Info("Incoming request",
				zap.String("method", r.Method),
				zap.String("url", r.URL.String()),
			)
			next.ServeHTTP(w, r)
		})
	}
}

package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vishenosik/gocherry/pkg/api"
	"github.com/vishenosik/gocherry/pkg/logs"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func RequestLogger(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			lrw := newLoggingResponseWriter(w)
			timeStart := time.Now()

			next.ServeHTTP(lrw, r)

			log := logger.With(
				slog.String("method", fmt.Sprintf("%s %s", r.Method, r.URL.Path)),
				slog.Int("code", lrw.statusCode),
				logs.Took(timeStart),
			)

			switch {
			case api.IsClientError(lrw.statusCode) || api.IsServerError(lrw.statusCode):
				log.Error("request failed with error")
			case api.IsRedirect(lrw.statusCode):
				log.Warn("request redirected")
			default:
				log.Info("request accepted")
			}
		})
	}
}

func newLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(statusCode int) {
	lrw.statusCode = statusCode
	lrw.ResponseWriter.WriteHeader(statusCode)
}

func SetHeaders() func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	}
}

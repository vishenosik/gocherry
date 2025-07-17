package http

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/vishenosik/gocherry/pkg/api"
	"github.com/vishenosik/gocherry/pkg/logs"
)

type requestLogger struct {
	http.ResponseWriter
	statusCode int
	log        *slog.Logger
}

type RequestLoggerOption func(*requestLogger)

func newRequestLogger() *requestLogger {
	log := logs.SetupLogger().With(appComponent())
	return &requestLogger{
		log: log,
	}
}

func (rl *requestLogger) WriteHeader(statusCode int) {
	rl.statusCode = statusCode
	rl.ResponseWriter.WriteHeader(statusCode)
}

func (rl *requestLogger) setWriter(w http.ResponseWriter) {
	rl.ResponseWriter = w
	rl.statusCode = http.StatusOK
}

func RequestLogger(opts ...RequestLoggerOption) func(next http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			rl := newRequestLogger()

			for _, opt := range opts {
				opt(rl)
			}

			rl.setWriter(w)

			timeStart := time.Now()

			next.ServeHTTP(rl, r)

			log := rl.log.With(
				slog.String("method", fmt.Sprintf("%s %s", r.Method, r.URL.Path)),
				slog.Int("code", rl.statusCode),
				logs.Took(timeStart),
			)

			switch {
			case api.IsClientError(rl.statusCode) || api.IsServerError(rl.statusCode):
				log.Error("request failed with error")
			case api.IsRedirect(rl.statusCode):
				log.Warn("request redirected")
			default:
				log.Info("request accepted")
			}
		})
	}
}

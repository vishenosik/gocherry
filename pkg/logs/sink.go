package logs

import (
	"log/slog"
)

type Sink struct {
	logger *slog.Logger
}

func NewSink(logger *slog.Logger) *Sink {
	return &Sink{
		logger: logger,
	}
}

func (sink *Sink) Info(level int, message string, keysAndValues ...any) {
	attrs := make([]any, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		attrs = append(attrs, slog.Attr{
			Key:   keysAndValues[i].(string),
			Value: slog.AnyValue(keysAndValues[i+1]),
		})
	}
	sink.logger.Info(message, attrs...)
}

func (sink *Sink) Error(err error, message string, keysAndValues ...any) {
	attrs := make([]any, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		attrs = append(attrs, slog.Attr{
			Key:   keysAndValues[i].(string),
			Value: slog.AnyValue(keysAndValues[i+1]),
		})
	}
	sink.logger.Error(message, append(attrs, Error(err))...)
}

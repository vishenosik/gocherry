package retry

import (
	"context"
	"log/slog"
	"time"

	"github.com/vishenosik/gocherry/pkg/logs"
)

type Connector interface {
	Connect(ctx context.Context) error
	Retryable(err error) bool
}

type ConnectorOption = func(c *connector)

type connector struct {
	conn Connector
	log  *slog.Logger
}

func NewConnector(conn Connector, opts ...ConnectorOption) *connector {
	c := defaultConnector(conn)
	for _, opt := range opts {
		opt(c)
	}
	return c
}

func (c *connector) Retry(ctx context.Context) error {

	backoff := NewFibonacci(1*time.Second, time.Second*5)
	if err := Do(ctx, backoff, func(ctx context.Context) error {

		c.log.Info("trying to connect")

		if err := c.conn.Connect(ctx); err != nil {

			msg := "failed to connect"

			if c.conn.Retryable(err) {
				c.log.Error(msg,
					slog.Int64("retry_in_seconds", backoff.RetryInSeconds()),
					logs.Error(err),
				)
				return RetryableError(err)
			}

			c.log.Error(msg, logs.Error(err))

			return err
		}

		c.log.Info("connected successfuly")
		return nil

	}); err != nil {
		return err
	}
	return nil
}

func defaultConnector(conn Connector) *connector {
	c := &connector{
		conn: conn,
		log:  slog.Default(),
	}
	return c
}

func WithLogger(log *slog.Logger) ConnectorOption {
	return func(c *connector) {
		if log != nil {
			c.log = log
		}
	}
}

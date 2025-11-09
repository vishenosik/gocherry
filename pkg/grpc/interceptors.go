package grpc

import (
	"context"
	"log/slog"
	"time"

	"github.com/vishenosik/gocherry/pkg/logs"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func WithCustomInterceptor(interceptors ...grpc.ServerOption) ServerOption {
	return func(srv *Server) {
		srv.interceptors = append(srv.interceptors, interceptors...)
	}
}

func WithLogInterceptors() ServerOption {
	return func(srv *Server) {
		srv.interceptors = append(srv.interceptors,
			// Unary
			grpc.UnaryInterceptor(LogUnaryRequest(srv.log)),
			// Stream
			grpc.StreamInterceptor(LogStreamRequest(srv.log)),
		)
	}
}

func LogUnaryRequest(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {

		timeStart := time.Now()
		resp, err := handler(ctx, req)

		if err != nil {
			st, _ := status.FromError(err)
			log.Error("request failed",
				slog.String("method", info.FullMethod),
				logs.Took(timeStart),
				slog.String("error", err.Error()),
				slog.Int("code", int(st.Code())),
			)
		} else {
			log.Info("request completed successfully",
				slog.String("method", info.FullMethod),
				logs.Took(timeStart),
			)
		}

		return resp, err
	}
}

func LogStreamRequest(log *slog.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		timeStart := time.Now()

		log.Info("stream started",
			slog.String("method", info.FullMethod),
		)

		err := handler(srv, ss)

		if err != nil {
			st, _ := status.FromError(err)
			log.Error("stream failed",
				slog.String("method", info.FullMethod),
				logs.Took(timeStart),
				slog.String("error", err.Error()),
				slog.Int("code", int(st.Code())),
			)
		} else {
			log.Info("stream completed",
				slog.String("method", info.FullMethod),
				logs.Took(timeStart),
			)
		}

		return err
	}
}

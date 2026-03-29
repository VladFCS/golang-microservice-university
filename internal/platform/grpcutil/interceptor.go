package grpcutil

import (
	"context"
	"log/slog"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func UnaryLoggingInterceptor(logger *slog.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		start := time.Now()
		clientAddr := ""
		if p, ok := peer.FromContext(ctx); ok && p.Addr != nil {
			clientAddr = p.Addr.String()
		}

		resp, err := handler(ctx, req)
		fields := []any{
			slog.String("grpc_method", info.FullMethod),
			slog.Duration("duration", time.Since(start)),
			slog.String("client_addr", clientAddr),
		}
		if err != nil {
			st := status.Convert(err)
			fields = append(fields, slog.String("code", st.Code().String()))
			logger.ErrorContext(ctx, "grpc request failed", fields...)
			return nil, err
		}

		logger.InfoContext(ctx, "grpc request completed", fields...)
		return resp, nil
	}
}

package logger

import (
	"google.golang.org/grpc"

	grpczerolog "github.com/grpc-ecosystem/go-grpc-middleware/providers/zerolog/v2"
	glogging "github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"github.com/rs/zerolog"
)

func UnaryServerInterceptor(logger zerolog.Logger) grpc.UnaryServerInterceptor {
	return glogging.UnaryServerInterceptor(grpczerolog.InterceptorLogger(logger))
}

func UnaryClientInterceptor(logger zerolog.Logger) grpc.UnaryClientInterceptor {
	return glogging.UnaryClientInterceptor(grpczerolog.InterceptorLogger(logger))
}

func StreamServerInterceptor(logger zerolog.Logger) grpc.StreamServerInterceptor {
	return glogging.StreamServerInterceptor(grpczerolog.InterceptorLogger(logger))
}

func StreamClientInterceptor(logger zerolog.Logger) grpc.StreamClientInterceptor {
	return glogging.StreamClientInterceptor(grpczerolog.InterceptorLogger(logger))
}

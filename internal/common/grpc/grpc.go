package grpc

import (
	"context"
	"log/slog"
	"net"
	"os"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewGRPCClientConn dials a gRPC server at addr. The returned connection is
// established lazily (on first RPC) and must be closed by the caller.
func NewGRPCClientConn(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

func RunGRPCServerOnAddr(addr string, registerServer func(server *grpc.Server)) {
	logger := slog.Default()

	grpcServer := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(interceptorLogger(logger)),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(interceptorLogger(logger)),
		),
	)
	registerServer(grpcServer)

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Error("failed to listen", "error", err, "addr", addr)
		os.Exit(1)
	}

	logger.Info("starting grpc server", "addr", addr)
	if err := grpcServer.Serve(listen); err != nil {
		logger.Error("grpc server stopped with error", "error", err)
		os.Exit(1)
	}
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

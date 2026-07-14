package grpc

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// NewGRPCClientConn dials a gRPC server at addr. The returned connection is
// established lazily (on first RPC) and must be closed by the caller.
func NewGRPCClientConn(addr string) (*grpc.ClientConn, error) {
	return grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
}

// NewGRPCServer builds a *grpc.Server with the standard logging
// interceptors wired in but not yet listening, so callers can register
// their services and control the run/shutdown lifecycle themselves (see
// commonHttp.NewEcho for the equivalent on the HTTP side).
func NewGRPCServer() *grpc.Server {
	logger := slog.Default()

	return grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			logging.UnaryServerInterceptor(interceptorLogger(logger)),
		),
		grpc.ChainStreamInterceptor(
			logging.StreamServerInterceptor(interceptorLogger(logger)),
		),
	)
}

// RunGRPCServerOnAddr listens on addr and serves server, blocking until it
// stops. A clean stop (server.GracefulStop/Stop) is reported as a nil
// error.
func RunGRPCServerOnAddr(server *grpc.Server, addr string) error {
	logger := slog.Default()

	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}

	logger.Info("starting grpc server", "addr", addr)
	if err := server.Serve(listen); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
		return fmt.Errorf("grpc server stopped with error: %w", err)
	}

	return nil
}

func interceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, level logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(level), msg, fields...)
	})
}

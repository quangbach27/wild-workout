package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	commonGrpc "workout/common/grpc"
	"workout/common/grpc/protobuf/trainer"
	"workout/common/grpc/protobuf/user"
	commonHttp "workout/common/http"
	"workout/common/log"
	"workout/training"
	grpcAdapter "workout/training/adapters/grpc"
	"workout/training/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Init(slog.LevelInfo)

	cfg := config.New()

	dbPgx, err := pgxpool.New(ctx, cfg.Database.DSN)
	if err != nil {
		panic(err)
	}

	trainerConn, err := commonGrpc.NewGRPCClientConn(cfg.App.TrainerGRPCAddress)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := trainerConn.Close(); err != nil {
			slog.Error("failed to close trainer gRPC connection", "error", err)
		}
	}()

	userConn, err := commonGrpc.NewGRPCClientConn(cfg.App.UserGRPCAddress)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := userConn.Close(); err != nil {
			slog.Error("failed to close user gRPC connection", "error", err)
		}
	}()

	trainerClient := trainer.NewTrainerServiceClient(trainerConn)
	userClient := user.NewUsersServiceClient(userConn)

	// TODO: swap StubAuthClient for a real *auth.Client once Firebase credentials are wired up.
	authClient := commonHttp.NewStubAuthClient("")

	externalService := training.ExternalServices{
		TrainerService: grpcAdapter.NewTrainerGrpc(trainerClient),
		UserService:    grpcAdapter.NewUsersGrpc(userClient),
		AuthClient:     authClient,
	}

	svc, err := training.NewService(ctx, dbPgx, externalService)
	if err != nil {
		panic(err)
	}

	if err := svc.Run(ctx, cfg.App); err != nil {
		panic(err)
	}
}

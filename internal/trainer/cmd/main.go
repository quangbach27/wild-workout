package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	commonHttp "workout/common/http"
	"workout/common/log"
	"workout/trainer"
	"workout/trainer/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Init(slog.LevelInfo)

	config := config.New()

	dbPgx, err := pgxpool.New(ctx, config.Database.DSN)
	if err != nil {
		panic(err)
	}

	// TODO: swap StubAuthClient for a real *auth.Client once Firebase credentials are wired up.
	externalServices := trainer.ExternalServices{
		AuthClient: commonHttp.NewStubAuthClient(""),
	}

	svc, err := trainer.New(ctx, dbPgx, externalServices)
	if err != nil {
		panic(err)
	}

	if err := svc.Run(ctx, config.App); err != nil {
		panic(err)
	}
}

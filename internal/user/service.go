package user

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"time"
	"workout/common"
	userPb "workout/common/grpc/protobuf/user"
	commonHttp "workout/common/http"
	"workout/common/log"
	"workout/user/adapters/db"
	"workout/user/app"
	"workout/user/config"
	portGrpc "workout/user/ports/grpc"
	portHttp "workout/user/ports/http"

	commonGrpc "workout/common/grpc"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
)

type ExternalServices struct {
	AuthClient commonHttp.AuthClient
}

type Service struct {
	echoRouter *echo.Echo
	grpcServer *grpc.Server

	pgxDb *pgxpool.Pool

	appHandler           *app.Handler
	userBalanceReadModel portGrpc.UserBalanceReadModel
}

func NewService(
	ctx context.Context,
	pgxDb *pgxpool.Pool,
	externalServices ExternalServices,
) (*Service, error) {
	e := commonHttp.NewEcho(externalServices.AuthClient)
	grpcServer := commonGrpc.NewGRPCServer()

	service := &Service{
		echoRouter: e,
		grpcServer: grpcServer,
		pgxDb:      pgxDb,
	}

	if err := service.init(ctx); err != nil {
		return nil, err
	}

	if err := service.registerHttp(); err != nil {
		return nil, err
	}

	if err := service.registerGrpc(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *Service) Run(
	ctx context.Context,
	appConfig config.App,
) error {
	defer s.pgxDb.Close()

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-gctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := s.echoRouter.Shutdown(shutdownCtx); err != nil {
			log.FromContext(ctx).Error("shutting down http server failed", "error", err)
		}
		s.grpcServer.GracefulStop()
		return nil
	})

	g.Go(func() error {
		if err := commonGrpc.RunGRPCServerOnAddr(s.grpcServer, appConfig.GRPCAddress); err != nil {
			return fmt.Errorf("grpc server stopped with error: %w", err)
		}
		return nil
	})

	g.Go(func() error {
		if err := s.echoRouter.Start(appConfig.HTTPAddress); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("starting http server failed: %w", err)
		}
		return nil
	})

	return g.Wait()
}

func (s *Service) name() string {
	return "user"
}

//go:embed adapters/db/migrations/*.sql
var embedMigrations embed.FS

func (s *Service) init(ctx context.Context) error {
	if err := common.MigrateDatabaseUp(
		ctx,
		s.name(),
		s.pgxDb,
		embedMigrations,
		"adapters/db/migrations",
	); err != nil {
		return err
	}

	userRepo := db.NewUserRepository(s.pgxDb)

	s.appHandler = app.NewHandler(userRepo)
	s.userBalanceReadModel = db.NewUserBalanceReadModel(s.pgxDb)

	return nil
}

func (s *Service) registerHttp() error {
	httpHandler := portHttp.NewHandler(s.appHandler)
	portHttp.Register(s.echoRouter, httpHandler)

	return nil
}

func (s *Service) registerGrpc() error {
	grpcHandler := portGrpc.NewHandler(s.appHandler, s.userBalanceReadModel)
	userPb.RegisterUsersServiceServer(s.grpcServer, grpcHandler)
	return nil
}

package trainer

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"time"
	"workout/common"
	commonGrpc "workout/common/grpc"
	trainerpb "workout/common/grpc/protobuf/trainer"
	commonHttp "workout/common/http"
	"workout/common/log"
	"workout/trainer/adapters/db"
	"workout/trainer/app/command"
	"workout/trainer/app/query"
	"workout/trainer/config"
	"workout/trainer/domain"
	portGrpc "workout/trainer/ports/grpc"
	portHttp "workout/trainer/ports/http"

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

	commandHandler *command.Handler
	queryHandler   *query.Handler
}

func New(
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
	return "trainer"
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

	hourFactoryCfg := domain.HourFactoryConfig{
		MaxWeeksInTheFutureToSet: 7,
		MinUtcHour:               8,
		MaxUtcHour:               17,
	}
	hourFactory := domain.MustNewHourFactory(hourFactoryCfg)
	hourReadModel := db.NewHourReadModel(s.pgxDb, hourFactoryCfg)
	hourRepo := db.NewHourRepository(s.pgxDb, hourFactory)

	s.commandHandler = command.NewHandler(hourRepo)
	s.queryHandler = query.NewHandler(hourReadModel, hourRepo)

	return nil
}

func (s *Service) registerHttp() error {
	return portHttp.Register(s.echoRouter, portHttp.NewHandler(s.commandHandler, s.queryHandler))
}

func (s *Service) registerGrpc() error {
	grpcHandler := portGrpc.NewHandler(s.commandHandler, s.queryHandler)
	trainerpb.RegisterTrainerServiceServer(s.grpcServer, grpcHandler)
	return nil
}

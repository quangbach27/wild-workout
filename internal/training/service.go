package training

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"net/http"
	"time"
	"workout/common"
	commonHttp "workout/common/http"
	"workout/common/log"
	dbAdapter "workout/training/adapters/db"
	"workout/training/app/command"
	"workout/training/app/query"
	"workout/training/config"
	portHttp "workout/training/ports/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"golang.org/x/sync/errgroup"
)

type ExternalServices struct {
	UserService    command.UserService
	TrainerService command.TrainerService
	AuthClient     commonHttp.AuthClient
}

type Service struct {
	echoRouter       *echo.Echo
	pgxDb            *pgxpool.Pool
	externalServices ExternalServices

	commandHandler *command.Handler
	queryHandler   *query.Handler
}

func NewService(
	ctx context.Context,
	pgxDb *pgxpool.Pool,
	externalServices ExternalServices,
) (*Service, error) {
	e := commonHttp.NewEcho(externalServices.AuthClient)
	if pgxDb == nil {
		return nil, errors.New("pgx can't be nil")
	}

	service := &Service{
		echoRouter:       e,
		pgxDb:            pgxDb,
		externalServices: externalServices,
	}

	if err := service.init(ctx); err != nil {
		return nil, err
	}

	if err := service.registerHttp(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *Service) Run(ctx context.Context, appConfig config.App) error {
	defer s.pgxDb.Close()

	g, gctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		<-gctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := s.echoRouter.Shutdown(shutdownCtx)
		if err != nil {
			log.FromContext(ctx).Error("shutting down http server failed")
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
	return "training"
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

	trainingRepo := dbAdapter.NewTraningRepository(s.pgxDb)
	trainingReadModel := dbAdapter.NewTrainingReadModel(s.pgxDb)

	s.commandHandler = command.NewHandler(trainingRepo, s.externalServices.UserService, s.externalServices.TrainerService)
	s.queryHandler = query.NewHandler(trainingReadModel)

	return nil
}

func (s *Service) registerHttp() error {
	handler := portHttp.NewHandler(s.commandHandler, s.queryHandler)
	return portHttp.Register(s.echoRouter, handler)
}

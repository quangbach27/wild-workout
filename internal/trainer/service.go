package trainer

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
	"workout/trainer/adapters/db"
	"workout/trainer/app/command"
	"workout/trainer/app/query"
	"workout/trainer/config"
	"workout/trainer/domain"
	portHttp "workout/trainer/ports/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
)

type Service struct {
	echoRouter *echo.Echo

	pgxDb *pgxpool.Pool

	commandHandler *command.Handler
	queryHandler   *query.Handler
}

func New(
	ctx context.Context,
	pgxDb *pgxpool.Pool,
) (*Service, error) {
	e := commonHttp.NewEcho()

	service := &Service{
		echoRouter: e,
		pgxDb:      pgxDb,
	}

	if err := service.init(ctx); err != nil {
		return nil, err
	}

	if err := service.registerHttp(); err != nil {
		return nil, err
	}

	return service, nil
}

func (s *Service) Run(
	ctx context.Context,
	appConfig config.App,
) error {
	defer s.pgxDb.Close()
	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		err := s.echoRouter.Shutdown(shutdownCtx)
		if err != nil {
			log.FromContext(ctx).Error("shutting down http server failed")
		}
	}()

	s.echoRouter.Server.WriteTimeout = 30 * time.Second
	s.echoRouter.Server.ReadHeaderTimeout = 30 * time.Second
	s.echoRouter.Server.ReadTimeout = 30 * time.Second
	s.echoRouter.Server.IdleTimeout = 60 * time.Second

	err := s.echoRouter.Start(appConfig.HTTPAddress)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("starting http server failed: %w", err)
	}

	return nil
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
	s.queryHandler = query.NewHandler(hourReadModel)

	return nil
}

func (s *Service) registerHttp() error {
	portHttp.Register(s.echoRouter, portHttp.NewHandler(s.commandHandler, s.queryHandler))
	return nil
}

func (s *Service) registerGrpc() error {
	return nil
}

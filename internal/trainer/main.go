package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/joho/godotenv"
	"github.com/quangbach27/wild-workout/internal/common/logs"
	"github.com/quangbach27/wild-workout/internal/common/server"
	"github.com/quangbach27/wild-workout/internal/trainer/adapter"
	"github.com/quangbach27/wild-workout/internal/trainer/app"
	"github.com/quangbach27/wild-workout/internal/trainer/app/command"
	"github.com/quangbach27/wild-workout/internal/trainer/app/query"
	"github.com/quangbach27/wild-workout/internal/trainer/domain/hour"
	"github.com/quangbach27/wild-workout/internal/trainer/ports"
)

func main() {
	// Load .env file
	err := godotenv.Load("../../.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	logs.Init()
	ctx := context.Background()
	application := newApplication(ctx)
	server.RunHTTPServer(func(router chi.Router) http.Handler {
		return ports.HandlerFromMux(
			ports.NewHttpServer(application),
			router,
		)
	})

}

func newApplication(ctx context.Context) app.Application {

	factoryConfig := hour.FactoryConfig{
		MaxWeeksInTheFutureToSet: 6,
		MinUtcHour:               12,
		MaxUtcHour:               20,
	}

	hourFactory := hour.MustNewFactory(factoryConfig)
	sqlDB := adapter.MustNewPostgresSQLConnection()
	hourRepo := adapter.NewPGHourRepository(sqlDB, hourFactory)

	return app.Application{
		Queries: app.Queries{
			TrainerAvailableHours: query.NewAvailableHourHandler(nil),
			HourAvailability:      query.NewHourAvailabilityHandler(hourRepo),
		},
		Commands: app.Commands{
			CancelTraining:       command.NewCancelTrainingHandler(hourRepo),
			ScheduleTraining:     command.NewScheduleTrainingHandler(hourRepo),
			MakeHoursAvailable:   command.NewMakeHoursAvailableHandler(hourRepo),
			MakeHoursUnavailable: command.NewMakeHoursUnavailableHandler(hourRepo),
		},
	}
}

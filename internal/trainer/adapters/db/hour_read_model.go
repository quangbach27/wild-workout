package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"workout/trainer/adapters/db/dbmodels"
	"workout/trainer/app/query"
	"workout/trainer/domain"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type HourReadModel struct {
	db            *pgxpool.Pool
	factoryConfig domain.HourFactoryConfig
}

func NewHourReadModel(
	db *pgxpool.Pool,
	factoryConfig domain.HourFactoryConfig,
) *HourReadModel {
	if db == nil {
		panic("db can't be nil")
	}

	if err := factoryConfig.Validate(); err != nil {
		panic(err)
	}

	return &HourReadModel{
		db:            db,
		factoryConfig: factoryConfig,
	}
}

func (r *HourReadModel) ListAvailableHours(ctx context.Context, from time.Time, to time.Time) ([]query.Date, error) {
	queries := dbmodels.New(r.db)

	dbHours, err := queries.ListHours(
		ctx,
		dbmodels.ListHoursParams{
			FromTime: from,
			ToTime:   to,
		},
	)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to list available hours: %w", err)
	}

	return r.buildDateQuery(dbHours, from, to), nil
}

func (r *HourReadModel) buildDateQuery(
	dbHours []dbmodels.TrainerHour,
	from time.Time,
	to time.Time,
) []query.Date {
	start := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, time.UTC)
	end := time.Date(to.Year(), to.Month(), to.Day(), 0, 0, 0, 0, time.UTC)

	numDays := int(end.Sub(start).Hours()/24) + 1
	dates := make([]query.Date, 0, numDays)

	idx := 0 

	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		hourCount := r.factoryConfig.MaxUtcHour - r.factoryConfig.MinUtcHour + 1
		hours := make([]query.Hour, 0, hourCount)
		hasFreeHour := false

		hourStart := time.Date(d.Year(), d.Month(), d.Day(), r.factoryConfig.MinUtcHour, 0, 0, 0, time.UTC)

		for h := 0; h < hourCount; h++ {
			t := hourStart.Add(time.Duration(h) * time.Hour)

			// advance idx past any stale entries (shouldn't normally happen, but safe)
			for idx < len(dbHours) && dbHours[idx].Hour.Before(t) {
				idx++
			}

			var queryHour query.Hour
			if idx < len(dbHours) && dbHours[idx].Hour.Equal(t) {
				hour := domain.UnmarshalHour(dbHours[idx].Hour, dbHours[idx].Availability)
				queryHour = query.Hour{
					Hour:                 hour.Time(),
					Available:            hour.IsAvailable(),
					HasTrainingScheduled: hour.HasTrainingScheduled(),
				}
				idx++
			} else {
				queryHour = query.Hour{
					Hour:                 t,
					Available:            false,
					HasTrainingScheduled: false,
				}
			}

			if queryHour.Available {
				hasFreeHour = true
			}
			hours = append(hours, queryHour)
		}

		dates = append(dates, query.Date{
			Date:         d,
			HasFreeHours: hasFreeHour,
			Hours:        hours,
		})
	}

	return dates
}

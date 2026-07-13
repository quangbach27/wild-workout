package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"workout/common"
	"workout/trainer/adapters/db/dbmodels"
	"workout/trainer/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

type HourRepo struct {
	db          *pgxpool.Pool
	hourFactory domain.HourFactory
}

func NewHourRepository(
	db *pgxpool.Pool,
	hourFactory domain.HourFactory,
) *HourRepo {
	if db == nil {
		panic("db can't be empty")
	}

	if hourFactory.IsZero() {
		panic("hour factory can't be empty")
	}

	return &HourRepo{db: db, hourFactory: hourFactory}
}

func (r HourRepo) GetHour(ctx context.Context, hour time.Time) (*domain.Hour, error) {
	queries := dbmodels.New(r.db)

	return r.getHour(ctx, hour, queries)

}

func (r HourRepo) getHour(
	ctx context.Context,
	hourTime time.Time,
	queries *dbmodels.Queries,
) (*domain.Hour, error) {
	dbHour, err := queries.GetHour(ctx, hourTime)
	if err != nil {
		// If the hour not exist in db, that mean the hour is not available
		if errors.Is(err, sql.ErrNoRows) {
			return r.hourFactory.NewNotAvailableHour(hourTime)
		}

		return nil, fmt.Errorf("failed to get db hour: %w", err)
	}

	return domain.UnmarshalHour(dbHour.Hour, dbHour.Availability), nil
}

func (r HourRepo) UpdateHour(
	ctx context.Context,
	hourTime time.Time,
	updateFn func(h *domain.Hour) (*domain.Hour, error),
) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)
		hour, err := r.getHour(ctx, hourTime, queries)
		if err != nil {
			return err
		}

		updatedHour, err := updateFn(hour)
		if err != nil {
			return fmt.Errorf("failed to update hour: %w", err)
		}

		err = queries.UpsertHour(
			ctx, dbmodels.UpsertHourParams{
				Hour:         updatedHour.Time(),
				Availability: updatedHour.Availability(),
			},
		)
		if err != nil {
			return fmt.Errorf("failed to upsert hour: %w", err)
		}

		return nil
	})
}

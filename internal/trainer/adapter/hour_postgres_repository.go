package adapter

import (
	"context"
	"database/sql"

	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"github.com/quangbach27/wild-workout/internal/trainer/domain/hour"
	"go.uber.org/multierr"
)

type DateModel struct {
	ID           int32       `db:"id"`
	Date         time.Time   `db:date`
	HasFreeHours bool        `db:has_free_hours`
	HourModels   []HourModel `db:-`
}

type HourModel struct {
	ID           int32     `db:"id"`
	Hour         time.Time `db:"hour"`
	Availability string    `db:"availability"`
}

type PGHourRepository struct {
	db          *sqlx.DB
	hourFactory hour.Factory
}

func NewPGHourRepository(db *sqlx.DB, hourFactory hour.Factory) PGHourRepository {
	return PGHourRepository{db, hourFactory}
}

// sqlContextGetter is an interface provided both by transaction and standard db connection
type sqlContextGetter interface {
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

func (repo PGHourRepository) GetHour(ctx context.Context, hourTime time.Time) (*hour.Hour, error) {
	return repo.getOrCreateHour(ctx, repo.db, hourTime, false)
}

func (repo PGHourRepository) getOrCreateHour(
	ctx context.Context,
	db sqlContextGetter,
	hourTime time.Time,
	forUpdate bool,
) (*hour.Hour, error) {
	hourModel := HourModel{}
	query := "SELECT * FROM `hours` WHERE `hour` = ?"
	if forUpdate {
		query += " FOR UPDATE"
	}

	err := db.GetContext(ctx, &hourModel, query, hourTime.UTC())
	if errors.Is(err, sql.ErrNoRows) {
		return repo.hourFactory.NewNotAvailableHour(hourTime)
	} else if err != nil {
		return nil, errors.Wrap(err, "unable to get hour from db")
	}

	availability, err := hour.NewAvailabilityFromString(hourModel.Availability)
	if err != nil {
		return nil, err
	}

	domainHour, err := repo.hourFactory.UnmarshalHourFromDatabase(hourModel.Hour.Local(), availability)
	if err != nil {
		return nil, err
	}

	return domainHour, nil
}

const mySQLDeadlockErrorCode = 1213

func (repo PGHourRepository) UpdateHour(
	ctx context.Context,
	hourTime time.Time,
	updateFn func(h *hour.Hour) (*hour.Hour, error),
) error {
	return repo.updateHour(ctx, hourTime, updateFn)
}

func (repo PGHourRepository) updateHour(
	ctx context.Context,
	hourTime time.Time,
	updateFn func(h *hour.Hour) (*hour.Hour, error),
) (err error) {
	tx, err := repo.db.Beginx()
	if err != nil {
		return errors.Wrap(err, "unable to start transaction")
	}

	// Defer is executed on function just before exit.
	// With defer, we are always sure that we will close our transaction properly.
	defer func() {
		// In `UpdateHour` we are using named return - `(err error)`.
		// Thanks to that, that can check if function exits with error.
		//
		// Even if function exits without error, commit still can return error.
		// In that case we can override nil to err `err = m.finish...`.
		err = repo.finishTransaction(err, tx)
	}()

	existingHour, err := repo.getOrCreateHour(ctx, tx, hourTime, true)
	if err != nil {
		return
	}

	updatedHour, err := updateFn(existingHour)
	if err != nil {
		return
	}

	if err := repo.upsertHour(tx, updatedHour); err != nil {
		return err
	}

	return nil
}

// upsertHour updates hour if hour already exists in the database.
// If your doesn't exists, it's inserted.
func (repo PGHourRepository) upsertHour(tx *sqlx.Tx, hourToUpdate *hour.Hour) error {
	updatedDbHour := HourModel{
		Hour:         hourToUpdate.Time().UTC(),
		Availability: hourToUpdate.Availability().String(),
	}

	_, err := tx.NamedExec(
		`INSERT INTO 
			hours (hour, availability) 
		VALUES 
			(:hour, :availability)
		ON CONFLICT (hour) DO UPDATE
		SET	availability = :availability`,
		updatedDbHour,
	)
	if err != nil {
		return errors.Wrap(err, "unable to upsert hour")
	}

	return nil
}

func (repo PGHourRepository) finishTransaction(err error, tx *sqlx.Tx) error {
	if err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			return multierr.Combine(err, rollbackErr)
		}

		return err
	} else {
		if commitErr := tx.Commit(); commitErr != nil {
			return errors.Wrap(err, "failed to commit tx")
		}

		return nil
	}
}

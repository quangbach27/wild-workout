package db

import (
	"context"
	"fmt"
	"time"
	"workout/common"
	"workout/training/adapters/db/dbmodels"
	"workout/training/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TrainingRepo struct {
	db *pgxpool.Pool
}

func NewTraningRepository(db *pgxpool.Pool) *TrainingRepo {
	if db == nil {
		panic("db can't be nil")
	}

	return &TrainingRepo{
		db: db,
	}
}
func (r *TrainingRepo) AddTraining(ctx context.Context, tr *domain.Training) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		return queries.CreateTraining(ctx, dbmodels.CreateTrainingParams{
			TrainingUuid: tr.UUID(),
			UserUuid:     tr.UserUUID(),
			Username:     tr.UserName(),
			Time:         tr.Time(),
			Notes:        common.ToPtr(tr.Notes()),
			Canceled:     tr.IsCanceled(),
		})
	})
}

func (r *TrainingRepo) GetTraining(
	ctx context.Context,
	trainingUUID domain.TrainingUUID,
	user domain.User,
) (*domain.Training, error) {
	queries := dbmodels.New(r.db)
	training, err := r.getTraining(ctx, trainingUUID, queries)
	if err != nil {
		return nil, err
	}

	err = domain.CanUserSeeTraining(user, training)
	if err != nil {
		return nil, common.NewForbiddenError("forbidden", "%s", err.Error())
	}

	return training, nil
}

func (r *TrainingRepo) getTraining(
	ctx context.Context,
	trainingUUID domain.TrainingUUID,
	queries *dbmodels.Queries,
) (*domain.Training, error) {
	dbTraining, err := queries.GetTraining(ctx, trainingUUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get training: %w", err)
	}

	return unmarshalTrainingDbToDomain(dbTraining), nil
}

func unmarshalTrainingDbToDomain(dbTraining dbmodels.TrainingTraining) *domain.Training {
	return domain.UnmarshalTrainingFromDB(
		dbTraining.TrainingUuid,
		dbTraining.UserUuid,
		dbTraining.Username,
		dbTraining.Time,
		common.SafeDeref(dbTraining.Notes, ""),
		common.SafeDeref(dbTraining.ProposedTime, time.Time{}),
		common.SafeDeref(dbTraining.MoveProposedBy, domain.UserType{}),
		dbTraining.Canceled,
	)
}

func (r *TrainingRepo) UpdateTraining(
	ctx context.Context,
	trainingUUID domain.TrainingUUID,
	user domain.User,
	updateFn func(ctx context.Context, tr *domain.Training) (*domain.Training, error),
) error {
	return common.UpdateInTx(ctx, r.db, func(ctx context.Context, tx pgx.Tx) error {
		queries := dbmodels.New(tx)

		training, err := r.getTraining(ctx, trainingUUID, queries)
		if err != nil {
			return err
		}

		if err = domain.CanUserSeeTraining(user, training); err != nil {
			return common.NewForbiddenError("forbidden", "%s", err.Error())
		}

		updatedTraining, err := updateFn(ctx, training)
		if err != nil {
			return err
		}

		return queries.UpdateTraining(ctx, marshalTrainingDomainToDbParams(updatedTraining))
	})
}

func marshalTrainingDomainToDbParams(tr *domain.Training) dbmodels.UpdateTrainingParams {
	params := dbmodels.UpdateTrainingParams{
		TrainingUuid: tr.UUID(),
		Time:         tr.Time(),
		Notes:        common.ToPtr(tr.Notes()),
		Canceled:     tr.IsCanceled(),
	}

	if proposedNewTime := tr.ProposedNewTime(); !proposedNewTime.IsZero() {
		params.ProposedTime = common.ToPtr(proposedNewTime)
	}

	if moveProposedBy := tr.MoveProposedBy(); !moveProposedBy.IsZero() {
		params.MoveProposedBy = common.ToPtr(moveProposedBy)
	}

	return params
}

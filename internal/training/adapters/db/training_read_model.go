package db

import (
	"context"
	"workout/training/adapters/db/dbmodels"
	"workout/training/app/query"
	"workout/training/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TrainingReadModel struct {
	db *pgxpool.Pool
}

func NewTrainingReadModel(db *pgxpool.Pool) *TrainingReadModel {
	if db == nil {
		panic("db can't be nil")
	}

	return &TrainingReadModel{
		db: db,
	}
}

func (r *TrainingReadModel) FindTrainingsForUser(ctx context.Context, userID domain.UserID) ([]query.Training, error) {
	queries := dbmodels.New(r.db)

	dbTrainings, err := queries.FindTrainingsForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return mapTrainingsDbToQuery(dbTrainings), nil
}

func (r *TrainingReadModel) ListAllTrainings(ctx context.Context) ([]query.Training, error) {
	queries := dbmodels.New(r.db)

	dbTrainings, err := queries.ListAllTrainings(ctx)
	if err != nil {
		return nil, err
	}

	return mapTrainingsDbToQuery(dbTrainings), nil
}

func mapTrainingsDbToQuery(dbTrainings []dbmodels.TrainingTraining) []query.Training {
	trainings := make([]query.Training, 0, len(dbTrainings))
	for _, dbTraining := range dbTrainings {
		trainings = append(trainings, mapTrainingDbToQuery(dbTraining))
	}

	return trainings
}

func mapTrainingDbToQuery(dbTraining dbmodels.TrainingTraining) query.Training {
	tr := unmarshalTrainingDbToDomain(dbTraining)

	var moveProposedBy *string
	if dbTraining.MoveProposedBy != nil {
		s := dbTraining.MoveProposedBy.String()
		moveProposedBy = &s
	}

	return query.Training{
		UUID:   tr.UUID(),
		UserID: tr.UserID(),
		User:   tr.UserName(),

		Time:  tr.Time(),
		Notes: tr.Notes(),

		ProposedTime:   dbTraining.ProposedTime,
		MoveProposedBy: moveProposedBy,

		CanBeCancelled: tr.CanBeCanceledForFree(),
	}
}

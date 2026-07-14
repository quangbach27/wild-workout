-- name: CreateTraining :exec
INSERT INTO training.trainings (
    training_uuid,
    user_id,
    username,
    time,
    notes,
    canceled
) VALUES (
    $1, $2, $3, $4, $5, $6
);

-- name: GetTraining :one
SELECT *
FROM training.trainings
WHERE training_uuid = $1;

-- name: UpdateTraining :exec
UPDATE training.trainings
SET
    time = $2,
    notes = $3,
    proposed_time = $4,
    move_proposed_by = $5,
    canceled = $6
WHERE training_uuid = $1;

-- name: ListAllTrainings :many
SELECT *
FROM training.trainings
WHERE canceled = false
ORDER BY time ASC;

-- name: FindTrainingsForUser :many
SELECT *
FROM training.trainings
WHERE user_id = $1 AND canceled = false
ORDER BY time ASC;

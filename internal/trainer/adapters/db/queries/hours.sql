-- name: GetHour :one
SELECT * FROM trainer.hours 
WHERE hour = $1;

-- name: UpsertHour :exec
INSERT INTO trainer.hours (hour, availability)
VALUES ($1, $2)
ON CONFLICT (hour) DO UPDATE
    SET availability = EXCLUDED.availability;
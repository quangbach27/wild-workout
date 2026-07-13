-- name: GetHour :one
SELECT * FROM trainer.hours 
WHERE hour = $1;

-- name: UpsertHour :exec
INSERT INTO trainer.hours (hour, availability)
VALUES ($1, $2)
ON CONFLICT (hour) DO UPDATE
    SET availability = EXCLUDED.availability;

-- name: ListHours :many
SELECT hour, availability
FROM trainer.hours
WHERE hour >= sqlc.arg(from_time)::timestamptz
  AND hour <=  sqlc.arg(to_time)::timestamptz
ORDER BY hour;
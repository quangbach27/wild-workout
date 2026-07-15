-- name: CreateUser :one
INSERT INTO "user".users (
    user_uuid,
    firebase_uid,
    username,
    role,
    balance
) VALUES (
    $1, $2, $3, $4, $5
)
RETURNING *;

-- name: GetUser :one
SELECT *
FROM "user".users
WHERE user_uuid = $1;

-- name: GetUserByFirebaseUID :one
SELECT *
FROM "user".users
WHERE firebase_uid = $1;

-- name: UpdateBalance :exec
UPDATE "user".users
SET balance = $2
WHERE firebase_uid = $1;

-- name: GetUserBalance :one
SELECT balance
FROM "user".users
WHERE firebase_uid = $1;

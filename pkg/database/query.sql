-- name: GetAllAccounts :many
SELECT *
FROM account
ORDER BY id;

-- name: GetAccount :one
SELECT *
FROM account
WHERE id = $1
LIMIT 1;

-- name: GetAccountByToken :one
SELECT *
FROM account
WHERE token = $1
LIMIT 1;

-- name: GetAccountByChatID :one
SELECT *
FROM account
WHERE chat_id = $1
LIMIT 1;

-- name: CreateAccount :one
INSERT INTO account (token, name, chat_id, status)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: UpdateAccountStatus :exec
UPDATE account
SET status = $2
WHERE id = $1;

-- name: GetLastUpdateByAccountID :many
SELECT *
FROM updates
WHERE account_id = $1
ORDER BY id DESC
LIMIT 1;

-- name: GetLatestUpdatesByAccountID :many
SELECT *
FROM updates
WHERE account_id = $1
ORDER BY id DESC
LIMIT 10;

-- name: CreateUpdate :one
INSERT INTO updates (account_id, created, data)
VALUES ($1, $2, $3)
RETURNING id;

-- name: GetAllFences :many
SELECT *
FROM fence;

-- name: CreateFence :one
INSERT INTO fence (name, longitude, latitude, radius)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: DeleteFence :exec
DELETE
FROM fence
WHERE id = $1;

-- name: GetMigrations :many
SELECT *
FROM migration
ORDER BY id;

-- name: CreateMigration :one
INSERT INTO migration (id, applied)
VALUES ($1, $2)
RETURNING id;

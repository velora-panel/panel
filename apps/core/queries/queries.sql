-- name: CreateUser :one
INSERT INTO users (email, password_hash, role) VALUES ($1,$2,$3) RETURNING *;
-- name: UserByEmail :one
SELECT * FROM users WHERE email=$1;
-- name: CreateSession :one
WITH account AS (SELECT id FROM users WHERE id=$2 AND password_hash=$4 AND NOT disabled FOR UPDATE)
INSERT INTO sessions (token_hash,user_id,expires_at) SELECT $1,id,$3 FROM account RETURNING token_hash;
-- name: SessionUser :one
SELECT users.* FROM users JOIN sessions ON sessions.user_id=users.id WHERE sessions.token_hash=$1 AND sessions.expires_at > now() AND NOT users.disabled;
-- name: DeleteSession :exec
DELETE FROM sessions WHERE token_hash=$1;
-- name: ListNodes :many
SELECT * FROM nodes ORDER BY created_at DESC, id DESC LIMIT 100;
-- name: NodeByID :one
SELECT * FROM nodes WHERE id=$1;
-- name: CreateNode :one
INSERT INTO nodes (name, description) VALUES ($1,$2) RETURNING *;
-- name: ListUsers :many
SELECT * FROM users ORDER BY created_at, id LIMIT 200;
-- name: UserByID :one
SELECT * FROM users WHERE id=$1;
-- name: CountActiveAdmins :one
SELECT count(*) FROM users WHERE role='admin' AND NOT disabled;
-- name: LockAccounts :exec
SELECT pg_advisory_xact_lock(76102941);
-- name: UpdateAccount :one
UPDATE users SET role=$2, disabled=$3 WHERE id=$1 RETURNING *;
-- name: UpdateProfile :one
UPDATE users SET email=$2 WHERE id=$1 AND password_hash=$3 AND NOT disabled RETURNING *;
-- name: UpdatePassword :one
UPDATE users SET password_hash=$2 WHERE id=$1 AND password_hash=$3 AND NOT disabled RETURNING *;
-- name: RevokeSessions :exec
DELETE FROM sessions WHERE user_id=$1 AND token_hash<>$2;
-- name: DeleteResets :exec
DELETE FROM password_resets WHERE user_id=$1;
-- name: CreateReset :exec
INSERT INTO password_resets(token_hash,user_id,expires_at) VALUES ($1,$2,$3);
-- name: ConsumeReset :one
DELETE FROM password_resets WHERE token_hash=$1 AND expires_at>now() RETURNING user_id;
-- name: ResetPassword :one
UPDATE users SET password_hash=$2 WHERE id=$1 AND NOT disabled RETURNING *;
-- name: SchemaInitialized :one
SELECT to_regclass('users') IS NOT NULL AND to_regclass('nodes') IS NOT NULL AND to_regclass('password_resets') IS NOT NULL AS initialized;

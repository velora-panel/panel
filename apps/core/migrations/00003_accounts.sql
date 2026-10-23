-- +goose Up
ALTER TABLE users ADD COLUMN disabled boolean NOT NULL DEFAULT false;
CREATE TABLE password_resets (
 token_hash text PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 expires_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX password_resets_user_id_idx ON password_resets(user_id);
CREATE INDEX password_resets_expires_at_idx ON password_resets(expires_at);
-- +goose Down
DROP TABLE password_resets;
ALTER TABLE users DROP COLUMN disabled;

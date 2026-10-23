-- +goose Up
CREATE TABLE users (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 email text NOT NULL UNIQUE CHECK (email = lower(email)),
 password_hash text NOT NULL,
 role text NOT NULL CHECK (role IN ('admin','user')),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE sessions (
 token_hash text PRIMARY KEY,
 user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 expires_at timestamptz NOT NULL,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX sessions_user_id_idx ON sessions(user_id);
CREATE INDEX sessions_expires_at_idx ON sessions(expires_at);
CREATE TABLE stations (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 name text NOT NULL UNIQUE CHECK (length(name) BETWEEN 1 AND 80),
 description text NOT NULL DEFAULT '' CHECK (length(description) <= 500),
 status text NOT NULL DEFAULT 'unconfigured' CHECK (status = 'unconfigured'),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE blueprints (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 name text NOT NULL UNIQUE,
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE TABLE instances (
 id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
 name text NOT NULL,
 station_id uuid NOT NULL REFERENCES stations(id),
 blueprint_id uuid REFERENCES blueprints(id),
 created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX instances_station_id_idx ON instances(station_id);
CREATE INDEX instances_blueprint_id_idx ON instances(blueprint_id);
-- +goose Down
DROP TABLE instances, blueprints, stations, sessions, users;

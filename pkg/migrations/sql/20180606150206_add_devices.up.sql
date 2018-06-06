CREATE TYPE disposition AS ENUM ('indoor', 'outdoor');

CREATE TABLE IF NOT EXISTS devices (
  id SERIAL PRIMARY KEY,
  token TEXT NOT NULL,
  user_id INTEGER NOT NULL REFERENCES users(id),
  private_key BYTEA NOT NULL,
  public_key TEXT NOT NULL,
  longitude DOUBLE PRECISION NOT NULL,
  latitude DOUBLE PRECISION NOT NULL,
  disposition disposition NOT NULL,
  created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE UNIQUE INDEX IF NOT EXISTS devices_token_idx
  ON devices(token);

CREATE INDEX IF NOT EXISTS devices_user_id_idx
  ON devices(user_id);
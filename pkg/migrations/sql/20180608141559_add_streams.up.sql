CREATE TABLE IF NOT EXISTS streams (
  id SERIAL PRIMARY KEY,
  device_id INTEGER NOT NULL REFERENCES devices(id),
  uid TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS streams_device_id_idx
  ON streams(device_id);

CREATE UNIQUE INDEX IF NOT EXISTS streams_uid_idx
  ON streams(uid);
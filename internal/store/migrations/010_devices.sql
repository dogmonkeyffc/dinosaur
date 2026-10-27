ALTER TABLE sessions ADD COLUMN kind TEXT NOT NULL DEFAULT 'browser' CHECK(kind IN ('browser','device'));
ALTER TABLE sessions ADD COLUMN device_id TEXT;
ALTER TABLE sessions ADD COLUMN device_name TEXT NOT NULL DEFAULT '';
CREATE UNIQUE INDEX sessions_device_id ON sessions(device_id) WHERE device_id IS NOT NULL;

CREATE TABLE device_grants (
 id TEXT PRIMARY KEY,
 secret_hash TEXT NOT NULL UNIQUE,
 kind TEXT NOT NULL CHECK(kind IN ('mobile','tv')),
 code TEXT UNIQUE,
 name TEXT NOT NULL,
 owner_hash TEXT REFERENCES sessions(token_hash) ON DELETE CASCADE,
 expires_at INTEGER NOT NULL
);
CREATE INDEX device_grants_expiry ON device_grants(expires_at);

CREATE TABLE server_identity (id TEXT NOT NULL);
INSERT INTO server_identity VALUES (lower(hex(randomblob(16))));

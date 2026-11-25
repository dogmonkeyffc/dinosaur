CREATE TABLE mobile_admin_grants (
 secret_hash TEXT PRIMARY KEY,
 parent_hash TEXT NOT NULL REFERENCES sessions(token_hash) ON DELETE CASCADE,
 expires_at INTEGER NOT NULL
);
CREATE UNIQUE INDEX mobile_admin_grants_parent ON mobile_admin_grants(parent_hash);
CREATE TABLE mobile_admin_sessions (
 token_hash TEXT PRIMARY KEY REFERENCES sessions(token_hash) ON DELETE CASCADE,
 parent_hash TEXT NOT NULL REFERENCES sessions(token_hash) ON DELETE CASCADE
);
CREATE INDEX mobile_admin_sessions_parent ON mobile_admin_sessions(parent_hash);
CREATE TRIGGER revoke_mobile_admin_sessions BEFORE DELETE ON sessions BEGIN
 DELETE FROM sessions WHERE token_hash IN (
  SELECT token_hash FROM mobile_admin_sessions WHERE parent_hash=OLD.token_hash
 );
END;

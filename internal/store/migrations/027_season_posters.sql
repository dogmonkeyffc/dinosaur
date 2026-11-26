CREATE TABLE season_posters (
 tmdb_id INTEGER NOT NULL,
 season_number INTEGER NOT NULL,
 poster_key TEXT NOT NULL DEFAULT '',
 retry_at INTEGER NOT NULL DEFAULT 0,
 PRIMARY KEY(tmdb_id,season_number)
);

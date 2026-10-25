CREATE TABLE tv_series (
 id INTEGER PRIMARY KEY,
 library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
 source_key TEXT NOT NULL,
 title TEXT NOT NULL,
 year INTEGER NOT NULL DEFAULT 0,
 metadata_json TEXT NOT NULL DEFAULT '{}',
 poster_key TEXT NOT NULL DEFAULT '',
 aliases TEXT NOT NULL DEFAULT '',
 UNIQUE(library_id,source_key)
);
CREATE TABLE tv_episodes (
 media_id INTEGER PRIMARY KEY REFERENCES media_items(id) ON DELETE CASCADE,
 series_id INTEGER NOT NULL REFERENCES tv_series(id) ON DELETE CASCADE,
 season_number INTEGER NOT NULL CHECK(season_number>=0),
 episode_number INTEGER NOT NULL CHECK(episode_number>0),
 episode_end INTEGER NOT NULL CHECK(episode_end>=episode_number)
);
CREATE INDEX tv_episode_order ON tv_episodes(series_id,season_number,episode_number,media_id);
ALTER TABLE scrape_jobs ADD COLUMN media_type TEXT NOT NULL DEFAULT 'movie';
ALTER TABLE scrape_job_items ADD COLUMN series_id INTEGER REFERENCES tv_series(id) ON DELETE SET NULL;

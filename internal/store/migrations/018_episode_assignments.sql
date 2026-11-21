CREATE TABLE episode_assignments (
 library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
 relative_path TEXT NOT NULL,
 series_folder TEXT NOT NULL,
 season_number INTEGER NOT NULL CHECK(season_number BETWEEN 0 AND 999),
 episode_number INTEGER NOT NULL CHECK(episode_number BETWEEN 1 AND 9999),
 PRIMARY KEY(library_id, relative_path)
);

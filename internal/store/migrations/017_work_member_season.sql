ALTER TABLE work_members ADD COLUMN season_number INTEGER NOT NULL DEFAULT 0
    CHECK (season_number BETWEEN 0 AND 999);

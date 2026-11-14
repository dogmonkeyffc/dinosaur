CREATE TABLE works (
 id INTEGER PRIMARY KEY,
 title TEXT NOT NULL CHECK(length(trim(title)) > 0),
 library_id INTEGER NOT NULL REFERENCES libraries(id) ON DELETE CASCADE,
 UNIQUE(id,library_id)
);
CREATE TABLE work_members (
 work_id INTEGER NOT NULL REFERENCES works(id) ON DELETE CASCADE,
 series_id INTEGER NOT NULL UNIQUE REFERENCES tv_series(id) ON DELETE CASCADE,
 position INTEGER NOT NULL,
 PRIMARY KEY(work_id,series_id),
 UNIQUE(work_id,position)
);
CREATE TRIGGER work_member_library_insert BEFORE INSERT ON work_members
WHEN (SELECT library_id FROM works WHERE id=NEW.work_id) !=
     (SELECT library_id FROM tv_series WHERE id=NEW.series_id)
BEGIN SELECT RAISE(ABORT,'cross-library work member'); END;

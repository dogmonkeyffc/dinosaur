CREATE TABLE watch_history (
 id INTEGER PRIMARY KEY AUTOINCREMENT,
 user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
 media_id INTEGER REFERENCES media_items(id) ON DELETE SET NULL,
 title TEXT NOT NULL,
 original_title TEXT NOT NULL DEFAULT '',
 year INTEGER NOT NULL DEFAULT 0,
 watched_on TEXT NOT NULL,
 rating INTEGER NOT NULL DEFAULT 0 CHECK(rating BETWEEN 0 AND 5),
 review TEXT NOT NULL DEFAULT '',
 source TEXT NOT NULL CHECK(source IN ('manual','douban','local')),
 source_id TEXT NOT NULL,
 poster_key TEXT NOT NULL DEFAULT '',
 created_at INTEGER NOT NULL,
 updated_at INTEGER NOT NULL,
 UNIQUE(user_id,source,source_id,watched_on)
);
CREATE INDEX watch_history_timeline ON watch_history(user_id,watched_on DESC,id DESC);
INSERT INTO watch_history(user_id,media_id,title,year,watched_on,source,source_id,created_at,updated_at)
 SELECT p.user_id,m.id,COALESCE(NULLIF(md.title,''),m.title),COALESCE(NULLIF(md.year,0),m.year),
 date(p.updated_at,'unixepoch','+8 hours'),'local',CAST(m.id AS TEXT)||':'||date(p.updated_at,'unixepoch','+8 hours'),p.updated_at,p.updated_at
 FROM playback_progress p JOIN media_items m ON m.id=p.media_id
 LEFT JOIN movie_metadata md ON md.media_id=m.id WHERE p.watched=1;

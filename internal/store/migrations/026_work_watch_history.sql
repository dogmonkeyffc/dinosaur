ALTER TABLE playback_progress ADD COLUMN completed_at INTEGER;
ALTER TABLE watch_history ADD COLUMN series_id INTEGER REFERENCES tv_series(id) ON DELETE SET NULL;
ALTER TABLE watch_history ADD COLUMN work_id INTEGER REFERENCES works(id) ON DELETE SET NULL;
CREATE UNIQUE INDEX watch_history_local_series ON watch_history(user_id,series_id)
 WHERE source='local' AND series_id IS NOT NULL;
CREATE UNIQUE INDEX watch_history_local_work ON watch_history(user_id,work_id)
 WHERE source='local' AND work_id IS NOT NULL;

CREATE VIEW history_episode_catalog AS
 SELECT ep.media_id,ep.series_id,ep.season_number,ep.episode_number,ep.episode_end,
 w.id AS work_id,COALESCE(w.title,NULLIF(json_extract(s.metadata_json,'$.title'),''),s.title) AS title,
 COALESCE(NULLIF(json_extract(s.metadata_json,'$.year'),0),s.year) AS year
 FROM tv_episodes ep JOIN media_items m ON m.id=ep.media_id JOIN tv_series s ON s.id=ep.series_id
 LEFT JOIN work_members wm ON wm.series_id=s.id LEFT JOIN works w ON w.id=wm.work_id
 WHERE m.missing=0 AND ep.season_number>0;

-- Legacy watch flags alone also included the old 90% threshold.
UPDATE playback_progress AS p SET completed_at=(
 SELECT MAX(h.created_at) FROM watch_history h
 WHERE h.user_id=p.user_id AND h.media_id=p.media_id AND h.source='local'
)
WHERE p.watched=1 AND EXISTS (
 SELECT 1 FROM tv_episodes e JOIN media_items m ON m.id=e.media_id
 WHERE e.media_id=p.media_id AND m.duration>0 AND p.position>=m.duration
);

-- Keep one dated record for each fully completed series or work.
WITH watched_slots AS (
 SELECT p.user_id,e.series_id,e.season_number,e.episode_number,e.episode_end
 FROM playback_progress p JOIN history_episode_catalog e ON e.media_id=p.media_id
 WHERE p.completed_at IS NOT NULL
 GROUP BY p.user_id,e.series_id,e.season_number,e.episode_number,e.episode_end
),
ranked AS (
 SELECT h.id,e.series_id,e.work_id,e.title,e.year,
 row_number() OVER (
  PARTITION BY h.user_id,e.work_id,CASE WHEN e.work_id IS NULL THEN e.series_id END
  ORDER BY h.watched_on DESC,h.id DESC
 ) AS choice
 FROM watch_history h JOIN history_episode_catalog e ON e.media_id=h.media_id
 JOIN playback_progress p ON p.user_id=h.user_id AND p.media_id=h.media_id
 WHERE h.source='local' AND p.completed_at IS NOT NULL AND NOT EXISTS (
  SELECT 1 FROM history_episode_catalog required
  LEFT JOIN watched_slots done ON done.user_id=h.user_id AND done.series_id=required.series_id
   AND done.season_number=required.season_number AND done.episode_number=required.episode_number
   AND done.episode_end=required.episode_end
  WHERE (required.work_id=e.work_id OR (e.work_id IS NULL AND required.series_id=e.series_id))
   AND done.user_id IS NULL
 )
)
UPDATE watch_history SET media_id=NULL,
 series_id=CASE WHEN ranked.work_id IS NULL THEN ranked.series_id END,work_id=ranked.work_id,
 source_id=CASE WHEN ranked.work_id IS NULL THEN 'series:'||ranked.series_id ELSE 'work:'||ranked.work_id END,
 title=ranked.title,year=CASE WHEN ranked.work_id IS NULL THEN ranked.year ELSE 0 END
FROM ranked WHERE watch_history.id=ranked.id AND ranked.choice=1;

DELETE FROM watch_history WHERE source='local' AND media_id IN (SELECT media_id FROM tv_episodes);

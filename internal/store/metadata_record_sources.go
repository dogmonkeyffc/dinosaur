package store

// Records have one movie or one entire series as their target. Episode files
// remain playback targets and must never be submitted as series metadata IDs.
const metadataRecordSources = `WITH sources AS (
 SELECT m.id AS media_id,0 AS series_id,m.id AS sort_id,'movie' AS media_type,
 m.library_id,l.name AS library_name,` + displayTitle + ` AS title,m.title AS local_title,
 m.relative_path AS filename,COALESCE(md.tmdb_id,0) AS tmdb_id,COALESCE(md.bangumi_id,0) AS bangumi_id,
 COALESCE(md.original_title,'') AS original_title,COALESCE(md.aliases,'') AS aliases,
 COALESCE(md.language,'') AS language,COALESCE(md.status,'') AS raw_status,
 COALESCE(md.message,'') AS message,COALESCE(md.updated_at,0) AS updated_at,
 COALESCE(md.candidates_json,'[]') AS candidates_json
 ` + mediaFrom + ` WHERE m.missing=0 AND l.type='movie'
 UNION ALL
 SELECT 0,s.id,s.id,'tv',s.library_id,l.name,` + seriesTitle + `,s.title,
 COALESCE((SELECT m.relative_path FROM tv_episodes e JOIN media_items m ON m.id=e.media_id
 WHERE e.series_id=s.id AND m.missing=0 ORDER BY e.season_number,e.episode_number,m.id LIMIT 1),''),
 COALESCE(json_extract(s.metadata_json,'$.tmdb_id'),0),
 COALESCE(json_extract(s.metadata_json,'$.bangumi_id'),0),
 COALESCE(json_extract(s.metadata_json,'$.original_title'),''),s.aliases,
 COALESCE(json_extract(s.metadata_json,'$.language'),''),
 COALESCE(json_extract(s.metadata_json,'$.status'),''),
 COALESCE(json_extract(s.metadata_json,'$.message'),''),
 COALESCE(json_extract(s.metadata_json,'$.updated_at'),0),
 COALESCE(json_extract(s.metadata_json,'$.candidates'),'[]')
 ` + seriesFrom + ` WHERE l.type='tv' AND ` + seriesVisible + `)
 `

const recordStatus = `CASE WHEN (tmdb_id>0 OR bangumi_id>0) THEN 'matched'
 WHEN raw_status='failed' THEN 'failed' WHEN raw_status='needs_review' THEN 'needs_review' ELSE 'unmatched' END`

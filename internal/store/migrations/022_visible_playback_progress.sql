CREATE VIEW visible_playback_progress AS
SELECT p.* FROM playback_progress p
JOIN media_items m ON m.id=p.media_id
WHERE NOT EXISTS (
 SELECT 1 FROM history_library_settings h
 WHERE h.user_id=p.user_id AND h.library_id=m.library_id AND h.record_history=0
);

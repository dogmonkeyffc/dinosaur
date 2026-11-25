DELETE FROM playback_progress
WHERE EXISTS (
 SELECT 1 FROM media_items m JOIN history_library_settings h ON h.library_id=m.library_id
 WHERE m.id=playback_progress.media_id AND h.user_id=playback_progress.user_id AND h.record_history=0
);
DELETE FROM watch_history
WHERE source='local' AND EXISTS (
 SELECT 1 FROM media_items m JOIN history_library_settings h ON h.library_id=m.library_id
 WHERE m.id=watch_history.media_id AND h.user_id=watch_history.user_id AND h.record_history=0
);

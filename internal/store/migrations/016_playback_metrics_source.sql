ALTER TABLE playback_metrics ADD COLUMN source_json TEXT;
ALTER TABLE playback_metrics ADD COLUMN reason TEXT NOT NULL DEFAULT '';

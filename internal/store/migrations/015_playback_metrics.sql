CREATE TABLE playback_metrics (
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  id TEXT NOT NULL,
  stream_id TEXT,
  mode TEXT NOT NULL,
  quality TEXT NOT NULL,
  burn_subtitle INTEGER CHECK(burn_subtitle IN (0, 1)),
  width INTEGER NOT NULL,
  height INTEGER NOT NULL,
  startup_ms INTEGER,
  stalls INTEGER NOT NULL,
  stall_ms INTEGER NOT NULL,
  played_ms INTEGER NOT NULL,
  buffer_ms INTEGER NOT NULL,
  dropped_frames INTEGER,
  total_frames INTEGER,
  updated_at_ns INTEGER NOT NULL,
  PRIMARY KEY (user_id, id)
);
CREATE INDEX playback_metrics_recency ON playback_metrics(updated_at_ns);

-- Keep speed reports independently: they can arrive before browser snapshots.
CREATE TABLE playback_transcode_speeds (
  user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  stream_id TEXT NOT NULL,
  speed REAL NOT NULL,
  updated_at_ns INTEGER NOT NULL,
  PRIMARY KEY (user_id, stream_id)
);
CREATE INDEX playback_transcode_speeds_recency ON playback_transcode_speeds(updated_at_ns);

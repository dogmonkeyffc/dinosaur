import type { PlaybackEngine } from './engine';

export function clearSubtitles(engine: PlaybackEngine) {
  engine.subtitleRevision++;
  for (const track of Array.from(engine.video.textTracks)) {
    track.mode = 'disabled';
    for (const cue of Array.from(track.cues || [])) track.removeCue(cue);
  }
  engine.video.querySelectorAll('track').forEach((track) => track.remove());
}

export function applySubtitle(engine: PlaybackEngine) {
  clearSubtitles(engine);
  const revision = engine.subtitleRevision;
  const id = engine.subtitle;
  if (!id) return;
  const track = document.createElement('track');
  const offset = engine.offset;
  track.kind = 'subtitles';
  track.label = engine.info.subtitles.find((item) => String(item.id) === id)?.label || '字幕';
  track.srclang = 'und';
  track.src = `/api/playback/${engine.info.item.id}/subtitles/${id}`;
  track.default = true;
  track.addEventListener('load', () => {
    if (revision !== engine.subtitleRevision || engine.disposed || !track.isConnected) {
      track.track.mode = 'disabled';
      return;
    }
    if (offset)
      for (const cue of Array.from(track.track.cues || [])) {
        if (cue.endTime <= offset) track.track.removeCue(cue);
        else {
          const end = cue.endTime - offset;
          cue.startTime = Math.max(0, cue.startTime - offset);
          cue.endTime = end;
        }
      }
    track.track.mode = 'showing';
  });
  track.addEventListener('error', () => {
    if (!engine.disposed && revision === engine.subtitleRevision)
      engine.callbacks.notify('字幕加载失败，请检查 UTF-8 编码的 SRT 或 WebVTT 文件。');
  });
  engine.video.append(track);
  track.track.mode = 'showing';
}

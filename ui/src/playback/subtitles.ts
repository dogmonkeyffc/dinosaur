import type { PlaybackEngine } from './engine';

export function clearSubtitles(engine: PlaybackEngine) {
  engine.subtitleRevision++;
  engine.subtitleWindow = -1;
  for (const track of Array.from(engine.video.textTracks)) {
    track.mode = 'disabled';
    for (const cue of Array.from(track.cues || [])) track.removeCue(cue);
  }
  engine.video.querySelectorAll('track').forEach((track) => track.remove());
}

export function refreshSubtitle(engine: PlaybackEngine) {
  const selected = engine.info.subtitles.find((item) => String(item.id) === engine.subtitle);
  if (
    engine.ready &&
    !engine.loading &&
    selected?.kind === 'embedded' &&
    selected.delivery === 'text' &&
    Math.floor(engine.position() / 60) !== engine.subtitleWindow
  )
    applySubtitle(engine, true);
}

export function applySubtitle(engine: PlaybackEngine, refresh = false) {
  if (!refresh) clearSubtitles(engine);
  const selected = engine.info.subtitles.find((item) => String(item.id) === engine.subtitle);
  if (!selected || selected.delivery === 'burn') return;
  const revision = ++engine.subtitleRevision;
  const track = document.createElement('track');
  const offset = engine.offset;
  const window = Math.floor(engine.position() / 60);
  engine.subtitleWindow = window;
  track.kind = 'subtitles';
  track.label = selected.label;
  track.srclang = selected.language || 'und';
  track.src =
    `/api/playback/${engine.info.item.id}/subtitles/${selected.id}` +
    (selected.kind === 'embedded' ? `?start=${window * 60}` : '');
  track.addEventListener('load', () => {
    if (revision !== engine.subtitleRevision || engine.disposed || !track.isConnected) {
      track.remove();
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
    for (const cue of Array.from(track.track.cues || [])) {
      if (cue instanceof VTTCue) {
        cue.line = -3;
        cue.snapToLines = true;
      }
    }
    engine.video.querySelectorAll('track').forEach((previous) => {
      if (previous !== track) {
        previous.track.mode = 'disabled';
        previous.remove();
      }
    });
    track.track.mode = 'showing';
  });
  track.addEventListener('error', () => {
    track.remove();
    if (!engine.disposed && revision === engine.subtitleRevision)
      engine.callbacks.notify('字幕加载失败，请重试或切换其他字幕轨。');
  });
  engine.video.append(track);
  track.track.mode = 'hidden';
}

'use strict';

let subtitleRevision = 0;
function clearPlayerSubtitles() {
  subtitleRevision++;
  if (!video) return;
  for (const track of Array.from(video.textTracks)) {
    track.mode = 'disabled';
    for (const cue of Array.from(track.cues || [])) track.removeCue(cue);
  }
  video.querySelectorAll('track').forEach(track => track.remove());
}
function applyPlayerSubtitle() {
  clearPlayerSubtitles();
  const revision = subtitleRevision;
  const id = player.subtitle;
  if (!player.info || id === '') return;
  const track = document.createElement('track');
  const offset = player.offset;
  track.kind = 'subtitles';
  track.label = player.info.subtitles.find(item => String(item.id) === id)?.label || '字幕';
  track.srclang = 'und';
  track.src = `/api/playback/${player.info.item.id}/subtitles/${id}`;
  track.default = true;
  track.addEventListener('load', () => {
    if (revision !== subtitleRevision || !track.isConnected || !player.info) {
      track.track.mode = 'disabled';
      return;
    }
    if (offset) for (const cue of Array.from(track.track.cues || [])) {
      if (cue.endTime <= offset) track.track.removeCue(cue);
      else {
        const end = cue.endTime - offset;
        cue.startTime = Math.max(0, cue.startTime - offset);
        cue.endTime = end;
      }
    }
    track.track.mode = 'showing';
  });
  track.addEventListener('error', () => notify('字幕加载失败，请检查文件是否为 UTF-8 编码的 SRT 或 WebVTT。', true));
  video.append(track);
  track.track.mode = 'showing';
}

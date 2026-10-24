'use strict';

function audioTrackLabel(track) {
  const languages = { chi: '中文', zho: '中文', jpn: '日语', fre: '法语', fra: '法语', eng: '英语' };
  return [track.title || languages[track.language] || track.language || `音轨 ${track.index}`,
    track.channels ? `${track.channels} 声道` : ''].filter(Boolean).join(' · ');
}
function playerSettings() {
  const info = player.info;
  const settings = [];
  if (info.audio_tracks.length) settings.push({
    name: 'audio-track', html: '音轨',
    tooltip: escapeHTML(audioTrackLabel(info.audio_tracks.find(track => track.index === player.audio) || info.audio_tracks[0])),
    selector: info.audio_tracks.map(track => ({
      html: escapeHTML(audioTrackLabel(track)), value: track.index, default: track.index === player.audio,
    })),
    onSelect: async item => {
      await changePlayerSource({ audio: item.value });
      return item.html;
    },
  });
  settings.push({
    name: 'subtitle-track', html: '字幕', tooltip: info.subtitles.length ? '关闭字幕' : '无外挂字幕',
    selector: [{ html: '关闭字幕', value: '', default: true }, ...info.subtitles.map(track => ({
      html: escapeHTML(track.label), value: String(track.id),
    }))],
    onSelect: item => {
      player.subtitle = item.value;
      applyPlayerSubtitle();
      return item.html;
    },
  }, {
    name: 'playback-mode', html: '播放方式', tooltip: '自动选择',
    selector: [
      { html: '自动选择', value: 'auto', default: true },
      { html: '兼容模式', value: 'compatible' },
    ],
    onSelect: async item => {
      await changePlayerSource({ mode: item.value });
      return item.html;
    },
  });
  return settings;
}

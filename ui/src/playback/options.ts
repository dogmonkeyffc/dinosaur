import type { PlaybackEngine } from './engine';
import { escapeHTML, type ArtInstance } from './artplayer';
import { closeSetting, settingKeyboard, syncOptions, toggleSetting } from './option-menu';

const icons = {
  audio: '<path d="M9 5 4 9H2v6h2l5 4V5Z"/><path d="M13 8a6 6 0 0 1 0 8m3-11a10 10 0 0 1 0 14"/>',
  subtitle: '<rect x="2" y="4" width="20" height="16" rx="2"/><path d="M6 12h4m4 0h4M6 16h7m3 0h2"/>',
  mode: '<path d="M4 6h16M4 12h16M4 18h16"/><circle cx="9" cy="6" r="2"/><circle cx="15" cy="12" r="2"/><circle cx="8" cy="18" r="2"/>',
  rate: '<path d="M4 18a9 9 0 1 1 16 0M12 12l5-5M5 18h14"/><circle cx="12" cy="12" r="1"/>',
};
type Key = keyof typeof icons;
interface Option {
  value: string;
  html: string;
}

export function settingsControls(engine: PlaybackEngine) {
  const info = engine.info;
  const languages: Record<string, string> = {
    chi: '中文',
    zho: '中文',
    jpn: '日语',
    fre: '法语',
    fra: '法语',
    eng: '英语',
  };
  return [
    control(
      engine,
      'audio',
      '音轨',
      10,
      info.audio_tracks.map((track) => ({
        value: String(track.index),
        html: escapeHTML(
          [
            track.title || languages[track.language] || track.language || `音轨 ${track.index}`,
            track.channels ? `${track.channels} 声道` : '',
          ]
            .filter(Boolean)
            .join(' · '),
        ),
      })),
    ),
    control(engine, 'subtitle', '字幕', 20, [
      { value: '', html: info.subtitles.length ? '关闭字幕' : '无外挂字幕' },
      ...info.subtitles.map((track) => ({ value: String(track.id), html: escapeHTML(track.label) })),
    ]),
    control(engine, 'mode', '播放方式', 30, [
      { value: 'auto', html: '自动选择' },
      { value: 'compatible', html: '兼容模式' },
    ]),
    control(
      engine,
      'rate',
      '倍速',
      40,
      [0.5, 0.75, 1, 1.25, 1.5].map((rate) => ({ value: String(rate), html: `${rate}×` })),
    ),
  ];
}

function control(engine: PlaybackEngine, key: Key, label: string, index: number, options: Option[]) {
  const disabled =
    (key === 'audio' && !engine.info.audio_tracks.length) ||
    (key === 'subtitle' && !engine.info.subtitles.length);
  const html = `<button class="player-setting-icon" type="button" aria-label="${label}" title="${label}" aria-haspopup="menu" aria-expanded="false" ${disabled ? 'disabled' : ''}><svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${icons[key]}</svg></button>`;
  let element: HTMLElement;
  return {
    name: `option-${key}`,
    position: 'right',
    index,
    html,
    selector: options.map((item) => ({ ...item, default: String(engine[key] ?? '') === item.value })),
    click(_control: unknown, event: MouseEvent) {
      if (disabled || (event.target as Element).closest('.art-selector-list')) return;
      toggleSetting(engine, element);
    },
    async onSelect(item: Option) {
      if (key === 'rate') {
        engine.rate = Number(item.value);
        engine.art.playbackRate = engine.rate;
      } else if (key === 'subtitle') engine.selectSubtitle(item.value);
      else await engine.changeSource({ [key]: key === 'audio' ? Number(item.value) : item.value });
      if (element.isConnected && !engine.disposed) {
        syncOptions(engine);
        closeSetting(element);
      }
      return html;
    },
    mounted(this: ArtInstance, node: HTMLElement) {
      element = node;
      element.dataset.settingKey = key;
      const menu = element.querySelector('.art-selector-list')!;
      menu.setAttribute('role', 'menu');
      menu.setAttribute('aria-label', label);
      element.querySelectorAll<HTMLElement>('.art-selector-item').forEach((item) => {
        item.setAttribute('role', 'menuitemradio');
        item.tabIndex = -1;
      });
      syncOptions(engine, element);
      this.events.proxy(element, 'keydown', (event) =>
        settingKeyboard(engine, element, event as KeyboardEvent),
      );
      this.events.proxy(element, 'focusout', (event) => {
        if (!element.contains((event as FocusEvent).relatedTarget as Node)) closeSetting(element);
      });
      this.events.proxy(document, 'click', (event) => {
        if (!element.contains(event.target as Node)) closeSetting(element);
      });
    },
  };
}

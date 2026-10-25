'use strict';

const playerSettingIcons = {
  audio: '<path d="M9 5 4 9H2v6h2l5 4V5Z"/><path d="M13 8a6 6 0 0 1 0 8m3-11a10 10 0 0 1 0 14"/>',
  subtitle: '<rect x="2" y="4" width="20" height="16" rx="2"/><path d="M6 12h4m4 0h4M6 16h7m3 0h2"/>',
  mode: '<path d="M4 6h16M4 12h16M4 18h16"/><circle cx="9" cy="6" r="2"/><circle cx="15" cy="12" r="2"/><circle cx="8" cy="18" r="2"/>',
  rate: '<path d="M4 18a9 9 0 1 1 16 0M12 12l5-5M5 18h14"/><circle cx="12" cy="12" r="1"/>',
};
function audioTrackLabel(track) {
  const languages = { chi: '中文', zho: '中文', jpn: '日语', fre: '法语', fra: '法语', eng: '英语' };
  return [track.title || languages[track.language] || track.language || `音轨 ${track.index}`,
    track.channels ? `${track.channels} 声道` : ''].filter(Boolean).join(' · ');
}
function playerSettingsControls() {
  const info = player.info;
  return [
    settingIconControl('audio', '音轨', 10, info.audio_tracks.map(track => ({ value: String(track.index), html: escapeHTML(audioTrackLabel(track)) }))),
    settingIconControl('subtitle', '字幕', 20, [{ value: '', html: info.subtitles.length ? '关闭字幕' : '无外挂字幕' }, ...info.subtitles.map(track => ({ value: String(track.id), html: escapeHTML(track.label) }))]),
    settingIconControl('mode', '播放方式', 30, [{ value: 'auto', html: '自动选择' }, { value: 'compatible', html: '兼容模式' }]),
    settingIconControl('rate', '倍速', 40, Artplayer.PLAYBACK_RATE.map(rate => ({ value: String(rate), html: `${rate}×` }))),
  ];
}
function settingIconControl(key, label, index, options) {
  const disabled = (key === 'audio' && !player.info.audio_tracks.length) || (key === 'subtitle' && !player.info.subtitles.length);
  const html = `<button class="player-setting-icon" type="button" aria-label="${label}" title="${label}" aria-haspopup="menu" aria-expanded="false" ${disabled ? 'disabled' : ''}><svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="1.7" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">${playerSettingIcons[key]}</svg></button>`;
  let element;
  return {
    name: `option-${key}`, position: 'right', index, html,
    selector: options.map(item => ({ ...item, default: String(player[key] ?? '') === item.value })),
    click(control, event) {
      if (disabled || event.target.closest('.art-selector-list')) return;
      togglePlayerSetting(element);
    },
    onSelect: async item => {
      if (key === 'rate') {
        player.rate = Number(item.value);
        art.playbackRate = player.rate;
      } else if (key === 'subtitle') {
        player.subtitle = item.value;
        applyPlayerSubtitle();
      } else await changePlayerSource({ [key]: key === 'audio' ? Number(item.value) : item.value });
      if (element.isConnected && player.info) {
        syncPlaybackOptions();
        closePlayerSetting(element);
      }
      return html;
    },
    mounted(node) {
      element = node;
      element.dataset.settingKey = key;
      const menu = element.querySelector('.art-selector-list');
      menu.setAttribute('role', 'menu');
      menu.setAttribute('aria-label', label);
      element.querySelectorAll('.art-selector-item').forEach(item => {
        item.setAttribute('role', 'menuitemradio');
        item.tabIndex = -1;
      });
      syncPlaybackOptions(element);
      this.events.proxy(element, 'keydown', event => playerSettingKeyboard(element, event));
      this.events.proxy(element, 'focusout', event => {
        if (!element.contains(event.relatedTarget)) closePlayerSetting(element);
      });
      this.events.proxy(document, 'click', event => {
        if (!element.contains(event.target)) closePlayerSetting(element);
      });
    },
  };
}
function closePlayerSetting(element) {
  element.classList.remove('is-menu-open');
  element.querySelector('button')?.setAttribute('aria-expanded', 'false');
}
function togglePlayerSetting(element, open = !element.classList.contains('is-menu-open')) {
  document.querySelectorAll('#player-art [data-setting-key]').forEach(closePlayerSetting);
  if (!open) return;
  const button = element.getBoundingClientRect();
  const frame = $('player-art').getBoundingClientRect();
  const menu = element.querySelector('.art-selector-list');
  const width = Math.min(220, frame.width - 16);
  const left = Math.max(frame.x + 8, Math.min(button.right - width, frame.right - width - 8));
  menu.style.width = `${width}px`;
  menu.style.left = `${left - button.x}px`;
  menu.style.maxHeight = `${Math.min(260, Math.max(80, frame.height - 110))}px`;
  element.classList.add('is-menu-open');
  element.querySelector('button').setAttribute('aria-expanded', 'true');
}
function playerSettingKeyboard(element, event) {
  event.stopPropagation();
  if (event.key === 'Escape') {
    event.preventDefault();
    closePlayerSetting(element);
    element.querySelector('button').focus();
  } else if (['ArrowDown', 'ArrowUp'].includes(event.key)) {
    event.preventDefault();
    if (!element.classList.contains('is-menu-open')) togglePlayerSetting(element, true);
    const items = [...element.querySelectorAll('.art-selector-item')];
    const index = items.indexOf(document.activeElement);
    items[(index + (event.key === 'ArrowDown' ? 1 : -1) + items.length) % items.length].focus();
  } else if (['Enter', ' '].includes(event.key) && event.target.matches('.art-selector-item')) {
    event.preventDefault();
    event.target.click();
  }
}
function syncPlaybackOptions(root = $('player-art')) {
  if (!root || !player.info) return;
  const controls = root.hasAttribute('data-setting-key') ? [root] : root.querySelectorAll('[data-setting-key]');
  for (const element of controls) {
    const selected = String(player[element.dataset.settingKey] ?? '');
    element.querySelectorAll('.art-selector-item').forEach(item => {
      const active = item.dataset.value === selected;
      item.classList.toggle('art-current', active);
      item.setAttribute('aria-checked', String(active));
    });
  }
}

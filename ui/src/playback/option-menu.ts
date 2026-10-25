import type { PlaybackEngine } from './engine';

export function closeSetting(element: HTMLElement) {
  element.classList.remove('is-menu-open');
  element.querySelector('button')?.setAttribute('aria-expanded', 'false');
}

export function toggleSetting(
  engine: PlaybackEngine,
  element: HTMLElement,
  open = !element.classList.contains('is-menu-open'),
) {
  engine.root.querySelectorAll<HTMLElement>('[data-setting-key]').forEach(closeSetting);
  if (!open) return;
  const button = element.getBoundingClientRect();
  const frame = engine.root.getBoundingClientRect();
  const menu = element.querySelector<HTMLElement>('.art-selector-list')!;
  const width = Math.min(220, frame.width - 16);
  const left = Math.max(frame.x + 8, Math.min(button.right - width, frame.right - width - 8));
  menu.style.width = `${width}px`;
  menu.style.left = `${left - button.x}px`;
  menu.style.maxHeight = `${Math.min(260, Math.max(80, frame.height - 110))}px`;
  element.classList.add('is-menu-open');
  element.querySelector('button')!.setAttribute('aria-expanded', 'true');
}

export function settingKeyboard(engine: PlaybackEngine, element: HTMLElement, event: KeyboardEvent) {
  event.stopPropagation();
  if (event.key === 'Escape') {
    event.preventDefault();
    closeSetting(element);
    element.querySelector('button')!.focus();
  } else if (['ArrowDown', 'ArrowUp'].includes(event.key)) {
    event.preventDefault();
    if (!element.classList.contains('is-menu-open')) toggleSetting(engine, element, true);
    const items = Array.from(element.querySelectorAll<HTMLElement>('.art-selector-item'));
    const index = items.indexOf(document.activeElement as HTMLElement);
    items[(index + (event.key === 'ArrowDown' ? 1 : -1) + items.length) % items.length]?.focus();
  } else if (['Enter', ' '].includes(event.key) && (event.target as Element).matches('.art-selector-item')) {
    event.preventDefault();
    (event.target as HTMLElement).click();
  }
}

export function syncOptions(engine: PlaybackEngine, root = engine.root) {
  const controls = root.hasAttribute('data-setting-key')
    ? [root]
    : root.querySelectorAll<HTMLElement>('[data-setting-key]');
  for (const element of controls) {
    const key = element.dataset.settingKey as 'audio' | 'subtitle' | 'mode' | 'rate';
    const selected = String(engine[key] ?? '');
    element.querySelectorAll<HTMLElement>('.art-selector-item').forEach((item) => {
      const active = item.dataset.value === selected;
      item.classList.toggle('art-current', active);
      item.setAttribute('aria-checked', String(active));
    });
  }
}

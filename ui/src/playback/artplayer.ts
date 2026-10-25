export interface ArtInstance {
  video: HTMLVideoElement;
  url: string;
  poster: string;
  playbackRate: number;
  notice: { show: string };
  events: { proxy: (target: EventTarget, event: string, handler: (event: Event) => void) => void };
  destroy: () => void;
}
interface ArtConstructor {
  new (options: Record<string, unknown>): ArtInstance;
  LOG_VERSION: boolean;
  CONTEXTMENU: boolean;
  RECONNECT_TIME_MAX: number;
  SETTING_WIDTH: number;
  PLAYBACK_RATE: number[];
}
declare global {
  interface Window {
    Artplayer: ArtConstructor;
  }
}
export const escapeHTML = (value: string) =>
  value.replace(
    /[&<>"']/g,
    (character) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' })[character]!,
  );

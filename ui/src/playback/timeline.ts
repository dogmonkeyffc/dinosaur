import type { PlaybackEngine } from './engine';

const nativeTime = Object.getOwnPropertyDescriptor(HTMLMediaElement.prototype, 'currentTime')!;
const nativeBuffered = Object.getOwnPropertyDescriptor(HTMLMediaElement.prototype, 'buffered')!;
const nativePlay = HTMLMediaElement.prototype.play;
export const currentTime = (video: HTMLVideoElement) => Number(nativeTime.get!.call(video));
export const setCurrentTime = (video: HTMLVideoElement, value: number) => nativeTime.set!.call(video, value);

export function createTimelineVideo(engine: PlaybackEngine) {
  const element = document.createElement('video');
  element.id = 'player-video';
  element.setAttribute('aria-label', '媒体视频');
  Object.defineProperties(element, {
    duration: { get: () => engine.info.item.duration },
    currentTime: {
      get: () => engine.seekTarget ?? engine.position(),
      set: (value: number) => engine.seek(value),
    },
    buffered: {
      get: () => {
        const ranges = nativeBuffered.get!.call(element) as TimeRanges;
        const offset = engine.offset;
        return {
          length: ranges.length,
          start: (index: number) => ranges.start(index) + offset,
          end: (index: number) => ranges.end(index) + offset,
        };
      },
    },
  });
  element.play = () => {
    if (engine.released || element.error || element.ended) {
      engine.start(element.ended ? 0 : engine.position());
      return Promise.resolve();
    }
    return nativePlay.call(element);
  };
  return element;
}

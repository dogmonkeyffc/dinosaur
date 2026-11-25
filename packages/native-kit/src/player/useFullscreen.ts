import { useEffect, useRef, useState } from 'react';
import * as Orientation from 'expo-screen-orientation';

export function useFullscreen(failed: (reason: string) => void) {
  const [fullscreen, setFullscreen] = useState(false);
  const previous = useRef<Orientation.OrientationLock | undefined>(undefined);
  const busy = useRef(false);
  useEffect(
    () => () => {
      if (previous.current !== undefined) void Orientation.lockAsync(previous.current).catch(() => {});
    },
    [],
  );
  async function toggle() {
    if (busy.current) return;
    busy.current = true;
    try {
      if (fullscreen) {
        await Orientation.lockAsync(previous.current ?? Orientation.OrientationLock.DEFAULT);
        previous.current = undefined;
        setFullscreen(false);
      } else {
        previous.current = await Orientation.getOrientationLockAsync();
        await Orientation.lockAsync(Orientation.OrientationLock.LANDSCAPE);
        setFullscreen(true);
      }
    } catch {
      failed('屏幕方向切换失败，请重试。');
    } finally {
      busy.current = false;
    }
  }
  return { fullscreen, toggle };
}

import { useState } from 'react';
import { Text, View } from 'react-native';
import { clock } from '@dinosaur/playback-core';
import { phoneStyles as css } from './phoneStyles';

export function SeekBar({
  time,
  duration,
  seek,
  touching,
}: {
  time: number;
  duration: number;
  seek: (delta: number) => void;
  touching: (value: boolean) => void;
}) {
  const [width, setWidth] = useState(1);
  const [preview, setPreview] = useState<number>();
  const target = (x: number) => Math.max(0, Math.min(1, x / width)) * duration;
  const position = Math.max(0, Math.min(preview ?? time, duration));
  return (
    <View style={css.timeline}>
      <Text style={css.time}>
        {clock(preview ?? time)} / {clock(duration)}
      </Text>
      <View
        accessible
        accessibilityRole="adjustable"
        accessibilityLabel="播放进度"
        accessibilityValue={{ min: 0, max: duration, now: position }}
        accessibilityActions={[
          { name: 'increment', label: '前进 10 秒' },
          { name: 'decrement', label: '后退 10 秒' },
        ]}
        onAccessibilityAction={(event) => seek(event.nativeEvent.actionName === 'increment' ? 10 : -10)}
        onLayout={(event) => setWidth(event.nativeEvent.layout.width)}
        style={css.scrubber}
        onStartShouldSetResponder={() => duration > 0}
        onResponderGrant={(event) => {
          touching(true);
          setPreview(target(event.nativeEvent.locationX));
        }}
        onResponderMove={(event) => setPreview(target(event.nativeEvent.locationX))}
        onResponderRelease={(event) => {
          seek(target(event.nativeEvent.locationX) - time);
          setPreview(undefined);
          touching(false);
        }}
        onResponderTerminate={() => {
          setPreview(undefined);
          touching(false);
        }}
      >
        <View style={css.track}>
          <View style={[css.fill, { width: `${duration > 0 ? (position / duration) * 100 : 0}%` }]} />
        </View>
      </View>
    </View>
  );
}

import { Pressable, StyleSheet, Text, View } from 'react-native';
import { PlaybackIcon, type PlaybackIconName } from './PlaybackIcon';

export function PlayerAction({
  label,
  icon,
  text,
  press,
  disabled = false,
}: {
  label: string;
  icon?: PlaybackIconName;
  text?: string;
  press: () => void;
  disabled?: boolean;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      disabled={disabled}
      onPress={press}
      style={({ pressed }) => [css.action, pressed && css.pressed, disabled && css.disabled]}
    >
      {icon ? <PlaybackIcon name={icon} /> : <Text style={css.label}>{text}</Text>}
    </Pressable>
  );
}

export function PhoneTransport({
  playing,
  toggle,
  seek,
  next,
  wide = false,
}: {
  playing: boolean;
  toggle: () => void;
  seek: (delta: number) => void;
  next?: () => void;
  wide?: boolean;
}) {
  return (
    <View style={css.transport}>
      <PlayerAction label={playing ? '暂停' : '播放'} icon={playing ? 'pause' : 'play'} press={toggle} />
      {wide && <PlayerAction label="后退 10 秒" icon="backward" press={() => seek(-10)} />}
      {wide && <PlayerAction label="前进 10 秒" icon="forward" press={() => seek(10)} />}
      {next && <PlayerAction label="下一集" icon="next" press={next} />}
    </View>
  );
}
const css = StyleSheet.create({
  transport: { flexDirection: 'row', alignItems: 'center' },
  action: { width: 40, height: 40, alignItems: 'center', justifyContent: 'center', borderRadius: 4 },
  pressed: { backgroundColor: '#ffffff18' },
  disabled: { opacity: 0.35 },
  label: { fontSize: 12, color: '#f2f6fc' },
});

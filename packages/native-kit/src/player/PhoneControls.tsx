import { Pressable, StyleSheet, Text, View } from 'react-native';

export function PlayerAction({
  label,
  text,
  press,
  primary = false,
}: {
  label: string;
  text: string;
  press: () => void;
  primary?: boolean;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      onPress={press}
      style={({ pressed }) => [css.action, primary && css.primary, pressed && css.pressed]}
    >
      <Text style={[css.label, primary && css.primaryLabel]}>{text}</Text>
    </Pressable>
  );
}

export function PhoneTransport({
  playing,
  toggle,
  seek,
  next,
}: {
  playing: boolean;
  toggle: () => void;
  seek: (delta: number) => void;
  next?: () => void;
}) {
  return (
    <View style={css.transport}>
      <PlayerAction label="后退 10 秒" text="↶ 10" press={() => seek(-10)} />
      <PlayerAction label={playing ? '暂停' : '播放'} text={playing ? 'Ⅱ' : '▶'} press={toggle} primary />
      <PlayerAction label="前进 30 秒" text="30 ↷" press={() => seek(30)} />
      {next && <PlayerAction label="下一集" text="下一集 ▷|" press={next} />}
    </View>
  );
}

const css = StyleSheet.create({
  transport: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 10,
    paddingVertical: 10,
  },
  action: {
    minWidth: 44,
    minHeight: 44,
    paddingHorizontal: 12,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: '#28292c',
  },
  primary: { width: 60, height: 60, paddingHorizontal: 0, backgroundColor: '#a8c7fa', borderRadius: 30 },
  pressed: { opacity: 0.65 },
  label: { fontSize: 13, fontWeight: '500', color: '#e5e5e9' },
  primaryLabel: { color: '#14233b', fontSize: 23 },
});

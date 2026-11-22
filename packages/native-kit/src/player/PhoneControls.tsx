import { Pressable, StyleSheet, Text, View } from 'react-native';
import { colors } from '../theme';

export function PlayerAction({
  label, text, press, primary = false,
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
  playing, toggle, seek,
}: {
  playing: boolean;
  toggle: () => void;
  seek: (delta: number) => void;
}) {
  return (
    <View style={css.transport}>
      <PlayerAction label="后退 10 秒" text="−10 秒" press={() => seek(-10)} />
      <PlayerAction label={playing ? '暂停' : '播放'} text={playing ? 'Ⅱ' : '▶'} press={toggle} primary />
      <PlayerAction label="前进 30 秒" text="+30 秒" press={() => seek(30)} />
    </View>
  );
}

export const phonePlayerStyles = StyleSheet.create({
  screen: { paddingHorizontal: 16, paddingTop: 48, paddingBottom: 20, gap: 10, backgroundColor: '#121510' },
  title: { flex: 1 },
  video: { overflow: 'hidden', backgroundColor: '#000', borderRadius: 12 },
  portraitVideo: { aspectRatio: 16 / 9, width: '100%' },
  landscapeVideo: { flex: 1 },
  options: { flexDirection: 'row', flexWrap: 'wrap', alignItems: 'center', gap: 4 },
});

const css = StyleSheet.create({
  transport: { flexDirection: 'row', alignItems: 'center', justifyContent: 'center', gap: 18 },
  action: {
    minWidth: 44,
    minHeight: 44,
    paddingHorizontal: 10,
    borderRadius: 24,
    alignItems: 'center',
    justifyContent: 'center',
  },
  primary: { width: 48, height: 48, paddingHorizontal: 0, backgroundColor: colors.accent },
  pressed: { opacity: 0.65 },
  label: { fontSize: 14, fontWeight: '600', color: colors.text },
  primaryLabel: { color: colors.background, fontSize: 18 },
});

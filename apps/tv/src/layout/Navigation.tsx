import { useState } from 'react';
import { Pressable, StyleSheet, Text, View } from 'react-native';
import type { MediaKind } from '@dinosaur/contracts';
import { tvColors } from './theme';

export const navigationWidth = 200;

function Destination({
  label, selected = false, preferred, press,
}: {
  label: string;
  selected?: boolean;
  preferred?: boolean;
  press: () => void;
}) {
  const [focused, setFocused] = useState(false);
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityState={{ selected }}
      hasTVPreferredFocus={preferred}
      onFocus={() => setFocused(true)}
      onBlur={() => setFocused(false)}
      onPress={press}
      style={[css.destination, selected && css.selected, focused && css.focused]}
    >
      <Text style={[css.label, (selected || focused) && css.activeLabel]}>{label}</Text>
    </Pressable>
  );
}

export function Navigation({
  kind, select, logout, preferred,
}: {
  kind: MediaKind;
  select: (kind: MediaKind) => void;
  logout: () => void;
  preferred: boolean;
}) {
  return (
    <View style={css.rail}>
      <Text style={css.brand}>Dinosaur</Text>
      <Destination label="电影" selected={kind === 'movie'} preferred={preferred && kind === 'movie'} press={() => select('movie')} />
      <Destination label="电视剧" selected={kind === 'tv'} preferred={preferred && kind === 'tv'} press={() => select('tv')} />
      <View style={css.spacer} />
      <Destination label="退出登录" press={logout} />
    </View>
  );
}

const css = StyleSheet.create({
  rail: {
    width: navigationWidth,
    paddingHorizontal: 20,
    paddingVertical: 32,
    gap: 8,
    backgroundColor: tvColors.container,
    borderRightWidth: 1,
    borderRightColor: tvColors.outline,
  },
  brand: { color: tvColors.text, fontSize: 24, fontWeight: '700', marginBottom: 16, paddingLeft: 12 },
  destination: {
    minHeight: 48,
    justifyContent: 'center',
    borderWidth: 2,
    borderColor: 'transparent',
    borderRadius: 24,
    paddingHorizontal: 14,
  },
  selected: { backgroundColor: tvColors.selected },
  focused: { backgroundColor: tvColors.selected, borderColor: tvColors.focus },
  label: { color: tvColors.muted, fontSize: 16, fontWeight: '500' },
  activeLabel: { color: tvColors.text, fontWeight: '700' },
  spacer: { flex: 1 },
});

import type { ReactNode } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Screen } from '@dinosaur/native-kit';
import { tvColors } from './theme';

export function Entry({ title, children }: { title: string; children: ReactNode }) {
  return (
    <Screen tv backgroundColor={tvColors.surface}>
      <View style={css.shell}>
        <View style={css.heading}>
          <Text style={css.brand}>Dinosaur</Text>
          <Text accessibilityRole="header" style={css.title}>{title}</Text>
        </View>
        <View style={css.panel}>{children}</View>
      </View>
    </Screen>
  );
}

const css = StyleSheet.create({
  shell: {
    width: '100%',
    maxWidth: 880,
    flexDirection: 'row',
    alignItems: 'stretch',
    alignSelf: 'center',
    marginVertical: 'auto',
    borderWidth: 1,
    borderColor: tvColors.outline,
    borderRadius: 24,
    overflow: 'hidden',
  },
  heading: { flex: 2, justifyContent: 'center', gap: 10, backgroundColor: tvColors.container, padding: 24 },
  brand: { color: tvColors.accent, fontSize: 17, fontWeight: '700' },
  title: { color: tvColors.text, fontSize: 28, fontWeight: '700' },
  panel: { flex: 3, gap: 16, padding: 24, backgroundColor: tvColors.raised },
});

import { useState } from 'react';
import { Modal, Pressable, ScrollView, StyleSheet, Text, View } from 'react-native';
import type { MediaLibrary } from '@dinosaur/contracts';
import { colors } from '@dinosaur/native-kit';

export function HomeHeader({ accounts, pair }: { accounts: () => void; pair: () => void }) {
  const [open, setOpen] = useState(false);
  function choose(action: () => void) {
    setOpen(false);
    action();
  }
  return (
    <View style={css.heading}>
      <Text accessibilityRole="header" style={css.brand}>Dinosaur</Text>
      <Pressable
        accessibilityLabel="账号菜单"
        accessibilityRole="button"
        accessibilityState={{ expanded: open }}
        onPress={() => setOpen(true)}
        style={css.menuButton}
      >
        <Text style={css.menuIcon}>⋯</Text>
      </Pressable>
      <Modal visible={open} transparent animationType="fade" onRequestClose={() => setOpen(false)}>
        <View style={css.overlay}>
          <Pressable
            accessibilityLabel="关闭账号菜单"
            accessibilityRole="button"
            style={StyleSheet.absoluteFill}
            onPress={() => setOpen(false)}
          />
          <View accessibilityViewIsModal style={css.menu}>
            <Pressable accessibilityRole="button" style={css.action} onPress={() => choose(accounts)}>
              <Text style={css.label}>切换账号</Text>
            </Pressable>
            <Pressable accessibilityRole="button" style={css.action} onPress={() => choose(pair)}>
              <Text style={css.label}>登录电视</Text>
            </Pressable>
          </View>
        </View>
      </Modal>
    </View>
  );
}

export function LibraryTabs({ libraries, selected, select, refresh }: {
  libraries: MediaLibrary[];
  selected?: number;
  select: (id: number) => void;
  refresh: () => void;
}) {
  return (
    <View style={css.heading}>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={css.tabs}>
        {libraries.map((library) => (
          <Pressable
            key={library.id}
            accessibilityRole="tab"
            accessibilityState={{ selected: selected === library.id }}
            onPress={() => select(library.id)}
            style={[css.tab, selected === library.id && css.selected]}
          >
            <Text style={css.label}>{library.name}</Text>
          </Pressable>
        ))}
      </ScrollView>
      <Pressable accessibilityRole="button" accessibilityLabel="刷新" onPress={refresh} style={css.menuButton}>
        <Text style={css.menuIcon}>↻</Text>
      </Pressable>
    </View>
  );
}

const css = StyleSheet.create({
  heading: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: 8 },
  brand: { fontSize: 27, fontWeight: '700', color: colors.text },
  menuButton: { width: 44, height: 44, borderRadius: 22, alignItems: 'center', justifyContent: 'center' },
  menuIcon: { fontSize: 25, color: colors.text },
  overlay: { flex: 1, alignItems: 'flex-end', paddingTop: 58, paddingHorizontal: 24, backgroundColor: '#0005' },
  menu: { minWidth: 156, padding: 6, borderRadius: 16, backgroundColor: colors.surface },
  action: { padding: 14, minHeight: 48 },
  label: { color: colors.text, fontWeight: '600', fontSize: 15 },
  tabs: { flexDirection: 'row', padding: 3, gap: 3, borderRadius: 24, backgroundColor: colors.surface },
  tab: { minHeight: 44, paddingHorizontal: 18, justifyContent: 'center', borderRadius: 22 },
  selected: { backgroundColor: '#384832' },
});

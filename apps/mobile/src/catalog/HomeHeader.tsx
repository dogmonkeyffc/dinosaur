import { useState, type RefObject } from 'react';
import { Modal, Pressable, ScrollView, StyleSheet, Text, TextInput, View } from 'react-native';
import type { MediaLibrary } from '@dinosaur/contracts';
import { Icon } from './HomeIcons';
import { homeStyles as css, palette } from './homeStyles';

export type HomeTab = 'home' | 'library' | 'search';

export function HomeHeader({
  accounts,
  pair,
  username,
  query,
  search,
  focusSearch,
  inputRef,
}: {
  accounts: () => void;
  pair: () => void;
  username: string;
  query: string;
  search: (value: string) => void;
  focusSearch: () => void;
  inputRef: RefObject<TextInput | null>;
}) {
  const [open, setOpen] = useState(false);
  function choose(action: () => void) {
    setOpen(false);
    action();
  }
  return (
    <View style={css.heading}>
      <View style={css.search}>
        <Icon name="search" />
        <TextInput
          ref={inputRef}
          accessibilityLabel="搜索媒体"
          placeholder="搜索电影、剧集"
          placeholderTextColor={palette.muted}
          value={query}
          onFocus={focusSearch}
          onChangeText={search}
          autoCapitalize="none"
          autoCorrect={false}
          returnKeyType="search"
          style={css.searchInput}
        />
        {!!query && (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel="清除搜索"
            onPress={() => search('')}
            style={css.clear}
          >
            <Text style={css.clearText}>×</Text>
          </Pressable>
        )}
      </View>
      <Pressable
        accessibilityLabel="账号菜单"
        accessibilityRole="button"
        accessibilityState={{ expanded: open }}
        onPress={() => setOpen(true)}
        style={css.avatarRing}
      >
        <View style={css.avatar}>
          <Text style={css.avatarText}>{username.slice(0, 1).toUpperCase()}</Text>
        </View>
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
            <Pressable accessibilityRole="button" style={css.menuAction} onPress={() => choose(accounts)}>
              <Text style={css.menuLabel}>切换账号</Text>
            </Pressable>
            <Pressable accessibilityRole="button" style={css.menuAction} onPress={() => choose(pair)}>
              <Text style={css.menuLabel}>登录电视</Text>
            </Pressable>
          </View>
        </View>
      </Modal>
    </View>
  );
}

export function LibraryTabs({
  libraries,
  selected,
  select,
}: {
  libraries: MediaLibrary[];
  selected?: number;
  select: (id: number) => void;
}) {
  return (
    <View style={css.tabBorder}>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={css.tabs}>
        {libraries.map((library) => (
          <Pressable
            key={library.id}
            accessibilityRole="tab"
            accessibilityState={{ selected: selected === library.id }}
            onPress={() => select(library.id)}
            style={css.tab}
          >
            <Text style={[css.tabLabel, selected === library.id && css.activeText]}>{library.name}</Text>
            {selected === library.id && <View style={css.underline} />}
          </Pressable>
        ))}
      </ScrollView>
    </View>
  );
}

export function HomeNavigation({ tab, select }: { tab: HomeTab; select: (tab: HomeTab) => void }) {
  const tabs = [
    { id: 'home', label: '首页', icon: 'home' },
    { id: 'library', label: '片库', icon: 'library' },
    { id: 'search', label: '搜索', icon: 'search' },
  ] as const;
  return (
    <View accessibilityRole="tablist" style={css.navigation}>
      {tabs.map((item) => (
        <Pressable
          key={item.id}
          accessibilityRole="tab"
          accessibilityLabel={item.label}
          accessibilityState={{ selected: tab === item.id }}
          onPress={() => select(item.id)}
          style={css.navItem}
        >
          <View
            key={tab === item.id ? 'active' : 'inactive'}
            collapsable={false}
            style={[css.navIcon, tab === item.id && css.activeIcon]}
          >
            <Icon name={item.icon} color={tab === item.id ? palette.accent : palette.muted} />
          </View>
          <Text style={[css.navLabel, tab === item.id && css.activeText]}>{item.label}</Text>
        </Pressable>
      ))}
    </View>
  );
}

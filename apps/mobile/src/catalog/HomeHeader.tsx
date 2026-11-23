import type { RefObject } from 'react';
import { Pressable, ScrollView, Text, TextInput, View } from 'react-native';
import type { MediaLibrary } from '@dinosaur/contracts';
import { Icon } from './HomeIcons';
import { AccountButton } from './AccountButton';
import { homeStyles as css, palette } from './homeStyles';

export type HomeTab = 'home' | 'library' | 'search';

export function HomeHeader({
  settings,
  username,
  query,
  search,
  focusSearch,
  inputRef,
}: {
  settings: () => void;
  username: string;
  query: string;
  search: (value: string) => void;
  focusSearch: () => void;
  inputRef: RefObject<TextInput | null>;
}) {
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
      <AccountButton username={username} open={settings} />
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

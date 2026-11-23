import type { RefObject } from 'react';
import {
  Image,
  Platform,
  Pressable,
  RefreshControl,
  ScrollView,
  StatusBar,
  Text,
  View,
  useWindowDimensions,
} from 'react-native';
import { Busy, ErrorText } from '@dinosaur/native-kit';
import { AccountButton } from './AccountButton';
import { Artwork, Pagination, type HomeMedia } from './HomeContent';
import { LibraryTabs } from './HomeHeader';
import type { MobileCatalog } from './useMobileCatalog';
import { libraryStyles as css } from './libraryStyles';
import { homeStyles, palette } from './homeStyles';

export function LibraryLanding({
  catalog: c,
  username,
  settings,
  open,
  scroll,
  select,
  changePage,
}: {
  catalog: MobileCatalog;
  username: string;
  settings: () => void;
  open: (item: HomeMedia) => void;
  scroll: RefObject<ScrollView | null>;
  select: (id: number) => void;
  changePage: (page: number) => void;
}) {
  const { height } = useWindowDimensions();
  const featured = c.list.items[0];
  const error = c.error || c.list.error;
  return (
    <ScrollView
      ref={scroll}
      contentContainerStyle={css.content}
      showsVerticalScrollIndicator={false}
      refreshControl={
        <RefreshControl
          refreshing={c.refreshing}
          onRefresh={c.refresh}
          colors={[palette.accent]}
          tintColor={palette.accent}
        />
      }
    >
      <View style={[css.hero, { minHeight: Math.max(300, Math.min(430, height * 0.48)) }]}>
        {!!featured?.poster_url && (
          <Image source={[c.client.resource(featured.poster_url)]} style={css.backdrop} resizeMode="cover" />
        )}
        <View style={css.shade} />
        <View
          style={[
            css.header,
            { paddingTop: (Platform.OS === 'android' ? (StatusBar.currentHeight ?? 24) : 52) + 12 },
          ]}
        >
          <Text accessibilityLabel="Dinosaur" style={css.logo}>
            D.
          </Text>
          <AccountButton username={username} open={settings} />
        </View>
        <View style={css.heroBody}>
          <Text accessibilityRole="header" numberOfLines={3} style={css.title}>
            {featured?.title ?? '片库'}
          </Text>
          {featured && (
            <View style={css.action}>
              <View style={css.thumbnails} importantForAccessibility="no-hide-descendants">
                {c.list.items.slice(1, 4).map((item) => (
                  <View key={item.id} style={css.thumbnail}>
                    <Artwork key={item.poster_url ?? item.id} item={item} client={c.client} />
                  </View>
                ))}
              </View>
              <Pressable
                accessibilityRole="button"
                accessibilityLabel={`查看${featured.title}`}
                style={css.button}
                onPress={() => open(featured)}
              >
                <Text style={css.buttonText}>查看</Text>
              </Pressable>
            </View>
          )}
        </View>
      </View>
      <LibraryTabs libraries={c.libraries} selected={c.library?.id} select={select} />
      {!!error && (
        <View style={css.feedback}>
          <ErrorText error={error} />
          <Pressable accessibilityRole="button" style={css.retry} onPress={c.refresh}>
            <Text style={homeStyles.posterTitle}>重试</Text>
          </Pressable>
        </View>
      )}
      {!c.refreshing && (c.loading || c.list.loading) && <Busy />}
      {!c.loading && c.libraries.length === 0 && <Text style={homeStyles.message}>暂无媒体库</Text>}
      {!c.list.loading && !error && c.library && !featured && (
        <Text style={homeStyles.message}>暂无媒体</Text>
      )}
      {c.list.items.length > 1 && (
        <View style={css.list}>
          <Text accessibilityRole="header" style={css.listTitle}>
            {c.library?.name}
          </Text>
          {c.list.items.slice(1).map((item) => (
            <Pressable
              key={item.id}
              accessibilityRole="button"
              accessibilityLabel={item.title}
              onPress={() => open(item)}
              style={({ pressed }) => [css.row, pressed && { opacity: 0.65 }]}
            >
              <View style={css.cover}>
                <Artwork key={item.poster_url ?? item.id} item={item} client={c.client} />
              </View>
              <Text numberOfLines={2} style={css.rowTitle}>
                {item.title}
              </Text>
              <Text style={css.arrow}>›</Text>
            </Pressable>
          ))}
        </View>
      )}
      <Pagination page={c.page} total={c.list.total} loading={c.list.loading} change={changePage} />
    </ScrollView>
  );
}

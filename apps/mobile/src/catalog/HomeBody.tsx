import type { RefObject } from 'react';
import { Pressable, RefreshControl, ScrollView, Text, View } from 'react-native';
import { Busy, ErrorText } from '@dinosaur/native-kit';
import type { useHomeVisibility } from '../settings/useHomeVisibility';
import { homeRecent } from '../settings/homeVisibility';
import { HomeFeed, MediaGrid, Pagination, type HomeMedia } from './HomeContent';
import type { MobileCatalog } from './useMobileCatalog';
import { homeStyles as css, palette } from './homeStyles';

export function HomeBody({
  catalog: c,
  home,
  visibility,
  scroll,
  open,
  play,
  browse,
  changePage,
}: {
  catalog: MobileCatalog;
  home: boolean;
  visibility: ReturnType<typeof useHomeVisibility>;
  scroll: RefObject<ScrollView | null>;
  open: (item: HomeMedia) => void;
  play: (id: number) => void;
  browse: () => void;
  changePage: (page: number) => void;
}) {
  const loading = c.loading || c.list.loading || (home && !visibility.ready);
  const error = c.error || c.list.error || (home ? visibility.error : '');
  return (
    <ScrollView
      ref={scroll}
      keyboardShouldPersistTaps="handled"
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
      {!!error && (
        <View style={css.feedback}>
          <ErrorText error={error} />
          {home && !visibility.ready && (
            <Pressable accessibilityRole="button" onPress={visibility.retry}>
              <Text style={css.posterTitle}>重试</Text>
            </Pressable>
          )}
        </View>
      )}
      {loading && !error && !c.refreshing && <Busy />}
      {!c.loading && c.libraries.length === 0 && <Text style={css.message}>暂无媒体库</Text>}
      {home ? (
        visibility.ready &&
        c.library && (
          <HomeFeed
            key={c.library.id}
            items={c.list.items}
            recent={homeRecent(c.list.recent, c.libraries, visibility.hidden)}
            title={c.library.name}
            client={c.client}
            open={open}
            play={play}
            browse={browse}
          />
        )
      ) : (
        <>
          <Text accessibilityRole="header" style={[css.sectionTitle, { paddingHorizontal: 24 }]}>
            {c.query.trim() ? '搜索结果' : '浏览片库'}
          </Text>
          <MediaGrid items={c.list.items} client={c.client} open={open} />
        </>
      )}
      {home && visibility.ready && !c.loading && c.libraries.length > 0 && !c.available.length && (
        <Text style={css.message}>暂无首页媒体库</Text>
      )}
      {c.library && !loading && !error && !c.list.items.length && (
        <Text style={css.message}>{c.query.trim() ? '没有找到相关媒体' : '暂无媒体'}</Text>
      )}
      {!home && c.library && (
        <Pagination page={c.page} total={c.list.total} loading={c.list.loading} change={changePage} />
      )}
    </ScrollView>
  );
}

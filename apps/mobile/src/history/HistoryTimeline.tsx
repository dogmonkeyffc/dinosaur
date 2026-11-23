import { Platform, RefreshControl, SectionList, StatusBar, Text, View, Pressable } from 'react-native';
import type { Client } from '@dinosaur/client';
import { Busy, ErrorText } from '@dinosaur/native-kit';
import { AccountButton } from '../catalog/AccountButton';
import { palette } from '../catalog/homeStyles';
import { HistoryCard } from './HistoryCard';
import { timelineSections } from './timeline';
import { useTimeline } from './useTimeline';
import { timelineStyles as css } from './timelineStyles';

export function HistoryTimeline({
  client,
  username,
  settings,
}: {
  client: Client;
  username: string;
  settings: () => void;
}) {
  const timeline = useTimeline(client);
  return (
    <View style={[css.screen, { paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 52 }]}>
      <View style={css.header}>
        <Text accessibilityRole="header" style={css.title}>
          观看记录
        </Text>
        <Text style={css.count}>{timeline.total}</Text>
        <AccountButton username={username} open={settings} />
      </View>
      <SectionList
        sections={timelineSections(timeline.items)}
        keyExtractor={(item) => String(item.id)}
        stickySectionHeadersEnabled
        showsVerticalScrollIndicator={false}
        renderSectionHeader={({ section }) => (
          <Text accessibilityRole="header" style={css.month}>
            {section.title}
          </Text>
        )}
        renderItem={({ item }) => <HistoryCard entry={item} client={client} />}
        refreshControl={
          <RefreshControl
            refreshing={timeline.refreshing}
            onRefresh={timeline.refresh}
            colors={[palette.accent]}
            tintColor={palette.accent}
          />
        }
        onEndReached={timeline.more}
        onEndReachedThreshold={0.35}
        ListEmptyComponent={
          !timeline.loading && !timeline.error ? (
            <Text style={[css.feedback, css.review]}>暂无观看记录</Text>
          ) : null
        }
        ListFooterComponent={
          <View style={css.feedback}>
            {!!timeline.error && (
              <>
                <ErrorText error={timeline.error} />
                <Pressable accessibilityRole="button" style={css.action} onPress={timeline.retry}>
                  <Text style={css.actionText}>重试</Text>
                </Pressable>
              </>
            )}
            {timeline.loading && !timeline.refreshing && <Busy />}
          </View>
        }
      />
    </View>
  );
}

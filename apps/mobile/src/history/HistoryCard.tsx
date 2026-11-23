import { useState } from 'react';
import { Image, Linking, Modal, Pressable, ScrollView, Text, View } from 'react-native';
import type { Client } from '@dinosaur/client';
import type { WatchHistoryEntry } from '@dinosaur/contracts';
import { timelineStyles as css } from './timelineStyles';

function source(entry: WatchHistoryEntry) {
  return entry.source === 'douban' ? '豆瓣' : entry.source === 'local' ? '站内' : '手动记录';
}

export function HistoryCard({ entry, client }: { entry: WatchHistoryEntry; client: Client }) {
  const [failed, setFailed] = useState(false);
  const [open, setOpen] = useState(false);
  const [linkError, setLinkError] = useState('');
  const stars = '★'.repeat(entry.rating) + '☆'.repeat(5 - entry.rating);
  return (
    <View style={css.row}>
      <View style={css.dateColumn}>
        <Text style={css.day}>{Number(entry.watched_on.slice(8))}</Text>
        <View style={css.dot} />
        <View style={css.line} />
      </View>
      <Pressable
        accessibilityRole="button"
        accessibilityLabel={`${entry.watched_on} ${entry.title}`}
        onPress={() => setOpen(true)}
        style={({ pressed }) => [css.card, pressed && { opacity: 0.7 }]}
      >
        {entry.poster_url && !failed ? (
          <Image
            source={[client.resource(entry.poster_url)]}
            style={css.poster}
            resizeMode="cover"
            onError={() => setFailed(true)}
          />
        ) : (
          <View style={css.poster}>
            <Text style={css.placeholder}>{entry.title.slice(0, 1)}</Text>
          </View>
        )}
        <View style={css.info}>
          <Text numberOfLines={2} style={css.itemTitle}>
            {entry.title}
          </Text>
          {entry.rating > 0 && (
            <Text accessibilityLabel={`${entry.rating} 星`} style={css.stars}>
              {stars}
            </Text>
          )}
          {!!entry.review && (
            <Text numberOfLines={2} style={css.review}>
              {entry.review}
            </Text>
          )}
          <Text style={css.source}>{source(entry)}</Text>
        </View>
      </Pressable>
      {open && (
        <Modal transparent animationType="fade" onRequestClose={() => setOpen(false)}>
          <View style={css.backdrop}>
            <View accessibilityViewIsModal style={css.detail}>
              <View style={css.closeRow}>
                <Text style={css.source}>{entry.watched_on}</Text>
                <Pressable
                  accessibilityRole="button"
                  accessibilityLabel="关闭观看记录"
                  style={css.action}
                  onPress={() => setOpen(false)}
                >
                  <Text style={css.close}>×</Text>
                </Pressable>
              </View>
              <ScrollView contentContainerStyle={css.detailContent}>
                <Text style={css.detailTitle}>{entry.title}</Text>
                {!!entry.original_title && <Text style={css.review}>{entry.original_title}</Text>}
                {entry.year > 0 && <Text style={css.review}>{entry.year}</Text>}
                {entry.rating > 0 && <Text style={css.stars}>{stars}</Text>}
                {!!entry.review && (
                  <Text selectable style={css.detailText}>
                    {entry.review}
                  </Text>
                )}
                {entry.source === 'douban' && /^[1-9][0-9]*$/.test(entry.source_id) ? (
                  <Pressable
                    accessibilityRole="link"
                    style={css.action}
                    onPress={() => {
                      Linking.openURL(`https://movie.douban.com/subject/${entry.source_id}/`).catch(() =>
                        setLinkError('无法打开豆瓣链接。'),
                      );
                    }}
                  >
                    <Text style={css.actionText}>豆瓣</Text>
                  </Pressable>
                ) : (
                  <Text style={css.source}>{source(entry)}</Text>
                )}
                {!!linkError && (
                  <Text accessibilityRole="alert" style={css.review}>
                    {linkError}
                  </Text>
                )}
              </ScrollView>
            </View>
          </View>
        </Modal>
      )}
    </View>
  );
}

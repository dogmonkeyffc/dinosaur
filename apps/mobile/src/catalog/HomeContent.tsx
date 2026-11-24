import { useState } from 'react';
import { Image, Pressable, ScrollView, Text, View, useWindowDimensions } from 'react-native';
import type { Client } from '@dinosaur/client';
import type { Movie, Series } from '@dinosaur/contracts';
import { homeStyles as css } from './homeStyles';
import { dp } from '../ui/density';

export type HomeMedia = Movie | Series;

export function Artwork({
  item,
  client,
  featured = false,
}: {
  item: HomeMedia;
  client: Client;
  featured?: boolean;
}) {
  const [failed, setFailed] = useState(false);
  const style = featured ? css.featureImage : css.posterImage;
  if (featured && item.poster_url && !failed)
    return (
      <View style={css.featureImage}>
        <Image
          source={[client.resource(item.poster_url)]}
          blurRadius={dp(18)}
          style={css.featureBackdrop}
          resizeMode="cover"
        />
        <Image
          source={[client.resource(item.poster_url)]}
          style={css.featureCover}
          resizeMode="cover"
          onError={() => setFailed(true)}
        />
      </View>
    );
  return item.poster_url && !failed ? (
    <Image
      source={[client.resource(item.poster_url)]}
      style={style}
      resizeMode="cover"
      onError={() => setFailed(true)}
      accessibilityIgnoresInvertColors
    />
  ) : (
    <View style={[style, css.placeholder]}>
      <Text style={css.placeholderText}>{item.title.slice(0, 1)}</Text>
    </View>
  );
}

function MediaCard({
  item,
  client,
  width,
  open,
}: {
  item: HomeMedia;
  client: Client;
  width: number;
  open: () => void;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={item.title}
      onPress={open}
      style={({ pressed }) => [css.poster, { width, opacity: pressed ? 0.7 : 1 }]}
    >
      <Artwork key={item.poster_url ?? item.id} item={item} client={client} />
      <Text numberOfLines={1} style={css.posterTitle}>
        {item.title}
      </Text>
    </Pressable>
  );
}

export function Shelf({
  title,
  items,
  client,
  open,
  more,
}: {
  title: string;
  items: HomeMedia[];
  client: Client;
  open: (item: HomeMedia) => void;
  more?: () => void;
}) {
  const { width } = useWindowDimensions();
  const cardWidth = dp(Math.min(150, Math.max(100, (width - 64) / 3.1)));
  if (!items.length) return null;
  return (
    <View style={css.section}>
      <View style={css.sectionHeading}>
        <Text accessibilityRole="header" style={css.sectionTitle}>
          {title}
        </Text>
        {more && (
          <Pressable
            accessibilityRole="button"
            accessibilityLabel={`查看全部${title}`}
            hitSlop={8}
            onPress={more}
            style={css.more}
          >
            <Text style={css.arrow}>→</Text>
          </Pressable>
        )}
      </View>
      <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={css.shelf}>
        {items.map((item) => (
          <MediaCard key={item.id} item={item} client={client} width={cardWidth} open={() => open(item)} />
        ))}
      </ScrollView>
    </View>
  );
}

export function HomeFeed({
  items,
  recent,
  title,
  client,
  open,
  play,
  browse,
}: {
  items: HomeMedia[];
  recent: Movie[];
  title: string;
  client: Client;
  open: (item: HomeMedia) => void;
  play: (id: number) => void;
  browse: () => void;
}) {
  const featured = items.find((item) => item.poster_url);
  return (
    <>
      <Shelf title="继续观看" items={recent.slice(0, 8)} client={client} open={(item) => play(item.id)} />
      <Shelf title={title} items={items.slice(0, 10)} client={client} open={open} more={browse} />
      {featured && (
        <Pressable
          accessibilityRole="button"
          accessibilityLabel={featured.title}
          onPress={() => open(featured)}
          style={({ pressed }) => [css.feature, pressed && { opacity: 0.8 }]}
        >
          <View style={css.featureArt}>
            <Artwork key={featured.poster_url} item={featured} client={client} featured />
          </View>
          <View style={css.featureFooter}>
            <Text numberOfLines={2} style={css.featureTitle}>
              {featured.title}
            </Text>
            <View style={css.featureArrow}>
              <Text style={css.arrow}>→</Text>
            </View>
          </View>
        </Pressable>
      )}
      <Shelf title="更多内容" items={items.slice(10)} client={client} open={open} more={browse} />
    </>
  );
}

export function MediaGrid({
  items,
  client,
  open,
}: {
  items: HomeMedia[];
  client: Client;
  open: (item: HomeMedia) => void;
}) {
  const { width } = useWindowDimensions();
  const columns = width > 700 ? 5 : width >= 360 ? 3 : 2;
  const cardWidth = Math.floor((width - dp(48) - dp(14) * (columns - 1)) / columns);
  return (
    <View style={css.grid}>
      {items.map((item) => (
        <MediaCard key={item.id} item={item} client={client} width={cardWidth} open={() => open(item)} />
      ))}
    </View>
  );
}

export function Pagination({
  page,
  total,
  loading,
  change,
}: {
  page: number;
  total: number;
  loading: boolean;
  change: (page: number) => void;
}) {
  if (total <= 24) return null;
  const pages = Math.ceil(total / 24);
  return (
    <View style={css.pagination}>
      {[-1, 0, 1].map((delta) => {
        if (delta === 0)
          return (
            <Text key={delta} style={css.posterTitle}>
              {page} / {pages}
            </Text>
          );
        const disabled = loading || page + delta < 1 || page + delta > pages;
        return (
          <Pressable
            key={delta}
            accessibilityRole="button"
            accessibilityLabel={delta < 0 ? '上一页' : '下一页'}
            hitSlop={4}
            disabled={disabled}
            accessibilityState={{ disabled }}
            onPress={() => change(page + delta)}
            style={[css.pageButton, disabled && css.disabled]}
          >
            <Text style={css.arrow}>{delta < 0 ? '←' : '→'}</Text>
          </Pressable>
        );
      })}
    </View>
  );
}

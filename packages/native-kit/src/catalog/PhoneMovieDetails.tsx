import { useState } from 'react';
import { Image, Pressable, StyleSheet, Text, View } from 'react-native';
import type { Client } from '@dinosaur/client';
import type { Movie } from '@dinosaur/contracts';
import { colors } from '../theme';

export function PhoneMovieDetails({ movie, client, play }: {
  movie: Movie;
  client: Client;
  play: () => void;
}) {
  const [expanded, setExpanded] = useState(false);
  const [overflow, setOverflow] = useState(false);
  const resume = movie.position > 0 && !movie.watched;
  const progress = resume && movie.duration > 0 ? Math.min(100, movie.position / movie.duration * 100) : 0;
  const facts = [movie.year || '', movie.duration > 0 ? `${Math.round(movie.duration / 60)} 分钟` : '']
    .filter(Boolean).join(' · ');
  const overview = movie.metadata?.overview.trim();

  return (
    <>
      <View style={css.hero}>
        {movie.poster_url && (
          <Image
            accessibilityIgnoresInvertColors
            source={[client.resource(movie.poster_url)]}
            style={css.poster}
            resizeMode="cover"
          />
        )}
        <View style={css.summary}>
          <Text accessibilityRole="header" style={css.title}>{movie.title}</Text>
          {!!facts && <Text style={css.facts}>{facts}</Text>}
          <Pressable
            accessibilityRole="button"
            onPress={play}
            style={({ pressed }) => [css.play, pressed && css.pressed]}
          >
            <Text style={css.playText}>{resume ? '继续观看' : '播放'}</Text>
          </Pressable>
          {progress > 0 && (
            <View
              accessible
              accessibilityRole="progressbar"
              accessibilityLabel="观看进度"
              accessibilityValue={{ now: Math.round(progress), min: 0, max: 100 }}
              style={css.progress}
            >
              <View style={[css.progressFill, { width: `${progress}%` }]} />
            </View>
          )}
        </View>
      </View>
      {!!overview && (
        <View style={css.overview}>
          <Text accessibilityRole="header" style={css.heading}>简介</Text>
          <Text
            style={css.copy}
            numberOfLines={expanded ? undefined : 4}
            onTextLayout={(event) => {
              if (event.nativeEvent.lines.length > 4) setOverflow(true);
            }}
          >{overview}</Text>
          {(overflow || overview.length > 105 || overview.split('\n').length > 4) && (
            <Pressable
              accessibilityRole="button"
              accessibilityState={{ expanded }}
              onPress={() => setExpanded((value) => !value)}
              style={css.disclosure}
            >
              <Text style={css.link}>{expanded ? '收起简介' : '展开简介'}</Text>
            </Pressable>
          )}
        </View>
      )}
    </>
  );
}

const css = StyleSheet.create({
  hero: { flexDirection: 'row', alignItems: 'center', gap: 18, paddingBottom: 8 },
  poster: { width: '39%', maxWidth: 156, aspectRatio: 2 / 3, borderRadius: 12, backgroundColor: colors.surface },
  summary: { flex: 1, alignItems: 'flex-start', gap: 8 },
  title: { color: colors.text, fontSize: 26, fontWeight: '700' },
  facts: { color: colors.muted, fontSize: 14, lineHeight: 21 },
  play: {
    minHeight: 44,
    paddingHorizontal: 18,
    borderRadius: 24,
    justifyContent: 'center',
    backgroundColor: colors.accent,
    marginTop: 4,
  },
  pressed: { opacity: 0.76 },
  playText: { color: colors.background, fontWeight: '700', fontSize: 15 },
  progress: { width: '100%', maxWidth: 168, height: 3, borderRadius: 2, marginTop: 3, backgroundColor: colors.border },
  progressFill: { height: '100%', borderRadius: 2, backgroundColor: colors.accent },
  overview: { borderTopWidth: 1, borderTopColor: colors.border, paddingTop: 16, gap: 8 },
  heading: { color: colors.text, fontSize: 17, fontWeight: '700' },
  copy: { color: colors.muted, fontSize: 15, lineHeight: 24 },
  disclosure: { alignSelf: 'flex-start', minHeight: 44, justifyContent: 'center' },
  link: { color: colors.accent, fontWeight: '600', fontSize: 14 },
});

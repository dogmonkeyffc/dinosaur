import { useState } from 'react';
import { Image, Pressable, Text, View } from 'react-native';
import type { Client } from '@dinosaur/client';
import type { Movie, Series } from '@dinosaur/contracts';
import { colors } from './theme';

export function Poster({
  item,
  client,
  width,
  press,
  preferred,
}: {
  item: Movie | Series;
  client: Client;
  width: number;
  press: () => void;
  preferred?: boolean;
}) {
  const [focused, setFocused] = useState(false);
  // RN TV 0.85 forwards Android image headers only when source is an array.
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={item.title}
      onPress={press}
      hasTVPreferredFocus={preferred}
      onFocus={() => setFocused(true)}
      onBlur={() => setFocused(false)}
      style={{
        width,
        gap: 10,
        padding: 4,
        borderRadius: 12,
        borderWidth: 3,
        borderColor: focused ? colors.accent : 'transparent',
      }}
    >
      {item.poster_url ? (
        <Image
          source={[client.resource(item.poster_url)]}
          style={{ width: '100%', aspectRatio: 2 / 3, borderRadius: 8, backgroundColor: colors.surface }}
          resizeMode="cover"
        />
      ) : (
        <View
          style={{
            aspectRatio: 2 / 3,
            borderRadius: 8,
            backgroundColor: colors.surface,
            justifyContent: 'center',
            alignItems: 'center',
          }}
        >
          <Text style={{ color: colors.accent, fontSize: 40 }}>D.</Text>
        </View>
      )}
      <Text numberOfLines={2} style={{ color: colors.text, fontSize: 16, lineHeight: 22 }}>
        {item.title}
      </Text>
    </Pressable>
  );
}

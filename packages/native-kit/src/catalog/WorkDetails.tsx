import { useRef } from 'react';
import { Text, View } from 'react-native';
import type { Client } from '@dinosaur/client';
import { Busy, Button, ErrorText, Screen } from '../components';
import { Poster } from '../Poster';
import { styles } from '../theme';
import { useWork } from './useWork';

export function WorkDetails({ client, id, tv, active, member, width, close, select }: {
  client: Client;
  id: number;
  tv: boolean;
  active: boolean;
  member?: number;
  width: number;
  close: () => void;
  select: (id: number) => void;
}) {
  const { detail, error, reload } = useWork(client, id);
  const scroll = useRef(0);
  return (
    <Screen key={String(active)} tv={tv} title={detail?.work.title || '作品详情'}
      scrollY={scroll.current} scrolled={(value) => { scroll.current = value; }}>
      <Button preferred={tv && active && !detail?.members.length} onPress={close}>返回</Button>
      <ErrorText error={error} />
      {error && <Button onPress={reload}>重试</Button>}
      {!detail && !error && <Busy />}
      {detail && (
        <>
          <Text style={styles.muted}>
            {detail.work.member_count} 篇章 · {detail.work.season_count} 季 · {detail.work.episode_count} 集
          </Text>
          <View style={[styles.row, { alignItems: 'flex-start' }]}>
            {detail.members.map((item, index) => (
              <Poster key={item.id} item={item} client={client} width={width}
                preferred={tv && active && (member === item.id || (member === undefined && index === 0))}
                press={() => select(item.id)} />
            ))}
          </View>
          {!detail.members.length && <Text style={styles.muted}>暂无篇章</Text>}
        </>
      )}
    </Screen>
  );
}

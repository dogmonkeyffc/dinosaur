import { useState } from 'react';
import { Text, View, useWindowDimensions } from 'react-native';
import type { Account, MediaKind } from '@dinosaur/contracts';
import {
  Busy,
  Button,
  ErrorText,
  Input,
  Player,
  Poster,
  Screen,
  styles,
  useClient,
  useMediaPage,
} from '@dinosaur/native-kit';
import { Details } from './Details';

function HomeSession({
  account,
  pair,
  accounts,
}: {
  account: Account;
  pair: () => void;
  accounts: () => void;
}) {
  const client = useClient(account);
  const [kind, setKind] = useState<MediaKind>('movie');
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<{ id: number; work: boolean }>();
  const [playing, setPlaying] = useState<number>();
  const list = useMediaPage(client, kind, page, query);
  const { width } = useWindowDimensions();
  const columns = width > 700 ? 4 : 2;
  const posterWidth = Math.floor((width - 48 - 8 * (columns - 1)) / columns);
  return (
    <View style={{ flex: 1 }}>
      {playing !== undefined && (
        <Player client={client} id={playing} close={() => {
          setPlaying(undefined);
          list.reload();
        }} />
      )}
      {selected && (
        <View style={{ flex: 1, display: playing === undefined ? 'flex' : 'none' }}
          accessibilityElementsHidden={playing !== undefined}
          importantForAccessibility={playing === undefined ? 'auto' : 'no-hide-descendants'}>
          <Details client={client} id={selected.id} work={selected.work} kind={kind}
            active={playing === undefined} close={() => setSelected(undefined)} play={setPlaying} />
        </View>
      )}
      {!selected && playing === undefined && (
        <Screen title="Dinosaur">
          <View style={styles.row}>
            <Button onPress={accounts}>切换账号</Button>
            <Button primary onPress={pair}>
              登录电视
            </Button>
          </View>
          {list.recent.length > 0 && (
            <View style={styles.card}>
              <Text style={styles.subtitle}>继续观看</Text>
              {list.recent.slice(0, 4).map((item) => (
                <Button key={item.id} onPress={() => setPlaying(item.id)}>
                  {item.title}
                </Button>
              ))}
            </View>
          )}
          <View style={styles.row}>
            <Button
              primary={kind === 'movie'}
              onPress={() => {
                setKind('movie');
                setPage(1);
              }}
            >
              电影
            </Button>
            <Button
              primary={kind === 'tv'}
              onPress={() => {
                setKind('tv');
                setPage(1);
              }}
            >
              电视剧
            </Button>
            <Button onPress={list.reload}>刷新</Button>
          </View>
          <Input
            accessibilityLabel="搜索媒体"
            placeholder="搜索片名"
            value={query}
            onChangeText={(value) => {
              setQuery(value);
              setPage(1);
            }}
          />
          <ErrorText error={list.error} />
          {list.loading && <Busy />}
          <View style={[styles.row, { gap: 8, alignItems: 'flex-start' }]}>
            {list.items.map((item) => (
              <Poster
                key={item.id}
                item={item}
                client={client}
                width={posterWidth}
                press={() => setSelected({ id: kind === 'tv' && 'work_id' in item && item.work_id ? item.work_id : item.id,
                    work: kind === 'tv' && 'work_id' in item && !!item.work_id })}
              />
            ))}
          </View>
          {!list.loading && !list.items.length && <Text style={styles.muted}>暂无媒体</Text>}
          <View style={styles.row}>
            <Button disabled={page <= 1 || list.loading} onPress={() => setPage((value) => value - 1)}>
              上一页
            </Button>
            <Text style={styles.muted}>
              {page} / {Math.max(1, Math.ceil(list.total / 24))}
            </Text>
            <Button
              disabled={page * 24 >= list.total || list.loading}
              onPress={() => setPage((value) => value + 1)}
            >
              下一页
            </Button>
          </View>
        </Screen>
      )}
    </View>
  );
}

export function Home(props: { account: Account; pair: () => void; accounts: () => void }) {
  return <HomeSession key={`${props.account.key}:${props.account.token}`} {...props} />;
}

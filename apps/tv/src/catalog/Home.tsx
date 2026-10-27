import { useEffect, useRef, useState } from 'react';
import { BackHandler, ScrollView, Text, View, useWindowDimensions } from 'react-native';
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

export function Home({ account, logout }: { account: Account; logout: () => void }) {
  const client = useClient(account);
  const [kind, setKind] = useState<MediaKind>('movie');
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<number>();
  const [playing, setPlaying] = useState<number>();
  const lastCard = useRef<number | undefined>(undefined);
  const scroll = useRef(0);
  const list = useMediaPage(client, kind, page, query);
  const { width } = useWindowDimensions();
  useEffect(() => {
    const handler = BackHandler.addEventListener('hardwareBackPress', () => {
      if (playing) return false;
      if (selected) {
        setSelected(undefined);
        return true;
      }
      return false;
    });
    return () => handler.remove();
  }, [selected, playing]);
  if (playing)
    return (
      <Player
        tv
        client={client}
        id={playing}
        close={() => {
          setPlaying(undefined);
          list.reload();
        }}
      />
    );
  if (selected)
    return (
      <Details
        client={client}
        id={selected}
        kind={kind}
        close={() => setSelected(undefined)}
        play={setPlaying}
      />
    );
  return (
    <Screen
      tv
      title="Dinosaur TV"
      scrollY={scroll.current}
      scrolled={(value) => {
        scroll.current = value;
      }}
    >
      <View style={styles.row}>
        <Text style={styles.text}>
          {account.user.username} · {account.server.name}
        </Text>
        <Button onPress={logout}>退出登录</Button>
      </View>
      {list.recent.length > 0 && (
        <View style={{ gap: 12 }}>
          <Text style={styles.subtitle}>继续观看</Text>
          <ScrollView horizontal contentContainerStyle={{ gap: 12 }}>
            {list.recent.map((item) => (
              <Button key={item.id} onPress={() => setPlaying(item.id)}>
                {item.title}
              </Button>
            ))}
          </ScrollView>
        </View>
      )}
      <View style={styles.row}>
        <Button
          preferred={!lastCard.current}
          primary={kind === 'movie'}
          onPress={() => {
            setKind('movie');
            setPage(1);
            lastCard.current = undefined;
          }}
        >
          电影
        </Button>
        <Button
          primary={kind === 'tv'}
          onPress={() => {
            setKind('tv');
            setPage(1);
            lastCard.current = undefined;
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
      <View style={[styles.row, { gap: 12, alignItems: 'flex-start' }]}>
        {list.items.map((item) => (
          <Poster
            key={item.id}
            client={client}
            item={item}
            width={Math.floor((width - 128) / 5)}
            preferred={lastCard.current === item.id}
            press={() => {
              lastCard.current = item.id;
              setSelected(item.id);
            }}
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
  );
}

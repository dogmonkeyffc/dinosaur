import { useRef, useState } from 'react';
import { ScrollView, Text, View, useWindowDimensions } from 'react-native';
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
import { Navigation, navigationWidth } from '../layout/Navigation';
import { tvColors } from '../layout/theme';

function HomeSession({ account, logout }: { account: Account; logout: () => void }) {
  const client = useClient(account);
  const [kind, setKind] = useState<MediaKind>('movie');
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<{ id: number; work: boolean }>();
  const [playing, setPlaying] = useState<number>();
  const lastCard = useRef<number | undefined>(undefined);
  const lastRecent = useRef<number | undefined>(undefined);
  const scroll = useRef(0);
  const list = useMediaPage(client, kind, page, query);
  const { width } = useWindowDimensions();
  const contentWidth = Math.max(160, width - navigationWidth - 80);
  const columns = Math.max(1, Math.floor((contentWidth + 12) / 144));
  const posterWidth = Math.floor((contentWidth - 12 * (columns - 1)) / columns);
  function selectKind(next: MediaKind) {
    if (next === kind) return;
    setKind(next);
    setPage(1);
    lastCard.current = undefined;
    lastRecent.current = undefined;
  }
  return (
    <View style={{ flex: 1 }}>
      {playing !== undefined && (
        <Player tv client={client} id={playing} close={() => {
          setPlaying(undefined);
          list.reload();
        }} />
      )}
      {selected && (
        <View style={{ flex: 1, display: playing === undefined ? 'flex' : 'none' }}
          accessibilityElementsHidden={playing !== undefined}
          importantForAccessibility={playing === undefined ? 'auto' : 'no-hide-descendants'}>
          <Details client={client} id={selected.id} work={selected.work} kind={kind}
            tv active={playing === undefined} close={() => setSelected(undefined)} play={setPlaying} />
        </View>
      )}
      {!selected && playing === undefined && (
        <View style={{ flex: 1, flexDirection: 'row', backgroundColor: tvColors.surface }}>
          <Navigation
            kind={kind}
            select={selectKind}
            logout={logout}
            preferred={!lastCard.current && !lastRecent.current}
          />
          <Screen
            tv
            backgroundColor={tvColors.surface}
            scrollY={scroll.current}
            scrolled={(value) => {
              scroll.current = value;
            }}
          >
            <View style={styles.row}>
              <View style={{ flex: 1 }}>
                <Input
                  accessibilityLabel="搜索媒体"
                  placeholder="搜索片名"
                  value={query}
                  onChangeText={(value) => {
                    setQuery(value);
                    setPage(1);
                  }}
                />
              </View>
              <Button onPress={list.reload}>刷新</Button>
            </View>
            {list.recent.length > 0 && (
              <View style={{ gap: 12 }}>
                <Text style={styles.subtitle}>继续观看</Text>
                <ScrollView horizontal contentContainerStyle={{ gap: 12 }}>
                  {list.recent.map((item) => (
                    <Button key={item.id} preferred={lastRecent.current === item.id} onPress={() => {
                      lastRecent.current = item.id;
                      lastCard.current = undefined;
                      setPlaying(item.id);
                    }}>
                      {item.title}
                    </Button>
                  ))}
                </ScrollView>
              </View>
            )}
            <ErrorText error={list.error} />
            {list.loading && <Busy />}
            <View style={[styles.row, { gap: 12, alignItems: 'flex-start' }]}>
              {list.items.map((item) => (
                <Poster
                  key={item.id}
                  client={client}
                  item={item}
                  width={posterWidth}
                  preferred={lastCard.current === item.id}
                  press={() => {
                    lastCard.current = item.id;
                    lastRecent.current = undefined;
                    setSelected({ id: kind === 'tv' && 'work_id' in item && item.work_id ? item.work_id : item.id,
                      work: kind === 'tv' && 'work_id' in item && !!item.work_id });
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
        </View>
      )}
    </View>
  );
}

export function Home(props: { account: Account; logout: () => void }) {
  return <HomeSession key={`${props.account.key}:${props.account.token}`} {...props} />;
}

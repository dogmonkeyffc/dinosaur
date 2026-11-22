import { useEffect, useState } from 'react';
import { ScrollView, Text, View, useWindowDimensions } from 'react-native';
import type { Account, MediaLibrary } from '@dinosaur/contracts';
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
import { HomeHeader, LibraryTabs } from './HomeHeader';

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
  const [libraryID, setLibraryID] = useState<number>();
  const [libraries, setLibraries] = useState<MediaLibrary[]>();
  const [libraryError, setLibraryError] = useState('');
  const [libraryRevision, setLibraryRevision] = useState(0);
  const library = libraries?.find((item) => item.id === libraryID) ?? libraries?.[0];
  const kind = library?.type ?? 'movie';
  useEffect(() => {
    let current = true;
    setLibraryError('');
    client.libraries().then(({ libraries: next }) => {
      if (current) setLibraries(next);
    }).catch((reason) => {
      if (current) setLibraryError(reason instanceof Error ? reason.message : '无法读取媒体库。');
    });
    return () => { current = false; };
  }, [client, libraryRevision]);
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<{ id: number; work: boolean }>();
  const [playing, setPlaying] = useState<number>();
  const list = useMediaPage(client, kind, page, query, library?.id ?? null);
  const { width } = useWindowDimensions();
  const columns = width > 700 ? 4 : width >= 380 ? 3 : 2;
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
        <Screen backgroundColor="#121510">
          <HomeHeader accounts={accounts} pair={pair} />
          <Input
            accessibilityLabel="搜索媒体"
            placeholder="搜索片名"
            value={query}
            onChangeText={(value) => {
              setQuery(value);
              setPage(1);
            }}
          />
          {list.recent.length > 0 && !query.trim() && (
            <View style={{ gap: 6 }}>
              <Text accessibilityRole="header" style={styles.subtitle}>继续观看</Text>
              <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={{ gap: 8 }}>
                {list.recent.slice(0, 8).map((item) => (
                  <Poster key={item.id} item={item} client={client} width={88} press={() => setPlaying(item.id)} />
                ))}
              </ScrollView>
            </View>
          )}
          <LibraryTabs libraries={libraries ?? []} selected={library?.id} select={(id) => {
            setLibraryID(id);
            setPage(1);
          }} refresh={() => {
            setLibraryRevision((value) => value + 1);
            list.reload();
          }} />
          <ErrorText error={libraryError || list.error} />
          {(!libraries && !libraryError || list.loading) && <Busy />}
          {libraries?.length === 0 && <Text style={styles.muted}>暂无媒体库</Text>}
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
          {library && !list.loading && !list.error && !list.items.length && <Text style={styles.muted}>暂无媒体</Text>}
          {library && list.total > 24 && <View style={styles.row}>
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
          </View>}
        </Screen>
      )}
    </View>
  );
}

export function Home(props: { account: Account; pair: () => void; accounts: () => void }) {
  return <HomeSession key={`${props.account.key}:${props.account.token}`} {...props} />;
}

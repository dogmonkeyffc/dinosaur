import { useEffect, useRef, useState } from 'react';
import {
  BackHandler,
  Keyboard,
  Platform,
  RefreshControl,
  ScrollView,
  StatusBar,
  Text,
  TextInput,
  View,
} from 'react-native';
import type { Account, MediaLibrary } from '@dinosaur/contracts';
import { Busy, ErrorText, Player, useClient, useMediaPage } from '@dinosaur/native-kit';
import { Details } from './Details';
import { HomeHeader, HomeNavigation, LibraryTabs, type HomeTab } from './HomeHeader';
import { HomeFeed, MediaGrid, Pagination, type HomeMedia } from './HomeContent';
import { homeStyles as css, palette } from './homeStyles';

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
  const [libraryLoading, setLibraryLoading] = useState(true);
  const [refreshing, setRefreshing] = useState(false);
  const [tab, setTab] = useState<HomeTab>('home');
  const library = libraries?.find((item) => item.id === libraryID) ?? libraries?.[0];
  const kind = library?.type ?? 'movie';
  const scroll = useRef<ScrollView>(null);
  const searchInput = useRef<TextInput>(null);
  useEffect(() => {
    let current = true;
    setLibraryError('');
    setLibraryLoading(true);
    client
      .libraries()
      .then(({ libraries: next }) => {
        if (current) setLibraries(next);
      })
      .catch((reason) => {
        if (current) setLibraryError(reason instanceof Error ? reason.message : '无法读取媒体库。');
      })
      .finally(() => {
        if (current) setLibraryLoading(false);
      });
    return () => {
      current = false;
    };
  }, [client, libraryRevision]);
  const [page, setPage] = useState(1);
  const [query, setQuery] = useState('');
  const [selected, setSelected] = useState<{ id: number; work: boolean }>();
  const [playing, setPlaying] = useState<number>();
  const list = useMediaPage(client, kind, page, query, library?.id ?? null);
  useEffect(() => {
    if (!libraryLoading && !list.loading) setRefreshing(false);
  }, [libraryLoading, list.loading]);
  useEffect(() => {
    if (selected || playing !== undefined || tab === 'home') return;
    const listener = BackHandler.addEventListener('hardwareBackPress', () => {
      navigate('home');
      return true;
    });
    return () => listener.remove();
  }, [tab, selected, playing]);
  function refresh() {
    if (refreshing) return;
    setRefreshing(true);
    setLibraryLoading(true);
    setLibraryRevision((value) => value + 1);
    list.reload();
  }
  function navigate(next: HomeTab) {
    if (next === 'search') searchInput.current?.focus();
    else Keyboard.dismiss();
    setTab(next);
    setQuery('');
    setPage(1);
    scroll.current?.scrollTo({ y: 0, animated: false });
  }
  function open(item: HomeMedia) {
    const work = kind === 'tv' && 'work_id' in item && !!item.work_id;
    setSelected({ id: work && 'work_id' in item ? item.work_id! : item.id, work });
  }
  function changePage(value: number) {
    setPage(value);
    scroll.current?.scrollTo({ y: 0, animated: false });
  }
  return (
    <View style={css.screen}>
      {playing !== undefined && (
        <Player
          client={client}
          id={playing}
          close={() => {
            setPlaying(undefined);
            list.reload();
          }}
        />
      )}
      {selected && (
        <View
          style={{ flex: 1, display: playing === undefined ? 'flex' : 'none' }}
          accessibilityElementsHidden={playing !== undefined}
          importantForAccessibility={playing === undefined ? 'auto' : 'no-hide-descendants'}
        >
          <Details
            client={client}
            id={selected.id}
            work={selected.work}
            kind={kind}
            active={playing === undefined}
            close={() => setSelected(undefined)}
            play={setPlaying}
          />
        </View>
      )}
      {!selected && playing === undefined && (
        <View style={[css.screen, { paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 52 }]}>
          <StatusBar barStyle="light-content" />
          <HomeHeader
            accounts={accounts}
            pair={pair}
            username={account.user.username}
            inputRef={searchInput}
            query={query}
            focusSearch={() => setTab('search')}
            search={(value) => {
              setQuery(value);
              setPage(1);
              setTab('search');
              scroll.current?.scrollTo({ y: 0, animated: false });
            }}
          />
          <LibraryTabs
            libraries={libraries ?? []}
            selected={library?.id}
            select={(id) => {
              setLibraryID(id);
              changePage(1);
            }}
          />
          <ScrollView
            ref={scroll}
            key={tab}
            keyboardShouldPersistTaps="handled"
            contentContainerStyle={css.content}
            showsVerticalScrollIndicator={false}
            refreshControl={
              <RefreshControl
                refreshing={refreshing}
                onRefresh={refresh}
                colors={[palette.accent]}
                tintColor={palette.accent}
              />
            }
          >
            {!!(
              libraryError ||
              list.error ||
              (!refreshing && ((!libraries && libraryLoading) || list.loading))
            ) && (
              <View style={css.feedback}>
                <ErrorText error={libraryError || list.error} />
                {!refreshing && ((!libraries && libraryLoading) || list.loading) && <Busy />}
              </View>
            )}
            {libraries?.length === 0 && <Text style={css.message}>暂无媒体库</Text>}
            {tab === 'home' ? (
              <HomeFeed
                key={library?.id}
                items={list.items}
                recent={list.recent}
                title={library?.name ?? '媒体库'}
                client={client}
                open={open}
                play={setPlaying}
                browse={() => navigate('library')}
              />
            ) : (
              <>
                <Text accessibilityRole="header" style={[css.sectionTitle, { paddingHorizontal: 24 }]}>
                  {tab === 'search' ? (query.trim() ? '搜索结果' : '浏览片库') : (library?.name ?? '片库')}
                </Text>
                <MediaGrid items={list.items} client={client} open={open} />
              </>
            )}
            {library && !list.loading && !list.error && !list.items.length && (
              <Text style={css.message}>{query.trim() ? '没有找到相关媒体' : '暂无媒体'}</Text>
            )}
            {tab !== 'home' && library && (
              <Pagination page={page} total={list.total} loading={list.loading} change={changePage} />
            )}
          </ScrollView>
          <HomeNavigation tab={tab} select={navigate} />
        </View>
      )}
    </View>
  );
}

export function Home(props: { account: Account; pair: () => void; accounts: () => void }) {
  return <HomeSession key={`${props.account.key}:${props.account.token}`} {...props} />;
}

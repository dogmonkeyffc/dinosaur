import { useEffect, useRef, useState } from 'react';
import { BackHandler, Keyboard, Platform, ScrollView, StatusBar, TextInput, View } from 'react-native';
import type { Account, MediaKind } from '@dinosaur/contracts';
import { Player } from '@dinosaur/native-kit';
import { AccountSettings } from '../settings/AccountSettings';
import { useHomeVisibility } from '../settings/useHomeVisibility';
import { HistoryTimeline } from '../history/HistoryTimeline';
import { Details } from './Details';
import { HomeHeader, HomeNavigation, LibraryTabs, type HomeTab } from './HomeHeader';
import type { HomeMedia } from './HomeContent';
import { HomeBody } from './HomeBody';
import { LibraryLanding } from './LibraryLanding';
import { useMobileCatalog } from './useMobileCatalog';
import { homeStyles as css } from './homeStyles';

function HomeSession({
  account,
  pair,
  accounts,
}: {
  account: Account;
  pair: () => void;
  accounts: () => void;
}) {
  const [tab, setTab] = useState<HomeTab>('home');
  const [settings, setSettings] = useState(false);
  const [selected, setSelected] = useState<{ id: number; work: boolean; kind: MediaKind }>();
  const [playing, setPlaying] = useState<number>();
  const visibility = useHomeVisibility(account.key);
  const c = useMobileCatalog(account, tab, visibility.hidden, visibility.ready);
  const scroll = useRef<ScrollView>(null);
  const searchInput = useRef<TextInput>(null);
  const searchRequested = useRef(false);
  useEffect(() => {
    if (tab === 'search' && searchRequested.current) {
      searchInput.current?.focus();
      searchRequested.current = false;
    }
  }, [tab]);
  useEffect(() => {
    if (selected || playing !== undefined || settings || tab === 'home') return;
    const listener = BackHandler.addEventListener('hardwareBackPress', () => {
      navigate('home');
      return true;
    });
    return () => listener.remove();
  }, [tab, selected, playing, settings]);
  function navigate(next: HomeTab) {
    searchRequested.current = next === 'search';
    if (next === 'search') searchInput.current?.focus();
    else Keyboard.dismiss();
    setTab(next);
    c.setQuery('');
    c.setPage(1);
    scroll.current?.scrollTo({ y: 0, animated: false });
  }
  function open(item: HomeMedia) {
    const work = c.kind === 'tv' && 'work_id' in item && !!item.work_id;
    setSelected({ id: work && 'work_id' in item ? item.work_id! : item.id, work, kind: c.kind });
  }
  function changePage(value: number) {
    c.setPage(value);
    scroll.current?.scrollTo({ y: 0, animated: false });
  }
  function selectLibrary(id: number) {
    c.selectLibrary(id);
    scroll.current?.scrollTo({ y: 0, animated: false });
  }
  function showSettings() {
    Keyboard.dismiss();
    setSettings(true);
  }
  return (
    <View style={css.screen}>
      {playing !== undefined && (
        <Player
          client={c.client}
          id={playing}
          close={() => {
            setPlaying(undefined);
            c.list.reload();
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
            client={c.client}
            id={selected.id}
            work={selected.work}
            kind={selected.kind}
            active={playing === undefined}
            close={() => setSelected(undefined)}
            play={setPlaying}
          />
        </View>
      )}
      {!selected && playing === undefined && (
        <View style={css.screen}>
          <StatusBar barStyle="light-content" />
          {tab === 'history' ? (
            <HistoryTimeline client={c.client} username={account.user.username} settings={showSettings} />
          ) : tab === 'library' ? (
            <LibraryLanding
              key="library"
              catalog={c}
              username={account.user.username}
              settings={showSettings}
              open={open}
              scroll={scroll}
              select={selectLibrary}
              changePage={changePage}
            />
          ) : (
            <>
              <View style={{ paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 52 }}>
                <HomeHeader
                  settings={showSettings}
                  username={account.user.username}
                  inputRef={searchInput}
                  query={c.query}
                  focusSearch={() => setTab('search')}
                  search={(value) => {
                    c.setQuery(value);
                    c.setPage(1);
                    setTab('search');
                    scroll.current?.scrollTo({ y: 0, animated: false });
                  }}
                />
                <LibraryTabs libraries={c.available} selected={c.library?.id} select={selectLibrary} />
              </View>
              <HomeBody
                key={tab}
                catalog={c}
                home={tab === 'home'}
                visibility={visibility}
                scroll={scroll}
                open={open}
                play={setPlaying}
                browse={() => navigate('library')}
                changePage={changePage}
              />
            </>
          )}
          <HomeNavigation tab={tab} select={navigate} />
        </View>
      )}
      {settings && (
        <AccountSettings
          account={account}
          client={c.client}
          libraries={c.libraries}
          visibility={visibility}
          close={() => setSettings(false)}
          accounts={accounts}
          pair={pair}
          loading={c.loading}
          error={c.error}
          retry={c.refresh}
        />
      )}
    </View>
  );
}

export function Home(props: { account: Account; pair: () => void; accounts: () => void }) {
  return <HomeSession key={`${props.account.key}:${props.account.token}`} {...props} />;
}

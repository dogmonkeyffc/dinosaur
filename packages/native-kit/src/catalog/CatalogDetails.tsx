import { useEffect, useState } from 'react';
import { BackHandler, View, useWindowDimensions } from 'react-native';
import { ErrorText, Screen } from '../components';
import type { Client } from '@dinosaur/client';
import type { MediaKind } from '@dinosaur/contracts';
import { WorkDetails } from './WorkDetails';
import { SeriesDetails } from './SeriesDetails';

export function CatalogDetails({ client, id, kind, work = false, tv = false,
  active = true, close, play }: {
  client: Client;
  id: number;
  kind: MediaKind;
  work?: boolean;
  tv?: boolean;
  active?: boolean;
  close: () => void;
  play: (id: number) => void;
}) {
  const [member, setMember] = useState<number>();
  const [lastMember, setLastMember] = useState<number>();
  const [season, setSeason] = useState<number>();
  const [lastSeason, setLastSeason] = useState<number>();
  const { width } = useWindowDimensions();
  const columns = tv ? 5 : width > 700 ? 4 : 2;
  const posterWidth = Math.max(80, Math.floor((width - (tv ? 128 : 56)) / columns) - 8);
  const back = () => {
    if (season !== undefined) setSeason(undefined);
    else if (work && member !== undefined) setMember(undefined);
    else close();
  };
  useEffect(() => {
    if (!active) return;
    const handler = BackHandler.addEventListener('hardwareBackPress', () => {
      back();
      return true;
    });
    return () => handler.remove();
  }, [active, season, member, work, close]);
  const showWork = work && member === undefined;
  if (!Number.isSafeInteger(id) || id <= 0) return (
    <Screen tv={tv} title="媒体详情" onBack={close} backPreferred={tv && active}>
      <ErrorText error="无效的媒体。" />
    </Screen>
  );
  return (
    <View style={{ flex: 1 }}>
      {work && (
        <View style={{ flex: 1, display: showWork ? 'flex' : 'none' }}
          accessibilityElementsHidden={!showWork || !active}
          importantForAccessibility={showWork && active ? 'auto' : 'no-hide-descendants'}>
          <WorkDetails client={client} id={id} tv={tv} active={active && showWork}
            member={lastMember} width={posterWidth} close={close} select={(value) => {
              setMember(value);
              setLastMember(value);
              setSeason(undefined);
              setLastSeason(undefined);
            }} />
        </View>
      )}
      {!showWork && (
        <SeriesDetails key={member ?? id} client={client} id={member ?? id} kind={kind}
          tv={tv} active={active} season={season} lastSeason={lastSeason} width={posterWidth}
          close={back} play={play} selectSeason={(value) => {
            setSeason(value);
            setLastSeason(value);
          }} />
      )}
    </View>
  );
}

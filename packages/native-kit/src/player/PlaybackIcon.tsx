import Svg, { Path } from 'react-native-svg';

const paths = {
  back: 'm10 5-7 7 7 7M3 12h18',
  play: 'm8 5 11 7-11 7Z',
  pause: 'M7 5v14M17 5v14',
  next: 'm5 4 11 8-11 8ZM19 4v16',
  backward: 'M4 8V3M4 8h5M4 8a9 9 0 1 1-1 8',
  forward: 'M20 8V3M20 8h-5M20 8a9 9 0 1 0 1 8',
  audio: 'M9 5 4 9H2v6h2l5 4V5ZM13 8a6 6 0 0 1 0 8m3-11a10 10 0 0 1 0 14',
  subtitle:
    'M4 4h16a2 2 0 0 1 2 2v12a2 2 0 0 1-2 2H4a2 2 0 0 1-2-2V6a2 2 0 0 1 2-2ZM6 12h4m4 0h4M6 16h7m3 0h2',
  mode: 'M4 6h16M4 12h16M4 18h16M9 4v4m6 2v4m-7 2v4',
  quality: 'M4 4h16v16H4ZM7 9v6m4-6v6M7 12h4m4-3h2a3 3 0 0 1 0 6h-2V9Z',
  rate: 'M4 18a9 9 0 1 1 16 0M12 12l5-5M5 18h14',
  fullscreen: 'M3 9V3h6m6 0h6v6m0 6v6h-6m-6 0H3v-6',
  collapse: 'M9 3v6H3m12-6v6h6m0 6h-6v6m-6 0v-6H3',
  close: 'm6 6 12 12M6 18 18 6',
};
export type PlaybackIconName = keyof typeof paths;

export function PlaybackIcon({
  name,
  color = '#f2f6fc',
  size = 22,
}: {
  name: PlaybackIconName;
  color?: string;
  size?: number;
}) {
  return (
    <Svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke={color}
      strokeWidth={name === 'pause' ? 3 : 1.8}
      strokeLinecap="round"
      strokeLinejoin="round"
    >
      <Path d={paths[name]} fill={name === 'play' || name === 'next' ? color : 'none'} />
    </Svg>
  );
}

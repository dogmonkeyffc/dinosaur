import { Box, Typography } from '@mui/material';
import type { ReactNode } from 'react';
import { bytes } from '../lib/format';
import type { PlaybackMetricRow, PlaybackMode, PlaybackQuality } from '../playback/metric-sample';

const modes: Record<PlaybackMode, string> = { direct: '直接播放', remux: '兼容封装', transcode: '兼容转码' };
const qualities: Record<PlaybackQuality, string> = {
  auto: '自动',
  original: '原画',
  '1080': '1080p',
  '720': '720p',
  '480': '480p',
};
const bitrates = new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 2 });

export const resolution = ({ width = 0, height = 0 }: { width?: number; height?: number }) =>
  width > 0 && height > 0 ? `${width} × ${height}` : '—';
export const modeLabel = (item: PlaybackMetricRow) =>
  modes[item.mode] + (item.burn_subtitle ? ' · 烧录字幕' : '');
export const qualityLabel = (item: PlaybackMetricRow) => qualities[item.quality];

function formatBitrate(bps = 0) {
  if (bps >= 1_000_000) return `${bitrates.format(bps / 1_000_000)} Mbps`;
  if (bps >= 1_000) return `${bitrates.format(bps / 1_000)} kbps`;
  return bps > 0 ? `${bitrates.format(bps)} bps` : '—';
}

function Fields({ entries }: { entries: [string, ReactNode][] }) {
  return (
    <Box
      component="dl"
      sx={{
        display: 'grid',
        gridTemplateColumns: 'max-content minmax(0, 1fr)',
        gap: '4px 12px',
        m: 0,
        '& dt': { color: 'text.secondary' },
        '& dd': { m: 0, overflowWrap: 'anywhere' },
      }}
    >
      {entries.map(([label, value]) => (
        <div key={label} style={{ display: 'contents' }}>
          <dt>{label}</dt>
          <dd>{value}</dd>
        </div>
      ))}
    </Box>
  );
}

export function PlaybackMetricDetails({ item }: { item: PlaybackMetricRow }) {
  const { source, reason } = item;
  return (
    <Box
      sx={{
        display: 'grid',
        gridTemplateColumns: { xs: '1fr', sm: 'repeat(2, minmax(0, 1fr))' },
        gap: 2,
        maxWidth: 'min(62rem, calc(100vw - 5rem))',
        whiteSpace: 'normal',
      }}
    >
      <div>
        <Typography component="h3" variant="subtitle2" sx={{ mb: 0.5 }}>
          片源
        </Typography>
        {source ? (
          <Fields
            entries={[
              ['标题', source.title || '—'],
              ['绝对路径', source.path ? <code>{source.path}</code> : '—'],
              ['分辨率', resolution(source)],
              ['视频编码', source.video_codec || '—'],
              ['音频编码', source.audio_codec || '—'],
              ['容器', source.container || '—'],
              ['码率', formatBitrate(source.bitrate_bps)],
              ['文件大小', source.size_bytes ? bytes(source.size_bytes) : '—'],
            ]}
          />
        ) : (
          <Typography variant="body2" color="text.secondary">
            未记录片源
          </Typography>
        )}
      </div>
      <div>
        <Typography component="h3" variant="subtitle2" sx={{ mb: 0.5 }}>
          播放输出
        </Typography>
        <Fields
          entries={[
            ['分辨率', resolution(item)],
            ['播放模式', modeLabel(item)],
            ['画质设置', qualityLabel(item)],
            ['播放原因', reason || '—'],
          ]}
        />
      </div>
    </Box>
  );
}

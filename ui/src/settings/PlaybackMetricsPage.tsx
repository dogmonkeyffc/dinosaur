import {
  Button,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
} from '@mui/material';
import { ErrorMessage, Heading, Loading } from '../components/common';
import { time } from '../lib/format';
import { useResource } from '../lib/hooks';
import type { PlaybackMetricRow, PlaybackMode, PlaybackQuality } from '../playback/metric-sample';

const modes: Record<PlaybackMode, string> = { direct: '直接播放', remux: '兼容封装', transcode: '兼容转码' };
const qualities: Record<PlaybackQuality, string> = {
  auto: '自动',
  original: '原画',
  '1080': '1080p',
  '720': '720p',
  '480': '480p',
};
const seconds = new Intl.NumberFormat('zh-CN', { maximumFractionDigits: 2 });
const duration = (ms: number | null | undefined) => (ms == null ? '—' : `${seconds.format(ms / 1000)} 秒`);

export function PlaybackMetricsPage() {
  const { data, error, loading, reload } = useResource<{ items: PlaybackMetricRow[] }>(
    '/api/admin/playback/metrics',
  );
  const items = data?.items ?? [];
  return (
    <section id="playback-metrics-section" aria-busy={loading}>
      <Heading title="播放指标">
        <Button variant="outlined" disabled={loading} onClick={reload}>
          刷新
        </Button>
      </Heading>
      <ErrorMessage error={error} />
      {loading && !items.length && <Loading />}
      {!loading && !error && !items.length && (
        <p className="muted" role="status">
          暂无播放指标
        </p>
      )}
      {!!items.length && (
        <TableContainer component={Paper} tabIndex={0} role="region" aria-label="最近播放指标">
          <Table size="small" aria-label="播放指标" sx={{ whiteSpace: 'nowrap' }}>
            <TableHead>
              <TableRow>
                <TableCell>更新时间</TableCell>
                <TableCell>播放模式</TableCell>
                <TableCell>画质</TableCell>
                <TableCell>分辨率</TableCell>
                <TableCell align="right">起播耗时</TableCell>
                <TableCell align="right">播放时长</TableCell>
                <TableCell align="right">卡顿次数</TableCell>
                <TableCell align="right">卡顿时长</TableCell>
                <TableCell align="right">剩余缓冲</TableCell>
                <TableCell align="right">丢帧 / 总帧</TableCell>
                <TableCell align="right">FFmpeg 转码速度</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {items.map((item) => (
                <TableRow key={item.id}>
                  <TableCell>{time(item.updated_at)}</TableCell>
                  <TableCell>
                    {modes[item.mode]}
                    {item.burn_subtitle ? ' · 烧录字幕' : ''}
                  </TableCell>
                  <TableCell>{qualities[item.quality]}</TableCell>
                  <TableCell>{item.width && item.height ? `${item.width} × ${item.height}` : '—'}</TableCell>
                  <TableCell align="right">{duration(item.startup_ms)}</TableCell>
                  <TableCell align="right">{duration(item.played_ms)}</TableCell>
                  <TableCell align="right">{item.stalls}</TableCell>
                  <TableCell align="right">{duration(item.stall_ms)}</TableCell>
                  <TableCell align="right">{duration(item.buffer_ms)}</TableCell>
                  <TableCell align="right">
                    {item.dropped_frames == null && item.total_frames == null
                      ? '—'
                      : `${item.dropped_frames ?? '—'} / ${item.total_frames ?? '—'}`}
                  </TableCell>
                  <TableCell align="right">
                    {item.transcode_speed == null ? '—' : `${seconds.format(item.transcode_speed)}×`}
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      )}
    </section>
  );
}

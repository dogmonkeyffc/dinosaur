import type { WatchHistoryEntry } from '@dinosaur/contracts';

export function timelineSections(items: WatchHistoryEntry[]) {
  const groups = new Map<string, WatchHistoryEntry[]>();
  for (const item of items) {
    const month = item.watched_on.slice(0, 7);
    const group = groups.get(month) ?? [];
    group.push(item);
    groups.set(month, group);
  }
  return [...groups.entries()]
    .sort(([a], [b]) => b.localeCompare(a))
    .map(([month, data]) => ({
      title: `${month.slice(0, 4)}年${Number(month.slice(5))}月`,
      data: data.sort((a, b) => b.watched_on.localeCompare(a.watched_on) || b.id - a.id),
    }));
}

export function mergeTimeline(previous: WatchHistoryEntry[], next: WatchHistoryEntry[]) {
  return [...new Map([...previous, ...next].map((item) => [item.id, item])).values()];
}

export interface Cue {
  start: number;
  end: number;
  text: string;
}
function seconds(raw: string) {
  const values = raw.replace(',', '.').split(':').map(Number);
  if (values.length < 2 || values.length > 3 || values.some((value) => !Number.isFinite(value))) return NaN;
  return values.reduce((total, value) => total * 60 + value, 0);
}
export function parseVTT(source: string): Cue[] {
  const result: Cue[] = [];
  for (const block of source.replace(/\r/g, '').split(/\n{2,}/)) {
    const lines = block.split('\n');
    const index = lines.findIndex((line) => line.includes(' --> '));
    if (index < 0) continue;
    const [from, to] = lines[index].split(' --> ');
    const start = seconds(from.trim());
    const end = seconds(to.trim().split(/\s/)[0]);
    if (!Number.isFinite(start) || !Number.isFinite(end) || start < 0 || end <= start) continue;
    const text = lines
      .slice(index + 1)
      .join('\n')
      .replace(/<[^>]*>/g, '')
      .replace(/&lt;/g, '<')
      .replace(/&gt;/g, '>')
      .replace(/&amp;/g, '&');
    result.push({ start, end, text });
  }
  return result.sort((a, b) => a.start - b.start);
}
export function activeCue(cues: Cue[], position: number) {
  return cues
    .filter((cue) => position >= cue.start && position < cue.end)
    .map((cue) => cue.text)
    .join('\n');
}

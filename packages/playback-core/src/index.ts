export class ProgressQueue {
  private pending: Promise<unknown> = Promise.resolve();
  constructor(private write: (position: number, watched?: boolean) => Promise<unknown>) {}
  save(position: number, watched?: boolean) {
    if (!Number.isFinite(position) || position < 0) return this.pending;
    this.pending = this.pending.catch(() => {}).then(() => this.write(position, watched));
    return this.pending;
  }
  flush() {
    return this.pending;
  }
}

export function sourcePosition(currentTime: number, offset: number, duration: number) {
  const position = Math.max(0, currentTime + offset);
  return duration > 0 ? Math.min(duration, position) : position;
}

export function clock(seconds: number) {
  const total = Math.floor(Math.max(0, seconds));
  return `${Math.floor(total / 3600) ? Math.floor(total / 3600) + ':' : ''}${String(Math.floor(total / 60) % 60).padStart(2, '0')}:${String(total % 60).padStart(2, '0')}`;
}

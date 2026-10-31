import { stateNames } from '../lib/format';

export function StatusBadge({ status }: { status: string }) {
  return <span className={`scrape-state scrape-state-${status}`}>{stateNames[status] || status}</span>;
}

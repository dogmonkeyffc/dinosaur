import { Chip } from '@mui/material';
import { stateNames } from '../lib/format';

export function StatusBadge({ status }: { status: string }) {
  const color =
    status === 'failed' || status === 'interrupted'
      ? 'error'
      : status === 'needs_review' || status === 'not_processed'
        ? 'warning'
        : 'default';
  return <Chip size="small" color={color} label={stateNames[status] || status} />;
}

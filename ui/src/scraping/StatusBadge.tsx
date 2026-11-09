import { Badge } from '@fluentui/react-components';
import { stateNames } from '../lib/format';

export function StatusBadge({ status }: { status: string }) {
  const color =
    status === 'failed' || status === 'interrupted'
      ? 'danger'
      : status === 'needs_review' || status === 'not_processed'
        ? 'warning'
        : 'informative';
  return (
    <Badge appearance="tint" color={color}>
      {stateNames[status] || status}
    </Badge>
  );
}

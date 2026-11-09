import { IconButton, Popover, Typography } from '@mui/material';
import { useId, useState } from 'react';
import { Icon } from './common';

export function InfoTip({ label, children }: { label: string; children: string }) {
  const [anchor, setAnchor] = useState<HTMLButtonElement | null>(null);
  const id = useId();
  return (
    <>
      <IconButton
        size="small"
        aria-label={label}
        aria-expanded={!!anchor}
        aria-controls={anchor ? id : undefined}
        onClick={(event) => setAnchor(event.currentTarget)}
      >
        <Icon name="info" />
      </IconButton>
      <Popover
        id={id}
        open={!!anchor}
        anchorEl={anchor}
        onClose={() => setAnchor(null)}
        anchorOrigin={{ vertical: 'bottom', horizontal: 'left' }}
      >
        <Typography sx={{ p: 2, maxWidth: 320 }} variant="body2">
          {children}
        </Typography>
      </Popover>
    </>
  );
}

import { IconButton, Menu, MenuItem } from '@mui/material';
import { useState } from 'react';

export function LibraryCardMenu({ name, scanning, grant, edit, remove }: {
  name: string;
  scanning: boolean;
  grant: () => void;
  edit: () => void;
  remove: () => void;
}) {
  const [anchor, setAnchor] = useState<HTMLElement | null>(null);
  function choose(action: () => void) {
    setAnchor(null);
    action();
  }
  return (
    <>
      <IconButton
        aria-label={`${name} 的操作`}
        aria-haspopup="menu"
        aria-expanded={Boolean(anchor)}
        onClick={(event) => setAnchor(event.currentTarget)}
        size="small"
      >
        <span aria-hidden="true">⋯</span>
      </IconButton>
      <Menu anchorEl={anchor} open={Boolean(anchor)} onClose={() => setAnchor(null)}>
        <MenuItem onClick={() => choose(grant)}>给账号赋权</MenuItem>
        <MenuItem disabled={scanning} onClick={() => choose(edit)}>编辑媒体库</MenuItem>
        <MenuItem disabled={scanning} onClick={() => choose(remove)} sx={{ color: 'error.main' }}>移除媒体库</MenuItem>
      </Menu>
    </>
  );
}

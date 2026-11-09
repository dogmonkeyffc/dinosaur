import {
  Alert,
  Button,
  CircularProgress,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  Stack,
  Typography,
} from '@mui/material';
import { Link } from 'react-router-dom';
import type { ReactNode } from 'react';
import sprite from '../assets/icons.svg?url';

export function Icon({ name }: { name: string }) {
  return (
    <svg className="icon" aria-hidden="true">
      <use href={`${sprite}#i-${name}`} />
    </svg>
  );
}

export function Brand() {
  return (
    <Link className="brand" to="/">
      <span className="brand-mark">
        <Icon name="logo" />
      </span>
      <span>Dinosaur</span>
    </Link>
  );
}

export function Heading({ title, children }: { title: ReactNode; children?: ReactNode }) {
  return (
    <header className="page-heading">
      <Typography variant="h4" component="h1">
        {title}
      </Typography>
      {children}
    </header>
  );
}

export function ErrorMessage({ error, id }: { error?: string; id?: string }) {
  return error ? (
    <Alert id={id} className="form-error" severity="error">
      {error}
    </Alert>
  ) : null;
}

export function Loading() {
  return (
    <Stack
      className="loading-message"
      role="status"
      direction="row"
      spacing={1}
      sx={{ alignItems: 'center' }}
    >
      <CircularProgress size={20} aria-hidden="true" />
      <span>正在读取…</span>
    </Stack>
  );
}

export function Pagination({
  page,
  total,
  size,
  onChange,
}: {
  page: number;
  total: number;
  size: number;
  onChange: (page: number) => void;
}) {
  if (total <= size) return null;
  return (
    <div className="movie-pagination">
      <Button variant="outlined" disabled={page <= 1} onClick={() => onChange(page - 1)}>
        上一页
      </Button>
      <span>
        共 {total} 条 · {page} / {Math.ceil(total / size)}
      </span>
      <Button variant="outlined" disabled={page * size >= total} onClick={() => onChange(page + 1)}>
        下一页
      </Button>
    </div>
  );
}

export function Modal({ title, children, close }: { title: string; children: ReactNode; close: () => void }) {
  return (
    <Dialog open fullWidth maxWidth="sm" onClose={close} aria-label={title}>
      <DialogTitle>
        <Stack direction="row" sx={{ alignItems: 'center', justifyContent: 'space-between' }}>
          {title}
          <IconButton aria-label="关闭" onClick={close}>
            <Icon name="close" />
          </IconButton>
        </Stack>
      </DialogTitle>
      <DialogContent>{children}</DialogContent>
    </Dialog>
  );
}

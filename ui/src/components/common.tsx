import {
  Button,
  Dialog,
  DialogBody,
  DialogContent,
  DialogSurface,
  DialogTitle,
  MessageBar,
  MessageBarBody,
  Spinner,
  useRestoreFocusSource,
} from '@fluentui/react-components';
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
    <a className="brand" href="/">
      <span className="brand-mark">
        <Icon name="logo" />
      </span>
      <span>
        dinosaur<span className="brand-period">.</span>
      </span>
    </a>
  );
}

export function Heading({ title, children }: { title: ReactNode; children?: ReactNode }) {
  return (
    <div className="page-heading">
      <h1>{title}</h1>
      {children}
    </div>
  );
}

export function ErrorMessage({ error, id }: { error?: string; id?: string }) {
  return error ? (
    <MessageBar id={id} className="form-error" intent="error">
      <MessageBarBody>{error}</MessageBarBody>
    </MessageBar>
  ) : null;
}

export function Loading() {
  return <Spinner className="loading-message" size="small" label="正在读取…" />;
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
      <Button className="button secondary" disabled={page <= 1} onClick={() => onChange(page - 1)}>
        上一页
      </Button>
      <span>
        共 {total} 条 · {page} / {Math.ceil(total / size)}
      </span>
      <Button className="button secondary" disabled={page * size >= total} onClick={() => onChange(page + 1)}>
        下一页
      </Button>
    </div>
  );
}

export function Modal({ title, children, close }: { title: string; children: ReactNode; close: () => void }) {
  const restoreFocus = useRestoreFocusSource();
  return (
    <Dialog
      open
      onOpenChange={(_, data) => {
        if (!data.open) close();
      }}
    >
      <DialogSurface className="modal">
        <DialogBody {...restoreFocus}>
          <DialogTitle
            action={
              <Button appearance="subtle" aria-label="关闭" onClick={close} icon={<Icon name="close" />} />
            }
          >
            {title}
          </DialogTitle>
          <DialogContent>{children}</DialogContent>
        </DialogBody>
      </DialogSurface>
    </Dialog>
  );
}

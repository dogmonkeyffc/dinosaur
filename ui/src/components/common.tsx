import { useEffect, useRef, type ReactNode } from 'react';
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

export function Heading({
  title,
  label,
  description,
  children,
}: {
  title: ReactNode;
  label?: string;
  description?: string;
  children?: ReactNode;
}) {
  return (
    <div className="page-heading">
      <div>
        <span className="small-label">{label}</span>
        <h1>{title}</h1>
        {description && <p className="muted">{description}</p>}
      </div>
      {children}
    </div>
  );
}

export function ErrorMessage({ error, id }: { error?: string; id?: string }) {
  return error ? (
    <p id={id} className="form-error" role="alert">
      {error}
    </p>
  ) : null;
}

export function Loading() {
  return (
    <p className="loading-message" role="status">
      正在读取…
    </p>
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
      <button className="button secondary" disabled={page <= 1} onClick={() => onChange(page - 1)}>
        上一页
      </button>
      <span>
        共 {total} 条 · {page} / {Math.ceil(total / size)}
      </span>
      <button className="button secondary" disabled={page * size >= total} onClick={() => onChange(page + 1)}>
        下一页
      </button>
    </div>
  );
}

export function Modal({ title, children, close }: { title: string; children: ReactNode; close: () => void }) {
  const dialog = useRef<HTMLDialogElement>(null);
  useEffect(() => {
    const element = dialog.current!;
    element.showModal();
    return () => element.close();
  }, []);
  return (
    <dialog ref={dialog} className="modal" aria-label={title} onCancel={close}>
      <div className="modal-heading">
        <h2>{title}</h2>
        <button className="icon-button" type="button" onClick={close} aria-label="关闭">
          <Icon name="close" />
        </button>
      </div>
      {children}
    </dialog>
  );
}

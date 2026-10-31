import { useId, useLayoutEffect, useRef, useState } from 'react';
import { Icon } from './common';
import '../styles/info-tip.css';

export function InfoTip({ label, children }: { label: string; children: string }) {
  const [open, setOpen] = useState(false);
  const id = useId();
  const root = useRef<HTMLSpanElement>(null);
  const button = useRef<HTMLButtonElement>(null);
  const bubble = useRef<HTMLSpanElement>(null);
  useLayoutEffect(() => {
    if (!open || !bubble.current || !button.current) return;
    const tooltip = bubble.current;
    tooltip.showPopover();
    const anchor = button.current.getBoundingClientRect();
    const bounds = tooltip.getBoundingClientRect();
    tooltip.style.left = `${Math.max(12, Math.min(anchor.left, innerWidth - bounds.width - 12))}px`;
    const top = anchor.bottom + 8;
    tooltip.style.top = `${Math.max(12, top + bounds.height < innerHeight - 12 ? top : anchor.top - bounds.height - 8)}px`;
    const outside = (event: PointerEvent) => {
      if (!root.current?.contains(event.target as Node)) setOpen(false);
    };
    const escape = (event: KeyboardEvent) => {
      if (event.key === 'Escape') {
        event.preventDefault();
        event.stopPropagation();
        setOpen(false);
      }
    };
    const close = () => setOpen(false);
    document.addEventListener('pointerdown', outside);
    document.addEventListener('keydown', escape, true);
    window.addEventListener('resize', close);
    window.addEventListener('scroll', close, true);
    return () => {
      tooltip.hidePopover();
      document.removeEventListener('pointerdown', outside);
      document.removeEventListener('keydown', escape, true);
      window.removeEventListener('resize', close);
      window.removeEventListener('scroll', close, true);
    };
  }, [open]);
  return (
    <span ref={root} className="info-tip">
      <button
        ref={button}
        type="button"
        aria-label={label}
        aria-expanded={open}
        aria-controls={id}
        aria-describedby={open ? id : undefined}
        onClick={() => setOpen(true)}
        onFocus={() => setOpen(true)}
        onBlur={() => setOpen(false)}
        onPointerEnter={(event) => {
          if (event.pointerType === 'mouse') setOpen(true);
        }}
        onPointerLeave={(event) => {
          if (event.pointerType === 'mouse' && !button.current?.matches(':focus-visible')) setOpen(false);
        }}
      >
        <Icon name="info" />
      </button>
      <span ref={bubble} id={id} role="tooltip" popover="manual" className="info-tip-content">
        {children}
      </span>
    </span>
  );
}
